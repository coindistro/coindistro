# Changelog

## v0.4.0-alpha (2026-07-22)

**Coindistro Alpha — Live Authentication & Dashboard**

First official alpha release of Coindistro. This milestone delivers a complete authentication flow, live user and admin dashboards, a production-grade backend architecture, and developer onboarding tooling.

### Added

- Live authentication (register, login, logout, refresh tokens, session persistence)
- Referral-only registration flow
- Genesis Membership system (cap 10,000)
- User dashboard with live widgets (welcome, avatar, roles, badges, referral stats, activity, security)
- Admin dashboard with live overview (user stats, system health, workers, scheduler, feature flags, audit)
- Admin RBAC: super_admin, admin, moderator
- Admin endpoints: `/admin/stats`, `/admin/system`, `/admin/workers`, `/admin/scheduler`, `/admin/users`
- Earn module foundation (products, participation, rewards engine, portfolio, admin APIs)
- Connected pages: Profile (view/update), Referrals, Notifications
- Role-based post-login routing (admin → `/admin`, user → `/app/dashboard`)
- Coindistro Design System (CDS) — tokens, themes, primitives, components, Storybook
- Frontend Shell — route groups, layouts, providers, API client, global UI

### Infrastructure

- Production-ready Go backend architecture (Gin, pgx, Redis, JWT, RBAC)
- PostgreSQL 16 with migrations
- Redis 7 for caching and sessions
- Docker + Docker Compose (production + development profiles)
- Prometheus metrics and OpenTelemetry tracing
- Structured logging (Zap)
- Health / ready / live probes
- Swagger / OpenAPI documentation
- Event bus (in-memory, pluggable)
- Background workers + job registry
- Task scheduler
- Feature flags
- Audit logging

### Developer Experience

- Bootstrap CLI (`go run ./scripts/bootstrap.go`) — creates admin/super_admin users
- Seeder CLI (`go run ./scripts/seed.go`) — demo users, Earn products, referrals, notifications
- Demo super admin: `admin@coindistro.com` / `Admin@123456`
- Demo user: `user1@coindistro.com` / `User@123456`
- Docker Compose-based development environment
- Makefile for common backend tasks
- npm workspaces monorepo

### Frontend

- Next.js 15 with App Router
- React 19
- TypeScript throughout
- TanStack Query for server state
- Tailwind CSS + CDS design tokens
- Protected routes with auto-redirect
- Toast notifications and command palette (Cmd+K)
- Responsive layouts (desktop, tablet, mobile)
- Dark/light theme support
- Global error boundaries (client + server), 404, maintenance, offline pages
- Breadcrumbs and keyboard navigation
- Production-quality placeholder pages for upcoming modules

### Coming Next

- Ledger Service
- Wallet Service
- Portfolio Engine
- P2P Marketplace
- Academy
- Trading Signals
- Payments
- Exchange
- AI Trading Bots

---

## v0.2.0 (2026-06-29)

### Added

- Production-grade backend foundations (Go, Gin, PostgreSQL, Redis)
- JWT access/refresh authentication
- RBAC (roles + permissions)
- Identity Service (registration, login, logout, email verification, password reset)
- Referral engine + dashboard
- Invitation credits system
- Genesis member system (cap 10,000)
- Session and device management
- Activity / security log
- Admin user management
- Account lockout after failed login attempts
- Earn Module (products, participation, rewards, portfolio, admin APIs)
- Audit logging, event bus, background workers, task scheduler
- Feature flags
- Prometheus metrics + OpenTelemetry
- Docker + Docker Compose
- Structured logging (Zap)
- Coindistro Design System (`@coindistro/cds`) — tokens, themes, UI primitives, Storybook
- Marketing landing page (Next.js, React, Tailwind CSS, Framer Motion)
- Swagger / OpenAPI documentation

### Infrastructure

- Modular monolith architecture for future service extraction
- Migration-based schema management
- Configuration via Viper + env overrides
- Email abstraction (SMTP / noop)
- Storage abstraction (local / memory)
- Queue abstraction