#!/bin/bash

# Data Collector API Test Script
# This script tests all API endpoints

set -e

API_BASE="http://localhost:8080/api/v1"
CONNECTOR_ID=""
JOB_ID=""

echo "================================"
echo "Data Collector API Test Suite"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function to print test results
test_result() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $1"
    else
        echo -e "${RED}✗ FAIL${NC}: $1"
        exit 1
    fi
}

echo -e "${BLUE}1. Testing Health Endpoint${NC}"
echo "----------------------------"
curl -s "$API_BASE/health" | jq .
test_result "Health check"
echo ""

echo -e "${BLUE}2. Creating Connector (Sandbox Mode)${NC}"
echo "-------------------------------------"
RESPONSE=$(curl -s -X POST "$API_BASE/connectors" \
  -H "Content-Type: application/json" \
  -d '{
    "exchange_id": "binance",
    "display_name": "Binance Testnet",
    "sandbox_mode": true,
    "rate_limit": {
      "limit": 1200,
      "period_ms": 60000
    }
  }')

echo "$RESPONSE" | jq .
CONNECTOR_ID=$(echo "$RESPONSE" | jq -r '.id')
test_result "Create connector (sandbox=true)"
echo "Connector ID: $CONNECTOR_ID"
echo ""

echo -e "${BLUE}3. Getting All Connectors${NC}"
echo "--------------------------"
curl -s "$API_BASE/connectors" | jq .
test_result "Get all connectors"
echo ""

echo -e "${BLUE}4. Getting Connector by ID${NC}"
echo "---------------------------"
curl -s "$API_BASE/connectors/$CONNECTOR_ID" | jq .
test_result "Get connector by ID"
echo ""

echo -e "${BLUE}5. Filtering Connectors (Sandbox Only)${NC}"
echo "---------------------------------------"
curl -s "$API_BASE/connectors?sandbox_mode=true" | jq .
test_result "Filter connectors by sandbox mode"
echo ""

echo -e "${BLUE}6. Toggle Sandbox Mode OFF${NC}"
echo "----------------------------"
curl -s -X PATCH "$API_BASE/connectors/$CONNECTOR_ID/sandbox" \
  -H "Content-Type: application/json" \
  -d '{"sandbox_mode": false}' | jq .
test_result "Toggle sandbox mode to false"
echo ""

echo -e "${BLUE}7. Toggle Sandbox Mode ON${NC}"
echo "---------------------------"
curl -s -X PATCH "$API_BASE/connectors/$CONNECTOR_ID/sandbox" \
  -H "Content-Type: application/json" \
  -d '{"sandbox_mode": true}' | jq .
test_result "Toggle sandbox mode to true"
echo ""

echo -e "${BLUE}8. Updating Connector${NC}"
echo "----------------------"
curl -s -X PUT "$API_BASE/connectors/$CONNECTOR_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Binance Updated",
    "rate_limit": {
      "limit": 2400
    }
  }' | jq .
test_result "Update connector"
echo ""

echo -e "${BLUE}9. Creating Job (BTC/USDT 1h)${NC}"
echo "------------------------------"
RESPONSE=$(curl -s -X POST "$API_BASE/jobs" \
  -H "Content-Type: application/json" \
  -d '{
    "connector_exchange_id": "binance",
    "symbol": "BTC/USDT",
    "timeframe": "1h",
    "status": "active"
  }')

echo "$RESPONSE" | jq .
JOB_ID=$(echo "$RESPONSE" | jq -r '.id')
test_result "Create job BTC/USDT 1h"
echo "Job ID: $JOB_ID"
echo ""

echo -e "${BLUE}10. Creating Job (ETH/USDT 15m)${NC}"
echo "--------------------------------"
curl -s -X POST "$API_BASE/jobs" \
  -H "Content-Type: application/json" \
  -d '{
    "connector_exchange_id": "binance",
    "symbol": "ETH/USDT",
    "timeframe": "15m",
    "status": "active"
  }' | jq .
test_result "Create job ETH/USDT 15m"
echo ""

echo -e "${BLUE}11. Getting All Jobs${NC}"
echo "--------------------"
curl -s "$API_BASE/jobs" | jq .
test_result "Get all jobs"
echo ""

echo -e "${BLUE}12. Getting Job by ID${NC}"
echo "---------------------"
curl -s "$API_BASE/jobs/$JOB_ID" | jq .
test_result "Get job by ID"
echo ""

echo -e "${BLUE}13. Filtering Jobs (by Exchange)${NC}"
echo "---------------------------------"
curl -s "$API_BASE/jobs?exchange_id=binance" | jq .
test_result "Filter jobs by exchange"
echo ""

echo -e "${BLUE}14. Filtering Jobs (by Symbol)${NC}"
echo "-------------------------------"
curl -s "$API_BASE/jobs?symbol=BTC/USDT" | jq .
test_result "Filter jobs by symbol"
echo ""

echo -e "${BLUE}15. Getting Jobs for Connector${NC}"
echo "-----------------------------------"
curl -s "$API_BASE/connectors/binance/jobs" | jq .
test_result "Get jobs for connector"
echo ""

echo -e "${BLUE}16. Pausing Job${NC}"
echo "----------------"
curl -s -X POST "$API_BASE/jobs/$JOB_ID/pause" | jq .
test_result "Pause job"
echo ""

echo -e "${BLUE}17. Resuming Job${NC}"
echo "-----------------"
curl -s -X POST "$API_BASE/jobs/$JOB_ID/resume" | jq .
test_result "Resume job"
echo ""

echo -e "${BLUE}18. Updating Job${NC}"
echo "-----------------"
curl -s -X PUT "$API_BASE/jobs/$JOB_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "timeframe": "4h"
  }' | jq .
test_result "Update job timeframe"
echo ""

echo -e "${BLUE}19. Deleting Job${NC}"
echo "-----------------"
curl -s -X DELETE "$API_BASE/jobs/$JOB_ID" -w "\nHTTP Status: %{http_code}\n"
test_result "Delete job"
echo ""

echo -e "${BLUE}20. Deleting Connector${NC}"
echo "-----------------------"
curl -s -X DELETE "$API_BASE/connectors/$CONNECTOR_ID" -w "\nHTTP Status: %{http_code}\n"
test_result "Delete connector"
echo ""

echo ""
echo "================================"
echo -e "${GREEN}All Tests Passed! ✓${NC}"
echo "================================"
