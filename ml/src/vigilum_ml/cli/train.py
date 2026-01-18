#!/usr/bin/env python3
"""
Training CLI for VIGILUM ML vulnerability detection model.

Usage:
    python -m vigilum_ml.cli.train --config config.json
    python -m vigilum_ml.cli.train --data data.parquet --output checkpoints/
"""

import argparse
import json
import sys
from pathlib import Path

import structlog

structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer(),
    ]
)

logger = structlog.get_logger(__name__)


def main():
    parser = argparse.ArgumentParser(description="Train VIGILUM vulnerability detection model")
    
    parser.add_argument(
        "--config",
        type=Path,
        help="Path to training configuration JSON file",
    )
    parser.add_argument(
        "--data",
        type=Path,
        help="Path to training data (parquet or JSON)",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path("checkpoints"),
        help="Output directory for checkpoints",
    )
    parser.add_argument(
        "--epochs",
        type=int,
        default=50,
        help="Number of training epochs",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=32,
        help="Training batch size",
    )
    parser.add_argument(
        "--learning-rate",
        type=float,
        default=1e-4,
        help="Learning rate",
    )
    parser.add_argument(
        "--device",
        type=str,
        default="cuda",
        choices=["cuda", "cpu"],
        help="Device to train on",
    )
    
    args = parser.parse_args()
    
    # Import here to avoid slow startup for help
    import torch
    from vigilum_ml.training import train_from_config, Trainer
    from vigilum_ml.dataset import (
        ContractDataset,
        load_labeled_dataset,
        create_train_val_split,
        collate_fn,
        VULN_TYPES,
    )
    from vigilum_ml.model import VulnerabilityDetector
    from vigilum_ml.models import ModelConfig, TrainingConfig
    from torch.utils.data import DataLoader
    
    if args.config:
        # Train from config file
        logger.info("Training from config", config=str(args.config))
        results = train_from_config(args.config)
    elif args.data:
        # Train with CLI arguments
        logger.info(
            "Training with CLI arguments",
            data=str(args.data),
            epochs=args.epochs,
            batch_size=args.batch_size,
        )
        
        # Check device
        device = torch.device(args.device if torch.cuda.is_available() else "cpu")
        if args.device == "cuda" and not torch.cuda.is_available():
            logger.warning("CUDA not available, falling back to CPU")
        
        # Load data
        samples = load_labeled_dataset(args.data)
        train_samples, val_samples = create_train_val_split(samples)
        
        logger.info(
            "Data loaded",
            train_samples=len(train_samples),
            val_samples=len(val_samples),
        )
        
        # Create configs
        model_config = ModelConfig()
        train_config = TrainingConfig(
            num_epochs=args.epochs,
            batch_size=args.batch_size,
            learning_rate=args.learning_rate,
        )
        
        # Create datasets
        train_dataset = ContractDataset(
            train_samples,
            max_seq_length=model_config.max_sequence_length,
        )
        val_dataset = ContractDataset(
            val_samples,
            max_seq_length=model_config.max_sequence_length,
        )
        
        # Create loaders
        train_loader = DataLoader(
            train_dataset,
            batch_size=train_config.batch_size,
            shuffle=True,
            collate_fn=collate_fn,
            num_workers=train_config.num_workers,
        )
        val_loader = DataLoader(
            val_dataset,
            batch_size=train_config.batch_size,
            shuffle=False,
            collate_fn=collate_fn,
            num_workers=train_config.num_workers,
        )
        
        # Create model
        model = VulnerabilityDetector(model_config, num_vuln_classes=len(VULN_TYPES))
        
        # Train
        trainer = Trainer(model, train_config, model_config, device)
        results = trainer.train(train_loader, val_loader, args.output)
    else:
        parser.print_help()
        sys.exit(1)
    
    # Print results
    logger.info(
        "Training complete",
        best_val_loss=results.get("best_val_loss"),
        total_time=f"{results.get('total_time_seconds', 0):.1f}s",
    )
    
    # Save training results
    results_path = args.output / "training_results.json"
    with open(results_path, "w") as f:
        # Convert non-serializable values
        serializable = {
            k: v if not isinstance(v, (list, tuple)) or len(v) < 100 else f"[{len(v)} values]"
            for k, v in results.items()
        }
        json.dump(serializable, f, indent=2)
    
    logger.info("Results saved", path=str(results_path))


if __name__ == "__main__":
    main()
