# Warehouse Service

A RESTful API service for warehouse and inventory management built with Go, Fiber framework, and MySQL.

## Features

- Warehouse management (CRUD operations)
- Inventory tracking with stock levels
- Inventory reservation system with database-level locking
- Race condition prevention for concurrent stock operations
- Clean architecture design (repository, usecase, handler)
- MySQL database with migrations
- Containerized with Docker
- Standardized error handling with custom error types
- Request ID tracking for API calls
- Structured JSON logging with context
- Timeout handling and context management
- Consistent API responses

## Prerequisites

To run this application, you need:

- Docker and Docker Compose
- Go 1.24+ (for local development)

## Running with Docker

1. Clone the repository:
   ```
   git clone <repository-url>
   cd user-service
   ```

2. Start the services using Docker Compose:
   ```
   docker-compose up -d
   ```

   This will:
   - Build the user service container
   - Start a MySQL container
   - Run database migrations
   - Start the application on port 3000

3. Access the API at:
   ```
   http://localhost:3000
   ```

## Authentication

All API endpoints are protected with API key authentication. To access the API, you need to include the API key in the `X-API-Key` header of your requests:

```
X-API-Key: warehouse-service-api-key
```

This simplified authentication approach is suitable for service-to-service communication where a shared API key is used. In a production environment, you would:

1. Use unique API keys for each client
2. Store API keys securely in a database
3. Implement rate limiting and monitoring
4. Consider additional authentication methods for user-specific operations

## API Flow

### Get Warehouse Flow

```mermaid
sequenceDiagram
    participant Client
    participant WarehouseService
    participant WarehouseDB
    
    Note over Client, WarehouseDB: GET /warehouses/:id - Get warehouse details
    
    Client->>WarehouseService: GET /warehouses/123
    Note right of Client: X-API-Key: warehouse-service-api-key
    
    WarehouseService->>WarehouseService: Validate API key
    
    alt Invalid API key
        WarehouseService-->>Client: 401 Unauthorized
        Note right of WarehouseService: { "success": false, "error": { "code": "UNAUTHORIZED", "message": "Invalid API key" } }
    else Valid API key
        WarehouseService->>WarehouseService: Validate warehouse ID parameter
        
        WarehouseService->>WarehouseDB: SELECT * FROM warehouses WHERE id = ?
        WarehouseDB-->>WarehouseService: Warehouse record or null
        
        alt Warehouse not found
            WarehouseService-->>Client: 404 Not Found
            Note right of WarehouseService: { "success": false, "error": { "code": "RESOURCE_NOT_FOUND", "message": "Warehouse not found" } }
        else Warehouse found
            WarehouseService->>WarehouseDB: SELECT COUNT(*) FROM warehouse_stock WHERE warehouse_id = ?
            WarehouseDB-->>WarehouseService: Total product count
            
            WarehouseService->>WarehouseDB: SELECT SUM(quantity) as total_items FROM warehouse_stock WHERE warehouse_id = ?
            WarehouseDB-->>WarehouseService: Total item count
            
            WarehouseService->>WarehouseService: Map to response DTO with statistics
            
            WarehouseService-->>Client: 200 OK
            Note right of WarehouseService: { "success": true, "data": { "id": 123, "name": "Central Warehouse", "location": "Downtown", "address": "789 Industrial Blvd", "is_active": true, "stats": { "total_products": 150, "total_items": 5000 }, "created_at": "2024-01-01T00:00:00Z" } }
        end
    end
```

### Create Warehouse Flow

```mermaid
sequenceDiagram
    participant Client
    participant WarehouseService
    participant WarehouseDB
    
    Note over Client, WarehouseDB: POST /warehouses - Create a new warehouse (Admin only)
    
    Client->>WarehouseService: POST /warehouses
    Note right of Client: X-API-Key: warehouse-service-api-key
    Note right of Client: { "name": "North Warehouse", "location": "North District", "address": "456 North Ave" }
    
    WarehouseService->>WarehouseService: Validate API key
    
    alt Invalid API key
        WarehouseService-->>Client: 401 Unauthorized
        Note right of WarehouseService: { "success": false, "error": { "code": "UNAUTHORIZED", "message": "Invalid API key" } }
    else Valid API key
        WarehouseService->>WarehouseService: Validate request body
        
        alt Invalid input
            WarehouseService-->>Client: 400 Bad Request
            Note right of WarehouseService: { "success": false, "error": { "code": "INVALID_INPUT", "message": "Invalid input data" } }
        else Valid input
            WarehouseService->>WarehouseDB: INSERT INTO warehouses (name, location, address, is_active)
            WarehouseDB-->>WarehouseService: Warehouse created with ID
            
            WarehouseService->>WarehouseService: Map to response DTO
            
            WarehouseService-->>Client: 201 Created
            Note right of WarehouseService: { "success": true, "data": { "id": 456, "name": "North Warehouse", "location": "North District", "address": "456 North Ave", "is_active": true, "created_at": "2024-01-01T00:00:00Z" } }
        end
    end
```

### Get Warehouse Stock Flow

```mermaid
sequenceDiagram
    participant Client
    participant WarehouseService
    participant WarehouseDB
    
    Note over Client, WarehouseDB: GET /warehouses/:id/stock - List stock in warehouse
    
    Client->>WarehouseService: GET /warehouses/123/stock?page=1&limit=20&product_id=456
    Note right of Client: X-API-Key: warehouse-service-api-key
    
    WarehouseService->>WarehouseService: Validate API key
    
    alt Invalid API key
        WarehouseService-->>Client: 401 Unauthorized
        Note right of WarehouseService: { "success": false, "error": { "code": "UNAUTHORIZED", "message": "Invalid API key" } }
    else Valid API key
        WarehouseService->>WarehouseService: Parse query parameters (pagination, filters)
        
        WarehouseService->>WarehouseDB: SELECT * FROM warehouses WHERE id = ? AND is_active = true
        WarehouseDB-->>WarehouseService: Verify warehouse exists and is active
        
        alt Warehouse not found or inactive
            WarehouseService-->>Client: 404 Not Found
            Note right of WarehouseService: { "success": false, "error": { "code": "RESOURCE_NOT_FOUND", "message": "Warehouse not found or inactive" } }
        else Warehouse found
            WarehouseService->>WarehouseDB: SELECT COUNT(*) FROM warehouse_stock ws WHERE ws.warehouse_id = ? AND (? = 0 OR ws.product_id = ?)
            WarehouseDB-->>WarehouseService: Total stock record count
            
            WarehouseService->>WarehouseDB: SELECT ws.*, p.name as product_name, p.sku FROM warehouse_stock ws LEFT JOIN products p ON ws.product_id = p.id WHERE ws.warehouse_id = ? AND (? = 0 OR ws.product_id = ?) ORDER BY ws.updated_at DESC LIMIT 20 OFFSET 0
            WarehouseDB-->>WarehouseService: Stock records with product info
            
            WarehouseService->>WarehouseService: Map to response DTOs
            
            WarehouseService-->>Client: 200 OK
            Note right of WarehouseService: { "success": true, "data": [{ "product_id": 456, "product_name": "Laptop", "sku": "LAP-001", "quantity": 50, "available_quantity": 45, "reserved_quantity": 5, "updated_at": "2024-01-01T00:00:00Z" }], "meta": { "page": 1, "limit": 20, "total": 150, "total_page": 8 } }
        end
    end
```

### Add Stock to Warehouse Flow

```mermaid
sequenceDiagram
    participant Client
    participant WarehouseService
    participant WarehouseDB
    participant MessageQueue
    participant ProductService
    
    Note over Client, ProductService: POST /warehouses/:id/stock - Add stock to warehouse
    
    Client->>WarehouseService: POST /warehouses/123/stock
    Note right of Client: X-API-Key: warehouse-service-api-key
    Note right of Client: { "product_id": 456, "quantity": 100, "reference": "PURCHASE-001", "notes": "New inventory received" }
    
    WarehouseService->>WarehouseService: Validate API key
    
    alt Invalid API key
        WarehouseService-->>Client: 401 Unauthorized
        Note right of WarehouseService: { "success": false, "error": { "code": "UNAUTHORIZED", "message": "Invalid API key" } }
    else Valid API key
        WarehouseService->>WarehouseService: Validate request body
        
        alt Invalid input
            WarehouseService-->>Client: 400 Bad Request
            Note right of WarehouseService: { "success": false, "error": { "code": "INVALID_INPUT", "message": "Invalid input data" } }
        else Valid input
            WarehouseService->>WarehouseDB: SELECT * FROM warehouses WHERE id = ? AND is_active = true
            WarehouseDB-->>WarehouseService: Verify warehouse exists and is active
            
            alt Warehouse not found or inactive
                WarehouseService-->>Client: 400 Bad Request
                Note right of WarehouseService: { "success": false, "error": { "code": "BUSINESS_RULE_VIOLATION", "message": "Cannot add stock to inactive warehouse" } }
            else Warehouse is active
                WarehouseService->>WarehouseDB: BEGIN TRANSACTION
                
                WarehouseService->>WarehouseDB: SELECT * FROM warehouse_stock WHERE warehouse_id = ? AND product_id = ? FOR UPDATE
                WarehouseDB-->>WarehouseService: Existing stock record or null
                
                alt Stock record exists
                    WarehouseService->>WarehouseDB: UPDATE warehouse_stock SET quantity = quantity + ?, available_quantity = available_quantity + ?, updated_at = NOW() WHERE warehouse_id = ? AND product_id = ?
                    WarehouseDB-->>WarehouseService: Stock updated
                else Stock record doesn't exist
                    WarehouseService->>WarehouseDB: INSERT INTO warehouse_stock (warehouse_id, product_id, quantity, available_quantity, reserved_quantity)
                    WarehouseDB-->>WarehouseService: Stock record created
                end
                
                WarehouseService->>WarehouseDB: INSERT INTO stock_movements (warehouse_id, product_id, movement_type, quantity, reference_type, reference_id, notes)
                WarehouseDB-->>WarehouseService: Movement logged
                
                WarehouseService->>WarehouseDB: COMMIT TRANSACTION
                
                WarehouseService->>MessageQueue: Publish StockAdded event
                MessageQueue->>ProductService: Update product stock aggregates
                
                WarehouseService->>WarehouseDB: SELECT ws.*, p.name as product_name FROM warehouse_stock ws LEFT JOIN products p ON ws.product_id = p.id WHERE ws.warehouse_id = ? AND ws.product_id = ?
                WarehouseDB-->>WarehouseService: Updated stock record with product info
                
                WarehouseService->>WarehouseService: Map to response DTO
                
                WarehouseService-->>Client: 200 OK
                Note right of WarehouseService: { "success": true, "data": { "warehouse_id": 123, "product_id": 456, "product_name": "Laptop", "quantity": 150, "available_quantity": 145, "reserved_quantity": 5, "updated_at": "2024-01-01T00:00:00Z" } }
            end
        end
    end
```

### Transfer Stock Between Warehouses Flow

```mermaid
sequenceDiagram
    participant Client
    participant WarehouseService
    participant WarehouseDB
    participant MessageQueue
    participant ProductService
    
    Note over Client, ProductService: POST /warehouses/transfer - Transfer stock between warehouses
    
    Client->>WarehouseService: POST /warehouses/transfer
    Note right of Client: X-API-Key: warehouse-service-api-key
    Note right of Client: { "source_warehouse_id": 123, "target_warehouse_id": 456, "product_id": 789, "quantity": 25, "notes": "Redistribution" }
    
    WarehouseService->>WarehouseService: Validate API key
    
    alt Invalid API key
        WarehouseService-->>Client: 401 Unauthorized
        Note right of WarehouseService: { "success": false, "error": { "code": "UNAUTHORIZED", "message": "Invalid API key" } }
    else Valid API key
        WarehouseService->>WarehouseService: Validate request body
        
        alt Invalid input
            WarehouseService-->>Client: 400 Bad Request
            Note right of WarehouseService: { "success": false, "error": { "code": "INVALID_INPUT", "message": "Invalid input data" } }
        else Valid input
            WarehouseService->>WarehouseDB: SELECT * FROM warehouses WHERE id IN (?, ?) AND is_active = true
            WarehouseDB-->>WarehouseService: Verify both warehouses exist and are active
            
            alt Source or target warehouse inactive
                WarehouseService-->>Client: 400 Bad Request
                Note right of WarehouseService: { "success": false, "error": { "code": "BUSINESS_RULE_VIOLATION", "message": "Source or target warehouse is inactive" } }
            else Both warehouses active
                WarehouseService->>WarehouseDB: BEGIN TRANSACTION
                
                Note over WarehouseService, WarehouseDB: Lock stocks in consistent order (by warehouse_id)
                WarehouseService->>WarehouseDB: SELECT * FROM warehouse_stock WHERE (warehouse_id = ? OR warehouse_id = ?) AND product_id = ? ORDER BY warehouse_id FOR UPDATE
                WarehouseDB-->>WarehouseService: Source and target stock records (with locks)
                
                WarehouseService->>WarehouseService: Find source and target stocks from result
                
                alt Insufficient source stock
                    WarehouseService->>WarehouseDB: ROLLBACK TRANSACTION
                    WarehouseService-->>Client: 400 Bad Request
                    Note right of WarehouseService: { "success": false, "error": { "code": "BUSINESS_RULE_VIOLATION", "message": "Insufficient stock in source warehouse" } }
                else Sufficient source stock
                    WarehouseService->>WarehouseDB: INSERT INTO stock_transfers (source_warehouse_id, target_warehouse_id, product_id, quantity, status, transfer_reference)
                    WarehouseDB-->>WarehouseService: Transfer record created
                    
                    WarehouseService->>WarehouseDB: UPDATE warehouse_stock SET quantity = quantity - ?, available_quantity = available_quantity - ? WHERE warehouse_id = ? AND product_id = ?
                    WarehouseDB-->>WarehouseService: Source stock decreased
                    
                    alt Target stock exists
                        WarehouseService->>WarehouseDB: UPDATE warehouse_stock SET quantity = quantity + ?, available_quantity = available_quantity + ? WHERE warehouse_id = ? AND product_id = ?
                        WarehouseDB-->>WarehouseService: Target stock increased
                    else Target stock doesn't exist
                        WarehouseService->>WarehouseDB: INSERT INTO warehouse_stock (warehouse_id, product_id, quantity, available_quantity, reserved_quantity)
                        WarehouseDB-->>WarehouseService: Target stock created
                    end
                    
                    WarehouseService->>WarehouseDB: INSERT INTO stock_movements (warehouse_id, product_id, movement_type, quantity, reference_type, reference_id)
                    Note right of WarehouseDB: Create movements for both source (transfer_out) and target (transfer_in)
                    WarehouseDB-->>WarehouseService: Movements logged
                    
                    WarehouseService->>WarehouseDB: UPDATE stock_transfers SET status = 'completed' WHERE id = ?
                    WarehouseDB-->>WarehouseService: Transfer marked as completed
                    
                    WarehouseService->>WarehouseDB: COMMIT TRANSACTION
                    
                    WarehouseService->>MessageQueue: Publish StockTransferred event
                    MessageQueue->>ProductService: Update product stock aggregates
                    
                    WarehouseService->>WarehouseService: Map to response DTO
                    
                    WarehouseService-->>Client: 200 OK
                    Note right of WarehouseService: { "success": true, "data": { "transfer_id": 101, "source_warehouse_id": 123, "target_warehouse_id": 456, "product_id": 789, "quantity": 25, "status": "completed", "transfer_reference": "TRF-12345", "created_at": "2024-01-01T00:00:00Z" } }
                end
            end
        end
    end
```

### Inventory Reservation Flow (with Database Locking)

```mermaid
sequenceDiagram
    participant Client
    participant ReservationHandler
    participant ReservationUsecase
    participant Repository
    participant Database
    
    Note over Client, Database: POST /inventory/reserve - Reserve inventory with locking
    
    Client->>ReservationHandler: POST /inventory/reserve
    Note right of Client: X-API-Key: warehouse-service-api-key
    Note right of Client: { "warehouse_id": 1, "product_id": 5, "quantity": 10 }
    
    ReservationHandler->>ReservationHandler: Validate API key
    
    alt Invalid API key
        ReservationHandler-->>Client: 401 Unauthorized
        Note right of ReservationHandler: { "success": false, "error": { "code": "UNAUTHORIZED", "message": "Invalid API key" } }
    else Valid API key
        ReservationHandler->>ReservationHandler: Validate request body
        
        alt Invalid request
            ReservationHandler-->>Client: 400 Bad Request
            Note right of ReservationHandler: { "success": false, "error": { "code": "INVALID_INPUT", "message": "Invalid request body" } }
        else Valid request
            ReservationHandler->>ReservationUsecase: ReserveStock(warehouseId, productId, quantity)
            
            ReservationUsecase->>ReservationUsecase: Start DB Transaction
            
            ReservationUsecase->>Repository: FindWarehouse(warehouseId)
            Repository->>Database: SELECT * FROM warehouses WHERE id = ?
            Database-->>Repository: Warehouse data
            Repository-->>ReservationUsecase: Warehouse entity
            
            alt Warehouse not found or inactive
                ReservationUsecase-->>ReservationHandler: Error: Warehouse not found or inactive
                ReservationHandler-->>Client: 404 Not Found
                Note right of ReservationHandler: { "success": false, "error": { "code": "RESOURCE_NOT_FOUND", "message": "Warehouse not found" } }
            else Warehouse found and active
                Note over Repository, Database: Critical section with DB locking
                
                ReservationUsecase->>Repository: ReserveStock(tx, warehouseId, productId, quantity)
                Repository->>Database: SELECT * FROM warehouse_stock WHERE warehouse_id = ? AND product_id = ? FOR UPDATE
                Note right of Repository: Lock stock record to prevent race conditions
                Database-->>Repository: Stock data (locked for this transaction)
                
                Repository->>Repository: Check if available quantity >= requested quantity
                
                alt Insufficient stock
                    Repository-->>ReservationUsecase: Error: Insufficient stock
                    ReservationUsecase-->>ReservationHandler: Error: Insufficient stock
                    ReservationUsecase->>ReservationUsecase: Rollback transaction
                    ReservationHandler-->>Client: 422 Unprocessable Entity
                    Note right of ReservationHandler: { "success": false, "error": { "code": "BUSINESS_RULE_VIOLATION", "message": "Insufficient stock: requested 10, available 5" } }
                else Sufficient stock
                    Repository->>Repository: Update reserved quantity
                    Repository->>Database: UPDATE warehouse_stock SET reserved_quantity = reserved_quantity + ? WHERE warehouse_id = ? AND product_id = ?
                    Database-->>Repository: Update confirmation
                    
                    ReservationUsecase->>Repository: Create reservation log
                    Repository->>Database: INSERT INTO reservation_logs (warehouse_id, product_id, quantity, status, reference) VALUES (?, ?, ?, 'pending', ?)
                    Database-->>Repository: Insert confirmation
                    
                    ReservationUsecase->>ReservationUsecase: Commit transaction
                    Note right of Repository: Release locks on commit
                    
                    ReservationUsecase-->>ReservationHandler: Reservation response with reference
                    ReservationHandler-->>Client: 200 OK
                    Note right of ReservationHandler: { "success": true, "data": { "warehouse_id": 1, "product_id": 5, "reserved_quantity": 10, "available_quantity": 90, "reference": "RSV-1-5-1234567890" } }
                end
            end
        end
    end
```

## API Endpoints

### Warehouse Management

#### Get Warehouse
```
GET /api/v1/warehouses/:id
```
Headers:
```
X-API-Key: warehouse-service-api-key
```

Response:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Main Warehouse",
    "location": "New York",
    "is_active": true,
    "created_at": "2025-05-17T09:23:37Z",
    "updated_at": "2025-05-17T09:23:37Z",
    "stats": {
      "product_count": 25,
      "total_items": 1500
    }
  }
}
```

#### Create Warehouse
```
POST /api/v1/warehouses
```
Headers:
```
X-API-Key: warehouse-service-api-key
```
Request Body:
```json
{
  "name": "Main Warehouse",
  "location": "New York",
  "is_active": true
}
```

### Inventory Reservation System

#### Reserve Stock
```
POST /api/v1/inventory/reserve
```
Headers:
```
X-API-Key: warehouse-service-api-key
```
Request Body:
```json
{
  "warehouse_id": 1,
  "product_id": 5,
  "quantity": 10
}
```

Response:
```json
{
  "success": true,
  "data": {
    "warehouse_id": 1,
    "product_id": 5,
    "reserved_quantity": 10,
    "available_quantity": 90,
    "total_quantity": 100,
    "reference": "RSV-1-5-1715969465",
    "status": "pending",
    "reservation_time": "2025-05-18T21:37:45+07:00"
  }
}
```

cURL Example:
```bash
curl -X POST 'http://localhost:3000/api/v1/inventory/reserve' \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: warehouse-service-api-key' \
  -d '{
    "warehouse_id": 1,
    "product_id": 5,
    "quantity": 10
  }'
```

#### Cancel Reservation
```
POST /api/v1/inventory/reserve/cancel
```
Headers:
```
X-API-Key: warehouse-service-api-key
```
Request Body:
```json
{
  "warehouse_id": 1,
  "product_id": 5,
  "quantity": 10,
  "reference": "RSV-1-5-1715969465"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "message": "Reservation cancelled successfully"
  }
}
```

cURL Example:
```bash
curl -X POST 'http://localhost:3000/api/v1/inventory/reserve/cancel' \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: warehouse-service-api-key' \
  -d '{
    "warehouse_id": 1,
    "product_id": 5,
    "quantity": 10,
    "reference": "RSV-1-5-1715969465"
  }'
```

#### Commit Reservation
```
POST /api/v1/inventory/reserve/commit
```
Headers:
```
X-API-Key: warehouse-service-api-key
```
Request Body:
```json
{
  "warehouse_id": 1,
  "product_id": 5,
  "quantity": 10,
  "reference": "RSV-1-5-1715969465"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "message": "Reservation committed successfully"
  }
}
```

cURL Example:
```bash
curl -X POST 'http://localhost:3000/api/v1/inventory/reserve/commit' \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: warehouse-service-api-key' \
  -d '{
    "warehouse_id": 1,
    "product_id": 5,
    "quantity": 10,
    "reference": "RSV-1-5-1715969465"
  }'
```

#### Get Reservation History
```
GET /api/v1/inventory/warehouses/:warehouse_id/products/:product_id/reservations?page=1&limit=20
```
Headers:
```
X-API-Key: warehouse-service-api-key
```

Response:
```json
{
  "success": true,
  "data": {
    "warehouse_id": 1,
    "product_id": 5,
    "total": 3,
    "page": 1,
    "limit": 20,
    "logs": [
      {
        "quantity": 10,
        "status": "committed",
        "reference": "RSV-1-5-1715969465",
        "created_at": "2025-05-18T21:37:45+07:00"
      },
      {
        "quantity": 5,
        "status": "cancelled",
        "reference": "RSV-1-5-1715969155",
        "created_at": "2025-05-18T21:32:35+07:00"
      },
      {
        "quantity": 8,
        "status": "pending",
        "reference": "RSV-1-5-1715968930",
        "created_at": "2025-05-18T21:28:50+07:00"
      }
    ]
  }
}
```

cURL Example:
```bash
curl -X GET 'http://localhost:3000/api/v1/inventory/warehouses/1/products/5/reservations?page=1&limit=20' \
  -H 'X-API-Key: warehouse-service-api-key'
```

### Error Response Format
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

## Local Development

1. Install dependencies:
   ```
   go mod download
   ```

2. Start a local MySQL instance or update the config.json with your database details

3. Run database migrations:
   ```
   make migrate-up
   ```

4. Run the application:
   ```
   make run
   ```
   
   Or run on a custom port (3001):
   ```
   make run-custom-port
   ```

5. Run unit tests:
   ```
   make test
   ```

6. Generate Swagger documentation:
   ```
   make swagger
   ```
   
   Then access the Swagger UI at: http://localhost:3000/swagger/

## Database Schema

### Entity Relationship Diagram

```mermaid
erDiagram
    Warehouse ||--o{ WarehouseStock : "contains"
    Warehouse ||--o{ ReservationLog : "has"
    Warehouse ||--o{ StockTransfer : "source_warehouse"
    Warehouse ||--o{ StockTransfer : "target_warehouse"
    
    Warehouse {
        uint id PK
        string name
        string location
        string address
        bool is_active
        datetime created_at
        datetime updated_at
    }
    
    WarehouseStock {
        uint id PK
        uint warehouse_id FK
        uint product_id
        int quantity
        int reserved_quantity
        datetime updated_at
    }
    
    ReservationLog {
        uint id PK
        uint warehouse_id FK
        uint product_id
        int quantity
        enum status "pending,committed,cancelled"
        string reference
        datetime created_at
    }
    
    StockTransfer {
        uint id PK
        uint source_warehouse_id FK
        uint target_warehouse_id FK
        uint product_id
        int quantity
        enum status "pending,completed,failed"
        string transfer_reference
        datetime created_at
        datetime updated_at
    }
```

### Notes on Database Implementation

1. **Missing Tables**: 
   - While our sequence diagrams reference a `stock_movements` table, this table is not yet implemented in our migrations or entity models.
   - The `products` table is referenced in queries but not yet implemented.

2. **Computed Fields**:
   - `available_quantity` in `WarehouseStock` is computed on the fly using the formula `quantity - reserved_quantity`.

3. **Type Inconsistencies**:
   - The `warehouse_id` and `product_id` in `reservation_logs` are defined as `BIGINT UNSIGNED` in the migration but as `uint` in the entity.
   
4. **Indexes**:
   - Additional indexes were added to `warehouse_stock` for better query performance.
   - The `reservation_logs` table has indexes on `warehouse_id`, `product_id`, `reference`, and `created_at`.

## Available Make Commands

For a complete list of available commands, run:
```
make help
```

Key commands include:
- `make build`: Build the application
- `make run`: Run the application
- `make test`: Run tests
- `make swagger`: Generate Swagger documentation
- `make docker-up`: Start the application with Docker Compose
- `make docker-down`: Stop Docker Compose services

## End-to-End Testing

The project includes end-to-end tests that verify the complete user flow by testing against a running service.

1. Run E2E tests with Docker (recommended):
   ```
   make test-e2e
   ```

2. Run E2E tests with logs for debugging:
   ```
   make test-e2e-with-logs
   ```

The E2E tests will:
- Start the service with a test database
- Test the user registration endpoint
- Test the user login endpoint with valid and invalid credentials
- Test authentication middleware
- Clean up after the tests

Test results are saved in the `test-results` directory.

## Project Structure

- `/cmd/web`: Main application entry point
- `/internal`: Internal application code
  - `/config`: Configuration handling
  - `/context`: Context management for timeouts and request tracking
  - `/delivery`: HTTP delivery layer
    - `/http/middleware`: HTTP middleware (logging, auth)
    - `/http/response`: Standardized response formatting
    - `/http/route`: API route definitions
  - `/entity`: Domain entities
  - `/errors`: Custom error types and error handling
  - `/handler`: HTTP handlers
  - `/model`: Data models
  - `/repository`: Data access layer
  - `/usecase`: Business logic layer
- `/db/migrations`: Database migration files
- `/e2e`: End-to-end tests
- `/mocks`: Mock interfaces for testing

## Inventory Reservation System and Race Condition Prevention

The warehouse service includes a robust inventory reservation system that uses database-level locking to prevent race conditions when multiple users try to reserve the same stock simultaneously.

### How It Works

1. **Pessimistic Locking**: When a reservation is made, the system uses SQL's `FOR UPDATE` clause to lock the stock record for the duration of the transaction. This prevents other transactions from modifying the same record until the first transaction completes.

2. **Transactional Integrity**: All reservation operations (reserve, cancel, commit) are executed within database transactions, ensuring atomicity and consistency.

3. **Reservation Lifecycle**:
   - **Reserve**: Locks the stock record, checks availability, increases reserved quantity
   - **Cancel**: Locks the stock record, decreases reserved quantity
   - **Commit**: Locks the stock record, decreases both reserved quantity and total quantity

4. **Audit Trail**: All reservation activities are logged in the `reservation_logs` table with timestamps and status.

### Implementation Details

The core of the locking mechanism is implemented in the `ReservationRepository` using GORM's locking API:

```go
// Lock the stock record with FOR UPDATE to prevent concurrent modifications
stock := new(entity.WarehouseStock)
result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
    First(stock)
```

This translates to the following SQL:

```sql
SELECT * FROM warehouse_stock 
WHERE warehouse_id = ? AND product_id = ? 
LIMIT 1 
FOR UPDATE;
```

The `FOR UPDATE` clause acquires an exclusive lock on the selected rows until the transaction is committed or rolled back, preventing race conditions and ensuring that inventory is never over-committed.

## Configuration

Configuration is stored in `config.json`. For Docker, use `config.docker.json`.

Key configurations:
- Web server port (default: 3000)
- Database connection parameters
- Logging level (0-6, with 6 being most verbose)

## Error Handling

The service uses a standardized error handling approach:
- Custom error types with error codes and HTTP status codes
- Consistent error response format
- Contextual error logging
- Request tracking with request IDs

## Logging

Logs are output in JSON format and include:
- Timestamp
- Log level
- Request ID for request tracing
- Contextual information
- Method and path for API requests
- Response status and latency
- SQL query tracing (in debug mode)