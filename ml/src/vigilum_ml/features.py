"""
Feature extraction from smart contract bytecode and source code.
"""

import re
from collections import Counter
from math import log2

import structlog

from vigilum_ml.models import ContractFeatures

logger = structlog.get_logger(__name__)

# EVM Opcode mappings (hex -> name)
OPCODES: dict[int, str] = {
    0x00: "STOP",
    0x01: "ADD",
    0x02: "MUL",
    0x03: "SUB",
    0x04: "DIV",
    0x05: "SDIV",
    0x06: "MOD",
    0x10: "LT",
    0x11: "GT",
    0x14: "EQ",
    0x15: "ISZERO",
    0x16: "AND",
    0x17: "OR",
    0x18: "XOR",
    0x19: "NOT",
    0x1A: "BYTE",
    0x1B: "SHL",
    0x1C: "SHR",
    0x20: "SHA3",
    0x30: "ADDRESS",
    0x31: "BALANCE",
    0x32: "ORIGIN",
    0x33: "CALLER",
    0x34: "CALLVALUE",
    0x35: "CALLDATALOAD",
    0x36: "CALLDATASIZE",
    0x37: "CALLDATACOPY",
    0x38: "CODESIZE",
    0x39: "CODECOPY",
    0x3A: "GASPRICE",
    0x3B: "EXTCODESIZE",
    0x3C: "EXTCODECOPY",
    0x3D: "RETURNDATASIZE",
    0x3E: "RETURNDATACOPY",
    0x3F: "EXTCODEHASH",
    0x40: "BLOCKHASH",
    0x41: "COINBASE",
    0x42: "TIMESTAMP",
    0x43: "NUMBER",
    0x44: "DIFFICULTY",
    0x45: "GASLIMIT",
    0x46: "CHAINID",
    0x47: "SELFBALANCE",
    0x48: "BASEFEE",
    0x50: "POP",
    0x51: "MLOAD",
    0x52: "MSTORE",
    0x53: "MSTORE8",
    0x54: "SLOAD",
    0x55: "SSTORE",
    0x56: "JUMP",
    0x57: "JUMPI",
    0x58: "PC",
    0x59: "MSIZE",
    0x5A: "GAS",
    0x5B: "JUMPDEST",
    0x5F: "PUSH0",
    # PUSH1-PUSH32: 0x60-0x7F
    # DUP1-DUP16: 0x80-0x8F
    # SWAP1-SWAP16: 0x90-0x9F
    0xA0: "LOG0",
    0xA1: "LOG1",
    0xA2: "LOG2",
    0xA3: "LOG3",
    0xA4: "LOG4",
    0xF0: "CREATE",
    0xF1: "CALL",
    0xF2: "CALLCODE",
    0xF3: "RETURN",
    0xF4: "DELEGATECALL",
    0xF5: "CREATE2",
    0xFA: "STATICCALL",
    0xFD: "REVERT",
    0xFE: "INVALID",
    0xFF: "SELFDESTRUCT",
}

# Known function selectors
KNOWN_SELECTORS: dict[str, str] = {
    "0x70a08231": "balanceOf(address)",
    "0xa9059cbb": "transfer(address,uint256)",
    "0x23b872dd": "transferFrom(address,address,uint256)",
    "0x095ea7b3": "approve(address,uint256)",
    "0x18160ddd": "totalSupply()",
    "0x06fdde03": "name()",
    "0x95d89b41": "symbol()",
    "0x313ce567": "decimals()",
    "0x8da5cb5b": "owner()",
    "0x715018a6": "renounceOwnership()",
    "0xf2fde38b": "transferOwnership(address)",
    "0x8456cb59": "pause()",
    "0x3f4ba83a": "unpause()",
    "0x5c975abb": "paused()",
}


class FeatureExtractor:
    """Extracts ML features from smart contract bytecode."""

    def __init__(self) -> None:
        self.logger = logger.bind(component="feature_extractor")

    def extract(self, bytecode_hex: str) -> ContractFeatures:
        """
        Extract features from hex-encoded bytecode.
        
        Args:
            bytecode_hex: Hex string of contract bytecode (with or without 0x prefix)
            
        Returns:
            ContractFeatures with extracted features
        """
        # Normalize bytecode
        bytecode_hex = bytecode_hex.lower()
        if bytecode_hex.startswith("0x"):
            bytecode_hex = bytecode_hex[2:]

        # Convert to bytes
        try:
            bytecode = bytes.fromhex(bytecode_hex)
        except ValueError as e:
            self.logger.error("Invalid bytecode hex", error=str(e))
            raise ValueError(f"Invalid bytecode hex: {e}") from e

        # Extract opcode distribution
        opcode_dist, opcodes_list = self._parse_opcodes(bytecode)

        # Calculate features
        features = ContractFeatures(
            bytecode_length=len(bytecode),
            opcode_distribution=opcode_dist,
            unique_opcodes=len(opcode_dist),
            jump_count=opcode_dist.get("JUMP", 0) + opcode_dist.get("JUMPI", 0),
            call_count=opcode_dist.get("CALL", 0),
            delegatecall_count=opcode_dist.get("DELEGATECALL", 0),
            staticcall_count=opcode_dist.get("STATICCALL", 0),
            selfdestruct_present=opcode_dist.get("SELFDESTRUCT", 0) > 0,
            sload_count=opcode_dist.get("SLOAD", 0),
            sstore_count=opcode_dist.get("SSTORE", 0),
            external_calls=self._count_external_calls(opcode_dist),
            ether_transfers=opcode_dist.get("CALL", 0),  # Simplified
            cyclomatic_complexity=self._estimate_complexity(opcode_dist),
            code_entropy=self._calculate_entropy(bytecode),
            function_selectors=self._extract_selectors(bytecode_hex),
            is_proxy=self._detect_proxy_pattern(opcode_dist, bytecode_hex),
            is_upgradeable=self._detect_upgradeable(bytecode_hex),
            has_ownership=self._detect_ownership(bytecode_hex),
            has_pausable=self._detect_pausable(bytecode_hex),
        )

        self.logger.info(
            "Features extracted",
            bytecode_length=features.bytecode_length,
            unique_opcodes=features.unique_opcodes,
            external_calls=features.external_calls,
        )

        return features

    def _parse_opcodes(self, bytecode: bytes) -> tuple[dict[str, int], list[str]]:
        """Parse bytecode into opcode distribution."""
        counter: Counter[str] = Counter()
        opcodes_list: list[str] = []
        i = 0

        while i < len(bytecode):
            op = bytecode[i]
            
            # Get opcode name
            if 0x60 <= op <= 0x7F:
                # PUSH1-PUSH32
                push_size = op - 0x5F
                name = f"PUSH{push_size}"
                i += push_size  # Skip push data
            elif 0x80 <= op <= 0x8F:
                # DUP1-DUP16
                name = f"DUP{op - 0x7F}"
            elif 0x90 <= op <= 0x9F:
                # SWAP1-SWAP16
                name = f"SWAP{op - 0x8F}"
            else:
                name = OPCODES.get(op, f"UNKNOWN_{hex(op)}")

            counter[name] += 1
            opcodes_list.append(name)
            i += 1

        return dict(counter), opcodes_list

    def _count_external_calls(self, opcode_dist: dict[str, int]) -> int:
        """Count total external calls."""
        return (
            opcode_dist.get("CALL", 0)
            + opcode_dist.get("DELEGATECALL", 0)
            + opcode_dist.get("STATICCALL", 0)
            + opcode_dist.get("CALLCODE", 0)
        )

    def _estimate_complexity(self, opcode_dist: dict[str, int]) -> float:
        """Estimate cyclomatic complexity from opcodes."""
        # Simplified: count decision points
        branches = opcode_dist.get("JUMPI", 0)
        return float(branches + 1)

    def _calculate_entropy(self, data: bytes) -> float:
        """Calculate Shannon entropy of bytecode."""
        if not data:
            return 0.0

        counter = Counter(data)
        length = len(data)
        entropy = 0.0

        for count in counter.values():
            if count > 0:
                prob = count / length
                entropy -= prob * log2(prob)

        return entropy

    def _extract_selectors(self, bytecode_hex: str) -> list[str]:
        """Extract function selectors from bytecode."""
        # Look for PUSH4 followed by 4 bytes (common selector pattern)
        selectors: list[str] = []
        pattern = r"63([0-9a-f]{8})"  # PUSH4 (0x63) + 4 bytes
        
        matches = re.findall(pattern, bytecode_hex)
        for match in matches:
            selector = f"0x{match}"
            selectors.append(selector)

        return list(set(selectors))  # Deduplicate

    def _detect_proxy_pattern(self, opcode_dist: dict[str, int], bytecode_hex: str) -> bool:
        """Detect if contract is a proxy."""
        # High DELEGATECALL usage suggests proxy
        if opcode_dist.get("DELEGATECALL", 0) > 0:
            # Check for minimal proxy pattern (EIP-1167)
            if "363d3d373d3d3d363d73" in bytecode_hex:
                return True
            # Check for OpenZeppelin proxy patterns
            if opcode_dist.get("DELEGATECALL", 0) >= 1 and len(bytecode_hex) < 1000:
                return True
        return False

    def _detect_upgradeable(self, bytecode_hex: str) -> bool:
        """Detect upgradeable proxy patterns."""
        # Look for implementation slot (EIP-1967)
        impl_slot = "360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"
        admin_slot = "b53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103"
        return impl_slot in bytecode_hex or admin_slot in bytecode_hex

    def _detect_ownership(self, bytecode_hex: str) -> bool:
        """Detect Ownable pattern."""
        # Look for owner() selector
        return "8da5cb5b" in bytecode_hex

    def _detect_pausable(self, bytecode_hex: str) -> bool:
        """Detect Pausable pattern."""
        # Look for paused() selector
        return "5c975abb" in bytecode_hex
