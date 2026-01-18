// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IThreatOracle} from "./interfaces/IThreatOracle.sol";

/**
 * @title ThreatOracle
 * @author VIGILUM Team
 * @notice On-chain oracle for publishing aggregated threat intelligence signals.
 * @dev Risk scores are 0-100 where:
 *      0-19  = INFO (no action needed)
 *      20-39 = LOW (monitor)
 *      40-59 = MEDIUM (review recommended)
 *      60-79 = HIGH (action recommended)
 *      80-100 = CRITICAL (immediate action required)
 */
contract ThreatOracle is IThreatOracle {
    // ═══════════════════════════════════════════════════════════════════════════
    // ERRORS
    // ═══════════════════════════════════════════════════════════════════════════
    
    /// @notice Caller is not an authorized oracle reporter.
    error NotOracleReporter();
    
    /// @notice Caller is not the owner.
    error NotOwner();
    
    /// @notice Invalid risk score (must be 0-100).
    error InvalidRiskScore();
    
    /// @notice Array lengths do not match.
    error ArrayLengthMismatch();
    
    /// @notice Batch size exceeds maximum.
    error BatchTooLarge();
    
    /// @notice Zero address not allowed.
    error ZeroAddress();

    // ═══════════════════════════════════════════════════════════════════════════
    // STATE
    // ═══════════════════════════════════════════════════════════════════════════
    
    /// @notice Maximum risk score value.
    uint8 public constant MAX_RISK_SCORE = 100;
    
    /// @notice Maximum batch size for updates.
    uint256 public constant MAX_BATCH_SIZE = 100;
    
    /// @notice Contract owner.
    address public owner;
    
    /// @notice Mapping of target address to risk score.
    mapping(address => uint8) public riskScores;
    
    /// @notice Mapping of target address to last update timestamp.
    mapping(address => uint256) public lastUpdates;
    
    /// @notice Mapping of authorized oracle reporters.
    mapping(address => bool) public oracleReporters;
    
    /// @notice Total number of targets tracked.
    uint256 public totalTargets;
    
    /// @notice Total number of updates made.
    uint256 public totalUpdates;

    // ═══════════════════════════════════════════════════════════════════════════
    // MODIFIERS
    // ═══════════════════════════════════════════════════════════════════════════
    
    modifier onlyOwner() {
        if (msg.sender != owner) revert NotOwner();
        _;
    }
    
    modifier onlyOracleReporter() {
        if (!oracleReporters[msg.sender]) revert NotOracleReporter();
        _;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // CONSTRUCTOR
    // ═══════════════════════════════════════════════════════════════════════════
    
    /**
     * @notice Initialize the ThreatOracle.
     * @param initialReporters Array of initial oracle reporter addresses.
     */
    constructor(address[] memory initialReporters) {
        owner = msg.sender;
        
        // Add owner as reporter by default
        oracleReporters[msg.sender] = true;
        emit OracleReporterUpdated(msg.sender, true);
        
        // Add initial reporters
        for (uint256 i = 0; i < initialReporters.length; i++) {
            if (initialReporters[i] != address(0)) {
                oracleReporters[initialReporters[i]] = true;
                emit OracleReporterUpdated(initialReporters[i], true);
            }
        }
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // ORACLE UPDATES
    // ═══════════════════════════════════════════════════════════════════════════
    
    /**
     * @notice Update the risk score for a single target.
     * @param target The target address to update.
     * @param score The new risk score (0-100).
     */
    function updateRiskScore(address target, uint8 score) external onlyOracleReporter {
        if (target == address(0)) revert ZeroAddress();
        if (score > MAX_RISK_SCORE) revert InvalidRiskScore();
        
        _updateScore(target, score);
    }
    
    /**
     * @notice Update risk scores for multiple targets in a single transaction.
     * @param targets Array of target addresses.
     * @param scores Array of risk scores (0-100).
     */
    function batchUpdateRiskScores(
        address[] calldata targets,
        uint8[] calldata scores
    ) external onlyOracleReporter {
        if (targets.length != scores.length) revert ArrayLengthMismatch();
        if (targets.length > MAX_BATCH_SIZE) revert BatchTooLarge();
        
        for (uint256 i = 0; i < targets.length; i++) {
            if (targets[i] == address(0)) continue; // Skip zero addresses
            if (scores[i] > MAX_RISK_SCORE) revert InvalidRiskScore();
            
            _updateScore(targets[i], scores[i]);
        }
    }
    
    /**
     * @dev Internal function to update a risk score.
     */
    function _updateScore(address target, uint8 score) internal {
        // Track new targets
        if (riskScores[target] == 0 && lastUpdates[target] == 0) {
            totalTargets++;
        }
        
        riskScores[target] = score;
        lastUpdates[target] = block.timestamp;
        totalUpdates++;
        
        emit RiskUpdated(target, score, block.timestamp);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // VIEW FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════
    
    /**
     * @notice Get the risk score for a target address.
     * @param target The target address to query.
     * @return The risk score (0-100).
     */
    function getRiskScore(address target) external view returns (uint8) {
        return riskScores[target];
    }
    
    /**
     * @notice Get the timestamp of the last update for a target.
     * @param target The target address to query.
     * @return The Unix timestamp of the last update.
     */
    function getLastUpdate(address target) external view returns (uint256) {
        return lastUpdates[target];
    }
    
    /**
     * @notice Check if an address is a registered oracle reporter.
     * @param account The address to check.
     * @return True if the address is an oracle reporter.
     */
    function isOracleReporter(address account) external view returns (bool) {
        return oracleReporters[account];
    }
    
    /**
     * @notice Get risk score and last update in a single call.
     * @param target The target address to query.
     * @return score The risk score (0-100).
     * @return timestamp The Unix timestamp of the last update.
     */
    function getTargetInfo(address target) external view returns (uint8 score, uint256 timestamp) {
        return (riskScores[target], lastUpdates[target]);
    }
    
    /**
     * @notice Categorize the risk level based on score.
     * @param target The target address to query.
     * @return level The risk level as a string.
     */
    function getRiskLevel(address target) external view returns (string memory level) {
        uint8 score = riskScores[target];
        
        if (score >= 80) return "CRITICAL";
        if (score >= 60) return "HIGH";
        if (score >= 40) return "MEDIUM";
        if (score >= 20) return "LOW";
        return "INFO";
    }
    
    /**
     * @notice Check if a target is considered high risk (score >= 60).
     * @param target The target address to check.
     * @return True if the target is high risk.
     */
    function isHighRisk(address target) external view returns (bool) {
        return riskScores[target] >= 60;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // ADMIN FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════
    
    /**
     * @notice Add or remove an oracle reporter.
     * @param reporter The reporter address.
     * @param isActive Whether the reporter should be active.
     */
    function setOracleReporter(address reporter, bool isActive) external onlyOwner {
        if (reporter == address(0)) revert ZeroAddress();
        
        oracleReporters[reporter] = isActive;
        emit OracleReporterUpdated(reporter, isActive);
    }
    
    /**
     * @notice Transfer ownership to a new address.
     * @param newOwner The new owner address.
     */
    function transferOwnership(address newOwner) external onlyOwner {
        if (newOwner == address(0)) revert ZeroAddress();
        owner = newOwner;
    }
    
    /**
     * @notice Clear the risk score for a target (reset to 0).
     * @param target The target address to clear.
     */
    function clearRiskScore(address target) external onlyOwner {
        if (riskScores[target] > 0) {
            riskScores[target] = 0;
            lastUpdates[target] = block.timestamp;
            emit RiskUpdated(target, 0, block.timestamp);
        }
    }
}
