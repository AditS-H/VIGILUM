# VIGILUM Tech Stack (2026 Production-Grade)

Goal: **Fast market deployment + production-hardened** from day one. No tech debt from "MVP in Python that we'll rewrite later."

---

## 1. Languages & Core Runtimes (Market-Optimized)

| Layer / Component        | Primary              | Alternative / Notes                                                        |
| ------------------------ | -------------------- | -------------------------------------------------------------------------- |
| Kernel / monitoring      | Rust + C             | eBPF for tracing; Rust for safety-critical paths                          |
| Backend services         | **Go** (primary)     | Fast, concurrent, great stdlib; Rust for cryptography/ZK-adjacent code   |
| ML & threat analysis     | Python (hardened)    | PyTorch + Polars (faster than Pandas); typing via pydantic v2             |
| Smart contracts          | Solidity (EVM)       | Vyper if we need security-first; multi-chain via abstract interfaces      |
| ZK circuits              | **Noir** (primary)   | Circom as fallback; Noir matures faster, better debugging, Rust-like      |
| Storage (immutable)      | IPFS + Arweave       | Content-addressed; consider Bunny CDN for hot path                        |
| Client SDKs              | **Rust** (primary)   | TypeScript for Web; Python bindings via PyO3; wasm for browsers           |

**Strategic choice:** Rust for crypto/ZK-adjacent, Go for services, Python for research/ML with hardening.

---

## 2. Backend & Services (Go-First, Rust for Security)

**Core services** (Go):
- Identity Firewall verifier service (`axum` web framework):
  - High-throughput proof verification.
  - Sub-50ms p99 latency target.
- Threat Oracle aggregator:
  - Feed ingestion, signal scoring, publishing.
- API gateway / load balancer.

**Security-critical services** (Rust):
- ZK proof generator for Human-Proof and Exploit-Proof.
  - Uses `noir-lang` + custom gadgets.
  - Isolated, containerized, rate-limited.
- Cryptographic key manager (secrets, signing).
- Anomaly detection engine (performance + safety).
 (Battle-Tested Stack)

- **Relational DB**:
  - **Postgres 15+** (primary):
    - All core entities: users, protocols, proofs, bounties, genomes.
    - JSONB for flexibility, native array types.
  - Connection pooling: `pgbouncer` or `sqlc` (from Go).
  
- **Analytics / time-series DB**:
  - **ClickHouse** for:
    - High-volume events (mempool traces, behavior samples).
    - Sub-second queries on billions of rows.
  - Bonus: Excellent ZK circuit cost tracking.

- **Cache layer**:
  - Redis 7+ for: (Modern Python + Rust)

- **Python (hardened)** for research and feature engineering:
  - `PyTorch 2.0+` for behavioral models (compile to Triton for inference).
  - `Polars` (not Pandas) for feature pipelines: 10x faster, better memory.
  - `networkx` + `rustworkx` for call graph analysis.
  - `scikit-learn` for baseline classifiers, but migrate to `tch-rs` (PyTorch Rust bindings) for production inference.

- **Rust inference** (production):
  - `ort` crate for ONNX model serving (convert PyTorch → ONNX).
  - Deploy as a microservice or embedded in Go services.
  - Sub-10ms inference latency for anomaly detection.

- **Experiment tracking**:
  - Weights & Biases (external) for rapid experimentation.
  - Commit best models to Git LFS + IPFS for reproducibility.

- **Feature stores**:
  - In-house Postgres + materialized views (MVP).
  - Upgrade to Feast or Tecton if feature complexity explodeZK proofs.
  - Arweave for long-term audit logs (compliance, evidence).
  - Local S3-compatible (MinIO) for internal backup

## 3. Data & Storage

- **Relational / analytics DB**:
  - Postgres for core (2026 Standards)

- **Chains (MVP target)**:
  - **Ethereum Sepolia** (testnet) → Ethereum (mainnet).
  - Arbitrum as tier-2 (lower costs, high throughput).
  - Multi-chain abstraction layer (provider pattern) for future Polygon, Optimism, etc.

- **Smart contract tooling**:
  - **Foundry** (primary): Faster tests, Rust-based, better DX than Hardhat.
  - Solidity 0.8.20+ (latest security patches).
  - Contracts:
    - `IdentityFirewall.sol` (verifier for human-proofs).
    - `RedTeamDAO.sol` (staking, reputation, rewards).
    - `ThreatOracle.sol` (signal consumer).
  - Upgradeable via UUPS pattern (not transparent proxy; cleaner).

- **ZK stack** (Noir primary):
  - **Noir** for:
    - Human-behavior ZK proofs (behavioral entropy circuit).
    - Exploit-existence proofs (custom gadgets).
  - Backend: barretenberg (Go prover; Aztec maintained).
  - Solidity SDKs (Rust-First + Polyglot)

- **Primary SDK** (Rust + WASM):
  - Core library in Rust; compiles to WebAssembly for browsers.
  - `wasm-bindgen` for TypeScript/JavaScript bindings.
  - Published as `@vigilum/sentinel-sdk` on npm.
  - Features:
    - Request and generate human-proofs locally (if behavioral data available).
    - Query Threat Oracle risk scores.
    - Sign and submit Proof-of-Exploit (for researchers).

- **TypeScript wrapper** (thin layer over Rust WASM):
  - Async/await API, type-safe.
  - Works in browsers and Node.js.

- **Python bindings** (PyO3):
  - `vigilum-sentinel` on PyPI.
  - For bots, data scientists, research workflows.
  - Lower performance; focus on ease of use.

- **Go client library** (in-repo):
  - For backend-to-backend integrations.
  - Used by Temporal job runners, test harn
---

## 5. Blockchain & ZK

- **Chains (initial)**:
  - One EVM-compatible testnet/mainnet (to be decided).
  - Support multi-chain by abstracting provider layer (ethers.js / web3.py).

- **Smart contract tooling**: (Cloud-Native, 2026)

- **Containerization**:
  - Docker (OCI standard).
  - Multi-stage builds: slim Go binaries (10–50 MB), optimized Python layers.

- **Orchestration**:
  - **Kubernetes** (EKS / GKE / AKS):
    - Helm for templating.
    - ArgoCD for GitOps-style deployments.
  - StatefulSets for IPFS nodes, Postgres replicas.

- **CI/CD**:
  - GitHub Actions (primary):
    - Lint, test, security s(Production-Grade)

- **Solidity static analysis**:
  - Slither + custom checks (via CI).
  - Mythril for advanced symbolic analysis (critical functions only).

- **Rust / Go security**:
  - `cargo-audit` (Rust dependencies).
  - Golangci-lint + gosec (Go).
  - `trivy` for container image scanning.

- **Python security**:
  - Bandit for common issues.
  - Type checking via Pyright (strict mode); pydantic v2 for validation.
  - Poetry for reproducible dependency locks.

- **Dependency management**:
  - Renovate for automated PR suggestions (all languages).
  - Weekly security audits of critical dependencies.

- **Secrets & access**:
  - Vault or cloud KMS (never in repos).
  - RBAC + audit logging for all access.
  - Automated rotation of API keys / signing keys.

- **Compliance (future)**:
  - SOC 2 readiness (logging, access controls, incident response).
  - GDPR considerations (minimal PII, data retention policies).

---

## 9. Market-Aligned Summary (Why This Stack)

| Why                          | What We Chose                                     |
| ---------------------------- | ------------------------------------------------- |
| **Speed to market**          | Go for 90% of backend; Rust for 10% security-critical. |
| **Scalability without pain** | Postgres + ClickHouse, Kubernetes, async (Temporal). |
| **Crypto + ZK maturity**     | Noir (emerging standard), Solidity (battle-tested). |
| **Security from day 1**      | Rust for secrets, Foundry for contracts, Vault for keys. |
| **Polyglot SDKs**            | Rust core → WASM (JS), Python, Go bindings. |
| **Modern DX**                | Foundry, cargo, uv (Python package manager). |
| **Observability + reliability** | Temporal for workflows, K8s for orchestration, Prometheus for signals. |

---

## 10. Open Decisions

- **Cloud provider**: AWS, GCP, or self-hosted? (Affects Vault choice, monitoring.)
- **IPFS strategy**: Full Kubo nodes vs light clients? (Cost vs decentralization trade-off.)
- **Polygon / Arbitrum prioritization**: MVP on Ethereum only, then multi-chain? (Time / complexity trade-off.)
## 6. Client / SDK

- **Languages**:
  - TypeScript for web dApps.
  - Python for scriptable integrations.
- **Packaging**:
  - npm package (`@vigilum/sentinel-sdk`) for front-end / Node.
  - PyPI package (`vigilum-sentinel`) for bots, research tools.

Features:
- Simple APIs to:
  - Request and attach human-proof to tx.
  - Pull latest risk scores / threat signals for addresses.

---

## 7. Infrastructure & DevOps

- Containerization:
  - Docker for all services.
- Orchestration (later phases):
  - Kubernetes or Nomad.
- CI/CD:
  - GitHub Actions (tests, lint, security scans, contract deployment).
- Observability:
  - Prometheus + Grafana for metrics.
  - Loki / OpenSearch for logs.

---

## 8. Security & Compliance Tooling

- Static analysis:
  - Slither / Mythril for Solidity.
  - Bandit / mypy for Python.
- Dependency security:
  - Dependabot / pip-audit.
- Secret management:
  - Environment variables + secret managers (HashiCorp Vault / cloud KMS).

---

## 9. Open Tech Decisions (To Be Decided)

- Which EVM chain to target first (for MVP experiments).
- Exact ZK stack (Circom vs Noir vs others) based on dev experience.
- Where to draw the line between Python vs Go/Rust for performance-critical components.
- Choice of cloud provider (or multi-cloud) and decentralization strategy for off-chain infra.
