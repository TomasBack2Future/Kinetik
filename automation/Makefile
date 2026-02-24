.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down

# Variables
APP_NAME=webhook-server
DOCKER_IMAGE=kinetik-automation
GO_FILES=$(shell find . -name '*.go' -type f)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) ./cmd/webhook-server

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/webhook-server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.out

lint: ## Run linter
	@echo "Running linter..."
	@go vet ./...
	@gofmt -s -w .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):latest .

docker-up: ## Start services with docker-compose
	@echo "Starting services..."
	@docker-compose up -d

docker-down: ## Stop services
	@echo "Stopping services..."
	@docker-compose down

docker-logs: ## View docker logs
	@docker-compose logs -f webhook-server

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	@psql $(DATABASE_URL) -f migrations/000000_grant_permissions.up.sql
	@psql $(DATABASE_URL) -f migrations/000001_init_schema.up.sql
	@psql $(DATABASE_URL) -f migrations/000002_add_token_tracking.up.sql

migrate-down: ## Run database migrations down (rollback latest)
	@echo "Rolling back latest migration..."
	@psql $(DATABASE_URL) -f migrations/000002_add_token_tracking.down.sql

migrate-down-all: ## Run all database migrations down
	@echo "Rolling back all migrations..."
	@psql $(DATABASE_URL) -f migrations/000002_add_token_tracking.down.sql
	@psql $(DATABASE_URL) -f migrations/000001_init_schema.down.sql

dev: docker-up ## Start development environment
	@echo "Development environment started"
	@echo "Webhook server: http://localhost:8080"
	@echo "Health check: http://localhost:8080/health"

.DEFAULT_GOAL := help
