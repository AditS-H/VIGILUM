// Package oracle implements threat intelligence aggregation and on-chain publishing.
package oracle

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════════
// ORACLE SERVICE
// ═══════════════════════════════════════════════════════════════════════════════

// Service orchestrates threat feed ingestion, aggregation, and publishing.
type Service struct {
	mu          sync.RWMutex
	feeds       []FeedFetcher
	aggregator  *Aggregator
	publisher   *Publisher
	storage     SignalStorage
	
	// Worker control
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	
	// Configuration
	pollInterval time.Duration
	publishThreshold uint8
}

// SignalStorage defines persistence interface for threat signals.
type SignalStorage interface {
	SaveFeedEvent(ctx context.Context, event FeedEvent) error
	SaveSignal(ctx context.Context, signal ThreatSignal) error
	SaveSignalUpdate(ctx context.Context, update SignalUpdate) error
	GetRecentEvents(ctx context.Context, since time.Time) ([]FeedEvent, error)
	GetSignalHistory(ctx context.Context, target TargetRef, limit int) ([]ThreatSignal, error)
}

// ServiceConfig contains service configuration.
type ServiceConfig struct {
	PollInterval      time.Duration
	PublishThreshold  uint8 // Minimum score to publish on-chain
	PublisherConfig   *PublisherConfig
	AggregatorThresholds AggregationThresholds
}

// DefaultServiceConfig returns sensible defaults.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		PollInterval:     5 * time.Minute,
		PublishThreshold: 50, // Only publish HIGH/CRITICAL
		AggregatorThresholds: DefaultThresholds(),
	}
}

// NewService creates a new oracle service.
func NewService(cfg ServiceConfig, storage SignalStorage) (*Service, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create aggregator
	aggregator := NewAggregator(cfg.AggregatorThresholds)

	// Create publisher if configured
	var publisher *Publisher
	if cfg.PublisherConfig != nil {
		var err error
		publisher, err = NewPublisher(*cfg.PublisherConfig)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create publisher: %w", err)
		}
	}

	return &Service{
		feeds:            make([]FeedFetcher, 0),
		aggregator:       aggregator,
		publisher:        publisher,
		storage:          storage,
		ctx:              ctx,
		cancel:           cancel,
		pollInterval:     cfg.PollInterval,
		publishThreshold: cfg.PublishThreshold,
	}, nil
}

// AddFeed registers a new threat feed.
func (s *Service) AddFeed(feed FeedFetcher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feeds = append(s.feeds, feed)
}

// Start begins the oracle service workers.
func (s *Service) Start() {
	// Start feed polling worker
	s.wg.Add(1)
	go s.feedPollingWorker()

	// Start publishing worker if configured
	if s.publisher != nil {
		s.wg.Add(1)
		go s.publishingWorker()
	}

	// Start cleanup worker
	s.wg.Add(1)
	go s.cleanupWorker()

	log.Printf("[Oracle] Service started with %d feeds", len(s.feeds))
}

// Stop gracefully shuts down the oracle service.
func (s *Service) Stop() {
	s.cancel()
	s.wg.Wait()
	
	if s.publisher != nil {
		s.publisher.Close()
	}
	
	log.Println("[Oracle] Service stopped")
}

// feedPollingWorker periodically fetches from all feeds.
func (s *Service) feedPollingWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	// Initial fetch
	s.fetchAllFeeds()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.fetchAllFeeds()
		}
	}
}

// fetchAllFeeds fetches from all registered feeds.
func (s *Service) fetchAllFeeds() {
	s.mu.RLock()
	feeds := s.feeds
	s.mu.RUnlock()

	var allEvents []FeedEvent

	for _, feed := range feeds {
		events, err := feed.Fetch(s.ctx)
		if err != nil {
			log.Printf("[Oracle] Error fetching from %s: %v", feed.Source(), err)
			continue
		}

		log.Printf("[Oracle] Fetched %d events from %s", len(events), feed.Source())

		// Save events to storage
		if s.storage != nil {
			for _, event := range events {
				if err := s.storage.SaveFeedEvent(s.ctx, event); err != nil {
					log.Printf("[Oracle] Error saving event: %v", err)
				}
			}
		}

		allEvents = append(allEvents, events...)
		feed.SetLastFetchTime(time.Now())
	}

	// Process events through aggregator
	if len(allEvents) > 0 {
		updatedSignals := s.aggregator.ProcessEvents(allEvents)
		
		// Save updated signals
		if s.storage != nil {
			for _, signal := range updatedSignals {
				if err := s.storage.SaveSignal(s.ctx, signal); err != nil {
					log.Printf("[Oracle] Error saving signal: %v", err)
				}
			}
		}

		// Queue high-risk signals for publishing
		if s.publisher != nil {
			for _, signal := range updatedSignals {
				if signal.RiskScore >= s.publishThreshold {
					if err := s.publisher.QueueUpdate(signal); err != nil {
						// Rate limited or other error, not fatal
						continue
					}
				}
			}
		}

		log.Printf("[Oracle] Processed %d events, updated %d signals", len(allEvents), len(updatedSignals))
	}
}

// publishingWorker periodically publishes queued signals on-chain.
func (s *Service) publishingWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			published, err := s.publisher.FlushQueue(s.ctx)
			if err != nil {
				log.Printf("[Oracle] Publishing error: %v", err)
			} else if published > 0 {
				log.Printf("[Oracle] Published %d signals on-chain", published)
			}
		}
	}
}

// cleanupWorker periodically prunes old events.
func (s *Service) cleanupWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			pruned := s.aggregator.PruneOldEvents(s.ctx)
			if pruned > 0 {
				log.Printf("[Oracle] Pruned %d old events", pruned)
			}
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// PUBLIC API
// ═══════════════════════════════════════════════════════════════════════════════

// GetSignal retrieves the current threat signal for an address.
func (s *Service) GetSignal(chainID domain.ChainID, addr domain.Address) *ThreatSignal {
	return s.aggregator.GetSignal(chainID, addr)
}

// GetSignals retrieves all signals above a minimum score.
func (s *Service) GetSignals(minScore uint8) []ThreatSignal {
	return s.aggregator.GetAllSignals(minScore)
}

// GetHighRiskTargets returns all targets with HIGH or CRITICAL risk.
func (s *Service) GetHighRiskTargets() []ThreatSignal {
	return s.aggregator.GetAllSignals(60) // 60+ is HIGH
}

// GetCriticalTargets returns all targets with CRITICAL risk.
func (s *Service) GetCriticalTargets() []ThreatSignal {
	return s.aggregator.GetAllSignals(80) // 80+ is CRITICAL
}

// IngestScannerResults adds internal scanner results to the feed.
func (s *Service) IngestScannerResults(vulns []domain.Vulnerability) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find or create internal feed
	var internalFeed *InternalFeed
	for _, feed := range s.feeds {
		if f, ok := feed.(*InternalFeed); ok {
			internalFeed = f
			break
		}
	}

	if internalFeed == nil {
		internalFeed = NewInternalFeed()
		s.feeds = append(s.feeds, internalFeed)
	}

	internalFeed.AddVulnerabilities(vulns)
}

// ForcePublish immediately publishes a signal on-chain.
func (s *Service) ForcePublish(ctx context.Context, signal ThreatSignal) (string, error) {
	if s.publisher == nil {
		return "", fmt.Errorf("publisher not configured")
	}

	tx, err := s.publisher.Publish(ctx, signal)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

// Stats returns service statistics.
func (s *Service) Stats() map[string]interface{} {
	stats := map[string]interface{}{
		"aggregator": s.aggregator.Stats(),
		"feeds":      len(s.feeds),
	}

	if s.publisher != nil {
		stats["publisher"] = s.publisher.Stats()
	}

	return stats
}
