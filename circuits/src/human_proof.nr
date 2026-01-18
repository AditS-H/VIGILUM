// Human Proof Circuit
// ====================
// Zero-knowledge proof of human-like on-chain behavior
// 
// This circuit proves that a wallet exhibits behavioral patterns consistent
// with a human operator without revealing the actual transaction data.
//
// Key Features:
// 1. Proves transaction timing entropy (humans have variable timing)
// 2. Proves gas usage variance (bots tend to use consistent gas)
// 3. Proves interaction diversity (humans interact with varied contracts)
// 4. Commitment to wallet address without revealing behavioral data

use std::hash::pedersen_hash;

// ═══════════════════════════════════════════════════════════════════════════════
// CONSTANTS
// ═══════════════════════════════════════════════════════════════════════════════

/// Minimum number of transactions required for proof
global MIN_TX_COUNT: u32 = 10;

/// Maximum transactions analyzed
global MAX_TX_COUNT: u32 = 100;

/// Minimum timing variance threshold (seconds^2)
/// Low variance = bot-like behavior
global MIN_TIMING_VARIANCE: u32 = 1000;

/// Maximum timing variance threshold
/// Extremely high variance = possibly abandoned wallet
global MAX_TIMING_VARIANCE: u32 = 10000000;

/// Minimum gas variance threshold
global MIN_GAS_VARIANCE: u32 = 100;

/// Minimum unique contracts interacted with
global MIN_CONTRACT_DIVERSITY: u32 = 3;

// ═══════════════════════════════════════════════════════════════════════════════
// DATA STRUCTURES
// ═══════════════════════════════════════════════════════════════════════════════

/// Public inputs for human proof verification
pub struct HumanProofPublicInputs {
    /// Commitment to the wallet address
    wallet_commitment: Field,
    /// Commitment to the behavioral features
    features_commitment: Field,
    /// Unix timestamp when proof was generated
    proof_timestamp: Field,
    /// Challenge nonce from the verifier
    challenge_nonce: Field,
}

/// Private witness data (only prover knows)
pub struct HumanProofWitness {
    /// The actual wallet address (as Field)
    wallet_address: Field,
    /// Secret used for wallet commitment
    wallet_secret: Field,
    /// Transaction timestamps (up to 20 most recent)
    tx_timestamps: [Field; 20],
    /// Number of actual transactions (may be less than 20)
    tx_count: u32,
    /// Gas values for each transaction
    gas_values: [Field; 20],
    /// Number of unique contracts interacted with
    unique_contracts: u32,
    /// Hash of contract addresses interacted with
    contracts_hash: Field,
    /// Nonce for features commitment
    features_nonce: Field,
}

/// Computed behavioral features
pub struct BehavioralFeatures {
    /// Mean time between transactions
    timing_mean: Field,
    /// Variance of timing intervals
    timing_variance: Field,
    /// Mean gas usage
    gas_mean: Field,
    /// Variance of gas usage
    gas_variance: Field,
    /// Number of unique contracts
    contract_diversity: u32,
}

// ═══════════════════════════════════════════════════════════════════════════════
// MAIN PROOF FUNCTION
// ═══════════════════════════════════════════════════════════════════════════════

/// Main entry point for human proof verification
/// 
/// This function verifies that:
/// 1. The wallet commitment is correctly computed
/// 2. The transaction count meets minimum requirements
/// 3. Timing variance is within human-like range
/// 4. Gas variance shows human-like variation
/// 5. Contract diversity meets minimum threshold
pub fn prove_human_behavior(
    public: HumanProofPublicInputs,
    witness: HumanProofWitness
) {
    // ═══════════════════════════════════════════════════════════════════════════
    // STEP 1: Verify wallet commitment
    // ═══════════════════════════════════════════════════════════════════════════
    let computed_wallet_commitment = pedersen_hash([
        witness.wallet_address,
        witness.wallet_secret
    ]);
    assert(computed_wallet_commitment == public.wallet_commitment);

    // ═══════════════════════════════════════════════════════════════════════════
    // STEP 2: Verify minimum transaction count
    // ═══════════════════════════════════════════════════════════════════════════
    assert(witness.tx_count >= MIN_TX_COUNT);
    assert(witness.tx_count as u32 <= MAX_TX_COUNT);

    // ═══════════════════════════════════════════════════════════════════════════
    // STEP 3: Compute and verify behavioral features
    // ═══════════════════════════════════════════════════════════════════════════
    let features = compute_behavioral_features(
        witness.tx_timestamps,
        witness.gas_values,
        witness.tx_count,
        witness.unique_contracts
    );

    // Verify timing variance is human-like
    assert(features.timing_variance as u32 >= MIN_TIMING_VARIANCE);
    assert(features.timing_variance as u32 <= MAX_TIMING_VARIANCE);

    // Verify gas variance shows variation (not a bot using fixed gas)
    assert(features.gas_variance as u32 >= MIN_GAS_VARIANCE);

    // Verify contract diversity
    assert(features.contract_diversity >= MIN_CONTRACT_DIVERSITY);

    // ═══════════════════════════════════════════════════════════════════════════
    // STEP 4: Verify features commitment
    // ═══════════════════════════════════════════════════════════════════════════
    let computed_features_commitment = pedersen_hash([
        features.timing_variance,
        features.gas_variance,
        features.contract_diversity as Field,
        witness.features_nonce
    ]);
    assert(computed_features_commitment == public.features_commitment);
}

// ═══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════════

/// Compute behavioral features from transaction data
fn compute_behavioral_features(
    timestamps: [Field; 20],
    gas_values: [Field; 20],
    tx_count: u32,
    unique_contracts: u32
) -> BehavioralFeatures {
    // Compute timing intervals and their statistics
    let (timing_mean, timing_variance) = compute_timing_stats(timestamps, tx_count);
    
    // Compute gas usage statistics
    let (gas_mean, gas_variance) = compute_gas_stats(gas_values, tx_count);

    BehavioralFeatures {
        timing_mean,
        timing_variance,
        gas_mean,
        gas_variance,
        contract_diversity: unique_contracts
    }
}

/// Compute mean and variance of timing intervals between transactions
fn compute_timing_stats(timestamps: [Field; 20], tx_count: u32) -> (Field, Field) {
    // Calculate intervals between consecutive transactions
    let mut total_interval: Field = 0;
    let interval_count = tx_count - 1;

    for i in 0..19 {
        if i < interval_count {
            let interval = timestamps[i + 1] - timestamps[i];
            total_interval = total_interval + interval;
        }
    }

    // Mean interval
    let mean = if interval_count > 0 {
        total_interval / (interval_count as Field)
    } else {
        0
    };

    // Variance calculation
    let mut variance_sum: Field = 0;
    for i in 0..19 {
        if i < interval_count {
            let interval = timestamps[i + 1] - timestamps[i];
            let diff = if interval > mean { interval - mean } else { mean - interval };
            variance_sum = variance_sum + (diff * diff);
        }
    }

    let variance = if interval_count > 0 {
        variance_sum / (interval_count as Field)
    } else {
        0
    };

    (mean, variance)
}

/// Compute mean and variance of gas values
fn compute_gas_stats(gas_values: [Field; 20], tx_count: u32) -> (Field, Field) {
    // Calculate sum
    let mut total: Field = 0;
    for i in 0..20 {
        if i < tx_count {
            total = total + gas_values[i];
        }
    }

    // Mean
    let mean = if tx_count > 0 {
        total / (tx_count as Field)
    } else {
        0
    };

    // Variance
    let mut variance_sum: Field = 0;
    for i in 0..20 {
        if i < tx_count {
            let diff = if gas_values[i] > mean { gas_values[i] - mean } else { mean - gas_values[i] };
            variance_sum = variance_sum + (diff * diff);
        }
    }

    let variance = if tx_count > 0 {
        variance_sum / (tx_count as Field)
    } else {
        0
    };

    (mean, variance)
}

/// Generate a commitment to behavioral features for storage
pub fn compute_features_commitment(
    timing_variance: Field,
    gas_variance: Field,
    contract_diversity: u32,
    nonce: Field
) -> Field {
    pedersen_hash([
        timing_variance,
        gas_variance,
        contract_diversity as Field,
        nonce
    ])
}

/// Generate a commitment to a wallet address
pub fn compute_wallet_commitment(
    wallet_address: Field,
    secret: Field
) -> Field {
    pedersen_hash([wallet_address, secret])
}

// ═══════════════════════════════════════════════════════════════════════════════
// TESTS
// ═══════════════════════════════════════════════════════════════════════════════

#[test]
fn test_wallet_commitment() {
    let address: Field = 0x742d35Cc6634C0532925a3b844Bc9e7595f7;
    let secret: Field = 123456789;
    
    let commitment = compute_wallet_commitment(address, secret);
    
    // Verify commitment is deterministic
    let commitment2 = compute_wallet_commitment(address, secret);
    assert(commitment == commitment2);
    
    // Different secret = different commitment
    let commitment3 = compute_wallet_commitment(address, secret + 1);
    assert(commitment != commitment3);
}

#[test]
fn test_features_commitment() {
    let timing_variance: Field = 5000;
    let gas_variance: Field = 10000;
    let diversity: u32 = 5;
    let nonce: Field = 999;
    
    let commitment = compute_features_commitment(
        timing_variance,
        gas_variance,
        diversity,
        nonce
    );
    
    // Should be non-zero
    assert(commitment != 0);
}

#[test]
fn test_timing_stats_human_like() {
    // Human-like timing: variable intervals
    let timestamps: [Field; 20] = [
        1704067200,  // Base timestamp
        1704067500,  // +300s
        1704068100,  // +600s
        1704068500,  // +400s
        1704069200,  // +700s
        1704069800,  // +600s
        1704070500,  // +700s
        1704071000,  // +500s
        1704071800,  // +800s
        1704072300,  // +500s
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0
    ];
    
    let (mean, variance) = compute_timing_stats(timestamps, 10);
    
    // Mean should be reasonable
    assert(mean > 0);
    
    // Variance should be above threshold (human-like)
    assert(variance as u32 >= MIN_TIMING_VARIANCE);
}

#[test]
fn test_gas_stats() {
    // Varied gas usage (human-like)
    let gas_values: [Field; 20] = [
        21000, 50000, 35000, 100000, 45000,
        30000, 75000, 21000, 60000, 25000,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0
    ];
    
    let (mean, variance) = compute_gas_stats(gas_values, 10);
    
    // Mean should be reasonable
    assert(mean > 0);
    
    // Variance should show variation
    assert(variance as u32 >= MIN_GAS_VARIANCE);
}

#[test]
fn test_valid_human_proof() {
    // Setup test data
    let wallet_address: Field = 0x742d35Cc6634C0532925a3b844Bc9e7595f7;
    let wallet_secret: Field = 12345;
    let features_nonce: Field = 67890;
    
    // Human-like timestamps with variance
    let tx_timestamps: [Field; 20] = [
        1704067200, 1704067500, 1704068100, 1704068500, 1704069200,
        1704069800, 1704070500, 1704071000, 1704071800, 1704072300,
        1704073000, 1704073700, 1704074500, 1704075000, 1704075800,
        0, 0, 0, 0, 0
    ];
    
    // Human-like gas values with variance
    let gas_values: [Field; 20] = [
        21000, 50000, 35000, 100000, 45000,
        30000, 75000, 21000, 60000, 25000,
        40000, 55000, 32000, 80000, 28000,
        0, 0, 0, 0, 0
    ];
    
    // Compute expected features
    let features = compute_behavioral_features(
        tx_timestamps,
        gas_values,
        15,
        5
    );
    
    // Compute commitments
    let wallet_commitment = compute_wallet_commitment(wallet_address, wallet_secret);
    let features_commitment = compute_features_commitment(
        features.timing_variance,
        features.gas_variance,
        features.contract_diversity,
        features_nonce
    );
    
    let public = HumanProofPublicInputs {
        wallet_commitment,
        features_commitment,
        proof_timestamp: 1704080000,
        challenge_nonce: 999
    };
    
    let witness = HumanProofWitness {
        wallet_address,
        wallet_secret,
        tx_timestamps,
        tx_count: 15,
        gas_values,
        unique_contracts: 5,
        contracts_hash: 0xabcdef,
        features_nonce
    };
    
    // This should pass all assertions
    prove_human_behavior(public, witness);
}
