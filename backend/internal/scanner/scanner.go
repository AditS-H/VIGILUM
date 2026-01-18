// Package scanner provides contract security analysis capabilities.
package scanner

import (
	"context"

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

// ScanAll runs all applicable scanners on a contract.
func (o *Orchestrator) ScanAll(ctx context.Context, contract *domain.Contract, opts *ScanOptions) (*domain.ScanReport, error) {
	if opts == nil {
		opts = DefaultScanOptions()
	}

	report := &domain.ScanReport{
		ContractID:      contract.ID,
		ScanType:        domain.ScanTypeFull,
		Status:          domain.ScanStatusRunning,
		Vulnerabilities: make([]domain.Vulnerability, 0),
	}

	// TODO: Run scanners concurrently with errgroup
	// TODO: Deduplicate findings
	// TODO: Calculate aggregate risk score
	// TODO: Determine overall threat level

	return report, nil
}

// Register adds a scanner to the orchestrator.
func (o *Orchestrator) Register(scanner Scanner) {
	o.scanners = append(o.scanners, scanner)
}
