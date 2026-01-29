// Package temporal implements worker registration for Temporal workflows.
package temporal

import (
	"log/slog"

	"go.temporal.io/sdk/worker"
)

// WorkerConfig contains worker configuration.
type WorkerConfig struct {
	TaskQueue string
}

// RegisterWorkflows registers all workflow definitions.
func RegisterWorkflows(w worker.Worker) {
	w.RegisterWorkflow(ProofVerificationWorkflow)
	w.RegisterWorkflow(ScanWorkflow)
}

// RegisterActivities registers all activity definitions.
func RegisterActivities(w worker.Worker) {
	w.RegisterActivity(FetchProofActivity)
	w.RegisterActivity(VerifyProofActivity)
	w.RegisterActivity(UpdateProofStatusActivity)
	w.RegisterActivity(FetchBytecodeActivity)
	w.RegisterActivity(ExtractFeaturesActivity)
	w.RegisterActivity(InferenceActivity)
	w.RegisterActivity(PublishOracleActivity)
}

// StartWorker starts the Temporal worker.
func StartWorker(logger *slog.Logger, client *Client, config WorkerConfig) (worker.Worker, error) {
	logger.Info("starting Temporal worker", "task_queue", config.TaskQueue)

	w := worker.New(client.client, config.TaskQueue, worker.Options{})

	RegisterWorkflows(w)
	RegisterActivities(w)

	err := w.Start()
	if err != nil {
		logger.Error("failed to start worker", "error", err)
		return nil, err
	}

	logger.Info("worker started successfully")
	return w, nil
}
