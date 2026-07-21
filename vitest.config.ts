import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["apps/web/src/**/*.test.{ts,tsx}"],
    exclude: ["node_modules", ".next", ".next"],
    globals: true,
  },
});
