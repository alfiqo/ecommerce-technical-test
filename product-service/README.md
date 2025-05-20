# Product Service

A RESTful API service for product management built with Go, Fiber framework, and MySQL, following Clean Architecture principles.

## Features

- List all products with pagination
- Get product details by ID
- Clean architecture design (repository, usecase, handler)
- MySQL database with migrations
- Containerized with Docker (configuration included)
- Comprehensive unit and end-to-end testing
- Good test coverage (repository: 82.6%, handler: 69.2%, usecase: 38.1%)

## Prerequisites

To run this application, you need:

- Docker and Docker Compose
- Go 1.24+ (for local development)

## Running with Docker

1. Clone the repository:
   ```
   git clone <repository-url>
   cd product-service
   ```

2. Start the services using Docker Compose:
   ```
   docker-compose up -d
   ```

   This will:
   - Build the product service container
   - Start a MySQL container
   - Run database migrations
   - Start the application on port 3001

3. Access the API at:
   ```
   http://localhost:3001
   ```

## API Endpoints

### Get Products (with pagination)
```
GET /api/v1/products?limit=10&offset=0
```

Response:
```json
{
  "data": {
    "products": [
      {
        "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "name": "Product Name",
        "description": "Product Description",
        "price": 99.99,
        "stock": 10,
        "category": "Category",
        "sku": "SKU-001",
        "image_url": "http://example.com/image.jpg",
        "created_at": "2025-05-17T10:00:00Z",
        "updated_at": "2025-05-17T10:00:00Z"
      }
    ],
    "count": 1,
    "limit": 10,
    "offset": 0
  }
}
```

### Get Product By ID
```
GET /api/v1/products/{id}
```

Response:
```json
{
  "data": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "name": "Product Name",
    "description": "Product Description",
    "price": 99.99,
    "stock": 10,
    "category": "Category",
    "sku": "SKU-001",
    "image_url": "http://example.com/image.jpg",
    "created_at": "2025-05-17T10:00:00Z",
    "updated_at": "2025-05-17T10:00:00Z"
  }
}
```

## Local Development

1. Install dependencies:
   ```
   go mod download
   ```

2. Start a local MySQL instance or update the config.json with your database details

3. Run the application:
   ```
   go run cmd/web/main.go
   ```

4. Run unit tests:
   ```
   go test ./...
   ```

## Project Structure

- `/cmd/web`: Main application entry point
- `/internal`: Internal application code
  - `/config`: Configuration handling
  - `/delivery`: HTTP delivery layer
  - `/entity`: Domain entities
  - `/handler`: HTTP handlers
  - `/model`: Data models and converters
  - `/repository`: Data access layer
  - `/usecase`: Business logic layer
- `/db/migrations`: Database migration files
- `/mocks`: Mock implementations for testing

## Configuration

Configuration is stored in `config.json`.

Key configurations:
- Web server port (default: 3001)
- Database connection parameters
- Logging level