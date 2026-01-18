// Package oracle implements threat intelligence aggregation and on-chain publishing.
package oracle

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════════
// SIGNAL AGGREGATOR
// ═══════════════════════════════════════════════════════════════════════════════

// Aggregator combines multiple threat feeds into unified risk signals.
type Aggregator struct {
	mu          sync.RWMutex
	signals     map[string]*ThreatSignal // key = chain:address
	events      map[string][]FeedEvent   // key = chain:address -> events
	sourceWeights map[FeedSource]float64
	thresholds    AggregationThresholds
}

// AggregationThresholds configure signal aggregation behavior.
type AggregationThresholds struct {
	MinSourcesForHigh     int     // Min sources to mark HIGH risk
	MinSourcesForCritical int     // Min sources to mark CRITICAL
	MinConfidence         float64 // Min confidence to include event
	EventDecayHours       int     // Hours after which event weight decays
	MaxEventsPerTarget    int     // Max events to track per target
}

// DefaultThresholds returns sensible defaults for aggregation.
func DefaultThresholds() AggregationThresholds {
	return AggregationThresholds{
		MinSourcesForHigh:     2,
		MinSourcesForCritical: 3,
		MinConfidence:         0.3,
		EventDecayHours:       168, // 1 week
		MaxEventsPerTarget:    100,
	}
}

// NewAggregator creates a new signal aggregator.
func NewAggregator(thresholds AggregationThresholds) *Aggregator {
	return &Aggregator{
		signals: make(map[string]*ThreatSignal),
		events:  make(map[string][]FeedEvent),
		sourceWeights: map[FeedSource]float64{
			FeedSourceGitHub:     0.6, // Lower - needs manual review
			FeedSourceNVD:        0.9, // High - official source
			FeedSourceChainAbuse: 0.7, // Medium-high - community validated
			FeedSourceBlockSec:   0.9, // High - security team
			FeedSourceCertiK:     0.9, // High - security audit firm
			FeedSourceSlowMist:   0.9, // High - security team
			FeedSourceForta:      0.8, // High - automated but ML-based
			FeedSourceInternal:   0.7, // Medium - our own scanner
		},
		thresholds: thresholds,
	}
}

// TargetKey generates a unique key for a target.
func TargetKey(chainID domain.ChainID, addr domain.Address) string {
	return string(chainID) + ":" + string(addr)
}

// ProcessEvents ingests new feed events and updates signals.
func (a *Aggregator) ProcessEvents(events []FeedEvent) []ThreatSignal {
	a.mu.Lock()
	defer a.mu.Unlock()

	updatedTargets := make(map[string]bool)

	for _, event := range events {
		// Filter low confidence events
		if event.Confidence < a.thresholds.MinConfidence {
			continue
		}

		// Process each target in the event
		for _, target := range event.Targets {
			if target.Address == "" {
				continue
			}

			key := TargetKey(target.ChainID, target.Address)
			updatedTargets[key] = true

			// Add event to history
			a.events[key] = append(a.events[key], event)

			// Trim to max events
			if len(a.events[key]) > a.thresholds.MaxEventsPerTarget {
				a.events[key] = a.events[key][len(a.events[key])-a.thresholds.MaxEventsPerTarget:]
			}
		}
	}

	// Recalculate signals for updated targets
	var updatedSignals []ThreatSignal
	for key := range updatedTargets {
		signal := a.calculateSignal(key)
		if signal != nil {
			a.signals[key] = signal
			updatedSignals = append(updatedSignals, *signal)
		}
	}

	return updatedSignals
}

// calculateSignal computes the aggregated threat signal for a target.
func (a *Aggregator) calculateSignal(key string) *ThreatSignal {
	events := a.events[key]
	if len(events) == 0 {
		return nil
	}

	// Collect unique sources
	sourceSet := make(map[FeedSource]bool)
	var eventIDs []string
	var latestTime, firstTime time.Time
	var totalWeight float64
	var maxSeverity domain.ThreatLevel

	now := time.Now()
	decayDuration := time.Duration(a.thresholds.EventDecayHours) * time.Hour

	for _, event := range events {
		// Apply time decay
		age := now.Sub(event.FetchedAt)
		decayFactor := 1.0
		if age > decayDuration {
			decayFactor = 0.1 // Heavily reduce weight of old events
		} else {
			decayFactor = 1.0 - (float64(age) / float64(decayDuration) * 0.5)
		}

		// Calculate weighted contribution
		sourceWeight := a.sourceWeights[event.Source]
		weight := event.Confidence * sourceWeight * decayFactor
		totalWeight += weight

		sourceSet[event.Source] = true
		eventIDs = append(eventIDs, event.ID)

		// Track timing
		if latestTime.IsZero() || event.FetchedAt.After(latestTime) {
			latestTime = event.FetchedAt
		}
		if firstTime.IsZero() || event.FetchedAt.Before(firstTime) {
			firstTime = event.FetchedAt
		}

		// Track max severity
		if severityRank(event.Severity) > severityRank(maxSeverity) {
			maxSeverity = event.Severity
		}
	}

	// Count unique sources
	var sources []FeedSource
	for source := range sourceSet {
		sources = append(sources, source)
	}
	numSources := len(sources)

	// Calculate risk score (0-100)
	baseScore := (totalWeight / float64(len(events))) * 100
	
	// Boost score based on number of sources
	sourceBoost := float64(numSources) * 10
	riskScore := baseScore + sourceBoost
	if riskScore > 100 {
		riskScore = 100
	}

	// Determine threat level
	threatLevel := a.calculateThreatLevel(int(riskScore), numSources, maxSeverity)

	// Calculate confidence
	confidence := 0.5 + float64(numSources)*0.15
	if confidence > 0.95 {
		confidence = 0.95
	}

	// Generate summary reason
	reason := a.generateReason(events, numSources)

	// Extract target info from first event
	var target TargetRef
	if len(events[0].Targets) > 0 {
		target = events[0].Targets[0]
	}

	return &ThreatSignal{
		Target:        target,
		RiskScore:     uint8(riskScore),
		ThreatLevel:   threatLevel,
		Confidence:    confidence,
		Sources:       sources,
		EventCount:    len(events),
		LatestEventAt: latestTime,
		FirstSeenAt:   firstTime,
		EventIDs:      eventIDs,
		SummaryReason: reason,
		UpdatedAt:     now,
	}
}

// calculateThreatLevel determines threat level from score and sources.
func (a *Aggregator) calculateThreatLevel(score int, numSources int, maxSeverity domain.ThreatLevel) domain.ThreatLevel {
	// CRITICAL: 3+ sources OR score > 80 with critical events
	if numSources >= a.thresholds.MinSourcesForCritical ||
		(score >= 80 && maxSeverity == domain.ThreatLevelCritical) {
		return domain.ThreatLevelCritical
	}

	// HIGH: 2+ sources OR score > 60
	if numSources >= a.thresholds.MinSourcesForHigh || score >= 60 {
		return domain.ThreatLevelHigh
	}

	// MEDIUM: score 40-60
	if score >= 40 {
		return domain.ThreatLevelMedium
	}

	// LOW: score 20-40
	if score >= 20 {
		return domain.ThreatLevelLow
	}

	return domain.ThreatLevelInfo
}

// severityRank converts ThreatLevel to numeric rank for comparison.
func severityRank(level domain.ThreatLevel) int {
	switch level {
	case domain.ThreatLevelCritical:
		return 5
	case domain.ThreatLevelHigh:
		return 4
	case domain.ThreatLevelMedium:
		return 3
	case domain.ThreatLevelLow:
		return 2
	case domain.ThreatLevelInfo:
		return 1
	default:
		return 0
	}
}

// generateReason creates a human-readable summary of why target is risky.
func (a *Aggregator) generateReason(events []FeedEvent, numSources int) string {
	// Group events by type
	typeCounts := make(map[ThreatEventType]int)
	for _, e := range events {
		typeCounts[e.Type]++
	}

	// Find most common type
	var topType ThreatEventType
	var topCount int
	for t, c := range typeCounts {
		if c > topCount {
			topType = t
			topCount = c
		}
	}

	// Build reason string
	switch topType {
	case ThreatEventExploit:
		return formatWithSources("Exploit PoC reported", numSources)
	case ThreatEventVuln:
		return formatWithSources("Vulnerability disclosed", numSources)
	case ThreatEventRugPull:
		return formatWithSources("Rug pull pattern detected", numSources)
	case ThreatEventPhishing:
		return formatWithSources("Phishing campaign linked", numSources)
	case ThreatEventFlashLoan:
		return formatWithSources("Flash loan attack detected", numSources)
	case ThreatEventBridgeHack:
		return formatWithSources("Bridge exploit reported", numSources)
	case ThreatEventOracleManip:
		return formatWithSources("Oracle manipulation detected", numSources)
	case ThreatEventMalware:
		return formatWithSources("Malware genome match", numSources)
	default:
		return formatWithSources("Suspicious activity reported", numSources)
	}
}

func formatWithSources(base string, numSources int) string {
	if numSources > 1 {
		return base + " by multiple sources"
	}
	return base
}

// GetSignal retrieves the current signal for a target.
func (a *Aggregator) GetSignal(chainID domain.ChainID, addr domain.Address) *ThreatSignal {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	key := TargetKey(chainID, addr)
	return a.signals[key]
}

// GetAllSignals returns all current signals above threshold.
func (a *Aggregator) GetAllSignals(minScore uint8) []ThreatSignal {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var signals []ThreatSignal
	for _, signal := range a.signals {
		if signal.RiskScore >= minScore {
			signals = append(signals, *signal)
		}
	}

	// Sort by risk score descending
	sort.Slice(signals, func(i, j int) bool {
		return signals[i].RiskScore > signals[j].RiskScore
	})

	return signals
}

// PruneOldEvents removes events older than the decay period.
func (a *Aggregator) PruneOldEvents(ctx context.Context) int {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := time.Now().Add(-time.Duration(a.thresholds.EventDecayHours*2) * time.Hour)
	pruned := 0

	for key, events := range a.events {
		var kept []FeedEvent
		for _, e := range events {
			if e.FetchedAt.After(cutoff) {
				kept = append(kept, e)
			} else {
				pruned++
			}
		}
		
		if len(kept) == 0 {
			delete(a.events, key)
			delete(a.signals, key)
		} else {
			a.events[key] = kept
		}
	}

	return pruned
}

// Stats returns aggregator statistics.
func (a *Aggregator) Stats() map[string]int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	totalEvents := 0
	for _, events := range a.events {
		totalEvents += len(events)
	}

	criticalSignals := 0
	highSignals := 0
	for _, signal := range a.signals {
		switch signal.ThreatLevel {
		case domain.ThreatLevelCritical:
			criticalSignals++
		case domain.ThreatLevelHigh:
			highSignals++
		}
	}

	return map[string]int{
		"total_targets":     len(a.signals),
		"total_events":      totalEvents,
		"critical_signals":  criticalSignals,
		"high_signals":      highSignals,
	}
}
