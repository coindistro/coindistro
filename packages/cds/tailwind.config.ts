import type { Config } from "tailwindcss";
import animate from "tailwindcss-animate";
import { cdsTailwindExtend } from "./src/tokens/tailwind";

const config: Config = {
  darkMode: ["class"],
  content: ["./src/**/*.{ts,tsx}", "./.storybook/**/*.{ts,tsx}"],
  theme: {
    container: {
      center: true,
      padding: "1rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      ...cdsTailwindExtend,
    },
  },
  plugins: [animate],
};

export default config;
