# User Service

A RESTful API service for user management built with Go, Fiber framework, and MySQL.

## Features

- User registration
- User login with authentication
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

## API Endpoints

### Register User
```
POST /api/v1/users
```
Request Body:
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "1234567890",
  "password": "securepassword"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "5df84b6f-8f5b-4a51-a106-e9a46b67c836",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "1234567890",
    "created_at": "2025-05-17T09:23:37Z",
    "updated_at": "2025-05-17T09:23:37Z"
  }
}
```

### Login
```
POST /api/v1/users/login
```
Request Body:
```json
{
  "email": "john@example.com",
  "password": "securepassword"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "token": "1db8a967-9809-41ca-b65e-49b3de48c7a4"
  }
}
```

### Get User Details
```
GET /api/v1/users/:id
```
Headers:
```
Authorization: Bearer <token>
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "5df84b6f-8f5b-4a51-a106-e9a46b67c836",
    "authenticated_as": "5df84b6f-8f5b-4a51-a106-e9a46b67c836",
    "message": "User details fetched successfully"
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