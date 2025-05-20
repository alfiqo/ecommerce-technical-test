#!/bin/bash

# Generate mocks for the warehouse usecase interface
mockgen -source=./internal/usecase/warehouse_usecase.go -destination=./mocks/usecase/warehouse_usecase_mock.go -package=usecase_mock

# Generate mocks for the warehouse repository interface
mockgen -source=./internal/repository/warehouse_repository.go -destination=./mocks/repository/warehouse_repository_mock.go -package=repository_mock

echo "Warehouse mocks generated successfully"