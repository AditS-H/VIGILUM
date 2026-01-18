# VIGILUM: Detailed Execution Plan

**Purpose:** Crystal-clear step-by-step blueprint to build VIGILUM. No hallucination. Every file, folder, and task explicitly specified.

---

## Part 1: Project Folder Structure (Complete)

```
vigilum/                                 # ROOT PROJECT FOLDER
├── README.md                            # Quick overview + getting started
├── LICENSE                              # Choose: Apache 2.0 or MIT
│
├── .github/
│   ├── workflows/
│   │   ├── lint-test-build.yml         # Go + Python + Rust lint/test/build
│   │   ├── contract-test.yml           # Foundry tests
│   │   ├── security-scan.yml           # Trivy + cargo-audit + Slither
│   │   ├── deploy-testnet.yml          # Deploy to Sepolia (manual gate)
│   │   └── deploy-prod.yml             # Deploy to mainnet (manual gate)
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   └── pull_request_template.md
│
├── docs/                                # Public documentation
│   ├── ARCHITECTURE.md                 # (link to architecture.md or copy)
│   ├── API_REFERENCE.md                # Auto-generated from OpenAPI/protobuf
│   ├── SDK_GUIDE.md                    # How to use Sentinel SDK
│   ├── CONTRIBUTING.md
│   ├── DEPLOYMENT.md                   # How to deploy VIGILUM infrastructure
│   ├── ZK_CIRCUITS.md                  # Noir circuit design and rationale
│   └── THREAT_MODEL.md                 # Security assumptions + threat analysis
│
├── backend/                             # Go + Rust backend services
│   ├── Dockerfile                       # Multi-stage: Go builder + runtime
│   ├── docker-compose.yml               # Local dev: postgres, redis, ipfs-node
│   ├── Makefile                         # Common tasks: build, test, lint, deploy
│   ├── go.mod / go.sum                  # Go dependencies
│   ├── Cargo.toml (at root)             # Rust workspace
│   │
│   ├── cmd/                             # Go service entry points
│   │   ├── identity-firewall/
│   │   │   └── main.go                  # Identity Firewall service
│   │   ├── threat-oracle/
│   │   │   └── main.go                  # Threat Oracle aggregator
│   │   ├── api-gateway/
│   │   │   └── main.go                  # API gateway / load balancer
│   │   └── cli/
│   │       └── main.go                  # CLI tool for admins
│   │
│   ├── internal/                        # Go private packages
│   │   ├── firewall/
│   │   │   ├── verifier.go              # Human-proof verification logic
│   │   │   ├── models.go                # Data structures
│   │   │   └── service.go               # Service layer
│   │   ├── oracle/
│   │   │   ├── aggregator.go            # Signal aggregation
│   │   │   ├── feeds.go                 # Feed ingestion (GitHub, Telegram, etc)
│   │   │   └── publisher.go             # Publish to Ethereum
│   │   ├── db/
│   │   │   ├── postgres.go              # Postgres connection pool
│   │   │   ├── clickhouse.go            # ClickHouse client
│   │   │   └── migrations/              # SQL migration files
│   │   ├── models/
│   │   │   ├── entities.go              # ORM / data models
│   │   │   └── events.go                # Event types
│   │   ├── config/
│   │   │   └── config.go                # Load env vars + config validation
│   │   ├── middleware/
│   │   │   ├── auth.go                  # Authentication + rate limiting
│   │   │   ├── logging.go               # Structured logging
│   │   │   └── errors.go                # Error handling
│   │   ├── temporal/
│   │   │   ├── activities.go            # Temporal activity definitions
│   │   │   ├── workflows.go             # Workflow definitions (genome analysis, etc)
│   │   │   └── client.go                # Temporal client setup
│   │   └── integration/
│   │       ├── ethereum.go              # Web3.go client for contract interactions
│   │       ├── ipfs.go                  # IPFS client
│   │       └── arweave.go               # Arweave client
│   │
│   ├── pkg/                             # Public Go packages (reusable)
│   │   ├── proof/
│   │   │   └── verifier.go              # Proof verification helpers
│   │   ├── crypto/
│   │   │   └── signatures.go            # ECDSA, BLS utilities
│   │   └── utils/
│   │       └── common.go                # Helpers
│   │
│   ├── crypto/                          # Rust workspace for cryptography
│   │   ├── Cargo.toml                   # Rust workspace root
│   │   ├── sentinel-crypto/             # Crate for core crypto operations
│   │   │   ├── Cargo.toml
│   │   │   ├── src/
│   │   │   │   ├── lib.rs               # Export public API
│   │   │   │   ├── proofs.rs            # Human-proof + Exploit-proof logic
│   │   │   │   ├── circuits.rs          # Noir circuit bindings
│   │   │   │   └── keys.rs              # Key derivation + management
│   │   │   └── tests/
│   │   │
│   │   └── zk-prover/                   # Crate for ZK proof generation
│   │       ├── Cargo.toml
│   │       ├── src/
│   │       │   ├── lib.rs
│   │       │   ├── human_proof.rs       # Human-behavior ZK proof generation
│   │       │   └── exploit_proof.rs     # Exploit existence proof generation
│   │       └── benches/
│   │
│   ├── tests/                           # Integration tests
│   │   ├── firewall_test.go
│   │   ├── oracle_test.go
│   │   └── integration_test.go
│   │
│   └── scripts/
│       ├── setup-db.sh                  # Initialize Postgres + ClickHouse
│       ├── seed-testdata.sh             # Populate test data
│       └── deploy.sh                    # Deployment script
│
├── contracts/                           # Solidity smart contracts
│   ├── package.json                     # npm for Foundry + dependencies
│   ├── foundry.toml                     # Foundry config
│   ├── Makefile                         # Common contract tasks
│   │
│   ├── src/
│   │   ├── IdentityFirewall.sol         # Main human-proof verifier
│   │   ├── RedTeamDAO.sol               # DAO for red-team + bounties
│   │   ├── ThreatOracle.sol             # Oracle signal consumer
│   │   ├── MalwareGenomeDB.sol          # On-chain genome hash storage
│   │   ├── ProofOfExploit.sol           # Exploit proof verifier + registry
│   │   ├── interfaces/
│   │   │   ├── IIdentityFirewall.sol
│   │   │   ├── IRedTeamDAO.sol
│   │   │   ├── IThreatOracle.sol
│   │   │   ├── IMalwareGenomeDB.sol
│   │   │   └── IProofOfExploit.sol
│   │   ├── libraries/
│   │   │   ├── VerificationHelpers.sol  # Common verification logic
│   │   │   ├── SafeMath.sol             # (or use newer OpenZeppelin)
│   │   │   └── Governance.sol           # DAO governance helpers
│   │   ├── mocks/
│   │   │   ├── MockThreatOracle.sol
│   │   │   └── MockToken.sol
│   │   └── upgradeable/
│   │       ├── IdentityFirewallV2.sol   # (for future versions)
│   │       └── ProxyAdmin.sol           # UUPS proxy admin
│   │
│   ├── test/
│   │   ├── IdentityFirewall.t.sol       # Foundry tests
│   │   ├── RedTeamDAO.t.sol
│   │   ├── ThreatOracle.t.sol
│   │   └── integration/
│   │       └── E2E.t.sol                # End-to-end scenario tests
│   │
│   ├── script/
│   │   ├── Deploy.s.sol                 # Deployment script (Foundry)
│   │   ├── UpgradeProxy.s.sol           # Upgrade script
│   │   └── LocalTestnet.s.sol           # Local test setup
│   │
│   └── .slitherignore                   # Slither config
│
├── circuits/                            # Noir ZK circuits
│   ├── Nargo.toml                       # Noir project config
│   ├── src/
│   │   ├── lib.nr                       # Export public circuit definitions
│   │   ├── human_proof.nr               # Behavioral entropy ZK circuit
│   │   ├── exploit_proof.nr             # Exploit existence ZK circuit
│   │   └── gadgets/
│   │       ├── hash.nr                  # Custom hash functions
│   │       ├── merkle.nr                # Merkle tree gadgets
│   │       └── signature.nr             # Signature verification
│   │
│   └── test/
│       ├── human_proof_test.nr
│       └── exploit_proof_test.nr
│
├── sdk/                                 # Rust core SDK (compiles to WASM)
│   ├── Cargo.toml                       # Rust SDK lib config
│   ├── src/
│   │   ├── lib.rs                       # Public API
│   │   ├── sentinel.rs                  # Sentinel SDK main logic
│   │   ├── behavior.rs                  # Local behavioral feature extraction
│   │   ├── proofs.rs                    # Proof generation / verification wrappers
│   │   ├── threat_oracle.rs             # Query threat signals
│   │   ├── contracts.rs                 # Ethers.rs wrapper for contract calls
│   │   └── wasm/
│   │       └── lib.rs                   # WASM entry points
│   │
│   ├── ts-bindings/                     # TypeScript wrappers (auto-generated)
│   │   ├── package.json
│   │   ├── src/
│   │   │   ├── index.ts                 # Main export
│   │   │   ├── types.ts                 # Type definitions
│   │   │   └── client.ts                # High-level client
│   │   └── tests/
│   │
│   └── python-bindings/                 # Python bindings (PyO3)
│       ├── pyproject.toml
│       ├── src/
│       │   ├── lib.rs                   # PyO3 glue code
│       │   └── bindings.rs              # Python function exports
│       └── tests/
│           └── test_sdk.py
│
├── ml/                                  # Python ML / analytics
│   ├── pyproject.toml                   # Poetry config
│   ├── poetry.lock                      # Locked dependencies
│   ├── Makefile                         # Common tasks
│   │
│   ├── src/
│   │   ├── vigilum_ml/
│   │   │   ├── __init__.py
│   │   │   ├── behavioral_model.py      # Identity Firewall model
│   │   │   ├── anomaly_detector.py      # On-chain anomaly detection
│   │   │   ├── genome_analyzer.py       # Malware genome clustering
│   │   │   ├── feature_engineering.py   # Feature extraction from traces
│   │   │   ├── models/
│   │   │   │   ├── __init__.py
│   │   │   │   ├── human_classifier.py  # Behavior classifier
│   │   │   │   └── anomaly_scorer.py    # Anomaly scoring
│   │   │   ├── utils/
│   │   │   │   ├── __init__.py
│   │   │   │   ├── trace_parser.py      # Parse transaction traces
│   │   │   │   ├── feature_cache.py     # Cache computed features
│   │   │   │   └── normalize.py         # Normalize raw data
│   │   │   └── inference/
│   │   │       ├── __init__.py
│   │   │       └── serve.py             # ONNX model serving
│   │
│   ├── notebooks/
│   │   ├── exploratory/
│   │   │   ├── behavioral_eda.ipynb     # Exploratory data analysis
│   │   │   └── genome_clustering.ipynb  # Genome representation experiments
│   │   └── experiments/
│   │       ├── train_human_classifier.ipynb
│   │       └── anomaly_threshold_tuning.ipynb
│   │
│   ├── data/
│   │   ├── raw/
│   │   │   ├── traces_sample.json       # Sample execution traces
│   │   │   └── exploits_sample.json     # Known exploits
│   │   └── processed/
│   │       └── .gitkeep
│   │
│   ├── tests/
│   │   ├── __init__.py
│   │   ├── test_behavioral_model.py
│   │   ├── test_anomaly_detector.py
│   │   └── test_feature_engineering.py
│   │
│   └── scripts/
│       ├── train_model.py               # Train classifier
│       ├── export_onnx.py               # Export to ONNX for Go service
│       └── evaluate.py                  # Model evaluation
│
├── infra/                               # Kubernetes + Terraform (later)
│   ├── k8s/
│   │   ├── namespace.yaml               # VIGILUM namespace
│   │   ├── backend/
│   │   │   ├── identity-firewall-deployment.yaml
│   │   │   ├── threat-oracle-deployment.yaml
│   │   │   └── api-gateway-deployment.yaml
│   │   ├── database/
│   │   │   ├── postgres-statefulset.yaml
│   │   │   └── clickhouse-statefulset.yaml
│   │   ├── services/
│   │   │   ├── redis-service.yaml
│   │   │   └── ipfs-service.yaml
│   │   ├── temporal/
│   │   │   ├── temporal-server-deployment.yaml
│   │   │   └── temporal-worker-deployment.yaml
│   │   ├── config/
│   │   │   ├── configmap.yaml
│   │   │   └── secrets.yaml              # (Never commit; use Sealed Secrets)
│   │   └── ingress/
│   │       └── ingress.yaml
│   │
│   ├── helm/                            # Helm charts (later)
│   │   └── vigilum-chart/
│   │       ├── Chart.yaml
│   │       ├── values.yaml
│   │       └── templates/
│   │
│   └── terraform/                       # IaC for cloud infra (later)
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       └── modules/
│           ├── networking/
│           ├── database/
│           └── compute/
│
├── tests/                               # System / integration tests
│   ├── e2e/
│   │   ├── test_human_proof_flow.go     # Full human-proof generation + verification
│   │   ├── test_exploit_reporting.go    # Full exploit reporting flow
│   │   ├── test_threat_oracle.go        # Threat signal publishing + consumption
│   │   └── test_self_healing.go         # Full detect→respond→heal
│   ├── load/
│   │   └── load_test.go                 # k6 / locust performance tests
│   └── chaos/
│       └── chaos_scenarios.yaml         # Chaos engineering tests (Kubernetes)
│
├── config/
│   ├── dev.env                          # Local development env vars
│   ├── staging.env                      # Staging environment
│   ├── prod.env                         # Production (secrets via Vault)
│   └── default.config.yaml              # Schema for all config options
│
├── scripts/
│   ├── setup-local-dev.sh               # One-command local dev setup
│   ├── deploy-testnet.sh                # Deploy to Sepolia
│   ├── deploy-prod.sh                   # Deploy to mainnet
│   └── monitor.sh                       # Check health of running services
│
├── logging/
│   └── .gitkeep                         # Logs stored here (not committed)
│
├── .gitignore
├── .env.example                         # Template for env vars
├── docker-compose.yml                   # (Alternative to k8s, for local dev)
├── Makefile                             # Root makefile with common tasks
│
└── (planning docs - existing)
    ├── architecture.md
    ├── techstack.md
    ├── requirements.md
    ├── roadmap.md
    ├── risks.md
    └── achieve.md
```

---

## Part 2: Phase-by-Phase Execution Checklist

### **PHASE 0: Foundation (Week 1–2)**
*Goal:* Set up repo structure, CI/CD, local dev environment.

#### 0.1 Repository Setup
- [ ] Create GitHub repo `vigilum` (or your provider)
- [ ] Set up main branch protection (require reviews + CI pass)
- [ ] Add .gitignore (Go, Rust, Python, Node, Solidity patterns)
- [ ] Add MIT or Apache 2.0 LICENSE file
- [ ] Create this folder structure locally (copy from Part 1)
- [ ] Commit initial structure with placeholder files

#### 0.2 CI/CD Pipeline (GitHub Actions)
- [ ] Create `.github/workflows/lint-test-build.yml`:
  - Go: `go fmt`, `go vet`, `golangci-lint`
  - Rust: `cargo fmt`, `cargo clippy`, `cargo test`
  - Python: `black`, `isort`, `pylint`, `pyright --strict`, `pytest`
- [ ] Create `.github/workflows/contract-test.yml`:
  - Run `forge test` for all Solidity
  - Run `slither` for static analysis
- [ ] Create `.github/workflows/security-scan.yml`:
  - `trivy` for container images
  - `cargo-audit` for Rust deps
  - `pip-audit` for Python deps
  - Dependabot / Renovate for auto-updates
- [ ] All workflows run on PR and merge to main

#### 0.3 Local Development Setup
- [ ] Create `Makefile` with targets:
  - `make setup` → installs all dependencies (Go, Rust, Python, Node, Foundry)
  - `make dev` → starts `docker-compose up` (postgres, redis, ipfs-node)
  - `make test` → runs all tests
  - `make lint` → runs all linters
- [ ] Create `.env.example` with all required env vars
- [ ] Create `docker-compose.yml` for local Postgres (15+), Redis (7+), IPFS Kubo node
- [ ] Create `setup-local-dev.sh` script:
  - Checks system dependencies
  - Copies `.env.example` to `.env`
  - Initializes databases
  - Seeds test data

#### 0.4 Documentation Skeleton
- [ ] README.md with:
  - Quick project overview (1 paragraph)
  - Tech stack summary
  - Quick-start guide (link to setup-local-dev.sh)
  - Folder structure overview
  - Contributing guidelines
- [ ] docs/CONTRIBUTING.md with:
  - Code style guide (Go, Rust, Python, Solidity)
  - Commit message convention
  - PR process
- [ ] docs/ARCHITECTURE.md (link to architecture.md)

#### 0.5 Commit & Checkpoint
- [ ] Commit: "feat: initial project structure + CI/CD pipelines"
- [ ] Verify all CI workflows pass on main branch

---

### **PHASE 1a: Identity Firewall Backend (Week 3–4)**
*Goal:* Identity Firewall service can accept and verify human-proofs locally.

#### 1a.1 Go Backend Structure
- [ ] Set up `backend/` Go module:
  - `go mod init github.com/vigilum/backend`
  - Add dependencies: `github.com/gin-gonic/gin`, `github.com/lib/pq`, `github.com/joho/godotenv`
- [ ] Create `backend/internal/config/config.go`:
  - Load DATABASE_URL, REDIS_URL, LOG_LEVEL, PORT from .env
  - Validate required vars on startup

#### 1a.2 Database Schema (Postgres)
- [ ] Create `backend/internal/db/migrations/001_init.sql`:
  - Table `users` (id, wallet_address, created_at)
  - Table `human_proofs` (id, user_id, proof_hash, verified, created_at, verified_at)
  - Table `threat_signals` (id, entity_address, signal_type, risk_score, updated_at)
- [ ] Create `backend/scripts/setup-db.sh`:
  - Runs migrations via sql-migrate or Flyway
  - Creates indexes
  - Seeds test data
- [ ] Run migrations locally

#### 1a.3 Identity Firewall Service
- [ ] Create `backend/cmd/identity-firewall/main.go`:
  - Starts HTTP server on PORT (default 8080)
  - Initializes DB connection pool
  - Sets up middleware (logging, error handling)
- [ ] Create `backend/internal/firewall/service.go`:
  - Method `VerifyProof(proofBytes []byte) (bool, error)`
  - Method `GenerateChallenge() (challenge string, error)`
  - Method `GetRiskScore(address string) (score float64, error)`
- [ ] Create `backend/internal/firewall/handlers.go`:
  - POST `/api/v1/firewall/verify-proof` → VerifyProof
  - GET `/api/v1/firewall/challenge` → GenerateChallenge
  - GET `/api/v1/firewall/risk/:address` → GetRiskScore
- [ ] Add structured logging (e.g., `github.com/sirupsen/logrus`)

#### 1a.4 Basic Proof Verification (Stub)
- [ ] For now: Accept any 32-byte proof, always return `true`
- [ ] Log proof received + verification event to Postgres
- [ ] Later: Replace with real Noir proof verification (Phase 2)

#### 1a.5 Tests & Commit
- [ ] Create `backend/tests/firewall_test.go`:
  - Test verify-proof endpoint
  - Test risk scoring
  - Mock DB calls
- [ ] Run `go test ./...` → all pass
- [ ] Commit: "feat(firewall): initial service + basic proof verification"

---

### **PHASE 1b: Sentinel SDK (Rust → WASM) (Week 5–6)**
*Goal:* Lightweight SDK that can run locally in wallets/dApps, extract features, stub proof generation.

#### 1b.1 Rust SDK Project
- [ ] Create `sdk/Cargo.toml`:
  - Add dependencies: `serde`, `web3`, `sha2`, `wasm-bindgen` (feature flag: `wasm`)
  - Set up workspace for `sdk` and `sdk/ts-bindings`
- [ ] Create `sdk/src/lib.rs`:
  - Export public API: `SentinelClient`, `ProofRequest`, `RiskSignal`

#### 1b.2 Behavioral Feature Extraction
- [ ] Create `sdk/src/behavior.rs`:
  - Struct `BehavioralFeatures`:
    - `transaction_count` (lifetime)
    - `avg_tx_interval` (seconds)
    - `gas_variance` (std dev)
    - `interaction_diversity` (number of unique contracts)
  - Function `extract_features(wallet_address) -> BehavioralFeatures`:
    - Query Ethereum RPC for last 100 txs of wallet
    - Compute features from traces
    - For MVP: hardcode some features or mock

#### 1b.3 WASM Bindings
- [ ] Create `sdk/src/wasm/lib.rs`:
  - `#[wasm_bindgen] pub async fn request_human_proof() -> ProofRequest`
  - `#[wasm_bindgen] pub fn get_risk_score(address: &str) -> f64`
- [ ] Build to WASM: `wasm-pack build --target web`
- [ ] Output goes to `sdk/pkg/` (add to .gitignore)

#### 1b.4 TypeScript Wrapper
- [ ] Create `sdk/ts-bindings/package.json` and TypeScript setup
- [ ] Create `sdk/ts-bindings/src/client.ts`:
  - Async function `requestHumanProof()`
  - Function `getRiskScore(address)`
  - Simple axios calls to backend Identity Firewall service
- [ ] Build: `npm run build` → produces JS module
- [ ] Publish stub to npm: `@vigilum/sentinel-sdk@0.0.1-alpha`

#### 1b.5 Tests & Commit
- [ ] Create `sdk/tests/test_sdk.rs`:
  - Test feature extraction
  - Test WASM initialization
- [ ] Run `cargo test` → all pass
- [ ] Commit: "feat(sdk): Rust WASM SDK + behavioral feature extraction"

---

### **PHASE 1c: Smart Contracts (Identity Firewall Verifier) (Week 7–8)**
*Goal:* On-chain verifier contract for human-proofs; testnet deployment.

#### 1c.1 Contract Scaffolding
- [ ] Create `contracts/foundry.toml`:
  - Set src = "src", test = "test"
  - Set solc version = "0.8.20"
- [ ] Add OpenZeppelin contracts: `forge install OpenZeppelin/openzeppelin-contracts`

#### 1c.2 Identity Firewall Contract
- [ ] Create `contracts/src/IdentityFirewall.sol`:
  ```solidity
  pragma solidity 0.8.20;
  
  contract IdentityFirewall {
      mapping(bytes32 => bool) public verifiedProofs;
      
      event ProofVerified(address indexed user, bytes32 proofHash);
      
      function verifyHumanProof(bytes calldata proof) public returns (bool) {
          // Stub: always accept for MVP
          bytes32 proofHash = keccak256(proof);
          verifiedProofs[proofHash] = true;
          emit ProofVerified(msg.sender, proofHash);
          return true;
      }
      
      function hasVerifiedProof(bytes32 proofHash) public view returns (bool) {
          return verifiedProofs[proofHash];
      }
  }
  ```

#### 1c.3 Tests
- [ ] Create `contracts/test/IdentityFirewall.t.sol`:
  - Test `verifyHumanProof()` accepts proof
  - Test event emission
  - Test proof lookup
- [ ] Run `forge test` → all pass

#### 1c.4 Testnet Deployment Script
- [ ] Create `contracts/script/Deploy.s.sol`:
  - Deploy IdentityFirewall to Sepolia
  - Output contract address
- [ ] Create `contracts/script/deploy.sh`:
  - Set PRIVATE_KEY + RPC_URL from env
  - Call `forge script Deploy.s.sol --broadcast --rpc-url $RPC_URL`
  - Verify contract on Etherscan
- [ ] Deploy to Sepolia testnet (manually)
- [ ] Save contract address to `IDENTITY_FIREWALL_ADDRESS` in `.env`

#### 1c.5 Commit
- [ ] Commit: "feat(contracts): IdentityFirewall verifier + Sepolia deployment"

---

### **PHASE 1d: Integration (Go ↔ Contract) (Week 9)**
*Goal:* Backend service calls Identity Firewall contract; end-to-end human-proof flow works.

#### 1d.1 Ethereum Client (Go)
- [ ] Create `backend/internal/integration/ethereum.go`:
  - Initialize `ethclient.Client` to Sepolia RPC
  - Load `IdentityFirewall` ABI
  - Method `CallVerifyProof(proof []byte) (bool, error)` → calls contract method

#### 1d.2 Update Firewall Service
- [ ] Modify `backend/internal/firewall/service.go`:
  - Add ethereum client dependency
  - `VerifyProof()` now calls contract via `eth.CallVerifyProof()`
  - Still logs event to Postgres

#### 1d.3 SDK Integration
- [ ] Update SDK `ts-bindings/src/client.ts`:
  - After user authorizes, SDK sends proof to backend
  - Backend verifies via contract
  - Backend returns result to SDK
  - SDK passes to wallet/dApp

#### 1d.4 E2E Test
- [ ] Create `tests/e2e/test_human_proof_flow.go`:
  - Generate mock proof
  - Call backend verify endpoint
  - Verify contract was called
  - Check Postgres log
- [ ] Run against local testnet (hardhat fork or Sepolia)

#### 1d.5 Commit
- [ ] Commit: "feat: end-to-end human-proof flow (SDK → Backend → Contract)"

---

### **PHASE 2a: Malware Genome Pipeline (Weeks 10–12)**
*Goal:* Ingest suspicious contracts, compute genomes, store on IPFS + on-chain hashes.

#### 2a.1 Genome Analyzer (Go)
- [ ] Create `backend/internal/genome/analyzer.go`:
  - Function `AnalyzeContract(bytecode []byte) (genome Genome, error)`
  - Extract features:
    - Opcode histogram
    - Call graph (if traces available)
    - Gas cost patterns
  - Return genome struct

#### 2a.2 Genome Hashing
- [ ] Create `backend/internal/genome/hasher.go`:
  - Function `HashGenome(genome) -> genomeHash`
  - Use SHA256(serialized genome)
  - Make deterministic

#### 2a.3 IPFS Storage
- [ ] Create `backend/internal/integration/ipfs.go`:
  - Initialize IPFS client to local node (or Infura)
  - Method `StoreGenome(genome) -> ipfsHash`
  - Upload serialized genome JSON

#### 2a.4 On-Chain Genome Registry
- [ ] Create `contracts/src/MalwareGenomeDB.sol`:
  ```solidity
  contract MalwareGenomeDB {
      mapping(bytes32 => GenomeRecord) public genomes;
      
      struct GenomeRecord {
          bytes32 genomeHash;
          string ipfsHash;
          uint256 timestamp;
          string label; // "known_exploit", "suspicious", etc
      }
      
      function registerGenome(bytes32 hash, string memory ipfsHash, string memory label) public {
          genomes[hash] = GenomeRecord(hash, ipfsHash, block.timestamp, label);
      }
  }
  ```

#### 2a.5 Backend Workflow
- [ ] Create Temporal workflow `backend/internal/temporal/workflows.go`:
  - `AnalyzeContractWorkflow()`:
    - Fetch contract bytecode
    - Compute genome
    - Store to IPFS
    - Call on-chain registry
    - Log to Postgres
- [ ] Create activity `backend/internal/temporal/activities.go` for each step

#### 2a.6 Tests & Commit
- [ ] Tests for genome extraction, hashing, IPFS upload
- [ ] Commit: "feat: malware genome pipeline (analysis → IPFS → on-chain)"

---

### **PHASE 2b: Threat Oracle Feeds (Weeks 13–14)**
*Goal:* Ingest public threat intel (GitHub PoCs, advisories), publish on-chain risk signals.

#### 2b.1 Feed Ingestion (Go)
- [ ] Create `backend/internal/oracle/feeds.go`:
  - Function `FetchGitHubPoCs() -> []Exploit`
  - Poll GitHub API for new repo creation patterns (e.g., "exploit for contract X")
  - Scrape public advisory databases (e.g., blockchain-specific security lists)
  - Parse + normalize to common schema

#### 2b.2 Signal Aggregation
- [ ] Create `backend/internal/oracle/aggregator.go`:
  - Function `AggregateSignals(feeds []FeedEvent) -> ThreatSignal`
  - Combine multiple feed signals into a single risk score
  - Example: if 3+ sources report exploit for contract X, mark as HIGH_RISK

#### 2b.3 On-Chain Oracle Contract
- [ ] Create `contracts/src/ThreatOracle.sol`:
  ```solidity
  contract ThreatOracle {
      mapping(address => uint256) public riskScores; // 0-100
      
      event RiskUpdated(address indexed target, uint256 riskScore);
      
      function updateRiskScore(address target, uint256 score) public onlyOracle {
          require(score <= 100);
          riskScores[target] = score;
          emit RiskUpdated(target, score);
      }
  }
  ```

#### 2b.4 Oracle Publisher (Go)
- [ ] Create `backend/internal/oracle/publisher.go`:
  - Temporal activity to call `updateRiskScore()` on chain
  - Rate-limit updates (e.g., once per hour per address)
  - Log all updates to Postgres

#### 2b.5 Tests & Commit
- [ ] Test feed ingestion, aggregation, on-chain publishing
- [ ] Commit: "feat: threat oracle with feed ingestion + on-chain signals"

---

### **PHASE 3a: ZK Circuits (Noir) – Human-Proof (Weeks 15–17)**
*Goal:* Replace stub proofs with real ZK human-proof circuit.

#### 3a.1 Circuit Design (Research)
- [ ] Document in `docs/ZK_CIRCUITS.md`:
  - Goal: prove "this wallet exhibits human-like transaction behavior"
  - Public inputs: wallet address
  - Private inputs: last 20 tx timestamps, gas values
  - Constraint: variance of tx times > threshold AND gas variance < threshold (heuristic)
- [ ] Review with ZK expert (e.g., from Noir team)

#### 3a.2 Noir Circuit Implementation
- [ ] Create `circuits/src/human_proof.nr`:
  - Expose `main()` function with constraints
  - Input: timestamps, gas values
  - Compute: mean, variance
  - Assert: variance > MIN_VARIANCE && variance < MAX_VARIANCE
- [ ] Create simple test fixtures

#### 3a.3 Rust Proof Generator
- [ ] Create `backend/crypto/zk-prover/src/human_proof.rs`:
  - Function `generate_human_proof(features: BehavioralFeatures) -> (proof, publicInputs)`
  - Calls Noir prover binary (or library)
  - Returns serialized proof

#### 3a.4 On-Chain Verifier
- [ ] Create Solidity verifier from Noir:
  - `nargo codegen-verifier` → outputs Solidity verifier contract
  - Integrate into `IdentityFirewall.sol`
  - `verifyHumanProof()` now calls real ZK verifier

#### 3a.5 End-to-End Test
- [ ] Create `tests/e2e/test_zk_human_proof.go`:
  - Generate proof locally
  - Send to backend
  - Backend verifies
  - Backend calls on-chain verifier
  - All checks pass

#### 3a.6 Commit
- [ ] Commit: "feat: Noir ZK circuit for human-proof + on-chain verifier"

---

### **PHASE 3b: Proof-of-Exploit Engine (Weeks 18–20)**
*Goal:* Researchers can submit ZK proofs of exploits; Red-Team DAO distributes rewards.

#### 3b.1 Circuit Design (Research)
- [ ] Document in `docs/ZK_CIRCUITS.md`:
  - Goal: prove "this execution trace violates property P of contract X"
  - Public inputs: contract address, property description
  - Private inputs: execution trace, state transition
  - Constraint: trace satisfies exploit preconditions

#### 3b.2 Noir Circuit
- [ ] Create `circuits/src/exploit_proof.nr`:
  - Accept trace as input
  - Assert specific state transition occurred
  - Compute proof hash

#### 3b.3 On-Chain Exploit Registry
- [ ] Create `contracts/src/ProofOfExploit.sol`:
  ```solidity
  contract ProofOfExploit {
      struct ExploitSubmission {
          bytes32 proofHash;
          address target;
          uint256 timestamp;
          string description;
          bool verified;
          uint256 bountyAmount;
      }
      
      mapping(bytes32 => ExploitSubmission) public submissions;
      
      function submitExploitProof(
          bytes calldata proof,
          address target,
          string memory description
      ) public returns (bytes32) {
          // Verify ZK proof
          // Check novelty vs MalwareGenomeDB
          // Create submission
          // Store genome
      }
  }
  ```

#### 3b.4 Red-Team DAO Contract
- [ ] Create `contracts/src/RedTeamDAO.sol`:
  - Staking mechanism (researchers deposit reputation tokens)
  - Reward distribution based on proof validity + impact
  - Slashing for duplicate / invalid proofs
  - Governance (parameters, approved bug classes)

#### 3b.5 Backend Integration
- [ ] Create `backend/internal/redteam/processor.go`:
  - Temporal workflow to process exploit submissions
  - Verify proof (off-chain re-execution check)
  - Calculate impact score
  - Recommend bounty amount
  - Trigger DAO payout

#### 3b.6 Tests & Commit
- [ ] Test proof submission, verification, bounty calculation
- [ ] Commit: "feat: Proof-of-Exploit engine + Red-Team DAO"

---

### **PHASE 4a: Self-Healing Hooks (Weeks 21–23)**
*Goal:* Protocols can integrate defensive patterns (pause, throttle, key-rotate) triggered by threat signals.

#### 4a.1 Library of Defensive Patterns
- [ ] Create `contracts/src/libraries/DefensivePatterns.sol`:
  - Modifier `whenNoExploit()` → checks ThreatOracle
  - Modifier `whenAdminNotCompromised()` → checks key-leak signal
  - Circuit breaker helper functions

#### 4a.2 Example Partner Integration Contract
- [ ] Create `contracts/src/examples/ExampleProtocol.sol`:
  - Minimal DEX or token contract
  - Uses `whenNoExploit()` modifier on swap function
  - References ThreatOracle for signal
  - Tests demonstrate auto-pause on exploit detection

#### 4a.3 Documentation
- [ ] Create `docs/INTEGRATION_GUIDE.md`:
  - Step-by-step for protocols to integrate VIGILUM
  - Code examples
  - Gas cost estimates
  - Failure modes + recommendations

#### 4a.4 E2E Self-Healing Test
- [ ] Create `tests/e2e/test_self_healing.go`:
  - Simulate exploit detection
  - Threat Oracle publishes HIGH_RISK signal
  - Protected contract pauses
  - After signal cleared, protocol resumes

#### 4a.5 Commit
- [ ] Commit: "feat: self-healing defensive patterns + integration guide"

---

### **PHASE 4b: ML Models in Production (Weeks 24–26)**
*Goal:* Trained human-behavior classifier deployed in backend; anomaly detection live.

#### 4b.1 Data Collection
- [ ] `ml/scripts/collect_data.py`:
  - Query Ethereum for wallet behavioral traces
  - Extract BehavioralFeatures for 10k+ wallets
  - Label as "human" (heuristic: high tx count + diverse contracts) or "bot"
  - Save to `ml/data/raw/`

#### 4b.2 Model Training
- [ ] `ml/src/vigilum_ml/models/human_classifier.py`:
  - Train RandomForest or XGBoost on features
  - 80/20 train/test split
  - Evaluate: precision, recall, F1
  - Store model to Git LFS

#### 4b.3 ONNX Export
- [ ] `ml/scripts/export_onnx.py`:
  - Convert trained model to ONNX format
  - Commit to repo for reproducibility

#### 4b.4 Go Inference Service
- [ ] `backend/internal/inference/serve.go`:
  - Use `ort` crate to load ONNX model
  - Expose function `PredictHumanScore(features) -> float64 [0, 1]`
  - Cache model in memory

#### 4b.5 Integration with Identity Firewall
- [ ] Update `backend/internal/firewall/service.go`:
  - For each proof request, run inference on user features
  - Only allow proof if score > 0.7 (configurable threshold)
  - Log score + decision to Postgres

#### 4b.6 Tests & Commit
- [ ] Test model prediction, inference latency
- [ ] Commit: "feat: ML-powered identity verification in production"

---

### **PHASE 5: Hardening & Testnet Launch (Weeks 27–30)**

#### 5.1 Security Audit Prep
- [ ] Run all security tools:
  - Go: golangci-lint, gosec
  - Rust: cargo-audit, clippy
  - Solidity: Slither + Mythril
  - Container: Trivy
- [ ] Fix high/critical issues
- [ ] Document known issues and mitigation

#### 5.2 Performance Optimization
- [ ] Benchmark identity-firewall service (target: <50ms p99 for verify)
- [ ] Benchmark ML inference (target: <10ms)
- [ ] Profile + optimize if needed

#### 5.3 Documentation
- [ ] Finalize:
  - README.md
  - docs/ARCHITECTURE.md
  - docs/API_REFERENCE.md (OpenAPI spec)
  - docs/DEPLOYMENT.md
  - docs/CONTRIBUTING.md

#### 5.4 Testnet Deployment
- [ ] Deploy all services to staging K8s cluster:
  - Identity Firewall service
  - Threat Oracle service
  - All contracts to Sepolia
  - IPFS node
  - Postgres + ClickHouse
- [ ] Run canary tests
- [ ] Monitor for 48+ hours

#### 5.5 Launch Announcement
- [ ] Blog post: "VIGILUM Testnet Alpha Released"
- [ ] GitHub discussions open for feedback

---

## Part 3: File Creation Checklist

### Essential Files to Create First
- [ ] README.md (root)
- [ ] .gitignore
- [ ] LICENSE
- [ ] Makefile (root)
- [ ] docker-compose.yml
- [ ] .env.example
- [ ] backend/go.mod
- [ ] backend/cmd/identity-firewall/main.go
- [ ] backend/internal/config/config.go
- [ ] backend/internal/db/migrations/001_init.sql
- [ ] contracts/src/IdentityFirewall.sol
- [ ] contracts/foundry.toml
- [ ] sdk/Cargo.toml
- [ ] ml/pyproject.toml
- [ ] .github/workflows/lint-test-build.yml

### Then Add Per Phase
- [ ] Phase 1a: database setup, firewall service, handlers
- [ ] Phase 1b: SDK structure, behavior extraction
- [ ] Phase 1c: contracts, tests, deploy scripts
- [ ] Phase 2a: genome analyzer, IPFS integration, Temporal workflows
- [ ] Phase 2b: threat oracle feeds
- [ ] Phase 3a: Noir circuits for human-proof
- [ ] Phase 3b: exploit proof engine, RedTeamDAO contract
- [ ] Phase 4a: defensive patterns, integration guide
- [ ] Phase 4b: ML training + inference

---

## Part 4: Success Criteria Per Phase

### Phase 1 (MVP)
- ✅ Human-proof flow: SDK generates stub proof → Backend verifies → Contract accepts
- ✅ All services run locally via docker-compose
- ✅ 90% test coverage for critical functions
- ✅ Deployed to Sepolia; manual testing successful

### Phase 2
- ✅ Genome hashing + IPFS storage working
- ✅ Threat Oracle publishes signals to contract
- ✅ E2E test demonstrates threat signal → contract state change

### Phase 3
- ✅ Real ZK proofs generated + verified
- ✅ Exploit-proof engine accepts researcher submissions
- ✅ Red-Team DAO distributes test payouts on testnet

### Phase 4
- ✅ Defensive patterns library usable by external protocols
- ✅ ML model in-service with >85% accuracy
- ✅ 2+ pilot protocols integrated

### Phase 5
- ✅ All security audits cleared
- ✅ P99 latency <100ms for all critical paths
- ✅ 48h testnet uptime with no major incidents
- ✅ 100+ community members on Discord/forum

---

## Part 5: Key Dependencies & Versions

**Go**
- Go 1.22+
- Gin v1.9+
- ethers-go / web3.go latest
- Temporal v1.20+

**Rust**
- Rust 1.70+
- Noir latest (from Aztec Labs)
- wasm-bindgen 0.2.90+
- web3 0.20+

**Python**
- Python 3.11+
- PyTorch 2.0+
- Polars 0.19+
- pydantic 2.0+

**Solidity**
- 0.8.20+
- OpenZeppelin 5.0+

**Infra**
- Postgres 15+
- ClickHouse 23+
- Redis 7+
- Docker 24+
- Kubernetes 1.27+

---

## Part 6: Common Pitfalls & Mitigations

| Pitfall                          | Mitigation                                                 |
| -------------------------------- | ---------------------------------------------------------- |
| Scope creep in Phase 1           | Stick to stub proofs; delay ML models until Phase 4b       |
| ZK circuit complexity            | Start with simple behavioral constraints; iterate after MVP |
| Ethereum gas costs               | Use Sepolia for testing; optimize contracts before mainnet |
| IPFS reliability                 | Use Infura or Pinata for backup; don't rely on single node |
| Secrets in repo                  | Use `.env.example`; Vault for prod; Sealed Secrets in K8s  |
| Tight coupling (Go ↔ Rust)       | Define gRPC or HTTP boundaries early                       |
| Test flakiness                   | Use testcontainers for DB tests; mock Ethereum calls       |
| Documentation lag                | Tie docs to code reviews; require doc updates per PR        |

---

## Part 7: Tools & Commands Reference

### Local Development
```bash
# Setup
./scripts/setup-local-dev.sh

# Run services
make dev

# Run tests
make test

# Deploy to Sepolia
./scripts/deploy-testnet.sh

# Monitor
./scripts/monitor.sh
```

### CI/CD
```bash
# Lint
cargo fmt --check && cargo clippy
go fmt ./... && golangci-lint run
black --check . && isort --check .

# Test
cargo test
go test ./...
pytest

# Deploy
# (triggered on merge to main after reviews)
```

### Contract Interactions
```bash
# Deploy
cd contracts && forge script script/Deploy.s.sol --broadcast --rpc-url $RPC_URL

# Test
forge test

# Verify on Etherscan
forge verify-contract --etherscan-api-key $ETHERSCAN_KEY 0x... src/IdentityFirewall.sol:IdentityFirewall
```

---

## Next: Implement Phase 0 Now

1. Create GitHub repo `vigilum`
2. Copy folder structure from Part 1
3. Create CI/CD workflows from Part 2
4. Run `make setup && make dev`
5. All tests pass locally
6. Commit: "feat: project foundation + CI/CD"

Then move to Phase 1a.

**Questions to clarify before proceeding:**
- Which chain to target first: Ethereum Sepolia or Arbitrum?
- Which cloud provider: AWS, GCP, or self-hosted?
- Team size and timeline: 1 person (3 months) or 5 people (6 weeks)?

