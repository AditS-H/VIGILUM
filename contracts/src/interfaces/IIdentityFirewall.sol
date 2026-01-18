// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/**
 * @title IIdentityFirewall
 * @notice Interface for the VIGILUM Identity Firewall
 * @dev Verifies zero-knowledge proofs of human-like behavior
 */
interface IIdentityFirewall {
    // ═══════════════════════════════════════════════════════════════════════════
    // EVENTS
    // ═══════════════════════════════════════════════════════════════════════════

    event ProofVerified(address indexed user, bytes32 indexed proofHash, uint256 timestamp);
    event ProofRevoked(address indexed user, bytes32 indexed proofHash, string reason);
    event ChallengeIssued(address indexed user, bytes32 challengeHash, uint256 expiresAt);

    // ═══════════════════════════════════════════════════════════════════════════
    // FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    function verifyHumanProof(bytes calldata proof) external returns (bool);
    function hasValidProof(address user) external view returns (bool);
    function getProofExpiry(bytes32 proofHash) external view returns (uint256);
    function revokeProof(bytes32 proofHash, string calldata reason) external;
}
