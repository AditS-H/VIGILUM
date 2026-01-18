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
