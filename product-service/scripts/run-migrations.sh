#!/bin/bash

set -e

echo "Waiting for the database to be ready..."
max_retries=30
retry_count=0

while [ $retry_count -lt $max_retries ]; do
    if mysql -h product-service-mysql-e2e --port=3306 -u root -ppassword --ssl=0 -e "SELECT 1" &> /dev/null; then
        echo "Database is ready!"
        break
    fi
    retry_count=$((retry_count+1))
    echo "Waiting for database connection... ($retry_count/$max_retries)"
    sleep 1
done

if [ $retry_count -eq $max_retries ]; then
    echo "Failed to connect to the database after $max_retries attempts."
    exit 1
fi

echo "Running migrations..."
# Check if the products table already exists
if mysql -h product-service-mysql-e2e --port=3306 -u root -ppassword --ssl=0 -e "SELECT 1 FROM information_schema.tables WHERE table_schema = 'product_service_test' AND table_name = 'products' LIMIT 1;" 2>/dev/null | grep -q 1; then
    echo "Tables already exist, skipping creation..."
else
    echo "Creating tables..."
    mysql -h product-service-mysql-e2e --port=3306 -u root -ppassword --ssl=0 product_service_test < /app/db/migrations/20250517154713_create_table_products.up.sql
fi

echo "Inserting test data..."
# Check if test data already exists
if mysql -h product-service-mysql-e2e --port=3306 -u root -ppassword --ssl=0 -e "SELECT COUNT(*) FROM products WHERE uuid IN ('f47ac10b-58cc-4372-a567-0e02b2c3d479', 'f47ac10b-58cc-4372-a567-0e02b2c3d480');" product_service_test 2>/dev/null | grep -q "2"; then
    echo "Test data already exists, skipping insertion..."
else
    echo "Inserting fresh test data..."
    # Clear existing test data to avoid conflicts
    mysql -h product-service-mysql-e2e --port=3306 -u root -ppassword --ssl=0 product_service_test -e "DELETE FROM products WHERE uuid IN ('f47ac10b-58cc-4372-a567-0e02b2c3d479', 'f47ac10b-58cc-4372-a567-0e02b2c3d480');"
    
    # Insert fresh test data
    mysql -h product-service-mysql-e2e --port=3306 -u root -ppassword --ssl=0 product_service_test << EOF
INSERT INTO products (uuid, name, description, base_price, category, sku, barcode, thumbnail_url, image_urls, status)
VALUES 
  ('f47ac10b-58cc-4372-a567-0e02b2c3d479', 'Test Product 1', 'Description for test product 1', 99.99, 'Test Category', 'TEST-SKU-001', 'BARCODE-001', 'http://example.com/image1.jpg', 'http://example.com/image1.jpg,http://example.com/image1-2.jpg', 'active'),
  ('f47ac10b-58cc-4372-a567-0e02b2c3d480', 'Test Product 2', 'Description for test product 2', 149.99, 'Test Category', 'TEST-SKU-002', 'BARCODE-002', 'http://example.com/image2.jpg', 'http://example.com/image2.jpg,http://example.com/image2-2.jpg', 'active');
EOF
fi

echo "Migrations and test data setup completed!"