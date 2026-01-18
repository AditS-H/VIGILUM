// Package oracle implements threat intelligence aggregation and on-chain publishing.
package oracle

import (
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════════
// FEED TYPES
// ═══════════════════════════════════════════════════════════════════════════════

// FeedSource identifies the origin of threat intelligence.
type FeedSource string

const (
	FeedSourceGitHub     FeedSource = "github"      // GitHub exploit repos
	FeedSourceNVD        FeedSource = "nvd"         // NIST NVD
	FeedSourceChainAbuse FeedSource = "chainabuse"  // Chainabuse.com
	FeedSourceBlockSec   FeedSource = "blocksec"    // BlockSec alerts
	FeedSourceCertiK     FeedSource = "certik"      // CertiK Skynet
	FeedSourceSlowMist   FeedSource = "slowmist"    // SlowMist hacked
	FeedSourceForta      FeedSource = "forta"       // Forta alerts
	FeedSourceInternal   FeedSource = "internal"    // Our own scanner
)

// FeedEvent represents a single piece of threat intelligence from a feed.
type FeedEvent struct {
	ID           string            `json:"id"`
	Source       FeedSource        `json:"source"`
	Type         ThreatEventType   `json:"type"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Severity     domain.ThreatLevel `json:"severity"`
	Targets      []TargetRef       `json:"targets"`
	Evidence     Evidence          `json:"evidence"`
	Tags         []string          `json:"tags"`
	Confidence   float64           `json:"confidence"`
	FetchedAt    time.Time         `json:"fetched_at"`
	PublishedAt  time.Time         `json:"published_at"`
	ExternalURL  string            `json:"external_url,omitempty"`
	RawData      []byte            `json:"-"`
}

// ThreatEventType categorizes the type of threat event.
type ThreatEventType string

const (
	ThreatEventExploit     ThreatEventType = "exploit"      // PoC or active exploit
	ThreatEventVuln        ThreatEventType = "vulnerability" // New vulnerability disclosed
	ThreatEventRugPull     ThreatEventType = "rug_pull"     // Rug pull detected
	ThreatEventPhishing    ThreatEventType = "phishing"     // Phishing campaign
	ThreatEventFlashLoan   ThreatEventType = "flash_loan"   // Flash loan attack
	ThreatEventBridgeHack  ThreatEventType = "bridge_hack"  // Bridge exploit
	ThreatEventOracleManip ThreatEventType = "oracle_manip" // Oracle manipulation
	ThreatEventMalware     ThreatEventType = "malware"      // Malware genome match
	ThreatEventSuspicious  ThreatEventType = "suspicious"   // Suspicious activity
)

// TargetRef identifies a target entity (contract, address, project).
type TargetRef struct {
	Type    TargetType      `json:"type"`
	ChainID domain.ChainID  `json:"chain_id,omitempty"`
	Address domain.Address  `json:"address,omitempty"`
	Name    string          `json:"name,omitempty"`
	Hash    domain.Hash     `json:"hash,omitempty"`
}

// TargetType categorizes what kind of target is affected.
type TargetType string

const (
	TargetTypeContract   TargetType = "contract"
	TargetTypeEOA        TargetType = "eoa"
	TargetTypeProject    TargetType = "project"
	TargetTypeBytecode   TargetType = "bytecode"
	TargetTypeSignature  TargetType = "signature"
)

// Evidence contains supporting data for a threat event.
type Evidence struct {
	CVE         string   `json:"cve,omitempty"`
	CVSS        float64  `json:"cvss,omitempty"`
	TxHashes    []string `json:"tx_hashes,omitempty"`
	ExploitCode string   `json:"exploit_code,omitempty"`
	IPAddresses []string `json:"ip_addresses,omitempty"`
	Signatures  []string `json:"signatures,omitempty"`
	Domains     []string `json:"domains,omitempty"`
	Screenshots []string `json:"screenshots,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// AGGREGATED SIGNALS
// ═══════════════════════════════════════════════════════════════════════════════

// ThreatSignal is the aggregated threat score for a target.
type ThreatSignal struct {
	Target          TargetRef         `json:"target"`
	RiskScore       uint8             `json:"risk_score"`       // 0-100
	ThreatLevel     domain.ThreatLevel `json:"threat_level"`
	Confidence      float64           `json:"confidence"`
	Sources         []FeedSource      `json:"sources"`
	EventCount      int               `json:"event_count"`
	LatestEventAt   time.Time         `json:"latest_event_at"`
	FirstSeenAt     time.Time         `json:"first_seen_at"`
	EventIDs        []string          `json:"event_ids"`
	SummaryReason   string            `json:"summary_reason"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// SignalUpdate represents an update to be published on-chain.
type SignalUpdate struct {
	TargetAddress domain.Address    `json:"target_address"`
	ChainID       domain.ChainID    `json:"chain_id"`
	RiskScore     uint8             `json:"risk_score"`
	Reason        string            `json:"reason"`
	Nonce         uint64            `json:"nonce"`
	Timestamp     time.Time         `json:"timestamp"`
	TxHash        domain.Hash       `json:"tx_hash,omitempty"`
	Status        PublishStatus     `json:"status"`
}

// PublishStatus tracks the state of an on-chain signal update.
type PublishStatus string

const (
	PublishStatusPending   PublishStatus = "pending"
	PublishStatusSubmitted PublishStatus = "submitted"
	PublishStatusConfirmed PublishStatus = "confirmed"
	PublishStatusFailed    PublishStatus = "failed"
)

// ═══════════════════════════════════════════════════════════════════════════════
// FEED CONFIGS
// ═══════════════════════════════════════════════════════════════════════════════

// FeedConfig contains configuration for a threat feed.
type FeedConfig struct {
	Source        FeedSource    `json:"source"`
	Enabled       bool          `json:"enabled"`
	PollInterval  time.Duration `json:"poll_interval"`
	APIKey        string        `json:"-"`
	BaseURL       string        `json:"base_url"`
	RateLimit     int           `json:"rate_limit_per_minute"`
	MaxResults    int           `json:"max_results"`
}
