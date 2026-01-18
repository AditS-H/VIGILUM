// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/**
 * @title IVigilumRegistry
 * @notice Interface for the VIGILUM security registry
 */
interface IVigilumRegistry {
    // ═══════════════════════════════════════════════════════════════════════════
    // EVENTS
    // ═══════════════════════════════════════════════════════════════════════════

    event ContractRegistered(address indexed contractAddr, bytes32 bytecodeHash, uint256 timestamp);
    event RiskScoreUpdated(address indexed contractAddr, uint256 oldScore, uint256 newScore);
    event ContractBlacklisted(address indexed contractAddr, string reason);

    // ═══════════════════════════════════════════════════════════════════════════
    // FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    function registerContract(address contractAddr, bytes32 bytecodeHash, uint256 initialRiskScore) external;
    function updateRiskScore(address contractAddr, uint256 newScore, uint32 vulnCount) external;
    function blacklistContract(address contractAddr, string calldata reason) external;
    function getRiskScore(address contractAddr) external view returns (uint256);
    function isBlacklisted(address contractAddr) external view returns (bool);
}
