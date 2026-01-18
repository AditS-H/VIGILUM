"""
VIGILUM ML - Machine Learning Pipeline for Smart Contract Security

This package provides:
- Neural network models for vulnerability detection (model.py)
- Feature extraction from bytecode (features.py)
- Dataset loading and preprocessing (dataset.py)
- Training pipeline (training.py)
- Inference service with ONNX support (inference/)
"""

__version__ = "0.1.0"

from vigilum_ml.models import (
    ModelConfig,
    TrainingConfig,
    ContractFeatures,
    ContractSample,
    VulnerabilityLabel,
    VulnerabilityPrediction,
    PredictionResult,
)
from vigilum_ml.features import FeatureExtractor
from vigilum_ml.model import VulnerabilityDetector, ContrastiveLoss

__all__ = [
    # Models
    "VulnerabilityDetector",
    "ContrastiveLoss",
    # Config
    "ModelConfig",
    "TrainingConfig",
    # Types
    "ContractFeatures",
    "ContractSample",
    "VulnerabilityLabel",
    "VulnerabilityPrediction",
    "PredictionResult",
    # Features
    "FeatureExtractor",
]
