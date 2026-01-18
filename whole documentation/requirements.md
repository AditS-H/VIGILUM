# VIGILUM Requirements (Draft)

This document captures **functional** and **non-functional** requirements for VIGILUM, especially for early phases.

---

## 1. Stakeholders

- Protocol / DAO teams
- Security researchers / red-teamers
- End users (wallet holders, dApp users)
- Infrastructure providers / node operators

---

## 2. Functional Requirements (MVP-Oriented)

### 2.1 Identity Firewall
- FR-IF-1: Wallets / dApps MUST be able to request a human-proof for a given address/session.
- FR-IF-2: System MUST generate a proof that can be verified on-chain without revealing raw behavioral data.
- FR-IF-3: On-chain contract MUST expose a function like `verifyHumanProof(proof)` that returns true/false.
- FR-IF-4: Protocols MUST be able to integrate checks with minimal code changes (modifiers / helper functions).

### 2.2 Malware Genome
- FR-MG-1: System MUST ingest contract bytecode and/or transaction traces.
- FR-MG-2: System MUST produce a deterministic genome representation from an execution trace.
- FR-MG-3: System MUST store genome hashes immutably and allow lookups.
- FR-MG-4: System SHOULD support similarity queries ("have we seen something like this?").

### 2.3 Proof-of-Exploit Engine
- FR-PE-1: Researchers MUST be able to define exploit scenarios offline.
- FR-PE-2: Engine MUST generate a proof that an exploit exists for a target contract and property.
- FR-PE-3: On-chain verifier MUST check validity without needing exploit details.
- FR-PE-4: Engine MUST integrate with Red-Team DAO for rewards and tracking.

### 2.4 Threat Oracle Layer
- FR-TO-1: System MUST ingest external signals (e.g., feeds, advisories, mempool anomalies).
- FR-TO-2: Oracles MUST write aggregated risk signals on-chain periodically.
- FR-TO-3: Protocols MUST be able to query risk signals on-chain in O(1) calls.
- FR-TO-4: Signals MUST be versioned and explainable at least at a coarse level.

### 2.5 Red-Team DAO
- FR-RD-1: Researchers MUST be able to stake and participate.
- FR-RD-2: DAO MUST track reputation and slash bad behavior.
- FR-RD-3: DAO MUST distribute rewards according to impact and novelty.
- FR-RD-4: Governance MUST be upgradeable (safely) for parameters and rules.

---

## 3. Non-Functional Requirements

### 3.1 Security
- NFR-S-1: Off-chain infrastructure MUST be hardened and monitored (auth, rate limits, logs).
- NFR-S-2: ZK circuits and contracts MUST undergo security review before mainnet use.
- NFR-S-3: Secret keys and sensitive configs MUST be stored using a secret manager, not plain env files.
- NFR-S-4: The system MUST avoid collecting unnecessary PII; behavior features MUST be non-identifying as much as possible.

### 3.2 Privacy
- NFR-P-1: Raw behavioral traces SHOULD stay local or be anonymized / aggregated.
- NFR-P-2: On-chain data SHOULD avoid doxxing or linking real-world identities.
- NFR-P-3: Proofs SHOULD leak minimal information beyond statement validity.

### 3.3 Performance
- NFR-PR-1: Human-proof generation MUST be fast enough not to kill UX (target: sub-second or tolerable async patterns).
- NFR-PR-2: On-chain verification MUST be gas-efficient enough to be usable in practice.
- NFR-PR-3: Genome analysis tasks MAY be async, but initial anomaly scoring SHOULD happen within minutes.

### 3.4 Reliability & Availability
- NFR-R-1: Core APIs SHOULD provide at least basic redundancy (no single-region SPOF in later phases).
- NFR-R-2: On-chain contracts MUST be designed for graceful degradation if off-chain infra is temporarily unavailable (e.g., fail-open or fail-safe per protocol choice).

### 3.5 Usability
- NFR-U-1: Integration for protocols MUST be simple (clear docs, SDKs, reference implementations).
- NFR-U-2: Wallet / dApp SDKs SHOULD default to safe behavior with minimal config.

---

## 4. Constraints & Assumptions

- Early versions target EVM chains.
- Off-chain infra can initially be centralized under the project team, with a roadmap towards more decentralization.
- Not all components need to be live at once; MVP focuses on Identity Firewall + basic threat memory.

---

## 5. Open Questions

- Exact boundary between local-only vs remote behavioral data.
- Regulatory considerations around storing and sharing exploit intel.
- How much of the oracle logic must be decentralized in MVP vs later.

This doc is a **living draft** – we will refine and link it to more detailed specs over time.
