#!/bin/sh

# Wait for MySQL to be ready
echo "Waiting for MySQL to be ready..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
  attempt=$((attempt+1))
  if mysql -h mysql -u root -e "SELECT 1;" >/dev/null 2>&1; then
    echo "MySQL is ready!"
    break
  fi
  echo "MySQL is not ready yet, waiting... ($attempt/$max_attempts)"
  sleep 2
  
  if [ $attempt -eq $max_attempts ]; then
    echo "Timed out waiting for MySQL"
    exit 1
  fi
done

# Create database tables manually
echo "Creating databases if they don't exist..."
mysql -h mysql -u root <<SQL
CREATE DATABASE IF NOT EXISTS shop_service;
CREATE DATABASE IF NOT EXISTS shop_service_test;
SQL

echo "Database setup completed successfully!"