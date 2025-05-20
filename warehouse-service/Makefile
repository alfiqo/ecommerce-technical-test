# Build variables
MIGRATE_DIR=db/migrations
MIGRATE_EXT=sql
APP_NAME=warehouse-service
BUILD_DIR=./build
GO_FILES=$(shell find . -name '*.go' | grep -v /vendor/)
DOCKER_IMAGE_NAME=warehouse-service
DOCKER_TAG=latest

# Go build flags
GOOS?=darwin
GOARCH?=amd64
BUILD_FLAGS=-ldflags="-s -w"

.PHONY: all build clean run test mock swagger docker-build docker-run migrate migrate-up migrate-down migrate-create help

all: clean build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/web/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/web/main.go

# Run with custom port
run-custom-port:
	@echo "Running $(APP_NAME) on custom port..."
	@PORT=3001 go run ./cmd/web/main.go

# Run tests
test:
	@echo "Running all tests..."
	@go test -v -cover ./internal/...

# Run e2e tests
test-e2e:
	@echo "Running end-to-end tests..."
	@./scripts/run-e2e-tests.sh

# Run e2e tests with logs
test-e2e-with-logs:
	@echo "Running end-to-end tests with logs..."
	@./scripts/run-e2e-tests.sh --show-logs

# Generate mocks
mock:
	@echo "Generating mocks..."
	# Generate mocks for the usecase interfaces
	@go run go.uber.org/mock/mockgen -source=./internal/usecase/warehouse_usecase.go -destination=./mocks/usecase/warehouse_usecase_mock.go -package=usecase
	@go run go.uber.org/mock/mockgen -source=./internal/usecase/reservation_usecase.go -destination=./mocks/usecase/reservation_usecase_mock.go -package=usecase

	# Generate mocks for the repository interfaces
	@go run go.uber.org/mock/mockgen -source=./internal/repository/warehouse_repository.go -destination=./mocks/repository/warehouse_repository_mock.go -package=repository
	@go run go.uber.org/mock/mockgen -source=./internal/repository/reservation_repository.go -destination=./mocks/repository/reservation_repository_mock.go -package=repository
	@echo "Mock generation complete"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/web/main.go
	@echo "Swagger documentation generated in docs/"

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -w $(GO_FILES)
	@echo "Formatting complete"

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...
	@echo "Vetting complete"

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 3000:3000 --name $(APP_NAME) -d $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)
	@echo "Docker container running on port 3000"

# Docker compose up
docker-up:
	@echo "Starting Docker Compose services..."
	@docker-compose up -d
	@echo "Docker Compose services started"

# Docker compose down
docker-down:
	@echo "Stopping Docker Compose services..."
	@docker-compose down
	@echo "Docker Compose services stopped"

# Create a new migration
migrate-create:
	@read -p "Enter table name: " table_name; \
	migrate create -ext $(MIGRATE_EXT) -dir $(MIGRATE_DIR) create_table_$$table_name
	@echo "Migration created in $(MIGRATE_DIR)"

# Run database migrations
migrate-up:
	@echo "Running database migrations..."
	@./scripts/run-migrations.sh up
	@echo "Migrations complete"

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	@./scripts/run-migrations.sh down
	@echo "Rollback complete"

# Display help information
help:
	@echo "Warehouse Service Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [command]"
	@echo ""
	@echo "Available Commands:"
	@echo "  all             Build and run the application"
	@echo "  build           Build the application"
	@echo "  clean           Clean build artifacts"
	@echo "  run             Run the application"
	@echo "  run-custom-port Run the application on port 3001"
	@echo "  test            Run tests"
	@echo "  test-e2e        Run end-to-end tests"
	@echo "  test-e2e-with-logs Run end-to-end tests with logs"
	@echo "  mock            Generate mock files for testing"
	@echo "  swagger         Generate Swagger documentation"
	@echo "  fmt             Format code"
	@echo "  vet             Vet code"
	@echo "  lint            Lint code"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-run      Run Docker container"
	@echo "  docker-up       Start Docker Compose services"
	@echo "  docker-down     Stop Docker Compose services"
	@echo "  migrate-create  Create a new migration"
	@echo "  migrate-up      Run database migrations"
	@echo "  migrate-down    Rollback database migrations"
	@echo "  help            Display this help message"

# Default target
.DEFAULT_GOAL := help

