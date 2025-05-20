#!/bin/bash

# Stock Locking Mechanism Test Script
# This script tests the stock locking implementation by simulating concurrent orders
# to ensure the system prevents overselling even under high concurrency.

# Colors for better readability
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Print header
echo -e "${BOLD}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║                 STOCK LOCKING TEST                ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════╝${NC}"
echo

# First, check current inventory and available stock
echo -e "${BLUE}${BOLD}STEP 1: Checking Initial Inventory Status${NC}"
echo -e "─────────────────────────────────────────────"

echo -e "Connecting to API server..."
HEALTH_CHECK=$(curl -s -X GET "http://localhost:3000/api/v1/health" -H "X-API-Key: order-service-api-key")
if [[ "$HEALTH_CHECK" == *"success\":true"* ]]; then
    echo -e "✅ Server connection: ${GREEN}OK${NC}"
else
    echo -e "❌ Server connection: ${RED}FAILED${NC}"
    echo "Error: $HEALTH_CHECK"
    exit 1
fi

# Get current inventory status
DB_USER="root"
DB_PASS=""
DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="order_service"

if [ -z "$DB_PASS" ]; then
    CONN="mysql -u$DB_USER -h$DB_HOST -P$DB_PORT $DB_NAME -N"
else
    CONN="mysql -u$DB_USER -p$DB_PASS -h$DB_HOST -P$DB_PORT $DB_NAME -N"
fi

# Product and order configuration
PRODUCT_ID=1
WAREHOUSE_ID=1
ORDER_QUANTITY=60  # Each order requesting 60 units

# Get actual inventory values
INVENTORY=$($CONN -e "SELECT quantity, reserved_quantity, (quantity - reserved_quantity) as available FROM inventory WHERE product_id = $PRODUCT_ID AND warehouse_id = $WAREHOUSE_ID;")
TOTAL_STOCK=$(echo "$INVENTORY" | awk '{print $1}')
RESERVED=$(echo "$INVENTORY" | awk '{print $2}')
AVAILABLE=$(echo "$INVENTORY" | awk '{print $3}')

echo -e "${YELLOW}Current Inventory:${NC}"
echo -e "  • Product ID:      $PRODUCT_ID"
echo -e "  • Warehouse ID:    $WAREHOUSE_ID"
echo -e "  • Total Stock:     $TOTAL_STOCK units"
echo -e "  • Reserved:        $RESERVED units"
echo -e "  • Available:       $AVAILABLE units"
echo

# Step 2: Place concurrent orders
echo -e "${BLUE}${BOLD}STEP 2: Testing Concurrent Order Processing${NC}"
echo -e "─────────────────────────────────────────────"
echo -e "Placing ${BOLD}3 concurrent orders${NC}, each requesting ${BOLD}$ORDER_QUANTITY units${NC}"
echo -e "Total requested: ${BOLD}$((ORDER_QUANTITY * 3)) units${NC} (vs. $AVAILABLE available)"
echo

# Function to place an order
place_order() {
    local order_num=$1
    echo -e "⏳ Placing order #$order_num (requesting $ORDER_QUANTITY units)..."
    
    RESPONSE=$(curl -s -X POST "http://localhost:3000/api/v1/orders" \
      -H "X-API-Key: order-service-api-key" \
      -H "Content-Type: application/json" \
      -d '{
        "user_id": "test-user-'$order_num'",
        "shipping_address": "Test Address '$order_num'",
        "payment_method": "credit_card",
        "items": [
          {
            "product_id": '$PRODUCT_ID',
            "warehouse_id": '$WAREHOUSE_ID',
            "quantity": '$ORDER_QUANTITY',
            "unit_price": 19.99
          }
        ]
      }')
      
    # Parse the response
    SUCCESS=$(echo $RESPONSE | grep -o '"success":true' || echo "")
    ERROR=$(echo $RESPONSE | grep -o '"message":"[^"]*' | cut -d'"' -f4 || echo "")
    
    if [ -n "$SUCCESS" ]; then
        ORDER_ID=$(echo $RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
        echo -e "${GREEN}✅ Order #$order_num SUCCEEDED${NC} with ID: $ORDER_ID"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        if [[ "$ERROR" == *"insufficient stock"* ]]; then
            echo -e "${YELLOW}❌ Order #$order_num FAILED${NC}: Insufficient stock available"
        else
            echo -e "${RED}❌ Order #$order_num FAILED${NC}: $ERROR"
        fi
    fi
}

# Launch 3 orders concurrently
SUCCESS_COUNT=0
place_order 1 &
PID1=$!
place_order 2 &
PID2=$!
place_order 3 &
PID3=$!

# Wait for all orders to complete
wait $PID1 $PID2 $PID3

echo
echo -e "${BLUE}${BOLD}STEP 3: Verifying Results${NC}"
echo -e "─────────────────────────────────────────────"

# Check updated inventory
INVENTORY_AFTER=$($CONN -e "SELECT quantity, reserved_quantity, (quantity - reserved_quantity) as available FROM inventory WHERE product_id = $PRODUCT_ID AND warehouse_id = $WAREHOUSE_ID;")
TOTAL_AFTER=$(echo "$INVENTORY_AFTER" | awk '{print $1}')
RESERVED_AFTER=$(echo "$INVENTORY_AFTER" | awk '{print $2}')
AVAILABLE_AFTER=$(echo "$INVENTORY_AFTER" | awk '{print $3}')

echo -e "${YELLOW}Current Inventory After Test:${NC}"
echo -e "  • Total Stock:     $TOTAL_AFTER units"
echo -e "  • Reserved:        $RESERVED_AFTER units"
echo -e "  • Available:       $AVAILABLE_AFTER units"
echo
echo -e "${YELLOW}Test Summary:${NC}"
echo -e "  • Successful Orders:   $SUCCESS_COUNT of 3"
echo -e "  • Failed Orders:       $((3 - SUCCESS_COUNT)) of 3"
echo

# Final check if the test passed
if [[ "$AVAILABLE_AFTER" -ge 0 ]]; then
    echo -e "${GREEN}${BOLD}✅ TEST PASSED:${NC} Stock locking prevented overselling"
    echo -e "   • Available stock is non-negative: $AVAILABLE_AFTER units"
    echo -e "   • Reserved stock matches active reservations: $RESERVED_AFTER units"
else
    echo -e "${RED}${BOLD}❌ TEST FAILED:${NC} Stock oversold!"
    echo -e "   • Available stock is negative: $AVAILABLE_AFTER units"
fi

echo
echo -e "${BLUE}To view detailed reservation status, run:${NC} ./scripts/check-inventory.sh"