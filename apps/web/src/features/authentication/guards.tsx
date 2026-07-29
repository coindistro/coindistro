"use client";

import * as React from "react";
import { useRouter, usePathname } from "next/navigation";
import { Spinner } from "@coindistro/cds";
import { useAuth } from "@/features/authentication/auth-provider";
import { hasAnyRole } from "@/features/authentication/roles";

function GuardSpinner({ label }: { label: string }) {
  return (
    <div className="flex min-h-[50vh] items-center justify-center">
      <Spinner label={label} />
    </div>
  );
}

/**
 * Requires an authenticated session. Anonymous users are sent to login.
 */
export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { loading, isAuthenticated } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  React.useEffect(() => {
    if (loading) return;
    if (!isAuthenticated) {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
    }
  }, [loading, isAuthenticated, router, pathname]);

  if (loading || !isAuthenticated) {
    return <GuardSpinner label="Checking session" />;
  }

  return <>{children}</>;
}

/**
 * Requires authentication and at least one of the given roles.
 * Unauthenticated → login. Authenticated without role → user dashboard.
 */
export function RoleGuard({
  children,
  roles,
  fallbackPath = "/dashboard",
}: {
  children: React.ReactNode;
  roles: string[];
  fallbackPath?: string;
}) {
  const { user, loading, isAuthenticated } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  const allowed = hasAnyRole(user?.roles, roles);

  React.useEffect(() => {
    if (loading) return;
    if (!isAuthenticated) {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
      return;
    }
    if (!allowed) {
      router.replace(`${fallbackPath}?error=forbidden`);
    }
  }, [loading, isAuthenticated, allowed, router, pathname, fallbackPath]);

  if (loading || !isAuthenticated) {
    return <GuardSpinner label="Checking session" />;
  }

  if (!allowed) {
    return <GuardSpinner label="Redirecting" />;
  }

  return <>{children}</>;
}

/** @deprecated Prefer AuthGuard / RoleGuard — kept for existing layouts. */
export function RequireAuth({
  children,
  roles,
}: {
  children: React.ReactNode;
  roles?: string[];
}) {
  if (roles?.length) {
    return <RoleGuard roles={roles}>{children}</RoleGuard>;
  }
  return <AuthGuard>{children}</AuthGuard>;
}
