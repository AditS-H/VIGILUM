// Package firewall provides HTTP handlers for Identity Firewall endpoints.
package firewall

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// Handler provides HTTP handlers for Identity Firewall endpoints.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new Identity Firewall HTTP handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With("handler", "firewall"),
	}
}

// RegisterRoutes registers the Identity Firewall routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/firewall/challenge", h.GetChallenge)
	mux.HandleFunc("POST /api/v1/firewall/verify-proof", h.VerifyProof)
	mux.HandleFunc("GET /api/v1/firewall/risk/{address}", h.GetRiskScore)
	mux.HandleFunc("GET /api/v1/firewall/stats", h.GetStats)
}

// ============================================================
// Request/Response types
// ============================================================

type verifyProofRequest struct {
	WalletAddress string `json:"wallet_address"`
	Proof         string `json:"proof"`         // hex-encoded
	PublicInputs  string `json:"public_inputs"` // hex-encoded
	ChainID       int64  `json:"chain_id"`
}

type verifyProofResponse struct {
	Valid     bool   `json:"valid"`
	ProofHash string `json:"proof_hash"`
	ExpiresAt string `json:"expires_at,omitempty"`
	TxHash    string `json:"tx_hash,omitempty"`
	Error     string `json:"error,omitempty"`
}

type riskResponse struct {
	Address     string  `json:"address"`
	RiskScore   float64 `json:"risk_score"`
	ThreatLevel string  `json:"threat_level"`
	IsHuman     bool    `json:"is_human"`
	LastProofAt string  `json:"last_proof_at,omitempty"`
	ProofCount  int     `json:"proof_count"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ============================================================
// Handlers
// ============================================================

// GetChallenge generates a new verification challenge.
// GET /api/v1/firewall/challenge
func (h *Handler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	challenge, err := h.service.GenerateChallenge(ctx)
	if err != nil {
		h.logger.Error("Failed to generate challenge", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to generate challenge", "CHALLENGE_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, challenge)
}

// VerifyProof verifies a human-likeness proof.
// POST /api/v1/firewall/verify-proof
func (h *Handler) VerifyProof(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req verifyProofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	// Validate request
	if req.WalletAddress == "" {
		h.writeError(w, http.StatusBadRequest, "wallet_address is required", "MISSING_FIELD")
		return
	}
	if req.Proof == "" {
		h.writeError(w, http.StatusBadRequest, "proof is required", "MISSING_FIELD")
		return
	}
	if req.ChainID == 0 {
		req.ChainID = 1 // Default to Ethereum mainnet
	}

	// Decode hex proof
	proofBytes, err := hexDecode(req.Proof)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid proof format (expected hex)", "INVALID_PROOF")
		return
	}

	publicInputsBytes, _ := hexDecode(req.PublicInputs)

	submission := &ProofSubmission{
		WalletAddress: req.WalletAddress,
		Proof:         proofBytes,
		PublicInputs:  publicInputsBytes,
		ChainID:       req.ChainID,
	}

	result, err := h.service.VerifyProof(ctx, submission)
	if err != nil {
		h.logger.Error("Proof verification failed", "error", err, "wallet", req.WalletAddress)
		h.writeError(w, http.StatusBadRequest, err.Error(), "VERIFICATION_FAILED")
		return
	}

	resp := verifyProofResponse{
		Valid:     result.Valid,
		ProofHash: result.ProofHash,
		TxHash:    result.TxHash,
	}
	if result.Valid {
		resp.ExpiresAt = result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// GetRiskScore retrieves risk assessment for an address.
// GET /api/v1/firewall/risk/{address}
func (h *Handler) GetRiskScore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := r.PathValue("address")
	if address == "" {
		h.writeError(w, http.StatusBadRequest, "address is required", "MISSING_PARAM")
		return
	}

	// Parse optional chain_id query param
	chainID := int64(1) // Default to Ethereum mainnet
	if chainIDStr := r.URL.Query().Get("chain_id"); chainIDStr != "" {
		var err error
		chainID, err = parseInt64(chainIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid chain_id", "INVALID_PARAM")
			return
		}
	}

	riskInfo, err := h.service.GetRiskScore(ctx, address, chainID)
	if err != nil {
		h.logger.Error("Failed to get risk score", "error", err, "address", address)
		h.writeError(w, http.StatusInternalServerError, "Failed to get risk score", "RISK_ERROR")
		return
	}

	resp := riskResponse{
		Address:     riskInfo.Address,
		RiskScore:   riskInfo.RiskScore,
		ThreatLevel: string(riskInfo.ThreatLevel),
		IsHuman:     riskInfo.IsHuman,
		ProofCount:  riskInfo.ProofCount,
	}
	if riskInfo.LastProofAt != nil {
		resp.LastProofAt = riskInfo.LastProofAt.Format("2006-01-02T15:04:05Z07:00")
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// GetStats returns Identity Firewall statistics.
// GET /api/v1/firewall/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.service.GetStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get stats", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get statistics", "STATS_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// ============================================================
// Helper methods
// ============================================================

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message, code string) {
	h.writeJSON(w, status, errorResponse{
		Error: message,
		Code:  code,
	})
}

func hexDecode(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	if len(s)%2 != 0 {
		s = "0" + s
	}
	result := make([]byte, len(s)/2)
	for i := 0; i < len(result); i++ {
		b, err := hexCharToByte(s[i*2])
		if err != nil {
			return nil, err
		}
		b2, err := hexCharToByte(s[i*2+1])
		if err != nil {
			return nil, err
		}
		result[i] = b<<4 | b2
	}
	return result, nil
}

func hexCharToByte(c byte) (byte, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	}
	return 0, &json.InvalidUnmarshalError{}
}

func parseInt64(s string) (int64, error) {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &json.InvalidUnmarshalError{}
		}
		n = n*10 + int64(c-'0')
	}
	return n, nil
}
