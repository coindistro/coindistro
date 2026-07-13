/**
 * Tailwind theme extensions derived from CDS tokens.
 */
export const cdsTailwindExtend = {
  colors: {
    border: "hsl(var(--cds-border))",
    input: "hsl(var(--cds-input))",
    ring: "hsl(var(--cds-ring))",
    background: "hsl(var(--cds-background))",
    foreground: "hsl(var(--cds-foreground))",
    primary: {
      DEFAULT: "hsl(var(--cds-primary))",
      foreground: "hsl(var(--cds-primary-foreground))",
    },
    secondary: {
      DEFAULT: "hsl(var(--cds-secondary))",
      foreground: "hsl(var(--cds-secondary-foreground))",
    },
    destructive: {
      DEFAULT: "hsl(var(--cds-danger))",
      foreground: "hsl(var(--cds-danger-foreground))",
    },
    success: {
      DEFAULT: "hsl(var(--cds-success))",
      foreground: "hsl(var(--cds-success-foreground))",
    },
    warning: {
      DEFAULT: "hsl(var(--cds-warning))",
      foreground: "hsl(var(--cds-warning-foreground))",
    },
    info: {
      DEFAULT: "hsl(var(--cds-info))",
      foreground: "hsl(var(--cds-info-foreground))",
    },
    muted: {
      DEFAULT: "hsl(var(--cds-muted))",
      foreground: "hsl(var(--cds-muted-foreground))",
    },
    accent: {
      DEFAULT: "hsl(var(--cds-accent))",
      foreground: "hsl(var(--cds-accent-foreground))",
    },
    popover: {
      DEFAULT: "hsl(var(--cds-popover))",
      foreground: "hsl(var(--cds-popover-foreground))",
    },
    card: {
      DEFAULT: "hsl(var(--cds-card))",
      foreground: "hsl(var(--cds-card-foreground))",
    },
    surface: "hsl(var(--cds-surface))",
    sidebar: {
      DEFAULT: "hsl(var(--cds-sidebar))",
      foreground: "hsl(var(--cds-sidebar-foreground))",
      border: "hsl(var(--cds-sidebar-border))",
      accent: "hsl(var(--cds-sidebar-accent))",
    },
    chart: {
      "1": "hsl(var(--cds-chart-1))",
      "2": "hsl(var(--cds-chart-2))",
      "3": "hsl(var(--cds-chart-3))",
      "4": "hsl(var(--cds-chart-4))",
      "5": "hsl(var(--cds-chart-5))",
    },
  },
  borderRadius: {
    lg: "var(--cds-radius)",
    md: "calc(var(--cds-radius) - 2px)",
    sm: "calc(var(--cds-radius) - 4px)",
  },
  fontFamily: {
    sans: ["var(--cds-font-sans)", "Inter", "system-ui", "sans-serif"],
    mono: ["var(--cds-font-mono)", "ui-monospace", "monospace"],
  },
  boxShadow: {
    "cds-sm": "var(--cds-shadow-sm)",
    "cds-md": "var(--cds-shadow-md)",
    "cds-lg": "var(--cds-shadow-lg)",
    "cds-glow": "var(--cds-shadow-glow)",
  },
  keyframes: {
    "cds-accordion-down": {
      from: { height: "0" },
      to: { height: "var(--radix-accordion-content-height)" },
    },
    "cds-accordion-up": {
      from: { height: "var(--radix-accordion-content-height)" },
      to: { height: "0" },
    },
    "cds-fade-in": {
      from: { opacity: "0" },
      to: { opacity: "1" },
    },
    "cds-slide-up": {
      from: { opacity: "0", transform: "translateY(8px)" },
      to: { opacity: "1", transform: "translateY(0)" },
    },
    "cds-pulse-soft": {
      "0%, 100%": { opacity: "1" },
      "50%": { opacity: "0.6" },
    },
  },
  animation: {
    "cds-accordion-down": "cds-accordion-down 0.2s ease-out",
    "cds-accordion-up": "cds-accordion-up 0.2s ease-out",
    "cds-fade-in": "cds-fade-in 0.2s ease-out",
    "cds-slide-up": "cds-slide-up 0.25s ease-out",
    "cds-pulse-soft": "cds-pulse-soft 2s ease-in-out infinite",
  },
  zIndex: {
    dropdown: "50",
    sticky: "100",
    overlay: "200",
    modal: "300",
    popover: "400",
    toast: "500",
    tooltip: "600",
    command: "700",
  },
  transitionDuration: {
    cds: "200ms",
    "cds-fast": "150ms",
    "cds-slow": "300ms",
  },
};
