// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Test, console2} from "forge-std/Test.sol";
import {VigilumRegistry} from "../src/VigilumRegistry.sol";

contract VigilumRegistryTest is Test {
    VigilumRegistry public registry;
    
    address public owner = address(this);
    address public oracle = address(0x1);
    address public user = address(0x2);
    address public targetContract = address(0x3);
    
    bytes32 public constant BYTECODE_HASH = keccak256("test bytecode");

    function setUp() public {
        registry = new VigilumRegistry();
        registry.authorizeOracle(oracle);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // REGISTRATION TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_RegisterContract() public {
        vm.prank(oracle);
        registry.registerContract(targetContract, BYTECODE_HASH, 2500);

        assertEq(registry.getRiskScore(targetContract), 2500);
        assertFalse(registry.isBlacklisted(targetContract));
    }

    function test_RegisterContract_RevertUnauthorized() public {
        vm.prank(user);
        vm.expectRevert(VigilumRegistry.Unauthorized.selector);
        registry.registerContract(targetContract, BYTECODE_HASH, 2500);
    }

    function test_RegisterContract_RevertInvalidScore() public {
        vm.prank(oracle);
        vm.expectRevert(VigilumRegistry.InvalidRiskScore.selector);
        registry.registerContract(targetContract, BYTECODE_HASH, 10001);
    }

    function test_RegisterContract_RevertAlreadyRegistered() public {
        vm.startPrank(oracle);
        registry.registerContract(targetContract, BYTECODE_HASH, 2500);
        
        vm.expectRevert(VigilumRegistry.ContractAlreadyRegistered.selector);
        registry.registerContract(targetContract, BYTECODE_HASH, 3000);
        vm.stopPrank();
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // UPDATE TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_UpdateRiskScore() public {
        vm.startPrank(oracle);
        registry.registerContract(targetContract, BYTECODE_HASH, 2500);
        registry.updateRiskScore(targetContract, 7500, 5);
        vm.stopPrank();

        assertEq(registry.getRiskScore(targetContract), 7500);
    }

    function test_UpdateRiskScore_RevertNotRegistered() public {
        vm.prank(oracle);
        vm.expectRevert(VigilumRegistry.ContractNotRegistered.selector);
        registry.updateRiskScore(targetContract, 5000, 2);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // BLACKLIST TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_BlacklistContract() public {
        vm.prank(oracle);
        registry.blacklistContract(targetContract, "Rug pull detected");

        assertTrue(registry.isBlacklisted(targetContract));
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // THREAT LEVEL TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_ThreatLevel_Critical() public {
        vm.prank(oracle);
        registry.registerContract(targetContract, BYTECODE_HASH, 8500);

        assertEq(uint256(registry.getThreatLevel(targetContract)), uint256(VigilumRegistry.ThreatLevel.CRITICAL));
    }

    function test_ThreatLevel_High() public {
        vm.prank(oracle);
        registry.registerContract(targetContract, BYTECODE_HASH, 7000);

        assertEq(uint256(registry.getThreatLevel(targetContract)), uint256(VigilumRegistry.ThreatLevel.HIGH));
    }

    function test_ThreatLevel_Medium() public {
        vm.prank(oracle);
        registry.registerContract(targetContract, BYTECODE_HASH, 5000);

        assertEq(uint256(registry.getThreatLevel(targetContract)), uint256(VigilumRegistry.ThreatLevel.MEDIUM));
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // OWNERSHIP TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function test_TransferOwnership() public {
        registry.transferOwnership(user);
        
        vm.prank(user);
        registry.acceptOwnership();

        assertEq(registry.owner(), user);
    }

    function test_TransferOwnership_RevertUnauthorized() public {
        vm.prank(user);
        vm.expectRevert(VigilumRegistry.Unauthorized.selector);
        registry.transferOwnership(user);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // FUZZ TESTS
    // ═══════════════════════════════════════════════════════════════════════════

    function testFuzz_RiskScore(uint256 score) public {
        score = bound(score, 0, 10000);
        
        vm.prank(oracle);
        registry.registerContract(targetContract, BYTECODE_HASH, score);

        assertEq(registry.getRiskScore(targetContract), score);
    }
}
