package scanner

import (
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vigilum/backend/internal/domain"
)

func TestMythrilScanner_NewMythrilScanner(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	scanner, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)
	assert.NotNil(t, scanner)
	assert.Equal(t, "myth", scanner.mythrilPath)
	assert.Equal(t, "/tmp/vigilum-mythril", scanner.workDir)
}

func TestMythrilScanner_MapMythrilSeverityToThreatLevel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)

	testCases := []struct {
		severity string
		expected domain.ThreatLevel
	}{
		{"Critical", domain.ThreatLevelCritical},
		{"High", domain.ThreatLevelHigh},
		{"Medium", domain.ThreatLevelMedium},
		{"Low", domain.ThreatLevelLow},
		{"Informational", domain.ThreatLevelInfo},
		{"unknown", domain.ThreatLevelLow},
	}

	for _, tc := range testCases {
		level := scanner.mapMythrilSeverityToThreatLevel(tc.severity)
		assert.Equal(t, tc.expected, level)
	}
}

func TestMythrilScanner_MapMythrilIssueToVulnType(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)

	testCases := []struct {
		issue    MythrilIssue
		expected domain.VulnType
	}{
		{MythrilIssue{SWCID: "SWC-107"}, domain.VulnReentrancy},
		{MythrilIssue{SWCID: "SWC-101", Title: "Integer Underflow"}, domain.VulnUnderflow},
		{MythrilIssue{SWCID: "SWC-101", Title: "Integer Overflow"}, domain.VulnOverflow},
		{MythrilIssue{SWCID: "SWC-115"}, domain.VulnTxOrigin},
		{MythrilIssue{Title: "Delegatecall to user supplied address"}, domain.VulnAccessControl},
		{MythrilIssue{Title: "Unchecked call return value"}, domain.VulnUncheckedCall},
		{MythrilIssue{Title: "Weak PRNG"}, domain.VulnWeakRandomness},
	}

	for _, tc := range testCases {
		vulnType := scanner.mapMythrilIssueToVulnType(tc.issue)
		assert.Equal(t, tc.expected, vulnType)
	}
}

func TestMythrilScanner_ParseMythrilOutput(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)

	jsonOutput := `{"issues":[{"title":"Reentrancy","type":"Reentrancy","severity":"High","swc-id":"SWC-107","lineno":42,"code":"call.value()"}]}`
	issues, err := scanner.parseMythrilOutput([]byte(jsonOutput))
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, "Reentrancy", issues[0].Title)
	assert.Equal(t, "SWC-107", issues[0].SWCID)
	assert.Equal(t, 42, issues[0].LineNo)
}

func TestMythrilScanner_PrepareAnalysisTarget(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)

	t.Run("source code", func(t *testing.T) {
		contract := &domain.Contract{
			ID:      "test-1",
			Address: "0x1234567890123456789012345678901234567890",
			ChainID: 1,
			SourceCode: "pragma solidity ^0.8.0; contract A { function foo() public {} }",
		}

		target, isBytecode, cleanup, err := scanner.prepareAnalysisTarget(contract)
		require.NoError(t, err)
		assert.False(t, isBytecode)

		_, statErr := os.Stat(target)
		assert.NoError(t, statErr)

		cleanup()
		_, statErr = os.Stat(target)
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("bytecode", func(t *testing.T) {
		contract := &domain.Contract{
			ID:       "test-2",
			Address:  "0x1234567890123456789012345678901234567890",
			ChainID:  1,
			Bytecode: []byte{0xde, 0xad, 0xbe, 0xef},
		}

		target, isBytecode, cleanup, err := scanner.prepareAnalysisTarget(contract)
		require.NoError(t, err)
		assert.True(t, isBytecode)
		assert.Equal(t, "deadbeef", strings.ToLower(target))
		cleanup()
	})
}

func TestMythrilScanner_CalculateRiskScore(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)

	t.Run("no vulnerabilities", func(t *testing.T) {
		vulns := []domain.Vulnerability{}
		score := scanner.calculateRiskScore(vulns)
		assert.Equal(t, 0.0, score)
	})

	t.Run("single high vulnerability", func(t *testing.T) {
		vulns := []domain.Vulnerability{{Severity: domain.ThreatLevelHigh, Confidence: 0.8}}
		score := scanner.calculateRiskScore(vulns)
		assert.Greater(t, score, 3.0)
		assert.LessOrEqual(t, score, 10.0)
	})
}

func TestMythrilScanner_CalculateMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scanner, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)

	vulns := []domain.Vulnerability{
		{Severity: domain.ThreatLevelCritical},
		{Severity: domain.ThreatLevelHigh},
		{Severity: domain.ThreatLevelHigh},
		{Severity: domain.ThreatLevelMedium},
		{Severity: domain.ThreatLevelLow},
		{Severity: domain.ThreatLevelInfo},
	}

	metrics := scanner.calculateMetrics(vulns, &domain.Contract{}, 2*time.Second)
	assert.Equal(t, 6, metrics.TotalIssues)
	assert.Equal(t, 1, metrics.CriticalCount)
	assert.Equal(t, 2, metrics.HighCount)
	assert.Equal(t, 1, metrics.MediumCount)
	assert.Equal(t, 1, metrics.LowCount)
	assert.Equal(t, 1, metrics.InfoCount)
}
