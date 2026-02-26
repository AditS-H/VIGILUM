package temporal

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"github.com/vigilum/backend/internal/domain"
)

// ScanContractWorkflow defines the contract scanning workflow
type ScanContractWorkflow struct {
	contractID string
	address    string
	chainID    int64
}

// ScanContractInput defines workflow input parameters
type ScanContractInput struct {
	ContractID  string
	Address     string
	ChainID     int64
	SourceCode  string
	Bytecode    []byte
	EnableML    bool
	Timeout     int
}

// ScanContractOutput defines workflow output
type ScanContractOutput struct {
	ScanReportID    string
	RiskScore       float64
	ThreatLevel     string
	VulnerabilityCount int
	Metrics         map[string]interface{}
	CompletedAt     time.Time
}

// ExecuteScanContractWorkflow is the main scanning workflow
func ExecuteScanContractWorkflow(ctx workflow.Context, input ScanContractInput) (*ScanContractOutput, error) {
	// Workflow options
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Duration(input.Timeout) * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, opts)

	// Step 1: Fetch contract from storage
	var contractData *domain.Contract
	err := workflow.ExecuteActivity(ctx, FetchContractActivity, input.ContractID).Get(ctx, &contractData)
	if err != nil {
		return nil, fmt.Errorf("fetch contract failed: %w", err)
	}

	// Step 2: Extract features for ML (parallel)
	var mlFeatures []float64
	mlErr := workflow.ExecuteActivity(ctx, ExtractFeaturesActivity, contractData).Get(ctx, &mlFeatures)
	// ML errors are non-fatal

	// Step 3: Run multi-engine scan
	var scanReport *domain.ScanReport
	scanInput := ScanInput{
		Contract:    contractData,
		EnableML:    input.EnableML && mlErr == nil,
		MLFeatures:  mlFeatures,
	}
	err = workflow.ExecuteActivity(ctx, PerformScanActivity, scanInput).Get(ctx, &scanReport)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Step 4: Store results
	var storeID string
	err = workflow.ExecuteActivity(ctx, StoreScanResultActivity, scanReport).Get(ctx, &storeID)
	if err != nil {
		// Log but don't fail workflow - scan happened but storage failed
		workflow.GetLogger(ctx).Error("storage failed", "error", err)
	}

	// Step 5: Notify or trigger further actions
	err = workflow.ExecuteActivity(ctx, NotifyResultsActivity, scanReport).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Warn("notification failed", "error", err)
	}

	// Return results
	output := &ScanContractOutput{
		ScanReportID:       scanReport.ID,
		RiskScore:          scanReport.RiskScore,
		ThreatLevel:        string(scanReport.ThreatLevel),
		VulnerabilityCount: len(scanReport.Vulnerabilities),
		Metrics: map[string]interface{}{
			"total_issues":  scanReport.Metrics.TotalIssues,
			"critical":      scanReport.Metrics.CriticalCount,
			"high":          scanReport.Metrics.HighCount,
			"medium":        scanReport.Metrics.MediumCount,
			"low":           scanReport.Metrics.LowCount,
			"info":          scanReport.Metrics.InfoCount,
		},
		CompletedAt: time.Now(),
	}

	return output, nil
}

// BatchScanWorkflow scans multiple contracts in parallel
func BatchScanWorkflow(ctx workflow.Context, contracts []ScanContractInput) ([]ScanContractOutput, error) {
	// Create multiple child workflows for contracts in batches
	results := make([]ScanContractOutput, 0, len(contracts))

	// Process in batches of 5 due to concurrency limits
	batchSize := 5
	for i := 0; i < len(contracts); i += batchSize {
		end := i + batchSize
		if end > len(contracts) {
			end = len(contracts)
		}
		batchContracts := contracts[i:end]

		// Run batch in parallel
		futures := make([]workflow.Future, 0, len(batchContracts))
		for _, contract := range batchContracts {
			future := workflow.ExecuteChildWorkflow(ctx, ExecuteScanContractWorkflow, contract)
			futures = append(futures, future)
		}

		// Collect batch results
		for _, future := range futures {
			var result ScanContractOutput
			err := future.Get(ctx, &result)
			if err != nil {
				workflow.GetLogger(ctx).Error("child workflow failed", "error", err)
				continue
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// PriorityQueueWorkflow processes contracts from priority queue
func PriorityQueueWorkflow(ctx workflow.Context) error {
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, opts)

	// Process queue items until context is cancelled
	for {
		// Fetch next item from priority queue
		var queueItem ScanContractInput
		err := workflow.ExecuteActivity(ctx, DequeueContractActivity).Get(ctx, &queueItem)
		if err != nil {
			// Retry after delay
			workflow.Sleep(ctx, 10*time.Second)
			continue
		}

		// Execute scan
		var output ScanContractOutput
		childErr := workflow.ExecuteChildWorkflow(ctx, ExecuteScanContractWorkflow, queueItem).Get(ctx, &output)
		if childErr != nil {
			workflow.GetLogger(ctx).Error("scan workflow failed", "error", childErr)
		}
	}
}

// ============================================================================
// Activity Definitions
// ============================================================================

// ScanInput wraps scan parameters
type ScanInput struct {
	Contract   *domain.Contract
	EnableML   bool
	MLFeatures []float64
}

// FetchContractActivity fetches contract from storage
func FetchContractActivity(ctx context.Context, contractID string) (*domain.Contract, error) {
	// In real implementation, fetch from database
	return &domain.Contract{
		ID:      domain.ContractID(contractID),
		ChainID: 1,
	}, nil
}

// PerformScanActivity executes the multi-engine scan
func PerformScanActivity(ctx context.Context, input ScanInput) (*domain.ScanReport, error) {
	// Would create orchestrator and run scan
	return &domain.ScanReport{
		ID:         fmt.Sprintf("temp_%d", time.Now().Unix()),
		ContractID: input.Contract.ID,
		Status:     domain.ScanStatusCompleted,
		RiskScore:  5.0,
		ThreatLevel: domain.ThreatLevelMedium,
		StartedAt:  time.Now(),
	}, nil
}

// StoreScanResultActivity persists results to database
func StoreScanResultActivity(ctx context.Context, report *domain.ScanReport) (string, error) {
	// Store in database
	return report.ID, nil
}

// NotifyResultsActivity sends notifications about scan results
func NotifyResultsActivity(ctx context.Context, report *domain.ScanReport) error {
	// Send notifications (websocket, email, webhook, etc)
	return nil
}

// DequeueContractActivity gets next contract from priority queue
func DequeueContractActivity(ctx context.Context) (*ScanContractInput, error) {
	// Get from Redis priority queue
	// Return nil if empty (non-fatal)
	return nil, fmt.Errorf("queue empty")
}
