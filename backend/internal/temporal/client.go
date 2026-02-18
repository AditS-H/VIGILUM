// Package temporal implements Temporal.io workflow orchestration for VIGILUM.
package temporal

import (
	"context"
	"log/slog"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/sdk/activity"
)

// ClientConfig contains Temporal client configuration.
type ClientConfig struct {
	HostPort    string
	Namespace   string
	TaskQueue   string
	Timeout     time.Duration
}

// DefaultClientConfig returns sensible defaults.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		HostPort:  "localhost:7233",
		Namespace: "default",
		TaskQueue: "vigilum-tasks",
		Timeout:   30 * time.Second,
	}
}

// Client wraps Temporal SDK client.
type Client struct {
	logger *slog.Logger
	client client.Client
	config ClientConfig
}

// NewClient creates a new Temporal client.
func NewClient(logger *slog.Logger, config ClientConfig) (*Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
	})
	if err != nil {
		logger.Error("failed to create Temporal client", "error", err)
		return nil, err
	}

	return &Client{
		logger: logger.With("service", "temporal"),
		client: c,
		config: config,
	}, nil
}

// ExecuteWorkflow starts a workflow execution.
func (c *Client) ExecuteWorkflow(ctx context.Context, workflowID string, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                     workflowID,
		TaskQueue:              c.config.TaskQueue,
		WorkflowExecutionTimeout: c.config.Timeout,
	}

	run, err := c.client.ExecuteWorkflow(ctx, workflowOptions, workflow, args...)
	if err != nil {
		c.logger.Error("failed to execute workflow", "workflow_id", workflowID, "error", err)
		return nil, err
	}

	c.logger.Info("workflow started", "workflow_id", workflowID)
	return run, nil
}

// GetWorkflowResult waits for workflow completion and returns result.
func (c *Client) GetWorkflowResult(ctx context.Context, workflowID string, runID string, valueType interface{}) error {
	run := c.client.GetWorkflow(ctx, workflowID, runID)
	err := run.Get(ctx, valueType)
	if err != nil {
		c.logger.Error("failed to get workflow result", "workflow_id", workflowID, "error", err)
		return err
	}
	return nil
}

// Close closes the Temporal client.
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// WORKFLOW DEFINITIONS
// ═══════════════════════════════════════════════════════════════════════════════

// ProofVerificationWorkflow orchestrates async proof verification.
func ProofVerificationWorkflow(ctx workflow.Context, proofID string, contractAddr string) (bool, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProofVerificationWorkflow started", "proof_id", proofID, "contract", contractAddr)

	// Activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Fetch proof from registry
	var proofData ProofData
	err := workflow.ExecuteActivity(ctx, FetchProofActivity, proofID).Get(ctx, &proofData)
	if err != nil {
		logger.Error("failed to fetch proof", "proof_id", proofID)
		return false, err
	}

	// Step 2: Verify WASM proof
	var isValid bool
	err = workflow.ExecuteActivity(ctx, VerifyProofActivity, proofData).Get(ctx, &isValid)
	if err != nil {
		logger.Error("failed to verify proof", "proof_id", proofID)
		return false, err
	}

	// Step 3: Update registry with result
	err = workflow.ExecuteActivity(ctx, UpdateProofStatusActivity, proofID, isValid).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update proof status", "proof_id", proofID)
		return false, err
	}

	logger.Info("ProofVerificationWorkflow completed", "proof_id", proofID, "valid", isValid)
	return isValid, nil
}

// ScanWorkflow orchestrates contract scanning and analysis.
func ScanWorkflow(ctx workflow.Context, contractAddr string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ScanWorkflow started", "contract", contractAddr)

	// Activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Fetch bytecode
	var bytecode string
	err := workflow.ExecuteActivity(ctx, FetchBytecodeActivity, contractAddr).Get(ctx, &bytecode)
	if err != nil {
		logger.Error("failed to fetch bytecode", "contract", contractAddr)
		return "", err
	}

	// Step 2: Extract features
	var features interface{}
	err = workflow.ExecuteActivity(ctx, ExtractFeaturesActivity, bytecode).Get(ctx, &features)
	if err != nil {
		logger.Error("failed to extract features", "contract", contractAddr)
		return "", err
	}

	// Step 3: Run ML inference
	var riskScore float64
	err = workflow.ExecuteActivity(ctx, InferenceActivity, features).Get(ctx, &riskScore)
	if err != nil {
		logger.Error("failed to run inference", "contract", contractAddr)
		return "", err
	}

	// Step 4: Publish to oracle
	var resultID string
	err = workflow.ExecuteActivity(ctx, PublishOracleActivity, contractAddr, riskScore).Get(ctx, &resultID)
	if err != nil {
		logger.Error("failed to publish to oracle", "contract", contractAddr)
		return "", err
	}

	logger.Info("ScanWorkflow completed", "contract", contractAddr, "risk_score", riskScore, "result_id", resultID)
	return resultID, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// ACTIVITY DEFINITIONS
// ═══════════════════════════════════════════════════════════════════════════════

// ProofData represents proof information.
type ProofData struct {
	ID           string
	ContractAddr string
	ProofBytes   []byte
	Timestamp    time.Time
}

// FetchProofActivity retrieves proof from registry.
func FetchProofActivity(ctx context.Context, proofID string) (ProofData, error) {
	// TODO: Implement actual fetch from VigilumRegistry contract
	logger := activity.GetLogger(ctx)
	logger.Info("fetching proof", "proof_id", proofID)
	return ProofData{
		ID:        proofID,
		Timestamp: time.Now(),
	}, nil
}

// VerifyProofActivity verifies WASM proof.
func VerifyProofActivity(ctx context.Context, proofData ProofData) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("verifying proof", "proof_id", proofData.ID)
	// TODO: Call WasmProverModule.VerifyHumanProof()
	return true, nil
}

// UpdateProofStatusActivity updates proof status on-chain.
func UpdateProofStatusActivity(ctx context.Context, proofID string, isValid bool) error {
	logger := activity.GetLogger(ctx)
	logger.Info("updating proof status", "proof_id", proofID, "valid", isValid)
	// TODO: Call EthereumClient to update VigilumRegistry
	return nil
}

// FetchBytecodeActivity retrieves contract bytecode.
func FetchBytecodeActivity(ctx context.Context, contractAddr string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("fetching bytecode", "contract", contractAddr)
	// TODO: Fetch from Etherscan or node
	return "", nil
}

// ExtractFeaturesActivity extracts features from bytecode.
func ExtractFeaturesActivity(ctx context.Context, bytecode string) (interface{}, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("extracting features", "bytecode_len", len(bytecode))
	// TODO: Call ML features.py
	return nil, nil
}

// InferenceActivity runs ML inference.
func InferenceActivity(ctx context.Context, features interface{}) (float64, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("running inference")
	// TODO: Call ML model.py for inference
	return 0.0, nil
}

// PublishOracleActivity publishes result to oracle.
func PublishOracleActivity(ctx context.Context, contractAddr string, riskScore float64) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("publishing to oracle", "contract", contractAddr, "risk_score", riskScore)
	// TODO: Call oracle.Service.PublishSignal()
	return "", nil
}
