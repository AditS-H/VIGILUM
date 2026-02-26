package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vigilum/backend/internal/domain"
	"github.com/vigilum/backend/internal/scanner"
)

// ScannerHandler handles scanning requests
type ScannerHandler struct {
	logger      *slog.Logger
	orchestrator *scanner.Orchestrator
}

// NewScannerHandler creates a new scanner handler
func NewScannerHandler(logger *slog.Logger, orchestrator *scanner.Orchestrator) *ScannerHandler {
	return &ScannerHandler{
		logger:      logger,
		orchestrator: orchestrator,
	}
}

// ScanRequest is the HTTP request for scanning
type ScanRequest struct {
	ContractID  string `json:"contract_id"`
	Address     string `json:"address"`
	ChainID     int64  `json:"chain_id"`
	SourceCode  string `json:"source_code,omitempty"`
	Bytecode    string `json:"bytecode,omitempty"`
	Timeout     int    `json:"timeout,omitempty"`
	EnableML    bool   `json:"enable_ml,omitempty"`
}

// ScanResponse is the HTTP response
type ScanResponse struct {
	ID              string                 `json:"id"`
	ContractID      string                 `json:"contract_id"`
	RiskScore       float64                `json:"risk_score"`
	ThreatLevel     string                 `json:"threat_level"`
	Status          string                 `json:"status"`
	Vulnerabilities []VulnResponse         `json:"vulnerabilities"`
	Metrics         map[string]interface{} `json:"metrics"`
	Duration        float64                `json:"duration_seconds"`
	CompletedAt     time.Time              `json:"completed_at"`
}

// VulnResponse is a vulnerability in the response
type VulnResponse struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	DetectedBy  string  `json:"detected_by"`
	Location    struct {
		File      string `json:"file"`
		StartLine int64  `json:"start_line"`
	} `json:"location"`
}

// ScanContract handles POST /api/v1/scan
func (h *ScannerHandler) ScanContract(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ContractID == "" {
		http.Error(w, "contract_id is required", http.StatusBadRequest)
		return
	}

	// Create contract domain object
	contract := &domain.Contract{
		ID:      domain.ContractID(req.ContractID),
		Address: domain.Address(req.Address),
		ChainID: domain.ChainID(req.ChainID),
	}

	if req.SourceCode != "" {
		contract.SourceCode = req.SourceCode
	}

	if req.Bytecode != "" {
		contract.Bytecode = []byte(req.Bytecode)
	}

	// Set scan options
	opts := &scanner.ScanOptions{
		Timeout:    req.Timeout,
		MaxDepth:   50,
		EnableML:   req.EnableML,
		IncludeInfo: false,
	}
	if opts.Timeout == 0 {
		opts.Timeout = 300 // Default timeout
	}

	h.logger.Info("Starting contract scan",
		"contract_id", req.ContractID,
		"address", req.Address,
	)

	// Execute scan
	startTime := time.Now()
	report, err := h.orchestrator.ScanAll(r.Context(), contract, opts)
	duration := time.Since(startTime)

	if err != nil && report == nil {
		h.logger.Error("scan failed", "error", err)
		http.Error(w, fmt.Sprintf("scan failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response
	response := ScanResponse{
		ID:          report.ID,
		ContractID:  string(report.ContractID),
		RiskScore:   report.RiskScore,
		ThreatLevel: string(report.ThreatLevel),
		Status:      string(report.Status),
		Duration:    duration.Seconds(),
	}

	if report.CompletedAt != nil {
		response.CompletedAt = *report.CompletedAt
	}

	// Convert vulnerabilities
	response.Vulnerabilities = make([]VulnResponse, len(report.Vulnerabilities))
	for i, vuln := range report.Vulnerabilities {
		response.Vulnerabilities[i] = VulnResponse{
			Type:        string(vuln.Type),
			Severity:    string(vuln.Severity),
			Title:       vuln.Title,
			Description: vuln.Description,
			Confidence:  vuln.Confidence,
			DetectedBy:  vuln.DetectedBy,
			Location: struct {
				File      string `json:"file"`
				StartLine int64  `json:"start_line"`
			}{
				File:      vuln.Location.File,
				StartLine: vuln.Location.StartLine,
			},
		}
	}

	// Convert metrics
	response.Metrics = map[string]interface{}{
		"total_issues":   report.Metrics.TotalIssues,
		"critical_count": report.Metrics.CriticalCount,
		"high_count":     report.Metrics.HighCount,
		"medium_count":   report.Metrics.MediumCount,
		"low_count":      report.Metrics.LowCount,
		"info_count":     report.Metrics.InfoCount,
	}

	h.logger.Info("Scan completed",
		"contract_id", req.ContractID,
		"risk_score", report.RiskScore,
		"threat_level", report.ThreatLevel,
		"vulnerabilities", len(report.Vulnerabilities),
		"duration_ms", duration.Milliseconds(),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthResponse is the health check response
type HealthResponse struct {
	Status   string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Scanners map[string]interface{} `json:"scanners"`
}

// Health handles GET /api/v1/health
func (h *ScannerHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Would check actual scanner health
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Scanners: map[string]interface{}{
			"slither": map[string]interface{}{
				"ready": true,
			},
			"mythril": map[string]interface{}{
				"ready": true,
			},
			"ml": map[string]interface{}{
				"ready": false, // Would check actual status
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StatsResponse contains system statistics
type StatsResponse struct {
	Timestamp      time.Time   `json:"timestamp"`
	ScansCompleted int64       `json:"scans_completed"`
	AvgRiskScore   float64     `json:"avg_risk_score"`
	IndexerStats   interface{} `json:"indexer_stats"`
}

// Stats handles GET /api/v1/stats
func (h *ScannerHandler) Stats(w http.ResponseWriter, r *http.Request) {
	response := StatsResponse{
		Timestamp:      time.Now(),
		ScansCompleted: 0, // Would query from database
		AvgRiskScore:   0.0, // Would calculate from DB
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
