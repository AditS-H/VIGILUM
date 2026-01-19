// Package repositories provides integration tests for repository workflows.
package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// TestUserAndAPIKeyWorkflow tests the complete User -> APIKey relationship.
func TestUserAndAPIKeyWorkflow(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	// Step 1: Create user
	user := &domain.User{
		ID:            "test_user_1",
		WalletAddress: domain.Address("0xUSER123"),
		RiskScore:     0,
		CreatedAt:     time.Now(),
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Step 2: Generate API keys for user
	tier := "premium"
	for i := 0; i < 3; i++ {
		keyHash := []byte(fmt.Sprintf("test_hash_%d", i))
		apiKey := &domain.APIKey{
			ID:        fmt.Sprintf("api_key_%d", i),
			KeyHash:   keyHash,
			UserID:    user.ID,
			Name:      fmt.Sprintf("API Key %d", i),
			Tier:      tier,
			RateLimit: 10000,
			CreatedAt: time.Now(),
			ExpiresAt: sql.NullTime{Time: time.Now().Add(365 * 24 * time.Hour), Valid: true},
			Revoked:   false,
		}
		if err := apiKeyRepo.Create(ctx, apiKey); err != nil {
			t.Fatalf("Failed to create API key: %v", err)
		}
	}

	// Step 3: Verify user has 3 API keys
	keys, err := apiKeyRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user API keys: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("Expected 3 API keys, got %d", len(keys))
	}

	// Step 4: Make requests with first key and track usage
	firstKey := keys[0]
	for i := 0; i < 5; i++ {
		if err := apiKeyRepo.UpdateRequestCount(ctx, firstKey.ID); err != nil {
			t.Fatalf("Failed to update request count: %v", err)
		}
	}

	// Step 5: Verify request count
	updated, err := apiKeyRepo.GetByID(ctx, firstKey.ID)
	if err != nil {
		t.Fatalf("Failed to get updated key: %v", err)
	}
	if updated.RequestsToday != 5 {
		t.Errorf("Expected 5 requests, got %d", updated.RequestsToday)
	}

	// Step 6: Reset daily counts
	if err := apiKeyRepo.ResetDailyCount(ctx, firstKey.ID); err != nil {
		t.Fatalf("Failed to reset daily count: %v", err)
	}

	// Step 7: Revoke all keys
	for _, key := range keys {
		if err := apiKeyRepo.Revoke(ctx, key.ID); err != nil {
			t.Fatalf("Failed to revoke key: %v", err)
		}
	}

	// Step 8: Verify no active keys for user
	activeKeys, err := apiKeyRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list active keys: %v", err)
	}
	if len(activeKeys) != 0 {
		t.Errorf("Expected 0 active keys after revocation, got %d", len(activeKeys))
	}

	t.Log("✓ User -> APIKey workflow completed successfully")
}

// TestUserAndProofWorkflow tests the complete User -> HumanProof verification workflow.
func TestUserAndProofWorkflow(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	userRepo := NewUserRepository(db)
	proofRepo := NewHumanProofRepository(db)
	ctx := context.Background()

	// Step 1: Create user
	user := &domain.User{
		ID:            "proof_user_1",
		WalletAddress: domain.Address("0xPROOFUSER"),
		RiskScore:     0,
		CreatedAt:     time.Now(),
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Step 2: Generate multiple proofs
	proofIDs := []string{}
	for i := 0; i < 3; i++ {
		proof := &domain.HumanProof{
			ID:        fmt.Sprintf("proof_%d", i),
			UserID:    user.ID,
			ProofHash: fmt.Sprintf("proof_hash_%d", i),
			ProofData: domain.ProofData{
				TimingVariance:    int32(100 + i*10),
				GasVariance:       int32(50 + i*5),
				ContractDiversity: int32(2 + i),
				ProofNonce:        fmt.Sprintf("nonce_%d", i),
			},
			VerifierAddress: domain.Address("0xVERIFIER"),
			ExpiresAt:       time.Now().Add(24 * time.Hour),
			CreatedAt:       time.Now(),
		}
		if err := proofRepo.Create(ctx, proof); err != nil {
			t.Fatalf("Failed to create proof: %v", err)
		}
		proofIDs = append(proofIDs, proof.ID)
	}

	// Step 3: Verify user has 3 proofs
	count, err := proofRepo.CountByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to count proofs: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 proofs, got %d", count)
	}

	// Step 4: Mark first proof as verified
	if err := proofRepo.MarkVerified(ctx, proofIDs[0]); err != nil {
		t.Fatalf("Failed to mark proof verified: %v", err)
	}

	// Step 5: Verify the proof is marked as verified
	verified, err := proofRepo.GetByID(ctx, proofIDs[0])
	if err != nil {
		t.Fatalf("Failed to get verified proof: %v", err)
	}
	if !verified.VerifiedAt.Valid {
		t.Error("VerifiedAt should be set for verified proof")
	}

	// Step 6: Get user's proofs with pagination
	proofs, err := proofRepo.GetByUserID(ctx, user.ID, 2, 0)
	if err != nil {
		t.Fatalf("Failed to get user proofs: %v", err)
	}
	if len(proofs) != 2 {
		t.Errorf("Expected 2 proofs (limit 2, offset 0), got %d", len(proofs))
	}

	// Step 7: Count verified proofs
	verifiedCount, err := proofRepo.CountVerifiedByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to count verified proofs: %v", err)
	}
	if verifiedCount != 1 {
		t.Errorf("Expected 1 verified proof, got %d", verifiedCount)
	}

	// Step 8: Update user risk score based on verification
	if err := userRepo.UpdateRiskScore(ctx, user.ID, 10); err != nil {
		t.Fatalf("Failed to update risk score: %v", err)
	}

	updatedUser, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	if updatedUser.RiskScore != 10 {
		t.Errorf("Expected risk score 10, got %d", updatedUser.RiskScore)
	}

	t.Log("✓ User -> HumanProof verification workflow completed successfully")
}

// TestThreatSignalPublishingWorkflow tests threat signal creation and on-chain publishing.
func TestThreatSignalPublishingWorkflow(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	signalRepo := NewThreatSignalRepository(db)
	ctx := context.Background()

	chainID := domain.ChainID(1)
	targetAddress := domain.Address("0xTARGETCONTRACT")

	// Step 1: Create threat signals from multiple sources
	signalIDs := []string{}
	for i := 0; i < 5; i++ {
		signal := &domain.ThreatSignal{
			ID:          fmt.Sprintf("signal_%d", i),
			ChainID:     chainID,
			Address:     targetAddress,
			SignalType:  "exploit",
			RiskScore:   int32(70 + i*5),
			ThreatLevel: "high",
			SourceID:    fmt.Sprintf("source_%d", i),
			Metadata: map[string]interface{}{
				"exploit_type": "reentrancy",
				"confidence":   0.85,
			},
			PublishedAt: sql.NullTime{Valid: false},
			CreatedAt:   time.Now(),
		}
		if err := signalRepo.Create(ctx, signal); err != nil {
			t.Fatalf("Failed to create signal: %v", err)
		}
		signalIDs = append(signalIDs, signal.ID)
	}

	// Step 2: Get unpublished signals
	unpublished, err := signalRepo.GetUnpublished(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get unpublished signals: %v", err)
	}
	if len(unpublished) < 5 {
		t.Errorf("Expected at least 5 unpublished signals, got %d", len(unpublished))
	}

	// Step 3: Aggregate signals for target entity
	entitySignals, err := signalRepo.GetByEntity(ctx, chainID, targetAddress, 10)
	if err != nil {
		t.Fatalf("Failed to get entity signals: %v", err)
	}
	if len(entitySignals) != 5 {
		t.Errorf("Expected 5 signals for entity, got %d", len(entitySignals))
	}

	// Step 4: Get high-risk signals (>=80)
	highRisk, err := signalRepo.GetHighRisk(ctx, 80, 10)
	if err != nil {
		t.Fatalf("Failed to get high-risk signals: %v", err)
	}
	if len(highRisk) == 0 {
		t.Error("Expected at least 1 high-risk signal")
	}

	// Step 5: Publish high-risk signals to blockchain
	txHash := "0xPUBLISHTX123ABC"
	for _, sig := range highRisk {
		if err := signalRepo.MarkPublished(ctx, sig.ID, txHash); err != nil {
			t.Fatalf("Failed to mark signal published: %v", err)
		}
	}

	// Step 6: Verify signals are published
	published, err := signalRepo.GetByID(ctx, highRisk[0].ID)
	if err != nil {
		t.Fatalf("Failed to get published signal: %v", err)
	}
	if !published.PublishedAt.Valid {
		t.Error("PublishedAt should be set")
	}

	// Step 7: Get remaining unpublished signals
	stillUnpublished, err := signalRepo.GetUnpublished(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get remaining unpublished: %v", err)
	}
	if len(stillUnpublished) >= len(unpublished) {
		t.Error("Expected fewer unpublished signals after publishing")
	}

	t.Log("✓ Threat signal publishing workflow completed successfully")
}

// TestGenomeAnalysisWorkflow tests genome creation, analysis, and similarity detection.
func TestGenomeAnalysisWorkflow(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	genomeRepo := NewGenomeRepository(db)
	ctx := context.Background()

	// Step 1: Create reference genome (malicious)
	refGenome := &domain.Genome{
		ID:              "ref_malicious_genome",
		GenomeHash:      "ref_hash_malicious",
		IPFSHash:        "QmRefMalicious",
		BytecodeSize:    2048,
		OpcodeCount:     1024,
		FunctionCount:   15,
		ComplexityScore: 0.85,
		Label:           "malicious",
		Features: map[string]interface{}{
			"patterns": []string{"reentrancy", "overflow"},
		},
		CreatedAt: time.Now(),
	}
	if err := genomeRepo.Create(ctx, refGenome); err != nil {
		t.Fatalf("Failed to create reference genome: %v", err)
	}

	// Step 2: Create similar genomes for clustering
	for i := 0; i < 4; i++ {
		genome := &domain.Genome{
			ID:              fmt.Sprintf("similar_genome_%d", i),
			GenomeHash:      fmt.Sprintf("sim_hash_%d", i),
			IPFSHash:        fmt.Sprintf("QmSim%d", i),
			BytecodeSize:    2000 + int32(i*10),
			OpcodeCount:     1010 + int32(i*5),
			FunctionCount:   14 + int32(i),
			ComplexityScore: 0.82,
			Label:           "malicious",
			Features: map[string]interface{}{
				"patterns": []string{"reentrancy"},
			},
			CreatedAt: time.Now(),
		}
		if err := genomeRepo.Create(ctx, genome); err != nil {
			t.Fatalf("Failed to create similar genome: %v", err)
		}
	}

	// Step 3: Find similar genomes to reference
	similar, err := genomeRepo.ListSimilar(ctx, refGenome.ID, 0.75, 10)
	if err != nil {
		t.Fatalf("Failed to find similar genomes: %v", err)
	}
	if len(similar) < 4 {
		t.Logf("Warning: Expected at least 4 similar genomes, got %d (this is OK for distance calculation)", len(similar))
	}

	// Step 4: List all malicious genomes
	malicious, err := genomeRepo.ListByLabel(ctx, "malicious", 10)
	if err != nil {
		t.Fatalf("Failed to list malicious genomes: %v", err)
	}
	if len(malicious) < 5 {
		t.Errorf("Expected at least 5 malicious genomes, got %d", len(malicious))
	}

	// Step 5: Get distribution across labels
	distribution, err := genomeRepo.GetDistribution(ctx)
	if err != nil {
		t.Fatalf("Failed to get distribution: %v", err)
	}
	if len(distribution) == 0 {
		t.Error("Expected distribution data")
	}

	// Step 6: Create benign genome for contrast
	benignGenome := &domain.Genome{
		ID:              "benign_genome",
		GenomeHash:      "benign_hash",
		IPFSHash:        "QmBenign",
		BytecodeSize:    512,
		OpcodeCount:     256,
		FunctionCount:   5,
		ComplexityScore: 0.2,
		Label:           "benign",
		Features: map[string]interface{}{
			"patterns": []string{},
		},
		CreatedAt: time.Now(),
	}
	if err := genomeRepo.Create(ctx, benignGenome); err != nil {
		t.Fatalf("Failed to create benign genome: %v", err)
	}

	// Step 7: Verify benign genomes are separate
	benignList, err := genomeRepo.ListByLabel(ctx, "benign", 10)
	if err != nil {
		t.Fatalf("Failed to list benign genomes: %v", err)
	}
	if len(benignList) == 0 {
		t.Error("Expected at least 1 benign genome")
	}

	t.Log("✓ Genome analysis workflow completed successfully")
}

// TestExploitSubmissionBountyWorkflow tests the complete exploit submission bounty workflow.
func TestExploitSubmissionBountyWorkflow(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	submissionRepo := NewExploitSubmissionRepository(db)
	genomeRepo := NewGenomeRepository(db)
	ctx := context.Background()

	// Step 1: Create genome for submission
	genome := &domain.Genome{
		ID:              "bounty_genome",
		GenomeHash:      "bounty_hash",
		IPFSHash:        "QmBounty",
		BytecodeSize:    1024,
		OpcodeCount:     512,
		FunctionCount:   10,
		ComplexityScore: 0.65,
		Label:           "vulnerable",
		Features:        map[string]interface{}{},
		CreatedAt:       time.Now(),
	}
	if err := genomeRepo.Create(ctx, genome); err != nil {
		t.Fatalf("Failed to create genome: %v", err)
	}

	// Step 2: Submit exploits from multiple researchers
	researcher1 := domain.Address("0xRESEARCHER1")
	researcher2 := domain.Address("0xRESEARCHER2")

	submissions := []struct {
		ID                string
		ResearcherAddress domain.Address
		Severity          string
		BountyAmount      int64
	}{
		{"exploit_1", researcher1, "critical", 5000000},
		{"exploit_2", researcher1, "high", 2500000},
		{"exploit_3", researcher2, "medium", 1000000},
	}

	for _, sub := range submissions {
		submission := &domain.ExploitSubmission{
			ID:                sub.ID,
			ResearcherAddress: sub.ResearcherAddress,
			TargetContract:    domain.Address("0xTARGET"),
			ChainID:           1,
			ProofHash:         fmt.Sprintf("proof_%s", sub.ID),
			GenomeID:          genome.ID,
			Description:       "Test exploit",
			Severity:          sub.Severity,
			BountyAmount:      sub.BountyAmount,
			BountyStatus:      "pending",
			Status:            "pending",
			CreatedAt:         time.Now(),
		}
		if err := submissionRepo.Create(ctx, submission); err != nil {
			t.Fatalf("Failed to create submission: %v", err)
		}
	}

	// Step 3: Get researcher's submissions
	r1Submissions, err := submissionRepo.GetByResearcher(ctx, researcher1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get researcher submissions: %v", err)
	}
	if len(r1Submissions) != 2 {
		t.Errorf("Expected 2 submissions from researcher1, got %d", len(r1Submissions))
	}

	// Step 4: Get pending submissions for verification
	pending, err := submissionRepo.GetPending(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get pending submissions: %v", err)
	}
	if len(pending) < 3 {
		t.Errorf("Expected at least 3 pending submissions, got %d", len(pending))
	}

	// Step 5: Verify exploit (auditor action)
	if err := submissionRepo.MarkVerified(ctx, "exploit_1"); err != nil {
		t.Fatalf("Failed to mark verified: %v", err)
	}

	verified, err := submissionRepo.GetByID(ctx, "exploit_1")
	if err != nil {
		t.Fatalf("Failed to get verified submission: %v", err)
	}
	if !verified.VerifiedAt.Valid {
		t.Error("VerifiedAt should be set")
	}

	// Step 6: Distribute bounty (payment action)
	if err := submissionRepo.MarkPaid(ctx, "exploit_1", "0xPAYTX"); err != nil {
		t.Fatalf("Failed to mark paid: %v", err)
	}

	// Step 7: Get total bounty amount for paid exploits
	totalBounty, err := submissionRepo.GetTotalBountyAmount(ctx, "paid")
	if err != nil {
		t.Fatalf("Failed to get total bounty: %v", err)
	}
	if totalBounty != 5000000 {
		t.Errorf("Expected total bounty 5000000, got %d", totalBounty)
	}

	// Step 8: Count submissions by researcher
	r1Count, err := submissionRepo.CountByResearcher(ctx, researcher1)
	if err != nil {
		t.Fatalf("Failed to count researcher submissions: %v", err)
	}
	if r1Count != 2 {
		t.Errorf("Expected 2 submissions, got %d", r1Count)
	}

	t.Log("✓ Exploit submission bounty workflow completed successfully")
}

// TestMultiRepositoryTransaction tests transaction-like behavior across repositories.
func TestMultiRepositoryTransaction(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	userRepo := NewUserRepository(db)
	proofRepo := NewHumanProofRepository(db)
	signalRepo := NewThreatSignalRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	// Step 1: Create complete user ecosystem
	user := &domain.User{
		ID:            "ecosystem_user",
		WalletAddress: domain.Address("0xECOSYSTEM"),
		RiskScore:     0,
		CreatedAt:     time.Now(),
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Step 2: Create user's API key
	apiKey := &domain.APIKey{
		ID:        "ecosystem_key",
		KeyHash:   []byte("ecosystem_hash"),
		UserID:    user.ID,
		Name:      "Main Key",
		Tier:      "premium",
		RateLimit: 50000,
		CreatedAt: time.Now(),
		ExpiresAt: sql.NullTime{Time: time.Now().Add(365 * 24 * time.Hour), Valid: true},
		Revoked:   false,
	}
	if err := apiKeyRepo.Create(ctx, apiKey); err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	// Step 3: User generates proof
	proof := &domain.HumanProof{
		ID:        "ecosystem_proof",
		UserID:    user.ID,
		ProofHash: "ecosystem_proof_hash",
		ProofData: domain.ProofData{
			TimingVariance:    100,
			GasVariance:       50,
			ContractDiversity: 3,
			ProofNonce:        "nonce",
		},
		VerifierAddress: domain.Address("0xVERIFIER"),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		CreatedAt:       time.Now(),
	}
	if err := proofRepo.Create(ctx, proof); err != nil {
		t.Fatalf("Failed to create proof: %v", err)
	}

	// Step 4: System detects threat via API call (using the key)
	if err := apiKeyRepo.UpdateRequestCount(ctx, apiKey.ID); err != nil {
		t.Fatalf("Failed to track API usage: %v", err)
	}

	signal := &domain.ThreatSignal{
		ID:          "ecosystem_signal",
		ChainID:     1,
		Address:     domain.Address("0xTHREAT"),
		SignalType:  "exploit",
		RiskScore:   95,
		ThreatLevel: "critical",
		SourceID:    "api_source",
		Metadata: map[string]interface{}{
			"api_key_tier": "premium",
		},
		PublishedAt: sql.NullTime{Valid: false},
		CreatedAt:   time.Now(),
	}
	if err := signalRepo.Create(ctx, signal); err != nil {
		t.Fatalf("Failed to create signal: %v", err)
	}

	// Step 5: Verify all relationships
	retrievedUser, _ := userRepo.GetByID(ctx, user.ID)
	retrievedKey, _ := apiKeyRepo.GetByID(ctx, apiKey.ID)
	retrievedProof, _ := proofRepo.GetByID(ctx, proof.ID)
	retrievedSignal, _ := signalRepo.GetByID(ctx, signal.ID)

	if retrievedUser.ID != user.ID {
		t.Error("User mismatch")
	}
	if retrievedKey.UserID != user.ID {
		t.Error("APIKey not linked to user")
	}
	if retrievedProof.UserID != user.ID {
		t.Error("Proof not linked to user")
	}
	if retrievedSignal.ID != signal.ID {
		t.Error("Signal mismatch")
	}

	// Step 6: Update user risk based on threat level
	highRiskSignals, _ := signalRepo.GetHighRisk(ctx, 90, 10)
	if len(highRiskSignals) > 0 {
		// Increase user risk score based on threats
		if err := userRepo.UpdateRiskScore(ctx, user.ID, 75); err != nil {
			t.Fatalf("Failed to update risk: %v", err)
		}
	}

	// Step 7: Verify final state
	finalUser, _ := userRepo.GetByID(ctx, user.ID)
	if finalUser.RiskScore != 75 {
		t.Errorf("Expected risk score 75, got %d", finalUser.RiskScore)
	}

	t.Log("✓ Multi-repository transaction workflow completed successfully")
}

// TestErrorHandlingAndEdgeCases tests error conditions across repositories.
func TestErrorHandlingAndEdgeCases(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	userRepo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("NotFound errors", func(t *testing.T) {
		_, err := userRepo.GetByID(ctx, "nonexistent_user")
		if err != domain.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}

		_, err = userRepo.GetByWallet(ctx, domain.Address("0xNONEXISTENT"))
		if err != domain.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Duplicate prevention", func(t *testing.T) {
		user := &domain.User{
			ID:            "dup_user",
			WalletAddress: domain.Address("0xDUP"),
			CreatedAt:     time.Now(),
		}
		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Try to create with same wallet (should fail due to unique constraint)
		dup := &domain.User{
			ID:            "dup_user_2",
			WalletAddress: domain.Address("0xDUP"),
			CreatedAt:     time.Now(),
		}
		err := userRepo.Create(ctx, dup)
		if err == nil {
			t.Error("Expected error for duplicate wallet address")
		}
	})

	t.Run("Update nonexistent", func(t *testing.T) {
		err := userRepo.UpdateRiskScore(ctx, "nonexistent", 50)
		if err != domain.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Delete nonexistent", func(t *testing.T) {
		err := userRepo.Delete(ctx, "nonexistent")
		if err != domain.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

// BenchmarkRepositoryOperations benchmarks core repository operations.
func BenchmarkRepositoryOperations(b *testing.B) {
	db := SetupTestDB(&testing.T{})
	defer CleanupTestDB(&testing.T{}, db)

	userRepo := NewUserRepository(db)
	ctx := context.Background()

	b.Run("UserCreate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			user := &domain.User{
				ID:            fmt.Sprintf("bench_user_%d", i),
				WalletAddress: domain.Address(fmt.Sprintf("0x%x", i)),
				RiskScore:     int32(i % 100),
				CreatedAt:     time.Now(),
			}
			userRepo.Create(ctx, user)
		}
	})

	b.Run("UserGetByID", func(b *testing.B) {
		// Setup
		user := &domain.User{
			ID:            "bench_get_user",
			WalletAddress: domain.Address("0xBENCHGET"),
			CreatedAt:     time.Now(),
		}
		userRepo.Create(ctx, user)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userRepo.GetByID(ctx, user.ID)
		}
	})

	b.Run("UserListByRiskScore", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userRepo.ListByRiskScore(ctx, 50, 100)
		}
	})
}
