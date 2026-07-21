# Coindistro Frontend Shell (Phase 1)

## Application structure

```
apps/web/src/
  app/
    (public)/          # Marketing & static site
    (auth)/            # Login, register, reset, verify
    (dashboard)/app/*  # Authenticated user portal
    (admin)/admin/*    # Admin control plane
  features/
    authentication/    # Auth API, provider, guards, schemas
    dashboard/         # User chrome + nav
    admin/             # Admin chrome + nav
    public/            # Public helpers
    shared/            # Providers, placeholder page, breadcrumbs, auth screens
  lib/
    api/               # Typed HTTP client
    config.ts
  components/          # Existing landing sections (unchanged design)
```

## Routing

| Area | Prefix | Layout |
|------|--------|--------|
| Public | `/`, `/about`, `/docs`, … | Navbar + Footer |
| Auth | `/login`, `/register`, … | CDS `AuthLayout` |
| User | `/app/*` | Dashboard shell + `RequireAuth` |
| Admin | `/admin/*` | Admin shell + admin roles |

## Authentication flow

1. Tokens stored in `localStorage` (`coindistro_access_token`, `coindistro_refresh_token`).
2. `AuthProvider` bootstraps session via `GET /api/v1/users/me`.
3. `api` client attaches Bearer token; on `401` tries `POST /api/v1/auth/refresh` once.
4. On hard failure, clears session and redirects to `/login?reason=session_expired`.
5. Protected routes use `RequireAuth` (optional `roles` for admin).

Configure API base:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## API client

- Location: `src/lib/api/client.ts`
- Methods: `api.get/post/put/patch/delete`
- Typed errors: `ApiError`
- Interceptors: auth header, refresh retry, unauthorized handler
- Keep page components free of raw `fetch`

## Providers

Root `AppProviders`:

- `CdsThemeProvider` (dark default, system enabled)
- TanStack Query
- Tooltip provider
- Auth provider
- Toast provider (context-based, auto-dismiss)
- Command palette provider (Cmd+K / Ctrl+K)

## Global UI

| Page | Route | Description |
|------|-------|-------------|
| Loading | `loading.tsx` | Full-screen spinner |
| 404 | `not-found.tsx` | Page not found with CDS EmptyState |
| 500 | `error.tsx` | Client error boundary |
| 500 (server) | `global-error.tsx` | Server error boundary |
| Maintenance | `/maintenance` | Scheduled maintenance screen |
| Offline | `/offline` | Offline indicator |

## Shared Components

- `Breadcrumbs` — Auto-generated from pathname segments
- `PlaceholderPage` — Production-quality module placeholder with stats, skeleton, and empty state
- `AuthSuccessScreen` — Success state for auth flows
- `AuthErrorScreen` — Error state for auth flows

## Design System

All new UI uses `@coindistro/cds`. Do **not** re-create buttons, cards, inputs, etc.

Tailwind content includes `packages/cds/src`.

## Adding a module later

1. Create feature folder under `features/<module>/`.
2. Replace placeholder page under `app/(dashboard)/app/<module>/page.tsx`.
3. Add API functions in `features/<module>/api.ts`.
4. Use React Query hooks for server state.
5. Keep navigation entry in `features/dashboard/nav.ts` (already present).

## Developer scripts

```bash
cd apps/web
npm install
npm run dev
npm run typecheck
npm run build
npm run test
```

## Testing

Tests are in `src/__tests__/`:

- `auth-routing.test.ts` — Validates route structure and redirect logic
- `api-client.test.ts` — API client unit tests (token management, error handling, refresh flow)

Run with: `npm run test`

## Phase 1 out of scope

Business logic for P2P, Earn, Wallet, Merchant, Academy, Signals, Trading, AI Bots — placeholders only.