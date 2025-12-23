#!/bin/bash
set -e

if [ -z "$API_URL" ]; then
  BASE_URL="http://localhost:8080"
else
  BASE_URL="$API_URL"
fi
DATE="2025-01-15"

# Auth Header
if [ -n "$API_KEY" ]; then
  AUTH_HEADER="-H X-API-Key:$API_KEY"
else
  AUTH_HEADER=""
fi

check_endpoint() {
  local endpoint=$1
  local expected_jq=$2
  
  echo "Testing $endpoint..."
  
  # Capture HTTP Status and Body
  response=$(curl -s -w "\n%{http_code}" $AUTH_HEADER "$BASE_URL$endpoint")
  http_code=$(echo "$response" | tail -n1)
  body=$(echo "$response" | sed '$d')

  if [ "$http_code" -ne 200 ]; then
    echo "❌ Failed: $endpoint returned $http_code"
    echo "Body: $body"
    exit 1
  fi

  # Validate JSON content
  if [ -n "$expected_jq" ]; then
    if echo "$body" | jq -e "$expected_jq" >/dev/null; then
       echo "✅ Verified: $endpoint matches expectation."
    else
       echo "❌ Failed: $endpoint JSON validation error."
       echo "Expected: $expected_jq"
       echo "Got: $body"
       exit 1
    fi
  else
    echo "✅ Verified: $endpoint returned 200 OK."
  fi
}

echo "Starting Robust Smoke Tests against $BASE_URL"

# 1. Health Check
check_endpoint "/health" '.status == "ok"'

# 2. Blotter (Expect non-empty array)
check_endpoint "/blotter?date=$DATE" 'length > 0'

# 3. Positions (Expect non-empty array)
check_endpoint "/positions?date=$DATE" 'length > 0'

# 4. Alarms (Expect array, check structure if populated)
check_endpoint "/alarms?date=$DATE" 'type == "array"'

echo "🎉 All Smoke Tests Passed!"
