#!/bin/bash

# Define colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting integration tests...${NC}"

# Check if the database exists
DB_NAME=${DB_NAME:-shop_service_test}
DB_USER=${DB_USER:-root}
DB_PASS=${DB_PASS:-""}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}

# Create test database if it doesn't exist
echo -e "${YELLOW}Ensuring test database exists...${NC}"
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER $([[ -n "$DB_PASS" ]] && echo "-p$DB_PASS") -e "CREATE DATABASE IF NOT EXISTS $DB_NAME;" || {
    echo -e "${RED}Failed to create test database.${NC}"
    exit 1
}
echo -e "${GREEN}Test database ready.${NC}"

# Run API server in the background
echo -e "${YELLOW}Starting API server...${NC}"
# Set environment variables for testing
export DB_NAME=$DB_NAME
export DB_USER=$DB_USER
export DB_PASS=$DB_PASS
export DB_HOST=$DB_HOST 
export DB_PORT=$DB_PORT

# Create a config.test.json for the application
cat > config.test.json << EOF
{
  "app": {
    "name": "shop-service-test"
  },
  "web": {
    "prefork": false,
    "port": 3002
  },
  "log": {
    "level": 6
  },
  "database": {
    "username": "$DB_USER",
    "password": "$DB_PASS",
    "host": "$DB_HOST",
    "port": $DB_PORT,
    "name": "$DB_NAME",
    "auto_migrate": true,
    "pool": {
      "idle": 10,
      "max": 100,
      "lifetime": 300
    }
  }
}
EOF

# Start the server in the background with a different port and config
export PORT=3002
export CONFIG_FILE=config.test.json
go run ./cmd/web/main.go &
SERVER_PID=$!

# Give the server time to start and initialize
echo -e "${YELLOW}Waiting for server to start...${NC}"
sleep 10  # Give more time for database initialization

# Function to cleanup on exit
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    kill $SERVER_PID 2>/dev/null
    wait $SERVER_PID 2>/dev/null
    echo -e "${GREEN}Server stopped.${NC}"
}

# Set trap to clean up on exit
trap cleanup EXIT INT TERM

# Run the integration tests
echo -e "${YELLOW}Running integration tests...${NC}"
go test -v ./e2e/...

# Check the test result
TEST_RESULT=$?
if [ $TEST_RESULT -eq 0 ]; then
    echo -e "${GREEN}Integration tests passed.${NC}"
else
    echo -e "${RED}Integration tests failed.${NC}"
fi

exit $TEST_RESULT