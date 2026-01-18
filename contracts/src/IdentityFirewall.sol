// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {IIdentityFirewall} from "./interfaces/IIdentityFirewall.sol";

/**
 * @title IdentityFirewall
 * @notice Verifies zero-knowledge proofs of human-like behavior
 * @dev Gates access to protected protocols by requiring proof of human-likeness
 * 
 * The Identity Firewall accepts ZK proofs that demonstrate:
 * - Consistent transaction timing patterns
 * - Natural gas usage variance
 * - Realistic interaction cadence
 * 
 * Without revealing any identifying information about the user.
 */
contract IdentityFirewall is IIdentityFirewall {
    // ═══════════════════════════════════════════════════════════════════════════
    // TYPES
    // ═══════════════════════════════════════════════════════════════════════════

    struct ProofRecord {
        address user;
        uint64 verifiedAt;
        uint64 expiresAt;
        bool isRevoked;
    }

    struct UserStatus {
        bytes32 latestProofHash;
        uint32 totalProofs;
        uint64 firstVerifiedAt;
        uint64 lastVerifiedAt;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // CONSTANTS
    // ═══════════════════════════════════════════════════════════════════════════

    /// @notice Default proof validity duration (24 hours)
    uint256 public constant DEFAULT_PROOF_DURATION = 24 hours;

    /// @notice Minimum proof validity duration (1 hour)
    uint256 public constant MIN_PROOF_DURATION = 1 hours;

    /// @notice Maximum proof validity duration (7 days)
    uint256 public constant MAX_PROOF_DURATION = 7 days;

    // ═══════════════════════════════════════════════════════════════════════════
    // STATE
    // ═══════════════════════════════════════════════════════════════════════════

    /// @notice Mapping from proof hash to proof record
    mapping(bytes32 => ProofRecord) public proofRecords;

    /// @notice Mapping from user address to their status
    mapping(address => UserStatus) public userStatuses;

    /// @notice Mapping of authorized verifiers (can submit proofs on behalf of users)
    mapping(address => bool) public authorizedVerifiers;

    /// @notice Contract owner
    address public owner;

    /// @notice Pending owner for 2-step transfer
    address public pendingOwner;

    /// @notice Current proof validity duration
    uint256 public proofDuration;

    /// @notice Total number of verified proofs
    uint256 public totalVerifiedProofs;

    /// @notice Whether the contract is paused
    bool public paused;

    // ═══════════════════════════════════════════════════════════════════════════
    // EVENTS (additional to interface)
    // ═══════════════════════════════════════════════════════════════════════════

    event VerifierAuthorized(address indexed verifier);
    event VerifierRevoked(address indexed verifier);
    event ProofDurationUpdated(uint256 oldDuration, uint256 newDuration);
    event Paused(address indexed by);
    event Unpaused(address indexed by);
    event OwnershipTransferStarted(address indexed currentOwner, address indexed pendingOwner);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    // ═══════════════════════════════════════════════════════════════════════════
    // ERRORS
    // ═══════════════════════════════════════════════════════════════════════════

    error Unauthorized();
    error ContractPaused();
    error InvalidProof();
    error ProofAlreadyExists();
    error ProofNotFound();
    error ProofExpired();
    error ProofAlreadyRevoked();
    error ProofIsRevoked();
    error InvalidDuration();
    error NoPendingOwner();
    error ZeroAddress();

    // ═══════════════════════════════════════════════════════════════════════════
    // MODIFIERS
    // ═══════════════════════════════════════════════════════════════════════════

    modifier onlyOwner() {
        if (msg.sender != owner) revert Unauthorized();
        _;
    }

    modifier onlyVerifier() {
        if (!authorizedVerifiers[msg.sender] && msg.sender != owner) revert Unauthorized();
        _;
    }

    modifier whenNotPaused() {
        if (paused) revert ContractPaused();
        _;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // CONSTRUCTOR
    // ═══════════════════════════════════════════════════════════════════════════

    constructor() {
        owner = msg.sender;
        authorizedVerifiers[msg.sender] = true;
        proofDuration = DEFAULT_PROOF_DURATION;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // EXTERNAL FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Verify a human-likeness proof
     * @dev For MVP: accepts any proof >= 32 bytes
     *      TODO: Replace with actual ZK verifier (Noir circuit)
     * @param proof The ZK proof bytes
     * @return True if proof is valid
     */
    function verifyHumanProof(bytes calldata proof) external whenNotPaused returns (bool) {
        return _verifyProof(msg.sender, proof);
    }

    /**
     * @notice Verify a proof on behalf of a user (for meta-transactions)
     * @param user The user address
     * @param proof The ZK proof bytes
     * @return True if proof is valid
     */
    function verifyProofFor(address user, bytes calldata proof) external onlyVerifier whenNotPaused returns (bool) {
        if (user == address(0)) revert ZeroAddress();
        return _verifyProof(user, proof);
    }

    /**
     * @notice Revoke a previously verified proof
     * @param proofHash The hash of the proof to revoke
     * @param reason Reason for revocation
     */
    function revokeProof(bytes32 proofHash, string calldata reason) external onlyOwner {
        ProofRecord storage record = proofRecords[proofHash];
        if (record.verifiedAt == 0) revert ProofNotFound();
        if (record.isRevoked) revert ProofAlreadyRevoked();

        record.isRevoked = true;

        emit ProofRevoked(record.user, proofHash, reason);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // VIEW FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Check if a user has a valid (non-expired, non-revoked) proof
     * @param user The user address to check
     * @return True if user has valid proof
     */
    function hasValidProof(address user) external view returns (bool) {
        UserStatus storage status = userStatuses[user];
        if (status.latestProofHash == bytes32(0)) return false;

        ProofRecord storage record = proofRecords[status.latestProofHash];
        return !record.isRevoked && block.timestamp < record.expiresAt;
    }

    /**
     * @notice Get the expiry timestamp for a proof
     * @param proofHash The proof hash
     * @return Expiry timestamp (0 if not found)
     */
    function getProofExpiry(bytes32 proofHash) external view returns (uint256) {
        return proofRecords[proofHash].expiresAt;
    }

    /**
     * @notice Check if a specific proof is valid
     * @param proofHash The proof hash to check
     * @return True if proof is valid
     */
    function isProofValid(bytes32 proofHash) external view returns (bool) {
        ProofRecord storage record = proofRecords[proofHash];
        return record.verifiedAt > 0 && !record.isRevoked && block.timestamp < record.expiresAt;
    }

    /**
     * @notice Get user verification status
     * @param user The user address
     * @return status The user's verification status
     */
    function getUserStatus(address user) external view returns (UserStatus memory status) {
        return userStatuses[user];
    }

    /**
     * @notice Require that caller has a valid proof (for use in other contracts)
     * @dev Reverts if caller doesn't have valid proof
     */
    function requireHumanProof() external view {
        UserStatus storage status = userStatuses[msg.sender];
        if (status.latestProofHash == bytes32(0)) revert InvalidProof();

        ProofRecord storage record = proofRecords[status.latestProofHash];
        if (record.isRevoked) revert ProofIsRevoked();
        if (block.timestamp >= record.expiresAt) revert ProofExpired();
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // ADMIN FUNCTIONS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Authorize a new verifier
     * @param verifier Address to authorize
     */
    function authorizeVerifier(address verifier) external onlyOwner {
        if (verifier == address(0)) revert ZeroAddress();
        authorizedVerifiers[verifier] = true;
        emit VerifierAuthorized(verifier);
    }

    /**
     * @notice Revoke verifier authorization
     * @param verifier Address to revoke
     */
    function revokeVerifier(address verifier) external onlyOwner {
        authorizedVerifiers[verifier] = false;
        emit VerifierRevoked(verifier);
    }

    /**
     * @notice Update proof validity duration
     * @param newDuration New duration in seconds
     */
    function setProofDuration(uint256 newDuration) external onlyOwner {
        if (newDuration < MIN_PROOF_DURATION || newDuration > MAX_PROOF_DURATION) {
            revert InvalidDuration();
        }
        uint256 oldDuration = proofDuration;
        proofDuration = newDuration;
        emit ProofDurationUpdated(oldDuration, newDuration);
    }

    /**
     * @notice Pause the contract
     */
    function pause() external onlyOwner {
        paused = true;
        emit Paused(msg.sender);
    }

    /**
     * @notice Unpause the contract
     */
    function unpause() external onlyOwner {
        paused = false;
        emit Unpaused(msg.sender);
    }

    /**
     * @notice Start ownership transfer (2-step)
     * @param newOwner Address of new owner
     */
    function transferOwnership(address newOwner) external onlyOwner {
        if (newOwner == address(0)) revert ZeroAddress();
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
     * @dev Internal proof verification logic
     * @param user The user address
     * @param proof The proof bytes
     * @return True if valid
     */
    function _verifyProof(address user, bytes calldata proof) internal returns (bool) {
        // ============================================================
        // STUB VERIFICATION (MVP)
        // For MVP, accept any proof that is at least 32 bytes
        // 
        // TODO: Replace with actual ZK verification:
        // 1. Parse proof into Noir proof structure
        // 2. Extract public inputs
        // 3. Call verifier contract generated by Noir
        // ============================================================
        if (proof.length < 32) revert InvalidProof();

        // Compute proof hash
        bytes32 proofHash = keccak256(proof);

        // Check if proof already exists
        if (proofRecords[proofHash].verifiedAt > 0) revert ProofAlreadyExists();

        // Calculate expiry
        uint64 expiresAt = uint64(block.timestamp + proofDuration);

        // Store proof record
        proofRecords[proofHash] = ProofRecord({
            user: user,
            verifiedAt: uint64(block.timestamp),
            expiresAt: expiresAt,
            isRevoked: false
        });

        // Update user status
        UserStatus storage status = userStatuses[user];
        status.latestProofHash = proofHash;
        status.totalProofs++;
        status.lastVerifiedAt = uint64(block.timestamp);
        if (status.firstVerifiedAt == 0) {
            status.firstVerifiedAt = uint64(block.timestamp);
        }

        // Increment total
        totalVerifiedProofs++;

        emit ProofVerified(user, proofHash, block.timestamp);

        return true;
    }
}
