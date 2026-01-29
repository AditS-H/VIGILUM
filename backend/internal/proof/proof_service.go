// Package zkproof implements zero-knowledge proof verification for human identity.
package zkproof

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// ProofChallenge represents a ZK proof challenge issued to a user.
type ProofChallenge struct {
	ChallengeID     string
	UserID          string
	ChallengeData   []byte      // Serialized challenge from prover
	IssuedAt        time.Time
	ExpiresAt       time.Time
	Attempts        int32
	MaxAttempts     int32
	Status          string // "pending", "verified", "expired", "failed"
	VerifierAddress domain.Address
}

// ProofResponse represents a user's response to a ZK proof challenge.
type ProofResponse struct {
	ChallengeID    string
	ProofData      []byte // Serialized proof from prover
	TimingVariance int32  // ms - variance in response timing
	GasVariance    int32  // gwei - variance in gas simulation
	ProofNonce     string // Random nonce for replay protection
	SubmittedAt    time.Time
}

// ProofVerificationResult contains the outcome of proof verification.
type ProofVerificationResult struct {
	IsValid            bool
	VerificationScore  float32  // 0.0-1.0
	TimingVariance     int32
	GasVariance        int32
	ContractDiversity  int32    // Number of unique contracts analyzed
	VerificationTime   int64    // milliseconds
	Message            string
	VerifiedAt         time.Time
}

// ProofServiceConfig holds configuration for the proof service.
type ProofServiceConfig struct {
	ProverPath         string        // Path to ZK prover executable/library
	MaxChallengeTime   time.Duration // How long a challenge is valid
	MaxProofAttempts   int32         // Max attempts before locking user
	MinVerificationScore float32       // Minimum acceptable verification score
	EnableTiming       bool          // Track timing variance
	EnableGasAnalysis  bool          // Analyze gas consumption
	ContractDiversity  int32         // Required unique contracts for full score
	EnableWasmVerification bool      // Enable real WASM verification (false = stub mode)
}

// ProofService manages ZK proof generation, challenges, and verification.
type ProofService struct {
	config   ProofServiceConfig
	logger   *slog.Logger
	proofs   map[string]*ProofChallenge
	verifier ProofVerifier
}

// ProofVerifier interface for pluggable verification backends.
type ProofVerifier interface {
	GenerateChallenge(ctx context.Context, userID string) (*ProofChallenge, error)
	VerifyProof(ctx context.Context, challenge *ProofChallenge, response *ProofResponse) (*ProofVerificationResult, error)
}

// NewProofService creates a new proof service instance.
func NewProofService(config ProofServiceConfig, logger *slog.Logger) *ProofService {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(nil, nil))
	}

	return &ProofService{
		config:   config,
		logger:   logger,
		proofs:   make(map[string]*ProofChallenge),
		verifier: NewLocalVerifier(config),
	}
}

// IssueChallenge creates a new ZK proof challenge for a user.
func (s *ProofService) IssueChallenge(ctx context.Context, userID string, verifierAddress domain.Address) (*ProofChallenge, error) {
	s.logger.InfoContext(ctx, "Issuing proof challenge", "user_id", userID)

	// Generate challenge via verifier
	challenge, err := s.verifier.GenerateChallenge(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("generate challenge: %w", err)
	}

	challenge.UserID = userID
	challenge.VerifierAddress = verifierAddress
	challenge.IssuedAt = time.Now()
	challenge.ExpiresAt = time.Now().Add(s.config.MaxChallengeTime)
	challenge.MaxAttempts = s.config.MaxProofAttempts
	challenge.Status = "pending"

	// Store in memory (in production, use repository)
	s.proofs[challenge.ChallengeID] = challenge

	s.logger.InfoContext(ctx, "Challenge issued successfully",
		"challenge_id", challenge.ChallengeID,
		"expires_at", challenge.ExpiresAt,
	)

	return challenge, nil
}

// SubmitProof validates a user's response to a ZK proof challenge.
func (s *ProofService) SubmitProof(ctx context.Context, response *ProofResponse) (*ProofVerificationResult, error) {
	s.logger.InfoContext(ctx, "Processing proof submission", "challenge_id", response.ChallengeID)

	// Retrieve challenge
	challenge, exists := s.proofs[response.ChallengeID]
	if !exists {
		return nil, fmt.Errorf("challenge not found: %s", response.ChallengeID)
	}

	// Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		challenge.Status = "expired"
		return nil, fmt.Errorf("challenge expired at %v", challenge.ExpiresAt)
	}

	// Check attempt limit
	if challenge.Attempts >= challenge.MaxAttempts {
		challenge.Status = "failed"
		return nil, fmt.Errorf("max attempts exceeded: %d", challenge.MaxAttempts)
	}

	// Increment attempt counter
	challenge.Attempts++

	// Verify the proof
	startTime := time.Now()
	result, err := s.verifier.VerifyProof(ctx, challenge, response)
	verificationTime := time.Since(startTime).Milliseconds()
	result.VerificationTime = verificationTime

	if err != nil {
		s.logger.ErrorContext(ctx, "Proof verification failed",
			"challenge_id", response.ChallengeID,
			"error", err,
		)
		return nil, fmt.Errorf("verify proof: %w", err)
	}

	// Check minimum verification score
	if result.IsValid && result.VerificationScore < s.config.MinVerificationScore {
		result.IsValid = false
		result.Message = fmt.Sprintf("verification score %.2f below minimum %.2f",
			result.VerificationScore, s.config.MinVerificationScore)
	}

	if result.IsValid {
		challenge.Status = "verified"
		result.VerifiedAt = time.Now()

		s.logger.InfoContext(ctx, "Proof verified successfully",
			"challenge_id", response.ChallengeID,
			"verification_score", result.VerificationScore,
			"verification_time_ms", result.VerificationTime,
		)
	} else {
		s.logger.WarnContext(ctx, "Proof verification failed",
			"challenge_id", response.ChallengeID,
			"reason", result.Message,
		)
	}

	return result, nil
}

// GetChallenge retrieves a challenge by ID.
func (s *ProofService) GetChallenge(ctx context.Context, challengeID string) (*ProofChallenge, error) {
	challenge, exists := s.proofs[challengeID]
	if !exists {
		return nil, fmt.Errorf("challenge not found: %s", challengeID)
	}

	// Check if expired
	if time.Now().After(challenge.ExpiresAt) && challenge.Status == "pending" {
		challenge.Status = "expired"
	}

	return challenge, nil
}

// IsChallengeValid checks if a challenge is still valid and pending.
func (s *ProofService) IsChallengeValid(ctx context.Context, challengeID string) bool {
	challenge, exists := s.proofs[challengeID]
	if !exists {
		return false
	}

	if challenge.Status != "pending" {
		return false
	}

	if time.Now().After(challenge.ExpiresAt) {
		return false
	}

	if challenge.Attempts >= challenge.MaxAttempts {
		return false
	}

	return true
}

// CalculateProofHash computes a deterministic hash for a proof.
func CalculateProofHash(userID string, proofData domain.ProofData) string {
	hash := sha256.New()
	hash.Write([]byte(userID))
	hash.Write([]byte(fmt.Sprintf("%d", proofData.TimingVariance)))
	hash.Write([]byte(fmt.Sprintf("%d", proofData.GasVariance)))
	hash.Write([]byte(fmt.Sprintf("%d", proofData.ContractDiversity)))
	hash.Write([]byte(proofData.ProofNonce))
	return hex.EncodeToString(hash.Sum(nil))
}

// LocalVerifier implements ProofVerifier with local verification logic.
type LocalVerifier struct {
	config ProofServiceConfig
}

// NewLocalVerifier creates a new local proof verifier.
func NewLocalVerifier(config ProofServiceConfig) *LocalVerifier {
	return &LocalVerifier{config: config}
}

// GenerateChallenge generates a new ZK proof challenge.
func (lv *LocalVerifier) GenerateChallenge(ctx context.Context, userID string) (*ProofChallenge, error) {
	// In production, this would call the actual ZK prover
	// For now, generate a mock challenge

	challengeID := fmt.Sprintf("challenge_%d", time.Now().UnixNano())

	// Mock challenge data (would be generated by prover)
	challengeData := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	}

	return &ProofChallenge{
		ChallengeID:   challengeID,
		ChallengeData: challengeData,
	}, nil
}

// VerifyProof verifies a proof response against a challenge.
func (lv *LocalVerifier) VerifyProof(ctx context.Context, challenge *ProofChallenge, response *ProofResponse) (*ProofVerificationResult, error) {
	result := &ProofVerificationResult{
		IsValid: false,
		Message: "",
	}

	// Validate proof data length
	if len(response.ProofData) == 0 {
		result.Message = "proof data is empty"
		return result, nil
	}

	// Verify proof structure
	if len(response.ProofData) < 32 {
		result.Message = "proof data too short"
		return result, nil
	}

	// Check timing variance
	if lv.config.EnableTiming {
		if response.TimingVariance < 0 || response.TimingVariance > 1000 {
			result.Message = fmt.Sprintf("invalid timing variance: %d", response.TimingVariance)
			return result, nil
		}
		result.TimingVariance = response.TimingVariance
	}

	// Check gas variance
	if lv.config.EnableGasAnalysis {
		if response.GasVariance < 0 || response.GasVariance > 10000 {
			result.Message = fmt.Sprintf("invalid gas variance: %d", response.GasVariance)
			return result, nil
		}
		result.GasVariance = response.GasVariance
	}

	// Simulate contract diversity check
	// In production, this would analyze actual contract interactions
	contractDiversity := int32(3) // Mock value
	if contractDiversity >= lv.config.ContractDiversity {
		result.ContractDiversity = contractDiversity
	}

	// Calculate verification score (0.0-1.0)
	score := float32(0.9) // Mock calculation

	// Apply penalties
	if response.TimingVariance > 500 {
		score -= 0.1
	}
	if response.GasVariance > 5000 {
		score -= 0.1
	}

	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	result.VerificationScore = score
	result.IsValid = score >= lv.config.MinVerificationScore
	result.Message = fmt.Sprintf("verification score: %.2f", score)

	return result, nil
}
