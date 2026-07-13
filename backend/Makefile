# ─── Coindistro Backend Makefile ───────────────────────

.PHONY: help build run test lint clean dev docker-up docker-down swagger

APP_NAME = coindistro-api
BUILD_DIR = ./bin
MAIN_PATH = ./cmd/api

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/api $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/api"

run: ## Run the server
	@echo "Starting $(APP_NAME)..."
	@go run $(MAIN_PATH) --config ./configs/config.yaml

dev: ## Run with hot reload using Air
	@echo "Starting $(APP_NAME) with hot reload..."
	@which air > /dev/null 2>&1 || (echo "Installing Air..." && go install github.com/air-verse/air@latest)
	@air

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-short: ## Run tests without race detector
	@echo "Running tests (short)..."
	@go test -v -short ./...

test-cover: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-up: ## Start Docker services
	@echo "Starting Docker services..."
	@docker compose up -d --build

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	@docker compose down

docker-logs: ## View Docker logs
	@docker compose logs -f

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null 2>&1 || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@swag init -g cmd/api/main.go -o docs

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy

all: deps lint test build ## Run all checks (deps, lint, test, build)