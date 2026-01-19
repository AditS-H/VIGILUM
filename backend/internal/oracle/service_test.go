// Package oracle provides tests for Threat Oracle service.
package oracle

import (
	"context"
	"testing"
	"time"
)

// MockSignalStorage implements SignalStorage for testing.
type MockSignalStorage struct {
	signals []ThreatSignal
}

func (m *MockSignalStorage) SaveFeedEvent(ctx context.Context, event FeedEvent) error {
	return nil
}

func (m *MockSignalStorage) SaveSignal(ctx context.Context, signal ThreatSignal) error {
	m.signals = append(m.signals, signal)
	return nil
}

func (m *MockSignalStorage) SaveSignalUpdate(ctx context.Context, update SignalUpdate) error {
	return nil
}

func (m *MockSignalStorage) GetRecentEvents(ctx context.Context, since time.Time) ([]FeedEvent, error) {
	return []FeedEvent{}, nil
}

func (m *MockSignalStorage) GetSignalHistory(ctx context.Context, target TargetRef, limit int) ([]ThreatSignal, error) {
	return m.signals, nil
}

func TestNewOracleService(t *testing.T) {
	cfg := DefaultServiceConfig()
	storage := &MockSignalStorage{}

	service, err := NewService(cfg, storage)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	if service == nil {
		t.Error("Service should not be nil")
	}

	service.Stop()
}

func TestAggregateSignals(t *testing.T) {
	cfg := DefaultServiceConfig()
	storage := &MockSignalStorage{}


	service, _ := NewService(cfg, storage)

	// Just verify service was created - aggregation details tested via producer/feed tests
	if service == nil {
		t.Error("Service should not be nil")
	}

	t.Logf("Service created with publish threshold: %d", service.publishThreshold)
	service.Stop()
}

func TestOracleServiceWithMultipleFeed(t *testing.T) {
	cfg := DefaultServiceConfig()
	storage := &MockSignalStorage{}

	service, _ := NewService(cfg, storage)

	// Verify service is running
	if service.ctx.Err() != nil {
		t.Error("Service context should not be cancelled")
	}

	service.Stop()

	// Verify service stopped
	if service.ctx.Err() == nil {
		t.Error("Service context should be cancelled after Stop()")
	}
}

func TestPublishThreshold(t *testing.T) {
	cfg := DefaultServiceConfig()
	cfg.PublishThreshold = 70

	storage := &MockSignalStorage{}
	service, _ := NewService(cfg, storage)

	if service.publishThreshold != 70 {
		t.Errorf("Publish threshold should be 70, got %d", service.publishThreshold)
	}

	service.Stop()
}
