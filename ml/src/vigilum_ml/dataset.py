"""
Dataset loading and preprocessing for vulnerability detection model training.
"""

import json
from pathlib import Path
from typing import Iterator

import polars as pl
import torch
from torch.utils.data import Dataset, IterableDataset

import structlog

from vigilum_ml.features import FeatureExtractor
from vigilum_ml.models import ContractSample, VulnerabilityLabel

logger = structlog.get_logger(__name__)


# Vulnerability type to index mapping
VULN_TYPES: list[str] = [
    "reentrancy",
    "access_control",
    "integer_overflow",
    "integer_underflow",
    "unchecked_call",
    "tx_origin",
    "timestamp_dependency",
    "flash_loan",
    "oracle_manipulation",
    "rug_pull",
    "honeypot",
    "logic_error",
    "denial_of_service",
]

VULN_TO_IDX: dict[str, int] = {v: i for i, v in enumerate(VULN_TYPES)}


class ContractDataset(Dataset):
    """
    In-memory dataset for contract bytecode and labels.
    """

    def __init__(
        self,
        samples: list[ContractSample],
        max_seq_length: int = 8192,
        feature_extractor: FeatureExtractor | None = None,
    ):
        self.samples = samples
        self.max_seq_length = max_seq_length
        self.feature_extractor = feature_extractor or FeatureExtractor()
        self.num_vuln_classes = len(VULN_TYPES)

    def __len__(self) -> int:
        return len(self.samples)

    def __getitem__(self, idx: int) -> dict:
        sample = self.samples[idx]
        
        # Convert bytecode to tensor
        bytecode_tensor = self._bytecode_to_tensor(sample.bytecode)
        
        # Create attention mask
        attention_mask = torch.ones(len(bytecode_tensor), dtype=torch.long)
        
        # Pad to max length
        if len(bytecode_tensor) < self.max_seq_length:
            pad_len = self.max_seq_length - len(bytecode_tensor)
            bytecode_tensor = torch.cat([
                bytecode_tensor,
                torch.full((pad_len,), 256, dtype=torch.long),  # Padding token
            ])
            attention_mask = torch.cat([
                attention_mask,
                torch.zeros(pad_len, dtype=torch.long),
            ])
        else:
            bytecode_tensor = bytecode_tensor[:self.max_seq_length]
            attention_mask = attention_mask[:self.max_seq_length]

        # Create vulnerability labels (multi-label)
        vuln_labels = torch.zeros(self.num_vuln_classes, dtype=torch.float)
        if sample.labels:
            for label in sample.labels:
                if label.vuln_type in VULN_TO_IDX:
                    vuln_labels[VULN_TO_IDX[label.vuln_type]] = 1.0

        return {
            "bytecode": bytecode_tensor,
            "attention_mask": attention_mask,
            "is_malicious": torch.tensor([1.0 if sample.is_malicious else 0.0]),
            "vulnerabilities": vuln_labels,
            "risk_score": torch.tensor([sample.risk_score / 100.0 if sample.risk_score else 0.0]),
            "contract_id": sample.contract_id,
        }

    def _bytecode_to_tensor(self, bytecode_hex: str) -> torch.Tensor:
        """Convert hex bytecode to tensor of byte values."""
        bytecode_hex = bytecode_hex.lower()
        if bytecode_hex.startswith("0x"):
            bytecode_hex = bytecode_hex[2:]
        
        try:
            bytecode = bytes.fromhex(bytecode_hex)
            return torch.tensor(list(bytecode), dtype=torch.long)
        except ValueError:
            # Return empty tensor for invalid bytecode
            return torch.tensor([], dtype=torch.long)


class StreamingContractDataset(IterableDataset):
    """
    Streaming dataset for large-scale training from parquet files.
    """

    def __init__(
        self,
        data_dir: Path,
        max_seq_length: int = 8192,
        shuffle: bool = True,
    ):
        self.data_dir = Path(data_dir)
        self.max_seq_length = max_seq_length
        self.shuffle = shuffle
        self.num_vuln_classes = len(VULN_TYPES)
        
        # Find all parquet files
        self.files = list(self.data_dir.glob("*.parquet"))
        if not self.files:
            raise ValueError(f"No parquet files found in {data_dir}")
        
        logger.info("StreamingDataset initialized", num_files=len(self.files))

    def __iter__(self) -> Iterator[dict]:
        worker_info = torch.utils.data.get_worker_info()
        
        files = self.files
        if worker_info is not None:
            # Split files across workers
            per_worker = len(files) // worker_info.num_workers
            start = worker_info.id * per_worker
            end = start + per_worker if worker_info.id < worker_info.num_workers - 1 else len(files)
            files = files[start:end]
        
        for file_path in files:
            df = pl.read_parquet(file_path)
            
            if self.shuffle:
                df = df.sample(fraction=1.0, shuffle=True)
            
            for row in df.iter_rows(named=True):
                yield self._process_row(row)

    def _process_row(self, row: dict) -> dict:
        """Process a single row from parquet file."""
        # Convert bytecode
        bytecode_hex = row.get("bytecode", "")
        bytecode_tensor = self._bytecode_to_tensor(bytecode_hex)
        
        # Create attention mask
        attention_mask = torch.ones(len(bytecode_tensor), dtype=torch.long)
        
        # Pad or truncate
        if len(bytecode_tensor) < self.max_seq_length:
            pad_len = self.max_seq_length - len(bytecode_tensor)
            bytecode_tensor = torch.cat([
                bytecode_tensor,
                torch.full((pad_len,), 256, dtype=torch.long),
            ])
            attention_mask = torch.cat([
                attention_mask,
                torch.zeros(pad_len, dtype=torch.long),
            ])
        else:
            bytecode_tensor = bytecode_tensor[:self.max_seq_length]
            attention_mask = attention_mask[:self.max_seq_length]

        # Parse vulnerability labels
        vuln_labels = torch.zeros(self.num_vuln_classes, dtype=torch.float)
        vulns = row.get("vulnerabilities", [])
        if isinstance(vulns, str):
            vulns = json.loads(vulns) if vulns else []
        for vuln_type in vulns:
            if vuln_type in VULN_TO_IDX:
                vuln_labels[VULN_TO_IDX[vuln_type]] = 1.0

        return {
            "bytecode": bytecode_tensor,
            "attention_mask": attention_mask,
            "is_malicious": torch.tensor([float(row.get("is_malicious", 0))]),
            "vulnerabilities": vuln_labels,
            "risk_score": torch.tensor([row.get("risk_score", 0) / 100.0]),
            "contract_id": row.get("contract_id", "unknown"),
        }

    def _bytecode_to_tensor(self, bytecode_hex: str) -> torch.Tensor:
        """Convert hex bytecode to tensor of byte values."""
        bytecode_hex = str(bytecode_hex).lower()
        if bytecode_hex.startswith("0x"):
            bytecode_hex = bytecode_hex[2:]
        
        try:
            bytecode = bytes.fromhex(bytecode_hex)
            return torch.tensor(list(bytecode), dtype=torch.long)
        except ValueError:
            return torch.tensor([], dtype=torch.long)


def collate_fn(batch: list[dict]) -> dict:
    """Custom collate function for batching."""
    return {
        "bytecode": torch.stack([x["bytecode"] for x in batch]),
        "attention_mask": torch.stack([x["attention_mask"] for x in batch]),
        "is_malicious": torch.stack([x["is_malicious"] for x in batch]),
        "vulnerabilities": torch.stack([x["vulnerabilities"] for x in batch]),
        "risk_score": torch.stack([x["risk_score"] for x in batch]),
        "contract_ids": [x["contract_id"] for x in batch],
    }


def load_labeled_dataset(data_path: Path) -> list[ContractSample]:
    """
    Load labeled dataset from JSON or Parquet file.
    
    Expected format:
    - JSON: List of ContractSample dicts
    - Parquet: DataFrame with columns matching ContractSample
    """
    data_path = Path(data_path)
    
    if data_path.suffix == ".json":
        with open(data_path) as f:
            data = json.load(f)
        return [ContractSample(**item) for item in data]
    
    elif data_path.suffix == ".parquet":
        df = pl.read_parquet(data_path)
        samples = []
        for row in df.iter_rows(named=True):
            labels = []
            vulns = row.get("vulnerabilities", [])
            if isinstance(vulns, str):
                vulns = json.loads(vulns) if vulns else []
            for vuln in vulns:
                labels.append(VulnerabilityLabel(vuln_type=vuln))
            
            samples.append(ContractSample(
                contract_id=row.get("contract_id", "unknown"),
                bytecode=row.get("bytecode", ""),
                source_code=row.get("source_code"),
                is_malicious=row.get("is_malicious", False),
                risk_score=row.get("risk_score"),
                labels=labels,
            ))
        return samples
    
    else:
        raise ValueError(f"Unsupported file format: {data_path.suffix}")


def create_train_val_split(
    samples: list[ContractSample],
    val_ratio: float = 0.1,
    seed: int = 42,
) -> tuple[list[ContractSample], list[ContractSample]]:
    """Split samples into train and validation sets."""
    import random
    random.seed(seed)
    random.shuffle(samples)
    
    split_idx = int(len(samples) * (1 - val_ratio))
    return samples[:split_idx], samples[split_idx:]
