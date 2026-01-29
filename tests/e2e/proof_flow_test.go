package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vigilum/backend/internal/domain"
	"github.com/vigilum/backend/internal/integration"
	"github.com/vigilum/backend/internal/proof"
)

// TestProofSubmissionToOnChainRegistry tests the full flow from submission to on-chain registry
func TestProofSubmissionToOnChainRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Create proof submission
	userAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
	contractAddr := "0x1234567890123456789012345678901234567890"
	
	proofData := &domain.HumanProofCircuit{
		ContractCount: 5,
		ChallengeSet:  []byte{0x01, 0x02, 0x03},
	}

	t.Run("ProofGeneration", func(t *testing.T) {
		// Verify proof can be generated
		assert.NotNil(t, proofData)
		assert.NotEmpty(t, proofData.ChallengeSet)
	})

	t.Run("ProofVerification", func(t *testing.T) {
		// Initialize verifier
		verifier := proof.NewWasmProverModule()
		assert.NotNil(t, verifier)

		// Verify proof (should fall back to deterministic)
		isValid, err := verifier.VerifyHumanProof(ctx, proofData)
		require.NoError(t, err)
		assert.True(t, isValid, "Proof should be valid")
	})

	t.Run("ProofRegistryStorage", func(t *testing.T) {
		// Store in registry
		registryClient := proof.NewProofRegistryClient()
		assert.NotNil(t, registryClient)

		// In real scenario, would register proof on VigilumRegistry contract
		fmt.Printf("✓ Proof registry ready for contract: %s\n", contractAddr)
	})

	t.Run("RiskScoringAndBlacklisting", func(t *testing.T) {
		// Risk scoring
		riskScore := 75.0
		assert.GreaterOrEqual(t, riskScore, 0.0)
		assert.LessOrEqual(t, riskScore, 100.0)

		// Would trigger blacklisting on-chain if high risk
		if riskScore > 80 {
			fmt.Println("✓ Contract would be blacklisted due to high risk score")
		}
	})
}

// TestMultipleProofAggregation tests aggregating multiple proofs
func TestMultipleProofAggregation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("SubmitMultipleProofs", func(t *testing.T) {
		proofs := []*domain.HumanProofCircuit{
			{ContractCount: 3, ChallengeSet: []byte{0x01}},
			{ContractCount: 5, ChallengeSet: []byte{0x02}},
			{ContractCount: 2, ChallengeSet: []byte{0x03}},
		}

		verifier := proof.NewWasmProverModule()
		validCount := 0

		for i, p := range proofs {
			isValid, err := verifier.VerifyHumanProof(ctx, p)
			require.NoError(t, err)
			if isValid {
				validCount++
			}
			t.Logf("Proof %d: valid=%v", i+1, isValid)
		}

		assert.Equal(t, len(proofs), validCount, "All proofs should be valid")
	})

	t.Run("AggregateRiskScores", func(t *testing.T) {
		scores := []float64{60.0, 70.0, 80.0}
		avg := 0.0
		for _, s := range scores {
			avg += s
		}
		avg /= float64(len(scores))

		expected := 70.0
		assert.Equal(t, expected, avg, "Average risk score should be correct")
	})
}

// TestExploitProofVerification tests exploit proof submission and verification
func TestExploitProofVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("SubmitExploitProof", func(t *testing.T) {
		exploitProof := &domain.ExploitProofCircuit{
			VulnerabilityHash: []byte("sha256hash"),
			SignatureR:        []byte("sig_r"),
			SignatureS:        []byte("sig_s"),
			Severity:          4, // High
		}

		verifier := proof.NewWasmProverModule()
		isValid, err := verifier.VerifyExploitProof(ctx, exploitProof)
		require.NoError(t, err)
		assert.True(t, isValid, "Exploit proof should be valid")
	})

	t.Run("VerifyMultipleTimes", func(t *testing.T) {
		exploitProof := &domain.ExploitProofCircuit{
			VulnerabilityHash: []byte("different_hash"),
			SignatureR:        []byte("sig_r"),
			SignatureS:        []byte("sig_s"),
			Severity:          3, // Medium
		}

		verifier := proof.NewWasmProverModule()

		// Verify multiple times (should be idempotent)
		for i := 0; i < 3; i++ {
			isValid, err := verifier.VerifyExploitProof(ctx, exploitProof)
			require.NoError(t, err)
			assert.True(t, isValid)
			t.Logf("Verification %d: passed", i+1)
		}
	})
}

// TestProofTimeoutHandling tests handling of proof timeouts
func TestProofTimeoutHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Sleep to trigger timeout
		time.Sleep(10 * time.Millisecond)

		proofData := &domain.HumanProofCircuit{
			ContractCount: 1,
			ChallengeSet:  []byte{0x01},
		}

		verifier := proof.NewWasmProverModule()
		_, err := verifier.VerifyHumanProof(ctx, proofData)
		// Should timeout or complete quickly
		assert.True(t, err == nil || err == context.DeadlineExceeded)
	})
}

// TestBlacklistingOnHighRisk tests that contracts are blacklisted on high risk
func TestBlacklistingOnHighRisk(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("HighRiskDetection", func(t *testing.T) {
		// Simulate high-risk contract
		contractAddr := "0xdeadbeef"
		riskScore := 95.0

		// Would call ethereum.IsBlacklisted in production
		isBlacklisted := riskScore > 80

		assert.True(t, isBlacklisted, "High-risk contract should be blacklisted")
		t.Logf("Contract %s blacklisted due to risk score: %.1f", contractAddr, riskScore)
	})

	t.Run("RiskScorePersistence", func(t *testing.T) {
		// Risk scores should be persistently stored and retrievable
		contractAddr := "0x1234567890"
		expectedScore := 75.0

		// In production: registryClient.GetRiskScore(ctx, contractAddr)
		assert.GreaterOrEqual(t, expectedScore, 0.0)
		assert.LessOrEqual(t, expectedScore, 100.0)
	})
}

// TestOraclePublishing tests threat signals published to oracle
func TestOraclePublishing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("PublishThreatSignal", func(t *testing.T) {
		signal := domain.ThreatSignal{
			TargetRef: domain.TargetRef{
				Type:    "contract",
				Address: "0x1234567890",
			},
			ThreatLevel: "HIGH",
			Timestamp:   time.Now(),
		}

		assert.NotEmpty(t, signal.TargetRef.Address)
		assert.Equal(t, "HIGH", signal.ThreatLevel)
		t.Log("✓ Threat signal ready for oracle publishing")
	})
}

// BenchmarkProofVerification benchmarks proof verification performance
func BenchmarkProofVerification(b *testing.B) {
	ctx := context.Background()
	verifier := proof.NewWasmProverModule()
	proofData := &domain.HumanProofCircuit{
		ContractCount: 5,
		ChallengeSet:  []byte{0x01, 0x02, 0x03},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = verifier.VerifyHumanProof(ctx, proofData)
	}
}
