# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o bin/rh-pos cmd/main.go
make build

# Run tests
go test -v ./...
go test -v -cover ./...
make test

# Lint
golangci-lint run
make lint

# Development (Docker Compose with hot reload via Air)
make dev           # docker-compose up --build
make dev-down

# Production
make prod          # docker-compose.prod.yml up -d
make prod-down
```

Migrations are run manually via the compiled binary: `./bin/rh-pos migrate up`.

## Architecture

Clean Architecture with four layers:

1. **`internal/domain/`** — Entities (`entities/`) and repository/service interfaces (`interfaces/`). No external dependencies. This is the source of truth for data shapes.
2. **`internal/repository/`** — GORM MySQL implementations of repository interfaces. Each repo takes `*gorm.DB` and `*slog.Logger`.
3. **`internal/usecase/`** — Business logic services implementing service interfaces from domain. Injected with repository interfaces and the MinIO storage client.
4. **`internal/handler/`** — Echo HTTP handlers. Injected with use case interfaces. Parse/validate requests, call use cases, return JSON.

Dependency direction: handler → usecase → repository → database. All inter-layer communication goes through interfaces defined in `internal/domain/interfaces/`.

## Key Packages

- **`internal/config/`** — Loads config from env vars (via `godotenv`). All required fields (`JWT_SECRET`, `ADMIN_USERNAME`, `ADMIN_PASSWORD`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`) cause startup failure if missing or using defaults.
- **`internal/pkg/hash/`** — `HashID`/`DecodeHashID` using `go-hashids`. Tenant IDs are hashed in JWT claims and in MinIO object paths. The salt is hardcoded (`__next_move_to_config__`) and should eventually be moved to config.
- **`internal/pkg/storage/minio/`** — MinIO client. All object keys are automatically namespaced under `<hashed_tenant_id>/<key>` using the `tenant_id` value from Go context.
- **`internal/pkg/middleware/`** — `AdminAuth` middleware uses HTTP Basic Auth checked against `ADMIN_USERNAME`/`ADMIN_PASSWORD` from config; sets `tenant_id=0` in Echo context for super-admin operations.

## Auth & Multi-tenancy

- **JWT** is used for `/api/*` routes. The JWT `SuccessHandler` in `internal/server/router.go` decodes the hashed `tenant_id` claim back to `uint` and stores it in both the Echo context (`c.Set`) and the Go request context (`context.WithValue`). This dual storage is intentional: use `c.Get("tenant_id")` in handlers, `ctx.Value("tenant_id")` in MinIO operations.
- **Admin routes** (`/admin/*`) use HTTP Basic Auth. Admin user has `tenant_id=0`.
- Product images stored in MinIO are scoped per tenant automatically.

## Configuration

Copy `.env.example` to `.env`. Required env vars beyond defaults:
- `JWT_SECRET` (must not be the default string)
- `ADMIN_USERNAME` / `ADMIN_PASSWORD`
- `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY`
- Database credentials

## API Routes Summary

| Group | Auth | Notable endpoints |
|-------|------|-------------------|
| `/auth` | None | `POST /login` |
| `/admin` | Basic Auth | CRUD tenants, create users |
| `/api` | JWT | Profile, products (with image upload), transactions, reports |

Product image endpoints: `POST /:id/image` (direct upload), `POST /:id/upload-url` (presigned PUT URL), `GET /:id/image/bytes` (direct download).
