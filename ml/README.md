# VIGILUM ML - Smart Contract Vulnerability Detection

Machine learning pipeline for detecting vulnerabilities in smart contract bytecode.

## Overview

VIGILUM ML provides a deep learning-based approach to smart contract security analysis:

- **Vulnerability Detection**: Multi-label classification for 10+ vulnerability types
- **Risk Scoring**: Continuous risk score prediction (0-1)
- **Malicious Contract Detection**: Binary classification for known malicious patterns
- **Embedding Generation**: Semantic embeddings for contract similarity search

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Raw Bytecode   │ ──► │   Transformer    │ ──► │  Multi-Head     │
│  (hex → tokens) │     │   Encoder        │     │  Output Layer   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                                       │
                        ┌──────────────────────────────┼──────────────────────────────┐
                        │                              │                               │
                        ▼                              ▼                               ▼
                 ┌─────────────┐              ┌─────────────┐               ┌─────────────────┐
                 │  Malicious  │              │   Vuln      │               │   Risk Score    │
                 │  Detection  │              │  Detection  │               │   Regression    │
                 └─────────────┘              └─────────────┘               └─────────────────┘
```

## Features

### Model
- Transformer-based architecture optimized for bytecode sequences
- Multi-task learning with shared representations
- Contrastive learning for improved embeddings

### Training
- Mixed precision training (fp16)
- Gradient accumulation for large effective batch sizes
- Cosine annealing learning rate schedule
- Early stopping with validation loss monitoring

### Inference
- PyTorch and ONNX runtime backends
- Batch processing support
- Embedding similarity search

## Installation

```bash
# Development installation
pip install -e ".[dev]"

# Production installation
pip install vigilum-ml
```

## Usage

### Training

```python
from vigilum_ml import VulnerabilityDetector, Trainer, TrainingConfig, ModelConfig
from vigilum_ml.dataset import load_labeled_dataset, create_train_val_split

# Load data
samples = load_labeled_dataset("data/contracts.parquet")
train_samples, val_samples = create_train_val_split(samples)

# Configure
model_config = ModelConfig(embed_dim=256, num_layers=6)
train_config = TrainingConfig(num_epochs=50, batch_size=32)

# Train
model = VulnerabilityDetector(model_config, num_vuln_classes=10)
trainer = Trainer(model, train_config, model_config)
results = trainer.train(train_loader, val_loader, output_dir)
```

### CLI Training

```bash
# From config file
python -m vigilum_ml.cli.train --config config.json

# With CLI arguments
python -m vigilum_ml.cli.train --data data.parquet --epochs 50 --batch-size 32
```

### Inference

```python
from vigilum_ml import InferenceService

# Load model
service = InferenceService.from_checkpoint("checkpoints/best_model.pt")

# Predict
result = service.predict("0x6080604052...")
print(f"Risk Score: {result.risk_score}")
print(f"Vulnerabilities: {result.vulnerabilities}")
print(f"Is Malicious: {result.is_malicious}")

# Get embedding for similarity search
embedding = service.get_embedding("0x6080604052...")
```

### ONNX Export

```python
from vigilum_ml import InferenceService

service = InferenceService.from_checkpoint("model.pt")
service.export_to_onnx("model.onnx")
```

## Vulnerability Types

| Type | Description |
|------|-------------|
| REENTRANCY | Reentrancy attack patterns |
| OVERFLOW | Integer overflow/underflow |
| ACCESS_CONTROL | Missing access restrictions |
| UNCHECKED_CALL | Unchecked external calls |
| FRONT_RUNNING | Front-running vulnerabilities |
| DENIAL_OF_SERVICE | DoS attack vectors |
| FLASH_LOAN | Flash loan attack patterns |
| ORACLE_MANIPULATION | Price oracle manipulation |
| TIMESTAMP | Timestamp dependency issues |
| SELFDESTRUCT | Dangerous selfdestruct usage |

## Data Format

### Parquet/JSON Schema

```json
{
  "address": "0x...",
  "bytecode": "0x6080604052...",
  "is_malicious": true,
  "vulnerabilities": ["reentrancy", "access_control"],
  "risk_score": 0.85
}
```

## Development

```bash
# Run tests
pytest tests/

# Type checking
mypy src/vigilum_ml

# Linting
ruff check src/vigilum_ml

# Format
ruff format src/vigilum_ml
```

## Project Structure

```
ml/
├── src/vigilum_ml/
│   ├── __init__.py          # Package exports
│   ├── model.py             # VulnerabilityDetector model
│   ├── models.py            # Pydantic data models
│   ├── features.py          # Feature extraction
│   ├── dataset.py           # Dataset loaders
│   ├── training.py          # Training pipeline
│   ├── cli/                 # CLI tools
│   │   └── train.py         # Training CLI
│   └── inference/           # Inference service
│       └── service.py       # InferenceService
├── tests/                   # Unit tests
├── notebooks/               # Jupyter notebooks
└── data/                    # Training data
```

## License

MIT License
