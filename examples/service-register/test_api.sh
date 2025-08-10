#!/bin/bash

# A simple script to test the API endpoints of the service-register example.
#
# Prerequisites:
# 1. The service-register example must be running.
# 2. httpie must be installed (e.g., `brew install httpie` or `pip install httpie`).

# --- Configuration ---
BASE_URL="http://localhost:3001"
TEST_ID="test-$(date +%s)" # Unique ID for the test data

# --- Helper Functions ---

echo_blue() {
    echo -e "\033[0;34m$1\033[0m"
}

# --- Test Execution ---

echo_blue "▶️ Testing Health Check endpoint..."
http GET ${BASE_URL}/health
sleep 1

echo_blue "▶️ Testing Service Discovery Status endpoint..."
http GET ${BASE_URL}/api/v1/service-discovery/status
sleep 1

echo_blue "▶️ Testing Redis Service Discovery endpoint..."
http GET ${BASE_URL}/api/v1/services/redis
sleep 1

echo_blue "▶️ Testing Redis Write endpoint (POST)..."
http POST ${BASE_URL}/api/v1/redis-test \
    id="${TEST_ID}" \
    message="Hello from httpie" \
    timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
sleep 1

echo_blue "▶️ Testing Redis Read endpoint (GET)..."
http GET ${BASE_URL}/api/v1/redis-test/${TEST_ID}

echo_blue "✅ All tests completed."
