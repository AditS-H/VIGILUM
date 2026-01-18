// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title IThreatOracle
 * @notice Interface for the ThreatOracle contract.
 */
interface IThreatOracle {
    /// @notice Emitted when a target's risk score is updated.
    event RiskUpdated(address indexed target, uint8 riskScore, uint256 timestamp);
    
    /// @notice Emitted when an oracle reporter is added or removed.
    event OracleReporterUpdated(address indexed reporter, bool isActive);

    /// @notice Get the risk score for a target address.
    function getRiskScore(address target) external view returns (uint8);
    
    /// @notice Get the timestamp of the last update for a target.
    function getLastUpdate(address target) external view returns (uint256);
    
    /// @notice Check if an address is a registered oracle reporter.
    function isOracleReporter(address account) external view returns (bool);
}
