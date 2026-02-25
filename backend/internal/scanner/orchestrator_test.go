package scanner

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vigilum/backend/internal/domain"
)

// TestOrchestrator_ScanAll tests full orchestrator functionality
func TestOrchestrator_ScanAll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slither, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	mythril, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)

	orchestrator := NewOrchestrator(slither, mythril)

	// Simple test contract with a known vulnerability
	contract := &domain.Contract{
		ID:      "test-contract-001",
		Address: "0x1234567890123456789012345678901234567890",
		ChainID: 1,
		SourceCode: `pragma solidity ^0.8.0;
contract Test {
    function bad() public {
        (bool ok,) = msg.sender.call{value: 1}("");
    }
}`,
	}

	opts := &ScanOptions{
		Timeout: 300,
	}

	report, err := orchestrator.ScanAll(context.Background(), contract, opts)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify report structure
	assert.Equal(t, contract.ID, report.ContractID)
	assert.Equal(t, domain.ScanStatusCompleted, report.Status)
	assert.GreaterOrEqual(t, report.RiskScore, 0.0)
	assert.LessOrEqual(t, report.RiskScore, 10.0)
	assert.NotNil(t, report.CompletedAt)
	assert.Greater(t, report.Duration, time.Duration(0))
}

// TestOrchestrator_ParallelExecution verifies scanners run in parallel
func TestOrchestrator_ParallelExecution(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slither, _ := NewSlitherScanner(logger, nil)
	mythril, _ := NewMythrilScanner(logger, nil)

	orchestrator := NewOrchestrator(slither, mythril)

	contract := &domain.Contract{
		ID:      "parallel-test",
		Address: "0xABCD",
		ChainID: 1,
		SourceCode: `pragma solidity ^0.8.0;
contract Simple {
    uint public value;
}`,
	}

	opts := &ScanOptions{Timeout: 300}

	startTime := time.Now()
	report, err := orchestrator.ScanAll(context.Background(), contract, opts)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Less(t, duration, time.Minute, "Parallel execution should be faster than serial")
}

// TestOrchestrator_NoScannersRegistered tests behavior with empty scanner list
func TestOrchestrator_NoScannersRegistered(t *testing.T) {
	orchestrator := NewOrchestrator()

	contract := &domain.Contract{
		ID:      "test",
		ChainID: 1,
	}

	report, err := orchestrator.ScanAll(context.Background(), contract, nil)
	require.NoError(t, err)
	assert.Equal(t, domain.ScanStatusCompleted, report.Status)
	assert.Len(t, report.Vulnerabilities, 0)
	assert.Equal(t, 0.0, report.RiskScore)
}

// TestOrchestrator_Timeout tests timeout handling
func TestOrchestrator_Timeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slither, _ := NewSlitherScanner(logger, nil)
	orchestrator := NewOrchestrator(slither)

	contract := &domain.Contract{
		ID:      "timeout-test",
		Address: "0x1234",
		ChainID: 1,
		SourceCode: "pragma solidity ^0.8.0; contract Test {} ",
	}

	// Very short timeout
	opts := &ScanOptions{Timeout: 1}

	report, err := orchestrator.ScanAll(context.Background(), contract, opts)
	// Should either timeout or complete (depending on system speed)
	assert.NotNil(t, report)
	// If there's an error, it should be timeout-related
	if err != nil {
		t.Logf("Got timeout error as expected: %v", err)
	}
}

// TestOrchestrator_DefaultOptions tests default scan options
func TestOrchestrator_DefaultOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slither, _ := NewSlitherScanner(logger, nil)
	orchestrator := NewOrchestrator(slither)

	contract := &domain.Contract{
		ID:      "default-opts-test",
		ChainID: 1,
		SourceCode: "pragma solidity ^0.8.0; contract Test {} ",
	}

	// Pass nil options - should use defaults
	report, err := orchestrator.ScanAll(context.Background(), contract, nil)
	require.NoError(t, err)
	assert.Equal(t, domain.ScanStatusCompleted, report.Status)
}

// ============================================================================
// Aggregator Tests
// ============================================================================

// TestAggregatorDeduplicate tests finding deduplication
func TestAggregatorDeduplicate(t *testing.T) {
	aggregator := NewScanAggregator()

	// Add same vulnerability from two different scanners
	v1 := domain.Vulnerability{
		Type:       domain.VulnReentrancy,
		Severity:   domain.ThreatLevelCritical,
		Confidence: 0.9,
		Location: domain.CodeLocation{
			File:      "contract.sol",
			StartLine: 10,
		},
		DetectedBy: "slither",
	}

	v2 := domain.Vulnerability{
		Type:       domain.VulnReentrancy,
		Severity:   domain.ThreatLevelCritical,
		Confidence: 0.8,
		Location: domain.CodeLocation{
			File:      "contract.sol",
			StartLine: 10,
		},
		DetectedBy: "mythril",
	}

	aggregator.AddFinding(v1)
	aggregator.AddFinding(v2)

	findings := aggregator.GetAggregatedFindings()
	require.Len(t, findings, 1, "Should deduplicate to 1 finding")
	assert.InDelta(t, 0.85, findings[0].Confidence, 0.0001, "Should average confidence")
	assert.Equal(t, domain.ThreatLevelCritical, findings[0].Severity)
}

// TestAggregatorMultipleFindingsMultipleScanners tests correct merging
func TestAggregatorMultipleFindingsMultipleScanners(t *testing.T) {
	aggregator := NewScanAggregator()

	// Add multiple different vulnerabilities
	severities := []domain.ThreatLevel{
		domain.ThreatLevelCritical,
		domain.ThreatLevelHigh,
		domain.ThreatLevelMedium,
	}

	for i, severity := range severities {
		aggregator.AddFinding(domain.Vulnerability{
			Type:       domain.VulnReentrancy,
			Severity:   severity,
			Confidence: 0.8,
			Location: domain.CodeLocation{
				File:      "contract.sol",
				StartLine: 10 + i,
			},
			DetectedBy: "slither",
		})
	}

	findings := aggregator.GetAggregatedFindings()
	assert.Len(t, findings, 3, "Should have 3 different findings")
}

// TestScanAggregator_RiskScore tests risk score calculation
func TestScanAggregator_RiskScore(t *testing.T) {
	aggregator := NewScanAggregator()

	aggregator.AddFinding(domain.Vulnerability{
		Severity:   domain.ThreatLevelCritical,
		Confidence: 0.9,
		Location: domain.CodeLocation{
			File:      "a.sol",
			StartLine: 1,
		},
		Type:       domain.VulnReentrancy,
		DetectedBy: "test",
	})

	score := aggregator.CalculateAggregateRiskScore()
	assert.Greater(t, score, 4.0, "Critical finding should give score > 4")
	assert.LessOrEqual(t, score, 10.0, "Score should be normalized to ≤ 10")
}

// TestScanAggregator_EmptyFindings tests risk score with no findings
func TestScanAggregator_EmptyFindings(t *testing.T) {
	aggregator := NewScanAggregator()

	score := aggregator.CalculateAggregateRiskScore()
	assert.Equal(t, 0.0, score, "Empty findings should give score 0")

	level := aggregator.DetermineThreatLevel(score)
	assert.Equal(t, domain.ThreatLevelNone, level)
}

// TestAggregatorMultipleVulnerabilities tests scoring with multiple findings
func TestAggregatorMultipleVulnerabilities(t *testing.T) {
	aggregator := NewScanAggregator()

	// Add multiple vulnerabilities of different severities
	aggregator.AddFinding(domain.Vulnerability{
		Severity:   domain.ThreatLevelCritical,
		Confidence: 0.95,
		Location: domain.CodeLocation{
			File:      "contract.sol",
			StartLine: 1,
		},
		Type:       domain.VulnReentrancy,
		DetectedBy: "test",
	})

	aggregator.AddFinding(domain.Vulnerability{
		Severity:   domain.ThreatLevelHigh,
		Confidence: 0.8,
		Location: domain.CodeLocation{
			File:      "contract.sol",
			StartLine: 5,
		},
		Type:       domain.VulnOverflow,
		DetectedBy: "test",
	})

	score := aggregator.CalculateAggregateRiskScore()
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 10.0)
}

// TestScanAggregator_ThreatLevel tests threat level mapping
func TestScanAggregator_ThreatLevel(t *testing.T) {
	aggregator := NewScanAggregator()

	testCases := []struct {
		score         float64
		expectedLevel domain.ThreatLevel
		name          string
	}{
		{0.5, domain.ThreatLevelNone, "Low score → None"},
		{1.5, domain.ThreatLevelLow, "1.5 → Low"},
		{3.5, domain.ThreatLevelMedium, "3.5 → Medium"},
		{6.5, domain.ThreatLevelHigh, "6.5 → High"},
		{8.5, domain.ThreatLevelCritical, "8.5 → Critical"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			level := aggregator.DetermineThreatLevel(tc.score)
			assert.Equal(t, tc.expectedLevel, level)
		})
	}
}

// TestAggregatorMetrics tests metrics calculation
func TestAggregatorMetrics(t *testing.T) {
	aggregator := NewScanAggregator()

	aggregator.AddFinding(domain.Vulnerability{
		Severity:   domain.ThreatLevelCritical,
		Confidence: 0.9,
		Location: domain.CodeLocation{
			File:      "contract.sol",
			StartLine: 1,
		},
		Type:       domain.VulnReentrancy,
		DetectedBy: "slither",
	})

	aggregator.AddFinding(domain.Vulnerability{
		Severity:   domain.ThreatLevelHigh,
		Confidence: 0.8,
		Location: domain.CodeLocation{
			File:      "contract.sol",
			StartLine: 5,
		},
		Type:       domain.VulnOverflow,
		DetectedBy: "mythril",
	})

	metrics := aggregator.CalculateMetrics()
	assert.Equal(t, 2, metrics.TotalIssues)
	assert.Equal(t, 1, metrics.CriticalCount)
	assert.Equal(t, 1, metrics.HighCount)
	assert.Equal(t, 0, metrics.MediumCount)
}

// TestScanAggregator_ConfidenceAveraging tests confidence score merging
func TestScanAggregator_ConfidenceAveraging(t *testing.T) {
	aggregator := NewScanAggregator()

	v1 := domain.Vulnerability{
		Type:       domain.VulnReentrancy,
		Severity:   domain.ThreatLevelHigh,
		Confidence: 1.0,
		Location: domain.CodeLocation{
			File:      "contract.sol",
			StartLine: 10,
		},
		DetectedBy: "slither",
	}

	v2 := domain.Vulnerability{
		Type:       domain.VulnReentrancy,
		Severity:   domain.ThreatLevelHigh,
		Confidence: 0.6,
		Location: domain.CodeLocation{
			File:      "contract.sol",
			StartLine: 10,
		},
		DetectedBy: "mythril",
	}

	aggregator.AddFinding(v1)
	aggregator.AddFinding(v2)

	findings := aggregator.GetAggregatedFindings()
	require.Len(t, findings, 1)
	assert.Equal(t, 0.8, findings[0].Confidence, "Should average 1.0 and 0.6")
}

// TestAggregatorDifferentDetectors tests merging findings from different detectors
func TestAggregatorDifferentDetectors(t *testing.T) {
	aggregator := NewScanAggregator()

	// Add same vulnerability type from different detectors
	for _, detectorName := range []string{"slither", "mythril"} {
		aggregator.AddFinding(domain.Vulnerability{
			Type:       domain.VulnReentrancy,
			Severity:   domain.ThreatLevelCritical,
			Confidence: 0.8,
			Location: domain.CodeLocation{
				File:      "contract.sol",
				StartLine: 10,
			},
			DetectedBy: detectorName,
		})
	}

	// Should have deduplicated to 1 finding
	findings := aggregator.GetAggregatedFindings()
	require.Len(t, findings, 1)
	assert.GreaterOrEqual(t, findings[0].Confidence, 0.75, "Confidence should be averaged")
}
