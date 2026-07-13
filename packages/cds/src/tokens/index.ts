/**
 * Coindistro Design System — design tokens (source of truth).
 * Runtime theming uses CSS variables defined in styles/globals.css.
 * This module documents token names for TypeScript consumers.
 */

export const cdsColors = {
  primary: "hsl(var(--cds-primary))",
  secondary: "hsl(var(--cds-secondary))",
  success: "hsl(var(--cds-success))",
  warning: "hsl(var(--cds-warning))",
  danger: "hsl(var(--cds-danger))",
  info: "hsl(var(--cds-info))",
  background: "hsl(var(--cds-background))",
  foreground: "hsl(var(--cds-foreground))",
  surface: "hsl(var(--cds-surface))",
  muted: "hsl(var(--cds-muted))",
  border: "hsl(var(--cds-border))",
  input: "hsl(var(--cds-input))",
  card: "hsl(var(--cds-card))",
  accent: "hsl(var(--cds-accent))",
  sidebar: "hsl(var(--cds-sidebar))",
} as const;

export const cdsSpacing = {
  0: "0",
  1: "0.25rem",
  2: "0.5rem",
  3: "0.75rem",
  4: "1rem",
  5: "1.25rem",
  6: "1.5rem",
  8: "2rem",
  10: "2.5rem",
  12: "3rem",
  16: "4rem",
  20: "5rem",
  24: "6rem",
} as const;

export const cdsRadius = {
  none: "0",
  sm: "calc(var(--cds-radius) - 4px)",
  md: "calc(var(--cds-radius) - 2px)",
  lg: "var(--cds-radius)",
  xl: "calc(var(--cds-radius) + 4px)",
  "2xl": "calc(var(--cds-radius) + 8px)",
  full: "9999px",
} as const;

export const cdsZIndex = {
  base: 0,
  dropdown: 50,
  sticky: 100,
  overlay: 200,
  modal: 300,
  popover: 400,
  toast: 500,
  tooltip: 600,
  command: 700,
} as const;

export const cdsBreakpoints = {
  sm: "640px",
  md: "768px",
  lg: "1024px",
  xl: "1280px",
  "2xl": "1400px",
  "3xl": "1920px",
} as const;

export const cdsTransitions = {
  fast: "150ms",
  base: "200ms",
  slow: "300ms",
  slower: "500ms",
} as const;

export const cdsTypography = {
  display: "text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight",
  h1: "text-3xl md:text-4xl font-bold tracking-tight",
  h2: "text-2xl md:text-3xl font-semibold tracking-tight",
  h3: "text-xl md:text-2xl font-semibold",
  h4: "text-lg font-semibold",
  h5: "text-base font-semibold",
  h6: "text-sm font-semibold",
  bodyLarge: "text-lg leading-relaxed",
  body: "text-base leading-normal",
  small: "text-sm leading-normal",
  caption: "text-xs leading-normal text-muted-foreground",
  overline: "text-xs font-medium uppercase tracking-wider text-muted-foreground",
  code: "font-mono text-sm",
} as const;

export const cdsCryptoAssets = {
  BTC: "#F7931A",
  ETH: "#627EEA",
  USDT: "#26A17B",
  USDC: "#2775CA",
  SOL: "#9945FF",
  BNB: "#F3BA2F",
  XRP: "#23292F",
  CDT: "#7C3AED",
} as const;

export type CdsColorToken = keyof typeof cdsColors;
export type CdsTypographyVariant = keyof typeof cdsTypography;
