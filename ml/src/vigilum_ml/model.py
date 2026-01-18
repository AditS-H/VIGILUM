"""
Neural network architecture for vulnerability detection.
"""

import torch
import torch.nn as nn
import torch.nn.functional as F

from vigilum_ml.models import ModelConfig


class BytecodeEmbedding(nn.Module):
    """Embeds raw bytecode into dense vectors."""

    def __init__(self, config: ModelConfig):
        super().__init__()
        self.config = config
        
        # Byte embedding (256 possible values + padding)
        self.byte_embed = nn.Embedding(257, config.embedding_dim, padding_idx=256)
        
        # Positional encoding
        self.pos_embed = nn.Embedding(config.max_sequence_length, config.embedding_dim)
        
        # Layer norm
        self.norm = nn.LayerNorm(config.embedding_dim)
        self.dropout = nn.Dropout(config.dropout)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """
        Args:
            x: Bytecode tensor of shape (batch, seq_len)
            
        Returns:
            Embedded tensor of shape (batch, seq_len, embedding_dim)
        """
        batch_size, seq_len = x.shape
        
        # Byte embeddings
        byte_emb = self.byte_embed(x)
        
        # Positional embeddings
        positions = torch.arange(seq_len, device=x.device).unsqueeze(0).expand(batch_size, -1)
        pos_emb = self.pos_embed(positions)
        
        # Combine and normalize
        emb = self.norm(byte_emb + pos_emb)
        return self.dropout(emb)


class VulnerabilityDetector(nn.Module):
    """
    Transformer-based vulnerability detection model.
    
    Architecture:
    1. Bytecode embedding layer
    2. Transformer encoder blocks
    3. Global pooling
    4. Multi-task classification heads
    """

    def __init__(self, config: ModelConfig, num_vuln_classes: int = 13):
        super().__init__()
        self.config = config
        self.num_vuln_classes = num_vuln_classes
        
        # Embedding
        self.embedding = BytecodeEmbedding(config)
        
        # Transformer encoder
        encoder_layer = nn.TransformerEncoderLayer(
            d_model=config.embedding_dim,
            nhead=config.num_heads,
            dim_feedforward=config.hidden_dim,
            dropout=config.dropout,
            activation=F.gelu,
            batch_first=True,
        )
        self.transformer = nn.TransformerEncoder(
            encoder_layer,
            num_layers=config.num_layers,
        )
        
        # Feature extraction heads
        self.feature_proj = nn.Sequential(
            nn.Linear(config.embedding_dim, config.hidden_dim),
            nn.GELU(),
            nn.Dropout(config.dropout),
            nn.Linear(config.hidden_dim, config.hidden_dim),
        )
        
        # Classification heads
        self.malicious_head = nn.Sequential(
            nn.Linear(config.hidden_dim, config.hidden_dim // 2),
            nn.GELU(),
            nn.Dropout(config.dropout),
            nn.Linear(config.hidden_dim // 2, 1),
        )
        
        self.vuln_head = nn.Sequential(
            nn.Linear(config.hidden_dim, config.hidden_dim // 2),
            nn.GELU(),
            nn.Dropout(config.dropout),
            nn.Linear(config.hidden_dim // 2, num_vuln_classes),
        )
        
        self.risk_head = nn.Sequential(
            nn.Linear(config.hidden_dim, config.hidden_dim // 2),
            nn.GELU(),
            nn.Dropout(config.dropout),
            nn.Linear(config.hidden_dim // 2, 1),
            nn.Sigmoid(),  # Output 0-1, scale to 0-100 later
        )

    def forward(
        self,
        bytecode: torch.Tensor,
        attention_mask: torch.Tensor | None = None,
    ) -> dict[str, torch.Tensor]:
        """
        Forward pass for vulnerability detection.
        
        Args:
            bytecode: Bytecode tensor (batch, seq_len)
            attention_mask: Optional mask for padding (batch, seq_len)
            
        Returns:
            Dictionary with:
            - is_malicious: (batch, 1) - Binary malicious probability
            - vulnerabilities: (batch, num_classes) - Per-class probabilities
            - risk_score: (batch, 1) - Risk score 0-1
            - features: (batch, hidden_dim) - Extracted features
        """
        # Embed bytecode
        x = self.embedding(bytecode)
        
        # Create attention mask for transformer
        if attention_mask is not None:
            # Convert to format expected by transformer
            src_key_padding_mask = ~attention_mask.bool()
        else:
            src_key_padding_mask = None
        
        # Transformer encoding
        encoded = self.transformer(x, src_key_padding_mask=src_key_padding_mask)
        
        # Global average pooling
        if attention_mask is not None:
            # Masked mean pooling
            mask = attention_mask.unsqueeze(-1).float()
            pooled = (encoded * mask).sum(dim=1) / mask.sum(dim=1).clamp(min=1)
        else:
            pooled = encoded.mean(dim=1)
        
        # Feature projection
        features = self.feature_proj(pooled)
        
        # Multi-task outputs
        is_malicious = torch.sigmoid(self.malicious_head(features))
        vulnerabilities = torch.sigmoid(self.vuln_head(features))
        risk_score = self.risk_head(features)
        
        return {
            "is_malicious": is_malicious,
            "vulnerabilities": vulnerabilities,
            "risk_score": risk_score,
            "features": features,
        }

    def get_embedding(self, bytecode: torch.Tensor) -> torch.Tensor:
        """Extract contract embedding for similarity search."""
        x = self.embedding(bytecode)
        encoded = self.transformer(x)
        pooled = encoded.mean(dim=1)
        return self.feature_proj(pooled)


class ContrastiveLoss(nn.Module):
    """
    Contrastive loss for learning similar/dissimilar contract embeddings.
    Useful for detecting clones and variants of known malicious contracts.
    """

    def __init__(self, margin: float = 1.0):
        super().__init__()
        self.margin = margin

    def forward(
        self,
        anchor: torch.Tensor,
        positive: torch.Tensor,
        negative: torch.Tensor,
    ) -> torch.Tensor:
        """
        Triplet loss: anchor should be closer to positive than negative.
        
        Args:
            anchor: Anchor embeddings (batch, dim)
            positive: Similar contract embeddings (batch, dim)
            negative: Dissimilar contract embeddings (batch, dim)
        """
        pos_dist = F.pairwise_distance(anchor, positive)
        neg_dist = F.pairwise_distance(anchor, negative)
        
        loss = F.relu(pos_dist - neg_dist + self.margin)
        return loss.mean()


class MultiTaskLoss(nn.Module):
    """Combined loss for multi-task vulnerability detection."""

    def __init__(
        self,
        malicious_weight: float = 1.0,
        vuln_weight: float = 1.0,
        risk_weight: float = 0.5,
    ):
        super().__init__()
        self.malicious_weight = malicious_weight
        self.vuln_weight = vuln_weight
        self.risk_weight = risk_weight
        
        self.bce = nn.BCELoss()
        self.mse = nn.MSELoss()

    def forward(
        self,
        outputs: dict[str, torch.Tensor],
        targets: dict[str, torch.Tensor],
    ) -> dict[str, torch.Tensor]:
        """
        Compute combined multi-task loss.
        
        Args:
            outputs: Model outputs
            targets: Ground truth labels
            
        Returns:
            Dictionary with individual and total losses
        """
        # Binary classification loss for malicious detection
        malicious_loss = self.bce(
            outputs["is_malicious"],
            targets["is_malicious"].float().unsqueeze(-1),
        )
        
        # Multi-label classification loss for vulnerabilities
        vuln_loss = self.bce(
            outputs["vulnerabilities"],
            targets["vulnerabilities"].float(),
        )
        
        # Regression loss for risk score
        risk_loss = self.mse(
            outputs["risk_score"].squeeze(-1),
            targets["risk_score"].float() / 100,  # Normalize to 0-1
        )
        
        # Weighted total
        total = (
            self.malicious_weight * malicious_loss
            + self.vuln_weight * vuln_loss
            + self.risk_weight * risk_loss
        )
        
        return {
            "total": total,
            "malicious": malicious_loss,
            "vulnerability": vuln_loss,
            "risk": risk_loss,
        }
