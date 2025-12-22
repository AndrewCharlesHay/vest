#!/bin/bash

if [ -z "$API_URL" ]; then
  BASE_URL="http://localhost:8080"
else
  BASE_URL="$API_URL"
fi
DATE="2025-01-15"

echo "Checking Health..."
curl -s -f "$BASE_URL/health" || echo "Health check failed"

echo "\n\nChecking Blotter..."
curl -s "$BASE_URL/blotter?date=$DATE" | jq .

echo "\n\nChecking Positions..."
curl -s "$BASE_URL/positions?date=$DATE" | jq .

echo "\n\nChecking Alarms..."
curl -s "$BASE_URL/alarms?date=$DATE" | jq .
