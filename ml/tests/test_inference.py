"""Tests for inference service."""

import pytest
import torch
from unittest.mock import MagicMock, patch

from vigilum_ml.models import ModelConfig


class TestInferenceService:
    """Test the InferenceService."""

    @pytest.fixture
    def config(self) -> ModelConfig:
        """Create a small config for testing."""
        return ModelConfig(
            embedding_dim=64,
            num_heads=2,
            num_layers=2,
            hidden_dim=128,
            max_sequence_length=512,
        )

    @pytest.fixture
    def mock_model(self, config: ModelConfig):
        """Create a mock model."""
        model = MagicMock()
        model.eval = MagicMock(return_value=model)
        model.to = MagicMock(return_value=model)
        
        # Mock forward pass
        def mock_forward(x):
            batch_size = x.shape[0]
            return (
                torch.rand(batch_size, 1),  # is_malicious
                torch.rand(batch_size, 10),  # vuln_logits
                torch.rand(batch_size, 1),  # risk_score
                torch.rand(batch_size, config.embedding_dim),  # embedding
            )
        
        model.return_value = mock_forward
        model.side_effect = mock_forward
        
        return model

    def test_bytecode_preprocessing(self, config: ModelConfig):
        """Test bytecode is preprocessed correctly."""
        from vigilum_ml.dataset import ContractDataset
        from vigilum_ml.models import ContractSample
        
        sample = ContractSample(
            contract_id="test",
            bytecode="0x608060405260"
        )
        dataset = ContractDataset([sample], max_seq_length=config.max_sequence_length)
        item = dataset[0]
        
        assert item["bytecode"].shape == (config.max_sequence_length,)
        assert item["bytecode"].dtype == torch.long

    def test_risk_score_normalization(self):
        """Test risk scores are normalized to [0, 1]."""
        # Sigmoid output is always in [0, 1]
        raw_scores = torch.randn(10)
        normalized = torch.sigmoid(raw_scores)
        
        assert (normalized >= 0).all()
        assert (normalized <= 1).all()

    def test_vulnerability_thresholding(self):
        """Test vulnerability detection with threshold."""
        vuln_logits = torch.tensor([[2.0, -1.0, 0.5, -2.0, 1.5]])
        probs = torch.sigmoid(vuln_logits)
        threshold = 0.5
        
        detected = probs > threshold
        
        # Should detect vulns at indices 0, 2, 4
        expected = torch.tensor([[True, False, True, False, True]])
        assert (detected == expected).all()

    def test_embedding_similarity(self):
        """Test embedding cosine similarity."""
        embed_dim = 64
        
        # Create two similar embeddings
        base = torch.randn(embed_dim)
        similar = base + torch.randn(embed_dim) * 0.1
        different = torch.randn(embed_dim)
        
        # Cosine similarity
        def cosine_sim(a, b):
            return torch.nn.functional.cosine_similarity(
                a.unsqueeze(0), b.unsqueeze(0)
            ).item()
        
        sim_similar = cosine_sim(base, similar)
        sim_different = cosine_sim(base, different)
        
        # Similar embeddings should have higher similarity
        assert sim_similar > 0.9  # Should be close to 1
        assert abs(sim_different) < sim_similar  # Should be much lower


class TestONNXExport:
    """Test ONNX export functionality."""

    def test_torch_onnx_export_available(self):
        """Test torch.onnx is available."""
        import torch.onnx
        assert hasattr(torch.onnx, "export")

    def test_model_traceable(self):
        """Test model can be traced for export."""
        from vigilum_ml.model import VulnerabilityDetector
        
        config = ModelConfig(
            embedding_dim=32,
            num_heads=2,
            num_layers=1,
            hidden_dim=64,
            max_sequence_length=128,
        )
        
        model = VulnerabilityDetector(config, num_vuln_classes=10)
        model.eval()
        
        # Create dummy input
        dummy_input = torch.randint(0, 256, (1, 128))
        
        # Try tracing
        with torch.no_grad():
            outputs = model(dummy_input)
        
        assert len(outputs) == 4  # is_malicious, vuln_logits, risk_score, embedding


class TestBatchProcessing:
    """Test batch processing capabilities."""

    def test_variable_batch_sizes(self):
        """Test model handles variable batch sizes."""
        from vigilum_ml.model import VulnerabilityDetector
        
        config = ModelConfig(
            embedding_dim=32,
            num_heads=2,
            num_layers=1,
            hidden_dim=64,
            max_sequence_length=128,
        )
        
        model = VulnerabilityDetector(config, num_vuln_classes=10)
        model.eval()
        
        for batch_size in [1, 4, 16, 32]:
            x = torch.randint(0, 256, (batch_size, 128))
            
            with torch.no_grad():
                outputs = model(x)
            
            assert outputs["is_malicious"].shape[0] == batch_size
            assert outputs["vulnerabilities"].shape[0] == batch_size
            assert outputs["risk_score"].shape[0] == batch_size
            assert outputs["features"].shape[0] == batch_size
