import { defineConfig } from "vitest/config";
import path from "path";

export default defineConfig({
  test: {
    include: ["apps/web/src/**/*.test.{ts,tsx}"],
    exclude: ["node_modules", ".next", ".next"],
    globals: true,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "apps/web/src"),
    },
  },
});