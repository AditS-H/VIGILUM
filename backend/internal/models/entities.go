// Package models defines core domain entities following OBJECT_DESIGN.md
package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// USER
// ============================================================

// User represents a wallet/user in the system
type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	WalletAddress string     `json:"wallet_address" db:"wallet_address"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	LastActivity  *time.Time `json:"last_activity,omitempty" db:"last_activity"`
	RiskScore     float64    `json:"risk_score" db:"risk_score"`
}

// NewUser creates a new user
func NewUser(walletAddress string) *User {
	return &User{
		ID:            uuid.New(),
		WalletAddress: walletAddress,
		CreatedAt:     time.Now().UTC(),
		RiskScore:     0.0,
	}
}

// ============================================================
// HUMAN PROOF
// ============================================================

// HumanProof represents a verified human-behavior proof
type HumanProof struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	ProofHash       []byte          `json:"proof_hash" db:"proof_hash"`
	ProofData       *ProofData      `json:"proof_data,omitempty" db:"proof_data"`
	Verified        bool            `json:"verified" db:"verified"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	VerifiedAt      *time.Time      `json:"verified_at,omitempty" db:"verified_at"`
	VerifierAddress string          `json:"verifier_address,omitempty" db:"verifier_address"`
	TxHash          string          `json:"tx_hash,omitempty" db:"tx_hash"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty" db:"expires_at"`
}

// ProofData contains the proof details (stored as JSONB in Postgres)
type ProofData struct {
	Proof        []byte            `json:"proof"`
	PublicInputs map[string]string `json:"public_inputs"`
	CircuitID    string            `json:"circuit_id"`
	ProverID     string            `json:"prover_id"`
}

// NewHumanProof creates a new human proof
func NewHumanProof(userID uuid.UUID, proofHash []byte, proofData *ProofData) *HumanProof {
	return &HumanProof{
		ID:        uuid.New(),
		UserID:    userID,
		ProofHash: proofHash,
		ProofData: proofData,
		Verified:  false,
		CreatedAt: time.Now().UTC(),
	}
}

// ============================================================
// THREAT SIGNAL
// ============================================================

// ThreatSignal represents a threat intelligence signal
type ThreatSignal struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	EntityAddress string                 `json:"entity_address" db:"entity_address"`
	SignalType    SignalType             `json:"signal_type" db:"signal_type"`
	RiskScore     int                    `json:"risk_score" db:"risk_score"` // 0-100
	Confidence    float64                `json:"confidence" db:"confidence"` // 0.0-1.0
	Source        string                 `json:"source" db:"source"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	PublishedAt   *time.Time             `json:"published_at,omitempty" db:"published_at"`
	TxHash        string                 `json:"tx_hash,omitempty" db:"tx_hash"`
}

// SignalType represents the type of threat signal
type SignalType string

const (
	SignalTypeExploitDetected   SignalType = "exploit_detected"
	SignalTypeKeyLeaked         SignalType = "key_leaked"
	SignalTypeAnomalyDetected   SignalType = "anomaly_detected"
	SignalTypePhishingAttempt   SignalType = "phishing_attempt"
	SignalTypeMaliciousContract SignalType = "malicious_contract"
	SignalTypeFlashLoanAttack   SignalType = "flash_loan_attack"
	SignalTypeReentrancy        SignalType = "reentrancy"
	SignalTypePriceManipulation SignalType = "price_manipulation"
)

// NewThreatSignal creates a new threat signal
func NewThreatSignal(entityAddress string, signalType SignalType, riskScore int, source string) *ThreatSignal {
	return &ThreatSignal{
		ID:            uuid.New(),
		EntityAddress: entityAddress,
		SignalType:    signalType,
		RiskScore:     riskScore,
		Source:        source,
		CreatedAt:     time.Now().UTC(),
		Metadata:      make(map[string]interface{}),
	}
}

// IsHighRisk returns true if risk score >= 70
func (s *ThreatSignal) IsHighRisk() bool {
	return s.RiskScore >= 70
}

// IsCritical returns true if risk score >= 90
func (s *ThreatSignal) IsCritical() bool {
	return s.RiskScore >= 90
}

// ============================================================
// GENOME
// ============================================================

// Genome represents a malware genome fingerprint
type Genome struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	GenomeHash      []byte           `json:"genome_hash" db:"genome_hash"`
	IPFSHash        string           `json:"ipfs_hash" db:"ipfs_hash"`
	ContractAddress string           `json:"contract_address,omitempty" db:"contract_address"`
	Label           GenomeLabel      `json:"label" db:"label"`
	Features        *GenomeFeatures  `json:"features,omitempty" db:"features"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	RegisteredOnChain bool           `json:"registered_on_chain" db:"registered_on_chain"`
	TxHash          string           `json:"tx_hash,omitempty" db:"tx_hash"`
}

// GenomeLabel represents the classification of a genome
type GenomeLabel string

const (
	GenomeLabelKnownExploit GenomeLabel = "known_exploit"
	GenomeLabelSuspicious   GenomeLabel = "suspicious"
	GenomeLabelBenign       GenomeLabel = "benign"
	GenomeLabelUnknown      GenomeLabel = "unknown"
)

// GenomeFeatures contains extracted features (stored as JSONB in Postgres)
type GenomeFeatures struct {
	OpcodeHistogram map[string]int `json:"opcode_histogram"`
	CallGraph       map[string][]string `json:"call_graph"`
	GasPatterns     []int          `json:"gas_patterns"`
	Complexity      float64        `json:"complexity"`
}

// NewGenome creates a new genome
func NewGenome(genomeHash []byte, ipfsHash string, label GenomeLabel) *Genome {
	return &Genome{
		ID:       uuid.New(),
		GenomeHash: genomeHash,
		IPFSHash: ipfsHash,
		Label:    label,
		CreatedAt: time.Now().UTC(),
		RegisteredOnChain: false,
		Features: &GenomeFeatures{
			OpcodeHistogram: make(map[string]int),
			CallGraph:       make(map[string][]string),
			GasPatterns:     []int{},
		},
	}
}

// ============================================================
// EXPLOIT SUBMISSION
// ============================================================

// ExploitSubmission represents a submitted exploit/proof-of-exploit
type ExploitSubmission struct {
	ID              uuid.UUID          `json:"id" db:"id"`
	ResearcherAddress string           `json:"researcher_address" db:"researcher_address"`
	TargetContract  string             `json:"target_contract" db:"target_contract"`
	ProofHash       []byte             `json:"proof_hash" db:"proof_hash"`
	Severity        Severity           `json:"severity" db:"severity"`
	Description     string             `json:"description" db:"description"`
	Status          SubmissionStatus   `json:"status" db:"status"`
	EstimatedBounty uint64             `json:"estimated_bounty" db:"estimated_bounty"` // wei
	CreatedAt       time.Time          `json:"created_at" db:"created_at"`
	VerifiedAt      *time.Time         `json:"verified_at,omitempty" db:"verified_at"`
	TxHash          string             `json:"tx_hash,omitempty" db:"tx_hash"`
}

// Severity represents exploit severity
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// SubmissionStatus represents the status of an exploit submission
type SubmissionStatus string

const (
	SubmissionStatusPending   SubmissionStatus = "pending"
	SubmissionStatusVerified  SubmissionStatus = "verified"
	SubmissionStatusRejected  SubmissionStatus = "rejected"
	SubmissionStatusPaid      SubmissionStatus = "paid"
)

// NewExploitSubmission creates a new exploit submission
func NewExploitSubmission(
	researcherAddr string,
	targetContract string,
	proofHash []byte,
	severity Severity,
	description string,
) *ExploitSubmission {
	return &ExploitSubmission{
		ID:                uuid.New(),
		ResearcherAddress: researcherAddr,
		TargetContract:    targetContract,
		ProofHash:         proofHash,
		Severity:          severity,
		Description:       description,
		Status:            SubmissionStatusPending,
		CreatedAt:         time.Now().UTC(),
	}
}

// ============================================================
// API KEY
// ============================================================

// APIKey represents an API key for rate limiting
type APIKey struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	KeyHash   []byte     `json:"key_hash" db:"key_hash"`
	Tier      APIKeyTier `json:"tier" db:"tier"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	LastUsed  *time.Time `json:"last_used,omitempty" db:"last_used"`
}

// APIKeyTier represents the tier of an API key
type APIKeyTier string

const (
	APIKeyTierFree       APIKeyTier = "free"
	APIKeyTierPaid       APIKeyTier = "paid"
	APIKeyTierEnterprise APIKeyTier = "enterprise"
)

// IsRevoked returns true if the API key has been revoked
func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}
