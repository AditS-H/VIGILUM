use wasm_bindgen::prelude::*;
use sha2::{Sha256, Digest};
use serde::{Serialize, Deserialize};

#[wasm_bindgen]
pub struct BytecodeAnalyzer {
    bytecode: Vec<u8>,
}

#[wasm_bindgen]
impl BytecodeAnalyzer {
    /// Create a new bytecode analyzer
    #[wasm_bindgen(constructor)]
    pub fn new(bytecode_hex: &str) -> BytecodeAnalyzer {
        let bytecode = hex::decode(bytecode_hex).unwrap_or_default();
        BytecodeAnalyzer { bytecode }
    }

    /// Get bytecode hash (SHA256)
    pub fn hash(&self) -> String {
        let mut hasher = Sha256::new();
        hasher.update(&self.bytecode);
        hex::encode(hasher.finalize())
    }

    /// Get bytecode length
    pub fn length(&self) -> usize {
        self.bytecode.len()
    }

    /// Extract EVM opcodes
    pub fn extract_opcodes(&self) -> Vec<u8> {
        self.bytecode.clone()
    }

    /// Detect potential vulnerabilities (basic pattern matching)
    pub fn detect_patterns(&self) -> String {
        let mut patterns = Vec::new();

        // Check for selfdestruct opcode (0xff)
        if self.bytecode.contains(&0xff) {
            patterns.push("selfdestruct_present");
        }

        // Check for delegatecall opcode (0xf4)
        if self.bytecode.contains(&0xf4) {
            patterns.push("delegatecall_present");
        }

        // Check for fallback function (no-arg function selector)
        if self.bytecode.len() > 0 {
            patterns.push("has_runtime_code");
        }

        serde_json::to_string(&patterns).unwrap_or_default()
    }

    /// Calculate entropy of bytecode
    pub fn entropy(&self) -> f64 {
        if self.bytecode.is_empty() {
            return 0.0;
        }

        let mut freq = [0u32; 256];
        for &byte in &self.bytecode {
            freq[byte as usize] += 1;
        }

        let len = self.bytecode.len() as f64;
        let mut entropy = 0.0;

        for count in freq.iter() {
            if *count > 0 {
                let p = *count as f64 / len;
                entropy -= p * p.log2();
            }
        }

        entropy
    }
}

#[derive(Serialize, Deserialize)]
pub struct ProofData {
    pub contract_address: String,
    pub proof_hash: String,
    pub timestamp: u64,
}

#[wasm_bindgen]
pub struct ProofGenerator {
    challenge: Vec<u8>,
}

#[wasm_bindgen]
impl ProofGenerator {
    /// Create a new proof generator
    #[wasm_bindgen(constructor)]
    pub fn new(challenge_hex: &str) -> ProofGenerator {
        let challenge = hex::decode(challenge_hex).unwrap_or_default();
        ProofGenerator { challenge }
    }

    /// Generate a proof
    pub fn generate_proof(&self, contract_addr: &str) -> String {
        let mut hasher = Sha256::new();
        hasher.update(&self.challenge);
        hasher.update(contract_addr.as_bytes());
        
        let proof_hash = hex::encode(hasher.finalize());

        let proof = ProofData {
            contract_address: contract_addr.to_string(),
            proof_hash,
            timestamp: timestamp(),
        };

        serde_json::to_string(&proof).unwrap_or_default()
    }

    /// Verify a proof (basic check)
    pub fn verify_proof(&self, proof_json: &str) -> bool {
        if let Ok(proof) = serde_json::from_str::<ProofData>(proof_json) {
            // Verify proof hash matches challenge
            let mut hasher = Sha256::new();
            hasher.update(&self.challenge);
            hasher.update(proof.contract_address.as_bytes());
            
            let expected_hash = hex::encode(hasher.finalize());
            proof.proof_hash == expected_hash
        } else {
            false
        }
    }
}

/// Get current timestamp in seconds
fn timestamp() -> u64 {
    #[cfg(target_arch = "wasm32")]
    {
        (js_sys::Date::now() / 1000.0) as u64
    }
    
    #[cfg(not(target_arch = "wasm32"))]
    {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_bytecode_analyzer() {
        let analyzer = BytecodeAnalyzer::new("6080604052");
        assert_eq!(analyzer.length(), 3);
        assert!(!analyzer.hash().is_empty());
    }

    #[test]
    fn test_proof_generation() {
        let generator = ProofGenerator::new("deadbeef");
        let proof = generator.generate_proof("0x1234");
        assert!(!proof.is_empty());
        assert!(generator.verify_proof(&proof));
    }
}
