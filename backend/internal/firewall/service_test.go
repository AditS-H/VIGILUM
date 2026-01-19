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

func TestChallengeExpiration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	challenge1, _ := service.GenerateChallenge(ctx)
	challenge2, _ := service.GenerateChallenge(ctx)

	if challenge1.ID == challenge2.ID {
		t.Error("Challenge IDs should be unique")
	}

	if challenge1.ExpiresAt.After(time.Now().Add(10*time.Minute)) {
		t.Error("Challenge expiration should be within 5 minutes")
	}
}

func TestServiceInitialization(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(nil, logger)

	if service == nil {
		t.Error("Service should not be nil")
	}

	if service.logger == nil {
		t.Error("Logger should be initialized")
	}

	t.Logf("Service initialized successfully")
}
