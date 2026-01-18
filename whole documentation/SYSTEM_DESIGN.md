# VIGILUM System Design

**Purpose:** Comprehensive system design covering service boundaries, data flows, APIs, deployment, scaling, and failure modes. Every decision explicitly documented to prevent architectural drift.

---

## Table of Contents
1. [System Overview](#1-system-overview)
2. [Service Architecture](#2-service-architecture)
3. [Data Architecture](#3-data-architecture)
4. [API Design](#4-api-design)
5. [Blockchain Integration](#5-blockchain-integration)
6. [ZK Proof System](#6-zk-proof-system)
7. [ML Pipeline](#7-ml-pipeline)
8. [Security Model](#8-security-model)
9. [Scaling Strategy](#9-scaling-strategy)
10. [Failure Modes & Recovery](#10-failure-modes--recovery)
11. [Monitoring & Observability](#11-monitoring--observability)
12. [Deployment Architecture](#12-deployment-architecture)

---

## 1. System Overview

### 1.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Wallets   │  │   dApps     │  │  Research Tools (CLI)   │ │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘ │
│         │                 │                     │                │
│         └─────────────────┴─────────────────────┘                │
│                           │                                      │
│                  ┌────────▼────────┐                            │
│                  │  Sentinel SDK   │                            │
│                  │  (Rust + WASM)  │                            │
│                  └────────┬────────┘                            │
└───────────────────────────┼─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                      API GATEWAY LAYER                           │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  API Gateway (Go)                                         │  │
│  │  - Rate limiting                                          │  │
│  │  - Authentication                                         │  │
│  │  - Load balancing                                         │  │
│  │  - Request routing                                        │  │
│  └─────┬───────────┬────────────┬──────────────┬─────────────┘  │
└────────┼───────────┼────────────┼──────────────┼────────────────┘
         │           │            │              │
┌────────▼───────────▼────────────▼──────────────▼────────────────┐
│                     CORE SERVICES LAYER (Go)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Identity   │  │    Threat    │  │   Genome Analyzer    │  │
│  │   Firewall   │  │    Oracle    │  │      Service         │  │
│  │   Service    │  │   Service    │  │                      │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │
│         │                  │                     │               │
│         │                  │                     │               │
│  ┌──────▼──────────────────▼─────────────────────▼───────────┐  │
│  │              Temporal Workflow Engine                      │  │
│  │  - Genome analysis workflows                               │  │
│  │  - Threat feed ingestion                                   │  │
│  │  - Exploit proof processing                                │  │
│  └────────────────────────┬───────────────────────────────────┘  │
└───────────────────────────┼──────────────────────────────────────┘
                            │
┌───────────────────────────▼──────────────────────────────────────┐
│                  CRYPTO & ZK LAYER (Rust)                        │
│  ┌──────────────────────┐  ┌────────────────────────────────┐   │
│  │  ZK Proof Generator  │  │   Cryptographic Key Manager    │   │
│  │  - Human-proof       │  │   - Signing operations         │   │
│  │  - Exploit-proof     │  │   - Key derivation             │   │
│  │  (Noir circuits)     │  │   - Vault integration          │   │
│  └──────────────────────┘  └────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                     DATA LAYER                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Postgres   │  │ ClickHouse  │  │  Redis   │  │   IPFS   │  │
│  │  (OLTP)     │  │  (OLAP)     │  │ (Cache)  │  │ (Genomes)│  │
│  └─────────────┘  └─────────────┘  └──────────┘  └──────────┘  │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                   BLOCKCHAIN LAYER                               │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Smart Contracts (Solidity)                              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐  │   │
│  │  │   Identity   │  │  Threat      │  │  Malware      │  │   │
│  │  │   Firewall   │  │  Oracle      │  │  Genome DB    │  │   │
│  │  └──────────────┘  └──────────────┘  └───────────────┘  │   │
│  │  ┌──────────────┐  ┌──────────────┐                     │   │
│  │  │ Red-Team DAO │  │Proof-of-     │                     │   │
│  │  │              │  │Exploit       │                     │   │
│  │  └──────────────┘  └──────────────┘                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                   Ethereum / Arbitrum                            │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                     ML LAYER (Python)                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Training Pipeline                                        │   │
│  │  - Behavioral model training                             │   │
│  │  - Anomaly detection tuning                              │   │
│  │  - Genome clustering                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Inference Service (ONNX via Go/Rust)                    │   │
│  │  - Real-time human scoring                               │   │
│  │  - Anomaly scoring                                       │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

### 1.2 Core Principles

1. **Separation of Concerns**: Each service has a single responsibility
2. **Fail-Safe Defaults**: Protocols can choose fail-open or fail-closed behavior
3. **Zero-Trust**: All inter-service communication is authenticated
4. **Immutability**: Critical data (genomes, proofs) stored immutably
5. **Observable**: Every operation logged, traced, and metered

---

## 2. Service Architecture

### 2.1 Identity Firewall Service

**Responsibility:** Verify human-like behavior proofs and provide risk scoring.

**Technology:** Go (Gin framework)

**Endpoints:**
- `POST /api/v1/firewall/verify-proof`
- `GET /api/v1/firewall/challenge`
- `GET /api/v1/firewall/risk/:address`
- `GET /api/v1/firewall/stats`

**Dependencies:**
- Postgres (proof logs, user records)
- Redis (rate limiting, challenge caching)
- Ethereum RPC (contract calls)
- ZK Prover (Rust service via gRPC)

**Scaling:**
- Stateless: horizontally scalable
- Database connection pooling (max 50 connections per instance)
- Redis for distributed rate limiting

**Performance Targets:**
- p50 latency: <20ms
- p99 latency: <50ms
- Throughput: 1000 requests/sec per instance

**Failure Modes:**
- Database down → return cached results (5min TTL) + alert
- ZK prover down → queue requests, retry with exponential backoff
- Ethereum RPC down → use backup RPC providers (Infura, Alchemy)

---

### 2.2 Threat Oracle Service

**Responsibility:** Aggregate threat intel from multiple feeds and publish on-chain risk signals.

**Technology:** Go

**Components:**
- Feed Ingestion Worker (pulls from GitHub, advisories, Telegram)
- Signal Aggregator (combines multiple sources)
- Risk Publisher (publishes to ThreatOracle contract)

**Data Flow:**
```
External Feeds → Ingestion Worker → Raw Events (ClickHouse)
                                            ↓
                                    Signal Aggregator
                                            ↓
                                    Risk Scores (Postgres)
                                            ↓
                                    Risk Publisher → Ethereum
```

**Scheduling:**
- Feed ingestion: Every 5 minutes
- Risk aggregation: Every 10 minutes
- On-chain publishing: Every 1 hour (gas optimization)

**Performance Targets:**
- Feed processing: <1 minute per source
- Signal delay: <15 minutes from external event to on-chain

**Failure Modes:**
- Feed source down → skip, log, continue with other sources
- Aggregation fails → use last known good state
- On-chain publish fails → retry up to 3 times, alert if persistent

---

### 2.3 Genome Analyzer Service

**Responsibility:** Analyze contract bytecode and transaction traces to generate malware genomes.

**Technology:** Go (orchestration) + Rust (analysis engine)

**Workflow (Temporal):**
```
1. Receive analysis request (contract address or tx hash)
2. Fetch bytecode / traces from Ethereum
3. Extract features:
   - Opcode histogram
   - Call graph
   - Gas patterns
   - State transitions
4. Compute genome hash (deterministic)
5. Check novelty vs existing genomes
6. Store genome to IPFS
7. Register genome hash on-chain (MalwareGenomeDB)
8. Update Postgres index
9. Notify subscribed services
```

**Performance Targets:**
- Simple contract analysis: <10 seconds
- Complex contract (1000+ opcodes): <60 seconds
- Throughput: 100 contracts/hour per worker

**Scaling:**
- Temporal workers: auto-scale 1-20 based on queue depth
- IPFS pinning: use Infura + Pinata for redundancy

**Failure Modes:**
- Bytecode fetch fails → retry 3 times, mark as failed
- IPFS pin fails → use backup service
- On-chain registration fails → queue for later retry

---

### 2.4 API Gateway

**Responsibility:** Single entry point for all client requests.

**Technology:** Go (net/http + middleware)

**Features:**
- Rate limiting (per IP, per API key)
- Authentication (API keys, JWT)
- Request routing to backend services
- Request/response logging
- CORS handling

**Rate Limits:**
- Unauthenticated: 10 req/min per IP
- API key (free): 100 req/min
- API key (paid): 1000 req/min

**Endpoints:**
```
/api/v1/firewall/*    → Identity Firewall Service
/api/v1/oracle/*      → Threat Oracle Service
/api/v1/genome/*      → Genome Analyzer Service
/api/v1/redteam/*     → Red-Team DAO Service
```

**Performance Targets:**
- Overhead: <5ms per request
- Throughput: 10,000 req/sec

---

### 2.5 ZK Proof Generator (Rust)

**Responsibility:** Generate ZK proofs for human-behavior and exploit-existence.

**Technology:** Rust (Noir circuits + barretenberg prover)

**Interface:** gRPC

**Methods:**
```protobuf
service ZKProver {
  rpc GenerateHumanProof(HumanProofRequest) returns (ProofResponse);
  rpc GenerateExploitProof(ExploitProofRequest) returns (ProofResponse);
  rpc VerifyProof(VerifyRequest) returns (VerifyResponse);
}
```

**Performance Targets:**
- Human-proof generation: <2 seconds
- Exploit-proof generation: <10 seconds
- Proof size: <2KB

**Scaling:**
- CPU-intensive: vertical scaling (16+ cores per instance)
- Queue-based: Temporal activities call this service

**Failure Modes:**
- Proof generation timeout (30s) → return error, client retries
- Invalid inputs → return validation error immediately

---

## 3. Data Architecture

### 3.1 Postgres Schema

**Purpose:** OLTP - transactional data, user records, proof logs.

**Tables:**

```sql
-- Users (wallets)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP,
    risk_score FLOAT DEFAULT 0.0,
    INDEX idx_wallet (wallet_address)
);

-- Human proofs
CREATE TABLE human_proofs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    proof_hash BYTEA NOT NULL,
    proof_data JSONB,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP,
    verifier_address VARCHAR(42), -- which contract verified
    INDEX idx_user_proofs (user_id, created_at DESC),
    INDEX idx_proof_hash (proof_hash)
);

-- Threat signals
CREATE TABLE threat_signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_address VARCHAR(42) NOT NULL,
    signal_type VARCHAR(50) NOT NULL, -- 'exploit_detected', 'key_leaked', etc
    risk_score INT CHECK (risk_score BETWEEN 0 AND 100),
    confidence FLOAT,
    source VARCHAR(100), -- 'github', 'telegram', 'manual'
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP, -- when written on-chain
    INDEX idx_entity_signals (entity_address, created_at DESC),
    INDEX idx_unpublished (published_at) WHERE published_at IS NULL
);

-- Malware genomes
CREATE TABLE genomes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    genome_hash BYTEA UNIQUE NOT NULL,
    ipfs_hash VARCHAR(100) NOT NULL,
    contract_address VARCHAR(42),
    label VARCHAR(50), -- 'known_exploit', 'suspicious', 'benign'
    features JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_genome_hash (genome_hash),
    INDEX idx_contract (contract_address)
);

-- Exploit submissions (from Red-Team DAO)
CREATE TABLE exploit_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    researcher_address VARCHAR(42) NOT NULL,
    target_contract VARCHAR(42) NOT NULL,
    proof_hash BYTEA NOT NULL,
    genome_id UUID REFERENCES genomes(id),
    description TEXT,
    severity VARCHAR(20), -- 'low', 'medium', 'high', 'critical'
    bounty_amount BIGINT,
    status VARCHAR(20), -- 'pending', 'verified', 'rejected', 'paid'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP,
    INDEX idx_researcher (researcher_address, created_at DESC),
    INDEX idx_status (status)
);

-- API keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash BYTEA UNIQUE NOT NULL,
    user_id UUID REFERENCES users(id),
    tier VARCHAR(20), -- 'free', 'paid'
    rate_limit INT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE,
    INDEX idx_key_hash (key_hash)
);
```

**Backup Strategy:**
- Automated daily backups (full + WAL archiving)
- Point-in-time recovery (PITR) enabled
- Replication: 1 primary + 2 read replicas

---

### 3.2 ClickHouse Schema

**Purpose:** OLAP - high-volume event analytics, behavioral traces.

**Tables:**

```sql
-- Transaction traces (for behavioral analysis)
CREATE TABLE tx_traces (
    timestamp DateTime,
    block_number UInt64,
    tx_hash String,
    from_address String,
    to_address String,
    value UInt256,
    gas_used UInt64,
    gas_price UInt64,
    input_data String,
    status UInt8,
    INDEX idx_from (from_address, timestamp),
    INDEX idx_to (to_address, timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, tx_hash);

-- Behavioral features (computed)
CREATE TABLE behavioral_features (
    wallet_address String,
    computed_at DateTime,
    tx_count UInt32,
    avg_tx_interval Float64,
    gas_variance Float64,
    interaction_diversity UInt32,
    unique_contracts UInt32,
    last_activity DateTime,
    INDEX idx_wallet (wallet_address, computed_at)
) ENGINE = ReplacingMergeTree(computed_at)
PARTITION BY toYYYYMM(computed_at)
ORDER BY (wallet_address, computed_at);

-- Threat feed events
CREATE TABLE threat_feed_events (
    timestamp DateTime,
    source String,
    event_type String,
    target_address String,
    severity String,
    metadata String, -- JSON
    INDEX idx_target (target_address, timestamp),
    INDEX idx_source (source, timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, source);
```

**Retention Policy:**
- tx_traces: 90 days
- behavioral_features: 180 days
- threat_feed_events: 365 days

---

### 3.3 Redis Structure

**Purpose:** Caching, rate limiting, real-time state.

**Key Patterns:**

```
# Rate limiting (per IP)
rate_limit:ip:{ip_address} → counter (TTL: 60s)

# Rate limiting (per API key)
rate_limit:key:{api_key_hash} → counter (TTL: 60s)

# Challenge cache (for human-proof flow)
challenge:{challenge_id} → {wallet_address, expires_at} (TTL: 300s)

# Risk score cache
risk_score:{address} → {score, updated_at} (TTL: 300s)

# Genome cache (recent lookups)
genome:{genome_hash} → {ipfs_hash, label} (TTL: 3600s)

# Lock for on-chain operations (prevent duplicates)
lock:publish:{signal_id} → 1 (TTL: 60s)
```

**Eviction Policy:** LRU (Least Recently Used)

**Persistence:** RDB snapshots every 5 minutes + AOF

---

### 3.4 IPFS Architecture

**Purpose:** Immutable storage for genome records and supporting data.

**Setup:**
- Primary node: Self-hosted Kubo node
- Backup: Infura IPFS pinning service
- Cold storage: Arweave for critical genomes

**Pinning Strategy:**
```
1. Upload genome JSON to local Kubo node
2. Get IPFS hash (CIDv1)
3. Pin to Infura (async)
4. If genome is verified exploit: also pin to Arweave
```

**Data Format (Genome JSON):**
```json
{
  "version": "1.0",
  "genome_hash": "0x...",
  "contract_address": "0x...",
  "analyzed_at": "2026-01-18T12:00:00Z",
  "features": {
    "opcode_histogram": {...},
    "call_graph": {...},
    "gas_patterns": {...}
  },
  "label": "known_exploit",
  "metadata": {...}
}
```

---

## 4. API Design

### 4.1 REST API Specification

**Base URL:** `https://api.vigilum.network/v1`

**Authentication:**
- Header: `Authorization: Bearer <api_key>`
- OR: `X-API-Key: <api_key>`

**Common Response Format:**
```json
{
  "success": true,
  "data": {...},
  "error": null,
  "timestamp": "2026-01-18T12:00:00Z"
}
```

**Error Response:**
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 42 seconds.",
    "details": {}
  },
  "timestamp": "2026-01-18T12:00:00Z"
}
```

---

### 4.2 Identity Firewall API

#### POST /firewall/verify-proof

**Purpose:** Verify a human-behavior ZK proof.

**Request:**
```json
{
  "proof": "0x...",  // hex-encoded proof bytes
  "public_inputs": {
    "wallet_address": "0x..."
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "verified": true,
    "proof_hash": "0x...",
    "tx_hash": "0x...",  // on-chain verification tx
    "risk_score": 0.15,   // 0.0-1.0
    "expires_at": "2026-01-19T12:00:00Z"
  }
}
```

**Status Codes:**
- 200: Proof verified successfully
- 400: Invalid proof format
- 401: Unauthorized (missing/invalid API key)
- 429: Rate limit exceeded
- 500: Internal server error

---

#### GET /firewall/challenge

**Purpose:** Get a challenge for proof generation.

**Response:**
```json
{
  "success": true,
  "data": {
    "challenge_id": "uuid",
    "challenge": "0x...",
    "expires_at": "2026-01-18T12:05:00Z"
  }
}
```

---

#### GET /firewall/risk/:address

**Purpose:** Get risk score for a wallet/contract address.

**Response:**
```json
{
  "success": true,
  "data": {
    "address": "0x...",
    "risk_score": 0.85,  // 0.0-1.0
    "risk_level": "high",  // low, medium, high, critical
    "signals": [
      {
        "type": "exploit_detected",
        "source": "github",
        "confidence": 0.9,
        "timestamp": "2026-01-18T11:30:00Z"
      }
    ],
    "updated_at": "2026-01-18T12:00:00Z"
  }
}
```

---

### 4.3 Threat Oracle API

#### GET /oracle/signals/:address

**Purpose:** Get all threat signals for an address.

**Response:**
```json
{
  "success": true,
  "data": {
    "address": "0x...",
    "signals": [
      {
        "id": "uuid",
        "type": "exploit_detected",
        "risk_score": 95,
        "confidence": 0.92,
        "source": "github",
        "created_at": "2026-01-18T10:00:00Z",
        "published_at": "2026-01-18T11:00:00Z"
      }
    ],
    "on_chain_risk_score": 95  // latest published score
  }
}
```

---

#### POST /oracle/subscribe

**Purpose:** Subscribe to threat signal updates via webhook.

**Request:**
```json
{
  "webhook_url": "https://your-protocol.com/webhook",
  "addresses": ["0x...", "0x..."],
  "signal_types": ["exploit_detected", "key_leaked"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "subscription_id": "uuid",
    "active": true
  }
}
```

---

### 4.4 Genome Analyzer API

#### POST /genome/analyze

**Purpose:** Request genome analysis for a contract.

**Request:**
```json
{
  "contract_address": "0x...",
  "priority": "normal"  // normal, high, urgent
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "analysis_id": "uuid",
    "status": "queued",
    "estimated_completion": "2026-01-18T12:05:00Z"
  }
}
```

---

#### GET /genome/status/:analysis_id

**Purpose:** Get status of genome analysis.

**Response:**
```json
{
  "success": true,
  "data": {
    "analysis_id": "uuid",
    "status": "completed",  // queued, processing, completed, failed
    "genome_hash": "0x...",
    "ipfs_hash": "Qm...",
    "label": "suspicious",
    "similarity": [
      {
        "genome_hash": "0x...",
        "similarity_score": 0.87,
        "label": "known_exploit"
      }
    ]
  }
}
```

---

### 4.5 Red-Team DAO API

#### POST /redteam/submit-exploit

**Purpose:** Submit a Proof-of-Exploit.

**Request:**
```json
{
  "target_contract": "0x...",
  "proof": "0x...",
  "description": "Reentrancy vulnerability in withdraw function",
  "severity": "critical"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "submission_id": "uuid",
    "status": "pending_verification",
    "estimated_bounty": "5000000000000000000"  // wei
  }
}
```

---

## 5. Blockchain Integration

### 5.1 Smart Contract Architecture

**Deployment Strategy:**
- UUPS Proxy pattern for upgradeability
- Separate implementation and proxy contracts
- ProxyAdmin managed by multisig (3-of-5)

**Gas Optimization:**
- Use events for non-critical data
- Pack variables in storage slots
- Batch operations where possible

---

### 5.2 IdentityFirewall Contract

**Key Functions:**

```solidity
interface IIdentityFirewall {
    /// @notice Verify a human-behavior ZK proof
    /// @param proof The ZK proof bytes
    /// @return verified True if proof is valid
    function verifyHumanProof(bytes calldata proof) external returns (bool verified);
    
    /// @notice Check if a proof hash has been verified
    /// @param proofHash The hash of the proof
    /// @return verified True if previously verified
    function hasVerifiedProof(bytes32 proofHash) external view returns (bool verified);
    
    /// @notice Get the last verification timestamp for an address
    /// @param user The wallet address
    /// @return timestamp The last verification time
    function lastVerification(address user) external view returns (uint256 timestamp);
}
```

**Events:**
```solidity
event ProofVerified(address indexed user, bytes32 indexed proofHash, uint256 timestamp);
event VerifierUpdated(address indexed oldVerifier, address indexed newVerifier);
```

**Gas Costs (estimated):**
- verifyHumanProof: ~150,000 gas
- hasVerifiedProof: ~5,000 gas (view)

---

### 5.3 ThreatOracle Contract

**Key Functions:**

```solidity
interface IThreatOracle {
    /// @notice Update risk score for a target address
    /// @param target The address to update
    /// @param riskScore Risk score (0-100)
    function updateRiskScore(address target, uint256 riskScore) external;
    
    /// @notice Get current risk score
    /// @param target The address to query
    /// @return score Current risk score
    function getRiskScore(address target) external view returns (uint256 score);
    
    /// @notice Check if address is flagged as high risk
    /// @param target The address to check
    /// @return isHighRisk True if risk >= threshold
    function isHighRisk(address target) external view returns (bool isHighRisk);
}
```

**Access Control:**
- Only authorized oracle publishers can call updateRiskScore
- Multi-oracle design: require 2-of-3 oracles to agree on high-risk designation

---

### 5.4 MalwareGenomeDB Contract

**Key Functions:**

```solidity
interface IMalwareGenomeDB {
    /// @notice Register a new genome
    /// @param genomeHash The genome hash
    /// @param ipfsHash The IPFS content ID
    /// @param label Classification label
    function registerGenome(
        bytes32 genomeHash,
        string calldata ipfsHash,
        string calldata label
    ) external;
    
    /// @notice Check if genome exists
    /// @param genomeHash The genome hash
    /// @return exists True if genome is registered
    function hasGenome(bytes32 genomeHash) external view returns (bool exists);
    
    /// @notice Get genome details
    /// @param genomeHash The genome hash
    /// @return ipfsHash, label, timestamp
    function getGenome(bytes32 genomeHash) 
        external 
        view 
        returns (string memory ipfsHash, string memory label, uint256 timestamp);
}
```

---

### 5.5 RedTeamDAO Contract

**Key Functions:**

```solidity
interface IRedTeamDAO {
    /// @notice Stake reputation tokens
    /// @param amount Amount to stake
    function stake(uint256 amount) external;
    
    /// @notice Submit an exploit proof
    /// @param target Target contract address
    /// @param proof ZK proof bytes
    /// @param description Vulnerability description
    /// @return submissionId Unique submission ID
    function submitExploit(
        address target,
        bytes calldata proof,
        string calldata description
    ) external returns (bytes32 submissionId);
    
    /// @notice Vote on an exploit submission (governance)
    /// @param submissionId The submission to vote on
    /// @param approve True to approve, false to reject
    function vote(bytes32 submissionId, bool approve) external;
    
    /// @notice Claim bounty reward
    /// @param submissionId The approved submission
    function claimBounty(bytes32 submissionId) external;
}
```

**Reward Calculation:**
```solidity
bounty = BASE_REWARD * severity_multiplier * novelty_bonus
severity_multiplier: low=1x, medium=2x, high=5x, critical=10x
novelty_bonus: first of its kind = +50%
```

---

### 5.6 On-Chain Event Processing

**Backend Event Listener:**

```go
// Subscribe to contract events
func (s *EthereumService) SubscribeToEvents(ctx context.Context) {
    // Identity Firewall
    proofVerifiedChan := make(chan *IdentityFirewallProofVerified)
    sub1, _ := s.identityFirewall.WatchProofVerified(nil, proofVerifiedChan, nil)
    
    // Threat Oracle
    riskUpdatedChan := make(chan *ThreatOracleRiskUpdated)
    sub2, _ := s.threatOracle.WatchRiskUpdated(nil, riskUpdatedChan, nil)
    
    // Process events
    for {
        select {
        case event := <-proofVerifiedChan:
            s.handleProofVerified(event)
        case event := <-riskUpdatedChan:
            s.handleRiskUpdated(event)
        case <-ctx.Done():
            sub1.Unsubscribe()
            sub2.Unsubscribe()
            return
        }
    }
}
```

---

## 6. ZK Proof System

### 6.1 Human-Behavior Proof Circuit

**Public Inputs:**
- `wallet_address` (address)
- `proof_timestamp` (uint64)

**Private Inputs:**
- `tx_timestamps[]` (array of last 20 tx timestamps)
- `gas_values[]` (array of last 20 tx gas values)
- `signature` (wallet signature over inputs)

**Constraints:**
```
1. Verify signature is valid for wallet_address
2. Compute mean and variance of tx_timestamps
3. Assert: variance(tx_timestamps) > MIN_VARIANCE
4. Assert: variance(tx_timestamps) < MAX_VARIANCE
5. Compute variance of gas_values
6. Assert: gas_variance < MAX_GAS_VARIANCE
7. Assert: all timestamps are in past
8. Assert: timestamps are chronologically ordered
```

**Circuit Complexity:**
- ~2000 constraints
- Proving time: ~1.5 seconds
- Verification time: ~5ms (on-chain)
- Proof size: ~1.2 KB

---

### 6.2 Exploit-Existence Proof Circuit

**Public Inputs:**
- `contract_address` (address)
- `property_violated` (enum: Reentrancy, Overflow, etc.)

**Private Inputs:**
- `execution_trace` (opcodes + state changes)
- `initial_state` (contract state before exploit)
- `final_state` (contract state after exploit)

**Constraints:**
```
1. Assert: initial_state is valid contract state
2. Simulate execution_trace on initial_state
3. Assert: execution produces final_state
4. Assert: property_violated condition is met in final_state
5. Assert: trace is non-trivial (>10 steps)
```

**Circuit Complexity:**
- ~10,000 constraints (depends on trace length)
- Proving time: ~8 seconds
- Verification time: ~10ms (on-chain)
- Proof size: ~2 KB

---

### 6.3 Proof Generation Workflow

```
SDK (Client Side)                  ZK Prover (Rust)              Backend (Go)
      │                                   │                            │
      ├─ Collect behavioral data          │                            │
      ├─ Extract features                 │                            │
      ├─ Call ZK Prover via gRPC ─────────►                            │
      │                                   ├─ Load circuit              │
      │                                   ├─ Generate witness          │
      │                                   ├─ Compute proof (~2s)       │
      │                                   ├─ Return proof ─────────────┤
      ◄─────────────────────────────────────────────────────────────────┤
      │                                   │                            │
      ├─ Send proof to backend ───────────────────────────────────────►│
      │                                   │                            ├─ Verify proof
      │                                   │                            ├─ Call contract
      │                                   │                            ├─ Log to DB
      ◄─ Return verification result ──────────────────────────────────┤
```

---

## 7. ML Pipeline

### 7.1 Training Pipeline

**Data Collection:**
```python
# ml/scripts/collect_data.py
def collect_behavioral_data():
    # Query Ethereum RPC
    wallets = get_active_wallets(limit=50000)
    
    for wallet in wallets:
        txs = get_transactions(wallet, limit=100)
        features = extract_features(txs)
        label = heuristic_labeling(features)  # human vs bot
        
        save_to_dataset(wallet, features, label)
```

**Feature Engineering:**
- Transaction count (lifetime)
- Mean transaction interval (seconds)
- Std dev of transaction intervals
- Gas usage variance
- Number of unique contracts interacted with
- Time-of-day distribution
- Day-of-week distribution
- Contract call diversity (ERC20, DeFi, NFT, etc.)

**Model Training:**
```python
# ml/src/vigilum_ml/models/human_classifier.py
from sklearn.ensemble import RandomForestClassifier
import polars as pl

# Load data
df = pl.read_parquet("data/processed/features.parquet")

# Train
X = df.select([pl.col("tx_count"), pl.col("avg_interval"), ...])
y = df.select("label")

model = RandomForestClassifier(n_estimators=100, max_depth=10)
model.fit(X, y)

# Evaluate
accuracy = model.score(X_test, y_test)
print(f"Accuracy: {accuracy}")

# Export to ONNX
import skl2onnx
onnx_model = skl2onnx.convert_sklearn(model)
with open("models/human_classifier.onnx", "wb") as f:
    f.write(onnx_model.SerializeToString())
```

**Model Versioning:**
- Models stored in Git LFS
- Semantic versioning: `human_classifier_v1.2.3.onnx`
- Metadata file tracks training date, accuracy, dataset size

---

### 7.2 Inference Pipeline

**Go Service Integration:**

```go
// backend/internal/inference/serve.go
package inference

import (
    ort "github.com/yalue/onnxruntime_go"
)

type InferenceService struct {
    session *ort.Session
}

func NewInferenceService(modelPath string) (*InferenceService, error) {
    session, err := ort.NewSession(modelPath)
    if err != nil {
        return nil, err
    }
    return &InferenceService{session: session}, nil
}

func (s *InferenceService) PredictHumanScore(features BehavioralFeatures) (float64, error) {
    // Prepare input tensor
    input := []float32{
        float32(features.TxCount),
        float32(features.AvgInterval),
        float32(features.GasVariance),
        float32(features.Diversity),
    }
    
    inputTensor, _ := ort.NewTensor(ort.NewShape(1, 4), input)
    defer inputTensor.Destroy()
    
    // Run inference
    outputs, err := s.session.Run([]ort.Value{inputTensor})
    if err != nil {
        return 0, err
    }
    defer outputs[0].Destroy()
    
    // Extract score
    outputData := outputs[0].GetData().([]float32)
    return float64(outputData[0]), nil
}
```

**Performance:**
- Inference latency: <5ms
- Throughput: 10,000 predictions/sec per instance
- Model size: ~50 MB (loaded into memory)

---

### 7.3 Model Monitoring

**Metrics Tracked:**
- Prediction distribution (score histogram)
- False positive rate (from user feedback)
- Model drift (compare input distributions over time)

**Retraining Triggers:**
- Scheduled: Every 30 days
- Accuracy drops below 80% on validation set
- Significant drift detected in input features

---

## 8. Security Model

### 8.1 Threat Vectors & Mitigations

| Threat | Mitigation |
|--------|------------|
| API key theft | Rate limiting, key rotation, IP whitelisting |
| DDoS attack | Rate limiting, CDN (Cloudflare), auto-scaling |
| Database injection | Parameterized queries, ORM, input validation |
| Proof forgery | ZK verification on-chain, deterministic checks |
| Oracle manipulation | Multi-oracle consensus, stake+slash mechanism |
| IPFS data loss | Redundant pinning (Infura + Pinata + Arweave) |
| Smart contract exploit | Audits, formal verification, bug bounty |
| Insider threat | Zero-trust, audit logs, least-privilege access |
| Key compromise | Vault for secrets, automated rotation, HSM for signing |

---

### 8.2 Authentication & Authorization

**API Authentication:**
- API keys stored as bcrypt hashes
- JWT tokens for user sessions (1 hour expiry)
- mTLS for service-to-service communication

**Smart Contract Access Control:**
```solidity
// OpenZeppelin AccessControl
contract IdentityFirewall is AccessControl {
    bytes32 public constant VERIFIER_ROLE = keccak256("VERIFIER_ROLE");
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE");
    
    modifier onlyVerifier() {
        require(hasRole(VERIFIER_ROLE, msg.sender), "Not authorized");
        _;
    }
}
```

---

### 8.3 Data Privacy

**Personal Data Handling:**
- Wallet addresses are public (blockchain data)
- Behavioral features aggregated, no raw tx data stored
- No KYC or PII collected
- GDPR compliance: right to erasure (delete user record, proofs remain on-chain)

**ZK Proofs:**
- Private inputs never leave client
- Only proof + public inputs sent to backend
- On-chain data reveals nothing about private inputs

---

## 9. Scaling Strategy

### 9.1 Horizontal Scaling

**Stateless Services:**
- Identity Firewall: 3-10 instances (auto-scale on CPU >70%)
- Threat Oracle: 2-5 instances
- API Gateway: 5-20 instances (auto-scale on requests/sec)

**Stateful Services:**
- Postgres: 1 primary + 2 read replicas
- ClickHouse: 3-node cluster (sharded by month)
- Redis: 3-node cluster (sentinel mode for HA)

---

### 9.2 Vertical Scaling

**ZK Prover:**
- CPU-intensive: 16-32 core instances
- Memory: 64GB+ RAM
- Not horizontally scalable (queue-based instead)

---

### 9.3 Database Scaling

**Postgres:**
- Partitioning: `human_proofs` partitioned by month
- Read replicas for analytics queries
- Connection pooling: PgBouncer (1000 connections → 50 DB connections)

**ClickHouse:**
- Partitioning by month (automatic TTL)
- Distributed tables across 3 nodes
- Materialized views for aggregations

---

### 9.4 Caching Strategy

**Redis Layers:**
- L1: Application-level cache (in-memory)
- L2: Redis cache (5-minute TTL)
- L3: Database

**Cache Invalidation:**
- On risk score update: invalidate `risk_score:{address}`
- On genome registration: invalidate `genome:{genome_hash}`

---

## 10. Failure Modes & Recovery

### 10.1 Service Failures

| Service | Failure Mode | Recovery Strategy | SLA |
|---------|--------------|-------------------|-----|
| Identity Firewall | Instance crash | Auto-restart, LB routes to healthy instance | 99.9% |
| Threat Oracle | Feed source down | Use cached data, alert ops team | 99.5% |
| Genome Analyzer | IPFS pin fails | Retry with backup provider (Infura) | 99.5% |
| ZK Prover | Proof gen timeout | Return error, client retries | 99.0% |
| API Gateway | Overload | Auto-scale, reject excess with 503 | 99.95% |

---

### 10.2 Database Failures

**Postgres Primary Failure:**
1. Failover to read replica (promoted to primary)
2. Update DNS / service discovery
3. Downtime: <60 seconds

**ClickHouse Node Failure:**
1. Queries automatically routed to healthy nodes
2. Data replicated 3x, no data loss
3. Downtime: 0 seconds (transparent)

**Redis Failure:**
1. Sentinel promotes new master
2. Clients reconnect automatically
3. Cache miss spike (acceptable)

---

### 10.3 Blockchain Failures

**Ethereum RPC Failure:**
- Primary: Infura
- Fallback 1: Alchemy
- Fallback 2: QuickNode
- Automatic failover (<1 second)

**High Gas Prices:**
- Queue on-chain operations
- Publish in batches during low-gas periods
- Alert if pending >24 hours

**Contract Exploit:**
- Emergency pause function (multisig)
- Upgrade to patched implementation (UUPS)
- Affected users notified via webhook

---

### 10.4 Disaster Recovery

**Backup Strategy:**
- Postgres: Daily full backup + continuous WAL archiving (S3)
- ClickHouse: Weekly snapshots (S3)
- Redis: RDB snapshots every 5 minutes
- Contracts: Immutable on-chain (no backup needed)
- IPFS: Multi-region pinning

**RTO (Recovery Time Objective):** 4 hours
**RPO (Recovery Point Objective):** 5 minutes

**DR Runbook:**
1. Restore Postgres from latest backup
2. Restore ClickHouse from snapshot
3. Redeploy services to DR region
4. Update DNS to DR region
5. Validate all services healthy
6. Resume normal operations

---

## 11. Monitoring & Observability

### 11.1 Metrics (Prometheus)

**System Metrics:**
- CPU, memory, disk usage per service
- Network I/O, latency
- Container health (Kubernetes)

**Application Metrics:**
```
# Identity Firewall
vigilum_firewall_proofs_verified_total (counter)
vigilum_firewall_verification_duration_seconds (histogram)
vigilum_firewall_risk_score_by_address (gauge)

# Threat Oracle
vigilum_oracle_signals_published_total (counter)
vigilum_oracle_feed_errors_total (counter)
vigilum_oracle_risk_updates_total (counter)

# Genome Analyzer
vigilum_genome_analyses_total (counter)
vigilum_genome_analysis_duration_seconds (histogram)
vigilum_genome_ipfs_pin_failures_total (counter)

# ZK Prover
vigilum_zkprover_proofs_generated_total (counter)
vigilum_zkprover_proof_generation_duration_seconds (histogram)
```

**Dashboards (Grafana):**
- Service health overview
- API latency percentiles (p50, p95, p99)
- Database query performance
- Blockchain interaction status
- ZK proof generation metrics

---

### 11.2 Logging (Loki)

**Structured Logging Format:**
```json
{
  "timestamp": "2026-01-18T12:00:00Z",
  "level": "info",
  "service": "identity-firewall",
  "trace_id": "abc123",
  "span_id": "def456",
  "message": "Proof verified successfully",
  "wallet_address": "0x...",
  "proof_hash": "0x...",
  "verification_duration_ms": 42
}
```

**Log Levels:**
- DEBUG: Detailed flow for troubleshooting
- INFO: Normal operations
- WARN: Recoverable errors, degraded state
- ERROR: Service errors requiring attention
- CRITICAL: System-wide failures

**Retention:**
- DEBUG logs: 7 days
- INFO/WARN logs: 30 days
- ERROR/CRITICAL logs: 90 days

---

### 11.3 Tracing (Jaeger)

**Distributed Traces:**
- Every API request generates a trace_id
- Spans created for each service hop
- Visualize full request flow: Gateway → Service → DB → Contract

**Example Trace:**
```
POST /firewall/verify-proof [200ms]
  ├─ API Gateway [5ms]
  ├─ Identity Firewall Service [195ms]
  │   ├─ Database query [10ms]
  │   ├─ ZK Prover gRPC call [150ms]
  │   ├─ Ethereum contract call [30ms]
  │   └─ Database insert [5ms]
```

---

### 11.4 Alerting (AlertManager)

**Critical Alerts (PagerDuty):**
- Service down for >2 minutes
- Database primary failure
- ZK Prover queue depth >100
- On-chain transaction failing for >1 hour
- Security event (unauthorized access attempt)

**Warning Alerts (Slack):**
- API latency p99 >100ms
- Error rate >1%
- Disk usage >80%
- Feed source unavailable for >10 minutes

---

## 12. Deployment Architecture

### 12.1 Kubernetes Cluster Layout

```
Namespace: vigilum-prod

Deployments:
├─ api-gateway (5 replicas, HPA: 5-20)
├─ identity-firewall (3 replicas, HPA: 3-10)
├─ threat-oracle (2 replicas)
├─ genome-analyzer (2 replicas)
├─ zk-prover (3 replicas, no HPA)
└─ temporal-server (3 replicas)

StatefulSets:
├─ postgres (1 primary + 2 replicas)
├─ clickhouse (3 nodes)
├─ redis (3 nodes, sentinel)
└─ ipfs-kubo (2 nodes)

Services:
├─ api-gateway-svc (LoadBalancer)
├─ identity-firewall-svc (ClusterIP)
├─ threat-oracle-svc (ClusterIP)
├─ genome-analyzer-svc (ClusterIP)
├─ zk-prover-svc (ClusterIP, gRPC)
├─ postgres-svc (ClusterIP, port 5432)
├─ clickhouse-svc (ClusterIP, port 8123)
└─ redis-svc (ClusterIP, port 6379)
```

---

### 12.2 Resource Allocation

```yaml
# api-gateway
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi

# identity-firewall
resources:
  requests:
    cpu: 1000m
    memory: 1Gi
  limits:
    cpu: 4000m
    memory: 4Gi

# zk-prover
resources:
  requests:
    cpu: 8000m
    memory: 16Gi
  limits:
    cpu: 16000m
    memory: 32Gi

# postgres
resources:
  requests:
    cpu: 4000m
    memory: 8Gi
  limits:
    cpu: 8000m
    memory: 16Gi
  storage: 500Gi (SSD)
```

---

### 12.3 Network Architecture

```
Internet
    │
    ▼
┌────────────────────┐
│  Cloudflare CDN    │
│  - DDoS protection │
│  - Rate limiting   │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  Load Balancer     │
│  (K8s Ingress)     │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  API Gateway       │
│  (Public-facing)   │
└─────────┬──────────┘
          │
    ┌─────┴─────┬─────────┬─────────┐
    ▼           ▼         ▼         ▼
┌───────┐  ┌─────────┐ ┌────────┐ ┌────────┐
│Firewall│ │ Oracle  │ │Genome  │ │RedTeam │
│        │ │         │ │        │ │        │
└───┬────┘ └────┬────┘ └───┬────┘ └───┬────┘
    │           │           │          │
    └───────────┴───────────┴──────────┘
                │
                ▼
    ┌───────────────────────┐
    │  Data Layer           │
    │  (Internal only)      │
    │  - Postgres           │
    │  - ClickHouse         │
    │  - Redis              │
    │  - IPFS               │
    └───────────────────────┘
```

---

### 12.4 CI/CD Pipeline

**GitHub Actions Workflow:**

```yaml
# .github/workflows/deploy-prod.yml
name: Deploy to Production

on:
  push:
    branches: [main]
    
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - Checkout code
      - Run unit tests (Go, Rust, Python, Solidity)
      - Run integration tests
      - Security scans (Trivy, Slither, cargo-audit)
      
  build:
    needs: test
    steps:
      - Build Docker images
      - Push to container registry
      - Tag with commit SHA
      
  deploy-contracts:
    needs: build
    steps:
      - Deploy contracts to Sepolia (manual approval)
      - Verify on Etherscan
      - Update .env with new contract addresses
      
  deploy-backend:
    needs: build
    steps:
      - Update K8s manifests with new image tags
      - Apply to cluster (via ArgoCD or kubectl)
      - Run smoke tests
      - Monitor for 10 minutes
      - Rollback if errors detected
```

---

### 12.5 Environment Strategy

**Environments:**
1. **Development (dev):** Local docker-compose or minikube
2. **Staging (staging):** K8s cluster + Sepolia testnet
3. **Production (prod):** K8s cluster + Ethereum mainnet

**Promotion Path:**
```
dev → staging (auto-deploy on PR merge) → prod (manual approval)
```

**Config Management:**
- ConfigMaps for non-sensitive config
- Sealed Secrets for sensitive values (Vault in prod)
- Different namespaces per environment

---

## Summary

This system design provides:
- **Clear service boundaries** with well-defined APIs
- **Scalable architecture** with horizontal and vertical scaling paths
- **Failure resilience** with multi-layer redundancy
- **Observable system** with comprehensive monitoring
- **Secure by default** with defense-in-depth

**Next Steps:**
1. Validate design with stakeholders
2. Create detailed object design (classes, structs, interfaces)
3. Begin Phase 0 implementation

**Open Questions for Review:**
- Ethereum mainnet vs Arbitrum for production?
- Self-hosted K8s or managed (EKS/GKE)?
- Target throughput: 1000 or 10,000 req/sec?
