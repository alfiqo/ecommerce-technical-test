#!/bin/bash

# This script performs a load test on the stock locking mechanism using Apache Bench
# It simulates multiple concurrent requests for the same inventory

# Colors for better readability
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Print header
echo -e "${BOLD}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║              STOCK LOCKING LOAD TEST              ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════╝${NC}"
echo

# Verify Apache Bench is installed
if ! command -v ab &> /dev/null; then
    echo -e "${RED}${BOLD}Error: Apache Bench (ab) is not installed.${NC}"
    echo "Please install it with:"
    echo "  - macOS: brew install httpd"
    echo "  - Ubuntu/Debian: apt-get install apache2-utils"
    echo "  - CentOS/RHEL: yum install httpd-tools"
    exit 1
fi

# Create a payload file for the order request
cat > order_payload.json << EOF
{
  "user_id": "load-test-user",
  "shipping_address": "123 Load Test St",
  "payment_method": "credit_card",
  "items": [
    {
      "product_id": 1,
      "warehouse_id": 1,
      "quantity": 10,
      "unit_price": 19.99
    }
  ]
}
EOF

# Set test parameters
CONCURRENCY=10
NUM_REQUESTS=50
PRODUCT_ID=1
WAREHOUSE_ID=1
QUANTITY_PER_REQUEST=10

# Calculate expected total stock needed
TOTAL_NEEDED=$((NUM_REQUESTS * QUANTITY_PER_REQUEST))

echo -e "${BLUE}${BOLD}TEST CONFIGURATION${NC}"
echo -e "─────────────────────────────────────────────"
echo -e "  • Concurrency:         ${BOLD}$CONCURRENCY${NC} users"
echo -e "  • Total requests:      ${BOLD}$NUM_REQUESTS${NC} orders"
echo -e "  • Product ID:          $PRODUCT_ID"
echo -e "  • Warehouse ID:        $WAREHOUSE_ID"
echo -e "  • Quantity per order:  $QUANTITY_PER_REQUEST units"
echo -e "  • Total stock needed:  ${BOLD}$TOTAL_NEEDED${NC} units"
echo

# Get current inventory status before test
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

# Get actual inventory values
INVENTORY=$($CONN -e "SELECT quantity, reserved_quantity, (quantity - reserved_quantity) as available FROM inventory WHERE product_id = $PRODUCT_ID AND warehouse_id = $WAREHOUSE_ID;")
TOTAL_STOCK=$(echo "$INVENTORY" | awk '{print $1}')
RESERVED=$(echo "$INVENTORY" | awk '{print $2}')
AVAILABLE=$(echo "$INVENTORY" | awk '{print $3}')

echo -e "${BLUE}${BOLD}CURRENT INVENTORY${NC}"
echo -e "─────────────────────────────────────────────"
echo -e "  • Total Stock:     ${BOLD}$TOTAL_STOCK${NC} units"
echo -e "  • Reserved:        ${BOLD}$RESERVED${NC} units" 
echo -e "  • Available:       ${BOLD}$AVAILABLE${NC} units"
echo

echo -e "${BLUE}${BOLD}RUNNING LOAD TEST${NC}"
echo -e "─────────────────────────────────────────────"
echo -e "Sending ${BOLD}$NUM_REQUESTS${NC} requests with ${BOLD}$CONCURRENCY${NC} concurrent users"
echo -e "Each order requests ${BOLD}$QUANTITY_PER_REQUEST${NC} units (total: ${BOLD}$TOTAL_NEEDED${NC})"
echo -e "Available inventory: ${BOLD}$AVAILABLE${NC} units"
echo

# Run Apache Bench test
echo -e "${YELLOW}Starting Apache Bench load test...${NC}"
echo
ab -n $NUM_REQUESTS -c $CONCURRENCY -p order_payload.json \
   -T 'application/json' \
   -H 'X-API-Key: order-service-api-key' \
   http://localhost:3000/api/v1/orders
echo

# Get inventory after test
INVENTORY_AFTER=$($CONN -e "SELECT quantity, reserved_quantity, (quantity - reserved_quantity) as available FROM inventory WHERE product_id = $PRODUCT_ID AND warehouse_id = $WAREHOUSE_ID;")
TOTAL_AFTER=$(echo "$INVENTORY_AFTER" | awk '{print $1}')
RESERVED_AFTER=$(echo "$INVENTORY_AFTER" | awk '{print $2}')
AVAILABLE_AFTER=$(echo "$INVENTORY_AFTER" | awk '{print $3}')

# Get successful orders count
SUCCESSFUL_ORDERS=$($CONN -e "SELECT COUNT(*) FROM orders o JOIN stock_reservations r ON o.id = r.order_id WHERE r.product_id = $PRODUCT_ID AND r.warehouse_id = $WAREHOUSE_ID AND r.is_active = 1 AND o.status = 'pending' AND o.created_at > DATE_SUB(NOW(), INTERVAL 2 MINUTE);")

echo -e "${BLUE}${BOLD}RESULTS SUMMARY${NC}"
echo -e "─────────────────────────────────────────────"
echo -e "  • Successful orders:    ${BOLD}$SUCCESSFUL_ORDERS${NC} of $NUM_REQUESTS"
echo -e "  • Failed orders:        ${BOLD}$((NUM_REQUESTS - SUCCESSFUL_ORDERS))${NC} of $NUM_REQUESTS"
echo
echo -e "${BLUE}${BOLD}INVENTORY AFTER TEST${NC}"
echo -e "─────────────────────────────────────────────"
echo -e "  • Total Stock:     ${BOLD}$TOTAL_AFTER${NC} units"
echo -e "  • Reserved:        ${BOLD}$RESERVED_AFTER${NC} units"
echo -e "  • Available:       ${BOLD}$AVAILABLE_AFTER${NC} units"
echo -e "  • Stock change:    ${BOLD}$((RESERVED_AFTER - RESERVED))${NC} units reserved during test"
echo

# Test verification
if [[ "$AVAILABLE_AFTER" -ge 0 ]]; then
    echo -e "${GREEN}${BOLD}✅ TEST PASSED:${NC} Stock locking prevented overselling"
    echo -e "   • Available stock is non-negative: $AVAILABLE_AFTER units"
    
    if [[ "$SUCCESSFUL_ORDERS" -gt 0 ]]; then
        echo -e "   • Successfully processed $SUCCESSFUL_ORDERS orders with stock locking"
    fi
    
    if [[ "$SUCCESSFUL_ORDERS" -lt "$NUM_REQUESTS" ]]; then
        echo -e "   • Correctly rejected orders when stock was depleted"
    fi
else
    echo -e "${RED}${BOLD}❌ TEST FAILED:${NC} Stock oversold!"
    echo -e "   • Available stock is negative: $AVAILABLE_AFTER units"
fi

echo
echo -e "${BLUE}To view detailed reservation status, run:${NC} ./scripts/check-inventory.sh"

# Clean up the payload file
rm order_payload.json