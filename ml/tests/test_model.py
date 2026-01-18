"""Tests for vulnerability detection model."""

import pytest
import torch
from unittest.mock import MagicMock

from vigilum_ml.model import VulnerabilityDetector, ContrastiveLoss
from vigilum_ml.models import ModelConfig


class TestVulnerabilityDetector:
    """Test the VulnerabilityDetector model."""

    @pytest.fixture
    def config(self) -> ModelConfig:
        """Create a small config for testing."""
        return ModelConfig(
            embedding_dim=64,
            num_heads=2,
            num_layers=2,
            hidden_dim=128,
            max_sequence_length=512,
            dropout=0.1,
        )

    @pytest.fixture
    def model(self, config: ModelConfig) -> VulnerabilityDetector:
        """Create a model for testing."""
        return VulnerabilityDetector(config, num_vuln_classes=10)

    def test_model_creation(self, model: VulnerabilityDetector):
        """Test model can be created."""
        assert model is not None
        assert isinstance(model, torch.nn.Module)

    def test_forward_pass(self, model: VulnerabilityDetector, config: ModelConfig):
        """Test model forward pass."""
        batch_size = 4
        seq_length = 256
        
        # Create random input (0-255 for bytes, 256 for padding)
        x = torch.randint(0, 256, (batch_size, seq_length))
        
        # Forward pass returns dict
        outputs = model(x)
        
        # Check outputs
        assert outputs["is_malicious"].shape == (batch_size, 1)
        assert outputs["vulnerabilities"].shape == (batch_size, 10)  # num_vuln_classes
        assert outputs["risk_score"].shape == (batch_size, 1)
        assert outputs["features"].shape == (batch_size, config.hidden_dim)

    def test_embedding_shape(self, model: VulnerabilityDetector, config: ModelConfig):
        """Test embedding dimension."""
        batch_size = 2
        seq_length = 128
        
        x = torch.randint(0, 256, (batch_size, seq_length))
        outputs = model(x)
        
        assert outputs["features"].shape == (batch_size, config.hidden_dim)

    def test_risk_score_bounds(self, model: VulnerabilityDetector, config: ModelConfig):
        """Test risk score is bounded by sigmoid."""
        batch_size = 8
        seq_length = 256
        
        x = torch.randint(0, 256, (batch_size, seq_length))
        outputs = model(x)
        
        risk_score = outputs["risk_score"]
        # Risk score should be between 0 and 1 (sigmoid output)
        assert (risk_score >= 0).all()
        assert (risk_score <= 1).all()

    def test_eval_mode(self, model: VulnerabilityDetector, config: ModelConfig):
        """Test model works in eval mode."""
        model.eval()
        
        x = torch.randint(0, 256, (2, 128))
        
        with torch.no_grad():
            outputs = model(x)
        
        assert outputs["is_malicious"] is not None
        assert outputs["vulnerabilities"] is not None

    def test_gradient_flow(self, model: VulnerabilityDetector, config: ModelConfig):
        """Test gradients can flow through the model."""
        x = torch.randint(0, 256, (2, 128))
        
        outputs = model(x)
        
        # Compute a simple loss
        loss = outputs["is_malicious"].sum() + outputs["vulnerabilities"].sum() + outputs["risk_score"].sum()
        loss.backward()
        
        # Check some parameters have gradients
        has_grad = False
        for param in model.parameters():
            if param.grad is not None and param.grad.abs().sum() > 0:
                has_grad = True
                break
        
        assert has_grad, "No gradients computed"


class TestContrastiveLoss:
    """Test the ContrastiveLoss function."""

    @pytest.fixture
    def loss_fn(self) -> ContrastiveLoss:
        """Create loss function."""
        return ContrastiveLoss(margin=1.0)

    def test_loss_creation(self, loss_fn: ContrastiveLoss):
        """Test loss function can be created."""
        assert loss_fn is not None

    def test_loss_computation(self, loss_fn: ContrastiveLoss):
        """Test loss can be computed."""
        embed_dim = 64
        batch_size = 8
        
        # Create anchor, positive, and negative embeddings
        anchor = torch.randn(batch_size, embed_dim)
        positive = anchor + torch.randn(batch_size, embed_dim) * 0.1  # Similar
        negative = torch.randn(batch_size, embed_dim)  # Different
        
        loss = loss_fn(anchor, positive, negative)
        
        assert loss.ndim == 0  # Scalar
        assert loss >= 0  # Loss should be non-negative

    def test_loss_decreases_with_similar_embeddings(self, loss_fn: ContrastiveLoss):
        """Test that loss is lower when positive pairs are similar."""
        embed_dim = 64
        batch_size = 4
        
        # Random anchor
        anchor = torch.randn(batch_size, embed_dim)
        negative = torch.randn(batch_size, embed_dim)
        
        # Very similar positive (should have low loss)
        close_positive = anchor + torch.randn(batch_size, embed_dim) * 0.01
        loss_close = loss_fn(anchor, close_positive, negative)
        
        # Random positive (higher loss expected)
        random_positive = torch.randn(batch_size, embed_dim)
        loss_random = loss_fn(anchor, random_positive, negative)
        
        # Close positive should have lower or equal loss
        assert loss_close <= loss_random + 0.5  # Some tolerance


class TestModelConfig:
    """Test ModelConfig."""

    def test_default_config(self):
        """Test default configuration."""
        config = ModelConfig()
        
        assert config.embedding_dim == 256
        assert config.hidden_dim > 0
        assert config.num_heads > 0
        assert config.num_layers > 0
        assert config.max_sequence_length > 0
        assert 0 <= config.dropout <= 1

    def test_custom_config(self):
        """Test custom configuration."""
        config = ModelConfig(
            embedding_dim=512,
            num_heads=8,
            num_layers=6,
        )
        
        assert config.embedding_dim == 512
        assert config.num_heads == 8
        assert config.num_layers == 6
