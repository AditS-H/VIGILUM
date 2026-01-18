// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Test, console2} from "forge-std/Test.sol";
import {IdentityFirewall} from "../src/IdentityFirewall.sol";
import {IIdentityFirewall} from "../src/interfaces/IIdentityFirewall.sol";

contract IdentityFirewallTest is Test {
    IdentityFirewall public firewall;

    address public owner = address(this);
    address public verifier = address(0x1);
    address public user = address(0x2);
    address public user2 = address(0x3);

    bytes public validProof = abi.encodePacked(
        bytes32(keccak256("test_proof")),
        bytes32(keccak256("public_input"))
    );
    bytes public shortProof = abi.encodePacked(bytes16(0));

    function setUp() public {
        firewall = new IdentityFirewall();
        firewall.authorizeVerifier(verifier);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // CONSTRUCTOR TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_Constructor() public view {
        assertEq(firewall.owner(), owner);
        assertTrue(firewall.authorizedVerifiers(owner));
        assertEq(firewall.proofDuration(), 24 hours);
        assertFalse(firewall.paused());
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // VERIFY PROOF TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_VerifyHumanProof() public {
        vm.prank(user);
        bool result = firewall.verifyHumanProof(validProof);

        assertTrue(result);
        assertTrue(firewall.hasValidProof(user));
        assertEq(firewall.totalVerifiedProofs(), 1);
    }

    function test_VerifyHumanProof_EmitsEvent() public {
        bytes32 expectedProofHash = keccak256(validProof);

        vm.expectEmit(true, true, false, true);
        emit IIdentityFirewall.ProofVerified(user, expectedProofHash, block.timestamp);

        vm.prank(user);
        firewall.verifyHumanProof(validProof);
    }

    function test_VerifyHumanProof_RevertInvalidProof() public {
        vm.prank(user);
        vm.expectRevert(IdentityFirewall.InvalidProof.selector);
        firewall.verifyHumanProof(shortProof);
    }

    function test_VerifyHumanProof_RevertProofAlreadyExists() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        vm.prank(user2);
        vm.expectRevert(IdentityFirewall.ProofAlreadyExists.selector);
        firewall.verifyHumanProof(validProof);
    }

    function test_VerifyProofFor() public {
        vm.prank(verifier);
        bool result = firewall.verifyProofFor(user, validProof);

        assertTrue(result);
        assertTrue(firewall.hasValidProof(user));
    }

    function test_VerifyProofFor_RevertUnauthorized() public {
        vm.prank(user);
        vm.expectRevert(IdentityFirewall.Unauthorized.selector);
        firewall.verifyProofFor(user2, validProof);
    }

    function test_VerifyProofFor_RevertZeroAddress() public {
        vm.prank(verifier);
        vm.expectRevert(IdentityFirewall.ZeroAddress.selector);
        firewall.verifyProofFor(address(0), validProof);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // PROOF VALIDITY TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_HasValidProof_ReturnsFalseWhenNoProof() public view {
        assertFalse(firewall.hasValidProof(user));
    }

    function test_HasValidProof_ReturnsFalseWhenExpired() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        // Fast forward past expiry
        vm.warp(block.timestamp + 25 hours);

        assertFalse(firewall.hasValidProof(user));
    }

    function test_HasValidProof_ReturnsFalseWhenRevoked() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        bytes32 proofHash = keccak256(validProof);
        firewall.revokeProof(proofHash, "test revocation");

        assertFalse(firewall.hasValidProof(user));
    }

    function test_GetProofExpiry() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        bytes32 proofHash = keccak256(validProof);
        uint256 expiry = firewall.getProofExpiry(proofHash);

        assertEq(expiry, block.timestamp + 24 hours);
    }

    function test_IsProofValid() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        bytes32 proofHash = keccak256(validProof);
        assertTrue(firewall.isProofValid(proofHash));
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // REVOKE PROOF TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_RevokeProof() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        bytes32 proofHash = keccak256(validProof);
        firewall.revokeProof(proofHash, "malicious activity detected");

        assertFalse(firewall.isProofValid(proofHash));
    }

    function test_RevokeProof_EmitsEvent() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        bytes32 proofHash = keccak256(validProof);

        vm.expectEmit(true, true, false, true);
        emit IIdentityFirewall.ProofRevoked(user, proofHash, "test");

        firewall.revokeProof(proofHash, "test");
    }

    function test_RevokeProof_RevertUnauthorized() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        bytes32 proofHash = keccak256(validProof);

        vm.prank(user);
        vm.expectRevert(IdentityFirewall.Unauthorized.selector);
        firewall.revokeProof(proofHash, "attempt");
    }

    function test_RevokeProof_RevertNotFound() public {
        vm.expectRevert(IdentityFirewall.ProofNotFound.selector);
        firewall.revokeProof(bytes32(0), "not found");
    }

    function test_RevokeProof_RevertAlreadyRevoked() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        bytes32 proofHash = keccak256(validProof);
        firewall.revokeProof(proofHash, "first");

        vm.expectRevert(IdentityFirewall.ProofAlreadyRevoked.selector);
        firewall.revokeProof(proofHash, "second");
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // REQUIRE HUMAN PROOF TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_RequireHumanProof() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        vm.prank(user);
        firewall.requireHumanProof(); // Should not revert
    }

    function test_RequireHumanProof_RevertNoProof() public {
        vm.prank(user);
        vm.expectRevert(IdentityFirewall.InvalidProof.selector);
        firewall.requireHumanProof();
    }

    function test_RequireHumanProof_RevertExpired() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        vm.warp(block.timestamp + 25 hours);

        vm.prank(user);
        vm.expectRevert(IdentityFirewall.ProofExpired.selector);
        firewall.requireHumanProof();
    }

    function test_RequireHumanProof_RevertRevoked() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        bytes32 proofHash = keccak256(validProof);
        firewall.revokeProof(proofHash, "revoked");

        vm.prank(user);
        vm.expectRevert(IdentityFirewall.ProofIsRevoked.selector);
        firewall.requireHumanProof();
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // USER STATUS TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_GetUserStatus() public {
        vm.prank(user);
        firewall.verifyHumanProof(validProof);

        IdentityFirewall.UserStatus memory status = firewall.getUserStatus(user);

        assertEq(status.latestProofHash, keccak256(validProof));
        assertEq(status.totalProofs, 1);
        assertEq(status.firstVerifiedAt, block.timestamp);
        assertEq(status.lastVerifiedAt, block.timestamp);
    }

    function test_GetUserStatus_MultipleProofs() public {
        bytes memory proof1 = abi.encodePacked(bytes32(keccak256("proof1")), bytes32(0));
        bytes memory proof2 = abi.encodePacked(bytes32(keccak256("proof2")), bytes32(0));

        uint256 firstTime = block.timestamp;

        vm.prank(user);
        firewall.verifyHumanProof(proof1);

        vm.warp(block.timestamp + 1 hours);

        vm.prank(user);
        firewall.verifyHumanProof(proof2);

        IdentityFirewall.UserStatus memory status = firewall.getUserStatus(user);

        assertEq(status.totalProofs, 2);
        assertEq(status.firstVerifiedAt, firstTime);
        assertEq(status.lastVerifiedAt, block.timestamp);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // ADMIN TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_AuthorizeVerifier() public {
        address newVerifier = address(0x123);
        firewall.authorizeVerifier(newVerifier);
        assertTrue(firewall.authorizedVerifiers(newVerifier));
    }

    function test_RevokeVerifier() public {
        firewall.revokeVerifier(verifier);
        assertFalse(firewall.authorizedVerifiers(verifier));
    }

    function test_SetProofDuration() public {
        uint256 newDuration = 12 hours;
        firewall.setProofDuration(newDuration);
        assertEq(firewall.proofDuration(), newDuration);
    }

    function test_SetProofDuration_RevertTooShort() public {
        vm.expectRevert(IdentityFirewall.InvalidDuration.selector);
        firewall.setProofDuration(30 minutes);
    }

    function test_SetProofDuration_RevertTooLong() public {
        vm.expectRevert(IdentityFirewall.InvalidDuration.selector);
        firewall.setProofDuration(8 days);
    }

    function test_Pause() public {
        firewall.pause();
        assertTrue(firewall.paused());
    }

    function test_Unpause() public {
        firewall.pause();
        firewall.unpause();
        assertFalse(firewall.paused());
    }

    function test_VerifyProof_RevertWhenPaused() public {
        firewall.pause();

        vm.prank(user);
        vm.expectRevert(IdentityFirewall.ContractPaused.selector);
        firewall.verifyHumanProof(validProof);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // OWNERSHIP TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_TransferOwnership() public {
        address newOwner = address(0x999);
        firewall.transferOwnership(newOwner);

        assertEq(firewall.pendingOwner(), newOwner);
        assertEq(firewall.owner(), owner);
    }

    function test_AcceptOwnership() public {
        address newOwner = address(0x999);
        firewall.transferOwnership(newOwner);

        vm.prank(newOwner);
        firewall.acceptOwnership();

        assertEq(firewall.owner(), newOwner);
        assertEq(firewall.pendingOwner(), address(0));
    }

    function test_AcceptOwnership_RevertNoPendingOwner() public {
        vm.prank(user);
        vm.expectRevert(IdentityFirewall.NoPendingOwner.selector);
        firewall.acceptOwnership();
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // FUZZ TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function testFuzz_VerifyProof(bytes memory proof) public {
        vm.assume(proof.length >= 32);
        
        // Make proof unique
        bytes memory uniqueProof = abi.encodePacked(keccak256(proof), bytes32(block.timestamp));

        vm.prank(user);
        bool result = firewall.verifyHumanProof(uniqueProof);

        assertTrue(result);
        assertTrue(firewall.hasValidProof(user));
    }

    function testFuzz_ProofDuration(uint256 duration) public {
        duration = bound(duration, 1 hours, 7 days);
        firewall.setProofDuration(duration);
        assertEq(firewall.proofDuration(), duration);
    }
}
