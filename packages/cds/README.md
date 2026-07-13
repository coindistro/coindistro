# @coindistro/cds

**Coindistro Design System** — production UI foundation for Landing, Dashboard, Admin, Earn, Wallet, P2P, Academy, Trading, Signals, and future apps.

## Quick start

```bash
cd packages/cds
npm install
npm run storybook   # http://localhost:6006
npm run typecheck
```

### In an app

```tsx
import { CdsThemeProvider, Button, CoinCard, DashboardShell } from "@coindistro/cds";
import "@coindistro/cds/styles";

export default function App({ children }) {
  return (
    <CdsThemeProvider defaultTheme="dark">
      {children}
    </CdsThemeProvider>
  );
}
```

Ensure the app Tailwind `content` globs include:

```js
"../../packages/cds/src/**/*.{ts,tsx}"
```

and extend theme from `@coindistro/cds/tailwind` (or duplicate `cdsTailwindExtend`).

## What’s included

- Design tokens (colors, type, radius, z-index, motion)
- Dark / light / system themes
- shadcn-style primitives (Button, Input, Dialog, Tabs, Select, …)
- Forms (password, search, currency, crypto address)
- Fintech cards (coin, wallet, earn product, portfolio, KYC, genesis/founder)
- Layout shells (auth, dashboard, page header)
- Data table + area/donut charts
- Storybook + accessibility addon

## Docs

See [docs/OVERVIEW.md](./docs/OVERVIEW.md).

## Quality bar

Reusable · Typed · Documented · Accessible · Responsive · Theme-aware · Composable
