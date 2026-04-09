# AGENTS.md

## Commands

```bash
# Build
go build -o bin/rh-pos cmd/main.go

# Test
go test -v ./...

# Lint
golangci-lint run

# Dev (Docker Compose with hot reload via Air)
make dev

# Migrations (run after build)
./bin/rh-pos migrate up
```

## Required Setup

1. Copy `.env.example` to `.env`
2. Set `JWT_SECRET` (must not be the default string)
3. Set `ADMIN_USERNAME` / `ADMIN_PASSWORD`
4. Set `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY`
5. Run `./bin/rh-pos migrate up` to apply database migrations

## Architecture

- **Clean Architecture**: handler → usecase → repository → database
- **Layers**: `internal/domain/` (entities + interfaces), `internal/repository/` (GORM), `internal/usecase/`, `internal/handler/`
- **Dependency injection**: All inter-layer communication via interfaces in `internal/domain/interfaces/`

## Multi-tenancy

- JWT for `/api/*` routes; tenant_id stored in Echo context (`c.Get("tenant_id")`) and Go context (`ctx.Value("tenant_id")`)
- Basic Auth for `/admin/*` routes; admin has `tenant_id=0`
- Tenant IDs are hashed for JWT claims and MinIO object paths (`internal/pkg/hash/`)

## Important Notes

- MinIO uses external Docker network `usernamesalah`; object keys automatically namespaced under `<hashed_tenant_id>/<key>`
- Migrations are manual (run via compiled binary), not automatic at startup
- Go follows cursor rules in `.cursor/rules/go.mdc`: no comments, be concise, no placeholders
- Use `context.Context` as first argument, `*slog.Logger` as second