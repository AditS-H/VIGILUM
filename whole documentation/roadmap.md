# VIGILUM Roadmap (High-Level)

This roadmap will evolve, but sets a **directional path** from research → MVP → ecosystem.

---

## Phase 0 – Exploration & Foundations

**Goals**
- Validate core problem: on-chain security is reactive, fragmented, and slow.
- Map competitor / adjacent solutions (audits, bug bounties, monitoring tools).
- Refine VIGILUM's unique positioning and primitives.

**Deliverables**
- Architecture, tech stack, and requirements docs (this folder).
- Initial ZK and genome modeling research notes.
- Short list of target chains and candidate protocols.

---

## Phase 1 – MVP: Identity Firewall + Basic Threat Memory

**Focus**
- Build something small but **end-to-end** that can be demoed.

**Core features**
- Sentinel SDK (alpha) for dApps / wallets.
- Identity Firewall on-chain contract with verifier.
- Simple behavioral model (even heuristic-based at first).
- Basic malware genome pipeline:
  - Ingest a few known exploit contracts.
  - Store genome hashes and simple labels.

**Outputs**
- Working demo with:
  - A testnet protocol integrating Identity Firewall.
  - Dashboard or CLI to show detection / gating events.

---

## Phase 2 – Proof-of-Exploit & Red-Team Alpha

**Focus**
- Attract serious researchers with a safer path to disclosure.

**Core features**
- First version of Proof-of-Exploit Engine for a narrow class of bugs.
- Red-Team DAO smart contracts:
  - Staking, basic reputation, and rewards.
- Integration with Malware Genome DB:
  - Each accepted exploit → new genome entry.

**Outputs**
- Small set of trusted researchers testing the flow.
- First payouts via Proof-of-Exploit.

---

## Phase 3 – Threat Oracle & Self-Healing Hooks

**Focus**
- Go from **detection** to **automatic mitigation**.

**Core features**
- Threat Oracle Layer that publishes risk scores on-chain.
- Library of defensive patterns for protocols, e.g.:
  - Governance freeze on key-leak signal.
  - Rate limit / pause on exploit-campaign signal.
- Deeper integration with 2–3 partner protocols to exercise self-healing loop.

**Outputs**
- Demonstrated real or realistic scenario of:
  - Attack attempt → detection → automated protocol defense.

---

## Phase 4 – Scaling & Decentralization

**Focus**
- Increase robustness, reach, and decentralization of off-chain parts.

**Core features**
- More scalable and robust infra (multi-region, more services in Go/Rust if needed).
- Progressive decentralization of oracles and some analysis components.
- Governance expansion in Red-Team DAO.

**Outputs**
- Multiple protocols using VIGILUM in production.
- Documented security posture and audited contracts / circuits.

---

## Phase 5 – Ecosystem & Research Frontier

**Focus**
- Position VIGILUM as a **shared security layer** across chains.

**Core features**
- Multi-chain support and cross-chain threat memory.
- Advanced ZK research applied to more subtle exploit classes.
- Community tooling and integrations (plugins, SDKs, open datasets).

**Outputs**
- Papers / public write-ups.
- Ecosystem of researchers, protocols, and infra partners building on top of VIGILUM.

---

## Notes

- Each phase should have a minimal acceptable scope; we can cut features but not the end-to-end story.
- Timelines will depend on team size and funding; to be added later.
