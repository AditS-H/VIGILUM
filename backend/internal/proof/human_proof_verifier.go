// Package proof provides human proof verification using ZK proofs.
package proof

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/vigilum/backend/internal/db/repositories"
	"github.com/vigilum/backend/internal/domain"
	zkproof "github.com/vigilum/backend/internal/proof/zkproof"
)

// HumanProofVerifier orchestrates proof generation, challenges, and verification.
type HumanProofVerifier struct {
	proofRepo    domain.HumanProofRepository
	userRepo     domain.UserRepository
	proofService *zkproof.ProofService
	logger       *slog.Logger
}

// NewHumanProofVerifier creates a new proof verifier.
func NewHumanProofVerifier(
	db *sql.DB,
	config zkproof.ProofServiceConfig,
	logger *slog.Logger,
) *HumanProofVerifier {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(nil, nil))
	}

	return &HumanProofVerifier{
		proofRepo:    repositories.NewHumanProofRepository(db),
		userRepo:     repositories.NewUserRepository(db),
		proofService: zkproof.NewProofService(config, logger),
		logger:       logger,
	}
}

// GenerateProofChallenge creates a new proof challenge for a user.
func (hpv *HumanProofVerifier) GenerateProofChallenge(
	ctx context.Context,
	userID string,
	verifierAddress domain.Address,
) (*zkproof.ProofChallenge, error) {
	hpv.logger.InfoContext(ctx, "Generating proof challenge", "user_id", userID)

	// Verify user exists
	user, err := hpv.userRepo.GetByID(ctx, userID)
	if err == domain.ErrNotFound {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Check if user is blacklisted
	if user.IsBlacklisted {
		return nil, fmt.Errorf("user is blacklisted: %s", userID)
	}

	// Issue challenge from proof service
	challenge, err := hpv.proofService.IssueChallenge(ctx, userID, verifierAddress)
	if err != nil {
		return nil, fmt.Errorf("issue challenge: %w", err)
	}

	hpv.logger.InfoContext(ctx, "Challenge generated",
		"challenge_id", challenge.ChallengeID,
		"user_id", userID,
	)

	return challenge, nil
}

// SubmitProofResponse processes a user's proof submission.
func (hpv *HumanProofVerifier) SubmitProofResponse(
	ctx context.Context,
	response *zkproof.ProofResponse,
) (*zkproof.ProofVerificationResult, error) {
	hpv.logger.InfoContext(ctx, "Processing proof response",
		"challenge_id", response.ChallengeID,
	)

	// Verify the proof
	result, err := hpv.proofService.SubmitProof(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("submit proof: %w", err)
	}

	// If verification successful, store proof in repository
	if result.IsValid {
		// Retrieve challenge to get user info
		challenge, err := hpv.proofService.GetChallenge(ctx, response.ChallengeID)
		if err != nil {
			return nil, fmt.Errorf("get challenge: %w", err)
		}

		// Calculate proof hash
		proofData := domain.ProofData{
			TimingVariance:    result.TimingVariance,
			GasVariance:       result.GasVariance,
			ContractDiversity: result.ContractDiversity,
			ProofNonce:        response.ProofNonce,
		}
		proofHash := zkproof.CalculateProofHash(challenge.UserID, proofData)

		// Create proof record
		proof := &domain.HumanProof{
			ID:              fmt.Sprintf("proof_%d", time.Now().UnixNano()),
			UserID:          challenge.UserID,
			ProofHash:       proofHash,
			ProofData:       proofData,
			VerifierAddress: challenge.VerifierAddress,
			ExpiresAt:       time.Now().Add(24 * 365 * time.Hour), // 1 year TTL
			VerifiedAt:      sql.NullTime{Time: result.VerifiedAt, Valid: true},
			CreatedAt:       time.Now(),
		}

		// Store proof in repository
		if err := hpv.proofRepo.Create(ctx, proof); err != nil {
			return nil, fmt.Errorf("store proof: %w", err)
		}

		// Update user risk score based on verification
		// Lower risk score for successful verification
		userRiskReduction := int32(10) // Reduce risk by 10 points
		newRiskScore := int32(0)
		if currentUser, err := hpv.userRepo.GetByID(ctx, challenge.UserID); err == nil {
			newRiskScore = currentUser.RiskScore - userRiskReduction
			if newRiskScore < 0 {
				newRiskScore = 0
			}

			if err := hpv.userRepo.UpdateRiskScore(ctx, challenge.UserID, newRiskScore); err != nil {
				hpv.logger.ErrorContext(ctx, "Failed to update user risk score",
					"user_id", challenge.UserID,
					"error", err,
				)
			}
		}

		hpv.logger.InfoContext(ctx, "Proof verified and stored",
			"proof_id", proof.ID,
			"user_id", challenge.UserID,
			"verification_score", result.VerificationScore,
			"new_risk_score", newRiskScore,
		)
	}

	return result, nil
}

// GetUserProofs retrieves all verified proofs for a user.
func (hpv *HumanProofVerifier) GetUserProofs(
	ctx context.Context,
	userID string,
	limit int,
	offset int,
) ([]*domain.HumanProof, error) {
	proofs, err := hpv.proofRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get proofs: %w", err)
	}

	return proofs, nil
}

// GetUserVerificationScore calculates a user's verification score based on proofs.
func (hpv *HumanProofVerifier) GetUserVerificationScore(
	ctx context.Context,
	userID string,
) (float32, error) {
	hpv.logger.DebugContext(ctx, "Calculating user verification score", "user_id", userID)

	// Count verified proofs
	verifiedCount, err := hpv.proofRepo.CountVerifiedByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count verified proofs: %w", err)
	}

	// Get total proofs
	totalCount, err := hpv.proofRepo.CountByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count total proofs: %w", err)
	}

	if totalCount == 0 {
		return 0, nil
	}

	// Calculate score as percentage of verified proofs
	score := float32(verifiedCount) / float32(totalCount)

	hpv.logger.DebugContext(ctx, "Verification score calculated",
		"user_id", userID,
		"verified", verifiedCount,
		"total", totalCount,
		"score", score,
	)

	return score, nil
}

// IsUserVerified checks if a user has at least one verified proof.
func (hpv *HumanProofVerifier) IsUserVerified(ctx context.Context, userID string) (bool, error) {
	verifiedCount, err := hpv.proofRepo.CountVerifiedByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("count verified proofs: %w", err)
	}

	return verifiedCount > 0, nil
}

// GetChallengeStatus retrieves the current status of a challenge.
func (hpv *HumanProofVerifier) GetChallengeStatus(
	ctx context.Context,
	challengeID string,
) (*zkproof.ProofChallenge, error) {
	challenge, err := hpv.proofService.GetChallenge(ctx, challengeID)
	if err != nil {
		return nil, fmt.Errorf("get challenge: %w", err)
	}

	return challenge, nil
}

// IsChallengeActive checks if a challenge is still active and can be submitted.
func (hpv *HumanProofVerifier) IsChallengeActive(ctx context.Context, challengeID string) bool {
	return hpv.proofService.IsChallengeValid(ctx, challengeID)
}

// GenerateProofMetrics generates analytics on proof verification system.
func (hpv *HumanProofVerifier) GenerateProofMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get total proofs in system
	// This would require a Count method on the repository
	// For now, return basic metrics

	metrics["system"] = map[string]interface{}{
		"proof_service_active": true,
		"timestamp":            time.Now(),
	}

	return metrics, nil
}

// ProofVerificationWorkflow represents a complete proof verification workflow.
type ProofVerificationWorkflow struct {
	Verifier *HumanProofVerifier
	Logger   *slog.Logger
}

// NewProofVerificationWorkflow creates a new verification workflow.
func NewProofVerificationWorkflow(
	verifier *HumanProofVerifier,
	logger *slog.Logger,
) *ProofVerificationWorkflow {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(nil, nil))
	}

	return &ProofVerificationWorkflow{
		Verifier: verifier,
		Logger:   logger,
	}
}

// ExecuteWorkflow runs a complete proof verification workflow.
func (pwf *ProofVerificationWorkflow) ExecuteWorkflow(
	ctx context.Context,
	userID string,
	verifierAddress domain.Address,
	submitProof func(ctx context.Context) (*zkproof.ProofResponse, error),
) (*zkproof.ProofVerificationResult, error) {
	pwf.Logger.InfoContext(ctx, "Starting proof verification workflow", "user_id", userID)

	// Step 1: Generate challenge
	challenge, err := pwf.Verifier.GenerateProofChallenge(ctx, userID, verifierAddress)
	if err != nil {
		return nil, fmt.Errorf("generate challenge: %w", err)
	}

	pwf.Logger.InfoContext(ctx, "Challenge generated",
		"challenge_id", challenge.ChallengeID,
	)

	// Step 2: User submits proof (via callback)
	response, err := submitProof(ctx)
	if err != nil {
		return nil, fmt.Errorf("submit proof: %w", err)
	}

	// Ensure response is linked to challenge
	response.ChallengeID = challenge.ChallengeID

	// Step 3: Verify proof
	result, err := pwf.Verifier.SubmitProofResponse(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("verify proof: %w", err)
	}

	if result.IsValid {
		pwf.Logger.InfoContext(ctx, "Proof verification workflow completed successfully",
			"user_id", userID,
			"verification_score", result.VerificationScore,
		)
	} else {
		pwf.Logger.WarnContext(ctx, "Proof verification workflow failed",
			"user_id", userID,
			"reason", result.Message,
		)
	}

	return result, nil
}
