// Package firewall provides tests for Identity Firewall service.
package firewall

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestGenerateChallenge(t *testing.T) {
	// Mock DB (simplified for testing)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	challenge, err := service.GenerateChallenge(ctx)
	if err != nil {
		t.Fatalf("GenerateChallenge failed: %v", err)
	}

	if challenge.ID == "" {
		t.Error("Challenge ID should not be empty")
	}
	if challenge.Challenge == "" {
		t.Error("Challenge should not be empty")
	}
	if challenge.ExpiresAt.Before(time.Now()) {
		t.Error("Challenge should not be expired")
	}

	t.Logf("Generated challenge: %s", challenge.ID)
}

func TestGetRiskInfo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	address := "0x742d35Cc6634C0532925a3b844Bc9e7595f42bE"
	info, err := service.GetRiskInfo(ctx, address)
	if err != nil {
		t.Fatalf("GetRiskInfo failed: %v", err)
	}

	if info.Address != address {
		t.Errorf("Expected address %s, got %s", address, info.Address)
	}

	if info.RiskScore < 0.0 || info.RiskScore > 1.0 {
		t.Errorf("Risk score should be 0.0-1.0, got %f", info.RiskScore)
	}

	t.Logf("Risk info: score=%.2f, is_human=%v", info.RiskScore, info.IsHuman)
}

func TestFirewallServiceStats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stats, err := service.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats == nil {
		t.Error("Stats should not be nil")
	}

	if stats.TotalProofs < 0 {
		t.Errorf("TotalProofs should be >= 0, got %d", stats.TotalProofs)
	}

	t.Logf("Stats: total=%d, unique_users=%d", stats.TotalProofs, stats.UniqueUsers)
}
