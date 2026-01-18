"""
Training pipeline for vulnerability detection model.
"""

import json
import time
from pathlib import Path
from typing import Any

import torch
import torch.nn as nn
from torch.optim import AdamW
from torch.optim.lr_scheduler import CosineAnnealingLR
from torch.utils.data import DataLoader

import structlog

from vigilum_ml.dataset import (
    ContractDataset,
    StreamingContractDataset,
    collate_fn,
    load_labeled_dataset,
    create_train_val_split,
    VULN_TYPES,
)
from vigilum_ml.model import VulnerabilityDetector, ContrastiveLoss
from vigilum_ml.models import ModelConfig, TrainingConfig

logger = structlog.get_logger(__name__)


class Trainer:
    """
    Training orchestrator for vulnerability detection model.
    """

    def __init__(
        self,
        model: VulnerabilityDetector,
        train_config: TrainingConfig,
        model_config: ModelConfig,
        device: torch.device | None = None,
    ):
        self.model = model
        self.train_config = train_config
        self.model_config = model_config
        self.device = device or torch.device("cuda" if torch.cuda.is_available() else "cpu")
        
        self.model.to(self.device)
        
        # Optimizers
        self.optimizer = AdamW(
            model.parameters(),
            lr=train_config.learning_rate,
            weight_decay=train_config.weight_decay,
        )
        
        # Loss functions
        self.bce_loss = nn.BCELoss()
        self.mse_loss = nn.MSELoss()
        self.contrastive_loss = ContrastiveLoss(margin=1.0)
        
        # Metrics tracking
        self.train_losses: list[float] = []
        self.val_losses: list[float] = []
        self.best_val_loss = float("inf")
        
        # Logging
        self.logger = logger.bind(component="trainer")

    def train(
        self,
        train_loader: DataLoader,
        val_loader: DataLoader | None = None,
        output_dir: Path | None = None,
    ) -> dict[str, Any]:
        """
        Train the model.
        
        Args:
            train_loader: Training data loader
            val_loader: Validation data loader (optional)
            output_dir: Directory to save checkpoints
            
        Returns:
            Training metrics dictionary
        """
        output_dir = Path(output_dir) if output_dir else Path("checkpoints")
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Learning rate scheduler
        scheduler = CosineAnnealingLR(
            self.optimizer,
            T_max=self.train_config.num_epochs,
            eta_min=self.train_config.learning_rate * 0.01,
        )
        
        self.logger.info(
            "Starting training",
            epochs=self.train_config.num_epochs,
            device=str(self.device),
            train_batches=len(train_loader),
        )
        
        start_time = time.time()
        
        for epoch in range(self.train_config.num_epochs):
            # Training phase
            train_loss = self._train_epoch(train_loader, epoch)
            self.train_losses.append(train_loss)
            
            # Validation phase
            val_loss = None
            if val_loader:
                val_loss = self._validate_epoch(val_loader)
                self.val_losses.append(val_loss)
                
                # Save best model
                if val_loss < self.best_val_loss:
                    self.best_val_loss = val_loss
                    self._save_checkpoint(output_dir / "best_model.pt", epoch)
            
            # Step scheduler
            scheduler.step()
            
            # Logging
            self.logger.info(
                "Epoch completed",
                epoch=epoch + 1,
                train_loss=f"{train_loss:.4f}",
                val_loss=f"{val_loss:.4f}" if val_loss else "N/A",
                lr=f"{scheduler.get_last_lr()[0]:.6f}",
            )
            
            # Periodic checkpoint
            if (epoch + 1) % self.train_config.save_every == 0:
                self._save_checkpoint(output_dir / f"checkpoint_epoch_{epoch+1}.pt", epoch)
        
        # Final save
        self._save_checkpoint(output_dir / "final_model.pt", self.train_config.num_epochs - 1)
        
        total_time = time.time() - start_time
        
        return {
            "train_losses": self.train_losses,
            "val_losses": self.val_losses,
            "best_val_loss": self.best_val_loss,
            "total_time_seconds": total_time,
            "final_epoch": self.train_config.num_epochs,
        }

    def _train_epoch(self, dataloader: DataLoader, epoch: int) -> float:
        """Train for one epoch."""
        self.model.train()
        total_loss = 0.0
        num_batches = 0
        
        for batch in dataloader:
            # Move to device
            bytecode = batch["bytecode"].to(self.device)
            attention_mask = batch["attention_mask"].to(self.device)
            is_malicious = batch["is_malicious"].to(self.device)
            vulnerabilities = batch["vulnerabilities"].to(self.device)
            risk_score = batch["risk_score"].to(self.device)
            
            # Forward pass
            self.optimizer.zero_grad()
            outputs = self.model(bytecode, attention_mask)
            
            # Calculate losses
            loss_malicious = self.bce_loss(outputs["is_malicious"], is_malicious)
            loss_vuln = self.bce_loss(outputs["vulnerabilities"], vulnerabilities)
            loss_risk = self.mse_loss(outputs["risk_score"], risk_score)
            
            # Combined loss with weighting
            loss = (
                self.train_config.malicious_weight * loss_malicious +
                self.train_config.vuln_weight * loss_vuln +
                self.train_config.risk_weight * loss_risk
            )
            
            # Backward pass
            loss.backward()
            
            # Gradient clipping
            torch.nn.utils.clip_grad_norm_(
                self.model.parameters(),
                self.train_config.gradient_clip,
            )
            
            self.optimizer.step()
            
            total_loss += loss.item()
            num_batches += 1
        
        return total_loss / max(num_batches, 1)

    def _validate_epoch(self, dataloader: DataLoader) -> float:
        """Validate for one epoch."""
        self.model.eval()
        total_loss = 0.0
        num_batches = 0
        
        with torch.no_grad():
            for batch in dataloader:
                bytecode = batch["bytecode"].to(self.device)
                attention_mask = batch["attention_mask"].to(self.device)
                is_malicious = batch["is_malicious"].to(self.device)
                vulnerabilities = batch["vulnerabilities"].to(self.device)
                risk_score = batch["risk_score"].to(self.device)
                
                outputs = self.model(bytecode, attention_mask)
                
                loss_malicious = self.bce_loss(outputs["is_malicious"], is_malicious)
                loss_vuln = self.bce_loss(outputs["vulnerabilities"], vulnerabilities)
                loss_risk = self.mse_loss(outputs["risk_score"], risk_score)
                
                loss = (
                    self.train_config.malicious_weight * loss_malicious +
                    self.train_config.vuln_weight * loss_vuln +
                    self.train_config.risk_weight * loss_risk
                )
                
                total_loss += loss.item()
                num_batches += 1
        
        return total_loss / max(num_batches, 1)

    def _save_checkpoint(self, path: Path, epoch: int) -> None:
        """Save model checkpoint."""
        checkpoint = {
            "epoch": epoch,
            "model_state_dict": self.model.state_dict(),
            "optimizer_state_dict": self.optimizer.state_dict(),
            "model_config": self.model_config.__dict__,
            "train_config": self.train_config.__dict__,
            "train_losses": self.train_losses,
            "val_losses": self.val_losses,
            "best_val_loss": self.best_val_loss,
        }
        torch.save(checkpoint, path)
        self.logger.info("Checkpoint saved", path=str(path))


def train_from_config(config_path: Path) -> dict[str, Any]:
    """
    Train model from configuration file.
    
    Config file should contain:
    - model_config: ModelConfig fields
    - train_config: TrainingConfig fields
    - data_path: Path to training data
    - output_dir: Path for checkpoints
    """
    with open(config_path) as f:
        config = json.load(f)
    
    # Load configs
    model_config = ModelConfig(**config.get("model_config", {}))
    train_config = TrainingConfig(**config.get("train_config", {}))
    
    # Load data
    data_path = Path(config["data_path"])
    if data_path.is_dir():
        # Streaming from multiple parquet files
        train_dataset = StreamingContractDataset(
            data_path / "train",
            max_seq_length=model_config.max_sequence_length,
        )
        val_dataset = StreamingContractDataset(
            data_path / "val",
            max_seq_length=model_config.max_sequence_length,
            shuffle=False,
        )
    else:
        # Single file dataset
        samples = load_labeled_dataset(data_path)
        train_samples, val_samples = create_train_val_split(samples)
        
        train_dataset = ContractDataset(
            train_samples,
            max_seq_length=model_config.max_sequence_length,
        )
        val_dataset = ContractDataset(
            val_samples,
            max_seq_length=model_config.max_sequence_length,
        )
    
    # Create data loaders
    train_loader = DataLoader(
        train_dataset,
        batch_size=train_config.batch_size,
        shuffle=True,
        collate_fn=collate_fn,
        num_workers=train_config.num_workers,
        pin_memory=True,
    )
    val_loader = DataLoader(
        val_dataset,
        batch_size=train_config.batch_size,
        shuffle=False,
        collate_fn=collate_fn,
        num_workers=train_config.num_workers,
        pin_memory=True,
    )
    
    # Create model
    model = VulnerabilityDetector(model_config, num_vuln_classes=len(VULN_TYPES))
    
    # Create trainer
    trainer = Trainer(model, train_config, model_config)
    
    # Train
    output_dir = Path(config.get("output_dir", "checkpoints"))
    return trainer.train(train_loader, val_loader, output_dir)


def load_model_from_checkpoint(
    checkpoint_path: Path,
    device: torch.device | None = None,
) -> VulnerabilityDetector:
    """Load trained model from checkpoint."""
    device = device or torch.device("cuda" if torch.cuda.is_available() else "cpu")
    
    checkpoint = torch.load(checkpoint_path, map_location=device)
    model_config = ModelConfig(**checkpoint["model_config"])
    
    model = VulnerabilityDetector(model_config, num_vuln_classes=len(VULN_TYPES))
    model.load_state_dict(checkpoint["model_state_dict"])
    model.to(device)
    model.eval()
    
    return model
