"use client";

import * as React from "react";
import { useRouter, usePathname } from "next/navigation";
import { Spinner } from "@coindistro/cds";
import { useAuth } from "@/features/authentication/auth-provider";

export function RequireAuth({
  children,
  roles,
}: {
  children: React.ReactNode;
  roles?: string[];
}) {
  const { user, loading, isAuthenticated } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  React.useEffect(() => {
    if (loading) return;
    if (!isAuthenticated) {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
      return;
    }
    if (roles?.length) {
      const userRoles = user?.roles ?? [];
      const ok = roles.some((r) => userRoles.includes(r));
      if (!ok) router.replace("/app/dashboard?error=forbidden");
    }
  }, [loading, isAuthenticated, user, roles, router, pathname]);

  if (loading || !isAuthenticated) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <Spinner label="Checking session" />
      </div>
    );
  }

  if (roles?.length) {
    const userRoles = user?.roles ?? [];
    const ok = roles.some((r) => userRoles.includes(r));
    if (!ok) {
      return (
        <div className="flex min-h-[50vh] items-center justify-center">
          <Spinner label="Redirecting" />
        </div>
      );
    }
  }

  return <>{children}</>;
}
