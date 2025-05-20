.PHONY: run test test-run test-e2e test-e2e-with-logs clean build

# Application settings
APP_NAME := product-service
BUILD_DIR := build

# Go settings
GO := go
GO_BUILD := $(GO) build
GO_TEST := $(GO) test
GO_CLEAN := $(GO) clean
GO_GET := $(GO) get
GO_MOD := $(GO) mod

# Docker settings
DOCKER_COMPOSE := docker-compose

# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO_BUILD) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/web

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GO) run ./cmd/web/main.go

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	$(GO_CLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Cleanup completed"

# Run tests
test:
	@echo "Running tests..."
	$(GO_TEST) -v -cover ./internal/...

# Run unit tests only
test-run:
	@echo "Running unit tests..."
	$(GO_TEST) -v -cover ./internal/...

# Run e2e tests with Docker
test-e2e:
	@echo "Running E2E tests with Docker..."
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yml up --build --abort-on-container-exit

# Run e2e tests with logs for debugging
test-e2e-with-logs:
	@echo "Running E2E tests with logs..."
	$(DOCKER_COMPOSE) -f docker-compose.e2e.yml up --build

# Generate mocks for testing
generate-mocks:
	@echo "Generating mocks..."
	mockery --dir=internal/repository --name=ProductRepositoryInterface --output=mocks/repository --filename=product_repository_mock.go
	mockery --dir=internal/usecase --name=ProductUseCaseInterface --output=mocks/usecase --filename=product_usecase_mock.go

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO_MOD) download

# Tidy go.mod file
tidy:
	@echo "Tidying go.mod..."
	$(GO_MOD) tidy

# Run database migrations (uses golang-migrate)
migrate-up:
	@echo "Running database migrations..."
	migrate -path db/migrations -database "mysql://root:password@tcp(localhost:3306)/product_service" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path db/migrations -database "mysql://root:password@tcp(localhost:3306)/product_service" down

# Start Docker containers
docker-up:
	@echo "Starting Docker containers..."
	$(DOCKER_COMPOSE) up -d

# Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE) down

# Help information
help:
	@echo "$(APP_NAME) Makefile help:"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build the application"
	@echo "  run                Run the application"
	@echo "  clean              Clean build artifacts"
	@echo "  test               Run all tests"
	@echo "  test-run           Run unit tests only"
	@echo "  test-e2e           Run E2E tests with Docker"
	@echo "  test-e2e-with-logs Run E2E tests with logs for debugging"
	@echo "  generate-mocks     Generate mocks for testing"
	@echo "  deps               Download dependencies"
	@echo "  tidy               Tidy go.mod file"
	@echo "  migrate-up         Run database migrations"
	@echo "  migrate-down       Rollback database migrations"
	@echo "  docker-up          Start Docker containers"
	@echo "  docker-down        Stop Docker containers"
	@echo "  help               Show this help information"