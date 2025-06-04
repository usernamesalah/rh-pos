.PHONY: help run build test clean docker-up docker-down migrate swagger-gen seed

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
run: ## Run the application with hot reload using Air
	docker compose build
	docker compose up

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/ tmp/ coverage.out coverage.html

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

# Database
migrate-up: ## Run database migrations
	goose -dir migrations mysql "$(DB_DSN)" up

migrate-down: ## Rollback database migrations
	goose -dir migrations mysql "$(DB_DSN)" down

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	goose -dir migrations create $(NAME) sql

seed: ## Seed database with initial data
	go run cmd/seed/main.go

# Swagger
swagger-gen: ## Generate Swagger documentation
	swag init -g cmd/main.go -o docs

# Linting
lint: ## Run golangci-lint
	golangci-lint run

# Format
fmt: ## Format Go code
	go fmt ./...

# Setup
setup: deps ## Setup development environment
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development environment setup complete!"

# Local development with MySQL
dev-up: ## Start local development (MySQL + App)
	docker-compose up mysql -d
	@echo "Waiting for MySQL to be ready..."
	@sleep 10
	make run

dev-down: ## Stop local development
	docker-compose down 