// Package genome provides HTTP handlers for Genome Analyzer endpoints.
package genome

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// Handler provides HTTP handlers for Genome Analyzer endpoints.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new Genome Analyzer HTTP handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With("handler", "genome"),
	}
}

// RegisterRoutes registers the Genome Analyzer routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/genome/analyze", h.AnalyzeContract)
	mux.HandleFunc("GET /api/v1/genome/status/{analysis_id}", h.GetStatus)
	mux.HandleFunc("GET /api/v1/genome/{genome_hash}", h.GetGenome)
	mux.HandleFunc("GET /api/v1/genome/{genome_hash}/similar", h.FindSimilar)
}

// AnalyzeContract initiates genome analysis.
// POST /api/v1/genome/analyze
func (h *Handler) AnalyzeContract(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	if req.ContractAddress == "" {
		h.writeError(w, http.StatusBadRequest, "contract_address is required", "MISSING_FIELD")
		return
	}

	result, err := h.service.AnalyzeContract(ctx, &req)
	if err != nil {
		h.logger.Error("Analysis failed", "error", err)
		h.writeError(w, http.StatusInternalServerError, err.Error(), "ANALYSIS_ERROR")
		return
	}

	h.writeJSON(w, http.StatusAccepted, result)
}

// GetStatus retrieves analysis status.
// GET /api/v1/genome/status/{analysis_id}
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	analysisID := strings.TrimPrefix(r.URL.Path, "/api/v1/genome/status/")

	result, err := h.service.GetAnalysisStatus(ctx, analysisID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Analysis not found", "NOT_FOUND")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// GetGenome retrieves genome by hash.
// GET /api/v1/genome/{genome_hash}
func (h *Handler) GetGenome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	genomeHash := strings.TrimPrefix(r.URL.Path, "/api/v1/genome/")
	genomeHash = strings.Split(genomeHash, "/")[0]

	result, err := h.service.GetGenomeHash(ctx, genomeHash)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Genome not found", "NOT_FOUND")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// FindSimilar finds similar genomes.
// GET /api/v1/genome/{genome_hash}/similar
func (h *Handler) FindSimilar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	genomeHash := strings.TrimPrefix(r.URL.Path, "/api/v1/genome/")
	genomeHash = strings.Split(genomeHash, "/")[0]

	threshold := 0.8
	if t := r.URL.Query().Get("threshold"); t != "" {
		// Parse threshold (simplified)
	}

	results, err := h.service.FindSimilar(ctx, genomeHash, threshold)
	if err != nil {
		h.logger.Error("Find similar failed", "error", err)
		h.writeError(w, http.StatusInternalServerError, err.Error(), "ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"similar": results,
	})
}

// Helper methods

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"code":  code,
	})
}
