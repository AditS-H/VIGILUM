// Package scanner provides Mythril symbolic execution integration.
package scanner

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// MythrilScanner performs symbolic execution analysis using Mythril.
type MythrilScanner struct {
	logger      *slog.Logger
	mythrilPath string
	workDir     string
	timeout     time.Duration
	maxDepth    int
}

// MythrilConfig contains configuration for Mythril scanner.
type MythrilConfig struct {
	MythrilPath string
	WorkDir     string
	Timeout     time.Duration
	MaxDepth    int
}

// DefaultMythrilConfig returns sensible defaults.
func DefaultMythrilConfig() *MythrilConfig {
	return &MythrilConfig{
		MythrilPath: "myth",
		WorkDir:     "/tmp/vigilum-mythril",
		Timeout:     10 * time.Minute,
		MaxDepth:    50,
	}
}

// NewMythrilScanner creates a new Mythril scanner instance.
func NewMythrilScanner(logger *slog.Logger, config *MythrilConfig) (*MythrilScanner, error) {
	if config == nil {
		config = DefaultMythrilConfig()
	}

	if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	if _, err := exec.LookPath(config.MythrilPath); err != nil {
		logger.Warn("Mythril not found in PATH", "path", config.MythrilPath, "error", err)
	}

	return &MythrilScanner{
		logger:      logger,
		mythrilPath: config.MythrilPath,
		workDir:     config.WorkDir,
		timeout:     config.Timeout,
		maxDepth:    config.MaxDepth,
	}, nil
}

// Name returns the scanner identifier.
func (m *MythrilScanner) Name() string {
	return "mythril"
}

// ScanType returns the type of analysis performed.
func (m *MythrilScanner) ScanType() domain.ScanType {
	return domain.ScanTypeSymbolic
}

// Scan performs symbolic execution analysis on a contract using Mythril.
func (m *MythrilScanner) Scan(ctx context.Context, contract *domain.Contract) (*ScanResult, error) {
	startTime := time.Now()
	m.logger.Info("Starting Mythril scan",
		"contract_id", contract.ID,
		"address", contract.Address,
		"chain_id", contract.ChainID,
	)

	target, isBytecode, cleanup, err := m.prepareAnalysisTarget(contract)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare analysis target: %w", err)
	}
	defer cleanup()

	output, err := m.runMythril(ctx, target, isBytecode)
	if err != nil {
		m.logger.Error("Mythril execution failed", "error", err)
		return nil, fmt.Errorf("mythril execution failed: %w", err)
	}

	findings, err := m.parseMythrilOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mythril output: %w", err)
	}

	vulnerabilities := m.mapFindingsToVulnerabilities(findings, contract)
	duration := time.Since(startTime)
	riskScore := m.calculateRiskScore(vulnerabilities)
	threatLevel := m.determineThreatLevel(riskScore)
	metrics := m.calculateMetrics(vulnerabilities, contract, duration)

	m.logger.Info("Mythril scan completed",
		"contract_id", contract.ID,
		"vulnerabilities", len(vulnerabilities),
		"risk_score", riskScore,
		"threat_level", threatLevel,
		"duration", duration,
	)

	return &ScanResult{
		Vulnerabilities: vulnerabilities,
		RiskScore:       riskScore,
		ThreatLevel:     threatLevel,
		Metrics:         metrics,
		RawOutput:       output,
		Metadata: map[string]any{
			"scanner":       "mythril",
			"version":       m.getMythrilVersion(),
			"analysis_time": duration.Seconds(),
			"max_depth":     m.maxDepth,
		},
	}, nil
}

// SupportedChains returns chains this scanner supports.
func (m *MythrilScanner) SupportedChains() []domain.ChainID {
	return []domain.ChainID{
		1,
		5,
		11155111,
		137,
		56,
		42161,
		10,
		8453,
	}
}

// IsHealthy checks if Mythril is operational.
func (m *MythrilScanner) IsHealthy(ctx context.Context) bool {
	if _, err := exec.LookPath(m.mythrilPath); err != nil {
		m.logger.Warn("Mythril health check failed: not found in PATH")
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, m.mythrilPath, "--version")
	if err := cmd.Run(); err != nil {
		m.logger.Warn("Mythril health check failed: version check", "error", err)
		return false
	}

	return true
}

// prepareAnalysisTarget prepares a source file or bytecode for analysis.
func (m *MythrilScanner) prepareAnalysisTarget(contract *domain.Contract) (string, bool, func(), error) {
	if contract.SourceCode != "" {
		filename := fmt.Sprintf("contract_%s_%d.sol", contract.Address, time.Now().Unix())
		path := filepath.Join(m.workDir, filename)

		if err := os.WriteFile(path, []byte(contract.SourceCode), 0644); err != nil {
			return "", false, func() {}, fmt.Errorf("failed to write contract file: %w", err)
		}

		cleanup := func() {
			_ = os.Remove(path)
		}

		return path, false, cleanup, nil
	}

	if len(contract.Bytecode) > 0 {
		bytecodeHex := hex.EncodeToString(contract.Bytecode)
		return bytecodeHex, true, func() {}, nil
	}

	return "", false, func() {}, fmt.Errorf("contract has no source code or bytecode")
}

// runMythril executes Mythril and returns JSON output.
func (m *MythrilScanner) runMythril(ctx context.Context, target string, isBytecode bool) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	args := []string{"analyze"}
	if isBytecode {
		args = append(args, "-x", target)
	} else {
		args = append(args, target)
	}

	args = append(args,
		"-o", "json",
		"--execution-timeout", fmt.Sprintf("%d", int(m.timeout.Seconds())),
		"--max-depth", fmt.Sprintf("%d", m.maxDepth),
	)

	cmd := exec.CommandContext(ctx, m.mythrilPath, args...)
	output, _ := cmd.CombinedOutput()

	if ctx.Err() != nil {
		return nil, fmt.Errorf("mythril execution timed out: %w", ctx.Err())
	}

	return output, nil
}

// MythrilOutput represents Mythril JSON output.
type MythrilOutput struct {
	Issues []MythrilIssue `json:"issues"`
}

// MythrilIssue represents a single Mythril finding.
type MythrilIssue struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	SWCID       string `json:"swc-id"`
	Function    string `json:"function"`
	Code        string `json:"code"`
	Address     string `json:"address"`
	LineNo      int    `json:"lineno"`
	Filename    string `json:"filename"`
}

// parseMythrilOutput parses Mythril JSON output.
func (m *MythrilScanner) parseMythrilOutput(data []byte) ([]MythrilIssue, error) {
	var output MythrilOutput
	if err := json.Unmarshal(data, &output); err != nil {
		m.logger.Warn("Failed to parse Mythril JSON, checking for plain text output")
		return []MythrilIssue{}, nil
	}

	return output.Issues, nil
}

// mapFindingsToVulnerabilities converts Mythril findings to domain vulnerabilities.
func (m *MythrilScanner) mapFindingsToVulnerabilities(findings []MythrilIssue, contract *domain.Contract) []domain.Vulnerability {
	vulnerabilities := make([]domain.Vulnerability, 0, len(findings))

	for _, finding := range findings {
		vulnType := m.mapMythrilIssueToVulnType(finding)
		severity := m.mapMythrilSeverityToThreatLevel(finding.Severity)

		vuln := domain.Vulnerability{
			ID:          fmt.Sprintf("mythril-%s-%d", finding.Type, time.Now().Unix()),
			ContractID:  contract.ID,
			Type:        vulnType,
			Severity:    severity,
			Title:       m.generateTitle(finding),
			Description: finding.Description,
			Location:    m.extractCodeLocation(finding),
			Confidence:  m.mapMythrilConfidence(severity),
			DetectedBy:  "mythril",
			DetectedAt:  time.Now(),
			IsConfirmed: false,
			IsFalsePos:  false,
			CWE:         finding.SWCID,
		}

		if advice := m.getRemediationAdvice(finding); advice != "" {
			vuln.Remediation = advice
		}

		vulnerabilities = append(vulnerabilities, vuln)
	}

	return vulnerabilities
}

// mapMythrilIssueToVulnType maps Mythril issues to domain vulnerability types.
func (m *MythrilScanner) mapMythrilIssueToVulnType(issue MythrilIssue) domain.VulnType {
	swc := strings.ToUpper(strings.TrimSpace(issue.SWCID))
	title := strings.ToLower(issue.Title)
	issueType := strings.ToLower(issue.Type)

	switch swc {
	case "SWC-107":
		return domain.VulnReentrancy
	case "SWC-101":
		if strings.Contains(title, "underflow") {
			return domain.VulnUnderflow
		}
		return domain.VulnOverflow
	case "SWC-104":
		return domain.VulnUncheckedCall
	case "SWC-115":
		return domain.VulnTxOrigin
	case "SWC-116":
		return domain.VulnTimestamp
	case "SWC-120":
		return domain.VulnWeakRandomness
	case "SWC-105":
		return domain.VulnAccessControl
	case "SWC-128":
		return domain.VulnDOS
	case "SWC-132":
		return domain.VulnStorageCollision
	}

	if strings.Contains(title, "reentrancy") || strings.Contains(issueType, "reentrancy") {
		return domain.VulnReentrancy
	}
	if strings.Contains(title, "tx.origin") || strings.Contains(issueType, "tx.origin") {
		return domain.VulnTxOrigin
	}
	if strings.Contains(title, "timestamp") {
		return domain.VulnTimestamp
	}
	if strings.Contains(title, "random") || strings.Contains(title, "prng") {
		return domain.VulnWeakRandomness
	}
	if strings.Contains(title, "delegatecall") {
		return domain.VulnAccessControl
	}
	if strings.Contains(title, "unchecked") {
		return domain.VulnUncheckedCall
	}
	if strings.Contains(title, "overflow") {
		return domain.VulnOverflow
	}
	if strings.Contains(title, "underflow") {
		return domain.VulnUnderflow
	}

	return domain.VulnLogicError
}

// mapMythrilSeverityToThreatLevel maps Mythril severity to threat level.
func (m *MythrilScanner) mapMythrilSeverityToThreatLevel(severity string) domain.ThreatLevel {
	switch strings.ToLower(severity) {
	case "critical":
		return domain.ThreatLevelCritical
	case "high":
		return domain.ThreatLevelHigh
	case "medium":
		return domain.ThreatLevelMedium
	case "low":
		return domain.ThreatLevelLow
	case "informational", "info":
		return domain.ThreatLevelInfo
	default:
		return domain.ThreatLevelLow
	}
}

// mapMythrilConfidence returns a default confidence score based on severity.
func (m *MythrilScanner) mapMythrilConfidence(severity domain.ThreatLevel) float64 {
	switch severity {
	case domain.ThreatLevelCritical:
		return 0.9
	case domain.ThreatLevelHigh:
		return 0.8
	case domain.ThreatLevelMedium:
		return 0.7
	case domain.ThreatLevelLow:
		return 0.5
	case domain.ThreatLevelInfo:
		return 0.4
	default:
		return 0.6
	}
}

// generateTitle creates a human-readable title from a Mythril finding.
func (m *MythrilScanner) generateTitle(finding MythrilIssue) string {
	if finding.Title != "" {
		return finding.Title
	}

	if finding.Type != "" {
		return strings.Title(strings.ReplaceAll(finding.Type, "_", " "))
	}

	return "Mythril Finding"
}

// extractCodeLocation extracts source code location from Mythril finding.
func (m *MythrilScanner) extractCodeLocation(finding MythrilIssue) domain.CodeLocation {
	location := domain.CodeLocation{}

	if finding.Filename != "" {
		location.File = finding.Filename
	}

	if finding.LineNo > 0 {
		location.StartLine = finding.LineNo
		location.EndLine = finding.LineNo
	}

	if finding.Code != "" {
		location.Snippet = finding.Code
	}

	return location
}

// getRemediationAdvice returns remediation advice for common Mythril findings.
func (m *MythrilScanner) getRemediationAdvice(finding MythrilIssue) string {
	swc := strings.ToUpper(strings.TrimSpace(finding.SWCID))

	adviceMap := map[string]string{
		"SWC-107": "Follow the checks-effects-interactions pattern. Update state before external calls. Consider using ReentrancyGuard.",
		"SWC-101": "Use SafeMath or Solidity 0.8+ checked arithmetic to prevent overflows and underflows.",
		"SWC-104": "Always check return values from low-level calls and revert on failure.",
		"SWC-115": "Avoid authorization based on tx.origin. Use msg.sender instead.",
		"SWC-116": "Avoid using block.timestamp for critical logic. Prefer block numbers or oracles.",
		"SWC-120": "Use secure randomness sources or VRF providers instead of block values.",
		"SWC-105": "Restrict selfdestruct and sensitive actions with access control.",
		"SWC-128": "Limit unbounded loops and external calls. Use pull patterns to avoid DoS.",
		"SWC-132": "Ensure proper storage layout for upgradeable contracts to prevent collisions.",
	}

	return adviceMap[swc]
}

// calculateRiskScore computes overall risk score from vulnerabilities.
func (m *MythrilScanner) calculateRiskScore(vulnerabilities []domain.Vulnerability) float64 {
	if len(vulnerabilities) == 0 {
		return 0.0
	}

	totalScore := 0.0
	weights := map[domain.ThreatLevel]float64{
		domain.ThreatLevelCritical: 10.0,
		domain.ThreatLevelHigh:     7.0,
		domain.ThreatLevelMedium:   4.0,
		domain.ThreatLevelLow:      1.0,
		domain.ThreatLevelInfo:     0.5,
	}

	for _, vuln := range vulnerabilities {
		weight := weights[vuln.Severity]
		totalScore += weight * vuln.Confidence
	}

	normalized := 10.0 * (1 - 1/(1 + totalScore/10))
	return normalized
}

// determineThreatLevel determines threat level from risk score.
func (m *MythrilScanner) determineThreatLevel(riskScore float64) domain.ThreatLevel {
	switch {
	case riskScore >= 8.0:
		return domain.ThreatLevelCritical
	case riskScore >= 6.0:
		return domain.ThreatLevelHigh
	case riskScore >= 3.0:
		return domain.ThreatLevelMedium
	case riskScore >= 1.0:
		return domain.ThreatLevelLow
	default:
		return domain.ThreatLevelNone
	}
}

// calculateMetrics computes scan metrics from vulnerabilities.
func (m *MythrilScanner) calculateMetrics(vulnerabilities []domain.Vulnerability, contract *domain.Contract, duration time.Duration) domain.ScanMetrics {
	metrics := domain.ScanMetrics{
		TotalIssues: len(vulnerabilities),
	}

	for _, vuln := range vulnerabilities {
		switch vuln.Severity {
		case domain.ThreatLevelCritical:
			metrics.CriticalCount++
		case domain.ThreatLevelHigh:
			metrics.HighCount++
		case domain.ThreatLevelMedium:
			metrics.MediumCount++
		case domain.ThreatLevelLow:
			metrics.LowCount++
		case domain.ThreatLevelInfo:
			metrics.InfoCount++
		}
	}

	return metrics
}

// getMythrilVersion returns the installed Mythril version.
func (m *MythrilScanner) getMythrilVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, m.mythrilPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}
