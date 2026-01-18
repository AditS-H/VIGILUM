# VIGILUM

<div align="center">

**Decentralized Blockchain Security Layer**

*Cloudflare + VirusTotal + IDS/IPS + Bug Bounty + AI SOC — but for Web3, on-chain, autonomous, and uncensorable.*

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Rust](https://img.shields.io/badge/Rust-1.75+-000000?style=flat&logo=rust)](https://www.rust-lang.org)
[![Solidity](https://img.shields.io/badge/Solidity-0.8.28-363636?style=flat&logo=solidity)](https://soliditylang.org)
[![Python](https://img.shields.io/badge/Python-3.12+-3776AB?style=flat&logo=python)](https://python.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## 🎯 What is VIGILUM?

VIGILUM is an **autonomous, decentralized security infrastructure** for blockchain ecosystems. Think of it as a security layer that:

- **Scans** every smart contract for vulnerabilities before you interact
- **Monitors** mempool for frontrunning, sandwich attacks, and malicious transactions
- **Rates** contracts with on-chain risk scores anyone can query
- **Proves** audits happened without revealing sensitive details (ZK)
- **Rewards** security researchers via decentralized bug bounties

### Core Features

| Feature | Description |
|---------|-------------|
| 🔍 **Multi-Engine Scanner** | Static analysis, symbolic execution, ML detection, fuzzing |
| 🛡️ **Real-time Protection** | Mempool monitoring, threat alerts, transaction simulation |
| 🧠 **AI Threat Detection** | Transformer models trained on exploit patterns, rug pulls, honeypots |
| 📜 **On-chain Registry** | Query risk scores directly from smart contracts |
| 🔐 **ZK Audit Proofs** | Privacy-preserving attestations via Noir circuits |
| 🏆 **Bug Bounty Protocol** | Decentralized vulnerability disclosure with token rewards |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        VIGILUM STACK                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   CLIENTS        SDK              API              WORKERS       │
│  ┌───────┐    ┌───────┐      ┌─────────┐      ┌─────────────┐  │
│  │ dApps │───▶│TS SDK │─────▶│  REST   │─────▶│  Scanner    │  │
│  │Wallets│    │Python │      │  gRPC   │      │  Indexer    │  │
│  └───────┘    └───────┘      └────┬────┘      │  ML Inference│  │
│                                   │           └──────┬──────┘  │
│                                   ▼                  │          │
│                          ┌────────────────┐         │          │
│                          │  Message Bus   │◀────────┘          │
│                          │    (NATS)      │                    │
│                          └───────┬────────┘                    │
│                                  │                              │
│   ┌──────────┬──────────┬───────┴───────┬──────────┐          │
│   ▼          ▼          ▼               ▼          ▼          │
│ ┌────┐   ┌─────┐   ┌────────┐    ┌───────┐   ┌────────┐      │
│ │ PG │   │Redis│   │Qdrant  │    │Click- │   │Temporal│      │
│ │    │   │Cache│   │Vectors │    │House  │   │Workflow│      │
│ └────┘   └─────┘   └────────┘    └───────┘   └────────┘      │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│                     BLOCKCHAIN LAYER                            │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ VigilumRegistry.sol │ ThreatOracle.sol │ BountyVault  │   │
│  └────────────────────────────────────────────────────────┘   │
│  Chains: Ethereum • Polygon • BSC • Arbitrum • Base           │
└────────────────────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
VIGILUM/
├── backend/                 # Go microservices
│   ├── cmd/api/            # API server (REST + gRPC)
│   ├── cmd/scanner/        # Vulnerability scanner worker
│   ├── cmd/indexer/        # Blockchain event indexer
│   └── internal/           # Core business logic
│       ├── domain/         # Entities (Contract, Vulnerability, Alert)
│       ├── scanner/        # Analysis engines
│       └── config/         # Configuration
│
├── contracts/              # Solidity (Foundry)
│   ├── src/VigilumRegistry.sol    # On-chain security registry
│   └── test/               # Foundry tests (15 passing)
│
├── circuits/               # Noir ZK circuits
│   └── src/proof_of_audit.nr      # Privacy-preserving proofs
│
├── ml/                     # Python ML pipeline
│   └── src/vigilum_ml/
│       ├── features.py     # Bytecode feature extraction
│       └── model.py        # PyTorch vulnerability detector
│
├── sdk/ts-sdk/            # TypeScript SDK
│   └── src/
│       ├── client.ts      # API client
│       └── analyzer.ts    # Local contract analysis
│
├── docker-compose.yml     # Full dev infrastructure
└── Makefile               # Build commands
```

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.23+** — Backend services
- **Python 3.12+** — ML pipeline  
- **Node.js 20+** — TypeScript SDK
- **Docker** — Infrastructure
- **Foundry** — Smart contracts
- **WSL** (Windows) — For Noir circuits

### 1. Start Infrastructure

```bash
# Start Postgres, Redis, Qdrant, ClickHouse, Temporal, Grafana, Jaeger
docker compose up -d
```

**Local URLs:**
| Service | URL |
|---------|-----|
| Temporal UI | http://localhost:8080 |
| Grafana | http://localhost:3000 |
| Jaeger Tracing | http://localhost:16686 |

### 2. Run Backend

```bash
# Terminal 1: API Server
cd backend && go run ./cmd/api

# Terminal 2: Scanner Worker
cd backend && go run ./cmd/scanner

# Terminal 3: Blockchain Indexer
cd backend && go run ./cmd/indexer
```

### 3. Build & Test Contracts

```bash
cd contracts

# Build
forge build

# Test (15 tests including fuzz)
forge test -vvv

# Deploy locally
anvil &                    # Start local chain
forge script script/Deploy.s.sol --broadcast
```

### 4. Setup ML Pipeline

```bash
cd ml
python -m venv .venv
.venv\Scripts\activate     # Windows
pip install -e ".[dev]"
pytest tests/ -v
```

### 5. Build SDK

```bash
cd sdk/ts-sdk
npm install
npm run build
```

---

## 🔧 Makefile Commands

```bash
make dev              # Start full dev environment
make build            # Build all components
make test             # Run all tests

make build-backend    # Build Go services
make build-contracts  # Build Solidity
make build-sdk        # Build TypeScript SDK

make run-api          # Start API server
make anvil            # Start local Ethereum node

make docker-up        # Start infrastructure
make docker-down      # Stop infrastructure
```

---

## 📖 Usage

### TypeScript SDK

```typescript
import { VigilumClient } from '@vigilum/sdk';

const client = new VigilumClient({
  apiKey: 'your-key',
  baseUrl: 'https://api.vigilum.io'
});

// Scan a contract
const result = await client.scan({
  address: '0xdead...',
  chainId: 1
});

console.log(`Risk: ${result.riskScore}/100`);
console.log(`Threats: ${result.threatLevel}`);
```

### Smart Contract Integration

```solidity
import {IVigilumRegistry} from "@vigilum/contracts/IVigilumRegistry.sol";

contract MyDeFi {
    IVigilumRegistry vigilum;
    
    modifier safeOnly(address target) {
        require(!vigilum.isBlacklisted(target), "Blocked");
        require(vigilum.getRiskScore(target) < 6000, "Risky");
        _;
    }
    
    function swap(address token) external safeOnly(token) {
        // Safe to proceed
    }
}
```

---

## 🧪 Test Results

```
Smart Contracts (Foundry): 15/15 passing ✅
├── VigilumRegistry registration
├── Risk score updates
├── Blacklist functionality
├── Threat level mapping
├── Ownership transfer (2-step)
└── Fuzz testing (256 runs)

Backend (Go): Compiles ✅
ML Pipeline: Models defined ✅
SDK: Types & client ready ✅
```

---

## 🗺️ Roadmap

| Phase | Status | Milestone |
|-------|--------|-----------|
| **0** | ✅ | Project structure, domain models, basic scanner |
| **1** | 🚧 | Database migrations, REST API, static analysis |
| **2** | ⏳ | ML training pipeline, real-time indexing |
| **3** | ⏳ | Mainnet deployment, ZK proofs |
| **4** | ⏳ | Bug bounty protocol, governance |

---

## 📄 License

MIT — See [LICENSE](LICENSE)

---

<div align="center">

**Securing Web3, One Contract at a Time** 🛡️

</div>
