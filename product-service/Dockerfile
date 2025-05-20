FROM golang:1.24-alpine

WORKDIR /app

# Install required build tools
RUN apk add --no-cache gcc musl-dev bash mysql-client

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Set environment variables
ENV CGO_ENABLED=1

# Expose the application port
EXPOSE 3001

# The command is specified in docker-compose.yml