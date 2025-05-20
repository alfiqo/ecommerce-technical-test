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

test-integration:
	./scripts/run-integration-tests.sh
	
mock:
# Generate mocks for the usecase interfaces

#Generate mocks for the repository interfaces


