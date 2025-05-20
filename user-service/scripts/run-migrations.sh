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
echo "Creating tables manually..."
mysql -h mysql -u root <<SQL
CREATE DATABASE IF NOT EXISTS user_service_test;
USE user_service_test;

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    uuid       CHAR(36) NOT NULL,           
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    phone      VARCHAR(50) UNIQUE,
    password   VARCHAR(100) NOT NULL,
    token      VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (uuid)
) ENGINE = InnoDB;
SQL

# Verify table was created
echo "Verifying table was created..."
if mysql -h mysql -u root -e "USE user_service_test; SHOW TABLES;" | grep -q "users"; then
  echo "Database setup completed successfully!"
else
  echo "Failed to create tables!"
  exit 1
fi