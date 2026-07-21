import type { Config } from "tailwindcss";
import animate from "tailwindcss-animate";
import { cdsTailwindExtend } from "../../packages/cds/src/tokens/tailwind";

const config: Config = {
  darkMode: ["class"],
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/cds/src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "1rem",
      screens: { "2xl": "1400px" },
    },
    extend: {
      ...cdsTailwindExtend,
    },
  },
  plugins: [animate],
};

export default config;
