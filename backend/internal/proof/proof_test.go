// Package proof tests the human proof verification system.
package proof

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/vigilum/backend/internal/db/repositories"
	"github.com/vigilum/backend/internal/domain"
	zkproof "github.com/vigilum/backend/internal/proof/zkproof"
)

// setupTestEnvironment initializes test database and repositories.
func setupTestEnvironment(t *testing.T) (*sql.DB, domain.UserRepository, domain.HumanProofRepository) {
	db := repositories.SetupTestDB(t)
	if err := repositories.TruncateTables(t, db); err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	proofRepo := repositories.NewHumanProofRepository(db)

	return db, userRepo, proofRepo
}

// TestProofServiceChallenge tests the basic challenge generation.
func TestProofServiceChallenge(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	config := zkproof.ProofServiceConfig{
		ProverPath:       "/path/to/prover",
		MaxChallengeTime: 5 * time.Minute,
		MaxProofAttempts: 3,
		MinVerificationScore: 0.7,
		EnableTiming:     true,
		EnableGasAnalysis: true,
		ContractDiversity: 2,
	}

	service := zkproof.NewProofService(config, logger)
	ctx := context.Background()

	t.Run("Issue challenge", func(t *testing.T) {
		userID := "test_user_1"
		verifier := domain.Address("0xVERIFIER")

		challenge, err := service.IssueChallenge(ctx, userID, verifier)
		if err != nil {
			t.Fatalf("IssueChallenge failed: %v", err)
		}

		if challenge.ChallengeID == "" {
			t.Error("Challenge ID should not be empty")
		}
		if challenge.UserID != userID {
			t.Errorf("User ID mismatch: expected %s, got %s", userID, challenge.UserID)
		}
		if challenge.Status != "pending" {
			t.Errorf("Challenge status should be pending, got %s", challenge.Status)
		}
	})

	t.Run("Get challenge", func(t *testing.T) {
		userID := "test_user_2"
		verifier := domain.Address("0xVERIFIER2")

		challenge, _ := service.IssueChallenge(ctx, userID, verifier)
		retrieved, err := service.GetChallenge(ctx, challenge.ChallengeID)
		if err != nil {
			t.Fatalf("GetChallenge failed: %v", err)
		}

		if retrieved.ChallengeID != challenge.ChallengeID {
			t.Error("Challenge ID mismatch")
		}
	})

	t.Run("Challenge validity", func(t *testing.T) {
		userID := "test_user_3"
		verifier := domain.Address("0xVERIFIER3")

		challenge, _ := service.IssueChallenge(ctx, userID, verifier)
		if !service.IsChallengeValid(ctx, challenge.ChallengeID) {
			t.Error("Challenge should be valid")
		}
	})

	t.Run("Challenge expiration", func(t *testing.T) {
		config := zkproof.ProofServiceConfig{
			MaxChallengeTime: 1 * time.Millisecond, // Expire immediately
		}
		service := zkproof.NewProofService(config, logger)

		userID := "test_user_4"
		verifier := domain.Address("0xVERIFIER4")

		challenge, _ := service.IssueChallenge(ctx, userID, verifier)
		time.Sleep(10 * time.Millisecond) // Wait for expiration

		if service.IsChallengeValid(ctx, challenge.ChallengeID) {
			t.Error("Challenge should be expired")
		}
	})
}

// TestProofVerification tests proof submission and verification.
func TestProofVerification(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	config := zkproof.ProofServiceConfig{
		MaxChallengeTime:    5 * time.Minute,
		MaxProofAttempts:    3,
		MinVerificationScore: 0.7,
		EnableTiming:        true,
		EnableGasAnalysis:   true,
		ContractDiversity:   2,
	}

	service := zkproof.NewProofService(config, logger)
	ctx := context.Background()

	t.Run("Valid proof submission", func(t *testing.T) {
		userID := "test_user_5"
		verifier := domain.Address("0xVERIFIER5")

		challenge, _ := service.IssueChallenge(ctx, userID, verifier)

		response := &zkproof.ProofResponse{
			ChallengeID:    challenge.ChallengeID,
			ProofData:      []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			TimingVariance: 100,
			GasVariance:    500,
			ProofNonce:     "nonce123",
			SubmittedAt:    time.Now(),
		}

		result, err := service.SubmitProof(ctx, response)
		if err != nil {
			t.Fatalf("SubmitProof failed: %v", err)
		}

		if result.VerificationScore < 0 || result.VerificationScore > 1 {
			t.Errorf("Invalid verification score: %f", result.VerificationScore)
		}
	})

	t.Run("Proof with excessive timing variance", func(t *testing.T) {
		userID := "test_user_6"
		verifier := domain.Address("0xVERIFIER6")

		challenge, _ := service.IssueChallenge(ctx, userID, verifier)

		response := &zkproof.ProofResponse{
			ChallengeID:    challenge.ChallengeID,
			ProofData:      []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			TimingVariance: 2000, // Excessive
			GasVariance:    500,
			ProofNonce:     "nonce456",
			SubmittedAt:    time.Now(),
		}

		result, err := service.SubmitProof(ctx, response)
		if err != nil {
			t.Fatalf("SubmitProof failed: %v", err)
		}

		// Should fail validation due to excessive timing variance
		if result.IsValid {
			t.Error("Proof with excessive timing variance should fail")
		}
	})

	t.Run("Max attempts exceeded", func(t *testing.T) {
		userID := "test_user_7"
		verifier := domain.Address("0xVERIFIER7")
		config := zkproof.ProofServiceConfig{
			MaxProofAttempts: 2,
		}
		service := zkproof.NewProofService(config, logger)

		challenge, _ := service.IssueChallenge(ctx, userID, verifier)

		// Submit max attempts
		for i := 0; i < 2; i++ {
			response := &zkproof.ProofResponse{
				ChallengeID:    challenge.ChallengeID,
				ProofData:      []byte{0x00},
				TimingVariance: 100,
				GasVariance:    500,
				ProofNonce:     fmt.Sprintf("nonce%d", i),
				SubmittedAt:    time.Now(),
			}
			service.SubmitProof(ctx, response)
		}

		// Next attempt should fail
		response := &zkproof.ProofResponse{
			ChallengeID:    challenge.ChallengeID,
			ProofData:      []byte{0x00},
			TimingVariance: 100,
			GasVariance:    500,
			ProofNonce:     "final_nonce",
			SubmittedAt:    time.Now(),
		}

		_, err := service.SubmitProof(ctx, response)
		if err == nil {
			t.Error("Expected error for max attempts exceeded")
		}
	})
}

// TestHumanProofVerifier tests the full verification workflow.
func TestHumanProofVerifier(t *testing.T) {
	db, userRepo, proofRepo := setupTestEnvironment(t)
	defer repositories.CleanupTestDB(t, db)

	logger := slog.New(slog.NewTextHandler(nil, nil))
	config := zkproof.ProofServiceConfig{
		MaxChallengeTime:    5 * time.Minute,
		MaxProofAttempts:    3,
		MinVerificationScore: 0.7,
		EnableTiming:        true,
		EnableGasAnalysis:   true,
		ContractDiversity:   2,
	}

	verifier := NewHumanProofVerifier(db, config, logger)
	ctx := context.Background()

	t.Run("Generate proof challenge", func(t *testing.T) {
		// Create user first
		user := &domain.User{
			ID:            "proof_user_1",
			WalletAddress: domain.Address("0xPROOFUSER1"),
			RiskScore:     50,
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		// Generate challenge
		challenge, err := verifier.GenerateProofChallenge(ctx, user.ID, domain.Address("0xVERIFIER"))
		if err != nil {
			t.Fatalf("GenerateProofChallenge failed: %v", err)
		}

		if challenge.ChallengeID == "" {
			t.Error("Challenge ID should not be empty")
		}
	})

	t.Run("Challenge for nonexistent user", func(t *testing.T) {
		_, err := verifier.GenerateProofChallenge(ctx, "nonexistent", domain.Address("0xVERIFIER"))
		if err == nil {
			t.Error("Expected error for nonexistent user")
		}
	})

	t.Run("Challenge for blacklisted user", func(t *testing.T) {
		user := &domain.User{
			ID:            "blacklisted_user",
			WalletAddress: domain.Address("0xBLACKLISTED"),
			IsBlacklisted: true,
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		_, err := verifier.GenerateProofChallenge(ctx, user.ID, domain.Address("0xVERIFIER"))
		if err == nil {
			t.Error("Expected error for blacklisted user")
		}
	})

	t.Run("Complete verification workflow", func(t *testing.T) {
		// Create user
		user := &domain.User{
			ID:            "workflow_user",
			WalletAddress: domain.Address("0xWORKFLOW"),
			RiskScore:     60,
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		// Generate challenge
		challenge, _ := verifier.GenerateProofChallenge(ctx, user.ID, domain.Address("0xVERIFIER"))

		// Submit proof
		response := &zkproof.ProofResponse{
			ChallengeID:    challenge.ChallengeID,
			ProofData:      []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			TimingVariance: 150,
			GasVariance:    800,
			ProofNonce:     "workflow_nonce",
			SubmittedAt:    time.Now(),
		}

		result, err := verifier.SubmitProofResponse(ctx, response)
		if err != nil {
			t.Fatalf("SubmitProofResponse failed: %v", err)
		}

		// Check verification result
		if result.IsValid {
			// Verify proof was stored
			proofs, _ := verifier.GetUserProofs(ctx, user.ID, 10, 0)
			if len(proofs) == 0 {
				t.Error("Proof should be stored in repository")
			}

			// Check user risk score was updated
			updated, _ := userRepo.GetByID(ctx, user.ID)
			if updated.RiskScore >= user.RiskScore {
				t.Error("Risk score should be reduced after verification")
			}
		}
	})

	t.Run("User verification score", func(t *testing.T) {
		user := &domain.User{
			ID:            "score_user",
			WalletAddress: domain.Address("0xSCOREUSER"),
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		// Initially no proofs
		score, _ := verifier.GetUserVerificationScore(ctx, user.ID)
		if score != 0 {
			t.Errorf("Expected score 0 for user with no proofs, got %f", score)
		}

		// Add verified proof
		proof := &domain.HumanProof{
			ID:              "score_proof_1",
			UserID:          user.ID,
			ProofHash:       "test_hash",
			ProofData:       domain.ProofData{},
			VerifierAddress: domain.Address("0xVERIFIER"),
			VerifiedAt:      sql.NullTime{Time: time.Now(), Valid: true},
			ExpiresAt:       time.Now().Add(24 * time.Hour),
			CreatedAt:       time.Now(),
		}
		proofRepo.Create(ctx, proof)

		// Score should reflect verified proof
		score, _ = verifier.GetUserVerificationScore(ctx, user.ID)
		if score != 1.0 {
			t.Errorf("Expected score 1.0 for fully verified user, got %f", score)
		}
	})

	t.Run("Is user verified", func(t *testing.T) {
		user := &domain.User{
			ID:            "verify_check_user",
			WalletAddress: domain.Address("0xVERIFYCHECK"),
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		// User should not be verified initially
		verified, _ := verifier.IsUserVerified(ctx, user.ID)
		if verified {
			t.Error("User should not be verified initially")
		}

		// Add verified proof
		proof := &domain.HumanProof{
			ID:              "verify_proof",
			UserID:          user.ID,
			ProofHash:       "verify_hash",
			ProofData:       domain.ProofData{},
			VerifierAddress: domain.Address("0xVERIFIER"),
			VerifiedAt:      sql.NullTime{Time: time.Now(), Valid: true},
			ExpiresAt:       time.Now().Add(24 * time.Hour),
			CreatedAt:       time.Now(),
		}
		proofRepo.Create(ctx, proof)

		// User should now be verified
		verified, _ = verifier.IsUserVerified(ctx, user.ID)
		if !verified {
			t.Error("User should be verified after proof")
		}
	})
}

// TestProofVerificationWorkflow tests the complete workflow execution.
func TestProofVerificationWorkflow(t *testing.T) {
	db, userRepo, _ := setupTestEnvironment(t)
	defer repositories.CleanupTestDB(t, db)

	logger := slog.New(slog.NewTextHandler(nil, nil))
	config := zkproof.ProofServiceConfig{
		MaxChallengeTime:    5 * time.Minute,
		MaxProofAttempts:    3,
		MinVerificationScore: 0.7,
	}

	verifier := NewHumanProofVerifier(db, config, logger)
	workflow := NewProofVerificationWorkflow(verifier, logger)
	ctx := context.Background()

	t.Run("Execute complete workflow", func(t *testing.T) {
		// Create user
		user := &domain.User{
			ID:            "workflow_test_user",
			WalletAddress: domain.Address("0xWORKFLOWTEST"),
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		// Execute workflow
		result, err := workflow.ExecuteWorkflow(
			ctx,
			user.ID,
			domain.Address("0xVERIFIER"),
			func(ctx context.Context) (*zkproof.ProofResponse, error) {
				return &zkproof.ProofResponse{
					ProofData:      []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
					TimingVariance: 100,
					GasVariance:    500,
					ProofNonce:     "workflow_test",
					SubmittedAt:    time.Now(),
				}, nil
			},
		)

		if err != nil {
			t.Fatalf("ExecuteWorkflow failed: %v", err)
		}

		if result == nil {
			t.Error("Result should not be nil")
		}
	})
}

// BenchmarkProofGeneration benchmarks challenge generation.
func BenchmarkProofGeneration(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	config := zkproof.ProofServiceConfig{
		MaxChallengeTime: 5 * time.Minute,
	}
	service := zkproof.NewProofService(config, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.IssueChallenge(ctx, fmt.Sprintf("user_%d", i), domain.Address("0xVERIFIER"))
	}
}

// BenchmarkProofVerification benchmarks proof verification.
func BenchmarkProofVerification(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	config := zkproof.ProofServiceConfig{
		MaxChallengeTime: 5 * time.Minute,
	}
	service := zkproof.NewProofService(config, logger)
	ctx := context.Background()

	challenge, _ := service.IssueChallenge(ctx, "bench_user", domain.Address("0xVERIFIER"))

	proofData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := &zkproof.ProofResponse{
			ChallengeID:    challenge.ChallengeID,
			ProofData:      proofData,
			TimingVariance: 100,
			GasVariance:    500,
			ProofNonce:     fmt.Sprintf("nonce_%d", i),
			SubmittedAt:    time.Now(),
		}
		service.SubmitProof(ctx, response)
	}
}
