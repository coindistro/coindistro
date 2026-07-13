# Coindistro Design System (CDS)

Official UI foundation for every Coindistro application.

## Philosophy

Trust · Innovation · Speed · Security · Simplicity · Professionalism

Dark mode is the **primary** brand experience. Light and system themes are fully supported via `CdsThemeProvider`.

## Package

```
@coindistro/cds
```

## Install (monorepo)

```bash
# from repo root
npm install
# apps consume:
# import { Button, CoinCard } from "@coindistro/cds"
# import "@coindistro/cds/styles"
```

## Structure

| Path | Purpose |
|------|---------|
| `src/tokens` | Design tokens + Tailwind mapping |
| `src/styles` | CSS variables (light/dark) |
| `src/components/ui` | Primitives (shadcn/Radix patterns) |
| `src/components/forms` | Form fields (RHF/Zod friendly) |
| `src/components/crypto` | Fintech / crypto composites |
| `src/components/layout` | Auth / dashboard shells |
| `src/components/navigation` | Sidebar, topbar |
| `src/components/tables` | Data table |
| `src/components/charts` | Recharts wrappers |
| `src/providers` | Theme provider |
| `src/hooks` | Theme, media query |

## Usage guidelines

1. **Never hardcode colors** — use semantic tokens (`primary`, `muted`, `destructive`).
2. **Keep API logic outside** CDS components — pass data via props.
3. Prefer composition (`Card` + `Badge` + `Button`) over one-off page markup.
4. Support keyboard focus and reduced motion (built into tokens).
5. Use Storybook as the component catalog before adding screens.

## Storybook

```bash
cd packages/cds
npm install
npm run storybook
```

## Accessibility

- Focus rings via `:focus-visible`
- ARIA on interactive controls
- `prefers-reduced-motion` respected
- Semantic HTML where possible

## Extending

1. Add tokens in `src/styles/globals.css` + `src/tokens`
2. Build primitives under `components/ui`
3. Compose product components under `crypto` / `dashboard` / `admin`
4. Export from `src/index.ts`
5. Add Storybook stories + docs

## Brand note

CDS uses Coindistro purple (`#7C3AED`) and cyan (`#06B6D4`) with a deep night background. Avoid cloning other exchanges — keep spacing clean and hierarchy clear.
