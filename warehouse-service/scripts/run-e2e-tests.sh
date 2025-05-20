#!/bin/bash

# Create test results directory if it doesn't exist
mkdir -p test-results

# Build and start the services
echo "Starting services for E2E testing..."
docker-compose -f docker-compose.e2e.yml up --build -d

# Show logs if requested
if [ "$1" = "--show-logs" ]; then
  echo "Showing service logs (press Ctrl+C to stop viewing logs, tests will continue)..."
  docker-compose -f docker-compose.e2e.yml logs -f &
  LOGS_PID=$!
  # Setup trap to kill logs process when script exits
  trap "kill $LOGS_PID 2>/dev/null" EXIT
fi

# Check if the test service exited
echo "Waiting for tests to complete..."
docker wait user-service-e2e-test

# Get the test container's exit code
TEST_EXIT_CODE=$(docker inspect user-service-e2e-test --format='{{.State.ExitCode}}')

# Save test output
echo "Saving test output..."
docker logs user-service-e2e-test > test-results/e2e-test-output.log

# Stop and remove containers
echo "Cleaning up containers..."
docker-compose -f docker-compose.e2e.yml down

echo "E2E test results are saved in test-results/e2e-test-output.log"

# Exit with the same code as the test container
exit $TEST_EXIT_CODE