// Package domain contains core business entities.
package domain

import (
	"time"
)

// ContractID is a unique identifier for a smart contract.
type ContractID string

// ChainID represents a blockchain network identifier.
type ChainID int64

// Address represents a blockchain address (20 bytes for EVM).
type Address string

// Hash represents a 32-byte hash.
type Hash string

// Contract represents a deployed smart contract.
type Contract struct {
	ID              ContractID `json:"id"`
	ChainID         ChainID    `json:"chain_id"`
	Address         Address    `json:"address"`
	Name            string     `json:"name,omitempty"`
	Bytecode        []byte     `json:"-"`
	BytecodeHash    Hash       `json:"bytecode_hash"`
	SourceCode      string     `json:"-"`
	ABI             string     `json:"abi,omitempty"`
	CompilerVersion string     `json:"compiler_version,omitempty"`
	IsVerified      bool       `json:"is_verified"`
	DeployedAt      time.Time  `json:"deployed_at"`
	DeployerAddress Address    `json:"deployer_address"`
	DeployTxHash    Hash       `json:"deploy_tx_hash"`
	RiskScore       float64    `json:"risk_score"`
	ThreatLevel     ThreatLevel `json:"threat_level"`
	Labels          []string   `json:"labels,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ThreatLevel categorizes the severity of a threat.
type ThreatLevel string

const (
	ThreatLevelCritical ThreatLevel = "critical"
	ThreatLevelHigh     ThreatLevel = "high"
	ThreatLevelMedium   ThreatLevel = "medium"
	ThreatLevelLow      ThreatLevel = "low"
	ThreatLevelInfo     ThreatLevel = "info"
	ThreatLevelNone     ThreatLevel = "none"
)

// Vulnerability represents a detected security issue.
type Vulnerability struct {
	ID           string       `json:"id"`
	ContractID   ContractID   `json:"contract_id"`
	Type         VulnType     `json:"type"`
	Severity     ThreatLevel  `json:"severity"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Location     CodeLocation `json:"location,omitempty"`
	Remediation  string       `json:"remediation,omitempty"`
	CWE          string       `json:"cwe,omitempty"`
	Confidence   float64      `json:"confidence"`
	DetectedBy   string       `json:"detected_by"`
	DetectedAt   time.Time    `json:"detected_at"`
	IsConfirmed  bool         `json:"is_confirmed"`
	IsFalsePos   bool         `json:"is_false_positive"`
}

// VulnType categorizes vulnerability types.
type VulnType string

const (
	VulnReentrancy         VulnType = "reentrancy"
	VulnOverflow           VulnType = "integer_overflow"
	VulnUnderflow          VulnType = "integer_underflow"
	VulnAccessControl      VulnType = "access_control"
	VulnUncheckedCall      VulnType = "unchecked_external_call"
	VulnTxOrigin           VulnType = "tx_origin"
	VulnTimestamp          VulnType = "timestamp_dependency"
	VulnFrontrunning       VulnType = "frontrunning"
	VulnFlashLoan          VulnType = "flash_loan_attack"
	VulnOracleManipulation VulnType = "oracle_manipulation"
	VulnRugPull            VulnType = "rug_pull_pattern"
	VulnHoneypot           VulnType = "honeypot"
	VulnPhishing           VulnType = "phishing_signature"
	VulnLogicError         VulnType = "logic_error"
	VulnPrecisionLoss      VulnType = "precision_loss"
	VulnWeakRandomness     VulnType = "weak_randomness"
	VulnDOS                VulnType = "denial_of_service"
	VulnStorageCollision   VulnType = "storage_collision"
)

// CodeLocation pinpoints a location in source code.
type CodeLocation struct {
	File      string `json:"file,omitempty"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	StartCol  int    `json:"start_col,omitempty"`
	EndCol    int    `json:"end_col,omitempty"`
	Snippet   string `json:"snippet,omitempty"`
}

// ScanReport represents a complete security scan result.
type ScanReport struct {
	ID              string          `json:"id"`
	ContractID      ContractID      `json:"contract_id"`
	ScanType        ScanType        `json:"scan_type"`
	Status          ScanStatus      `json:"status"`
	RiskScore       float64         `json:"risk_score"`
	ThreatLevel     ThreatLevel     `json:"threat_level"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Metrics         ScanMetrics     `json:"metrics"`
	StartedAt       time.Time       `json:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	Duration        time.Duration   `json:"duration,omitempty"`
	Error           string          `json:"error,omitempty"`
}

// ScanType categorizes scan methodologies.
type ScanType string

const (
	ScanTypeStatic   ScanType = "static"
	ScanTypeDynamic  ScanType = "dynamic"
	ScanTypeML       ScanType = "ml_inference"
	ScanTypeSymbolic ScanType = "symbolic"
	ScanTypeFuzz     ScanType = "fuzz"
	ScanTypeFull     ScanType = "full"
)

// ScanStatus represents the current state of a scan.
type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
	ScanStatusCancelled ScanStatus = "cancelled"
)

// ScanMetrics contains quantitative scan results.
type ScanMetrics struct {
	TotalIssues      int     `json:"total_issues"`
	CriticalCount    int     `json:"critical_count"`
	HighCount        int     `json:"high_count"`
	MediumCount      int     `json:"medium_count"`
	LowCount         int     `json:"low_count"`
	InfoCount        int     `json:"info_count"`
	CodeCoverage     float64 `json:"code_coverage,omitempty"`
	PathsExplored    int     `json:"paths_explored,omitempty"`
	InstructionsExec int     `json:"instructions_executed,omitempty"`
}

// Transaction represents a blockchain transaction.
type Transaction struct {
	Hash        Hash       `json:"hash"`
	ChainID     ChainID    `json:"chain_id"`
	BlockNumber uint64     `json:"block_number"`
	BlockHash   Hash       `json:"block_hash"`
	From        Address    `json:"from"`
	To          *Address   `json:"to,omitempty"` // nil for contract creation
	Value       string     `json:"value"`
	GasPrice    string     `json:"gas_price"`
	GasUsed     uint64     `json:"gas_used"`
	Input       []byte     `json:"-"`
	InputHex    string     `json:"input"`
	Nonce       uint64     `json:"nonce"`
	Timestamp   time.Time  `json:"timestamp"`
	IsContract  bool       `json:"is_contract_creation"`
	IsSuspicious bool      `json:"is_suspicious"`
	RiskScore   float64    `json:"risk_score"`
}

// Alert represents a security alert triggered by the system.
type Alert struct {
	ID          string      `json:"id"`
	Type        AlertType   `json:"type"`
	Severity    ThreatLevel `json:"severity"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	ContractID  *ContractID `json:"contract_id,omitempty"`
	TxHash      *Hash       `json:"tx_hash,omitempty"`
	ChainID     ChainID     `json:"chain_id"`
	Address     *Address    `json:"address,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	AckedAt     *time.Time  `json:"acknowledged_at,omitempty"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty"`
}

// AlertType categorizes alert sources.
type AlertType string

const (
	AlertTypeScan        AlertType = "scan_result"
	AlertTypeRealtime    AlertType = "realtime_detection"
	AlertTypeMempool     AlertType = "mempool_threat"
	AlertTypeAnomaly     AlertType = "anomaly_detection"
	AlertTypeReputation  AlertType = "reputation_change"
)
