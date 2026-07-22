import { describe, it, expect } from "vitest";
import { isAdminRole, postLoginPath } from "@/lib/api/types";

describe("auth helpers", () => {
  it("detects admin roles", () => {
    expect(isAdminRole(["user"])).toBe(false);
    expect(isAdminRole(["admin"])).toBe(true);
    expect(isAdminRole(["super_admin"])).toBe(true);
    expect(isAdminRole(["moderator"])).toBe(true);
    expect(isAdminRole(null)).toBe(false);
  });

  it("routes after login by role", () => {
    expect(postLoginPath(["user"])).toBe("/app/dashboard");
    expect(postLoginPath(["admin"])).toBe("/admin");
    expect(postLoginPath(["user"], "/app/profile")).toBe("/app/profile");
    expect(postLoginPath(["admin"], "/admin/users")).toBe("/admin/users");
    expect(postLoginPath(["admin"], "https://evil.com")).toBe("/admin");
  });
});
