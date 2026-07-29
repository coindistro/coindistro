"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import {
  loadAuthTokensFromStorage,
  setAuthTokens,
  setUnauthorizedHandler,
  getAccessToken,
} from "@/lib/api/client";
import type { AuthUser } from "@/lib/api/types";
import * as authApi from "@/features/authentication/api";
import {
  isAdminRole,
  isMerchantRole,
  permissionsForRoles,
} from "@/features/authentication/roles";
import { appConfig } from "@/lib/config";

interface AuthContextValue {
  user: AuthUser | null;
  roles: string[];
  permissions: string[];
  accessToken: string | null;
  loading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isMerchant: boolean;
  login: (email: string, password: string) => Promise<AuthUser>;
  register: (input: authApi.RegisterInput) => Promise<AuthUser>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<AuthUser | null>;
  /** Clear local session without calling the API (e.g. forced expiry). */
  clearSession: () => void;
}

const AuthContext = React.createContext<AuthContextValue | null>(null);

function persistUser(user: AuthUser | null) {
  if (typeof window === "undefined") return;
  if (user) localStorage.setItem(appConfig.userKey, JSON.stringify(user));
  else localStorage.removeItem(appConfig.userKey);
}

function readStoredUser(): AuthUser | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(appConfig.userKey);
    return raw ? (JSON.parse(raw) as AuthUser) : null;
  } catch {
    return null;
  }
}

function applySession(user: AuthUser, access: string, refresh: string) {
  setAuthTokens(access, refresh);
  persistUser(user);
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [user, setUser] = React.useState<AuthUser | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [tokenTick, setTokenTick] = React.useState(0);

  const clearSession = React.useCallback(() => {
    setAuthTokens(null, null);
    persistUser(null);
    setUser(null);
    setTokenTick((t) => t + 1);
  }, []);

  React.useEffect(() => {
    loadAuthTokensFromStorage();
    setUnauthorizedHandler(() => {
      clearSession();
      router.replace("/login?reason=session_expired");
    });

    const bootstrap = async () => {
      const token = getAccessToken();
      const cached = readStoredUser();
      if (cached) setUser(cached);
      if (!token) {
        setLoading(false);
        return;
      }
      try {
        // getMe uses the API client which auto-refreshes on 401 once
        const me = await authApi.getMe();
        setUser(me);
        persistUser(me);
      } catch {
        clearSession();
      } finally {
        setLoading(false);
      }
    };
    void bootstrap();
  }, [clearSession, router]);

  const roles = user?.roles ?? [];
  const permissions = React.useMemo(() => permissionsForRoles(roles), [roles]);
  // tokenTick forces re-read after login/logout
  const accessToken = tokenTick >= 0 ? getAccessToken() : null;

  const value = React.useMemo<AuthContextValue>(
    () => ({
      user,
      roles,
      permissions,
      accessToken,
      loading,
      isAuthenticated: !!user && !!getAccessToken(),
      isAdmin: isAdminRole(roles),
      isMerchant: isMerchantRole(roles),
      clearSession,
      async login(email, password) {
        const data = await authApi.login({ email, password });
        applySession(data.user, data.access_token, data.refresh_token);
        setUser(data.user);
        setTokenTick((t) => t + 1);
        return data.user;
      },
      async register(input) {
        const data = await authApi.register(input);
        // Tokens may be issued even when email is pending verification
        // so the user can call resend-verification while signed in.
        applySession(data.user, data.access_token, data.refresh_token);
        setUser(data.user);
        setTokenTick((t) => t + 1);
        return data.user;
      },
      async logout() {
        try {
          await authApi.logout();
        } finally {
          clearSession();
          router.replace("/login");
        }
      },
      async refreshUser() {
        const me = await authApi.getMe();
        setUser(me);
        persistUser(me);
        return me;
      },
    }),
    [user, roles, permissions, accessToken, loading, clearSession, router],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = React.useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}

/** Convenience: whether the current user may access admin routes. */
export function useIsAdmin() {
  return useAuth().isAdmin;
}
