// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {IVigilumRegistry} from "./interfaces/IVigilumRegistry.sol";

/**
 * @title VigilumRegistry
 * @notice Central registry for contract security metadata, risk scores, and threat intelligence
 * @dev Stores on-chain security assessments and integrates with oracle feeds
 */
contract VigilumRegistry is IVigilumRegistry {
    // ═══════════════════════════════════════════════════════════════════════════
    // TYPES
    // ═══════════════════════════════════════════════════════════════════════════

    struct ContractMetadata {
        uint256 riskScore;          // 0-10000 (basis points)
        ThreatLevel threatLevel;
        uint64 lastScanTimestamp;
        uint32 vulnCount;
        bytes32 bytecodeHash;
        bool isVerified;
        bool isBlacklisted;
    }

    enum ThreatLevel {
        NONE,
        INFO,
        LOW,
        MEDIUM,
        HIGH,
        CRITICAL
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // STATE
    // ═══════════════════════════════════════════════════════════════════════════

    /// @notice Mapping from contract address to security metadata
    mapping(address => ContractMetadata) public contractData;

    /// @notice Mapping from bytecode hash to known addresses
    mapping(bytes32 => address[]) public bytecodeToContracts;

    /// @notice Authorized security oracles
    mapping(address => bool) public authorizedOracles;

    /// @notice Contract owner
    address public owner;

    /// @notice Proposed new owner
    address public pendingOwner;

    // ═══════════════════════════════════════════════════════════════════════════
    // EVENTS (inherits ContractRegistered, RiskScoreUpdated, ContractBlacklisted from interface)
    // ═══════════════════════════════════════════════════════════════════════════

    event ThreatLevelChanged(address indexed contractAddr, ThreatLevel oldLevel, ThreatLevel newLevel);
    event OracleAuthorized(address indexed oracle);
    event OracleRevoked(address indexed oracle);
    event OwnershipTransferStarted(address indexed currentOwner, address indexed pendingOwner);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    // ═══════════════════════════════════════════════════════════════════════════
    // ERRORS
    // ═══════════════════════════════════════════════════════════════════════════

    error Unauthorized();
    error InvalidAddress();
    error InvalidRiskScore();
    error ContractAlreadyRegistered();
    error ContractNotRegistered();
    error NoPendingOwner();

    // ═══════════════════════════════════════════════════════════════════════════
    // MODIFIERS
    // ═══════════════════════════════════════════════════════════════════════════

    modifier onlyOwner() {
        if (msg.sender != owner) revert Unauthorized();
        _;
    }

    modifier onlyOracle() {
        if (!authorizedOracles[msg.sender]) revert Unauthorized();
        _;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // CONSTRUCTOR
    // ═══════════════════════════════════════════════════════════════════════════

    constructor() {
        owner = msg.sender;
        authorizedOracles[msg.sender] = true;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // EXTERNAL FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Register a new contract with initial security metadata
     * @param contractAddr Address of the contract to register
     * @param bytecodeHash Hash of the contract's bytecode
     * @param initialRiskScore Initial risk score (0-10000)
     */
    function registerContract(
        address contractAddr,
        bytes32 bytecodeHash,
        uint256 initialRiskScore
    ) external onlyOracle {
        if (contractAddr == address(0)) revert InvalidAddress();
        if (initialRiskScore > 10000) revert InvalidRiskScore();
        if (contractData[contractAddr].lastScanTimestamp != 0) revert ContractAlreadyRegistered();

        contractData[contractAddr] = ContractMetadata({
            riskScore: initialRiskScore,
            threatLevel: _scoreToThreatLevel(initialRiskScore),
            lastScanTimestamp: uint64(block.timestamp),
            vulnCount: 0,
            bytecodeHash: bytecodeHash,
            isVerified: false,
            isBlacklisted: false
        });

        bytecodeToContracts[bytecodeHash].push(contractAddr);

        emit ContractRegistered(contractAddr, bytecodeHash, block.timestamp);
    }

    /**
     * @notice Update the risk score for a registered contract
     * @param contractAddr Address of the contract
     * @param newScore New risk score (0-10000)
     * @param vulnCount Number of vulnerabilities found
     */
    function updateRiskScore(
        address contractAddr,
        uint256 newScore,
        uint32 vulnCount
    ) external onlyOracle {
        if (newScore > 10000) revert InvalidRiskScore();
        
        ContractMetadata storage data = contractData[contractAddr];
        if (data.lastScanTimestamp == 0) revert ContractNotRegistered();

        uint256 oldScore = data.riskScore;
        ThreatLevel oldLevel = data.threatLevel;
        ThreatLevel newLevel = _scoreToThreatLevel(newScore);

        data.riskScore = newScore;
        data.threatLevel = newLevel;
        data.lastScanTimestamp = uint64(block.timestamp);
        data.vulnCount = vulnCount;

        emit RiskScoreUpdated(contractAddr, oldScore, newScore);
        
        if (oldLevel != newLevel) {
            emit ThreatLevelChanged(contractAddr, oldLevel, newLevel);
        }
    }

    /**
     * @notice Blacklist a malicious contract
     * @param contractAddr Address to blacklist
     * @param reason Reason for blacklisting
     */
    function blacklistContract(address contractAddr, string calldata reason) external onlyOracle {
        contractData[contractAddr].isBlacklisted = true;
        contractData[contractAddr].threatLevel = ThreatLevel.CRITICAL;
        emit ContractBlacklisted(contractAddr, reason);
    }

    /**
     * @notice Mark a contract as verified
     * @param contractAddr Address of the verified contract
     */
    function setVerified(address contractAddr, bool verified) external onlyOracle {
        contractData[contractAddr].isVerified = verified;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // VIEW FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Get the risk score for a contract
     * @param contractAddr Address to query
     * @return Risk score (0-10000)
     */
    function getRiskScore(address contractAddr) external view returns (uint256) {
        return contractData[contractAddr].riskScore;
    }

    /**
     * @notice Get the threat level for a contract
     * @param contractAddr Address to query
     * @return Threat level enum
     */
    function getThreatLevel(address contractAddr) external view returns (ThreatLevel) {
        return contractData[contractAddr].threatLevel;
    }

    /**
     * @notice Check if a contract is blacklisted
     * @param contractAddr Address to check
     * @return True if blacklisted
     */
    function isBlacklisted(address contractAddr) external view returns (bool) {
        return contractData[contractAddr].isBlacklisted;
    }

    /**
     * @notice Get all contracts with matching bytecode
     * @param bytecodeHash Hash to look up
     * @return Array of contract addresses
     */
    function getContractsByBytecode(bytes32 bytecodeHash) external view returns (address[] memory) {
        return bytecodeToContracts[bytecodeHash];
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // ADMIN FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Authorize a new security oracle
     * @param oracle Address to authorize
     */
    function authorizeOracle(address oracle) external onlyOwner {
        if (oracle == address(0)) revert InvalidAddress();
        authorizedOracles[oracle] = true;
        emit OracleAuthorized(oracle);
    }

    /**
     * @notice Revoke oracle authorization
     * @param oracle Address to revoke
     */
    function revokeOracle(address oracle) external onlyOwner {
        authorizedOracles[oracle] = false;
        emit OracleRevoked(oracle);
    }

    /**
     * @notice Start ownership transfer (2-step)
     * @param newOwner Address of new owner
     */
    function transferOwnership(address newOwner) external onlyOwner {
        if (newOwner == address(0)) revert InvalidAddress();
        pendingOwner = newOwner;
        emit OwnershipTransferStarted(owner, newOwner);
    }

    /**
     * @notice Accept ownership transfer
     */
    function acceptOwnership() external {
        if (msg.sender != pendingOwner) revert NoPendingOwner();
        emit OwnershipTransferred(owner, pendingOwner);
        owner = pendingOwner;
        pendingOwner = address(0);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // INTERNAL FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @dev Convert risk score to threat level
     */
    function _scoreToThreatLevel(uint256 score) internal pure returns (ThreatLevel) {
        if (score >= 8000) return ThreatLevel.CRITICAL;
        if (score >= 6000) return ThreatLevel.HIGH;
        if (score >= 4000) return ThreatLevel.MEDIUM;
        if (score >= 2000) return ThreatLevel.LOW;
        if (score > 0) return ThreatLevel.INFO;
        return ThreatLevel.NONE;
    }
}
