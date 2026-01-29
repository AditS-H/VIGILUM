# VIGILUM Backend API - Manual Testing Guide
# Use these curl commands to test the API manually

# Environment
BASE_URL="http://localhost:8080/api/v1"
USER_ID="test-user-manual-001"
VERIFIER_ADDRESS="0x1234567890123456789012345678901234567890"

## ============================================
## 1. HEALTH CHECK
## ============================================

curl -X GET "$BASE_URL/health"


## ============================================
## 2. GENERATE PROOF CHALLENGE
## ============================================

# Save challenge response to use in next tests
curl -X POST "$BASE_URL/proofs/challenges" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'"$USER_ID"'",
    "verifier_address": "'"$VERIFIER_ADDRESS"'"
  }' | jq .

# Expected response:
# {
#   "challenge_id": "challenge_...",
#   "challenge_data": "0x...",
#   "issued_at": "2024-01-28T...",
#   "expires_at": "2024-01-28T..."
# }


## ============================================
## 3. GET CHALLENGE STATUS
## ============================================

# Replace CHALLENGE_ID with the value from step 2
curl -X GET "$BASE_URL/proofs/challenges/{CHALLENGE_ID}"


## ============================================
## 4. SUBMIT PROOF FOR VERIFICATION
## ============================================

# Submit proof for the challenge from step 2
curl -X POST "$BASE_URL/proofs/verify" \
  -H "Content-Type: application/json" \
  -d '{
    "challenge_id": "{CHALLENGE_ID}",
    "proof_data": "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
    "timing_variance": 50,
    "gas_variance": 2000,
    "proof_nonce": "test_nonce_'"$(date +%s)"'"
  }' | jq .

# Expected response:
# {
#   "proof_id": "proof_...",
#   "verification_score": 0.85,
#   "is_valid": true,
#   "verified_at": "2024-01-28T..."
# }


## ============================================
## 5. GET USER PROOFS (WITH PAGINATION)
## ============================================

# Get paginated list of user's proofs
curl -X GET "$BASE_URL/proofs?user_id=$USER_ID&page=1&limit=10" | jq .

# Expected response:
# {
#   "proofs": [
#     {
#       "id": "proof_...",
#       "user_id": "user_...",
#       "verified": true,
#       "verification_score": 0.85,
#       "created_at": "2024-01-28T...",
#       "verified_at": "2024-01-28T..."
#     }
#   ],
#   "pagination": {
#     "page": 1,
#     "limit": 10,
#     "total": 5,
#     "total_pages": 1
#   }
# }


## ============================================
## 6. GET USER VERIFICATION SCORE
## ============================================

curl -X GET "$BASE_URL/users/verification-score?user_id=$USER_ID" | jq .

# Expected response:
# {
#   "user_id": "user_...",
#   "verification_score": 0.85,
#   "is_verified": true,
#   "risk_score": 25,
#   "last_verified_at": "2024-01-28T...",
#   "proof_count": 5,
#   "verified_proof_count": 4
# }


## ============================================
## 7. FIREWALL VERIFY PROOF (ALIAS)
## ============================================

# Alternative endpoint for proof verification
curl -X POST "$BASE_URL/firewall/verify-proof" \
  -H "Content-Type: application/json" \
  -d '{
    "challenge_id": "{CHALLENGE_ID}",
    "proof_data": "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
    "timing_variance": 40,
    "gas_variance": 1800,
    "proof_nonce": "firewall_nonce_'"$(date +%s)"'"
  }' | jq .


## ============================================
## 8. FIREWALL GET CHALLENGE
## ============================================

curl -X GET "$BASE_URL/firewall/challenge?user_id=$USER_ID"


## ============================================
## 9. TEST CORS HEADERS
## ============================================

# Check if CORS headers are present
curl -X OPTIONS "$BASE_URL/proofs/challenges" \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST" \
  -v

# Expected headers in response:
# Access-Control-Allow-Origin: *
# Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT, DELETE
# Access-Control-Allow-Headers: Content-Type, ...


## ============================================
## TEST SCENARIOS
## ============================================

# Scenario 1: Complete workflow
# 1. Generate challenge
# 2. Submit proof
# 3. Get verification score
# 4. Generate another challenge (pagination test)
# 5. Submit another proof (same user)
# 6. Verify pagination returns correct counts

# Scenario 2: Invalid inputs
# - Submit proof with invalid challenge ID
# - Submit proof with malformed proof data (not valid hex)
# - Get score for non-existent user
# - Get proofs with invalid pagination params

# Scenario 3: Pagination
# - Create multiple proofs (5+)
# - Get page 1 with limit 2
# - Verify total_pages calculation
# - Get page 2
# - Get with limit larger than total

# Scenario 4: Risk scoring
# - Create user with multiple verified proofs
# - Verify risk_score calculation
# - Check verification_score reflects verified ratio

# Scenario 5: Timestamp tracking
# - Create proofs at different times
# - Verify last_verified_at is most recent
# - Verify created_at timestamps are preserved
