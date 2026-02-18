// Package scanner provides tests for the Slither scanner integration.
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

// TestSlitherScanner_NewSlitherScanner tests scanner initialization.
func TestSlitherScanner_NewSlitherScanner(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("default config", func(t *testing.T) {
		scanner, err := NewSlitherScanner(logger, nil)
		require.NoError(t, err)
		assert.NotNil(t, scanner)
		assert.Equal(t, "slither", scanner.Name())
		assert.Equal(t, domain.ScanTypeStatic, scanner.ScanType())
	})

	t.Run("custom config", func(t *testing.T) {
		config := &SlitherConfig{
			SlitherPath:   "/usr/bin/slither",
			WorkDir:       "/tmp/test-slither",
			Timeout:       3 * time.Minute,
			EnabledChecks: []string{"reentrancy-eth", "arbitrary-send"},
		}

		scanner, err := NewSlitherScanner(logger, config)
		require.NoError(t, err)
		assert.Equal(t, config.SlitherPath, scanner.slitherPath)
		assert.Equal(t, config.WorkDir, scanner.workDir)
		assert.Equal(t, config.Timeout, scanner.timeout)
		assert.Equal(t, 2, len(scanner.enabledChecks))
	})

	t.Run("work directory created", func(t *testing.T) {
		tempDir := "/tmp/slither-test-" + time.Now().Format("20060102150405")
		config := &SlitherConfig{
			WorkDir: tempDir,
		}

		scanner, err := NewSlitherScanner(logger, config)
		require.NoError(t, err)

		// Verify directory exists
		info, err := os.Stat(tempDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Cleanup
		os.RemoveAll(tempDir)
	})
}

// TestSlitherScanner_SupportedChains tests chain support.
func TestSlitherScanner_SupportedChains(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	chains := scanner.SupportedChains()
	assert.Greater(t, len(chains), 0)
	assert.Contains(t, chains, domain.ChainID(1))  // Ethereum Mainnet
	assert.Contains(t, chains, domain.ChainID(137)) // Polygon
}

// TestSlitherScanner_IsHealthy tests health check.
func TestSlitherScanner_IsHealthy(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	ctx := context.Background()
	
	// Note: This test may fail if Slither is not installed
	// In CI/CD, ensure Slither is available or skip this test
	healthy := scanner.IsHealthy(ctx)
	t.Logf("Slither health check: %v", healthy)
}

// TestSlitherScanner_MapSlitherCheckToVulnType tests vulnerability type mapping.
func TestSlitherScanner_MapSlitherCheckToVulnType(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	testCases := []struct {
		check    string
		expected domain.VulnType
	}{
		{"reentrancy-eth", domain.VulnReentrancy},
		{"reentrancy-no-eth", domain.VulnReentrancy},
		{"arbitrary-send", domain.VulnAccessControl},
		{"tx-origin", domain.VulnTxOrigin},
		{"unchecked-send", domain.VulnUncheckedCall},
		{"timestamp", domain.VulnTimestamp},
		{"weak-prng", domain.VulnWeakRandomness},
		{"unknown-check", domain.VulnLogicError}, // Default
	}

	for _, tc := range testCases {
		t.Run(tc.check, func(t *testing.T) {
			vulnType := scanner.mapSlitherCheckToVulnType(tc.check)
			assert.Equal(t, tc.expected, vulnType)
		})
	}
}

// TestSlitherScanner_MapSlitherImpactToSeverity tests severity mapping.
func TestSlitherScanner_MapSlitherImpactToSeverity(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	testCases := []struct {
		impact   string
		expected domain.ThreatLevel
	}{
		{"High", domain.ThreatLevelCritical},
		{"high", domain.ThreatLevelCritical},
		{"Medium", domain.ThreatLevelHigh},
		{"Low", domain.ThreatLevelMedium},
		{"Informational", domain.ThreatLevelInfo},
		{"unknown", domain.ThreatLevelMedium}, // Default
	}

	for _, tc := range testCases {
		t.Run(tc.impact, func(t *testing.T) {
			severity := scanner.mapSlitherImpactToSeverity(tc.impact)
			assert.Equal(t, tc.expected, severity)
		})
	}
}

// TestSlitherScanner_MapSlitherConfidenceToFloat tests confidence mapping.
func TestSlitherScanner_MapSlitherConfidenceToFloat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	testCases := []struct {
		confidence string
		expected   float64
	}{
		{"High", 0.9},
		{"high", 0.9},
		{"Medium", 0.7},
		{"Low", 0.5},
		{"unknown", 0.6}, // Default
	}

	for _, tc := range testCases {
		t.Run(tc.confidence, func(t *testing.T) {
			conf := scanner.mapSlitherConfidenceToFloat(tc.confidence)
			assert.InDelta(t, tc.expected, conf, 0.01)
		})
	}
}

// TestSlitherScanner_CalculateRiskScore tests risk score calculation.
func TestSlitherScanner_CalculateRiskScore(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	t.Run("no vulnerabilities", func(t *testing.T) {
		vulns := []domain.Vulnerability{}
		score := scanner.calculateRiskScore(vulns)
		assert.Equal(t, 0.0, score)
	})

	t.Run("single critical vulnerability", func(t *testing.T) {
		vulns := []domain.Vulnerability{
			{
				Severity:   domain.ThreatLevelCritical,
				Confidence: 0.9,
			},
		}
		score := scanner.calculateRiskScore(vulns)
		assert.Greater(t, score, 5.0)
		assert.LessOrEqual(t, score, 10.0)
	})

	t.Run("multiple vulnerabilities", func(t *testing.T) {
		vulns := []domain.Vulnerability{
			{Severity: domain.ThreatLevelCritical, Confidence: 0.9},
			{Severity: domain.ThreatLevelHigh, Confidence: 0.8},
			{Severity: domain.ThreatLevelMedium, Confidence: 0.7},
			{Severity: domain.ThreatLevelLow, Confidence: 0.5},
		}
		score := scanner.calculateRiskScore(vulns)
		assert.Greater(t, score, 0.0)
		assert.LessOrEqual(t, score, 10.0)
		t.Logf("Risk score for %d vulns: %.2f", len(vulns), score)
	})
}

// TestSlitherScanner_DetermineThreatLevel tests threat level determination.
func TestSlitherScanner_DetermineThreatLevel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	testCases := []struct {
		riskScore float64
		expected  domain.ThreatLevel
	}{
		{0.0, domain.ThreatLevelNone},
		{0.5, domain.ThreatLevelNone},
		{1.5, domain.ThreatLevelLow},
		{3.5, domain.ThreatLevelMedium},
		{6.5, domain.ThreatLevelHigh},
		{8.5, domain.ThreatLevelCritical},
		{10.0, domain.ThreatLevelCritical},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			level := scanner.determineThreatLevel(tc.riskScore)
			assert.Equal(t, tc.expected, level)
		})
	}
}

// TestSlitherScanner_PrepareContractFile tests contract file preparation.
func TestSlitherScanner_PrepareContractFile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := &SlitherConfig{
		WorkDir: "/tmp/slither-test-prepare",
	}
	scanner, err := NewSlitherScanner(logger, config)
	require.NoError(t, err)
	defer os.RemoveAll(config.WorkDir)

	t.Run("with source code", func(t *testing.T) {
		contract := &domain.Contract{
			ID:      "test-1",
			Address: "0x1234567890123456789012345678901234567890",
			ChainID: 1,
			SourceCode: `// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract SimpleStorage {
    uint256 public value;
    
    function setValue(uint256 _value) public {
        value = _value;
    }
}`,
		}

		filepath, err := scanner.prepareContractFile(contract)
		require.NoError(t, err)
		defer os.Remove(filepath)

		// Verify file exists and contains source code
		content, err := os.ReadFile(filepath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "SimpleStorage")
		assert.Contains(t, string(content), "setValue")
	})

	t.Run("with bytecode only", func(t *testing.T) {
		contract := &domain.Contract{
			ID:       "test-2",
			Address:  "0x1234567890123456789012345678901234567890",
			ChainID:  1,
			Bytecode: []byte{0x60, 0x80, 0x60, 0x40}, // Sample bytecode
		}

		filepath, err := scanner.prepareContractFile(contract)
		require.NoError(t, err)
		defer os.Remove(filepath)

		// Verify file exists with generated placeholder
		content, err := os.ReadFile(filepath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "UnknownContract")
		assert.Contains(t, string(content), contract.Address.(string))
	})

	t.Run("no source or bytecode", func(t *testing.T) {
		contract := &domain.Contract{
			ID:      "test-3",
			Address: "0x1234567890123456789012345678901234567890",
			ChainID: 1,
		}

		_, err := scanner.prepareContractFile(contract)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no source code or bytecode")
	})
}

// TestSlitherScanner_GenerateMinimalSource tests minimal source generation.
func TestSlitherScanner_GenerateMinimalSource(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	contract := &domain.Contract{
		ID:       "test",
		Address:  "0x" + "ab"*20,
		ChainID:  1,
		Bytecode: make([]byte, 1024),
	}

	source := scanner.generateMinimalSource(contract)
	assert.Contains(t, source, "pragma solidity")
	assert.Contains(t, source, contract.Address.(string))
	assert.Contains(t, source, "1024 bytes")
}

// TestSlitherScanner_GetRemediationAdvice tests remediation advice.
func TestSlitherScanner_GetRemediationAdvice(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	testCases := []struct {
		check    string
		hasAdvice bool
	}{
		{"reentrancy-eth", true},
		{"tx-origin", true},
		{"unchecked-lowlevel", true},
		{"unknown-check", false},
	}

	for _, tc := range testCases {
		t.Run(tc.check, func(t *testing.T) {
			advice := scanner.getRemediationAdvice(tc.check)
			if tc.hasAdvice {
				assert.NotEmpty(t, advice)
			} else {
				assert.Empty(t, advice)
			}
		})
	}
}

// TestSlitherScanner_CalculateMetrics tests metrics calculation.
func TestSlitherScanner_CalculateMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	vulns := []domain.Vulnerability{
		{Severity: domain.ThreatLevelCritical},
		{Severity: domain.ThreatLevelCritical},
		{Severity: domain.ThreatLevelHigh},
		{Severity: domain.ThreatLevelMedium},
		{Severity: domain.ThreatLevelLow},
		{Severity: domain.ThreatLevelInfo},
	}

	contract := &domain.Contract{
		Bytecode: make([]byte, 2048),
	}

	metrics := scanner.calculateMetrics(vulns, contract, 5*time.Second)

	assert.Equal(t, 6, metrics.TotalIssues)
	assert.Equal(t, 2, metrics.CriticalCount)
	assert.Equal(t, 1, metrics.HighCount)
	assert.Equal(t, 1, metrics.MediumCount)
	assert.Equal(t, 1, metrics.LowCount)
	assert.Equal(t, 1, metrics.InfoCount)
}

// TestSlitherScanner_ParseSlitherOutput tests JSON parsing.
func TestSlitherScanner_ParseSlitherOutput(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	t.Run("valid JSON", func(t *testing.T) {
		jsonOutput := []byte(`{
			"success": true,
			"results": {
				"detectors": [
					{
						"check": "reentrancy-eth",
						"confidence": "High",
						"impact": "High",
						"description": "Reentrancy vulnerability found",
						"elements": []
					}
				]
			}
		}`)

		findings, err := scanner.parseSlitherOutput(jsonOutput)
		require.NoError(t, err)
		assert.Len(t, findings, 1)
		assert.Equal(t, "reentrancy-eth", findings[0].Check)
		assert.Equal(t, "High", findings[0].Confidence)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		jsonOutput := []byte(`not valid json`)

		findings, err := scanner.parseSlitherOutput(jsonOutput)
		require.NoError(t, err) // Should not error, returns empty findings
		assert.Len(t, findings, 0)
	})

	t.Run("error in output", func(t *testing.T) {
		jsonOutput := []byte(`{
			"success": false,
			"error": "Compilation failed"
		}`)

		_, err := scanner.parseSlitherOutput(jsonOutput)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Compilation failed")
	})
}

// Benchmark tests
func BenchmarkSlitherScanner_CalculateRiskScore(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, _ := NewSlitherScanner(logger, nil)

	vulns := []domain.Vulnerability{
		{Severity: domain.ThreatLevelCritical, Confidence: 0.9},
		{Severity: domain.ThreatLevelHigh, Confidence: 0.8},
		{Severity: domain.ThreatLevelMedium, Confidence: 0.7},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.calculateRiskScore(vulns)
	}
}
