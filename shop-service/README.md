# Shop Service

A RESTful API service for shop management built with Go, Fiber framework, and MySQL.

## Features

- Clean architecture design (repository, usecase, handler)
- MySQL database with migrations
- Containerized with Docker
- Standardized error handling with custom error types
- Request ID tracking for API calls
- Structured JSON logging with context
- Timeout handling and context management
- Consistent API responses
- API key authentication

## Prerequisites

To run this application, you need:

- Docker and Docker Compose
- Go 1.24+ (for local development)

## Running with Docker

1. Clone the repository:
   ```
   git clone <repository-url>
   cd shop-service
   ```

2. Start the services using Docker Compose:
   ```
   docker-compose up -d
   ```

   This will:
   - Build the shop service container
   - Start a MySQL container
   - Run database migrations
   - Start the application on port 3000

3. Access the API at:
   ```
   http://localhost:3000
   ```

## API Endpoints

### Health Check
```
GET /api/v1/health
```

Response:
```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
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

3. Run the application:
   ```
   make run
   ```

4. Run unit tests:
   ```
   make test-run
   ```

## End-to-End Testing

The project includes end-to-end tests that verify the API endpoints by testing against a running service.

1. Run E2E tests with Docker (recommended):
   ```
   make test-e2e
   ```

2. Run E2E tests with logs for debugging:
   ```
   make test-e2e-with-logs
   ```

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