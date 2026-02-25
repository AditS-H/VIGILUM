// Package scanner provides contract security analysis capabilities.
package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
	"github.com/vigilum/backend/internal/domain"
)

// Scanner is the interface for security scanners.
type Scanner interface {
	// Name returns the scanner identifier.
	Name() string
	
	// ScanType returns the type of analysis performed.
	ScanType() domain.ScanType
	
	// Scan performs security analysis on a contract.
	Scan(ctx context.Context, contract *domain.Contract) (*ScanResult, error)
	
	// SupportedChains returns chains this scanner supports.
	SupportedChains() []domain.ChainID
	
	// IsHealthy checks if the scanner is operational.
	IsHealthy(ctx context.Context) bool
}

// ScanResult contains the output of a security scan.
type ScanResult struct {
	Vulnerabilities []domain.Vulnerability
	RiskScore       float64
	ThreatLevel     domain.ThreatLevel
	Metrics         domain.ScanMetrics
	RawOutput       []byte // Scanner-specific output
	Metadata        map[string]any
}

// ScanOptions configures scan behavior.
type ScanOptions struct {
	Timeout      int  // Seconds
	MaxDepth     int  // For symbolic execution
	FuzzRounds   int  // For fuzzing
	EnableML     bool // Use ML models
	IncludeInfo  bool // Include informational findings
}

// DefaultScanOptions returns sensible defaults.
func DefaultScanOptions() *ScanOptions {
	return &ScanOptions{
		Timeout:     300,
		MaxDepth:    50,
		FuzzRounds:  1000,
		EnableML:    true,
		IncludeInfo: false,
	}
}

// Orchestrator coordinates multiple scanners.
type Orchestrator struct {
	scanners []Scanner
}

// NewOrchestrator creates a new scanner orchestrator.
func NewOrchestrator(scanners ...Scanner) *Orchestrator {
	return &Orchestrator{
		scanners: scanners,
	}
}

// ScanAll runs all applicable scanners on a contract in parallel.
// Findings are deduplicated and aggregated using weighted scoring.
// Individual scanner failures do not block the overall scan.
func (o *Orchestrator) ScanAll(ctx context.Context, contract *domain.Contract, opts *ScanOptions) (*domain.ScanReport, error) {
	if opts == nil {
		opts = DefaultScanOptions()
	}

	startTime := time.Now()

	report := &domain.ScanReport{
		ID:              fmt.Sprintf("scan_%d_%s", time.Now().Unix(), contract.ID),
		ContractID:      contract.ID,
		ScanType:        domain.ScanTypeFull,
		Status:          domain.ScanStatusRunning,
		Vulnerabilities: make([]domain.Vulnerability, 0),
		StartedAt:       startTime,
	}

	// Create a timeout context if specified
	scanCtx := ctx
	var cancel context.CancelFunc
	if opts != nil && opts.Timeout > 0 {
		scanCtx, cancel = context.WithTimeout(ctx, time.Duration(opts.Timeout)*time.Second)
		defer cancel()
	}

	// Create aggregator for combining results
	aggregator := NewScanAggregator()

	// Run scanners concurrently with errgroup
	g, groupCtx := errgroup.WithContext(scanCtx)

	// Run each scanner in a goroutine
	for _, scanner := range o.scanners {
		scanner := scanner // Capture for closure

		g.Go(func() error {
			result, err := scanner.Scan(groupCtx, contract)
			if err != nil {
				// Log error but don't fail entire scan - continue with other scanners
				slog.Warn("Scanner failed",
					"scanner", scanner.Name(),
					"contract", contract.ID,
					"error", err,
				)
				return nil // Don't return error - let other scanners finish
			}

			if result == nil {
				return nil
			}

			// Add all findings from this scanner to aggregator
			for _, vuln := range result.Vulnerabilities {
				aggregator.AddFinding(vuln)
			}

			return nil
		})
	}

	// Wait for all scanners to complete
	if err := g.Wait(); err != nil {
		report.Status = domain.ScanStatusFailed
		report.Error = fmt.Sprintf("scan error: %v", err)
		return report, err
	}

	// Get aggregated results
	report.Vulnerabilities = aggregator.GetAggregatedFindings()
	report.RiskScore = aggregator.CalculateAggregateRiskScore()
	report.ThreatLevel = aggregator.DetermineThreatLevel(report.RiskScore)
	report.Metrics = aggregator.CalculateMetrics()

	// Update status and timing
	duration := time.Since(startTime)
	report.Status = domain.ScanStatusCompleted
	now := time.Now()
	report.CompletedAt = &now
	report.Duration = duration

	return report, nil
}

// Register adds a scanner to the orchestrator.
func (o *Orchestrator) Register(scanner Scanner) {
	o.scanners = append(o.scanners, scanner)
}
