// Package repositories tests all repository implementations.
package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// TestUserRepository tests the UserRepository implementation.
func TestUserRepository(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve user", func(t *testing.T) {
		user := &domain.User{
			ID:             "user1",
			WalletAddress: "0x1234567890",
			RiskScore:      50,
			IsBlacklisted:  false,
			CreatedAt:      time.Now(),
		}

		_, err := repo.Create(ctx, user.WalletAddress)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.WalletAddress != user.WalletAddress {
			t.Errorf("wallet mismatch: expected %s, got %s", user.WalletAddress, retrieved.WalletAddress)
		}
	})

	t.Run("GetByWallet", func(t *testing.T) {
		user := &domain.User{
			ID:            "user2",
			WalletAddress: "0xABCDEF",
			RiskScore:     75,
			CreatedAt:     time.Now(),
		}
		repo.Create(ctx, user.WalletAddress)

		retrieved, err := repo.GetByWallet(ctx, user.WalletAddress)
		if err != nil {
			t.Fatalf("GetByWallet failed: %v", err)
		}

		if retrieved.ID != user.ID {
			t.Errorf("ID mismatch: expected %s, got %s", user.ID, retrieved.ID)
		}
	})

	t.Run("Update risk score", func(t *testing.T) {
		user := &domain.User{
			ID:            "user3",
			WalletAddress: "0x3333",
			RiskScore:     10,
			CreatedAt:     time.Now(),
		}
		repo.Create(ctx, user.WalletAddress)

		err := repo.UpdateRiskScore(ctx, user.ID, 99)
		if err != nil {
			t.Fatalf("UpdateRiskScore failed: %v", err)
		}

		updated, _ := repo.GetByID(ctx, user.ID)
		if updated.RiskScore != 99 {
			t.Errorf("risk score not updated: expected 99, got %d", updated.RiskScore)
		}
	})

	t.Run("Blacklist operations", func(t *testing.T) {
		user := &domain.User{
			ID:            "user4",
			WalletAddress: "0x4444",
			RiskScore:     0,
			CreatedAt:     time.Now(),
		}
		repo.Create(ctx, user.WalletAddress)

		err := repo.Blacklist(ctx, user.ID)
		if err != nil {
			t.Fatalf("Blacklist failed: %v", err)
		}

		blacklisted, _ := repo.GetByID(ctx, user.ID)
		if !blacklisted.IsBlacklisted {
			t.Error("user not blacklisted")
		}

		repo.RemoveBlacklist(ctx, user.ID)
		unblacklisted, _ := repo.GetByID(ctx, user.ID)
		if unblacklisted.IsBlacklisted {
			t.Error("user still blacklisted after removal")
		}
	})

	t.Run("ListByRiskScore", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			user := &domain.User{
				ID:            fmt.Sprintf("user_risk_%d", i),
				WalletAddress: fmt.Sprintf("0x%d", i),
				RiskScore: float64(i * 20),
				CreatedAt:     time.Now(),
			}
			repo.Create(ctx, user.WalletAddress)
		}

		users, err := repo.ListByRiskScore(ctx, 50, 10)
		if err != nil {
			t.Fatalf("ListByRiskScore failed: %v", err)
		}

		if len(users) == 0 {
			t.Error("expected users with risk score >= 50")
		}

		for _, u := range users {
			if u.RiskScore < 50 {
				t.Errorf("returned user with risk score below threshold: %d", u.RiskScore)
			}
		}
	})

	t.Run("Delete user", func(t *testing.T) {
		user := &domain.User{
			ID:            "user_delete",
			WalletAddress: "0xDELETE",
			CreatedAt:     time.Now(),
		}
		repo.Create(ctx, user.WalletAddress)

		err := repo.Delete(ctx, user.ID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, user.ID)
		if err != domain.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// TestHumanProofRepository tests HumanProofRepository implementation.
func TestHumanProofRepository(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewHumanProofRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve proof", func(t *testing.T) {
		proof := &domain.HumanProof{
			ID:              "proof1",
			UserID:          "user1",
			ProofHash:       "hash123",
			ProofData: &domain.ProofData{
				TimingVariance:     100,
				GasVariance:        50,
				ContractDiversity:  3,
				ProofNonce: 1,
			},
			VerifierAddress: "0xVERIFIER",
			ExpiresAt: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			CreatedAt:       time.Now(),
		}

		err := repo.Create(ctx, proof)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, proof.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.ProofHash != proof.ProofHash {
			t.Errorf("proof hash mismatch")
		}
	})

	t.Run("GetByUserID with pagination", func(t *testing.T) {
		userID := "user_proof"
		for i := 0; i < 5; i++ {
			proof := &domain.HumanProof{
				ID:              fmt.Sprintf("proof_%d", i),
				UserID:          userID,
				ProofHash:       fmt.Sprintf("hash%d", i),
				ProofData:       domain.ProofData{TimingVariance: int32(i * 10)},
				ExpiresAt: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
				CreatedAt:       time.Now(),
			}
			repo.Create(ctx, proof)
		}

		proofs, err := repo.GetByUserID(ctx, userID, 2, 0)
		if err != nil {
			t.Fatalf("GetByUserID failed: %v", err)
		}

		if len(proofs) != 2 {
			t.Errorf("expected 2 proofs, got %d", len(proofs))
		}
	})

	t.Run("MarkVerified", func(t *testing.T) {
		proof := &domain.HumanProof{
			ID:              "proof_verify",
			UserID:          "user_verify",
			ProofHash:       "verify_hash",
			ProofData:       domain.ProofData{},
			VerifierAddress: "0x0",
			ExpiresAt: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			CreatedAt:       time.Now(),
		}
		repo.Create(ctx, proof)

		err := repo.MarkVerified(ctx, proof.ID)
		if err != nil {
			t.Fatalf("MarkVerified failed: %v", err)
		}

		verified, _ := repo.GetByID(ctx, proof.ID)
		if !verified.VerifiedAt.Valid {
			t.Error("VerifiedAt not set")
		}
	})

	t.Run("CountByUserID", func(t *testing.T) {
		userID := "count_user"
		for i := 0; i < 3; i++ {
			proof := &domain.HumanProof{
				ID:        fmt.Sprintf("count_proof_%d", i),
				UserID:    userID,
				ProofHash: fmt.Sprintf("ch%d", i),
				ProofData: &domain.ProofData{},
				ExpiresAt: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
				CreatedAt: time.Now(),
			}
			repo.Create(ctx, proof)
		}

		count, err := repo.CountByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("CountByUserID failed: %v", err)
		}

		if count < 3 {
			t.Errorf("expected at least 3 proofs, got %d", count)
		}
	})
}

// TestThreatSignalRepository tests ThreatSignalRepository implementation.
func TestThreatSignalRepository(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewThreatSignalRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve signal", func(t *testing.T) {
		signal := &domain.ThreatSignal{
			ID:           "signal1",
			ChainID:      1,
			Address:      domain.Address("0xTARGET"),
			SignalType:   "exploit",
			RiskScore:    85,
			ThreatLevel:  "critical",
			SourceID:     "source1",
			Metadata:     map[string]interface{}{"reason": "detected_pattern"},
			PublishedAt:  sql.NullTime{},
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, signal)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, signal.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.ThreatLevel != signal.ThreatLevel {
			t.Errorf("threat level mismatch")
		}
	})

	t.Run("GetByEntity", func(t *testing.T) {
		chainID := domain.ChainID(137)
		addr := domain.Address("0xENTITY")
		for i := 0; i < 3; i++ {
			signal := &domain.ThreatSignal{
				ID:          fmt.Sprintf("signal_entity_%d", i),
				ChainID:     chainID,
				Address:     addr,
				SignalType:  "malware",
				RiskScore: float64(70 + i*5),
				ThreatLevel: "high",
				SourceID:    fmt.Sprintf("src%d", i),
				Metadata:    map[string]interface{}{},
				CreatedAt:   time.Now(),
			}
			repo.Create(ctx, signal)
		}

		signals, err := repo.GetByEntity(ctx, chainID, addr, 10)
		if err != nil {
			t.Fatalf("GetByEntity failed: %v", err)
		}

		if len(signals) < 3 {
			t.Errorf("expected at least 3 signals, got %d", len(signals))
		}
	})

	t.Run("GetUnpublished", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			signal := &domain.ThreatSignal{
				ID:          fmt.Sprintf("unpub_%d", i),
				ChainID:     1,
				Address:     domain.Address(fmt.Sprintf("0x%d", i)),
				SignalType:  "vulnerability",
				RiskScore:   90,
				ThreatLevel: "critical",
				SourceID:    "source",
				Metadata:    map[string]interface{}{},
				PublishedAt: sql.NullTime{Valid: false},
				CreatedAt:   time.Now(),
			}
			repo.Create(ctx, signal)
		}

		unpub, err := repo.GetUnpublished(ctx, 5)
		if err != nil {
			t.Fatalf("GetUnpublished failed: %v", err)
		}

		if len(unpub) == 0 {
			t.Error("expected unpublished signals")
		}
	})

	t.Run("MarkPublished", func(t *testing.T) {
		signal := &domain.ThreatSignal{
			ID:          "pub_signal",
			ChainID:     1,
			Address:     domain.Address("0xPUB"),
			SignalType:  "warning",
			RiskScore:   50,
			ThreatLevel: "medium",
			SourceID:    "src",
			Metadata:    map[string]interface{}{},
			CreatedAt:   time.Now(),
		}
		repo.Create(ctx, signal)

		txHash := "0xPUBLISHTX"
		err := repo.MarkPublished(ctx, signal.ID, txHash)
		if err != nil {
			t.Fatalf("MarkPublished failed: %v", err)
		}

		pub, _ := repo.GetByID(ctx, signal.ID)
		if !pub.PublishedAt.Valid {
			t.Error("PublishedAt not set")
		}
	})

	t.Run("GetHighRisk", func(t *testing.T) {
		signal := &domain.ThreatSignal{
			ID:          "high_risk",
			ChainID:     1,
			Address:     domain.Address("0xHIGH"),
			SignalType:  "exploit",
			RiskScore:   95,
			ThreatLevel: "critical",
			SourceID:    "src",
			Metadata:    map[string]interface{}{},
			CreatedAt:   time.Now(),
		}
		repo.Create(ctx, signal)

		high, err := repo.GetHighRisk(ctx, 90, 5)
		if err != nil {
			t.Fatalf("GetHighRisk failed: %v", err)
		}

		if len(high) == 0 {
			t.Error("expected high risk signals")
		}
	})
}

// TestGenomeRepository tests GenomeRepository implementation.
func TestGenomeRepository(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewGenomeRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve genome", func(t *testing.T) {
		genome := &domain.Genome{
			ID:              "genome1",
			GenomeHash:      "ghash123",
			IPFSHash:        "Qmabc123",
			BytecodeSize:    1024,
			OpcodeCount:     512,
			FunctionCount:   10,
			ComplexityScore: 0.75,
			Label:           "malicious",
			Features: map[string]interface{}{
				"patterns": []string{"pattern1", "pattern2"},
			},
			CreatedAt: time.Now(),
		}

		err := repo.Create(ctx, genome)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, genome.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.GenomeHash != genome.GenomeHash {
			t.Errorf("genome hash mismatch")
		}
	})

	t.Run("ListByLabel", func(t *testing.T) {
		label := "benign"
		for i := 0; i < 4; i++ {
			genome := &domain.Genome{
				ID:              fmt.Sprintf("genome_label_%d", i),
				GenomeHash:      fmt.Sprintf("gh%d", i),
				IPFSHash:        fmt.Sprintf("Qm%d", i),
				BytecodeSize:    512 + int32(i*100),
				OpcodeCount:     256 + int32(i*50),
				FunctionCount:   5 + int32(i),
				ComplexityScore: 0.5,
				Label:           label,
				Features:        map[string]interface{}{},
				CreatedAt:       time.Now(),
			}
			repo.Create(ctx, genome)
		}

		genomes, err := repo.ListByLabel(ctx, label, 10)
		if err != nil {
			t.Fatalf("ListByLabel failed: %v", err)
		}

		if len(genomes) < 4 {
			t.Errorf("expected at least 4 genomes, got %d", len(genomes))
		}
	})

	t.Run("GetByContractAddress", func(t *testing.T) {
		addr := domain.Address("0xCONTRACT")
		genome := &domain.Genome{
			ID:              "genome_contract",
			GenomeHash:      "gch_addr",
			IPFSHash:        "Qm_addr",
			ContractAddress: addr,
			BytecodeSize:    2048,
			OpcodeCount:     1024,
			FunctionCount:   20,
			ComplexityScore: 0.85,
			Label:           "monitored",
			Features:        map[string]interface{}{},
			CreatedAt:       time.Now(),
		}
		repo.Create(ctx, genome)

		retrieved, err := repo.GetByContractAddress(ctx, addr)
		if err != nil {
			t.Fatalf("GetByContractAddress failed: %v", err)
		}

		if retrieved.ID != genome.ID {
			t.Errorf("ID mismatch")
		}
	})

	t.Run("ListSimilar", func(t *testing.T) {
		refGenome := &domain.Genome{
			ID:              "ref_genome",
			GenomeHash:      "ref_hash",
			IPFSHash:        "Qmref",
			BytecodeSize:    1000,
			OpcodeCount:     500,
			FunctionCount:   15,
			ComplexityScore: 0.70,
			Label:           "suspicious",
			Features:        map[string]interface{}{},
			CreatedAt:       time.Now(),
		}
		repo.Create(ctx, refGenome)

		// Create similar genome
		similar := &domain.Genome{
			ID:              "similar_genome",
			GenomeHash:      "sim_hash",
			IPFSHash:        "Qmsim",
			BytecodeSize:    1050,
			OpcodeCount:     510,
			FunctionCount:   14,
			ComplexityScore: 0.68,
			Label:           "suspicious",
			Features:        map[string]interface{}{},
			CreatedAt:       time.Now(),
		}
		repo.Create(ctx, similar)

		genomes, err := repo.ListSimilar(ctx, refGenome.ID, 0.6, 5)
		if err != nil {
			t.Fatalf("ListSimilar failed: %v", err)
		}

		// Should find at least the reference itself
		if len(genomes) == 0 {
			t.Error("expected similar genomes")
		}
	})

	t.Run("GetDistribution", func(t *testing.T) {
		labels := []string{"malicious", "benign", "unknown"}
		for _, label := range labels {
			for i := 0; i < 2; i++ {
				genome := &domain.Genome{
					ID:              fmt.Sprintf("dist_%s_%d", label, i),
					GenomeHash:      fmt.Sprintf("dh_%s_%d", label, i),
					IPFSHash:        fmt.Sprintf("Qdm_%s_%d", label, i),
					BytecodeSize:    512,
					OpcodeCount:     256,
					FunctionCount:   8,
					ComplexityScore: 0.5,
					Label:           label,
					Features:        map[string]interface{}{},
					CreatedAt:       time.Now(),
				}
				repo.Create(ctx, genome)
			}
		}

		dist, err := repo.GetDistribution(ctx)
		if err != nil {
			t.Fatalf("GetDistribution failed: %v", err)
		}

		if len(dist) == 0 {
			t.Error("expected distribution data")
		}
	})
}

// TestExploitSubmissionRepository tests ExploitSubmissionRepository implementation.
func TestExploitSubmissionRepository(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewExploitSubmissionRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve submission", func(t *testing.T) {
		submission := &domain.ExploitSubmission{
			ID:              "exploit1",
			ResearcherAddress: domain.Address("0xRESEARCHER"),
			TargetContract:    domain.Address("0xTARGET"),
			ChainID:           1,
			ProofHash:         "proof123",
			GenomeID:          "genome1",
			Description:       "Reentrancy vulnerability",
			Severity:          "critical",
			BountyAmount:      1000000,
			BountyStatus:      "pending",
			Status:            "pending",
			CreatedAt:         time.Now(),
		}

		err := repo.Create(ctx, submission)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, submission.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.Severity != submission.Severity {
			t.Errorf("severity mismatch")
		}
	})

	t.Run("GetByResearcher", func(t *testing.T) {
		researcher := domain.Address("0xRES2")
		for i := 0; i < 3; i++ {
			submission := &domain.ExploitSubmission{
				ID:                fmt.Sprintf("exploit_res_%d", i),
				ResearcherAddress: researcher,
				TargetContract:    domain.Address(fmt.Sprintf("0x%d", i)),
				ChainID:           1,
				ProofHash:         fmt.Sprintf("ph%d", i),
				GenomeID:          "gen",
				Description:       "test",
				Severity:          "high",
				BountyAmount:      500000,
				BountyStatus:      "pending",
				Status:            "pending",
				CreatedAt:         time.Now(),
			}
			repo.Create(ctx, submission)
		}

		submissions, err := repo.GetByResearcher(ctx, researcher, 10, 0)
		if err != nil {
			t.Fatalf("GetByResearcher failed: %v", err)
		}

		if len(submissions) < 3 {
			t.Errorf("expected at least 3 submissions, got %d", len(submissions))
		}
	})

	t.Run("MarkVerified and MarkPaid", func(t *testing.T) {
		submission := &domain.ExploitSubmission{
			ID:                "exploit_verify",
			ResearcherAddress: domain.Address("0xRES3"),
			TargetContract:    domain.Address("0xTARG3"),
			ChainID:           1,
			ProofHash:         "pv123",
			GenomeID:          "gen",
			Description:       "test",
			Severity:          "medium",
			BountyAmount:      250000,
			BountyStatus:      "pending",
			Status:            "pending",
			CreatedAt:         time.Now(),
		}
		repo.Create(ctx, submission)

		err := repo.MarkVerified(ctx, submission.ID)
		if err != nil {
			t.Fatalf("MarkVerified failed: %v", err)
		}

		verified, _ := repo.GetByID(ctx, submission.ID)
		if !verified.VerifiedAt.Valid {
			t.Error("VerifiedAt not set")
		}

		err = repo.MarkPaid(ctx, submission.ID, "0xPAYTX")
		if err != nil {
			t.Fatalf("MarkPaid failed: %v", err)
		}

		paid, _ := repo.GetByID(ctx, submission.ID)
		if !paid.PaidAt.Valid {
			t.Error("PaidAt not set")
		}
	})

	t.Run("GetPending", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			submission := &domain.ExploitSubmission{
				ID:                fmt.Sprintf("pending_%d", i),
				ResearcherAddress: domain.Address("0xPEND"),
				TargetContract:    domain.Address("0xPT"),
				ChainID:           1,
				ProofHash:         fmt.Sprintf("pending%d", i),
				GenomeID:          "gen",
				Description:       "test",
				Severity:          "low",
				BountyAmount:      100000,
				BountyStatus:      "pending",
				Status:            "pending",
				CreatedAt:         time.Now(),
			}
			repo.Create(ctx, submission)
		}

		pending, err := repo.GetPending(ctx, 10)
		if err != nil {
			t.Fatalf("GetPending failed: %v", err)
		}

		if len(pending) == 0 {
			t.Error("expected pending submissions")
		}
	})
}

// TestAPIKeyRepository tests APIKeyRepository implementation.
func TestAPIKeyRepository(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewAPIKeyRepository(db)
	ctx := context.Background()

	t.Run("Create and retrieve key", func(t *testing.T) {
		keyHash := []byte("hash123456789")
		key := &domain.APIKey{
			ID:           "key1",
			KeyHash:      keyHash,
			UserID:       "user1",
			Name:         "Production API Key",
			Tier:         "premium",
			RateLimit:    10000,
			CreatedAt:    time.Now(),
			ExpiresAt:    sql.NullTime{Time: time.Now().Add(365 * 24 * time.Hour), Valid: true},
			Revoked:      false,
		}

		err := repo.Create(ctx, key)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetByHash(ctx, keyHash)
		if err != nil {
			t.Fatalf("GetByHash failed: %v", err)
		}

		if retrieved.Name != key.Name {
			t.Errorf("name mismatch")
		}
	})

	t.Run("GetByUserID", func(t *testing.T) {
		userID := "user_keys"
		for i := 0; i < 3; i++ {
			key := &domain.APIKey{
				ID:        fmt.Sprintf("key_user_%d", i),
				KeyHash:   []byte(fmt.Sprintf("hash%d", i)),
				UserID:    userID,
				Name:      fmt.Sprintf("Key %d", i),
				Tier:      "standard",
				RateLimit: 5000,
				CreatedAt: time.Now(),
				ExpiresAt: sql.NullTime{Time: time.Now().Add(90 * 24 * time.Hour), Valid: true},
				Revoked:   false,
			}
			repo.Create(ctx, key)
		}

		keys, err := repo.GetByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("GetByUserID failed: %v", err)
		}

		if len(keys) < 3 {
			t.Errorf("expected at least 3 keys, got %d", len(keys))
		}
	})

	t.Run("UpdateRequestCount", func(t *testing.T) {
		key := &domain.APIKey{
			ID:        "key_count",
			KeyHash:   []byte("count_hash"),
			UserID:    "user_count",
			Name:      "Count Key",
			Tier:      "basic",
			RateLimit: 1000,
			CreatedAt: time.Now(),
			ExpiresAt: sql.NullTime{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
			Revoked:   false,
		}
		repo.Create(ctx, key)

		for i := 0; i < 5; i++ {
			repo.UpdateRequestCount(ctx, key.ID)
		}

		updated, _ := repo.GetByID(ctx, key.ID)
		if updated.RequestsToday != 5 {
			t.Errorf("expected 5 requests, got %d", updated.RequestsToday)
		}
	})

	t.Run("ResetDailyCount", func(t *testing.T) {
		key := &domain.APIKey{
			ID:        "key_reset",
			KeyHash:   []byte("reset_hash"),
			UserID:    "user_reset",
			Name:      "Reset Key",
			Tier:      "premium",
			RateLimit: 50000,
			CreatedAt: time.Now(),
			ExpiresAt: sql.NullTime{Time: time.Now().Add(365 * 24 * time.Hour), Valid: true},
			Revoked:   false,
		}
		repo.Create(ctx, key)

		repo.UpdateRequestCount(ctx, key.ID)
		repo.ResetDailyCount(ctx, key.ID)

		reset, _ := repo.GetByID(ctx, key.ID)
		if reset.RequestsToday != 0 {
			t.Errorf("expected 0 requests after reset, got %d", reset.RequestsToday)
		}
	})

	t.Run("Revoke", func(t *testing.T) {
		key := &domain.APIKey{
			ID:        "key_revoke",
			KeyHash:   []byte("revoke_hash"),
			UserID:    "user_revoke",
			Name:      "Revoke Key",
			Tier:      "standard",
			RateLimit: 5000,
			CreatedAt: time.Now(),
			ExpiresAt: sql.NullTime{Time: time.Now().Add(180 * 24 * time.Hour), Valid: true},
			Revoked:   false,
		}
		repo.Create(ctx, key)

		err := repo.Revoke(ctx, key.ID)
		if err != nil {
			t.Fatalf("Revoke failed: %v", err)
		}

		revoked, _ := repo.GetByID(ctx, key.ID)
		if !revoked.Revoked {
			t.Error("key not revoked")
		}

		// Should not be retrievable by hash after revocation
		_, err = repo.GetByHash(ctx, key.KeyHash)
		if err != domain.ErrNotFound {
			t.Error("revoked key still retrievable by hash")
		}
	})

	t.Run("ListByTier", func(t *testing.T) {
		tier := "enterprise"
		for i := 0; i < 2; i++ {
			key := &domain.APIKey{
				ID:        fmt.Sprintf("key_tier_%d", i),
				KeyHash:   []byte(fmt.Sprintf("tier_hash%d", i)),
				UserID:    fmt.Sprintf("user_tier_%d", i),
				Name:      fmt.Sprintf("Enterprise Key %d", i),
				Tier:      tier,
				RateLimit: 100000,
				CreatedAt: time.Now(),
				ExpiresAt: sql.NullTime{Time: time.Now().Add(365 * 24 * time.Hour), Valid: true},
				Revoked:   false,
			}
			repo.Create(ctx, key)
		}

		keys, err := repo.ListByTier(ctx, tier)
		if err != nil {
			t.Fatalf("ListByTier failed: %v", err)
		}

		if len(keys) < 2 {
			t.Errorf("expected at least 2 enterprise keys, got %d", len(keys))
		}
	})
}

// setupTestDB creates a test database connection.
func setupTestDB(t *testing.T) *sql.DB {
	db := SetupTestDB(t)
	
	// Clear any existing data
	if err := TruncateTables(t, db); err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}
	
	return db
}

// cleanupTestDB closes the test database connection.
func cleanupTestDB(t *testing.T, db *sql.DB) {
	CleanupTestDB(t, db)
}



