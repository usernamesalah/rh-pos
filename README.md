# RH POS API

Backend API for a multi-tenant point of sale system built with Go, Echo, GORM, and MySQL.

## Overview

This service provides:

- JWT-protected `/api/*` endpoints for day-to-day POS operations
- Basic Auth protected `/admin/*` endpoints for tenant bootstrap and admin setup
- product, transaction, discount campaign, warranty, report, tenant, and user management
- Swagger/OpenAPI docs at `/swagger/index.html`
- local filesystem or MinIO-backed file storage

## Architecture

The codebase follows clean architecture:

```text
handler -> usecase -> repository -> database
```

Main directories:

```text
cmd/                 application entrypoint
internal/domain/     entities and interfaces
internal/usecase/    business logic
internal/repository/ data access with GORM
internal/handler/    HTTP handlers
internal/server/     router and middleware setup
internal/pkg/        shared packages
migrations/          SQL migrations
docs/                Swagger and project docs
```

## Requirements

- Go 1.24+
- MySQL 8+
- Docker and Docker Compose for containerized development
- Make optional but convenient

## Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

Required values before startup:

- `JWT_SECRET` must be set to a secure non-default value
- `HASHID_SALT` is required and must be at least 16 characters
- `ADMIN_USERNAME` and `ADMIN_PASSWORD` are required
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` must point to a working MySQL database
- if `STORAGE_TYPE=minio`, `MINIO_ACCESS_KEY` and `MINIO_SECRET_KEY` are required

Useful optional values:

- `SERVER_HOST` default `0.0.0.0`
- `SERVER_PORT` default `8080`
- `CORS_ALLOWED_ORIGINS` comma-separated list
- `LOG_LEVEL` default `info`
- `LOCAL_STORAGE_PATH` when `STORAGE_TYPE=local`
- `STORAGE_BASE_URL` for local file URLs

Storage options:

- `STORAGE_TYPE=local` for simple local development
- `STORAGE_TYPE=minio` for object storage

## Local Development

Build the binary:

```bash
go build -o bin/rh-pos cmd/main.go
```

Run migrations:

```bash
./bin/rh-pos migrate up
```

Start the server:

```bash
./bin/rh-pos
```

On Windows, use `bin/rh-pos.exe` instead.

## Docker Development

The repository includes a Docker Compose setup with hot reload via Air:

```bash
make dev
```

Useful commands:

```bash
make dev-down
make dev-logs
make prod
make prod-down
make prod-logs
```

Note: `docker-compose.yml` expects the external Docker network `usernamesalah` to exist.

Create it if needed:

```bash
docker network create usernamesalah
```

## Build, Test, Lint

Build:

```bash
make build
```

Test:

```bash
go test -v ./...
```

Lint:

```bash
golangci-lint run
```

## Database Commands

The compiled binary exposes migration and seed commands:

```bash
./bin/rh-pos migrate up
./bin/rh-pos migrate down
./bin/rh-pos migrate status
./bin/rh-pos seed
```

Or via `make`:

```bash
make migrate-up
make migrate-down
make migrate-status
make seed
```

## API Access

Public endpoints:

- `GET /health`
- `GET /swagger/index.html`
- `GET /warranty/search`
- `GET /warranty/:transaction_id`

Authentication:

- `POST /auth/login` returns a JWT for `/api/*`
- `/admin/*` uses HTTP Basic Auth with `ADMIN_USERNAME` and `ADMIN_PASSWORD`
- `/api/*` uses JWT Bearer auth

Main API groups:

- `/api/products`
- `/api/transactions`
- `/api/reports`
- `/api/discount-campaigns`
- `/api/users`
- `/api/my-tenant`

## Multi-Tenancy Notes

- tenant context is derived from JWT claims on `/api/*`
- IDs exposed by the API are hashed IDs, not raw database IDs
- report and product/campaign access is tenant-scoped

## Troubleshooting

- If startup fails, verify `.env` values first
- If migrations fail, ensure the binary was built before running migration commands
- If `make dev` fails, check that the external Docker network `usernamesalah` exists
- If `STORAGE_TYPE=minio`, verify MinIO credentials and endpoint configuration
- If JWT auth fails, make sure the frontend is sending `Authorization: Bearer <token>`

## License

This project is licensed under the MIT License. See `LICENSE` if present in your deployment context.
