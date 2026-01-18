// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import {ThreatOracle} from "../src/ThreatOracle.sol";
import {IThreatOracle} from "../src/interfaces/IThreatOracle.sol";

contract ThreatOracleTest is Test {
    ThreatOracle public oracle;
    
    address public owner = address(this);
    address public reporter1 = address(0x1);
    address public reporter2 = address(0x2);
    address public user = address(0x3);
    address public target1 = address(0x100);
    address public target2 = address(0x200);
    address public target3 = address(0x300);

    // Events for testing
    event RiskUpdated(address indexed target, uint8 riskScore, uint256 timestamp);
    event OracleReporterUpdated(address indexed reporter, bool isActive);

    function setUp() public {
        // Deploy with initial reporters
        address[] memory initialReporters = new address[](2);
        initialReporters[0] = reporter1;
        initialReporters[1] = reporter2;
        
        oracle = new ThreatOracle(initialReporters);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // DEPLOYMENT TESTS
    // ═══════════════════════════════════════════════════════════════════════════
    
    function test_InitialState() public view {
        assertEq(oracle.owner(), owner);
        assertTrue(oracle.isOracleReporter(owner));
        assertTrue(oracle.isOracleReporter(reporter1));
        assertTrue(oracle.isOracleReporter(reporter2));
        assertFalse(oracle.isOracleReporter(user));
        assertEq(oracle.totalTargets(), 0);
        assertEq(oracle.totalUpdates(), 0);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // UPDATE RISK SCORE TESTS
    // ═══════════════════════════════════════════════════════════════════════════
    
    function test_UpdateRiskScore() public {
        vm.expectEmit(true, false, false, true);
        emit RiskUpdated(target1, 75, block.timestamp);
        
        oracle.updateRiskScore(target1, 75);
        
        assertEq(oracle.getRiskScore(target1), 75);
        assertEq(oracle.getLastUpdate(target1), block.timestamp);
        assertEq(oracle.totalTargets(), 1);
        assertEq(oracle.totalUpdates(), 1);
    }
    
    function test_UpdateRiskScore_AsReporter() public {
        vm.prank(reporter1);
        oracle.updateRiskScore(target1, 50);
        
        assertEq(oracle.getRiskScore(target1), 50);
    }
    
    function test_UpdateRiskScore_RevertNotReporter() public {
        vm.prank(user);
        vm.expectRevert(ThreatOracle.NotOracleReporter.selector);
        oracle.updateRiskScore(target1, 50);
    }
    
    function test_UpdateRiskScore_RevertInvalidScore() public {
        vm.expectRevert(ThreatOracle.InvalidRiskScore.selector);
        oracle.updateRiskScore(target1, 101);
    }
    
    function test_UpdateRiskScore_RevertZeroAddress() public {
        vm.expectRevert(ThreatOracle.ZeroAddress.selector);
        oracle.updateRiskScore(address(0), 50);
    }
    
    function test_UpdateRiskScore_MaxScore() public {
        oracle.updateRiskScore(target1, 100);
        assertEq(oracle.getRiskScore(target1), 100);
    }
    
    function test_UpdateRiskScore_ZeroScore() public {
        oracle.updateRiskScore(target1, 50);
        oracle.updateRiskScore(target1, 0);
        assertEq(oracle.getRiskScore(target1), 0);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // BATCH UPDATE TESTS
    // ═══════════════════════════════════════════════════════════════════════════
    
    function test_BatchUpdateRiskScores() public {
        address[] memory targets = new address[](3);
        targets[0] = target1;
        targets[1] = target2;
        targets[2] = target3;
        
        uint8[] memory scores = new uint8[](3);
        scores[0] = 80;
        scores[1] = 60;
        scores[2] = 40;
        
        oracle.batchUpdateRiskScores(targets, scores);
        
        assertEq(oracle.getRiskScore(target1), 80);
        assertEq(oracle.getRiskScore(target2), 60);
        assertEq(oracle.getRiskScore(target3), 40);
        assertEq(oracle.totalTargets(), 3);
        assertEq(oracle.totalUpdates(), 3);
    }
    
    function test_BatchUpdateRiskScores_RevertArrayMismatch() public {
        address[] memory targets = new address[](3);
        uint8[] memory scores = new uint8[](2);
        
        vm.expectRevert(ThreatOracle.ArrayLengthMismatch.selector);
        oracle.batchUpdateRiskScores(targets, scores);
    }
    
    function test_BatchUpdateRiskScores_RevertBatchTooLarge() public {
        address[] memory targets = new address[](101);
        uint8[] memory scores = new uint8[](101);
        
        vm.expectRevert(ThreatOracle.BatchTooLarge.selector);
        oracle.batchUpdateRiskScores(targets, scores);
    }
    
    function test_BatchUpdateRiskScores_SkipsZeroAddress() public {
        address[] memory targets = new address[](3);
        targets[0] = target1;
        targets[1] = address(0); // Should be skipped
        targets[2] = target2;
        
        uint8[] memory scores = new uint8[](3);
        scores[0] = 80;
        scores[1] = 50;
        scores[2] = 60;
        
        oracle.batchUpdateRiskScores(targets, scores);
        
        assertEq(oracle.getRiskScore(target1), 80);
        assertEq(oracle.getRiskScore(target2), 60);
        assertEq(oracle.totalTargets(), 2);
    }
    
    function test_BatchUpdateRiskScores_MaxBatch() public {
        address[] memory targets = new address[](100);
        uint8[] memory scores = new uint8[](100);
        
        for (uint256 i = 0; i < 100; i++) {
            targets[i] = address(uint160(i + 1000));
            scores[i] = uint8(i);
        }
        
        oracle.batchUpdateRiskScores(targets, scores);
        
        assertEq(oracle.totalTargets(), 100);
        assertEq(oracle.getRiskScore(address(1000)), 0);
        assertEq(oracle.getRiskScore(address(1050)), 50);
        assertEq(oracle.getRiskScore(address(1099)), 99);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // VIEW FUNCTION TESTS
    // ═══════════════════════════════════════════════════════════════════════════
    
    function test_GetTargetInfo() public {
        oracle.updateRiskScore(target1, 75);
        
        (uint8 score, uint256 timestamp) = oracle.getTargetInfo(target1);
        assertEq(score, 75);
        assertEq(timestamp, block.timestamp);
    }
    
    function test_GetRiskLevel_Critical() public {
        oracle.updateRiskScore(target1, 80);
        assertEq(oracle.getRiskLevel(target1), "CRITICAL");
        
        oracle.updateRiskScore(target1, 100);
        assertEq(oracle.getRiskLevel(target1), "CRITICAL");
    }
    
    function test_GetRiskLevel_High() public {
        oracle.updateRiskScore(target1, 60);
        assertEq(oracle.getRiskLevel(target1), "HIGH");
        
        oracle.updateRiskScore(target1, 79);
        assertEq(oracle.getRiskLevel(target1), "HIGH");
    }
    
    function test_GetRiskLevel_Medium() public {
        oracle.updateRiskScore(target1, 40);
        assertEq(oracle.getRiskLevel(target1), "MEDIUM");
        
        oracle.updateRiskScore(target1, 59);
        assertEq(oracle.getRiskLevel(target1), "MEDIUM");
    }
    
    function test_GetRiskLevel_Low() public {
        oracle.updateRiskScore(target1, 20);
        assertEq(oracle.getRiskLevel(target1), "LOW");
        
        oracle.updateRiskScore(target1, 39);
        assertEq(oracle.getRiskLevel(target1), "LOW");
    }
    
    function test_GetRiskLevel_Info() public {
        assertEq(oracle.getRiskLevel(target1), "INFO"); // No score set
        
        oracle.updateRiskScore(target1, 0);
        assertEq(oracle.getRiskLevel(target1), "INFO");
        
        oracle.updateRiskScore(target1, 19);
        assertEq(oracle.getRiskLevel(target1), "INFO");
    }
    
    function test_IsHighRisk() public {
        assertFalse(oracle.isHighRisk(target1));
        
        oracle.updateRiskScore(target1, 59);
        assertFalse(oracle.isHighRisk(target1));
        
        oracle.updateRiskScore(target1, 60);
        assertTrue(oracle.isHighRisk(target1));
        
        oracle.updateRiskScore(target1, 100);
        assertTrue(oracle.isHighRisk(target1));
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // ADMIN FUNCTION TESTS
    // ═══════════════════════════════════════════════════════════════════════════
    
    function test_SetOracleReporter_Add() public {
        address newReporter = address(0x999);
        
        vm.expectEmit(true, false, false, true);
        emit OracleReporterUpdated(newReporter, true);
        
        oracle.setOracleReporter(newReporter, true);
        assertTrue(oracle.isOracleReporter(newReporter));
        
        // Verify new reporter can update
        vm.prank(newReporter);
        oracle.updateRiskScore(target1, 50);
        assertEq(oracle.getRiskScore(target1), 50);
    }
    
    function test_SetOracleReporter_Remove() public {
        assertTrue(oracle.isOracleReporter(reporter1));
        
        vm.expectEmit(true, false, false, true);
        emit OracleReporterUpdated(reporter1, false);
        
        oracle.setOracleReporter(reporter1, false);
        assertFalse(oracle.isOracleReporter(reporter1));
        
        // Verify removed reporter cannot update
        vm.prank(reporter1);
        vm.expectRevert(ThreatOracle.NotOracleReporter.selector);
        oracle.updateRiskScore(target1, 50);
    }
    
    function test_SetOracleReporter_RevertNotOwner() public {
        vm.prank(reporter1);
        vm.expectRevert(ThreatOracle.NotOwner.selector);
        oracle.setOracleReporter(user, true);
    }
    
    function test_SetOracleReporter_RevertZeroAddress() public {
        vm.expectRevert(ThreatOracle.ZeroAddress.selector);
        oracle.setOracleReporter(address(0), true);
    }
    
    function test_TransferOwnership() public {
        address newOwner = address(0x888);
        
        oracle.transferOwnership(newOwner);
        assertEq(oracle.owner(), newOwner);
        
        // Old owner can no longer admin
        vm.expectRevert(ThreatOracle.NotOwner.selector);
        oracle.setOracleReporter(user, true);
        
        // New owner can admin
        vm.prank(newOwner);
        oracle.setOracleReporter(user, true);
        assertTrue(oracle.isOracleReporter(user));
    }
    
    function test_TransferOwnership_RevertNotOwner() public {
        vm.prank(reporter1);
        vm.expectRevert(ThreatOracle.NotOwner.selector);
        oracle.transferOwnership(reporter1);
    }
    
    function test_TransferOwnership_RevertZeroAddress() public {
        vm.expectRevert(ThreatOracle.ZeroAddress.selector);
        oracle.transferOwnership(address(0));
    }
    
    function test_ClearRiskScore() public {
        oracle.updateRiskScore(target1, 75);
        assertEq(oracle.getRiskScore(target1), 75);
        
        vm.expectEmit(true, false, false, true);
        emit RiskUpdated(target1, 0, block.timestamp);
        
        oracle.clearRiskScore(target1);
        assertEq(oracle.getRiskScore(target1), 0);
    }
    
    function test_ClearRiskScore_NoOp() public {
        // Clearing a zero score should not emit event
        oracle.clearRiskScore(target1);
        assertEq(oracle.getRiskScore(target1), 0);
    }
    
    function test_ClearRiskScore_RevertNotOwner() public {
        vm.prank(reporter1);
        vm.expectRevert(ThreatOracle.NotOwner.selector);
        oracle.clearRiskScore(target1);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // FUZZ TESTS
    // ═══════════════════════════════════════════════════════════════════════════
    
    function testFuzz_UpdateRiskScore(address target, uint8 score) public {
        vm.assume(target != address(0));
        vm.assume(score <= 100);
        
        oracle.updateRiskScore(target, score);
        assertEq(oracle.getRiskScore(target), score);
    }
    
    function testFuzz_MultipleUpdates(address target, uint8[5] memory scores) public {
        vm.assume(target != address(0));
        
        for (uint256 i = 0; i < 5; i++) {
            uint8 score = scores[i] % 101; // Ensure valid score
            oracle.updateRiskScore(target, score);
            assertEq(oracle.getRiskScore(target), score);
        }
        
        // Only one target tracked
        assertEq(oracle.totalTargets(), 1);
        // 5 updates made
        assertEq(oracle.totalUpdates(), 5);
    }

    // ═══════════════════════════════════════════════════════════════════════════
    // GAS BENCHMARKS
    // ═══════════════════════════════════════════════════════════════════════════
    
    function test_Gas_SingleUpdate() public {
        uint256 gasBefore = gasleft();
        oracle.updateRiskScore(target1, 75);
        uint256 gasUsed = gasBefore - gasleft();
        
        // Should be reasonably efficient (cold storage write is ~22k, plus overhead)
        assertLt(gasUsed, 150000);
    }
    
    function test_Gas_BatchUpdate10() public {
        address[] memory targets = new address[](10);
        uint8[] memory scores = new uint8[](10);
        
        for (uint256 i = 0; i < 10; i++) {
            targets[i] = address(uint160(i + 1));
            scores[i] = uint8(i * 10);
        }
        
        uint256 gasBefore = gasleft();
        oracle.batchUpdateRiskScores(targets, scores);
        uint256 gasUsed = gasBefore - gasleft();
        
        // Batch of 10 cold storage writes
        assertLt(gasUsed, 600000);
    }
}
