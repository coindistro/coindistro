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
import { appConfig } from "@/lib/config";

interface AuthContextValue {
  user: AuthUser | null;
  loading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (input: authApi.RegisterInput) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
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

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [user, setUser] = React.useState<AuthUser | null>(null);
  const [loading, setLoading] = React.useState(true);

  const clearSession = React.useCallback(() => {
    setAuthTokens(null, null);
    persistUser(null);
    setUser(null);
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

  const value = React.useMemo<AuthContextValue>(
    () => ({
      user,
      loading,
      isAuthenticated: !!user && !!getAccessToken(),
      async login(email, password) {
        const data = await authApi.login({ email, password });
        setUser(data.user);
        persistUser(data.user);
      },
      async register(input) {
        const data = await authApi.register(input);
        setUser(data.user);
        persistUser(data.user);
      },
      async logout() {
        await authApi.logout();
        clearSession();
        router.replace("/login");
      },
      async refreshUser() {
        const me = await authApi.getMe();
        setUser(me);
        persistUser(me);
      },
    }),
    [user, loading, clearSession, router],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = React.useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
