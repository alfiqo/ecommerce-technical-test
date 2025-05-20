#!/bin/bash

# This script connects to the database and displays inventory information
# to verify that stock locking is working correctly

# Set database connection parameters directly
DB_USER="root"
DB_PASS=""
DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="order_service"

echo "===== Inventory Status Check ====="

# If mysql client is available
if command -v mysql &> /dev/null; then
    echo "Connecting to database to check inventory..."
    
    # Build connection string (handling empty password)
    if [ -z "$DB_PASS" ]; then
        CONN="mysql -u$DB_USER -h$DB_HOST -P$DB_PORT $DB_NAME"
    else
        CONN="mysql -u$DB_USER -p$DB_PASS -h$DB_HOST -P$DB_PORT $DB_NAME"
    fi
    
    # Run the query - we show all product inventory
    echo "Current inventory status:"
    $CONN -e "SELECT product_id, warehouse_id, quantity as total_stock, 
              reserved_quantity as reserved, 
              (quantity - reserved_quantity) as available 
              FROM inventory 
              ORDER BY product_id, warehouse_id;"
    
    # Show reservations
    echo
    echo "Active reservations:"
    $CONN -e "SELECT o.id as order_id, o.status, 
              r.product_id, r.warehouse_id, r.quantity, r.is_active,
              r.expires_at
              FROM stock_reservations r
              JOIN orders o ON r.order_id = o.id
              WHERE r.is_active = 1
              ORDER BY r.product_id, r.warehouse_id;"
else
    echo "MySQL client not found. Please install it to run database checks."
    echo "You can manually check using the MySQL client with these commands:"
    echo
    echo "  mysql -u$DB_USER -p -h$DB_HOST -P$DB_PORT $DB_NAME"
    echo
    echo "Once connected, run these queries:"
    echo "  SELECT product_id, warehouse_id, quantity as total_stock,"
    echo "         reserved_quantity as reserved,"
    echo "         (quantity - reserved_quantity) as available"
    echo "  FROM inventory"
    echo "  ORDER BY product_id, warehouse_id;"
    echo
    echo "  SELECT o.id as order_id, o.status,"
    echo "         r.product_id, r.warehouse_id, r.quantity, r.is_active,"
    echo "         r.expires_at"
    echo "  FROM stock_reservations r"
    echo "  JOIN orders o ON r.order_id = o.id"
    echo "  WHERE r.is_active = 1"
    echo "  ORDER BY r.product_id, r.warehouse_id;"
fi

echo
echo "Notes on interpreting the results:"
echo "1. The 'available' column should never be negative"
echo "2. The sum of all 'reserved' quantities for active reservations of a product/warehouse"
echo "   should match the 'reserved_quantity' in the inventory table"
echo "3. After processing payments, inventory 'quantity' should decrease and 'reserved_quantity'"
echo "   should also decrease accordingly"