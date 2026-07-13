# Coindistro

**One Platform. Everything Crypto.**

Coindistro is Africa's next-generation crypto financial ecosystem — a unified platform for trading, payments, education, automation, and digital asset management.

| | |
|---|---|
| **Current version** | `v0.2.0` |
| **Status** | Backend foundation + Identity Service complete |
| **License** | MIT |

---

## Current Project Status

| Layer | Status | Notes |
|-------|--------|--------|
| Marketing website | ✅ Complete | Next.js landing page (`apps/web`) |
| Backend infrastructure | ✅ Complete | Go monolith with production-grade foundations |
| Identity Service | ✅ Complete | Auth, sessions, devices, referrals, invitations, Genesis |
| Earn Module | ✅ Complete | Products, participation, rewards engine, portfolio, admin APIs |
| User Service | 🔜 Next (`v0.3.0`) | Profiles, KYC, preferences |
| Wallet / Payments / Exchange | 📋 Planned | Schema prepared; modules not yet implemented |
| Academy / Signals / Bots | 📋 Planned | Feature-flag placeholders + SQL schema |

**Milestone `v0.2.0`** ships a production-ready backend foundation and a full Identity Service, ready for the next product modules.

---

## Completed Modules

### Backend foundation

- HTTP server (Gin) with graceful shutdown
- PostgreSQL (pgx) + Redis
- JWT access / refresh authentication
- RBAC (roles + permissions)
- Audit logging
- Event bus (in-memory, pluggable)
- Background workers + job registry
- Task scheduler
- Feature flags
- Email abstraction (SMTP / noop)
- Storage abstraction (local / memory)
- Queue abstraction
- Prometheus metrics
- OpenTelemetry tracing hooks
- Health / ready / live probes
- Swagger / OpenAPI scaffolding
- Docker + Docker Compose
- Structured logging (Zap)
- Configuration (Viper + env overrides)

### Identity Service

- Invite-only / referral-gated registration
- Login, logout, token refresh
- Email verification + password reset / change
- Session management (list, terminate, terminate all)
- Trusted device management
- Referral engine + dashboard
- Invitation credits + send/list invitations
- Genesis member system (cap 10,000)
- Founder account / badge support
- Activity / security log
- Admin user management (status, roles, credits)
- Lockout after failed login attempts

### Earn Module (`/api/v1/earn`)

- Product catalog (flexible, fixed, stablecoin, AI smart, signal vault, launchpool, learn & earn, referral)
- User portfolio overview and participation lifecycle (join, add funds, withdraw, exit)
- Pluggable reward engine (no custody / real yield execution)
- Launchpool, Learn & Earn, referral reward APIs (contracts for Academy / Identity)
- Admin product management, participants, analytics
- Feature flags, RBAC (`earn.*`), events, workers, scheduler tasks, Prometheus metrics

### Design System (`@coindistro/cds`)

- Official Coindistro Design System package under `packages/cds`
- Tokens, themes (dark/light/system), shadcn/Radix primitives
- Forms, layout shells, tables, charts, crypto/fintech components
- Storybook catalog: `npm run cds:storybook`

### Frontend

- Coindistro marketing landing page (ecosystem, markets, signals, bots, academy, security, roadmap, CTA)
- Dark / light theme
- Responsive layout + animations

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     apps/web (Next.js)                       │
│                   Marketing / public site                    │
└────────────────────────────┬────────────────────────────────┘
                             │ future API client
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                 backend (Go monolith / Gin)                  │
│  Middleware: RequestID → Logger → Recovery → CORS → Gzip    │
│             → RateLimiter → Metrics → Auth → RBAC           │
├─────────────────────────────────────────────────────────────┤
│  /health  /ready  /live  /metrics  /swagger                 │
│  /api/v1/...  (versioned business APIs)                     │
├─────────────────────────────────────────────────────────────┤
│  Identity │ Auth │ RBAC │ Audit │ Events │ Workers │ Flags │
│  (live)   │ JWT  │      │       │ Bus    │ Sched.  │       │
├─────────────────────────────────────────────────────────────┤
│           PostgreSQL 16          │         Redis 7           │
└─────────────────────────────────────────────────────────────┘
```

The backend is intentionally structured as a **modular monolith** so domains (wallet, payments, exchange, academy, signals, bots) can be extracted into services later without rewriting core infrastructure.

```
Coindistro/
├── apps/
│   └── web/                 # Next.js marketing site
├── backend/
│   ├── cmd/api/             # API entrypoint
│   ├── configs/             # Default config
│   ├── internal/            # App code (auth, identity, middleware, …)
│   ├── migrations/          # SQL migrations
│   ├── Dockerfile
│   └── README.md            # Backend-specific docs
├── packages/                # Shared packages (reserved)
├── docker-compose.yml       # Production-oriented compose
├── docker-compose.dev.yml   # Dev compose (hot reload web)
└── package.json             # npm workspaces root
```

---

## Tech Stack

| Area | Technology |
|------|------------|
| Web | Next.js 14, React 18, Tailwind CSS, Framer Motion |
| API | Go 1.24+, Gin |
| Database | PostgreSQL 16 (pgx) |
| Cache | Redis 7 |
| Auth | JWT (access + refresh) |
| Observability | Prometheus, OpenTelemetry, Zap |
| Containers | Docker, Docker Compose |

---

## Quick Start

### Prerequisites

- Node.js 20+
- Go 1.24+
- Docker & Docker Compose (recommended for Postgres + Redis)

### Web (marketing site)

```bash
npm install
npm run dev
# → http://localhost:3000
```

### Backend API

```bash
cd backend
cp .env.example .env
# Start Postgres + Redis (from backend/ or repo root compose)
docker compose up -d postgres redis
# Apply migrations
psql -h localhost -U coindistro -d coindistro -f migrations/001_initial_schema.sql
psql -h localhost -U coindistro -d coindistro -f migrations/002_identity_service.sql
# Run API
make run
# or with hot reload
make dev
# → http://localhost:8080
```

### Full stack via Docker (web + data services)

```bash
# Development (web hot reload)
npm run docker:dev

# Production-style web + Postgres + Redis
npm run docker:up
```

> Backend API Docker service lives under `backend/docker-compose.yml`. Wire it into the root compose when deploying the full stack.

---

## API Snapshot (Identity)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | No | Register (referral code required) |
| POST | `/api/v1/auth/login` | No | Login |
| POST | `/api/v1/auth/refresh` | Yes | Refresh tokens |
| POST | `/api/v1/auth/logout` | Yes | Logout |
| GET | `/api/v1/auth/verify-email` | No | Verify email |
| POST | `/api/v1/auth/forgot-password` | No | Request password reset |
| POST | `/api/v1/auth/reset-password` | No | Complete password reset |
| PUT | `/api/v1/auth/change-password` | Yes | Change password |
| GET | `/api/v1/users/me` | Yes | Current profile |
| PUT | `/api/v1/users/me` | Yes | Update profile |
| GET/DELETE | `/api/v1/sessions…` | Yes | Session management |
| GET/DELETE | `/api/v1/devices…` | Yes | Device management |
| GET | `/api/v1/referrals…` | Yes | Referral dashboard |
| GET/POST | `/api/v1/invitations…` | Yes | Invitations |
| GET | `/health` | No | Health check |

See `backend/README.md` for full infrastructure and admin routes.

---

## Roadmap

| Version | Focus | Status |
|---------|--------|--------|
| **v0.1.x** | Marketing landing page | ✅ Done |
| **v0.2.0** | Backend foundation + Identity Service | ✅ Done |
| **v0.2.1+** | Earn Module (product architecture + APIs) | ✅ Current |
| **v0.3.0** | User Service (profiles, KYC, preferences) | 🔜 Next |
| **v0.4.x** | Wallet + Payments foundations | Planned |
| **v0.5.x** | Merchant / Pay APIs | Planned |
| Later | Exchange, Signals, Bots, Academy product APIs | Planned |
| Scale | Extract microservices (Exchange/Wallet in Rust, etc.) | Future |

### Upcoming modules

1. **User Service** — extended profiles, KYC workflows, preferences  
2. **Wallet Service** — balances, deposits, withdrawals, ledger  
3. **Payment Service** — merchant rails, invoices, settlements  
4. **Exchange** — spot / futures / P2P (high performance path)  
5. **Signals & Bots** — market signals and automation  
6. **Academy** — courses, progress, certifications  
7. **Notifications** — email, push, SMS  

---

## Contribution Guide

1. **Fork** the repository and create a feature branch from `main`.
2. Keep changes focused — one feature or fix per PR.
3. Backend:
   - `cd backend && go fmt ./... && go vet ./... && go test ./...`
   - Prefer Conventional Commits (`feat:`, `fix:`, `docs:`, …).
4. Frontend:
   - `npm run lint` and `npm run build` from the repo root (or `apps/web`).
5. Do **not** commit secrets, `.env` files, binaries, or `node_modules`.
6. Open a PR against `main` with a clear description and test notes.
7. For larger features, open an issue first and align on design.

### Development standards

- No production secrets in git (use env vars / secret managers).
- API changes should remain versioned under `/api/v1`.
- New domain modules should follow the existing `internal/<domain>` layout (handlers / service / store / models).
- Feature-flag new product surfaces until ready for general availability.

---

## Environment

Copy examples — never commit real credentials:

- Root: `.env.example`
- Backend: `backend/.env.example`

Default local DB/Redis credentials in compose and `configs/config.yaml` are **development placeholders** (`coindistro` / `change-me-in-production`). Override them for any non-local environment.

---

## Links

- Backend deep dive: [`backend/README.md`](./backend/README.md)
- Site: [coindistro.com](https://coindistro.com) (when deployed)

---

## License

MIT
