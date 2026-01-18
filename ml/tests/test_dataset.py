"""Tests for dataset module."""

import pytest
import torch
from pathlib import Path

from vigilum_ml.dataset import (
    ContractDataset,
    VULN_TYPES,
    VULN_TO_IDX,
    load_labeled_dataset,
    create_train_val_split,
)
from vigilum_ml.models import ContractSample, VulnerabilityLabel


class TestContractDataset:
    """Test ContractDataset class."""

    @pytest.fixture
    def sample_contracts(self) -> list[ContractSample]:
        """Create sample contracts for testing."""
        return [
            ContractSample(
                contract_id="test1",
                bytecode="0x6080604052",
                is_malicious=True,
                risk_score=85.0,
                labels=[VulnerabilityLabel(vuln_type="reentrancy")],
            ),
            ContractSample(
                contract_id="test2",
                bytecode="60806040526004",
                is_malicious=False,
                risk_score=20.0,
                labels=[],
            ),
        ]

    def test_dataset_creation(self, sample_contracts: list[ContractSample]):
        """Test dataset can be created."""
        dataset = ContractDataset(sample_contracts, max_seq_length=128)
        
        assert len(dataset) == 2

    def test_dataset_getitem(self, sample_contracts: list[ContractSample]):
        """Test dataset __getitem__."""
        dataset = ContractDataset(sample_contracts, max_seq_length=128)
        
        item = dataset[0]
        
        assert "bytecode" in item
        assert "is_malicious" in item
        assert "vulnerabilities" in item
        assert "risk_score" in item
        assert item["bytecode"].shape == (128,)
        assert item["bytecode"].dtype == torch.long

    def test_bytecode_conversion(self, sample_contracts: list[ContractSample]):
        """Test bytecode is converted correctly."""
        dataset = ContractDataset(sample_contracts, max_seq_length=128)
        
        item = dataset[0]
        # First bytes should match 0x6080604052
        # 0x60 = 96, 0x80 = 128, 0x60 = 96, 0x40 = 64, 0x52 = 82
        assert item["bytecode"][0].item() == 0x60
        assert item["bytecode"][1].item() == 0x80

    def test_vulnerability_labels(self, sample_contracts: list[ContractSample]):
        """Test vulnerability labels are encoded correctly."""
        dataset = ContractDataset(sample_contracts, max_seq_length=128)
        
        item = dataset[0]  # Has reentrancy
        vuln_idx = VULN_TO_IDX["reentrancy"]
        assert item["vulnerabilities"][vuln_idx].item() == 1.0

        item2 = dataset[1]  # No vulns
        assert item2["vulnerabilities"].sum().item() == 0.0


class TestVulnTypes:
    """Test vulnerability type definitions."""

    def test_vuln_types_defined(self):
        """Test vulnerability types are defined."""
        assert len(VULN_TYPES) > 0

    def test_common_vuln_types(self):
        """Test common vulnerability types are present."""
        expected_substrings = ["reentrancy", "overflow", "access_control"]
        
        for substring in expected_substrings:
            assert any(substring in v.lower() for v in VULN_TYPES), f"Missing {substring}"

    def test_vuln_to_idx_mapping(self):
        """Test vulnerability to index mapping."""
        assert len(VULN_TO_IDX) == len(VULN_TYPES)
        
        for vuln, idx in VULN_TO_IDX.items():
            assert VULN_TYPES[idx] == vuln


class TestContractSample:
    """Test ContractSample model."""

    def test_sample_creation(self):
        """Test creating a contract sample."""
        sample = ContractSample(
            contract_id="test",
            bytecode="0x6080604052",
            is_malicious=True,
            risk_score=85.0,
            labels=[VulnerabilityLabel(vuln_type="reentrancy")],
        )
        
        assert sample.contract_id == "test"
        assert sample.is_malicious is True
        assert sample.risk_score == 85.0
        assert len(sample.labels) == 1

    def test_sample_with_optional_fields(self):
        """Test sample with optional fields."""
        sample = ContractSample(
            contract_id="test",
            bytecode="0x6080604052",
        )
        
        assert sample.is_malicious is False
        assert sample.labels == []
        assert sample.risk_score is None


class TestTrainValSplit:
    """Test train/val split function."""

    def test_split_ratio(self):
        """Test split creates correct ratio."""
        samples = [
            ContractSample(contract_id=f"test{i}", bytecode="0x60")
            for i in range(100)
        ]
        
        train, val = create_train_val_split(samples, val_ratio=0.2)
        
        assert len(train) == 80
        assert len(val) == 20

    def test_split_deterministic(self):
        """Test split is deterministic with seed."""
        # Create two separate lists for the two calls
        samples1 = [
            ContractSample(contract_id=f"test{i}", bytecode="0x60")
            for i in range(50)
        ]
        samples2 = [
            ContractSample(contract_id=f"test{i}", bytecode="0x60")
            for i in range(50)
        ]
        
        train1, val1 = create_train_val_split(samples1, seed=42)
        train2, val2 = create_train_val_split(samples2, seed=42)
        
        assert [s.contract_id for s in train1] == [s.contract_id for s in train2]
