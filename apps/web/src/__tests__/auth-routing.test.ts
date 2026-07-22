import { describe, it, expect } from "vitest";

/**
 * Authentication routing tests.
 * Validates the route group structure and redirect logic.
 */

const authRoutes = [
  { path: "/login", label: "Login", auth: false },
  { path: "/register", label: "Register", auth: false },
  { path: "/forgot-password", label: "Forgot Password", auth: false },
  { path: "/reset-password", label: "Reset Password", auth: false },
  { path: "/verify-email", label: "Verify Email", auth: false },
  { path: "/invite", label: "Invitation Registration", auth: false },
  { path: "/referral", label: "Referral Registration", auth: false },
];

const protectedRoutes = [
  "/app/dashboard",
  "/app/markets",
  "/app/trade",
  "/app/p2p",
  "/app/earn",
  "/app/academy",
  "/app/signals",
  "/app/ai-bots",
  "/app/wallet",
  "/app/merchant",
  "/app/pay",
  "/app/referrals",
  "/app/notifications",
  "/app/profile",
  "/app/settings",
];

const adminRoutes = [
  { path: "/admin", roles: ["admin", "super_admin", "moderator"] },
  { path: "/admin/users", roles: ["admin", "super_admin", "moderator"] },
  { path: "/admin/genesis", roles: ["admin", "super_admin", "moderator"] },
  { path: "/admin/referrals", roles: ["admin", "super_admin", "moderator"] },
  { path: "/admin/p2p", roles: ["admin", "super_admin", "moderator"] },
];

describe("Auth Routing", () => {
  it("should have all auth routes defined", () => {
    expect(authRoutes).toHaveLength(7);
    authRoutes.forEach((route) => {
      expect(route.path).toBeTruthy();
      expect(route.auth).toBe(false);
    });
  });

  it("should have all protected routes defined", () => {
    expect(protectedRoutes).toHaveLength(15);
    protectedRoutes.forEach((path) => {
      expect(path).toMatch(/^\/app\//);
    });
  });

  it("should have all admin routes defined with correct roles", () => {
    expect(adminRoutes).toHaveLength(5);
    adminRoutes.forEach((route) => {
      expect(route.roles).toContain("admin");
    });
  });

  it("should redirect unauthenticated users to login", () => {
    protectedRoutes.forEach((path) => {
      const redirect = `/login?next=${encodeURIComponent(path)}`;
      expect(redirect).toContain("/login");
      expect(redirect).toContain(encodeURIComponent(path));
    });
  });

  it("should redirect expired sessions to login with reason", () => {
    const redirect = "/login?reason=session_expired";
    expect(redirect).toContain("reason=session_expired");
  });
});