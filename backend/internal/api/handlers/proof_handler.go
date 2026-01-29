// Package handlers implements HTTP handlers for proof verification endpoints.
package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vigilum/backend/internal/domain"
	"github.com/vigilum/backend/internal/proof"
	zkproof "github.com/vigilum/backend/internal/proof/zkproof"
)

// ProofHandler handles proof verification HTTP requests.
type ProofHandler struct {
	verifier *proof.HumanProofVerifier
	logger   *slog.Logger
}

// NewProofHandler creates a new proof handler.
func NewProofHandler(
	verifier *proof.HumanProofVerifier,
	logger *slog.Logger,
) *ProofHandler {
	return &ProofHandler{
		verifier: verifier,
		logger:   logger,
	}
}

// Request/Response DTOs

// GenerateChallengeRequest is the request body for challenge generation.
type GenerateChallengeRequest struct {
	UserID            string `json:"user_id" binding:"required"`
	VerifierAddress   string `json:"verifier_address" binding:"required"`
}

// GenerateChallengeResponse is the response for challenge generation.
type GenerateChallengeResponse struct {
	ChallengeID string    `json:"challenge_id"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	TTL         int       `json:"ttl_seconds"`
}

// SubmitProofRequest is the request body for proof submission.
type SubmitProofRequest struct {
	ChallengeID    string `json:"challenge_id" binding:"required"`
	ProofData      string `json:"proof_data" binding:"required"` // hex-encoded
	TimingVariance int64  `json:"timing_variance" binding:"required"`
	GasVariance    int64  `json:"gas_variance" binding:"required"`
	ProofNonce     string `json:"proof_nonce" binding:"required"`
}

// SubmitProofResponse is the response for proof submission.
type SubmitProofResponse struct {
	IsValid              bool      `json:"is_valid"`
	VerificationScore    float64   `json:"verification_score"`
	VerificationResult   string    `json:"verification_result"`
	RiskScoreReduction   int       `json:"risk_score_reduction"`
	ProofID              string    `json:"proof_id,omitempty"`
	VerifiedAt           time.Time `json:"verified_at,omitempty"`
	Message              string    `json:"message"`
}

// GetUserProofsResponse is the response for retrieving user proofs.
type GetUserProofsResponse struct {
	UserID       string               `json:"user_id"`
	ProofCount   int                  `json:"proof_count"`
	Proofs       []ProofInfo          `json:"proofs"`
	AverageScore float64              `json:"average_score"`
	PageInfo     PaginationInfo       `json:"page_info"`
}

// ProofInfo contains proof information.
type ProofInfo struct {
	ID                  string    `json:"id"`
	ProofHash           string    `json:"proof_hash"`
	VerificationScore   float64   `json:"verification_score"`
	VerifiedAt          time.Time `json:"verified_at,omitempty"`
	ExpiresAt           time.Time `json:"expires_at"`
	CreatedAt           time.Time `json:"created_at"`
	VerifierAddress     string    `json:"verifier_address"`
}

// GetVerificationScoreResponse is the response for user verification score.
type GetVerificationScoreResponse struct {
	UserID                string    `json:"user_id"`
	VerificationScore     float64   `json:"verification_score"`
	ProofCount            int       `json:"proof_count"`
	VerifiedProofCount    int       `json:"verified_proof_count"`
	IsVerified            bool      `json:"is_verified"`
	LastVerifiedAt        time.Time `json:"last_verified_at,omitempty"`
	RiskScore             int       `json:"risk_score"`
}

// PaginationInfo contains pagination metadata.
type PaginationInfo struct {
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	Total     int `json:"total"`
	TotalPage int `json:"total_pages"`
}

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error      string    `json:"error"`
	Message    string    `json:"message"`
	StatusCode int       `json:"status_code"`
	Timestamp  time.Time `json:"timestamp"`
}

// GenerateChallenge handles POST /api/v1/proofs/challenges
func (ph *ProofHandler) GenerateChallenge(c *gin.Context) {
	var req GenerateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.logger.Warn("Invalid request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "invalid_request",
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
			Timestamp:  time.Now(),
		})
		return
	}

	// Validate user ID format
	if !isValidUserID(req.UserID) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "invalid_user_id",
			Message:    "User ID must be non-empty string",
			StatusCode: http.StatusBadRequest,
			Timestamp:  time.Now(),
		})
		return
	}

	// Validate verifier address format
	if !isValidAddress(req.VerifierAddress) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "invalid_verifier_address",
			Message:    "Verifier address must be valid Ethereum address",
			StatusCode: http.StatusBadRequest,
			Timestamp:  time.Now(),
		})
		return
	}

	// Generate challenge
	challenge, err := ph.verifier.GenerateProofChallenge(c.Request.Context(), req.UserID, domain.Address(req.VerifierAddress))
	if err != nil {
		ph.logger.Error("Failed to generate challenge",
			slog.String("user_id", req.UserID),
			slog.Any("error", err),
		)

		statusCode := http.StatusInternalServerError
		errType := "internal_error"
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
			errType = "user_not_found"
		} else if strings.Contains(err.Error(), "blacklisted") {
			statusCode = http.StatusForbidden
			errType = "user_blacklisted"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:      errType,
			Message:    err.Error(),
			StatusCode: statusCode,
			Timestamp:  time.Now(),
		})
		return
	}

	expiresAt := challenge.ExpiresAt
	ttl := int(expiresAt.Sub(time.Now()).Seconds())

	ph.logger.Info("Challenge generated successfully",
		slog.String("challenge_id", challenge.ChallengeID),
		slog.String("user_id", req.UserID),
	)

	c.JSON(http.StatusOK, GenerateChallengeResponse{
		ChallengeID: challenge.ChallengeID,
		IssuedAt:    challenge.IssuedAt,
		ExpiresAt:   expiresAt,
		TTL:         ttl,
	})
}

// SubmitProof handles POST /api/v1/proofs/verify
func (ph *ProofHandler) SubmitProof(c *gin.Context) {
	var req SubmitProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "invalid_request",
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
			Timestamp:  time.Now(),
		})
		return
	}

	// Parse hex-encoded proof data
	proofData, err := hexToBytes(req.ProofData)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "invalid_proof_data",
			Message:    "Proof data must be valid hex string",
			StatusCode: http.StatusBadRequest,
			Timestamp:  time.Now(),
		})
		return
	}

	// Create proof response
	proofResponse := &zkproof.ProofResponse{
		ChallengeID:    req.ChallengeID,
		ProofData:      proofData,
		TimingVariance: req.TimingVariance,
		GasVariance:    req.GasVariance,
		ProofNonce:     req.ProofNonce,
		SubmittedAt:    time.Now(),
	}

	// Submit proof for verification
	result, err := ph.verifier.SubmitProofResponse(c.Request.Context(), proofResponse)
	if err != nil {
		ph.logger.Error("Failed to verify proof",
			slog.String("challenge_id", req.ChallengeID),
			slog.Any("error", err),
		)

		statusCode := http.StatusInternalServerError
		errType := "verification_failed"
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
			errType = "challenge_not_found"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:      errType,
			Message:    err.Error(),
			StatusCode: statusCode,
			Timestamp:  time.Now(),
		})
		return
	}

	// Calculate risk reduction
	riskReduction := int(result.VerificationScore * 10)
	if result.VerificationScore < 0.7 {
		riskReduction = 0
	}

	resultMessage := "Proof verification failed"
	if result.IsValid {
		resultMessage = "Proof verified successfully"
	}

	ph.logger.Info("Proof verification completed",
		slog.String("challenge_id", req.ChallengeID),
		slog.Bool("is_valid", result.IsValid),
		slog.Float64("verification_score", result.VerificationScore),
	)

	c.JSON(http.StatusOK, SubmitProofResponse{
		IsValid:            result.IsValid,
		VerificationScore:  result.VerificationScore,
		VerificationResult: resultMessage,
		RiskScoreReduction: riskReduction,
		VerifiedAt:         result.VerifiedAt,
		Message:            resultMessage,
	})
}

// GetUserProofs handles GET /api/v1/proofs?user_id=xxx&page=1&limit=10
func (ph *ProofHandler) GetUserProofs(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "missing_user_id",
			Message:    "user_id query parameter is required",
			StatusCode: http.StatusBadRequest,
			Timestamp:  time.Now(),
		})
		return
	}

	// Parse pagination
	page := 1
	if p := c.Query("page"); p != "" {
		if val, err := parseInt(p); err == nil && val > 0 {
			page = val
		}
	}

	pageSize := 10
	if ps := c.Query("limit"); ps != "" {
		if val, err := parseInt(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	offset := (page - 1) * pageSize

	// Get total count for pagination
	totalCount, err := ph.verifier.GetUserProofCount(c.Request.Context(), userID)
	if err != nil {
		ph.logger.Error("Failed to get user proof count",
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:      "internal_error",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
			Timestamp:  time.Now(),
		})
		return
	}

	// Retrieve proofs
	proofs, err := ph.verifier.GetUserProofs(c.Request.Context(), userID, pageSize, offset)
	if err != nil {
		ph.logger.Error("Failed to retrieve user proofs",
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:      "internal_error",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
			Timestamp:  time.Now(),
		})
		return
	}

	// Convert to response DTOs
	proofInfos := make([]ProofInfo, len(proofs))
	totalVerified := 0
	for i, p := range proofs {
		proofInfos[i] = ProofInfo{
			ID:                p.ID,
			ProofHash:         p.ProofHash,
			VerificationScore: p.ProofData.VerificationScore,
			VerifiedAt:        p.VerifiedAt.Time,
			ExpiresAt:         p.ExpiresAt,
			CreatedAt:         p.CreatedAt,
			VerifierAddress:   string(p.VerifierAddress),
		}
		if p.VerifiedAt.Valid {
			totalVerified++
		}
	}

	// Calculate average score
	avgScore := 0.0
	if len(proofs) > 0 {
		totalScore := 0.0
		for _, p := range proofs {
			totalScore += p.ProofData.VerificationScore
		}
		avgScore = totalScore / float64(len(proofs))
	}

	// Calculate total pages
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	ph.logger.Info("User proofs retrieved",
		slog.String("user_id", userID),
		slog.Int("proof_count", len(proofs)),
		slog.Int64("total_count", totalCount),
	)

	c.JSON(http.StatusOK, GetUserProofsResponse{
		UserID:       userID,
		ProofCount:   int(totalCount),
		Proofs:       proofInfos,
		AverageScore: avgScore,
		PageInfo: PaginationInfo{
			Page:      page,
			PageSize:  pageSize,
			Total:     int(totalCount),
			TotalPage: totalPages,
		},
	})
}

// GetVerificationScore handles GET /api/v1/verification-score?user_id=xxx
func (ph *ProofHandler) GetVerificationScore(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "missing_user_id",
			Message:    "user_id query parameter is required",
			StatusCode: http.StatusBadRequest,
			Timestamp:  time.Now(),
		})
		return
	}

	// Get verification score
	score, err := ph.verifier.GetUserVerificationScore(c.Request.Context(), userID)
	if err != nil {
		ph.logger.Error("Failed to get verification score",
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:      "internal_error",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
			Timestamp:  time.Now(),
		})
		return
	}

	// Check if user is verified
	isVerified, err := ph.verifier.IsUserVerified(c.Request.Context(), userID)
	if err != nil {
		ph.logger.Error("Failed to check user verification status",
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
	}

	// Get user verification metadata (risk score, last verified time, counts)
	metadata, err := ph.verifier.GetUserVerificationMetadata(c.Request.Context(), userID)
	if err != nil {
		ph.logger.Error("Failed to get user verification metadata",
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:      "internal_error",
			Message:    "Failed to retrieve user metadata",
			StatusCode: http.StatusInternalServerError,
			Timestamp:  time.Now(),
		})
		return
	}

	ph.logger.Info("Verification score retrieved",
		slog.String("user_id", userID),
		slog.Float64("score", score),
		slog.Int64("proof_count", metadata.ProofCount),
		slog.Int64("verified_count", metadata.VerifiedCount),
	)

	c.JSON(http.StatusOK, GetVerificationScoreResponse{
		UserID:             userID,
		VerificationScore:  score,
		IsVerified:         isVerified,
		RiskScore:          int(metadata.RiskScore),
		LastVerifiedAt:     metadata.LastVerifiedAt,
		ProofCount:         int(metadata.ProofCount),
		VerifiedProofCount: int(metadata.VerifiedCount),
	})
}

// GetChallengeStatus handles GET /api/v1/proofs/challenges/:challenge_id
func (ph *ProofHandler) GetChallengeStatus(c *gin.Context) {
	challengeID := c.Param("challenge_id")
	if challengeID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:      "missing_challenge_id",
			Message:    "challenge_id parameter is required",
			StatusCode: http.StatusBadRequest,
			Timestamp:  time.Now(),
		})
		return
	}

	// Get challenge status
	status, err := ph.verifier.GetChallengeStatus(c.Request.Context(), challengeID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:      "challenge_not_found",
			Message:    err.Error(),
			StatusCode: http.StatusNotFound,
			Timestamp:  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"challenge_id": challengeID,
		"status":       status,
		"timestamp":    time.Now(),
	})
}

// Health checks the health of the proof service
func (ph *ProofHandler) Health(c *gin.Context) {
	metrics := ph.verifier.GenerateProofMetrics(c.Request.Context())
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"service":  "proof-verification",
		"metrics":  metrics,
		"timestamp": time.Now(),
	})
}

// Helper functions

func isValidUserID(userID string) bool {
	return userID != "" && len(userID) <= 255
}

func isValidAddress(address string) bool {
	// Simple Ethereum address validation (0x + 40 hex chars)
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return false
	}
	for _, c := range address[2:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// hexToBytes converts a hex string to bytes using encoding/hex.
// Strips the 0x prefix if present.
func hexToBytes(hexStr string) ([]byte, error) {
	// Strip 0x or 0X prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")
	hexStr = strings.TrimPrefix(hexStr, "0X")
	
	// Decode using standard library
	return hex.DecodeString(hexStr)
}

func parseInt(s string) (int, error) {
	var val int
	_, err := json.Unmarshal([]byte(s), &val)
	return val, err
}
