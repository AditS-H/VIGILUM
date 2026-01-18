# VIGILUM Risks & Assumptions (Draft)

We explicitly track **risks**, **assumptions**, and potential **failure modes**.

---

## 1. Technical Risks

- **R1 – ZK complexity**
  - Risk: Proof-of-Exploit and behavior-based proofs may be harder than expected.
  - Mitigation: Start with narrow, well-defined use cases; collaborate with ZK experts; rely on audited libraries.

- **R2 – False positives / negatives**
  - Risk: Identity Firewall may misclassify humans as bots (or vice versa).
  - Mitigation: Conservative thresholds; multi-signal decision-making; allow protocols to choose fail-open vs fail-closed.

- **R3 – Oracle vulnerabilities**
  - Risk: Threat Oracle may become a single point of failure or attack.
  - Mitigation: Multiple independent data sources; crypto-economic incentives; transparency in how signals are generated.

- **R4 – Performance / gas costs**
  - Risk: On-chain verification may be too expensive or slow.
  - Mitigation: Optimize circuits and contracts; offload heavy work off-chain; iterate with gas benchmarks early.

---

## 2. Game-Theoretic / Economic Risks

- **R5 – Adversarial researchers**
  - Risk: Attackers game the Red-Team DAO (fake reports, collusion, griefing).
  - Mitigation: Staking + slashing; careful reward curves; social + technical defenses; gradual access levels.

- **R6 – Perverse incentives**
  - Risk: Incentive design may encourage hoarding of zero-days or partial disclosure.
  - Mitigation: Reward schemes that value early disclosure and high impact; penalties for late or withheld information when damage is clear.

---

## 3. Adoption & Ecosystem Risks

- **R7 – Integration friction**
  - Risk: Protocols may find it too complex to integrate VIGILUM.
  - Mitigation: Provide dead-simple SDKs, reference contracts, and migration guides; start with a small, friendly set of partners.

- **R8 – Market skepticism**
  - Risk: Ecosystem is skeptical of new security primitives.
  - Mitigation: Transparent metrics; case studies; collaborations with respected auditors and researchers.

---

## 4. Legal / Ethical Risks

- **R9 – Handling of exploit data**
  - Risk: Storing and sharing exploit intel might have legal or ethical implications.
  - Mitigation: Avoid distributing working exploits; focus on proofs and anonymized genomes; seek legal advice where needed.

- **R10 – Jurisdictional issues**
  - Risk: Running a global security network crosses many jurisdictions.
  - Mitigation: Keep core infra neutral; design governance to be global and open.

---

## 5. Assumptions (To Validate)

- Attackers will respond to incentives and use Proof-of-Exploit instead of going fully black-hat in at least some cases.
- Protocols are willing to offload some security logic to a shared layer.
- ZK technology is mature enough for our chosen narrow use cases in a 1–3 year horizon.

---

## 6. Next Steps

- Prioritize risks (impact × likelihood).
- Link each high-priority risk to concrete mitigation tasks in technical specs and roadmap.
- Revisit this document regularly as design and experiments evolve.
