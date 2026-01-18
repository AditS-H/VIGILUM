"""
Core domain models for VIGILUM ML pipeline.
"""

from datetime import datetime
from enum import Enum
from typing import Any

from pydantic import BaseModel, Field


class ThreatLevel(str, Enum):
    """Severity classification for detected threats."""
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    INFO = "info"
    NONE = "none"


class VulnerabilityType(str, Enum):
    """Categories of smart contract vulnerabilities."""
    REENTRANCY = "reentrancy"
    INTEGER_OVERFLOW = "integer_overflow"
    INTEGER_UNDERFLOW = "integer_underflow"
    ACCESS_CONTROL = "access_control"
    UNCHECKED_CALL = "unchecked_external_call"
    TX_ORIGIN = "tx_origin"
    TIMESTAMP_DEPENDENCY = "timestamp_dependency"
    FRONTRUNNING = "frontrunning"
    FLASH_LOAN = "flash_loan_attack"
    ORACLE_MANIPULATION = "oracle_manipulation"
    RUG_PULL = "rug_pull_pattern"
    HONEYPOT = "honeypot"
    PHISHING = "phishing_signature"


class ContractFeatures(BaseModel):
    """Extracted features from smart contract bytecode/source."""
    
    # Bytecode statistics
    bytecode_length: int = Field(..., ge=0)
    opcode_distribution: dict[str, int] = Field(default_factory=dict)
    unique_opcodes: int = Field(..., ge=0)
    
    # Control flow features
    jump_count: int = Field(..., ge=0)
    call_count: int = Field(..., ge=0)
    delegatecall_count: int = Field(..., ge=0)
    staticcall_count: int = Field(..., ge=0)
    selfdestruct_present: bool = False
    
    # Storage features
    sload_count: int = Field(..., ge=0)
    sstore_count: int = Field(..., ge=0)
    
    # External interaction features
    external_calls: int = Field(..., ge=0)
    ether_transfers: int = Field(..., ge=0)
    
    # Complexity metrics
    cyclomatic_complexity: float = Field(default=0.0, ge=0)
    code_entropy: float = Field(default=0.0, ge=0, le=8)
    
    # Function signatures (first 4 bytes)
    function_selectors: list[str] = Field(default_factory=list)
    
    # Known patterns
    is_proxy: bool = False
    is_upgradeable: bool = False
    has_ownership: bool = False
    has_pausable: bool = False


class PredictionResult(BaseModel):
    """ML model prediction output."""
    
    contract_address: str
    chain_id: int
    
    # Overall assessment
    risk_score: float = Field(..., ge=0, le=100)
    threat_level: ThreatLevel
    confidence: float = Field(..., ge=0, le=1)
    
    # Per-vulnerability predictions
    vulnerability_probabilities: dict[VulnerabilityType, float] = Field(default_factory=dict)
    
    # Top predicted vulnerabilities
    predicted_vulnerabilities: list[VulnerabilityType] = Field(default_factory=list)
    
    # Model metadata
    model_version: str
    inference_time_ms: float
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class TrainingSample(BaseModel):
    """Single training sample for ML model."""
    
    # Input features
    features: ContractFeatures
    bytecode_embedding: list[float] = Field(default_factory=list)
    
    # Labels
    is_malicious: bool
    vulnerability_labels: list[VulnerabilityType] = Field(default_factory=list)
    risk_score: float = Field(..., ge=0, le=100)
    
    # Metadata
    contract_address: str
    chain_id: int
    source: str  # Dataset source
    verified: bool = False  # Human-verified label


class ModelConfig(BaseModel):
    """Configuration for ML model training and inference."""
    
    # Architecture
    embedding_dim: int = 256
    hidden_dim: int = 512
    num_layers: int = 4
    num_heads: int = 8
    dropout: float = 0.1
    
    # Training
    batch_size: int = 32
    learning_rate: float = 1e-4
    weight_decay: float = 1e-5
    max_epochs: int = 100
    early_stopping_patience: int = 10
    
    # Inference
    confidence_threshold: float = 0.5
    max_sequence_length: int = 8192  # Max bytecode length
    
    # Paths
    model_path: str = "./models/vulnerability_detector.onnx"
    checkpoint_path: str = "./checkpoints/"


class InferenceRequest(BaseModel):
    """Request for ML inference."""
    
    contract_address: str
    chain_id: int
    bytecode: str  # Hex-encoded
    source_code: str | None = None


# Additional types for training and inference

class VulnerabilityLabel(BaseModel):
    """Label for a vulnerability in training data."""
    vuln_type: str
    severity: str | None = None
    confidence: float = 1.0


class ContractSample(BaseModel):
    """Training sample for vulnerability detection."""
    contract_id: str
    bytecode: str
    source_code: str | None = None
    is_malicious: bool = False
    risk_score: float | None = None
    labels: list[VulnerabilityLabel] = Field(default_factory=list)


class VulnerabilityPrediction(BaseModel):
    """Single vulnerability prediction from the model."""
    vuln_type: str
    confidence: float = Field(..., ge=0, le=1)
    severity: str = "medium"


class TrainingConfig(BaseModel):
    """Configuration for model training."""
    
    # Training parameters
    num_epochs: int = 50
    batch_size: int = 32
    learning_rate: float = 1e-4
    weight_decay: float = 1e-5
    gradient_clip: float = 1.0
    num_workers: int = 4
    
    # Loss weights
    malicious_weight: float = 1.0
    vuln_weight: float = 1.0
    risk_weight: float = 0.5
    
    # Checkpointing
    save_every: int = 5
    

# Redefine PredictionResult with more flexible fields
class PredictionResult(BaseModel):
    """ML model prediction output (flexible version for inference service)."""
    
    # Overall assessment
    is_malicious: bool = False
    malicious_probability: float = Field(0.0, ge=0, le=1)
    risk_score: int = Field(0, ge=0, le=100)
    
    # Vulnerability predictions
    vulnerabilities: list[VulnerabilityPrediction] = Field(default_factory=list)
    
    # Embedding for similarity search
    embedding: list[float] | None = None
    
    # Contract features
    contract_features: ContractFeatures | None = None

    abi: dict[str, Any] | None = None


class InferenceResponse(BaseModel):
    """Response from ML inference."""
    
    request_id: str
    prediction: PredictionResult
    features: ContractFeatures
    processing_time_ms: float
