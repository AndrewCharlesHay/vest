#!/bin/bash

if [ -z "$API_URL" ]; then
  BASE_URL="http://localhost:8080"
else
  BASE_URL="$API_URL"
fi
DATE="2025-01-15"

echo "Checking Health..."
curl -s -f "$BASE_URL/health" || echo "Health check failed"

# Auth Header
if [ -n "$API_KEY" ]; then
  AUTH_HEADER="-H \"X-API-Key: $API_KEY\""
else
  AUTH_HEADER=""
fi

echo "\n\nChecking Blotter..."
eval curl -s $AUTH_HEADER "$BASE_URL/blotter?date=$DATE" | jq .

echo "\n\nChecking Positions..."
eval curl -s $AUTH_HEADER "$BASE_URL/positions?date=$DATE" | jq .

echo "\n\nChecking Alarms..."
eval curl -s $AUTH_HEADER "$BASE_URL/alarms?date=$DATE" | jq .
