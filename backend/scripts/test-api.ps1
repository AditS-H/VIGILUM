# VIGILUM Backend API Test Script
# Tests all major endpoints with various scenarios

$BaseUrl = "http://localhost:8080/api/v1"
$UserId = "test-user-$(Get-Date -UFormat '%s')"
$VerifierAddress = "0x1234567890123456789012345678901234567890"

Write-Host "================================"
Write-Host "VIGILUM Backend API Test Suite"
Write-Host "================================"
Write-Host "Base URL: $BaseUrl"
Write-Host "User ID: $UserId"
Write-Host ""

# Test 1: Health Check
Write-Host "[TEST 1] Health Check"
Write-Host "GET /health"
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/health" -Method Get
    $response | ConvertTo-Json
} catch {
    Write-Host "FAILED: $_"
}
Write-Host ""

# Test 2: Generate Challenge
Write-Host "[TEST 2] Generate Challenge"
Write-Host "POST /proofs/challenges"
$body = @{
    user_id = $UserId
    verifier_address = $VerifierAddress
} | ConvertTo-Json

try {
    $challengeResponse = Invoke-RestMethod -Uri "$BaseUrl/proofs/challenges" `
        -Method Post `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body
    $challengeResponse | ConvertTo-Json
    $challengeId = $challengeResponse.challenge_id
    Write-Host "Challenge ID: $challengeId"
} catch {
    Write-Host "FAILED: $_"
    $challengeId = $null
}
Write-Host ""

# Test 3: Get Challenge Status
if ($challengeId) {
    Write-Host "[TEST 3] Get Challenge Status"
    Write-Host "GET /proofs/challenges/$challengeId"
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/proofs/challenges/$challengeId" -Method Get
        $response | ConvertTo-Json
    } catch {
        Write-Host "FAILED: $_"
    }
    Write-Host ""
}

# Test 4: Submit Proof
if ($challengeId) {
    Write-Host "[TEST 4] Submit Proof"
    Write-Host "POST /proofs/verify"
    $proofData = "0x" + ([byte[]]::new(32) | ForEach-Object { "{0:x2}" -f $_ } | Join-String)
    $proofBody = @{
        challenge_id = $challengeId
        proof_data = $proofData
        timing_variance = 50
        gas_variance = 2000
        proof_nonce = "nonce_$(Get-Date -UFormat '%s')"
    } | ConvertTo-Json

    try {
        $proofResponse = Invoke-RestMethod -Uri "$BaseUrl/proofs/verify" `
            -Method Post `
            -Headers @{"Content-Type" = "application/json"} `
            -Body $proofBody
        $proofResponse | ConvertTo-Json
    } catch {
        Write-Host "FAILED: $_"
    }
    Write-Host ""
}

# Test 5: Get User Proofs with Pagination
Write-Host "[TEST 5] Get User Proofs (Pagination)"
Write-Host "GET /proofs?user_id=$UserId&page=1&limit=10"
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/proofs?user_id=$UserId&page=1&limit=10" -Method Get
    $response | ConvertTo-Json
} catch {
    Write-Host "FAILED: $_"
}
Write-Host ""

# Test 6: Get Verification Score
Write-Host "[TEST 6] Get User Verification Score"
Write-Host "GET /users/verification-score?user_id=$UserId"
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/users/verification-score?user_id=$UserId" -Method Get
    $response | ConvertTo-Json
} catch {
    Write-Host "FAILED: $_"
}
Write-Host ""

# Test 7: Firewall Endpoints
Write-Host "[TEST 7] Firewall - Verify Proof (Alias)"
Write-Host "POST /firewall/verify-proof"
$firewallProofData = "0x" + ([byte[]]::new(32) | ForEach-Object { "{0:x2}" -f $_ } | Join-String)
$fwBody = @{
    challenge_id = "challenge_test_$(Get-Date -UFormat '%s')"
    proof_data = $firewallProofData
    timing_variance = 30
    gas_variance = 1500
    proof_nonce = "fw_nonce_$(Get-Date -UFormat '%s')"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/firewall/verify-proof" `
        -Method Post `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $fwBody
    $response | ConvertTo-Json
} catch {
    Write-Host "FAILED: $_"
}
Write-Host ""

# Test 8: CORS Headers
Write-Host "[TEST 8] CORS Preflight Check"
Write-Host "OPTIONS /proofs/challenges"
try {
    $response = Invoke-WebRequest -Uri "$BaseUrl/proofs/challenges" `
        -Method Options `
        -Headers @{
            "Origin" = "http://localhost:5173"
            "Access-Control-Request-Method" = "POST"
        } `
        -SkipHttpErrorCheck
    Write-Host "Response Headers:"
    $response.Headers.GetEnumerator() | Where-Object { $_.Key -like "*Access-Control*" } | ForEach-Object {
        Write-Host "$($_.Key): $($_.Value)"
    }
} catch {
    Write-Host "FAILED: $_"
}
Write-Host ""

Write-Host "================================"
Write-Host "Test Suite Complete"
Write-Host "================================"
Write-Host ""
Write-Host "Summary:"
Write-Host "- Test 1: Health check endpoint"
Write-Host "- Test 2: Challenge generation"
Write-Host "- Test 3: Challenge status retrieval"
Write-Host "- Test 4: Proof submission and verification"
Write-Host "- Test 5: Paginated proof listing"
Write-Host "- Test 6: Verification score retrieval"
Write-Host "- Test 7: Firewall endpoint"
Write-Host "- Test 8: CORS headers validation"
