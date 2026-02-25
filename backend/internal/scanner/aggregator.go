package scanner

import (
	"fmt"

	"github.com/vigilum/backend/internal/domain"
)

// ScanAggregator combines results from multiple scanners with deduplication and weighted aggregation
type ScanAggregator struct {
	findings map[string]domain.Vulnerability // Key: hash(location + type)
	weights  map[string]float64               // Scanner name -> weight
}

// NewScanAggregator creates a new aggregator with default scanner weights
func NewScanAggregator() *ScanAggregator {
	return &ScanAggregator{
		findings: make(map[string]domain.Vulnerability),
		weights: map[string]float64{
			"slither":  0.40,  // Static analysis is mature and comprehensive
			"mythril":  0.40,  // Symbolic execution catches deep semantic issues
			"static":   0.15,  // Pattern matching (less reliable but fast)
			"ml":       0.05,  // ML as signal boost (not production-ready yet)
		},
	}
}

// AddFinding adds a vulnerability finding from a scanner
// If a duplicate finding exists, it merges confidence scores
func (sa *ScanAggregator) AddFinding(finding domain.Vulnerability) {
	key := sa.keyForFinding(finding)

	if existing, exists := sa.findings[key]; exists {
		// Finding already exists - merge results
		if finding.Severity > existing.Severity {
			// Use the more severe version
			sa.findings[key] = finding
		} else if finding.Severity == existing.Severity {
			// Same severity - average confidence
			existing.Confidence = (existing.Confidence + finding.Confidence) / 2
			sa.findings[key] = existing
		}
		// Note: Multiple detectors found this vulnerability, but we can't store
		// that in Vulnerability struct, so we just use the higher severity/confidence
	} else {
		// New finding - add it
		sa.findings[key] = finding
	}
}

// GetAggregatedFindings returns all deduplicated findings
func (sa *ScanAggregator) GetAggregatedFindings() []domain.Vulnerability {
	findings := make([]domain.Vulnerability, 0, len(sa.findings))
	for _, finding := range sa.findings {
		findings = append(findings, finding)
	}
	return findings
}

// CalculateAggregateRiskScore computes weighted risk score from all findings
// Uses same logarithmic algorithm as individual scanners for consistency (0-10 scale)
func (sa *ScanAggregator) CalculateAggregateRiskScore() float64 {
	if len(sa.findings) == 0 {
		return 0.0
	}

	totalScore := 0.0
	// Weights for each threat level
	weights := map[domain.ThreatLevel]float64{
		domain.ThreatLevelCritical: 10.0,
		domain.ThreatLevelHigh:     7.0,
		domain.ThreatLevelMedium:   4.0,
		domain.ThreatLevelLow:      1.0,
		domain.ThreatLevelInfo:     0.5,
		domain.ThreatLevelNone:     0.0,
	}

	// Accumulate weighted severity
	for _, finding := range sa.findings {
		weight := weights[finding.Severity]
		// Factor in confidence - lower confidence reduces impact
		totalScore += weight * finding.Confidence
	}

	// Normalize to 0-10 scale using same algorithm as Slither/Mythril
	// Formula: 10 * (1 - 1/(1 + score/10))
	// This gives non-linear growth as score increases
	normalized := 10.0 * (1 - 1/(1+totalScore/10))
	if normalized > 10.0 {
		normalized = 10.0
	}
	return normalized
}

// DetermineThreatLevel maps risk score to threat level
func (sa *ScanAggregator) DetermineThreatLevel(riskScore float64) domain.ThreatLevel {
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

// CalculateMetrics combines all findings and generates summary metrics
func (sa *ScanAggregator) CalculateMetrics() domain.ScanMetrics {
	metrics := domain.ScanMetrics{
		TotalIssues: len(sa.findings),
	}

	for _, finding := range sa.findings {
		switch finding.Severity {
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

// keyForFinding creates a deduplication key for a finding
// Key is: file:line:type (ignores detector name to deduplicate across scanners)
func (sa *ScanAggregator) keyForFinding(finding domain.Vulnerability) string {
	return fmt.Sprintf("%s:%d:%s",
		finding.Location.File,
		finding.Location.StartLine,
		finding.Type,
	)
}
