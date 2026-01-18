# VIGILUM: What We Want to Achieve

This file captures **goals**, **success criteria**, and **milestones** for VIGILUM as a startup-grade project.

---

## 1. Vision-Level Outcomes

- Make "on-chain immune systems" a **standard primitive** for Web3 protocols.
- Turn every hack into **shared defense memory**, not just a post-mortem.
- Enable **responsible, cryptographically verifiable disclosure** via Proof-of-Exploit.
- Provide DAOs and protocols with **continuous security**, not one-off audits.

---

## 2. Product Outcomes

### 2.1 For Protocols / DAOs
- Reduce successful exploit rate and impact.
- Reduce reaction time from "incident detected" to "mitigation applied" from days → minutes.
- Offer an easy on-ramp:
  - Simple contract integrations (1–3 function calls / modifiers).
  - Clear dashboards / APIs for current threat posture.

### 2.2 For Security Researchers
- Provide a **safer, more profitable** path than going full black-hat.
- Allow them to:
  - Prove exploits without leaking them.
  - Get paid automatically and transparently.
  - Build on-chain security reputation.

### 2.3 For End Users / Wallets
- Reduce exposure to compromised protocols and phishing.
- Give **visible, interpretable signals** about risk before they sign.
- Keep UX smooth (no heavy CAPTCHAs or KYC).

---

## 3. Research Outcomes

- Novel ZK constructions or applications for:
  - Human-like behavior proofs.
  - Exploit existence proofs.
- New representations for "malware genomes" on-chain:
  - Stable under minor mutations.
  - Useful for clustering families of exploits.
- Publishable results (papers / blog posts) that still respect operational security.

---

## 4. Quantitative Success Metrics (First Draft)

- **Adoption**
  - N protocols integrated (e.g., 3–5 for early stage, 10+ later).
  - Number of transactions passing through Identity Firewall.
- **Security impact**
  - Number of detected and prevented exploits.
  - Time-to-detection and time-to-mitigation vs industry baseline.
- **Researcher ecosystem**
  - Number of active red-team participants.
  - Number of valid Proof-of-Exploit submissions.
  - Ratio of valid to invalid (slashed) submissions.

---

## 5. Milestones (High-Level)

### M0 – Planning & Research
- Clarify architecture and tech stack (this phase).
- Identify initial chains and protocols to target.
- Draft initial threat models and ZK research directions.

### M1 – MVP (Alpha)
- Identity Firewall prototype:
  - Minimal Sentinel SDK.
  - On-chain verifier for human-proof.
- Basic Malware Genome pipeline:
  - Sandbox execution and fingerprinting.
  - Storage of genome hashes.
- Integrate with 1–2 testnet protocols for experimentation.

### M2 – Proof-of-Exploit + Threat Feeds
- Implement basic Proof-of-Exploit Engine for a limited class of bugs.
- Simple Red-Team DAO with staking and payouts.
- Connect initial threat intel feeds (e.g., GitHub PoCs, public advisories).

### M3 – Self-Healing Features
- Add automatic response mechanisms to partner protocols:
  - Governance freeze.
  - Key rotation flows.
  - Transaction throttling / circuit breakers.
- Demonstrate end-to-end "detect → prove → respond → heal" in the wild or realistic test scenarios.

### M4 – Scaling & Ecosystem
- Harden infrastructure for reliability and decentralization.
- Onboard multiple protocols / DAOs.
- Formalize partnerships (auditors, bug bounty platforms, infra providers).

---

## 6. Non-Goals (At Least for Now)

- Replacing base-layer consensus or building a new L1.
- Becoming a general-purpose DeFi protocol; focus stays on **security**.
- Building full-blown centralized dashboards for everything before core primitives are working.

---

## 7. Next Steps for This Doc

- Refine metrics to be more concrete once we pick target chains and protocols.
- Attach timelines / quarters once team capacity and funding assumptions are clearer.
- Link each milestone to specific technical specs and tasks in future docs.
