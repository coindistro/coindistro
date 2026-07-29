/** Canonical role helpers for Coindistro authentication routing and guards. */

export const ADMIN_ROLES = ["super_admin", "admin", "moderator"] as const;
export const MERCHANT_ROLES = ["merchant"] as const;

export type AdminRole = (typeof ADMIN_ROLES)[number];
export type MerchantRole = (typeof MERCHANT_ROLES)[number];

export function hasAnyRole(
  userRoles?: string[] | null,
  required?: readonly string[] | string[] | null,
): boolean {
  if (!required?.length) return true;
  if (!userRoles?.length) return false;
  return required.some((r) => userRoles.includes(r));
}

export function isAdminRole(roles?: string[] | null): boolean {
  return hasAnyRole(roles, ADMIN_ROLES);
}

export function isMerchantRole(roles?: string[] | null): boolean {
  return hasAnyRole(roles, MERCHANT_ROLES);
}

export function isSuperAdminRole(roles?: string[] | null): boolean {
  return hasAnyRole(roles, ["super_admin"]);
}

/**
 * Home path after login/register based on roles.
 * Uses short paths that map (via redirects) onto the app shells.
 */
export function defaultHomeForRoles(roles?: string[] | null): string {
  if (isAdminRole(roles) && (roles?.includes("super_admin") || roles?.includes("admin"))) {
    return "/admin";
  }
  // Moderators still land on admin control plane
  if (isAdminRole(roles)) {
    return "/admin";
  }
  if (isMerchantRole(roles)) {
    return "/merchant";
  }
  return "/dashboard";
}

/**
 * Resolve post-auth destination.
 * Honors safe relative `next` only when the user is allowed to open that area.
 */
export function postLoginPath(
  roles?: string[] | null,
  next?: string | null,
): string {
  const home = defaultHomeForRoles(roles);

  if (!next || !next.startsWith("/") || next.startsWith("//")) {
    return home;
  }

  if (next.startsWith("/admin") && !isAdminRole(roles)) {
    return home;
  }
  if (
    (next.startsWith("/merchant") || next.startsWith("/app/merchant")) &&
    !isMerchantRole(roles) &&
    !isAdminRole(roles)
  ) {
    return home;
  }

  return next;
}

/** Lightweight permission list derived from roles (RBAC expansion point). */
export function permissionsForRoles(roles?: string[] | null): string[] {
  const set = new Set<string>();
  if (!roles?.length) return [];
  for (const role of roles) {
    set.add(`role:${role}`);
    if (role === "super_admin") {
      set.add("admin.access");
      set.add("admin.full");
    } else if (role === "admin" || role === "moderator") {
      set.add("admin.access");
    } else if (role === "merchant") {
      set.add("merchant.access");
    } else if (role === "user") {
      set.add("user.access");
    }
  }
  return Array.from(set);
}
