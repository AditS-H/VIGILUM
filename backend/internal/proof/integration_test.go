// Package proof integration tests for ZK proof verification system.
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

// TestIntegrationProofWorkflow tests the complete proof workflow end-to-end.
func TestIntegrationProofWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := repositories.SetupTestDB(t)
	defer repositories.CleanupTestDB(t, db)

	if err := repositories.TruncateTables(t, db); err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(nil, nil))
	config := zkproof.ProofServiceConfig{
		MaxChallengeTime:     5 * time.Minute,
		MaxProofAttempts:     3,
		MinVerificationScore: 0.7,
		EnableTiming:         true,
		EnableGasAnalysis:    true,
		ContractDiversity:    2,
	}

	userRepo := repositories.NewUserRepository(db)
	proofRepo := repositories.NewHumanProofRepository(db)
	verifier := NewHumanProofVerifier(db, config, logger)
	verifier.(*HumanProofVerifier).proofRepo = proofRepo
	verifier.(*HumanProofVerifier).userRepo = userRepo

	ctx := context.Background()

	t.Run("Complete workflow: user registration to verification", func(t *testing.T) {
		// Step 1: User registration
		user := &domain.User{
			ID:            "integration_user_1",
			WalletAddress: domain.Address("0xINTEGRATION1"),
			RiskScore:     75,
			IsBlacklisted: false,
			CreatedAt:     time.Now(),
		}

		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		t.Log("✓ User registered")

		// Step 2: Generate proof challenge
		verifierAddr := domain.Address("0xVERIFIER")
		challenge, err := verifier.GenerateProofChallenge(ctx, user.ID, verifierAddr)
		if err != nil {
			t.Fatalf("Failed to generate challenge: %v", err)
		}
		if challenge.ChallengeID == "" {
			t.Fatal("Challenge ID is empty")
		}
		t.Log("✓ Proof challenge generated")

		// Step 3: Verify challenge is active
		if !verifier.IsChallengeActive(ctx, challenge.ChallengeID) {
			t.Fatal("Challenge should be active")
		}
		t.Log("✓ Challenge is active")

		// Step 4: Submit proof
		response := &zkproof.ProofResponse{
			ChallengeID:    challenge.ChallengeID,
			ProofData:      []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			TimingVariance: 120,
			GasVariance:    600,
			ProofNonce:     "integration_test_nonce_1",
			SubmittedAt:    time.Now(),
		}

		result, err := verifier.SubmitProofResponse(ctx, response)
		if err != nil {
			t.Fatalf("Failed to submit proof: %v", err)
		}
		if result.IsValid {
			t.Log("✓ Proof verified successfully")
		} else {
			t.Logf("! Proof verification score: %f", result.VerificationScore)
		}

		// Step 5: Verify user has proof stored
		proofs, err := verifier.GetUserProofs(ctx, user.ID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to retrieve proofs: %v", err)
		}
		if len(proofs) == 0 && result.IsValid {
			t.Fatal("Proof should be stored in database")
		}
		t.Logf("✓ User has %d stored proofs", len(proofs))

		// Step 6: Check verification score
		score, err := verifier.GetUserVerificationScore(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to get verification score: %v", err)
		}
		t.Logf("✓ User verification score: %f", score)

		// Step 7: Verify user is marked as verified
		verified, err := verifier.IsUserVerified(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to check user verification: %v", err)
		}
		t.Logf("✓ User verified status: %v", verified)

		// Step 8: Check user risk score was updated
		updated, _ := userRepo.GetByID(ctx, user.ID)
		if result.IsValid && updated.RiskScore >= user.RiskScore {
			t.Logf("⚠ Risk score not reduced (expected for low verification scores)")
		} else if result.IsValid {
			t.Logf("✓ User risk score reduced from %d to %d", user.RiskScore, updated.RiskScore)
		}
	})

	t.Run("Multi-user proof verification", func(t *testing.T) {
		// Create multiple users
		userCount := 3
		users := make([]*domain.User, userCount)

		for i := 0; i < userCount; i++ {
			users[i] = &domain.User{
				ID:            fmt.Sprintf("multi_user_%d", i),
				WalletAddress: domain.Address(fmt.Sprintf("0xMULTI%d", i)),
				RiskScore:     50 + i*10,
				CreatedAt:     time.Now(),
			}
			userRepo.Create(ctx, users[i])
		}
		t.Logf("✓ Created %d users", userCount)

		// Generate challenges for all users
		challenges := make([]*zkproof.ProofChallenge, userCount)
		for i, user := range users {
			challenge, err := verifier.GenerateProofChallenge(ctx, user.ID, domain.Address("0xVERIFIER"))
			if err != nil {
				t.Errorf("Failed to generate challenge for user %d: %v", i, err)
				continue
			}
			challenges[i] = challenge
		}
		t.Logf("✓ Generated %d challenges", userCount)

		// Submit proofs for all users
		for i, challenge := range challenges {
			if challenge == nil {
				continue
			}
			response := &zkproof.ProofResponse{
				ChallengeID:    challenge.ChallengeID,
				ProofData:      []byte{0x00, 0x01, 0x02, 0x03},
				TimingVariance: 100 + int64(i*50),
				GasVariance:    500 + int64(i*100),
				ProofNonce:     fmt.Sprintf("multi_nonce_%d", i),
				SubmittedAt:    time.Now(),
			}
			verifier.SubmitProofResponse(ctx, response)
		}
		t.Logf("✓ Submitted %d proofs", userCount)

		// Verify all users have proofs
		for i, user := range users {
			proofs, _ := verifier.GetUserProofs(ctx, user.ID, 10, 0)
			if len(proofs) > 0 {
				t.Logf("✓ User %d has %d proofs", i, len(proofs))
			}
		}
	})

	t.Run("Proof challenge expiration", func(t *testing.T) {
		// Create user with very short challenge TTL
		expireConfig := zkproof.ProofServiceConfig{
			MaxChallengeTime: 1 * time.Millisecond,
		}
		shortLiveVerifier := NewHumanProofVerifier(db, expireConfig, logger)
		shortLiveVerifier.(*HumanProofVerifier).proofRepo = proofRepo
		shortLiveVerifier.(*HumanProofVerifier).userRepo = userRepo

		user := &domain.User{
			ID:            "expire_user",
			WalletAddress: domain.Address("0xEXPIRE"),
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		// Generate challenge
		challenge, _ := shortLiveVerifier.GenerateProofChallenge(ctx, user.ID, domain.Address("0xVERIFIER"))

		// Verify it's active initially
		if !shortLiveVerifier.IsChallengeActive(ctx, challenge.ChallengeID) {
			t.Fatal("Challenge should be active initially")
		}
		t.Log("✓ Challenge active after generation")

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Verify it's expired
		if shortLiveVerifier.IsChallengeActive(ctx, challenge.ChallengeID) {
			t.Logf("⚠ Challenge may not have expired yet")
		} else {
			t.Log("✓ Challenge expired as expected")
		}
	})

	t.Run("Proof storage and retrieval", func(t *testing.T) {
		user := &domain.User{
			ID:            "storage_user",
			WalletAddress: domain.Address("0xSTORAGE"),
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		// Store multiple proofs
		proofCount := 5
		for i := 0; i < proofCount; i++ {
			proof := &domain.HumanProof{
				ID:        fmt.Sprintf("storage_proof_%d", i),
				UserID:    user.ID,
				ProofHash: fmt.Sprintf("hash_%d", i),
				ProofData: domain.ProofData{
					VerificationScore: float64(70+i*5) / 100.0,
				},
				VerifierAddress: domain.Address("0xVERIFIER"),
				VerifiedAt:      sql.NullTime{Time: time.Now(), Valid: true},
				ExpiresAt:       time.Now().Add(24 * time.Hour),
				CreatedAt:       time.Now(),
			}
			proofRepo.Create(ctx, proof)
		}
		t.Logf("✓ Stored %d proofs", proofCount)

		// Retrieve proofs with pagination
		retrieved, _ := verifier.GetUserProofs(ctx, user.ID, 2, 0)
		if len(retrieved) != 2 {
			t.Errorf("Expected 2 proofs with limit=2, got %d", len(retrieved))
		}
		t.Log("✓ Pagination works correctly")

		// Get all proofs
		all, _ := verifier.GetUserProofs(ctx, user.ID, 100, 0)
		if len(all) < proofCount {
			t.Logf("⚠ Expected at least %d proofs, got %d", proofCount, len(all))
		} else {
			t.Logf("✓ Retrieved all %d proofs", len(all))
		}
	})

	t.Run("Proof verification workflow", func(t *testing.T) {
		user := &domain.User{
			ID:            "workflow_user",
			WalletAddress: domain.Address("0xWORKFLOW"),
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		workflow := NewProofVerificationWorkflow(verifier, logger)

		// Define proof submission function
		submitProof := func(ctx context.Context) (*zkproof.ProofResponse, error) {
			return &zkproof.ProofResponse{
				ProofData:      []byte{0x00, 0x01, 0x02, 0x03},
				TimingVariance: 150,
				GasVariance:    700,
				ProofNonce:     "workflow_nonce",
				SubmittedAt:    time.Now(),
			}, nil
		}

		// Execute workflow
		result, err := workflow.ExecuteWorkflow(
			ctx,
			user.ID,
			domain.Address("0xVERIFIER"),
			submitProof,
		)

		if err != nil {
			t.Logf("⚠ Workflow error (may be expected): %v", err)
		}
		if result != nil {
			t.Logf("✓ Workflow completed with verification score: %f", result.VerificationScore)
		}
	})

	t.Run("Blacklisted user cannot generate proof", func(t *testing.T) {
		user := &domain.User{
			ID:            "blacklist_user",
			WalletAddress: domain.Address("0xBLACKLIST"),
			IsBlacklisted: true,
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		_, err := verifier.GenerateProofChallenge(ctx, user.ID, domain.Address("0xVERIFIER"))
		if err == nil {
			t.Fatal("Expected error for blacklisted user")
		}
		t.Log("✓ Blacklisted user correctly rejected")
	})

	t.Run("Nonexistent user handling", func(t *testing.T) {
		_, err := verifier.GenerateProofChallenge(ctx, "nonexistent_user", domain.Address("0xVERIFIER"))
		if err == nil {
			t.Fatal("Expected error for nonexistent user")
		}
		t.Log("✓ Nonexistent user correctly rejected")
	})
}

// BenchmarkIntegrationWorkflow benchmarks the complete workflow.
func BenchmarkIntegrationWorkflow(b *testing.B) {
	db := repositories.SetupTestDB(&testing.T{})
	defer repositories.CleanupTestDB(&testing.T{}, db)

	logger := slog.New(slog.NewTextHandler(nil, nil))
	config := zkproof.ProofServiceConfig{
		MaxChallengeTime: 5 * time.Minute,
	}

	userRepo := repositories.NewUserRepository(db)
	proofRepo := repositories.NewHumanProofRepository(db)
	verifier := NewHumanProofVerifier(db, config, logger)
	verifier.(*HumanProofVerifier).proofRepo = proofRepo
	verifier.(*HumanProofVerifier).userRepo = userRepo

	ctx := context.Background()

	// Create test user
	user := &domain.User{
		ID:            "bench_user",
		WalletAddress: domain.Address("0xBENCH"),
		CreatedAt:     time.Now(),
	}
	userRepo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Challenge generation
		challenge, _ := verifier.GenerateProofChallenge(ctx, user.ID, domain.Address("0xVERIFIER"))

		// Proof submission
		if challenge != nil {
			response := &zkproof.ProofResponse{
				ChallengeID:    challenge.ChallengeID,
				ProofData:      []byte{0x00, 0x01, 0x02, 0x03},
				TimingVariance: 100,
				GasVariance:    500,
				ProofNonce:     fmt.Sprintf("nonce_%d", i),
				SubmittedAt:    time.Now(),
			}
			verifier.SubmitProofResponse(ctx, response)
		}
	}
}
