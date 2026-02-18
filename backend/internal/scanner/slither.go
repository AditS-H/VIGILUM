// Package scanner provides Slither static analysis integration.
package scanner

import (
	"context"
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

// SlitherScanner performs static analysis using Slither.
type SlitherScanner struct {
	logger       *slog.Logger
	slitherPath  string // Path to slither executable
	workDir      string // Temporary directory for analysis
	timeout      time.Duration
	enabledChecks []string // Specific detectors to run
}

// SlitherConfig contains configuration for Slither scanner.
type SlitherConfig struct {
	SlitherPath   string
	WorkDir       string
	Timeout       time.Duration
	EnabledChecks []string
}

// DefaultSlitherConfig returns sensible defaults.
func DefaultSlitherConfig() *SlitherConfig {
	return &SlitherConfig{
		SlitherPath:   "slither", // Assumes slither is in PATH
		WorkDir:       "/tmp/vigilum-slither",
		Timeout:       5 * time.Minute,
		EnabledChecks: []string{}, // Empty = all detectors
	}
}

// NewSlitherScanner creates a new Slither scanner instance.
func NewSlitherScanner(logger *slog.Logger, config *SlitherConfig) (*SlitherScanner, error) {
	if config == nil {
		config = DefaultSlitherConfig()
	}

	// Create work directory if it doesn't exist
	if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Verify slither is available
	if _, err := exec.LookPath(config.SlitherPath); err != nil {
		logger.Warn("Slither not found in PATH", "path", config.SlitherPath, "error", err)
		// Don't fail - allow scanner to be created but mark as unhealthy
	}

	return &SlitherScanner{
		logger:       logger,
		slitherPath:  config.SlitherPath,
		workDir:      config.WorkDir,
		timeout:      config.Timeout,
		enabledChecks: config.EnabledChecks,
	}, nil
}

// Name returns the scanner identifier.
func (s *SlitherScanner) Name() string {
	return "slither"
}

// ScanType returns the type of analysis performed.
func (s *SlitherScanner) ScanType() domain.ScanType {
	return domain.ScanTypeStatic
}

// Scan performs static analysis on a contract using Slither.
func (s *SlitherScanner) Scan(ctx context.Context, contract *domain.Contract) (*ScanResult, error) {
	startTime := time.Now()
	s.logger.Info("Starting Slither scan",
		"contract_id", contract.ID,
		"address", contract.Address,
		"chain_id", contract.ChainID,
	)

	// Create temporary Solidity file
	contractFile, err := s.prepareContractFile(contract)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare contract file: %w", err)
	}
	defer os.Remove(contractFile)

	// Run Slither
	slitherOutput, err := s.runSlither(ctx, contractFile)
	if err != nil {
		s.logger.Error("Slither execution failed", "error", err)
		return nil, fmt.Errorf("slither execution failed: %w", err)
	}

	// Parse Slither JSON output
	findings, err := s.parseSlitherOutput(slitherOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse slither output: %w", err)
	}

	// Convert to domain vulnerabilities
	vulnerabilities := s.mapFindingsToVulnerabilities(findings, contract)

	// Calculate risk score and metrics
	duration := time.Since(startTime)
	riskScore := s.calculateRiskScore(vulnerabilities)
	threatLevel := s.determineThreatLevel(riskScore)
	metrics := s.calculateMetrics(vulnerabilities, contract, duration)

	s.logger.Info("Slither scan completed",
		"contract_id", contract.ID,
		"vulnerabilities", len(vulnerabilities),
		"risk_score", riskScore,
		"threat_level", threatLevel,
		"duration", duration,
	)

	return &ScanResult{
		Vulnerabilities: vulnerabilities,
		RiskScore:       riskScore,
		ThreatLevel:     s.determineThreatLevel(riskScore),
		Metrics:         metrics,
		RawOutput:       slitherOutput,
		Metadata: map[string]any{
			"scanner":        "slither",
			"version":        s.getSlitherVersion(),
			"detectors_used": s.enabledChecks,
			"analysis_time":  duration.Seconds(),
		},
	}, nil
}

// SupportedChains returns chains this scanner supports.
func (s *SlitherScanner) SupportedChains() []domain.ChainID {
	// Slither works on Solidity source code, chain-agnostic
	return []domain.ChainID{
		1,     // Ethereum Mainnet
		5,     // Goerli
		11155111, // Sepolia
		137,   // Polygon
		56,    // BSC
		42161, // Arbitrum One
		10,    // Optimism
		8453,  // Base
	}
}

// IsHealthy checks if Slither is operational.
func (s *SlitherScanner) IsHealthy(ctx context.Context) bool {
	// Check if slither executable exists
	if _, err := exec.LookPath(s.slitherPath); err != nil {
		s.logger.Warn("Slither health check failed: not found in PATH")
		return false
	}

	// Try to run slither --version
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.slitherPath, "--version")
	if err := cmd.Run(); err != nil {
		s.logger.Warn("Slither health check failed: version check", "error", err)
		return false
	}

	return true
}

// prepareContractFile creates a temporary Solidity file for analysis.
func (s *SlitherScanner) prepareContractFile(contract *domain.Contract) (string, error) {
	// Create unique filename
	filename := fmt.Sprintf("contract_%s_%d.sol", contract.Address, time.Now().Unix())
	filepath := filepath.Join(s.workDir, filename)

	// If we have source code, use it
	var content string
	if contract.SourceCode != "" {
		content = contract.SourceCode
	} else if len(contract.Bytecode) > 0 {
		// If only bytecode available, create a minimal wrapper
		// Note: Slither works best with source code
		content = s.generateMinimalSource(contract)
	} else {
		return "", fmt.Errorf("contract has no source code or bytecode")
	}

	// Write to file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write contract file: %w", err)
	}

	return filepath, nil
}

// generateMinimalSource creates minimal Solidity source for bytecode-only contracts.
func (s *SlitherScanner) generateMinimalSource(contract *domain.Contract) string {
	// This is a fallback - Slither needs source code for best results
	return fmt.Sprintf(`// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

// Contract at address: %s
// Deployed on chain: %d
// Note: This is a placeholder - Slither requires source code for accurate analysis
contract UnknownContract {
    // Bytecode size: %d bytes
    // Analysis may be limited without source code
    fallback() external payable {}
    receive() external payable {}
}
`, contract.Address, contract.ChainID, len(contract.Bytecode))
}

// runSlither executes Slither and returns JSON output.
func (s *SlitherScanner) runSlither(ctx context.Context, contractFile string) ([]byte, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Build slither command
	args := []string{
		contractFile,
		"--json", "-", // Output JSON to stdout
		"--solc-disable-warnings", // Reduce noise
	}

	// Add specific detectors if configured
	if len(s.enabledChecks) > 0 {
		detectors := strings.Join(s.enabledChecks, ",")
		args = append(args, "--detect", detectors)
	}

	// Run slither
	cmd := exec.CommandContext(ctx, s.slitherPath, args...)
	output, _ := cmd.CombinedOutput()

	// Slither returns non-zero exit code if vulnerabilities found
	// This is expected, so we only error on context timeout/cancel
	if ctx.Err() != nil {
		return nil, fmt.Errorf("slither execution timed out: %w", ctx.Err())
	}

	return output, nil
}

// SlitherOutput represents the JSON structure from Slither.
type SlitherOutput struct {
	Success bool              `json:"success"`
	Error   string            `json:"error,omitempty"`
	Results SlitherResults    `json:"results"`
}

type SlitherResults struct {
	Detectors []SlitherDetector `json:"detectors"`
}

type SlitherDetector struct {
	Check       string                 `json:"check"`
	Confidence  string                 `json:"confidence"`
	Impact      string                 `json:"impact"`
	Description string                 `json:"description"`
	Elements    []SlitherElement       `json:"elements"`
	FirstMarkdownElement string        `json:"first_markdown_element,omitempty"`
	AdditionalFields map[string]interface{} `json:"additional_fields,omitempty"`
}

type SlitherElement struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	SourceMapping SlitherSourceMapping `json:"source_mapping"`
}

type SlitherSourceMapping struct {
	Start      int    `json:"start"`
	Length     int    `json:"length"`
	Filename   string `json:"filename_relative"`
	Lines      []int  `json:"lines"`
}

// parseSlitherOutput parses Slither's JSON output.
func (s *SlitherScanner) parseSlitherOutput(data []byte) ([]SlitherDetector, error) {
	var output SlitherOutput
	if err := json.Unmarshal(data, &output); err != nil {
		// If JSON parsing fails, Slither might have output plain text
		s.logger.Warn("Failed to parse Slither JSON, checking for plain text output")
		return []SlitherDetector{}, nil
	}

	if !output.Success && output.Error != "" {
		return nil, fmt.Errorf("slither error: %s", output.Error)
	}

	return output.Results.Detectors, nil
}

// mapFindingsToVulnerabilities converts Slither findings to domain vulnerabilities.
func (s *SlitherScanner) mapFindingsToVulnerabilities(findings []SlitherDetector, contract *domain.Contract) []domain.Vulnerability {
	vulnerabilities := make([]domain.Vulnerability, 0, len(findings))

	for _, finding := range findings {
		vuln := domain.Vulnerability{
			ID:          fmt.Sprintf("slither-%s-%d", finding.Check, time.Now().Unix()),
			ContractID:  contract.ID,
			Type:        s.mapSlitherCheckToVulnType(finding.Check),
			Severity:    s.mapSlitherImpactToSeverity(finding.Impact),
			Title:       s.generateTitle(finding),
			Description: finding.Description,
			Location:    s.extractCodeLocation(finding),
			Confidence:  s.mapSlitherConfidenceToFloat(finding.Confidence),
			DetectedBy:  "slither",
			DetectedAt:  time.Now(),
			IsConfirmed: false,
			IsFalsePos:  false,
		}

		// Add remediation advice if available
		if advice := s.getRemediationAdvice(finding.Check); advice != "" {
			vuln.Remediation = advice
		}

		vulnerabilities = append(vulnerabilities, vuln)
	}

	return vulnerabilities
}

// mapSlitherCheckToVulnType maps Slither detector names to domain vulnerability types.
func (s *SlitherScanner) mapSlitherCheckToVulnType(check string) domain.VulnType {
	// Map common Slither checks to domain VulnType constants
	checkMap := map[string]domain.VulnType{
		"reentrancy-eth":          domain.VulnReentrancy,
		"reentrancy-no-eth":       domain.VulnReentrancy,
		"reentrancy-benign":       domain.VulnReentrancy,
		"arbitrary-send":          domain.VulnAccessControl,
		"suicidal":                domain.VulnAccessControl,
		"unprotected-upgrade":     domain.VulnAccessControl,
		"tx-origin":               domain.VulnTxOrigin,
		"unchecked-lowlevel":      domain.VulnUncheckedCall,
		"unchecked-send":          domain.VulnUncheckedCall,
		"timestamp":               domain.VulnTimestamp,
		"weak-prng":               domain.VulnWeakRandomness,
		"divide-before-multiply":  domain.VulnPrecisionLoss,
		"incorrect-equality":      domain.VulnLogicError,
		"incorrect-shift":         domain.VulnLogicError,
	}

	if vulnType, ok := checkMap[check]; ok {
		return vulnType
	}

	// Default to logic error for unmapped checks
	return domain.VulnLogicError
}

// mapSlitherImpactToSeverity maps Slither impact levels to domain threat level.
func (s *SlitherScanner) mapSlitherImpactToSeverity(impact string) domain.ThreatLevel {
	switch strings.ToLower(impact) {
	case "high":
		return domain.ThreatLevelCritical
	case "medium":
		return domain.ThreatLevelHigh
	case "low":
		return domain.ThreatLevelMedium
	case "informational":
		return domain.ThreatLevelInfo
	default:
		return domain.ThreatLevelLow
	}
}

// mapSlitherConfidenceToFloat converts Slither confidence to 0-1 scale.
func (s *SlitherScanner) mapSlitherConfidenceToFloat(confidence string) float64 {
	switch strings.ToLower(confidence) {
	case "high":
		return 0.9
	case "medium":
		return 0.7
	case "low":
		return 0.5
	default:
		return 0.6
	}
}

// generateTitle creates a human-readable title from Slither finding.
func (s *SlitherScanner) generateTitle(finding SlitherDetector) string {
	// Clean up the check name for display
	title := strings.ReplaceAll(finding.Check, "-", " ")
	title = strings.Title(title)
	
	// Extract function name if available
	if len(finding.Elements) > 0 {
		for _, elem := range finding.Elements {
			if elem.Type == "function" && elem.Name != "" {
				title += fmt.Sprintf(" in %s()", elem.Name)
				break
			}
		}
	}

	return title
}

// extractCodeLocation extracts source code location from Slither finding.
func (s *SlitherScanner) extractCodeLocation(finding SlitherDetector) domain.CodeLocation {
	location := domain.CodeLocation{}

	if len(finding.Elements) == 0 {
		return location
	}

	// Get the first element with source mapping
	elem := finding.Elements[0]
	mapping := elem.SourceMapping

	location.File = mapping.Filename

	if len(mapping.Lines) > 0 {
		location.StartLine = mapping.Lines[0]
		location.EndLine = mapping.Lines[len(mapping.Lines)-1]
	}

	// Add snippet if description contains code
	if finding.FirstMarkdownElement != "" {
		location.Snippet = finding.FirstMarkdownElement
	}

	return location
}

// getRemediationAdvice returns remediation advice for common vulnerabilities.
func (s *SlitherScanner) getRemediationAdvice(check string) string {
	adviceMap := map[string]string{
		"reentrancy-eth": "Follow the checks-effects-interactions pattern. Update state before external calls. Consider using ReentrancyGuard from OpenZeppelin.",
		"reentrancy-no-eth": "Update state variables before making external calls to prevent reentrancy attacks.",
		"controlled-delegatecall": "Avoid delegatecall with user-controlled addresses. Use a whitelist of approved targets.",
		"arbitrary-send": "Implement access control to restrict who can trigger token transfers.",
		"suicidal": "Add access control to selfdestruct. Consider removing it entirely if not needed.",
		"unprotected-upgrade": "Add proper access control to upgrade functions. Use OpenZeppelin's UpgradeableProxy pattern.",
		"tx-origin": "Replace tx.origin with msg.sender for authorization checks.",
		"unchecked-lowlevel": "Always check the return value of low-level calls (call, delegatecall, staticcall).",
		"timestamp": "Avoid using block.timestamp for critical logic. Use block numbers or external oracles.",
	}

	return adviceMap[check]
}

// calculateRiskScore computes overall risk score from vulnerabilities.
func (s *SlitherScanner) calculateRiskScore(vulnerabilities []domain.Vulnerability) float64 {
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
		// Factor in confidence
		totalScore += weight * vuln.Confidence
	}

	// Normalize to 0-10 scale (diminishing returns for many vulnerabilities)
	// Using logarithmic scale to prevent score explosion
	normalized := 10.0 * (1 - 1/(1 + totalScore/10))
	
	return normalized
}

// determineThreatLevel determines threat level from risk score.
func (s *SlitherScanner) determineThreatLevel(riskScore float64) domain.ThreatLevel {
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
func (s *SlitherScanner) calculateMetrics(vulnerabilities []domain.Vulnerability, contract *domain.Contract, duration time.Duration) domain.ScanMetrics {
	metrics := domain.ScanMetrics{
		TotalIssues: len(vulnerabilities),
	}

	// Count by severity
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

// getSlitherVersion returns the installed Slither version.
func (s *SlitherScanner) getSlitherVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.slitherPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}
