# VIGILUM Architecture

## 1. Purpose

Define how VIGILUM wraps existing blockchains with an "immune system": detect → prove → respond → heal → learn → evolve, without replacing the underlying chains.

---

## 2. Top-Level View

**Actors**
- Users / wallets / DAOs
- Protocols / dApps
- Red-team researchers
- Off-chain observers (threat intel, scanners)

**Layers**
- Sentinel SDK (client-side / off-chain)
- VIGILUM Network (coordination + intelligence)
  - Identity Firewall
  - Malware Genome DB
  - Proof-of-Exploit Engine
  - Threat Oracle Layer
  - Red-Team DAO
- Protected smart contracts (integrated protocols)

**High-Level Flow (conceptual)**

```text
User / Wallet / DAO
        ↓
Sentinel SDK (local + off-chain)
        ↓
VIGILUM Network
 ├─ Identity Firewall
 ├─ Malware Genome DB
 ├─ Exploit Proof Engine
 ├─ Threat Oracle Layer
 ├─ Red-Team DAO
        ↓
Protected Smart Contracts
```

---

## 3. Core Components

### 3.1 Sentinel SDK
- Runs locally with wallet / dApp.
- Extracts behavioral features (timing, gas patterns, interaction cadence).
- Builds non-identifying feature vectors.
- Interacts with Identity Firewall and Threat Oracle APIs.
- Provides hooks for dApps / wallets to:
  - Request human-proof.
  - Query current threat level for addresses / contracts.

### 3.2 Identity Firewall
- Off-chain engine + on-chain verification contracts.
- Maintains models to classify "human-like" vs "bot / sybil" behavior.
- Generates Zero-Knowledge Proofs (ZKPs) of human-like behavior without leaking raw features.
- On-chain `Sentinel` contract exposes `verifyHumanProof(proof)`.

### 3.3 Malware Genome DB
- Off-chain threat-analysis pipeline that:
  - Sandboxes suspicious contracts / transactions.
  - Extracts opcode-level, call-graph, and gas-profile fingerprints.
  - Normalizes into a "genome" representation.
- Stores immutable genome hashes on-chain (IPFS / Arweave for raw data).
- Exposes query APIs:
  - "Have we seen this genome or close variants before?"
  - "What impact class is associated?"

### 3.4 Proof-of-Exploit Engine
- Allows security researchers to prove existence of a reproducible exploit **without** revealing exploit details.
- Workflow:
  - Researcher defines exploit scenario off-chain.
  - Engine generates a ZK proof attesting:
    - A state-transition exists that violates property P of contract X.
    - Exploit is reproducible given public contract state.
  - No exploit payload or PoC bytecode is revealed.
- On-chain verifier contract checks:
  - Validity of proof.
  - Uniqueness / novelty (against Malware Genome and previous proofs).
  - Impact level (low / medium / critical).

### 3.5 Threat Oracle Layer
- Aggregates signals from:
  - Dark web / leak forums (off-chain scrapers).
  - GitHub PoCs and disclosures.
  - Mempool anomaly detectors.
  - On-chain events from Malware Genome and Proof-of-Exploit.
- Produces **actionable on-chain signals**, e.g.:
  - `admin_key_leaked` (boolean or risk score).
  - `exploit_campaign_detected` for certain protocol classes.
- Feeds into:
  - Protected contracts (governance freeze, rate limiting, circuit breakers).
  - Sentinel SDK (for UI and warnings).

### 3.6 Red-Team DAO
- Governance and incentive layer for continuous attack simulation.
- Tracks:
  - Researcher identities (pseudonymous, reputation-based).
  - Stakes and slashing events.
  - Submitted Proof-of-Exploit artifacts.
- Controls:
  - Bounty rates and reward curves.
  - Access levels to more sensitive tooling / simulations.

### 3.7 Protected Smart Contracts
- Integrate VIGILUM as a security dependency, not as a fork.
- Examples:
  - `require(Sentinel.verifyHumanProof(proof));`
  - `if (ThreatOracle.adminKeyLeaked(protocolId)) freeze_governance();`
  - `if (ThreatOracle.exploitRisk(contract) > RISK_THRESHOLD) { pause(); }`

---

## 4. Data Flows (First Draft)

### 4.1 Identity Firewall Flow
1. User interacts with a VIGILUM-enabled wallet / dApp.
2. Sentinel SDK collects behavioral traces locally.
3. Local model derives a feature vector and generates a ZKP of "human-likeness".
4. dApp submits proof with transaction.
5. On-chain Identity Firewall contract verifies proof and gates access.

### 4.2 Exploit Reporting Flow
1. Researcher discovers potential exploit against contract X.
2. Off-chain environment simulates exploit and computes malware genome.
3. Proof-of-Exploit Engine generates:
   - ZK proof of exploit.
   - Genome hash.
4. Researcher submits proof + genome to Red-Team DAO contract.
5. DAO verifies, checks novelty, and triggers payout.
6. Malware Genome DB updated; Threat Oracle pushes alerts to protocols.

### 4.3 Self-Healing Loop Flow
1. Anomaly detector flags suspicious cluster of transactions.
2. Malware Genome pipeline classifies pattern as known / unknown exploit.
3. Threat Oracle raises risk signal for affected protocols.
4. Protected contracts:
   - Freeze / throttle actions.
   - Rotate keys if oracle indicates key compromise.
   - Trigger governance emergency mode.
5. Once patch is deployed and new genome is registered, defenses propagate to other protocols using similar patterns.

---

## 5. Deployment & Topology (Initial Assumptions)

- **Blockchains supported**: start with one EVM chain (e.g., testnet) → expand to multi-chain.
- **Off-chain infrastructure**:
  - Microservices or modular monolith for early MVP.
  - Containerized (Docker) and orchestrated (Kubernetes) in later phases.
- **Data storage**:
  - Hot path: Postgres / ClickHouse for events and features.
  - Cold / immutable: IPFS + Arweave for genome records and evidence.
- **ZK infrastructure**:
  - Dedicated provers (Circom / Noir) running in secure environments.
  - On-chain verifier contracts for each proof system.

---

## 6. Trust & Threat Model (Sketch)

- Assume:
  - Attackers are rational, adaptive, and well-funded.
  - Insider threats and collusion attempts exist.
  - On-chain state is public; off-chain signals may be manipulated.
- Minimize trust by:
  - Making proofs verifiable on-chain.
  - Staking + slashing in Red-Team DAO.
  - Multiple independent oracles and data sources.

Open questions (to refine later):
- How decentralized should off-chain components be in early versions?
- How to guard against oracle manipulation and censorship?
- How to version and deprecate old genome schemas?

---

## 7. Open TODOs for Architecture

- Formalize security properties for each component (e.g., Identity Firewall soundness / completeness trade-offs).
- Decide boundary between on-chain vs off-chain logic per feature.
- Design minimal viable architecture for the **MVP phase** (likely just Identity Firewall + basic threat memory).
- Add sequence diagrams for 3–4 critical flows once we lock terminology.
