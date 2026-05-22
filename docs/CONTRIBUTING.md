# Contributing Guide

<!-- AUTO-GENERATED: commands table generated from Makefile -->

## Prerequisites

- Go 1.24+
- MySQL 8+
- Docker and Docker Compose
- `golangci-lint` for linting
- `make` (optional but convenient)

## Setup

```bash
cp .env.example .env
# Edit .env — set JWT_SECRET, HASHID_SALT, ADMIN_USERNAME, ADMIN_PASSWORD, DB_*
docker network create usernamesalah
make dev
```

## Available Commands

<!-- AUTO-GENERATED -->
| Command | Description |
|---------|-------------|
| `make dev` | Start development environment (Docker + hot reload via Air) |
| `make dev-down` | Stop development environment |
| `make dev-logs` | View development logs |
| `make prod` | Start production environment (detached) |
| `make prod-down` | Stop production environment |
| `make prod-logs` | View production logs |
| `make build` | Build binary to `bin/rh-pos` |
| `make build-prod` | Build production Docker image |
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-status` | Show migration status |
| `make seed` | Seed database with initial data |
| `make test` | Run tests (`go test -v ./...`) |
| `make test-coverage` | Run tests with coverage |
| `make lint` | Run `golangci-lint run` |
| `make clean` | Remove `tmp/`, `bin/`, prune Docker |
<!-- /AUTO-GENERATED -->

## Testing

```bash
make test
# or
go test -v ./...
go test -v -cover ./...
```

Write tests in `_test.go` files in the same package. See `internal/usecase/transaction_service_test.go` for examples.

## Code Style

- Run `golangci-lint run` before submitting
- Follow clean architecture: handler → usecase → repository. No layer should import a layer above it.
- All inter-layer communication goes through interfaces in `internal/domain/interfaces/`
- Use structured logging (`slog`) — no `fmt.Printf` in production paths
- Context keys via `internal/pkg/ctxkey` — no raw string context keys

## PR Checklist

- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `make test` passes
- [ ] New behavior covered by tests
- [ ] Migration added if schema changed
- [ ] `.env.example` updated if new env var added
