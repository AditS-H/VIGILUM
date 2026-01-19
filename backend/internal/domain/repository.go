// Package domain contains core business entities and repository interfaces.
package domain

import (
	"context"
)

// ContractRepository defines the interface for contract persistence.
type ContractRepository interface {
	// Create stores a new contract.
	Create(ctx context.Context, contract *Contract) error
	
	// GetByID retrieves a contract by its ID.
	GetByID(ctx context.Context, id ContractID) (*Contract, error)
	
	// GetByAddress retrieves a contract by chain and address.
	GetByAddress(ctx context.Context, chainID ChainID, address Address) (*Contract, error)
	
	// Update modifies an existing contract.
	Update(ctx context.Context, contract *Contract) error
	
	// Delete removes a contract (soft delete).
	Delete(ctx context.Context, id ContractID) error
	
	// List retrieves contracts with pagination.
	List(ctx context.Context, filter ContractFilter, page, pageSize int) ([]*Contract, int, error)
	
	// FindByBytecodeHash finds contracts with matching bytecode.
	FindByBytecodeHash(ctx context.Context, hash Hash) ([]*Contract, error)
	
	// FindHighRisk retrieves contracts above a risk threshold.
	FindHighRisk(ctx context.Context, minScore float64, limit int) ([]*Contract, error)
}

// ContractFilter specifies filtering options for contract queries.
type ContractFilter struct {
	ChainID      *ChainID
	ThreatLevel  *ThreatLevel
	MinRiskScore *float64
	MaxRiskScore *float64
	IsVerified   *bool
	Labels       []string
	DeployerAddr *Address
}

// VulnerabilityRepository defines the interface for vulnerability persistence.
type VulnerabilityRepository interface {
	// Create stores a new vulnerability.
	Create(ctx context.Context, vuln *Vulnerability) error
	
	// CreateBatch stores multiple vulnerabilities.
	CreateBatch(ctx context.Context, vulns []*Vulnerability) error
	
	// GetByID retrieves a vulnerability by ID.
	GetByID(ctx context.Context, id string) (*Vulnerability, error)
	
	// GetByContract retrieves all vulnerabilities for a contract.
	GetByContract(ctx context.Context, contractID ContractID) ([]*Vulnerability, error)
	
	// Update modifies a vulnerability (e.g., mark as false positive).
	Update(ctx context.Context, vuln *Vulnerability) error
	
	// Delete removes a vulnerability.
	Delete(ctx context.Context, id string) error
	
	// FindByType retrieves vulnerabilities of a specific type.
	FindByType(ctx context.Context, vulnType VulnType, limit int) ([]*Vulnerability, error)
	
	// GetStats returns vulnerability statistics.
	GetStats(ctx context.Context, filter VulnStatsFilter) (*VulnStats, error)
}

// VulnStatsFilter specifies filtering for vulnerability statistics.
type VulnStatsFilter struct {
	ChainID    *ChainID
	ContractID *ContractID
	TimeRange  *TimeRange
}

// TimeRange represents a time interval.
type TimeRange struct {
	Start int64 // Unix timestamp
	End   int64
}

// VulnStats contains aggregated vulnerability statistics.
type VulnStats struct {
	Total          int            `json:"total"`
	BySeverity     map[ThreatLevel]int `json:"by_severity"`
	ByType         map[VulnType]int    `json:"by_type"`
	ConfirmedCount int            `json:"confirmed_count"`
	FalsePosCount  int            `json:"false_positive_count"`
}

// ScanReportRepository defines the interface for scan report persistence.
type ScanReportRepository interface {
	// Create stores a new scan report.
	Create(ctx context.Context, report *ScanReport) error
	
	// GetByID retrieves a scan report by ID.
	GetByID(ctx context.Context, id string) (*ScanReport, error)
	
	// GetByContract retrieves scan reports for a contract.
	GetByContract(ctx context.Context, contractID ContractID, limit int) ([]*ScanReport, error)
	
	// Update modifies a scan report.
	Update(ctx context.Context, report *ScanReport) error
	
	// GetLatest retrieves the most recent scan for a contract.
	GetLatest(ctx context.Context, contractID ContractID, scanType ScanType) (*ScanReport, error)
}

// TransactionRepository defines the interface for transaction persistence.
type TransactionRepository interface {
	// Create stores a transaction.
	Create(ctx context.Context, tx *Transaction) error
	
	// CreateBatch stores multiple transactions.
	CreateBatch(ctx context.Context, txs []*Transaction) error
	
	// GetByHash retrieves a transaction by hash.
	GetByHash(ctx context.Context, chainID ChainID, hash Hash) (*Transaction, error)
	
	// GetByAddress retrieves transactions involving an address.
	GetByAddress(ctx context.Context, chainID ChainID, addr Address, limit int) ([]*Transaction, error)
	
	// FindSuspicious retrieves suspicious transactions.
	FindSuspicious(ctx context.Context, chainID ChainID, limit int) ([]*Transaction, error)
}

// AlertRepository defines the interface for alert persistence.
type AlertRepository interface {
	// Create stores a new alert.
	Create(ctx context.Context, alert *Alert) error
	
	// GetByID retrieves an alert by ID.
	GetByID(ctx context.Context, id string) (*Alert, error)
	
	// List retrieves alerts with filtering.
	List(ctx context.Context, filter AlertFilter, page, pageSize int) ([]*Alert, int, error)
	
	// Update modifies an alert.
	Update(ctx context.Context, alert *Alert) error
	
	// Acknowledge marks an alert as acknowledged.
	Acknowledge(ctx context.Context, id string) error
	
	// Resolve marks an alert as resolved.
	Resolve(ctx context.Context, id string) error
	
	// GetUnacknowledged retrieves pending alerts.
	GetUnacknowledged(ctx context.Context, severity *ThreatLevel, limit int) ([]*Alert, error)
}

// AlertFilter specifies filtering for alert queries.
type AlertFilter struct {
	Type          *AlertType
	Severity      *ThreatLevel
	ChainID       *ChainID
	IsAcknowledged *bool
	IsResolved    *bool
	TimeRange     *TimeRange
}

// ============================================================
// USER REPOSITORY
// ============================================================

// UserRepository handles User persistence.
type UserRepository interface {
	// Create inserts a new user.
	Create(ctx context.Context, wallet string) (*User, error)

	// GetByWallet retrieves a user by wallet address.
	GetByWallet(ctx context.Context, wallet string) (*User, error)

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id string) (*User, error)

	// Update modifies an existing user's risk score and metadata.
	Update(ctx context.Context, id string, u *User) error

	// UpdateRiskScore updates the risk score for a user.
	UpdateRiskScore(ctx context.Context, id string, score float64) error

	// UpdateLastActivity updates the last activity timestamp.
	UpdateLastActivity(ctx context.Context, id string) error

	// Blacklist marks a user as blacklisted.
	Blacklist(ctx context.Context, id string) error

	// RemoveBlacklist removes a user from blacklist.
	RemoveBlacklist(ctx context.Context, id string) error

	// Delete removes a user (admin only).
	Delete(ctx context.Context, id string) error

	// ListByRiskScore returns users above a risk threshold, ordered by risk descending.
	ListByRiskScore(ctx context.Context, threshold float64, limit int) ([]*User, error)

	// ListBlacklisted returns all blacklisted users.
	ListBlacklisted(ctx context.Context, limit int, offset int) ([]*User, error)

	// Count returns the total number of users.
	Count(ctx context.Context) (int64, error)
}

// ============================================================
// HUMAN PROOF REPOSITORY
// ============================================================

// HumanProofRepository handles HumanProof persistence.
type HumanProofRepository interface {
	// Create inserts a new human proof record.
	Create(ctx context.Context, proof *HumanProof) error

	// GetByID retrieves a proof by ID.
	GetByID(ctx context.Context, id string) (*HumanProof, error)

	// GetByHash retrieves a proof by proof hash.
	GetByHash(ctx context.Context, hash []byte) (*HumanProof, error)

	// GetByUserID retrieves all proofs for a user, ordered by created_at DESC.
	GetByUserID(ctx context.Context, userID string, limit int, offset int) ([]*HumanProof, error)

	// Update modifies an existing proof (typically marking as verified).
	Update(ctx context.Context, id string, proof *HumanProof) error

	// MarkVerified marks a proof as verified by a contract.
	MarkVerified(ctx context.Context, id string, verifierAddr string, txHash string) error

	// Delete removes a proof record.
	Delete(ctx context.Context, id string) error

	// DeleteExpired removes all expired proofs.
	DeleteExpired(ctx context.Context) (int64, error)

	// CountByUserID returns the number of proofs for a user.
	CountByUserID(ctx context.Context, userID string) (int64, error)

	// CountVerifiedByUserID returns the number of verified proofs for a user.
	CountVerifiedByUserID(ctx context.Context, userID string) (int64, error)
}

// ============================================================
// THREAT SIGNAL REPOSITORY
// ============================================================

// ThreatSignalRepository handles ThreatSignal persistence.
type ThreatSignalRepository interface {
	// Create inserts a new threat signal.
	Create(ctx context.Context, signal *ThreatSignal) error

	// GetByID retrieves a signal by ID.
	GetByID(ctx context.Context, id string) (*ThreatSignal, error)

	// GetByEntity retrieves signals for an entity on a chain, ordered by risk descending.
	GetByEntity(ctx context.Context, chainID ChainID, address Address, limit int) ([]*ThreatSignal, error)

	// GetUnpublished retrieves signals not yet published on-chain.
	GetUnpublished(ctx context.Context, limit int) ([]*ThreatSignal, error)

	// Update modifies an existing signal.
	Update(ctx context.Context, id string, signal *ThreatSignal) error

	// MarkPublished marks a signal as published on-chain.
	MarkPublished(ctx context.Context, id string) error

	// Delete removes a threat signal.
	Delete(ctx context.Context, id string) error

	// GetHighRisk returns all signals above a risk threshold.
	GetHighRisk(ctx context.Context, threshold int, limit int) ([]*ThreatSignal, error)

	// GetByCriticalSignalType returns signals of critical threat types.
	GetByCriticalSignalType(ctx context.Context, limit int) ([]*ThreatSignal, error)

	// Count returns the total number of signals.
	Count(ctx context.Context) (int64, error)

	// CountByEntity returns the number of signals for an entity.
	CountByEntity(ctx context.Context, chainID ChainID, address Address) (int64, error)
}

// ============================================================
// GENOME REPOSITORY
// ============================================================

// GenomeRepository handles Genome (malware fingerprint) persistence.
type GenomeRepository interface {
	// Create inserts a new genome.
	Create(ctx context.Context, genome *Genome) error

	// GetByID retrieves a genome by ID.
	GetByID(ctx context.Context, id string) (*Genome, error)

	// GetByHash retrieves a genome by genome hash.
	GetByHash(ctx context.Context, hash []byte) (*Genome, error)

	// GetByContractAddress retrieves genomes for a contract.
	GetByContractAddress(ctx context.Context, chainID ChainID, address Address) (*Genome, error)

	// Update modifies an existing genome.
	Update(ctx context.Context, id string, genome *Genome) error

	// Delete removes a genome.
	Delete(ctx context.Context, id string) error

	// ListByLabel retrieves genomes with a specific label.
	ListByLabel(ctx context.Context, label string, limit int, offset int) ([]*Genome, error)

	// ListSimilar retrieves genomes similar to a given one (for clustering analysis).
	ListSimilar(ctx context.Context, genomeID string, threshold float64, limit int) ([]*Genome, error)

	// Count returns the total number of genomes.
	Count(ctx context.Context) (int64, error)

	// CountByLabel returns the number of genomes with a specific label.
	CountByLabel(ctx context.Context, label string) (int64, error)

	// GetDistribution returns count of genomes per label.
	GetDistribution(ctx context.Context) (map[string]int64, error)
}

// ============================================================
// EXPLOIT SUBMISSION REPOSITORY
// ============================================================

// ExploitSubmissionRepository handles ExploitSubmission (bug bounty) persistence.
type ExploitSubmissionRepository interface {
	// Create inserts a new exploit submission.
	Create(ctx context.Context, submission *ExploitSubmission) error

	// GetByID retrieves a submission by ID.
	GetByID(ctx context.Context, id string) (*ExploitSubmission, error)

	// GetByResearcher retrieves all submissions from a researcher.
	GetByResearcher(ctx context.Context, researcher Address, limit int, offset int) ([]*ExploitSubmission, error)

	// GetByTarget retrieves submissions for a target contract.
	GetByTarget(ctx context.Context, chainID ChainID, target Address, limit int) ([]*ExploitSubmission, error)

	// GetByStatus retrieves submissions with a specific status.
	GetByStatus(ctx context.Context, status string, limit int, offset int) ([]*ExploitSubmission, error)

	// GetPending retrieves all pending submissions.
	GetPending(ctx context.Context, limit int) ([]*ExploitSubmission, error)

	// Update modifies an existing submission.
	Update(ctx context.Context, id string, submission *ExploitSubmission) error

	// UpdateStatus changes the submission status.
	UpdateStatus(ctx context.Context, id string, status string) error

	// MarkVerified marks a submission as verified (by auditor).
	MarkVerified(ctx context.Context, id string) error

	// MarkPaid marks a submission as paid (bounty distributed).
	MarkPaid(ctx context.Context, id string, txHash string) error

	// Delete removes a submission.
	Delete(ctx context.Context, id string) error

	// CountByResearcher returns the number of submissions from a researcher.
	CountByResearcher(ctx context.Context, researcher Address) (int64, error)

	// CountByStatus returns the number of submissions in a status.
	CountByStatus(ctx context.Context, status string) (int64, error)

	// GetTotalBountyAmount returns total bounties for submissions in a status.
	GetTotalBountyAmount(ctx context.Context, status string) (int64, error)
}

// ============================================================
// API KEY REPOSITORY
// ============================================================

// APIKeyRepository handles APIKey (rate limiting tier) persistence.
type APIKeyRepository interface {
	// Create inserts a new API key.
	Create(ctx context.Context, key *APIKey) error

	// GetByHash retrieves an API key by hashed key value.
	GetByHash(ctx context.Context, keyHash []byte) (*APIKey, error)

	// GetByID retrieves an API key by ID.
	GetByID(ctx context.Context, id string) (*APIKey, error)

	// GetByUserID retrieves all API keys for a user.
	GetByUserID(ctx context.Context, userID string) ([]*APIKey, error)

	// Update modifies an existing API key.
	Update(ctx context.Context, id string, key *APIKey) error

	// UpdateLastUsed updates the last_used timestamp.
	UpdateLastUsed(ctx context.Context, id string) error

	// UpdateRequestCount increments the daily request counter.
	UpdateRequestCount(ctx context.Context, id string) error

	// ResetDailyCount resets the requests_today counter.
	ResetDailyCount(ctx context.Context, id string) error

	// Revoke marks an API key as revoked.
	Revoke(ctx context.Context, id string) error

	// Delete removes an API key.
	Delete(ctx context.Context, id string) error

	// ListByUserID retrieves all non-revoked keys for a user.
	ListByUserID(ctx context.Context, userID string) ([]*APIKey, error)

	// ListByTier retrieves all active keys in a tier.
	ListByTier(ctx context.Context, tier string) ([]*APIKey, error)

	// GetExpiring retrieves keys expiring soon (next N days).
	GetExpiring(ctx context.Context, days int) ([]*APIKey, error)

	// Count returns the total number of active API keys.
	Count(ctx context.Context) (int64, error)

	// CountByTier returns the number of active keys in a tier.
	CountByTier(ctx context.Context, tier string) (int64, error)
}
