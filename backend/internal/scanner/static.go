// Package scanner provides static analysis capabilities.
package scanner

import (
	"context"
	"encoding/hex"
	"log/slog"
	"strings"

	"github.com/vigilum/backend/internal/domain"
)

// StaticAnalyzer performs static code analysis on bytecode and source.
type StaticAnalyzer struct {
	logger          *slog.Logger
	patternDetector *PatternDetector
}

// NewStaticAnalyzer creates a new static analyzer.
func NewStaticAnalyzer(logger *slog.Logger) *StaticAnalyzer {
	return &StaticAnalyzer{
		logger:          logger.With("scanner", "static"),
		patternDetector: NewPatternDetector(),
	}
}

// Name returns the scanner identifier.
func (s *StaticAnalyzer) Name() string {
	return "static-analyzer"
}

// ScanType returns the scan methodology.
func (s *StaticAnalyzer) ScanType() domain.ScanType {
	return domain.ScanTypeStatic
}

// SupportedChains returns all EVM-compatible chains.
func (s *StaticAnalyzer) SupportedChains() []domain.ChainID {
	return []domain.ChainID{1, 137, 56, 42161, 10, 8453} // ETH, Polygon, BSC, Arbitrum, Optimism, Base
}

// IsHealthy checks if the scanner is operational.
func (s *StaticAnalyzer) IsHealthy(ctx context.Context) bool {
	return true
}

// Scan performs static analysis on the contract.
func (s *StaticAnalyzer) Scan(ctx context.Context, contract *domain.Contract) (*ScanResult, error) {
	s.logger.Info("Starting static analysis",
		"contract_id", contract.ID,
		"address", contract.Address,
		"chain_id", contract.ChainID,
	)

	result := &ScanResult{
		Vulnerabilities: make([]domain.Vulnerability, 0),
		Metrics: domain.ScanMetrics{
			TotalIssues: 0,
		},
		Metadata: make(map[string]any),
	}

	// Run analysis passes
	vulns := make([]domain.Vulnerability, 0)

	// Check for known vulnerability patterns
	if contract.Bytecode != nil {
		bytecodeVulns := s.analyzeBytecode(ctx, contract)
		vulns = append(vulns, bytecodeVulns...)
	}

	// Check source code if available
	if contract.SourceCode != "" {
		sourceVulns := s.analyzeSource(ctx, contract)
		vulns = append(vulns, sourceVulns...)
	}

	result.Vulnerabilities = vulns
	result.RiskScore = s.calculateRiskScore(vulns)
	result.ThreatLevel = s.determineThreatLevel(result.RiskScore)
	result.Metrics = s.calculateMetrics(vulns)

	s.logger.Info("Static analysis complete",
		"contract_id", contract.ID,
		"vulnerabilities", len(vulns),
		"risk_score", result.RiskScore,
	)

	return result, nil
}

// analyzeBytecode checks bytecode for known vulnerability patterns.
func (s *StaticAnalyzer) analyzeBytecode(ctx context.Context, contract *domain.Contract) []domain.Vulnerability {
	vulns := make([]domain.Vulnerability, 0)
	
	if len(contract.Bytecode) == 0 {
		return vulns
	}

	bytecodeHex := hex.EncodeToString(contract.Bytecode)

	// ═══════════════════════════════════════════════════════════════════════════
	// DELEGATECALL Detection
	// ═══════════════════════════════════════════════════════════════════════════
	// DELEGATECALL opcode = 0xf4
	if strings.Contains(bytecodeHex, "f4") {
		vulns = append(vulns, domain.Vulnerability{
			Type:        domain.VulnLogicError,
			Severity:    domain.ThreatLevelInfo,
			Confidence:  0.5,
			Title:       "DELEGATECALL Detected",
			Description: "Contract uses delegatecall - verify implementation is secure",
			Location:    domain.CodeLocation{File: "bytecode"},
			Remediation: "Ensure delegatecall targets are trusted and cannot be manipulated",
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// SELFDESTRUCT Detection
	// ═══════════════════════════════════════════════════════════════════════════
	// SELFDESTRUCT opcode = 0xff
	if strings.Contains(bytecodeHex, "ff") && s.looksLikeSelfDestruct(bytecodeHex) {
		vulns = append(vulns, domain.Vulnerability{
			Type:        domain.VulnAccessControl,
			Severity:    domain.ThreatLevelHigh,
			Confidence:  0.6,
			Title:       "SELFDESTRUCT Detected",
			Description: "Contract can be destroyed - verify access control",
			Location:    domain.CodeLocation{File: "bytecode"},
			Remediation: "Ensure selfdestruct is properly protected or consider removing it",
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// CREATE2 Detection (proxy/factory patterns)
	// ═══════════════════════════════════════════════════════════════════════════
	// CREATE2 opcode = 0xf5
	if strings.Contains(bytecodeHex, "f5") {
		vulns = append(vulns, domain.Vulnerability{
			Type:        domain.VulnLogicError,
			Severity:    domain.ThreatLevelInfo,
			Confidence:  0.5,
			Title:       "CREATE2 Detected",
			Description: "Contract uses CREATE2 - may be a factory or proxy",
			Location:    domain.CodeLocation{File: "bytecode"},
			Remediation: "Review CREATE2 usage for address collision risks",
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// Proxy Pattern Detection
	// ═══════════════════════════════════════════════════════════════════════════
	if s.looksLikeProxy(contract.Bytecode) {
		vulns = append(vulns, domain.Vulnerability{
			Type:        domain.VulnLogicError,
			Severity:    domain.ThreatLevelInfo,
			Confidence:  0.8,
			Title:       "Proxy Contract Detected",
			Description: "This appears to be a proxy contract - analyze implementation",
			Location:    domain.CodeLocation{File: "bytecode"},
			Remediation: "Verify implementation contract and upgrade mechanism",
		})
	}

	return vulns
}

// looksLikeSelfDestruct checks if FF is likely SELFDESTRUCT vs other uses.
func (s *StaticAnalyzer) looksLikeSelfDestruct(bytecode string) bool {
	// SELFDESTRUCT typically follows an address push
	// PUSH20 (0x73) + address + SELFDESTRUCT (0xff)
	return strings.Contains(bytecode, "73") && strings.Contains(bytecode, "ff")
}

// looksLikeProxy detects proxy contract patterns.
func (s *StaticAnalyzer) looksLikeProxy(bytecode []byte) bool {
	if len(bytecode) < 50 {
		return false
	}
	
	bytecodeHex := hex.EncodeToString(bytecode)
	
	// ERC1967 implementation slot: 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc
	if strings.Contains(bytecodeHex, "360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc") {
		return true
	}
	
	// Minimal proxy (EIP-1167) signature
	if strings.HasPrefix(bytecodeHex, "363d3d373d3d3d363d73") {
		return true
	}
	
	// Short contract that mainly does delegatecall
	if len(bytecode) < 100 && strings.Contains(bytecodeHex, "f4") {
		return true
	}
	
	return false
}

// analyzeSource checks source code for vulnerability patterns.
func (s *StaticAnalyzer) analyzeSource(ctx context.Context, contract *domain.Contract) []domain.Vulnerability {
	if contract.SourceCode == "" {
		return make([]domain.Vulnerability, 0)
	}

	// Use pattern detector for source analysis
	return s.patternDetector.DetectFromSource(ctx, contract.SourceCode)
}

// calculateRiskScore computes overall risk from vulnerabilities.
func (s *StaticAnalyzer) calculateRiskScore(vulns []domain.Vulnerability) float64 {
	if len(vulns) == 0 {
		return 0.0
	}

	weights := map[domain.ThreatLevel]float64{
		domain.ThreatLevelCritical: 10.0,
		domain.ThreatLevelHigh:     7.0,
		domain.ThreatLevelMedium:   4.0,
		domain.ThreatLevelLow:      1.0,
		domain.ThreatLevelInfo:     0.1,
	}

	var totalWeight float64
	for _, v := range vulns {
		totalWeight += weights[v.Severity] * v.Confidence
	}

	// Normalize to 0-100 scale
	score := totalWeight * 10
	if score > 100 {
		score = 100
	}
	return score
}

// determineThreatLevel maps risk score to threat level.
func (s *StaticAnalyzer) determineThreatLevel(score float64) domain.ThreatLevel {
	switch {
	case score >= 80:
		return domain.ThreatLevelCritical
	case score >= 60:
		return domain.ThreatLevelHigh
	case score >= 40:
		return domain.ThreatLevelMedium
	case score >= 20:
		return domain.ThreatLevelLow
	case score > 0:
		return domain.ThreatLevelInfo
	default:
		return domain.ThreatLevelNone
	}
}

// calculateMetrics aggregates vulnerability counts.
func (s *StaticAnalyzer) calculateMetrics(vulns []domain.Vulnerability) domain.ScanMetrics {
	metrics := domain.ScanMetrics{
		TotalIssues: len(vulns),
	}

	for _, v := range vulns {
		switch v.Severity {
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
