# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user-service ./cmd/web/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata mysql-client

# Copy binary from builder stage
COPY --from=builder /app/user-service .

# Copy config file
COPY config.docker.json ./config.json

# Copy migrations
COPY db/migrations ./db/migrations

# Copy scripts
COPY scripts/run-migrations.sh ./scripts/run-migrations.sh
RUN chmod +x ./scripts/run-migrations.sh

# Expose web port
EXPOSE 3000

# Set up a healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

# Command to run the application
CMD ["./user-service"]