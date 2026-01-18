// Reputation Circuit
// ==================
// Privacy-preserving reputation proofs for security researchers and auditors

use std::hash::pedersen_hash;

/// Public reputation attestation
pub struct ReputationAttestation {
    researcher_commitment: Field,  // Commitment to researcher identity
    reputation_tier: Field,        // 1=Bronze, 2=Silver, 3=Gold, 4=Platinum
    min_audits: Field,             // Minimum audits completed
    valid_since: Field,            // Timestamp when reputation was earned
}

/// Private reputation data
pub struct ReputationWitness {
    researcher_secret: Field,      // Researcher's private identifier
    total_audits: Field,           // Actual audit count
    confirmed_vulns: Field,        // Confirmed vulnerabilities found
    reputation_score: Field,       // Internal reputation score
    nonce: Field,                  // Random value for commitment
}

/// Proves researcher has minimum reputation tier without revealing exact stats
pub fn prove_reputation_tier(
    public: ReputationAttestation,
    witness: ReputationWitness
) {
    // Verify researcher identity commitment
    let computed_commitment = pedersen_hash([witness.researcher_secret, witness.nonce]);
    assert(computed_commitment == public.researcher_commitment);
    
    // Verify audit count meets minimum
    assert(witness.total_audits as u32 >= public.min_audits as u32);
    
    // Verify tier requirements
    // Bronze: 1+ audits, Silver: 10+ audits, Gold: 50+ audits, Platinum: 100+ audits
    let tier = public.reputation_tier as u32;
    
    if tier == 1 {
        assert(witness.total_audits as u32 >= 1);
    } else if tier == 2 {
        assert(witness.total_audits as u32 >= 10);
        assert(witness.confirmed_vulns as u32 >= 5);
    } else if tier == 3 {
        assert(witness.total_audits as u32 >= 50);
        assert(witness.confirmed_vulns as u32 >= 25);
    } else if tier == 4 {
        assert(witness.total_audits as u32 >= 100);
        assert(witness.confirmed_vulns as u32 >= 50);
    }
}

/// Proves a researcher found a vulnerability without revealing their identity
pub fn prove_anonymous_discovery(
    vuln_hash: Field,
    researcher_commitment: Field,
    researcher_secret: Field,
    discovery_proof: Field
) {
    // Verify researcher commitment
    let computed = pedersen_hash([researcher_secret]);
    assert(computed == researcher_commitment);
    
    // Verify discovery proof links researcher to vulnerability
    let expected_proof = pedersen_hash([vuln_hash, researcher_secret]);
    assert(expected_proof == discovery_proof);
}

/// Computes a researcher identity commitment
pub fn compute_researcher_commitment(
    secret: Field,
    nonce: Field
) -> Field {
    pedersen_hash([secret, nonce])
}

/// Computes a discovery proof for bounty claims
pub fn compute_discovery_proof(
    vuln_hash: Field,
    researcher_secret: Field
) -> Field {
    pedersen_hash([vuln_hash, researcher_secret])
}

#[test]
fn test_gold_tier_proof() {
    let secret = 99999;
    let nonce = 11111;
    
    let commitment = pedersen_hash([secret, nonce]);
    
    let public = ReputationAttestation {
        researcher_commitment: commitment,
        reputation_tier: 3, // Gold
        min_audits: 50,
        valid_since: 1704067200,
    };
    
    let witness = ReputationWitness {
        researcher_secret: secret,
        total_audits: 75,
        confirmed_vulns: 30,
        reputation_score: 8500,
        nonce: nonce,
    };
    
    prove_reputation_tier(public, witness);
}
