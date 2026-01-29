// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/**
 * @title RedTeamDAO
 * @notice DAO for security researchers and red teamers
 * @dev Manages proposals, voting, exploit tracking, and reputation scoring
 */

interface IProofOfExploit {
    function submitExploit(
        address target,
        string calldata description,
        bytes calldata proofBytes
    ) external returns (bytes32);
}

contract RedTeamDAO {
    /// @notice Team member role
    struct Member {
        address member;
        uint256 reputation;
        uint256 exploitsFound;
        uint256 joinedAt;
        bool isActive;
    }

    /// @notice Proposal structure
    struct Proposal {
        uint256 id;
        address proposer;
        string title;
        string description;
        uint256 votesFor;
        uint256 votesAgainst;
        uint256 createdAt;
        uint256 deadline;
        bool executed;
        mapping(address => bool) voted;
    }

    /// @notice Exploit tracking
    struct ExploitRecord {
        bytes32 id;
        address discoverer;
        address targetContract;
        string description;
        uint256 severity; // 1-5
        uint256 bountyAmount;
        uint256 discoveredAt;
        bool verified;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // STATE VARIABLES
    // ═══════════════════════════════════════════════════════════════════════════

    address public owner;
    address public proofOfExploitContract;

    mapping(address => Member) public members;
    mapping(uint256 => Proposal) public proposals;
    mapping(bytes32 => ExploitRecord) public exploits;

    address[] public memberList;
    uint256[] public proposalIds;
    bytes32[] public exploitIds;

    uint256 public nextProposalId = 1;
    uint256 public memberCount = 0;

    // Configuration
    uint256 public constant MIN_REPUTATION = 100;
    uint256 public constant PROPOSAL_DURATION = 7 days;
    uint256 public constant MIN_VOTERS = 5;

    // ═══════════════════════════════════════════════════════════════════════════
    // EVENTS
    // ═══════════════════════════════════════════════════════════════════════════

    event MemberJoined(address indexed member, uint256 reputation);
    event ProposalCreated(uint256 indexed proposalId, address indexed proposer, string title);
    event VoteCasted(uint256 indexed proposalId, address indexed voter, bool support);
    event ProposalExecuted(uint256 indexed proposalId);
    event ExploitRecorded(bytes32 indexed exploitId, address indexed discoverer, uint256 severity);
    event ReputationUpdated(address indexed member, uint256 newReputation);
    event BountyPaid(address indexed member, uint256 amount);

    // ═══════════════════════════════════════════════════════════════════════════
    // CONSTRUCTOR
    // ═══════════════════════════════════════════════════════════════════════════

    constructor() {
        owner = msg.sender;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // MEMBERSHIP
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Join the DAO (reputation starts at 0)
     */
    function joinDAO() external {
        require(!members[msg.sender].isActive, "Already a member");

        members[msg.sender] = Member({
            member: msg.sender,
            reputation: 0,
            exploitsFound: 0,
            joinedAt: block.timestamp,
            isActive: true
        });

        memberList.push(msg.sender);
        memberCount++;

        emit MemberJoined(msg.sender, 0);
    }

    /**
     * @notice Get member reputation
     */
    function getReputation(address member) external view returns (uint256) {
        return members[member].reputation;
    }

    /**
     * @notice Increase member reputation (only owner)
     */
    function addReputation(address member, uint256 amount) external onlyOwner {
        require(members[member].isActive, "Member not active");
        members[member].reputation += amount;
        emit ReputationUpdated(member, members[member].reputation);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // PROPOSALS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Create a new proposal
     */
    function createProposal(
        string calldata title,
        string calldata description
    ) external onlyMember returns (uint256) {
        require(members[msg.sender].reputation >= MIN_REPUTATION, "Insufficient reputation");

        uint256 proposalId = nextProposalId++;

        Proposal storage prop = proposals[proposalId];
        prop.id = proposalId;
        prop.proposer = msg.sender;
        prop.title = title;
        prop.description = description;
        prop.createdAt = block.timestamp;
        prop.deadline = block.timestamp + PROPOSAL_DURATION;
        prop.executed = false;

        proposalIds.push(proposalId);

        emit ProposalCreated(proposalId, msg.sender, title);
        return proposalId;
    }

    /**
     * @notice Vote on a proposal
     */
    function vote(uint256 proposalId, bool support) external onlyMember {
        Proposal storage prop = proposals[proposalId];
        require(block.timestamp < prop.deadline, "Proposal expired");
        require(!prop.voted[msg.sender], "Already voted");

        prop.voted[msg.sender] = true;

        if (support) {
            prop.votesFor++;
        } else {
            prop.votesAgainst++;
        }

        emit VoteCasted(proposalId, msg.sender, support);
    }

    /**
     * @notice Get proposal details
     */
    function getProposal(uint256 proposalId) external view returns (
        address proposer,
        string memory title,
        uint256 votesFor,
        uint256 votesAgainst,
        bool executed
    ) {
        Proposal storage prop = proposals[proposalId];
        return (prop.proposer, prop.title, prop.votesFor, prop.votesAgainst, prop.executed);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // EXPLOIT TRACKING
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Record a discovered exploit
     */
    function recordExploit(
        address targetContract,
        string calldata description,
        uint256 severity,
        uint256 bountyAmount
    ) external onlyMember returns (bytes32) {
        require(severity >= 1 && severity <= 5, "Invalid severity");

        bytes32 exploitId = keccak256(
            abi.encodePacked(msg.sender, targetContract, block.timestamp)
        );

        exploits[exploitId] = ExploitRecord({
            id: exploitId,
            discoverer: msg.sender,
            targetContract: targetContract,
            description: description,
            severity: severity,
            bountyAmount: bountyAmount,
            discoveredAt: block.timestamp,
            verified: false
        });

        exploitIds.push(exploitId);
        members[msg.sender].exploitsFound++;

        // Award reputation based on severity
        uint256 reputationReward = severity * 50;
        members[msg.sender].reputation += reputationReward;

        emit ExploitRecorded(exploitId, msg.sender, severity);
        emit ReputationUpdated(msg.sender, members[msg.sender].reputation);

        return exploitId;
    }

    /**
     * @notice Get exploit details
     */
    function getExploit(bytes32 exploitId) external view returns (
        address discoverer,
        address targetContract,
        string memory description,
        uint256 severity,
        uint256 bountyAmount,
        bool verified
    ) {
        ExploitRecord storage exploit = exploits[exploitId];
        return (
            exploit.discoverer,
            exploit.targetContract,
            exploit.description,
            exploit.severity,
            exploit.bountyAmount,
            exploit.verified
        );
    }

    /**
     * @notice Verify an exploit (only owner)
     */
    function verifyExploit(bytes32 exploitId) external onlyOwner {
        require(!exploits[exploitId].verified, "Already verified");
        exploits[exploitId].verified = true;
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // STATISTICS
    // ═══════════════════════════════════════════════════════════════════════════

    /**
     * @notice Get leaderboard (top 10 members by reputation)
     */
    function getLeaderboard() external view returns (
        address[] memory topMembers,
        uint256[] memory topReputation
    ) {
        uint256 len = memberList.length > 10 ? 10 : memberList.length;
        topMembers = new address[](len);
        topReputation = new uint256[](len);

        // Simple bubble sort for demo (in production use off-chain sorting)
        for (uint256 i = 0; i < len; i++) {
            uint256 maxRep = 0;
            address maxMember = address(0);

            for (uint256 j = 0; j < memberList.length; j++) {
                if (members[memberList[j]].reputation > maxRep) {
                    bool alreadyInTop = false;
                    for (uint256 k = 0; k < i; k++) {
                        if (topMembers[k] == memberList[j]) {
                            alreadyInTop = true;
                            break;
                        }
                    }
                    if (!alreadyInTop) {
                        maxRep = members[memberList[j]].reputation;
                        maxMember = memberList[j];
                    }
                }
            }

            if (maxMember != address(0)) {
                topMembers[i] = maxMember;
                topReputation[i] = maxRep;
            }
        }

        return (topMembers, topReputation);
    }

    /**
     * @notice Get member stats
     */
    function getMemberStats(address member) external view returns (
        uint256 reputation,
        uint256 exploitsFound,
        uint256 joinedAt
    ) {
        require(members[member].isActive, "Member not active");
        Member storage m = members[member];
        return (m.reputation, m.exploitsFound, m.joinedAt);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // MODIFIERS
    // ═══════════════════════════════════════════════════════════════════════════

    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }

    modifier onlyMember() {
        require(members[msg.sender].isActive, "Not a member");
        _;
    }
}
