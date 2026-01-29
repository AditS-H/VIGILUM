# VIGILUM Platform Documentation

## Table of Contents

1. [Quick Start](#quick-start)
2. [Architecture Overview](#architecture-overview)
3. [API Reference](#api-reference)
4. [Smart Contracts](#smart-contracts)
5. [Deployment Guide](#deployment-guide)
6. [Development Guide](#development-guide)
7. [Troubleshooting](#troubleshooting)
8. [Security Considerations](#security-considerations)

---

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+
- Python 3.12+
- Rust 1.70+
- Docker & Docker Compose
- PostgreSQL 15+

### Development Setup (5 minutes)

```bash
# Clone repository
git clone https://github.com/vigilum/vigilum.git
cd vigilum

# Backend setup
cd backend
go mod download
CGO_ENABLED=0 go build ./cmd/api

# Frontend setup
cd ../sdk/ts-sdk/demo
npm install
npm run dev  # http://localhost:5173

# ML pipeline setup
cd ../../..
python -m venv venv
source venv/bin/activate
pip install -e ml/

# Backend in another terminal
cd backend
VIGILUM_ENV=development go run ./cmd/api/main.go
```

Access:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8000
- **API Docs**: http://localhost:8000/swagger

---

## Architecture Overview

### System Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (React)                         │
│  - Proof submission UI                                      │
│  - Contract analysis dashboard                              │
│  - Risk scoring visualization                               │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTP/WebSocket
┌──────────────────────▼──────────────────────────────────────┐
│                   API Gateway (Go)                           │
│  - Rate limiting                                             │
│  - Authentication                                           │
│  - Request aggregation                                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
        ▼              ▼              ▼
   ┌─────────┐  ┌─────────┐  ┌─────────┐
   │ Scanner │  │ Firewall│  │  Oracle │
   │ Service │  │ Service │  │ Service │
   └────┬────┘  └────┬────┘  └────┬────┘
        │             │             │
        └─────────────┼─────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
        ▼             ▼             ▼
   ┌─────────┐  ┌─────────┐  ┌─────────┐
   │   WASM  │  │ Temporal│  │  ML     │
   │ Verifier│  │ Workflows│ │ Inference
   └────┬────┘  └────┬────┘  └────┬────┘
        │             │             │
        └─────────────┼─────────────┘
                      │
        ┌─────────────┼──────────────────────┐
        │             │                      │
        ▼             ▼                      ▼
   ┌──────────┐  ┌──────────┐         ┌───────────┐
   │PostgreSQL│  │  Ethereum │         │  Jae ger  │
   │ Database │  │   Node    │         │ Tracing   │
   └──────────┘  └──────────┘         └───────────┘
```

### Service Responsibilities

| Service | Purpose | Tech Stack |
|---------|---------|-----------|
| **API Gateway** | HTTP routing, rate limiting | Go, Gin |
| **Scanner** | Contract bytecode analysis | Go, ML integration |
| **Firewall** | Proof verification, access control | Go, Zero-Knowledge |
| **Oracle** | Threat intelligence aggregation | Go, Feed sources |
| **ML Pipeline** | Risk scoring models | Python, PyTorch |
| **Temporal** | Async workflow orchestration | Temporal SDK |
| **WASM Verifier** | Zero-knowledge proof verification | Rust, Noir circuits |

---

## API Reference

### Authentication

All API endpoints require wallet signature authentication.

```bash
# Sign challenge with wallet
challenge=$(curl -s http://localhost:8000/auth/challenge | jq -r .challenge)
signature=$(cast sig "$(echo -n $challenge | tr -d '\n')")

# Include in requests
curl -H "Authorization: Bearer $signature" \
     -H "X-Challenge: $challenge" \
     http://localhost:8000/api/proofs
```

### Proof Submission

#### `POST /api/proofs/submit`

Submit a proof for contract analysis.

**Request:**
```json
{
  "contract_address": "0x1234567890123456789012345678901234567890",
  "proof_data": "0x...",
  "proof_type": "human" | "exploit",
  "description": "Description of proof"
}
```

**Response:**
```json
{
  "proof_id": "0x...",
  "status": "pending",
  "created_at": "2025-01-15T10:00:00Z",
  "estimated_verification_time": 300
}
```

#### `GET /api/proofs/:proof_id`

Get proof status and verification results.

**Response:**
```json
{
  "proof_id": "0x...",
  "status": "verified",
  "contract_address": "0x1234...",
  "risk_score": 75.0,
  "verified_at": "2025-01-15T10:05:00Z",
  "on_chain_tx": "0xabcd..."
}
```

### Risk Scoring

#### `GET /api/contracts/:address/risk`

Get current risk score for a contract.

**Response:**
```json
{
  "contract_address": "0x1234...",
  "risk_score": 75.0,
  "risk_level": "HIGH",
  "factors": {
    "code_complexity": 0.8,
    "vulnerability_patterns": 0.6,
    "historical_exploits": 0.9
  },
  "last_update": "2025-01-15T10:00:00Z"
}
```

#### `GET /api/contracts/:address/blacklist-status`

Check if contract is blacklisted.

**Response:**
```json
{
  "contract_address": "0x1234...",
  "is_blacklisted": true,
  "reason": "Critical vulnerability detected",
  "blacklisted_at": "2025-01-15T09:00:00Z",
  "blacklist_tx": "0xdeadbeef..."
}
```

### Health & Metrics

#### `GET /health`

System health status.

**Response:**
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "database": "connected",
  "ethereum": "connected",
  "temporal": "connected",
  "timestamp": "2025-01-15T10:00:00Z"
}
```

#### `GET /metrics`

Prometheus metrics (Grafana compatible).

```
# HELP vigilum_proofs_verified_total Total proofs verified
# TYPE vigilum_proofs_verified_total counter
vigilum_proofs_verified_total{type="human"} 1234
vigilum_proofs_verified_total{type="exploit"} 567

# HELP vigilum_proof_verification_duration_seconds Proof verification duration
# TYPE vigilum_proof_verification_duration_seconds histogram
vigilum_proof_verification_duration_seconds_bucket{le="0.1"} 500
```

---

## Smart Contracts

### VigilumRegistry

Main registry for risk scores and blacklisting.

```solidity
// Submit a contract for analysis
registerContract(address contractAddr, bytes32 bytecodeHash, uint16 riskScore)

// Update risk score
updateRiskScore(address contractAddr, uint16 newScore, uint32 vulnerabilityCount)

// Query risk
getRiskScore(address contractAddr) -> uint16

// Check blacklist
isBlacklisted(address contractAddr) -> bool
```

### RedTeamDAO

Decentralized autonomous organization for security researchers.

```solidity
// Join DAO (reputation starts at 0)
joinDAO()

// Create a proposal
createProposal(string title, string description) -> uint256

// Vote on proposal
vote(uint256 proposalId, bool support)

// Record discovered exploit
recordExploit(address targetContract, string description, uint256 severity, uint256 bounty) -> bytes32

// Get member reputation
getReputation(address member) -> uint256
```

### ProofOfExploit

Smart contract for exploit proof submission and verification.

```solidity
// Submit exploit proof
submitProof(
  address targetContract,
  uint256 severity,
  string description,
  bytes proofData
) -> bytes32

// Verify proof (requires 3+ verifiers)
verifyProof(bytes32 proofId, bool approved, string notes)

// Claim reward (after verification approval)
claimReward(bytes32 proofId)

// Get contract vulnerability statistics
getContractStats(address targetContract) -> (uint256 total, uint256 verified, uint256[] severities)
```

---

## Deployment Guide

### Sepolia Testnet Deployment

```bash
# 1. Set environment variables
export RPC_URL="https://sepolia.infura.io/v3/YOUR_INFURA_KEY"
export PRIVATE_KEY="0x..."
export ETHERSCAN_API_KEY="YOUR_ETHERSCAN_KEY"

# 2. Deploy VigilumRegistry
cd contracts
forge script script/DeployVigilumRegistry.s.sol:DeployVigilumRegistry \
  --rpc-url $RPC_URL \
  --private-key $PRIVATE_KEY \
  --broadcast \
  --verify \
  --etherscan-api-key $ETHERSCAN_API_KEY

# 3. Deploy DAO contracts
forge script script/DeployDAOContracts.s.sol:DeployDAOContracts \
  --rpc-url $RPC_URL \
  --private-key $PRIVATE_KEY \
  --broadcast

# 4. Update backend with contract addresses
export VIGILUM_REGISTRY="0x..."
export REDTEAM_DAO="0x..."
export PROOF_OF_EXPLOIT="0x..."
```

### Kubernetes Deployment

```bash
# 1. Create namespace and secrets
kubectl apply -f infra/k8s/config/namespace.yaml

kubectl create secret generic vigilum-secrets \
  --from-literal=eth-rpc-url=$ETH_RPC_URL \
  --from-literal=private-key=$PRIVATE_KEY \
  --from-literal=database-url=$DATABASE_URL \
  -n vigilum

# 2. Deploy services
kubectl apply -f infra/k8s/backend/
kubectl apply -f infra/k8s/frontend/
kubectl apply -f infra/k8s/database/
kubectl apply -f infra/k8s/temporal/

# 3. Verify rollout
kubectl rollout status deployment/vigilum-backend -n vigilum
kubectl rollout status deployment/vigilum-frontend -n vigilum

# 4. Check endpoints
kubectl get endpoints -n vigilum
kubectl port-forward svc/vigilum-backend 8000:8000 -n vigilum
```

---

## Development Guide

### Adding a New Service

1. **Create service package**:
   ```bash
   mkdir backend/internal/myservice
   touch backend/internal/myservice/service.go
   ```

2. **Implement service interface**:
   ```go
   package myservice

   type Service struct {
       logger *slog.Logger
       // dependencies
   }

   func New(logger *slog.Logger) *Service {
       return &Service{logger: logger}
   }
   ```

3. **Register with API**:
   ```go
   // In backend/cmd/api/main.go
   myService := myservice.New(logger)
   router.POST("/api/myservice/endpoint", myService.HandleRequest)
   ```

4. **Add tests**:
   ```bash
   touch backend/internal/myservice/service_test.go
   ```

### Running Tests

```bash
# Backend
cd backend
go test ./... -v -race

# Frontend
cd sdk/ts-sdk
npm test

# Contracts
cd contracts
forge test -vvv

# ML Pipeline
cd ml
pytest -v tests/
```

### Code Style

- **Go**: golangci-lint, goimports
- **TypeScript**: ESLint, Prettier
- **Solidity**: Solhint, Slither
- **Python**: Black, Ruff

---

## Troubleshooting

### Backend won't start

```bash
# Check database connection
psql $DATABASE_URL -c "SELECT 1"

# Check Ethereum RPC
cast call 0x0000000000000000000000000000000000000000 --rpc-url $ETH_RPC_URL

# Check Temporal
temporal operator namespace describe -n default
```

### Frontend build fails

```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install

# Check Node version
node --version  # Should be 18+

# Try build again
npm run build
```

### Proof verification timeout

```bash
# Check WASM module loading
curl http://localhost:8000/health

# Verify proof circuit compilation
cd circuits && nargo compile

# Check available resources
ps aux | grep api  # CPU/memory usage
```

---

## Security Considerations

⚠️ **Never commit private keys**
- Use `.env` files (add to `.gitignore`)
- Use environment variables in CI/CD
- Rotate keys regularly

⚠️ **HTTPS only in production**
- Obtain TLS certificates (Let's Encrypt)
- Configure ingress with TLS
- Set `Strict-Transport-Security` header

⚠️ **Database backups**
```bash
# Daily backups
pg_dump $DATABASE_URL > backup-$(date +%Y%m%d).sql

# Restore from backup
psql $DATABASE_URL < backup-20250115.sql
```

⚠️ **Monitor logs for security events**
```bash
# Watch for suspicious activity
kubectl logs -f deployment/vigilum-backend -n vigilum | grep -i "error\|warning\|blacklist"
```

---

**Documentation Version**: 1.0  
**Last Updated**: 2025-01-15  
**Next Review**: 2025-02-15
