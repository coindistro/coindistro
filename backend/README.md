# Coindistro Backend

The official Go backend for [Coindistro](https://coindistro.com) — Africa's next-generation crypto financial ecosystem.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.24+ |
| Framework | Gin |
| Database | PostgreSQL 16 (via pgx) |
| Cache | Redis 7 |
| Auth | JWT (access + refresh tokens) |
| Config | Viper |
| Logging | Zap |
| Validation | go-playground/validator |
| Docs | Swagger/OpenAPI |
| Metrics | Prometheus |
| Tracing | OpenTelemetry (OTLP) |
| Queue | In-memory (pluggable: Redis Streams, NATS, Kafka, RabbitMQ) |
| Containerization | Docker |
| Hot Reload | Air |

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Router (Gin)                     │
├─────────────────────────────────────────────────────────────┤
│  Middleware: RequestID → Logger → Recovery → CORS → Gzip    │
│             → RateLimiter → Metrics → Auth → RBAC           │
├─────────────────────────────────────────────────────────────┤
│  Routes: /health, /ready, /live, /metrics, /swagger         │
│          /api/v1/... (versioned)                             │
├─────────────────────────────────────────────────────────────┤
│  Infrastructure Components:                                  │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌────────┐ │
│  │ DB   │ │Redis │ │Auth  │ │RBAC  │ │Event │ │Queue   │ │
│  │(pgx) │ │(go-  │ │(JWT) │ │(Perms│ │Bus   │ │(Pub/   │ │
│  │      │ │redis)│ │      │ │/Roles│ │      │ │Sub)    │ │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └────────┘ │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌────────┐ │
│  │Work- │ │Sched-│ │Feat. │ │Email │ │Stor- │ │Tele-   │ │
│  │ers   │ │uler  │ │Flags │ │(SMTP)│ │age   │ │metry   │ │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
backend/
├── cmd/
│   └── api/                  # Application entry point
├── configs/                  # Configuration files
├── docs/                     # Swagger documentation
├── internal/
│   ├── audit/                # Audit logging subsystem
│   ├── auth/                 # JWT authentication infrastructure
│   ├── cache/                # Redis client and caching interface
│   ├── common/               # Shared utilities (pagination, etc.)
│   ├── config/               # Configuration loading (Viper)
│   ├── database/             # PostgreSQL connection and pooling
│   ├── email/                # Email abstraction (SMTP/Noop)
│   ├── errors/               # Standardized error types
│   ├── events/               # Event-driven architecture (pub/sub)
│   ├── featureflags/         # Feature flag system
│   ├── health/               # Health check endpoints
│   ├── logger/               # Structured logging (Zap)
│   ├── metrics/              # Prometheus metrics collection
│   ├── middleware/           # HTTP middleware stack
│   ├── monitoring/           # Prometheus/Grafana/Loki configs
│   ├── queue/                # Queue abstraction (pub/sub/DLQ)
│   ├── rbac/                 # Role-Based Access Control
│   ├── response/             # Unified API response format
│   ├── routes/               # Route registration and API versioning
│   ├── scheduler/            # Task scheduling framework
│   ├── server/               # HTTP server with DI container
│   ├── storage/              # File storage abstraction
│   ├── telemetry/            # OpenTelemetry distributed tracing
│   ├── uuid/                 # UUIDv7 generation utilities
│   ├── validation/           # Request validation
│   ├── workers/              # Background worker framework
│   └── modules/              # Future module placeholders
├── migrations/               # SQL migration files
├── pkg/                      # Shared packages
├── scripts/                  # Utility scripts
├── Dockerfile                # Multi-stage production build
├── docker-compose.yml        # Local development environment
├── Makefile                  # Build automation
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make (optional)

### 1. Clone and Setup

```bash
cd backend
cp .env.example .env
```

### 2. Start Dependencies (PostgreSQL + Redis)

```bash
docker compose up -d postgres redis
```

### 3. Run Database Migrations

```bash
psql -h localhost -U coindistro -d coindistro -f migrations/001_initial_schema.sql
```

### 4. Start the Server

```bash
# With hot reload (recommended for development)
make dev

# Or without hot reload
make run
```

The server will start at `http://localhost:8080`.

## API Endpoints

### Infrastructure Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Full health check (DB + Redis) |
| GET | `/ready` | Readiness probe |
| GET | `/live` | Liveness probe |
| GET | `/metrics` | Prometheus metrics |
| GET | `/swagger/*any` | Swagger documentation |

### API v1

All business endpoints are versioned under `/api/v1`.

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/health` | No | API health check |
| GET | `/api/v1/features` | No | List feature flags |
| GET | `/api/v1/users/me` | Yes | Get current user profile |
| PUT | `/api/v1/users/me` | Yes | Update current user profile |
| POST | `/api/v1/auth/refresh` | Yes | Refresh access token |
| POST | `/api/v1/auth/logout` | Yes | Logout |
| PUT | `/api/v1/auth/password` | Yes | Change password |

### Module Placeholder Routes

Routes protected by feature flags return 503 when disabled.

| Path | Feature Flag | Description |
|------|-------------|-------------|
| `/api/v1/kyc` | `kyc.enabled` | KYC submissions |
| `/api/v1/merchant` | `merchant.enabled` | Merchant accounts |
| `/api/v1/academy` | `academy.enabled` | Educational courses |
| `/api/v1/signals` | `signals.enabled` | Trading signals |
| `/api/v1/bots` | `bots.enabled` | Trading bots |
| `/api/v1/wallet` | `wallet.enabled` | Wallet management |
| `/api/v1/payments` | `payments.enabled` | Payment processing |
| `/api/v1/notifications` | `notifications.enabled` | Notifications |
| `/api/v1/analytics` | `analytics.enabled` | Analytics |

### Admin Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/admin/dashboard` | Admin | Admin dashboard |
| GET | `/api/v1/admin/users` | Admin | List all users |
| GET | `/api/v1/admin/roles` | Admin | List all roles |
| GET | `/api/v1/admin/roles/:role/permissions` | Admin | Get role permissions |
| GET | `/api/v1/admin/features` | Admin | List feature flags |
| PUT | `/api/v1/admin/features/:flag` | Admin | Toggle feature flag |

## RBAC (Role-Based Access Control)

### Roles

| Role | Description |
|------|-------------|
| `super_admin` | Full system access, all permissions |
| `admin` | Administrative access (most permissions) |
| `compliance_officer` | KYC/merchant review and audit access |
| `support` | Customer support read access |
| `merchant` | Merchant account management |
| `instructor` | Academy course creation and management |
| `trader` | Trading, bots, and wallet management |
| `student` | Academy read access |
| `user` | Default user (profile, wallet, academy, signals) |

### Permissions

Permissions follow the format `resource.action` (e.g., `users.read`, `merchant.approve`).

**User Management:** `users.read`, `users.write`, `users.delete`
**Academy:** `academy.create`, `academy.update`, `academy.delete`, `academy.read`
**Signals:** `signals.publish`, `signals.delete`, `signals.read`
**Merchant:** `merchant.approve`, `merchant.read`, `merchant.write`
**Wallet:** `wallet.read`, `wallet.write`, `wallet.freeze`
**Bots:** `bots.manage`, `bots.read`
**KYC:** `kyc.access`, `kyc.approve`
**Payments:** `payments.read`, `payments.write`, `payments.approve`
**Admin:** `admin.access`, `admin.settings`, `admin.users`, `admin.audit`, `admin.feature_flags`

### Usage

```go
// Check if a role has a specific permission
rbac.HasPermission(rbac.RoleAdmin, rbac.PermUsersRead)  // true
rbac.HasPermission(rbac.RoleUser, rbac.PermMerchantApprove)  // false

// Get all permissions for a role
perms := rbac.GetPermissions(rbac.RoleTrader)

// Middleware usage
r.Use(middleware.RequireRole("admin", "super_admin"))
r.Use(middleware.RequirePermission("merchant.approve"))
```

## Audit Logging

Every security-sensitive action generates an audit event. The audit event store is
pluggable — implement the `audit.Store` interface to store logs in your database
or external system.

### Audited Actions

- Authentication: login, logout, password change, email change, token refresh
- User management: user creation, deletion, role/permission changes
- KYC: submission, approval, rejection
- Merchant: creation, approval, suspension
- Wallet: creation, deposits, withdrawals (request/approval/rejection), freeze/unfreeze
- Trading: signal publication, bot start/stop, trade execution
- Admin: settings changes, feature flag changes

### Usage

```go
// Manual audit logging
event := audit.NewEvent(actorID, audit.ActionLogin).
    WithUserID(userID).
    WithIP(clientIP).
    WithUserAgent(userAgent).
    WithOutcome("success").
    Build()
auditLogger.Record(ctx, event)
```

## Event Bus

The event bus provides a publish-subscribe abstraction for domain events.
The transport is pluggable — future backends include NATS, Kafka, RabbitMQ.

### Event Types

- `user.registered`, `user.verified`, `user.logged_in`
- `kyc.submitted`, `kyc.approved`, `kyc.rejected`
- `merchant.created`, `merchant.approved`
- `wallet.created`, `wallet.deposit_completed`
- `signal.published`
- `payment.completed`, `payment.failed`
- `notification.requested`

### Usage

```go
// Subscribe to events
eventBus.Subscribe(events.EventUserRegistered, events.HandlerFunc(func(ctx context.Context, event events.Event) error {
    // Handle user registration event
    return nil
}))

// Publish events
event := events.NewEvent(events.EventUserRegistered, "auth-service", map[string]interface{}{
    "user_id": "abc123",
    "email":   "user@example.com",
})
eventBus.Publish(ctx, event)
```

## Background Workers

The worker framework supports asynchronous job processing with a pool of workers.

### Job Types

- `email.send`, `email.verification`, `email.password_reset`
- `notification.push`, `notification.sms`
- `signal.broadcast`
- `payment.process`, `payment.settle`
- `blockchain.sync`, `blockchain.confirm`
- `report.generate`, `report.daily`
- `system.cleanup`, `system.health_check`

### Usage

```go
// Register job handler
jobRegistry.Register(workers.JobSendEmail, func(ctx context.Context, job workers.Job) error {
    // Send email logic
    return nil
})

// Submit job
pool.Submit(workers.Job{
    ID:   uuid.NewString(),
    Type: workers.JobSendEmail,
    Payload: map[string]interface{}{
        "to":      "user@example.com",
        "subject": "Welcome!",
    },
})
```

## Scheduler

The scheduler manages recurring and one-time tasks.

### Default Task Intervals

| Task | Interval |
|------|----------|
| `market_sync` | 30s |
| `daily_report` | 24h |
| `leaderboard` | 1h |
| `portfolio_calc` | 5m |
| `certificate_gen` | 10m |
| `cleanup` | 1h |
| `health_check` | 30s |
| `blockchain_sync` | 1m |
| `payment_settlement` | 1m |
| `signal_expiry` | 5m |

### Usage

```go
scheduler.AddTask(scheduler.Task{
    ID:       scheduler.TaskMarketSync,
    Name:     "Market Data Sync",
    Interval: 30 * time.Second,
    Handler: func(ctx context.Context) error {
        // Sync market data
        return nil
    },
})
scheduler.Start()
```

## Feature Flags

Centralized feature flag management. Features can be toggled at runtime
without code changes.

### Default Flags

| Flag | Default | Description |
|------|---------|-------------|
| `exchange.enabled` | true | Enable exchange trading |
| `academy.enabled` | true | Enable academy |
| `signals.enabled` | true | Enable trading signals |
| `merchant.enabled` | true | Enable merchant services |
| `bots.enabled` | true | Enable trading bots |
| `payments.enabled` | true | Enable payments |
| `bank.enabled` | false | Enable banking (future) |
| `maintenance_mode` | false | Maintenance mode |

### Usage

```go
// Check flag
if featureFlags.IsEnabled(featureflags.FlagExchange) {
    // Enable exchange routes
}

// Toggle via API
PUT /api/v1/admin/features/bots.enabled
{"enabled": false}
```

## Monitoring

### Prometheus Metrics

Available at `GET /metrics`. Metrics include:

**HTTP:** request count, duration, request/response size
**Database:** query latency, errors, pool size
**Redis:** command latency, errors, operations
**Workers:** queue depth, job count, duration, errors
**Scheduler:** task count, duration, errors
**Events:** published/handled counts
**Cache:** hit/miss counts
**System:** memory, goroutines, CPU, open FDs
**Business:** active users, transactions, deposits, withdrawals, signals, active bots

### Grafana Dashboard

A pre-configured Grafana dashboard is available at `internal/monitoring/grafana_dashboard.json`.

### Prometheus Configuration

Prometheus config is at `internal/monitoring/prometheus.yml`.

## Distributed Tracing

OpenTelemetry integration with OTLP HTTP exporter. Traces HTTP requests,
database operations, Redis operations, and background workers.

### Usage

```go
// Start a trace
ctx, span := tracer.StartSpan(ctx, "process_payment")
defer span.End()

// Add attributes
telemetry.SetSpanAttributes(ctx, telemetry.AttrUserID.WithString(userID))

// Record error
telemetry.SetSpanError(ctx, err)
```

## Email

Abstraction for sending emails. Supports SMTP and noop (development) providers.
Future providers: Resend, SendGrid, AWS SES.

### Config

```yaml
email:
  provider: "smtp"    # smtp | noop
  smtp:
    host: "smtp.example.com"
    port: 587
    username: "user"
    password: "pass"
    from: "noreply@coindistro.com"
```

## Storage

Abstraction for file storage. Supports local filesystem and in-memory providers.
Future providers: Amazon S3, Cloudflare R2, MinIO.

### Usage

```go
file, err := storage.Upload(ctx, "users/avatars/user123.jpg", imageReader)
reader, err := storage.Download(ctx, "users/avatars/user123.jpg")
exists, err := storage.Exists(ctx, "users/avatars/user123.jpg")
```

## Queue

Message queue abstraction with publish/consume, retry, and dead letter queue support.
Currently in-memory; future backends include Redis Streams, NATS JetStream, RabbitMQ.

### Topics

- `email.send`, `email.verification`, `email.password_reset`
- `signal.broadcast`, `bot.execution`, `trade.settlement`
- `payment.process`, `payment.settlement`, `payment.refund`
- `blockchain.sync`, `blockchain.confirm`
- `system.cleanup`, `system.report`

## UUIDv7

Time-ordered UUIDs with millisecond precision, ideal for database primary keys
and time-based sorting.

### Usage

```go
id := uuid.New()
fmt.Println(id.String())  // 018f3a7e-5b3c-7a00-8000-000000000000

id2 := uuid.NewString()    // string directly
parsed, _ := uuid.Parse("018f3a7e-5b3c-7a00-8000-000000000000")
```

## Configuration

Configuration is loaded from `configs/config.yaml` with environment variable overrides
prefixed with `COINDISTRO_`.

### Key Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `COINDISTRO_ENV` | `development` | Environment |
| `COINDISTRO_PORT` | `8080` | Server port |
| `COINDISTRO_DB_HOST` | `localhost` | PostgreSQL host |
| `COINDISTRO_REDIS_HOST` | `localhost` | Redis host |
| `COINDISTRO_JWT_ACCESS_SECRET` | - | JWT secret |
| `COINDISTRO_LOG_LEVEL` | `info` | Log level |
| `COINDISTRO_TELEMETRY_ENABLED` | `false` | Enable tracing |
| `COINDISTRO_SMTP_HOST` | - | SMTP host |
| `COINDISTRO_METRICS_ENABLED` | `true` | Enable Prometheus |

## Makefile Commands

```bash
make help        # Show available commands
make build       # Build the binary
make run         # Run the server
make dev         # Run with hot reload (Air)
make test        # Run all tests
make lint        # Run linter
make swagger     # Generate Swagger docs
make docker-up   # Start all Docker services
make docker-down # Stop Docker services
make deps        # Download dependencies
make tidy        # Tidy go modules
make all         # Run all checks (deps, lint, test, build)
```

## Docker

### Start Everything

```bash
docker compose up -d --build
```

### Services

| Service | Port | Description |
|---------|------|-------------|
| API | 8080 | Go backend |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache |

## Earn Module

Routes are under `/api/v1/earn` (feature flag: `earn.enabled`).

| Area | Paths |
|------|--------|
| Products | `GET /earn/products`, `GET /earn/products/:id` |
| Portfolio | `GET /earn/portfolio` |
| Participation | `POST /earn/products/:id/join`, `GET /earn/participations`, add-funds / withdraw / exit |
| Rewards / history | `GET /earn/rewards`, `GET /earn/history` |
| Launchpool / Learn | `GET /earn/launchpool`, `GET /earn/learn`, `POST /earn/learn/:id/complete` |
| Referral | `GET /earn/referral/rewards` |
| Admin | `/earn/admin/products`, analytics, participants, launchpool, learn |

Migration: `migrations/003_earn_service.sql`

Reward calculations are **display/estimate only** — no custody or on-chain execution.

## Future Microservice Roadmap

The current codebase is a monolith designed for clean extraction into services:

1. **Auth Service** (Go) — Authentication, authorization, RBAC
2. **User Service** (Go) — User management, profiles, KYC
3. **Exchange Service** (Rust) — Order matching, trading engine
4. **Wallet Service** (Rust) — Wallet management, blockchain sync
5. **Payment Service** (Go) — Payment processing, merchant services
6. **Notification Service** (Go) — Push, email, SMS notifications
7. **Signal Service** (Go/Rust) — Trading signals, market analysis
8. **Academy Service** (Go) — Course management, certifications
9. **Analytics Service** (Go) — Reporting, dashboards, insights

Each service can use the shared packages in `pkg/` and communicate via the event bus.

## License

MIT