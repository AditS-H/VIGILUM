# VIGILUM Phase 14: Composite Scanner - Implementation Guide

**Objective:** Implement `Orchestrator.ScanAll()` to run Slither + Mythril + Static + ML scanners in parallel

**Status:** Slither ✅ + Mythril ✅ + Static ⚠️ + ML 🔴 → Need orchestrator

---

## Current Code Structure

### What Already Exists

**Scanner Interface** (`backend/internal/scanner/scanner.go`):
```go
type Scanner interface {
    Name() string
    ScanType() domain.ScanType
    Scan(ctx context.Context, contract *domain.Contract) (*ScanResult, error)
    SupportedChains() []domain.ChainID
    IsHealthy(ctx context.Context) bool
}

type ScanResult struct {
    Vulnerabilities []domain.Vulnerability
    RiskScore       float64
    ThreatLevel     domain.ThreatLevel
    Metrics         domain.ScanMetrics
    RawOutput       []byte
    Metadata        map[string]any
}

type Orchestrator struct {
    scanners []Scanner
}

func NewOrchestrator(scanners ...Scanner) *Orchestrator {
    return &Orchestrator{scanners: scanners}
}

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

func (o *Orchestrator) Register(scanner Scanner) {
    o.scanners = append(o.scanners, scanner)
}
```

### What's Implemented

✅ **SlitherScanner** (`slither.go` - 538 LOC)
- Compiles and tests pass
- Returns `ScanResult` with vulnerabilities

✅ **MythrilScanner** (`mythril.go` - 450 LOC)
- Compiles and tests pass
- Returns `ScanResult` with vulnerabilities

⚠️ **StaticScanner** (`static.go` - exists but minimal)
- Basic pattern matching only
- Needs enhancement

---

## Implementation Steps

### Step 1: Create the Aggregator

**File:** `backend/internal/scanner/aggregator.go` (NEW)

```go
package scanner

import (
	"fmt"
	"github.com/vigilum/backend/internal/domain"
)

// ScanAggregator combines results from multiple scanners
type ScanAggregator struct {
	findings map[string]domain.Vulnerability // Key: hash(location + type)
	weights map[string]float64 // Scanner name -> weight
}

// NewScanAggregator creates a new aggregator with default weights
func NewScanAggregator() *ScanAggregator {
	return &ScanAggregator{
		findings: make(map[string]domain.Vulnerability),
		weights: map[string]float64{
			"slither":  0.40,  // Static analysis is mature
			"mythril":  0.40,  // Symbolic execution catches deep bugs
			"static":   0.15,  // Pattern matching (less reliable)
			"ml":       0.05,  // ML as signal boost (not production-ready)
		},
	}
}

// AddFinding adds a vulnerability finding from a scanner
func (sa *ScanAggregator) AddFinding(finding domain.Vulnerability) {
	key := sa.keyForFinding(finding)
	
	if existing, exists := sa.findings[key]; exists {
		// Merge: take highest severity and combine confidence
		if finding.Severity > existing.Severity {
			sa.findings[key] = finding
		} else if finding.Severity == existing.Severity {
			// Average confidence
			existing.Confidence = (existing.Confidence + finding.Confidence) / 2
			sa.findings[key] = existing
		}
		// Update metadata to show multiple detectors
		if existing.Metadata == nil {
			existing.Metadata = make(map[string]any)
		}
		existing.Metadata["detectors"] = append(
			existing.Metadata["detectors"].([]string),
			finding.DetectedBy,
		)
	} else {
		finding.Metadata = map[string]any{
			"detectors": []string{finding.DetectedBy},
		}
		sa.findings[key] = finding
	}
}

// GetAggregatedFindings returns deduplicated findings
func (sa *ScanAggregator) GetAggregatedFindings() []domain.Vulnerability {
	findings := make([]domain.Vulnerability, 0, len(sa.findings))
	for _, finding := range sa.findings {
		findings = append(findings, finding)
	}
	return findings
}

// CalculateAggregateRiskScore computes weighted risk from all findings
func (sa *ScanAggregator) CalculateAggregateRiskScore() float64 {
	if len(sa.findings) == 0 {
		return 0.0
	}

	totalScore := 0.0
	weights := map[domain.ThreatLevel]float64{
		domain.ThreatLevelCritical: 10.0,
		domain.ThreatLevelHigh:     7.0,
		domain.ThreatLevelMedium:   4.0,
		domain.ThreatLevelLow:      1.0,
		domain.ThreatLevelInfo:     0.5,
		domain.ThreatLevelNone:     0.0,
	}

	for _, finding := range sa.findings {
		weight := weights[finding.Severity]
		// Factor in confidence
		totalScore += weight * finding.Confidence
	}

	// Normalize to 0-10 scale (same algorithm as individual scanners)
	normalized := 10.0 * (1 - 1/(1+totalScore/10))
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

// CalculateMetrics combines metrics from all findings
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
func (sa *ScanAggregator) keyForFinding(finding domain.Vulnerability) string {
	// Key: file:line + type (ignores detector name)
	return fmt.Sprintf("%s:%d:%s",
		finding.Location.File,
		finding.Location.StartLine,
		finding.Type,
	)
}
```

---

### Step 2: Update Orchestrator.ScanAll()

**File:** `backend/internal/scanner/scanner.go` (MODIFY)

Replace the `ScanAll` method:

```go
import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
	"github.com/vigilum/backend/internal/domain"
)

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

	// Create aggregator for results
	aggregator := NewScanAggregator()

	// Run scanners concurrently with errgroup
	g, groupCtx := errgroup.WithContext(scanCtx)

	// Run each scanner in a goroutine
	for _, scanner := range o.scanners {
		scanner := scanner // Capture for closure
		
		g.Go(func() error {
			result, err := scanner.Scan(groupCtx, contract)
			if err != nil {
				// Log error but don't fail entire scan
				slog.Warn("Scanner failed",
					"scanner", scanner.Name(),
					"contract", contract.ID,
					"error", err,
				)
				return nil // Continue even if one scanner fails
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
	report.CompletedAt = &time.Now()
	report.Duration = duration

	return report, nil
}
```

---

### Step 3: Update Imports

**File:** `backend/internal/scanner/scanner.go` (ADD IMPORT)

```go
import (
	"golang.org/x/sync/errgroup"
)
```

**Install package if needed:**
```bash
cd backend
go get golang.org/x/sync/errgroup
```

---

### Step 4: Write Tests

**File:** `backend/internal/scanner/orchestrator_test.go` (NEW)

```go
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

func TestOrchestrator_ScanAll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slither, err := NewSlitherScanner(logger, nil)
	require.NoError(t, err)

	mythril, err := NewMythrilScanner(logger, nil)
	require.NoError(t, err)

	orchestrator := NewOrchestrator(slither, mythril)

	contract := &domain.Contract{
		ID:      "test-contract",
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
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, domain.ScanStatusCompleted, report.Status)
	assert.GreaterOrEqual(t, report.RiskScore, 0.0)
	assert.LessOrEqual(t, report.RiskScore, 10.0)
}

func TestScanAggregator_Deduplicate(t *testing.T) {
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
	assert.Len(t, findings, 1, "Should deduplicate to 1 finding")
	assert.Equal(t, 0.85, findings[0].Confidence, "Should average confidence")
}

func TestScanAggregator_RiskScore(t *testing.T) {
	aggregator := NewScanAggregator()

	aggregator.AddFinding(domain.Vulnerability{
		Severity:   domain.ThreatLevelCritical,
		Confidence: 0.9,
		Location:   domain.CodeLocation{File: "a.sol", StartLine: 1},
		Type:       domain.VulnReentrancy,
		DetectedBy: "test",
	})

	score := aggregator.CalculateAggregateRiskScore()
	assert.Greater(t, score, 4.0, "Critical finding should give score >4")
	assert.Less(t, score, 10.0, "Score should be normalized to <10")
}

func TestScanAggregator_ThreatLevel(t *testing.T) {
	aggregator := NewScanAggregator()

	testCases := []struct {
		score         float64
		expectedLevel domain.ThreatLevel
	}{
		{0.5, domain.ThreatLevelNone},
		{1.5, domain.ThreatLevelLow},
		{3.5, domain.ThreatLevelMedium},
		{6.5, domain.ThreatLevelHigh},
		{8.5, domain.ThreatLevelCritical},
	}

	for _, tc := range testCases {
		level := aggregator.DetermineThreatLevel(tc.score)
		assert.Equal(t, tc.expectedLevel, level)
	}
}

func TestOrchestrator_Timeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slither, _ := NewSlitherScanner(logger, nil)
	orchestrator := NewOrchestrator(slither)

	contract := &domain.Contract{
		ID:      "test",
		Address: "0xABCD",
		ChainID: 1,
		SourceCode: "pragma solidity ^0.8.0; contract Test {} ",
	}

	opts := &ScanOptions{
		Timeout: 1, // 1 second timeout
	}

	report, err := orchestrator.ScanAll(context.Background(), contract, opts)
	// Should either timeout or complete (depending on system speed)
	assert.NotNil(t, report)
	if err != nil {
		assert.Contains(t, err.Error(), "timeout")
	}
}
```

---

### Step 5: Run Tests

```bash
cd backend

# Run orchestrator tests
go test ./internal/scanner -v -run TestOrchestrator

# Run aggregator tests
go test ./internal/scanner -v -run TestScanAggregator

# Run all scanner tests
go test ./internal/scanner -v

# Check coverage
go test ./internal/scanner -cover
```

---

## Expected Output

When running the tests, you should see:

```
=== RUN   TestOrchestrator_ScanAll
=== RUN   TestScanAggregator_Deduplicate
=== RUN   TestScanAggregator_RiskScore
=== RUN   TestScanAggregator_ThreatLevel
=== RUN   TestOrchestrator_Timeout
--- PASS: TestOrchestrator_ScanAll
--- PASS: TestScanAggregator_Deduplicate
--- PASS: TestScanAggregator_RiskScore
--- PASS: TestScanAggregator_ThreatLevel
--- PASS: TestOrchestrator_Timeout

PASS
ok      github.com/vigilum/backend/internal/scanner     X.XXXs
```

---

## Integration with API

Once the orchestrator is working, integrate it into the API:

**File:** `backend/internal/api/handlers/scanner_handler.go` (NEW or UPDATE)

```go
func (h *ScannerHandler) ScanContract(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	contract := &domain.Contract{
		ID:         req.ContractID,
		Address:    domain.Address(req.Address),
		ChainID:    req.ChainID,
		SourceCode: req.SourceCode,
		Bytecode:   req.Bytecode,
	}

	opts := &ScanOptions{
		Timeout:    300,
		MaxDepth:   50,
		EnableML:   true,
		IncludeInfo: false,
	}

	report, err := h.orchestrator.ScanAll(r.Context(), contract, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
```

---

## Verification Checklist

- [ ] `aggregator.go` created with 200+ LOC
- [ ] `scanner.go` Orchestrator.ScanAll() implemented
- [ ] All imports added
- [ ] Tests written and passing
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/scanner -v` shows all tests passing
- [ ] Risk score aggregation works correctly
- [ ] De-duplication working (confirmed in tests)
- [ ] Timeout handling works
- [ ] Error handling for individual scanner failures

---

## Success Criteria

✅ **Code Quality:**
- No compilation errors
- Tests passing: 100%
- Code coverage: >80%

✅ **Functionality:**
- Runs all scanners in parallel
- Deduplicates findings correctly
- Aggregates risk scores properly
- Handles timeouts gracefully
- Continues if one scanner fails

✅ **Performance:**
- Total scan time ≤ 30 seconds
- Memory usage reasonable
- No goroutine leaks

