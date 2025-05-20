MIGRATE_DIR=db/migrations
MIGRATE_EXT=sql

migrate-create:
	@read -p "Enter table name: " table_name; \
	migrate create -ext $(MIGRATE_EXT) -dir $(MIGRATE_DIR) create_table_$$table_name

run:
	go run ./cmd/web/main.go

test-run:
	go test -v -cover ./internal/...
	
test-e2e:
	./scripts/run-e2e-tests.sh

test-e2e-with-logs:
	./scripts/run-e2e-tests.sh --show-logs
	
mock:
# Generate mocks for the usecase interfaces
	mockgen -source=./internal/usecase/user_usecase.go -destination=./mocks/usecase/user_usecase_mock.go -package=usecase_mock
	mockgen -source=./internal/usecase/order_usecase.go -destination=./mocks/usecase/order_usecase_mock.go -package=usecase_mock
	mockgen -source=./internal/usecase/inventory_usecase.go -destination=./mocks/usecase/inventory_usecase_mock.go -package=usecase_mock

#Generate mocks for the repository interfaces
	mockgen -source=./internal/repository/user_repository.go -destination=./mocks/repository/user_repository_mock.go -package=repository_mock
	mockgen -source=./internal/repository/order_repository.go -destination=./mocks/repository/order_repository_mock.go -package=repository_mock
	mockgen -source=./internal/repository/inventory_repository.go -destination=./mocks/repository/inventory_repository_mock.go -package=repository_mock

# Stock locking tests
test-stock:
	@echo "Running basic stock locking test..."
	./scripts/test-stock-lock.sh

test-stock-load:
	@echo "Running stock locking load test..."
	./scripts/load-test.sh

check-inventory:
	@echo "Checking current inventory status..."
	./scripts/check-inventory.sh

# Combined stock test target
test-stock-all: check-inventory test-stock test-stock-load check-inventory
	@echo "All stock locking tests completed."


