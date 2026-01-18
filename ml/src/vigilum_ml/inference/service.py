"""
Inference service for vulnerability detection.
"""

from pathlib import Path
from typing import Any

import torch
import onnx
import onnxruntime as ort

import structlog

from vigilum_ml.dataset import VULN_TYPES, VULN_TO_IDX
from vigilum_ml.features import FeatureExtractor
from vigilum_ml.model import VulnerabilityDetector
from vigilum_ml.models import (
    ModelConfig,
    PredictionResult,
    VulnerabilityPrediction,
)
from vigilum_ml.training import load_model_from_checkpoint

logger = structlog.get_logger(__name__)


class InferenceService:
    """
    High-performance inference service for vulnerability detection.
    
    Supports both PyTorch and ONNX runtime backends.
    """

    def __init__(
        self,
        model_path: Path | str,
        use_onnx: bool = True,
        device: str = "cpu",
        max_seq_length: int = 8192,
    ):
        self.model_path = Path(model_path)
        self.use_onnx = use_onnx
        self.device = device
        self.max_seq_length = max_seq_length
        
        self.feature_extractor = FeatureExtractor()
        self.logger = logger.bind(component="inference_service")
        
        # Load model
        if use_onnx and self.model_path.suffix == ".onnx":
            self._load_onnx_model()
        else:
            self._load_pytorch_model()

    def _load_pytorch_model(self) -> None:
        """Load PyTorch model from checkpoint."""
        self.logger.info("Loading PyTorch model", path=str(self.model_path))
        
        device = torch.device(self.device)
        self.model = load_model_from_checkpoint(self.model_path, device)
        self.model.eval()
        self.onnx_session = None

    def _load_onnx_model(self) -> None:
        """Load ONNX model for optimized inference."""
        self.logger.info("Loading ONNX model", path=str(self.model_path))
        
        # Configure ONNX runtime
        providers = ["CPUExecutionProvider"]
        if self.device == "cuda":
            providers = ["CUDAExecutionProvider", "CPUExecutionProvider"]
        
        self.onnx_session = ort.InferenceSession(
            str(self.model_path),
            providers=providers,
        )
        self.model = None

    def predict(self, bytecode_hex: str) -> PredictionResult:
        """
        Predict vulnerabilities in contract bytecode.
        
        Args:
            bytecode_hex: Hex-encoded contract bytecode
            
        Returns:
            PredictionResult with all predictions
        """
        # Preprocess bytecode
        bytecode_tensor = self._preprocess_bytecode(bytecode_hex)
        
        # Run inference
        if self.onnx_session:
            outputs = self._predict_onnx(bytecode_tensor)
        else:
            outputs = self._predict_pytorch(bytecode_tensor)
        
        # Postprocess outputs
        return self._postprocess_outputs(outputs, bytecode_hex)

    def predict_batch(self, bytecodes: list[str]) -> list[PredictionResult]:
        """Predict vulnerabilities for multiple contracts."""
        return [self.predict(bc) for bc in bytecodes]

    def _preprocess_bytecode(self, bytecode_hex: str) -> torch.Tensor:
        """Convert bytecode hex to tensor."""
        bytecode_hex = bytecode_hex.lower()
        if bytecode_hex.startswith("0x"):
            bytecode_hex = bytecode_hex[2:]
        
        try:
            bytecode = bytes.fromhex(bytecode_hex)
        except ValueError:
            bytecode = b""
        
        # Convert to tensor
        tensor = torch.tensor(list(bytecode), dtype=torch.long)
        
        # Pad or truncate
        if len(tensor) < self.max_seq_length:
            pad_len = self.max_seq_length - len(tensor)
            tensor = torch.cat([
                tensor,
                torch.full((pad_len,), 256, dtype=torch.long),
            ])
        else:
            tensor = tensor[:self.max_seq_length]
        
        return tensor.unsqueeze(0)  # Add batch dimension

    def _predict_pytorch(self, bytecode_tensor: torch.Tensor) -> dict[str, Any]:
        """Run inference with PyTorch model."""
        device = next(self.model.parameters()).device
        bytecode_tensor = bytecode_tensor.to(device)
        
        with torch.no_grad():
            outputs = self.model(bytecode_tensor)
        
        return {
            "is_malicious": outputs["is_malicious"].cpu().numpy()[0, 0],
            "vulnerabilities": outputs["vulnerabilities"].cpu().numpy()[0],
            "risk_score": outputs["risk_score"].cpu().numpy()[0, 0],
            "features": outputs["features"].cpu().numpy()[0],
        }

    def _predict_onnx(self, bytecode_tensor: torch.Tensor) -> dict[str, Any]:
        """Run inference with ONNX runtime."""
        # Prepare inputs
        inputs = {
            "bytecode": bytecode_tensor.numpy(),
            "attention_mask": (bytecode_tensor != 256).long().numpy(),
        }
        
        # Run inference
        outputs = self.onnx_session.run(None, inputs)
        
        return {
            "is_malicious": outputs[0][0, 0],
            "vulnerabilities": outputs[1][0],
            "risk_score": outputs[2][0, 0],
            "features": outputs[3][0],
        }

    def _postprocess_outputs(
        self,
        outputs: dict[str, Any],
        bytecode_hex: str,
    ) -> PredictionResult:
        """Convert raw model outputs to structured prediction result."""
        # Extract vulnerability predictions
        vuln_probs = outputs["vulnerabilities"]
        vulnerabilities = []
        
        for i, vuln_type in enumerate(VULN_TYPES):
            if vuln_probs[i] > 0.3:  # Threshold for reporting
                vulnerabilities.append(VulnerabilityPrediction(
                    vuln_type=vuln_type,
                    confidence=float(vuln_probs[i]),
                    severity=self._get_severity(vuln_type, vuln_probs[i]),
                ))
        
        # Sort by confidence
        vulnerabilities.sort(key=lambda x: x.confidence, reverse=True)
        
        # Calculate contract features for additional context
        try:
            features = self.feature_extractor.extract(bytecode_hex)
        except Exception:
            features = None
        
        return PredictionResult(
            is_malicious=float(outputs["is_malicious"]) > 0.5,
            malicious_probability=float(outputs["is_malicious"]),
            vulnerabilities=vulnerabilities,
            risk_score=int(float(outputs["risk_score"]) * 100),
            embedding=outputs["features"].tolist() if "features" in outputs else None,
            contract_features=features,
        )

    def _get_severity(self, vuln_type: str, confidence: float) -> str:
        """Determine severity level based on vulnerability type and confidence."""
        # Critical vulnerabilities
        if vuln_type in ["reentrancy", "access_control", "flash_loan", "oracle_manipulation"]:
            if confidence > 0.8:
                return "critical"
            elif confidence > 0.6:
                return "high"
            else:
                return "medium"
        
        # High-severity vulnerabilities
        elif vuln_type in ["rug_pull", "honeypot", "integer_overflow", "integer_underflow"]:
            if confidence > 0.8:
                return "high"
            elif confidence > 0.5:
                return "medium"
            else:
                return "low"
        
        # Medium-severity vulnerabilities
        else:
            if confidence > 0.8:
                return "medium"
            else:
                return "low"

    def get_embedding(self, bytecode_hex: str) -> list[float]:
        """
        Get contract embedding for similarity search.
        
        Returns a fixed-size vector representation of the contract.
        """
        result = self.predict(bytecode_hex)
        return result.embedding or []

    def similarity(self, bytecode1: str, bytecode2: str) -> float:
        """
        Calculate similarity between two contracts.
        
        Returns cosine similarity (0-1) between contract embeddings.
        """
        emb1 = self.get_embedding(bytecode1)
        emb2 = self.get_embedding(bytecode2)
        
        if not emb1 or not emb2:
            return 0.0
        
        import numpy as np
        
        vec1 = np.array(emb1)
        vec2 = np.array(emb2)
        
        dot = np.dot(vec1, vec2)
        norm1 = np.linalg.norm(vec1)
        norm2 = np.linalg.norm(vec2)
        
        if norm1 == 0 or norm2 == 0:
            return 0.0
        
        return float(dot / (norm1 * norm2))


def export_to_onnx(
    model: VulnerabilityDetector,
    output_path: Path,
    max_seq_length: int = 8192,
) -> None:
    """
    Export PyTorch model to ONNX format for optimized inference.
    
    Args:
        model: Trained VulnerabilityDetector model
        output_path: Path to save ONNX model
        max_seq_length: Maximum sequence length for inputs
    """
    model.eval()
    
    # Create dummy inputs
    dummy_bytecode = torch.randint(0, 256, (1, max_seq_length), dtype=torch.long)
    dummy_mask = torch.ones(1, max_seq_length, dtype=torch.long)
    
    # Export
    torch.onnx.export(
        model,
        (dummy_bytecode, dummy_mask),
        output_path,
        input_names=["bytecode", "attention_mask"],
        output_names=["is_malicious", "vulnerabilities", "risk_score", "features"],
        dynamic_axes={
            "bytecode": {0: "batch_size"},
            "attention_mask": {0: "batch_size"},
        },
        opset_version=17,
    )
    
    # Verify exported model
    onnx_model = onnx.load(output_path)
    onnx.checker.check_model(onnx_model)
    
    logger.info("Model exported to ONNX", path=str(output_path))
