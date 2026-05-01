#!/bin/bash

# AgileOS Load Testing Script (Bash Alternative to k6)
# Uses Apache Bench (ab) and curl for load testing
# Install: sudo apt-get install apache2-utils curl jq

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8081}"
API_BASE="$BASE_URL/api/v1"
CONCURRENT_USERS="${CONCURRENT_USERS:-50}"
TOTAL_REQUESTS="${TOTAL_REQUESTS:-1000}"
RESULTS_DIR="load-test-results"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 AgileOS Load Testing Script${NC}"
echo "=================================="
echo "Target: $BASE_URL"
echo "Concurrent Users: $CONCURRENT_USERS"
echo "Total Requests: $TOTAL_REQUESTS"
echo ""

# Create results directory
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$RESULTS_DIR/load-test-$TIMESTAMP.txt"

# Check if required tools are installed
command -v ab >/dev/null 2>&1 || { echo -e "${RED}Error: Apache Bench (ab) is not installed. Install with: sudo apt-get install apache2-utils${NC}"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo -e "${RED}Error: curl is not installed${NC}"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${YELLOW}Warning: jq is not installed. JSON parsing will be limited${NC}"; }

# Function to log results
log_result() {
    echo "$1" | tee -a "$REPORT_FILE"
}

# Function to get auth token
get_auth_token() {
    local username="$1"
    local password="$2"
    
    local response=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    if command -v jq >/dev/null 2>&1; then
        echo "$response" | jq -r '.access_token'
    else
        echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4
    fi
}

# Start testing
log_result ""
log_result "$(date): Starting Load Test"
log_result "=================================="

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Health Check Endpoint${NC}"
log_result ""
log_result "Test 1: Health Check Endpoint"
log_result "------------------------------"

ab -n "$TOTAL_REQUESTS" -c "$CONCURRENT_USERS" -g "$RESULTS_DIR/health-$TIMESTAMP.tsv" \
    "$BASE_URL/health" 2>&1 | tee -a "$REPORT_FILE"

# Test 2: Login Performance
echo -e "${YELLOW}Test 2: Login Performance${NC}"
log_result ""
log_result "Test 2: Login Performance"
log_result "-------------------------"

# Create temp file with login payload
LOGIN_PAYLOAD='{"username":"admin","password":"password123"}'
echo "$LOGIN_PAYLOAD" > "$RESULTS_DIR/login-payload.json"

ab -n "$TOTAL_REQUESTS" -c "$CONCURRENT_USERS" \
    -p "$RESULTS_DIR/login-payload.json" \
    -T "application/json" \
    -g "$RESULTS_DIR/login-$TIMESTAMP.tsv" \
    "$API_BASE/auth/login" 2>&1 | tee -a "$REPORT_FILE"

# Get token for authenticated tests
echo -e "${YELLOW}Getting authentication token...${NC}"
TOKEN=$(get_auth_token "admin" "password123")

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
    echo -e "${RED}Failed to get authentication token. Skipping authenticated tests.${NC}"
    log_result "Failed to get authentication token. Skipping authenticated tests."
else
    echo -e "${GREEN}✓ Authentication successful${NC}"
    
    # Test 3: Get Workflows (Authenticated)
    echo -e "${YELLOW}Test 3: Get Workflows (Authenticated)${NC}"
    log_result ""
    log_result "Test 3: Get Workflows (Authenticated)"
    log_result "--------------------------------------"
    
    ab -n "$TOTAL_REQUESTS" -c "$CONCURRENT_USERS" \
        -H "Authorization: Bearer $TOKEN" \
        -g "$RESULTS_DIR/workflows-$TIMESTAMP.tsv" \
        "$API_BASE/workflows" 2>&1 | tee -a "$REPORT_FILE"
    
    # Test 4: Get Analytics (Authenticated)
    echo -e "${YELLOW}Test 4: Get Analytics (Authenticated)${NC}"
    log_result ""
    log_result "Test 4: Get Analytics (Authenticated)"
    log_result "--------------------------------------"
    
    ab -n "$TOTAL_REQUESTS" -c "$CONCURRENT_USERS" \
        -H "Authorization: Bearer $TOKEN" \
        -g "$RESULTS_DIR/analytics-$TIMESTAMP.tsv" \
        "$API_BASE/analytics/overview" 2>&1 | tee -a "$REPORT_FILE"
    
    # Test 5: Start Process (Write Operation)
    echo -e "${YELLOW}Test 5: Start Process (Write Operation)${NC}"
    log_result ""
    log_result "Test 5: Start Process (Write Operation)"
    log_result "----------------------------------------"
    
    # Create process payload
    PROCESS_PAYLOAD='{"workflow_id":"purchase_approval","initiated_by":"admin","data":{"amount":5000,"description":"Load test purchase","department":"IT"}}'
    echo "$PROCESS_PAYLOAD" > "$RESULTS_DIR/process-payload.json"
    
    # Reduce concurrent users for write operations to avoid overwhelming the system
    WRITE_CONCURRENT=$((CONCURRENT_USERS / 2))
    WRITE_REQUESTS=$((TOTAL_REQUESTS / 5))
    
    ab -n "$WRITE_REQUESTS" -c "$WRITE_CONCURRENT" \
        -p "$RESULTS_DIR/process-payload.json" \
        -T "application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -g "$RESULTS_DIR/process-$TIMESTAMP.tsv" \
        "$API_BASE/process/start" 2>&1 | tee -a "$REPORT_FILE"
fi

# Summary
echo ""
echo -e "${GREEN}=================================="
echo "Load Test Completed"
echo "==================================${NC}"
log_result ""
log_result "=================================="
log_result "Load Test Completed"
log_result "=================================="
log_result "Results saved to: $REPORT_FILE"
log_result "TSV data files saved to: $RESULTS_DIR/"

# Calculate summary statistics
echo -e "${BLUE}Summary Statistics:${NC}"
log_result ""
log_result "Summary Statistics:"

if [ -f "$RESULTS_DIR/health-$TIMESTAMP.tsv" ]; then
    HEALTH_AVG=$(awk '{sum+=$5; count++} END {print sum/count}' "$RESULTS_DIR/health-$TIMESTAMP.tsv" 2>/dev/null || echo "N/A")
    echo -e "Health Check Avg Response Time: ${HEALTH_AVG}ms"
    log_result "Health Check Avg Response Time: ${HEALTH_AVG}ms"
fi

if [ -f "$RESULTS_DIR/login-$TIMESTAMP.tsv" ]; then
    LOGIN_AVG=$(awk '{sum+=$5; count++} END {print sum/count}' "$RESULTS_DIR/login-$TIMESTAMP.tsv" 2>/dev/null || echo "N/A")
    echo -e "Login Avg Response Time: ${LOGIN_AVG}ms"
    log_result "Login Avg Response Time: ${LOGIN_AVG}ms"
fi

if [ -f "$RESULTS_DIR/workflows-$TIMESTAMP.tsv" ]; then
    WORKFLOW_AVG=$(awk '{sum+=$5; count++} END {print sum/count}' "$RESULTS_DIR/workflows-$TIMESTAMP.tsv" 2>/dev/null || echo "N/A")
    echo -e "Workflows Avg Response Time: ${WORKFLOW_AVG}ms"
    log_result "Workflows Avg Response Time: ${WORKFLOW_AVG}ms"
fi

echo ""
echo -e "${GREEN}✓ Load test completed successfully${NC}"
echo -e "View detailed results: cat $REPORT_FILE"

# Cleanup temp files
rm -f "$RESULTS_DIR/login-payload.json" "$RESULTS_DIR/process-payload.json"
