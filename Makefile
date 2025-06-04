.PHONY: dev build prod down clean test migrate seed help

# Default target
.DEFAULT_GOAL := help

# Variables
DOCKER_COMPOSE_DEV = docker-compose
DOCKER_COMPOSE_PROD = docker-compose -f docker-compose.prod.yml

# Development commands
dev: ## Start development environment
	$(DOCKER_COMPOSE_DEV) up --build

dev-down: ## Stop development environment
	$(DOCKER_COMPOSE_DEV) down

dev-logs: ## View development logs
	$(DOCKER_COMPOSE_DEV) logs -f

# Production commands
prod: ## Start production environment
	$(DOCKER_COMPOSE_PROD) up -d

prod-down: ## Stop production environment
	$(DOCKER_COMPOSE_PROD) down

prod-logs: ## View production logs
	$(DOCKER_COMPOSE_PROD) logs -f

# Build commands
build: ## Build the application
	go build -o bin/main cmd/main.go

build-prod: ## Build production Docker image
	$(DOCKER_COMPOSE_PROD) build

# Database commands
migrate: ## Run database migrations
	go run cmd/seed/main.go migrate

seed: ## Seed the database with initial data
	go run cmd/seed/main.go seed

# Testing commands
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -cover ./...

# Utility commands
clean: ## Clean up temporary files and Docker resources
	rm -rf tmp/
	rm -rf bin/
	docker system prune -f

lint: ## Run linter
	golangci-lint run

# Help command
help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST) 