// Package genome provides tests for Genome Analyzer service.
package genome

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestAnalyzeContract(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := &AnalysisRequest{
		ContractAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f42bE",
		Bytecode:        []byte{0x60, 0x80, 0x60, 0x40}, // Simple bytecode
		Priority:        "normal",
	}

	result, err := service.AnalyzeContract(ctx, req)
	if err != nil {
		t.Fatalf("AnalyzeContract failed: %v", err)
	}

	if result.AnalysisID == "" {
		t.Error("AnalysisID should not be empty")
	}

	if result.GenomeHash == "" {
		t.Error("GenomeHash should not be empty")
	}

	if result.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result.Status)
	}

	if result.RiskScore < 0.0 || result.RiskScore > 1.0 {
		t.Errorf("RiskScore should be 0.0-1.0, got %f", result.RiskScore)
	}

	t.Logf("Analysis completed: id=%s, risk=%.2f", result.AnalysisID[:8], result.RiskScore)
}

func TestGetAnalysisStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	analysisID := "abc123def456"

	result, err := service.GetAnalysisStatus(ctx, analysisID)
	if err != nil {
		t.Fatalf("GetAnalysisStatus failed: %v", err)
	}

	if result.AnalysisID != analysisID {
		t.Errorf("Expected ID %s, got %s", analysisID, result.AnalysisID)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result.Status)
	}

	t.Logf("Status retrieved: %s", result.Status)
}

func TestGetGenomeHash(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	genomeHash := "abc123def456abc123def456abc123def456"

	result, err := service.GetGenomeHash(ctx, genomeHash)
	if err != nil {
		t.Fatalf("GetGenomeHash failed: %v", err)
	}

	if result.GenomeHash != genomeHash {
		t.Errorf("Expected hash %s, got %s", genomeHash, result.GenomeHash)
	}

	if result.IPFSHash == "" {
		t.Error("IPFSHash should not be empty")
	}

	// Check bounds before slicing
	if len(result.IPFSHash) >= 16 {
		t.Logf("Genome retrieved: ipfs=%s", result.IPFSHash[:16])
	} else {
		t.Logf("Genome retrieved: ipfs=%s", result.IPFSHash)
	}
}

func TestFindSimilar(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	genomeHash := "abc123def456"
	threshold := 0.85

	results, err := service.FindSimilar(ctx, genomeHash, threshold)
	if err != nil {
		t.Fatalf("FindSimilar failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Should return at least one similar genome")
	}

	for _, result := range results {
		if result.SimilarityScore < threshold {
			t.Errorf("Similarity score %.2f below threshold %.2f", result.SimilarityScore, threshold)
		}
	}

	t.Logf("Found %d similar genomes", len(results))
}

func TestAnalysisMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	bytecode := []byte{0x60, 0x80, 0x60, 0x40, 0x52, 0x34, 0x15}

	req := &AnalysisRequest{
		ContractAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f42bE",
		Bytecode:        bytecode,
	}

	result, err := service.AnalyzeContract(ctx, req)
	if err != nil {
		t.Fatalf("AnalyzeContract failed: %v", err)
	}

	metrics := result.Metrics
	if metrics.BytecodeSize != len(bytecode) {
		t.Errorf("BytecodeSize mismatch: expected %d, got %d", len(bytecode), metrics.BytecodeSize)
	}

	if metrics.OpcodeCount <= 0 {
		t.Error("OpcodeCount should be > 0")
	}

	if metrics.Complexity < 0.0 || metrics.Complexity > 1.0 {
		t.Errorf("Complexity should be 0.0-1.0, got %f", metrics.Complexity)
	}

	t.Logf("Metrics: size=%d, opcodes=%d, complexity=%.2f",
		metrics.BytecodeSize, metrics.OpcodeCount, metrics.Complexity)
}
