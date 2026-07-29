import { describe, it, expect } from "vitest";
import {
  isAdminRole,
  isMerchantRole,
  postLoginPath,
  defaultHomeForRoles,
  permissionsForRoles,
} from "@/features/authentication/roles";
import { mapAuthError } from "@/features/authentication/auth-errors";
import { ApiError } from "@/lib/api/types";

describe("auth helpers", () => {
  it("detects admin roles", () => {
    expect(isAdminRole(["user"])).toBe(false);
    expect(isAdminRole(["admin"])).toBe(true);
    expect(isAdminRole(["super_admin"])).toBe(true);
    expect(isAdminRole(["moderator"])).toBe(true);
    expect(isAdminRole(null)).toBe(false);
  });

  it("detects merchant roles", () => {
    expect(isMerchantRole(["merchant"])).toBe(true);
    expect(isMerchantRole(["user"])).toBe(false);
  });

  it("routes after login by role", () => {
    expect(defaultHomeForRoles(["user"])).toBe("/dashboard");
    expect(defaultHomeForRoles(["admin"])).toBe("/admin");
    expect(defaultHomeForRoles(["super_admin"])).toBe("/admin");
    expect(defaultHomeForRoles(["merchant"])).toBe("/merchant");

    expect(postLoginPath(["user"])).toBe("/dashboard");
    expect(postLoginPath(["admin"])).toBe("/admin");
    expect(postLoginPath(["merchant"])).toBe("/merchant");
    expect(postLoginPath(["user"], "/app/profile")).toBe("/app/profile");
    expect(postLoginPath(["admin"], "/admin/users")).toBe("/admin/users");
    // Non-admin cannot force admin next
    expect(postLoginPath(["user"], "/admin")).toBe("/dashboard");
    expect(postLoginPath(["admin"], "https://evil.com")).toBe("/admin");
  });

  it("derives permissions from roles", () => {
    const perms = permissionsForRoles(["super_admin", "user"]);
    expect(perms).toContain("admin.access");
    expect(perms).toContain("admin.full");
    expect(perms).toContain("role:super_admin");
  });
});

describe("mapAuthError", () => {
  it("maps status codes", () => {
    expect(mapAuthError(new ApiError(401, "INVALID_CREDENTIALS", "Invalid email or password"))).toBe(
      "Invalid email or password",
    );
    expect(mapAuthError(new ApiError(403, "ACCOUNT_NOT_VERIFIED", "Please verify your email"))).toContain(
      "verify",
    );
    expect(mapAuthError(new ApiError(423, "ACCOUNT_LOCKED", "Account is locked"))).toBe(
      "Account is locked",
    );
    expect(mapAuthError(new ApiError(500, "INTERNAL", "boom"))).toBe(
      "Unexpected server error",
    );
  });
});
