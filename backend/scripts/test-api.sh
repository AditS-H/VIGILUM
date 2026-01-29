#!/bin/bash
# VIGILUM Backend API Test Script
# Tests all major endpoints with various scenarios

BASE_URL="http://localhost:8080/api/v1"
USER_ID="test-user-$(date +%s)"
VERIFIER_ADDRESS="0x1234567890123456789012345678901234567890"

echo "================================"
echo "VIGILUM Backend API Test Suite"
echo "================================"
echo "Base URL: $BASE_URL"
echo "User ID: $USER_ID"
echo ""

# Test 1: Health Check
echo "[TEST 1] Health Check"
echo "GET /health"
curl -s -X GET "$BASE_URL/health" | jq . || echo "FAILED"
echo ""

# Test 2: Generate Challenge
echo "[TEST 2] Generate Challenge"
echo "POST /proofs/challenges"
CHALLENGE_RESPONSE=$(curl -s -X POST "$BASE_URL/proofs/challenges" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"verifier_address\": \"$VERIFIER_ADDRESS\"
  }")
echo "$CHALLENGE_RESPONSE" | jq .
CHALLENGE_ID=$(echo "$CHALLENGE_RESPONSE" | jq -r '.challenge_id // empty')
echo "Challenge ID: $CHALLENGE_ID"
echo ""

# Test 3: Get Challenge Status
if [ -n "$CHALLENGE_ID" ]; then
  echo "[TEST 3] Get Challenge Status"
  echo "GET /proofs/challenges/$CHALLENGE_ID"
  curl -s -X GET "$BASE_URL/proofs/challenges/$CHALLENGE_ID" | jq .
  echo ""
fi

# Test 4: Submit Proof
if [ -n "$CHALLENGE_ID" ]; then
  echo "[TEST 4] Submit Proof"
  echo "POST /proofs/verify"
  PROOF_RESPONSE=$(curl -s -X POST "$BASE_URL/proofs/verify" \
    -H "Content-Type: application/json" \
    -d "{
      \"challenge_id\": \"$CHALLENGE_ID\",
      \"proof_data\": \"0x$(xxd -p -l 32 /dev/zero | tr -d '\n')\",
      \"timing_variance\": 50,
      \"gas_variance\": 2000,
      \"proof_nonce\": \"nonce_$(date +%s)\"
    }")
  echo "$PROOF_RESPONSE" | jq .
  PROOF_ID=$(echo "$PROOF_RESPONSE" | jq -r '.proof_id // empty')
  echo ""
fi

# Test 5: Get User Proofs with Pagination
echo "[TEST 5] Get User Proofs (Pagination)"
echo "GET /proofs?user_id=$USER_ID&page=1&limit=10"
PROOFS_RESPONSE=$(curl -s -X GET "$BASE_URL/proofs?user_id=$USER_ID&page=1&limit=10")
echo "$PROOFS_RESPONSE" | jq .
echo ""

# Test 6: Get Verification Score
echo "[TEST 6] Get User Verification Score"
echo "GET /users/verification-score?user_id=$USER_ID"
SCORE_RESPONSE=$(curl -s -X GET "$BASE_URL/users/verification-score?user_id=$USER_ID")
echo "$SCORE_RESPONSE" | jq .
echo ""

# Test 7: Firewall Endpoints
echo "[TEST 7] Firewall - Verify Proof (Alias)"
echo "POST /firewall/verify-proof"
curl -s -X POST "$BASE_URL/firewall/verify-proof" \
  -H "Content-Type: application/json" \
  -d "{
    \"challenge_id\": \"challenge_test_$(date +%s)\",
    \"proof_data\": \"0x$(xxd -p -l 32 /dev/zero | tr -d '\n')\",
    \"timing_variance\": 30,
    \"gas_variance\": 1500,
    \"proof_nonce\": \"fw_nonce_$(date +%s)\"
  }" | jq .
echo ""

# Test 8: CORS Preflight
echo "[TEST 8] CORS Preflight Check"
echo "OPTIONS /proofs/challenges"
curl -s -X OPTIONS "$BASE_URL/proofs/challenges" \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST" \
  -v 2>&1 | grep -i "access-control"
echo ""

echo "================================"
echo "Test Suite Complete"
echo "================================"
echo ""
echo "Summary:"
echo "- Test 1: Health check endpoint"
echo "- Test 2: Challenge generation"
echo "- Test 3: Challenge status retrieval"
echo "- Test 4: Proof submission and verification"
echo "- Test 5: Paginated proof listing"
echo "- Test 6: Verification score retrieval"
echo "- Test 7: Firewall endpoint"
echo "- Test 8: CORS headers validation"
