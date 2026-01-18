# VIGILUM Object Design

**Purpose:** Complete specification of every class, struct, interface, enum, and method across all languages. Copy-paste ready. No ambiguity.

---

## Table of Contents
1. [Go Backend Objects](#1-go-backend-objects)
2. [Rust Crypto & SDK Objects](#2-rust-crypto--sdk-objects)
3. [Solidity Smart Contract Objects](#3-solidity-smart-contract-objects)
4. [Python ML Objects](#4-python-ml-objects)
5. [Protobuf/gRPC Definitions](#5-protobufgrpc-definitions)
6. [TypeScript SDK Objects](#6-typescript-sdk-objects)

---

## 1. Go Backend Objects

### 1.1 Configuration (`backend/internal/config/`)

```go
// config.go
package config

import (
    "time"
)

// Config holds all application configuration
type Config struct {
    Server      ServerConfig      `json:"server"`
    Database    DatabaseConfig    `json:"database"`
    Redis       RedisConfig       `json:"redis"`
    Ethereum    EthereumConfig    `json:"ethereum"`
    IPFS        IPFSConfig        `json:"ipfs"`
    ZKProver    ZKProverConfig    `json:"zk_prover"`
    Temporal    TemporalConfig    `json:"temporal"`
    Logging     LoggingConfig     `json:"logging"`
    RateLimiter RateLimiterConfig `json:"rate_limiter"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
    Host            string        `json:"host" default:"0.0.0.0"`
    Port            int           `json:"port" default:"8080"`
    ReadTimeout     time.Duration `json:"read_timeout" default:"30s"`
    WriteTimeout    time.Duration `json:"write_timeout" default:"30s"`
    ShutdownTimeout time.Duration `json:"shutdown_timeout" default:"10s"`
    MaxRequestSize  int64         `json:"max_request_size" default:"1048576"` // 1MB
}

// DatabaseConfig holds Postgres connection settings
type DatabaseConfig struct {
    Host            string        `json:"host" required:"true"`
    Port            int           `json:"port" default:"5432"`
    User            string        `json:"user" required:"true"`
    Password        string        `json:"password" required:"true"`
    Database        string        `json:"database" required:"true"`
    SSLMode         string        `json:"ssl_mode" default:"require"`
    MaxOpenConns    int           `json:"max_open_conns" default:"50"`
    MaxIdleConns    int           `json:"max_idle_conns" default:"10"`
    ConnMaxLifetime time.Duration `json:"conn_max_lifetime" default:"1h"`
}

// DSN returns the connection string
func (c DatabaseConfig) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
    )
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
    Host         string        `json:"host" required:"true"`
    Port         int           `json:"port" default:"6379"`
    Password     string        `json:"password"`
    DB           int           `json:"db" default:"0"`
    PoolSize     int           `json:"pool_size" default:"100"`
    MinIdleConns int           `json:"min_idle_conns" default:"10"`
    DialTimeout  time.Duration `json:"dial_timeout" default:"5s"`
    ReadTimeout  time.Duration `json:"read_timeout" default:"3s"`
    WriteTimeout time.Duration `json:"write_timeout" default:"3s"`
}

// Addr returns Redis address
func (c RedisConfig) Addr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// EthereumConfig holds Ethereum RPC settings
type EthereumConfig struct {
    RPCURL              string        `json:"rpc_url" required:"true"`
    FallbackRPCURLs     []string      `json:"fallback_rpc_urls"`
    ChainID             int64         `json:"chain_id" required:"true"`
    PrivateKey          string        `json:"private_key"` // For signing txs
    GasLimit            uint64        `json:"gas_limit" default:"500000"`
    GasPriceMultiplier  float64       `json:"gas_price_multiplier" default:"1.1"`
    ConfirmationBlocks  int           `json:"confirmation_blocks" default:"2"`
    PollingInterval     time.Duration `json:"polling_interval" default:"12s"`
    
    // Contract addresses
    IdentityFirewallAddress string `json:"identity_firewall_address"`
    ThreatOracleAddress     string `json:"threat_oracle_address"`
    MalwareGenomeDBAddress  string `json:"malware_genome_db_address"`
    RedTeamDAOAddress       string `json:"red_team_dao_address"`
    ProofOfExploitAddress   string `json:"proof_of_exploit_address"`
}

// IPFSConfig holds IPFS client settings
type IPFSConfig struct {
    APIURL        string        `json:"api_url" default:"http://localhost:5001"`
    GatewayURL    string        `json:"gateway_url" default:"https://ipfs.io/ipfs/"`
    PinTimeout    time.Duration `json:"pin_timeout" default:"60s"`
    InfuraAPIKey  string        `json:"infura_api_key"`
    InfuraSecret  string        `json:"infura_secret"`
    PinataAPIKey  string        `json:"pinata_api_key"`
    PinataSecret  string        `json:"pinata_secret"`
}

// ZKProverConfig holds ZK prover service settings
type ZKProverConfig struct {
    GRPCAddress       string        `json:"grpc_address" default:"localhost:50051"`
    Timeout           time.Duration `json:"timeout" default:"30s"`
    MaxRetries        int           `json:"max_retries" default:"3"`
    RetryBackoff      time.Duration `json:"retry_backoff" default:"1s"`
    EnableTLS         bool          `json:"enable_tls" default:"false"`
    TLSCertPath       string        `json:"tls_cert_path"`
}

// TemporalConfig holds Temporal workflow settings
type TemporalConfig struct {
    HostPort        string `json:"host_port" default:"localhost:7233"`
    Namespace       string `json:"namespace" default:"vigilum"`
    TaskQueue       string `json:"task_queue" default:"vigilum-tasks"`
    WorkerCount     int    `json:"worker_count" default:"4"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
    Level      string `json:"level" default:"info"` // debug, info, warn, error
    Format     string `json:"format" default:"json"` // json, text
    Output     string `json:"output" default:"stdout"` // stdout, file
    FilePath   string `json:"file_path"`
    MaxSize    int    `json:"max_size" default:"100"` // MB
    MaxBackups int    `json:"max_backups" default:"10"`
    MaxAge     int    `json:"max_age" default:"30"` // days
}

// RateLimiterConfig holds rate limiting settings
type RateLimiterConfig struct {
    Enabled               bool          `json:"enabled" default:"true"`
    DefaultRatePerMinute  int           `json:"default_rate_per_minute" default:"10"`
    FreeAPIKeyRate        int           `json:"free_api_key_rate" default:"100"`
    PaidAPIKeyRate        int           `json:"paid_api_key_rate" default:"1000"`
    BurstMultiplier       int           `json:"burst_multiplier" default:"2"`
    CleanupInterval       time.Duration `json:"cleanup_interval" default:"5m"`
}

// Load loads configuration from environment and files
func Load() (*Config, error)

// Validate validates all configuration values
func (c *Config) Validate() error
```

---

### 1.2 Domain Models (`backend/internal/models/`)

```go
// entities.go
package models

import (
    "time"
    "github.com/google/uuid"
)

// ========== User ==========

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

// Validate validates user fields
func (u *User) Validate() error

// ========== Human Proof ==========

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
}

// ProofData contains the proof details (stored as JSONB)
type ProofData struct {
    Proof        []byte            `json:"proof"`
    PublicInputs map[string]string `json:"public_inputs"`
    CircuitID    string            `json:"circuit_id"`
    ProverID     string            `json:"prover_id"`
}

// NewHumanProof creates a new human proof
func NewHumanProof(userID uuid.UUID, proofHash []byte, proofData *ProofData) *HumanProof

// ========== Threat Signal ==========

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
    SignalTypeExploitDetected    SignalType = "exploit_detected"
    SignalTypeKeyLeaked          SignalType = "key_leaked"
    SignalTypeAnomalyDetected    SignalType = "anomaly_detected"
    SignalTypePhishingAttempt    SignalType = "phishing_attempt"
    SignalTypeMaliciousContract  SignalType = "malicious_contract"
    SignalTypeFlashLoanAttack    SignalType = "flash_loan_attack"
    SignalTypeReentrancy         SignalType = "reentrancy"
    SignalTypePriceManipulation  SignalType = "price_manipulation"
)

// NewThreatSignal creates a new threat signal
func NewThreatSignal(entityAddress string, signalType SignalType, riskScore int, source string) *ThreatSignal

// IsHighRisk returns true if risk score >= 70
func (s *ThreatSignal) IsHighRisk() bool {
    return s.RiskScore >= 70
}

// IsCritical returns true if risk score >= 90
func (s *ThreatSignal) IsCritical() bool {
    return s.RiskScore >= 90
}

// ========== Genome ==========

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

// GenomeFeatures contains extracted features (stored as JSONB)
type GenomeFeatures struct {
    OpcodeHistogram     map[string]int    `json:"opcode_histogram"`
    CallGraph           *CallGraph        `json:"call_graph"`
    GasPatterns         *GasPatterns      `json:"gas_patterns"`
    StateTransitions    []StateTransition `json:"state_transitions"`
    BytecodeLength      int               `json:"bytecode_length"`
    UniqueOpcodes       int               `json:"unique_opcodes"`
    JumpDensity         float64           `json:"jump_density"`
    ExternalCallCount   int               `json:"external_call_count"`
    StorageAccessCount  int               `json:"storage_access_count"`
}

// CallGraph represents the contract call graph
type CallGraph struct {
    Nodes []CallGraphNode `json:"nodes"`
    Edges []CallGraphEdge `json:"edges"`
}

type CallGraphNode struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Type     string `json:"type"` // function, external, internal
    Selector string `json:"selector,omitempty"`
}

type CallGraphEdge struct {
    From string `json:"from"`
    To   string `json:"to"`
    Type string `json:"type"` // call, delegatecall, staticcall
}

// GasPatterns represents gas usage patterns
type GasPatterns struct {
    TotalGas       uint64             `json:"total_gas"`
    AvgGasPerOp    float64            `json:"avg_gas_per_op"`
    GasVariance    float64            `json:"gas_variance"`
    HotspotOpcodes map[string]uint64  `json:"hotspot_opcodes"`
}

// StateTransition represents a state change
type StateTransition struct {
    Slot      string `json:"slot"`
    OldValue  string `json:"old_value"`
    NewValue  string `json:"new_value"`
    Operation string `json:"operation"` // SSTORE, SLOAD
}

// NewGenome creates a new genome
func NewGenome(genomeHash []byte, ipfsHash string, label GenomeLabel) *Genome

// ComputeHash computes the deterministic hash of features
func (g *Genome) ComputeHash() []byte

// ========== Exploit Submission ==========

// ExploitSubmission represents a Red-Team DAO submission
type ExploitSubmission struct {
    ID                 uuid.UUID        `json:"id" db:"id"`
    ResearcherAddress  string           `json:"researcher_address" db:"researcher_address"`
    TargetContract     string           `json:"target_contract" db:"target_contract"`
    ProofHash          []byte           `json:"proof_hash" db:"proof_hash"`
    GenomeID           *uuid.UUID       `json:"genome_id,omitempty" db:"genome_id"`
    Description        string           `json:"description" db:"description"`
    Severity           Severity         `json:"severity" db:"severity"`
    BountyAmount       uint64           `json:"bounty_amount" db:"bounty_amount"` // in wei
    Status             SubmissionStatus `json:"status" db:"status"`
    CreatedAt          time.Time        `json:"created_at" db:"created_at"`
    VerifiedAt         *time.Time       `json:"verified_at,omitempty" db:"verified_at"`
    PaidAt             *time.Time       `json:"paid_at,omitempty" db:"paid_at"`
    TxHash             string           `json:"tx_hash,omitempty" db:"tx_hash"`
    VotesFor           int              `json:"votes_for" db:"votes_for"`
    VotesAgainst       int              `json:"votes_against" db:"votes_against"`
}

// Severity represents exploit severity level
type Severity string

const (
    SeverityLow      Severity = "low"
    SeverityMedium   Severity = "medium"
    SeverityHigh     Severity = "high"
    SeverityCritical Severity = "critical"
)

// Multiplier returns the bounty multiplier for this severity
func (s Severity) Multiplier() float64 {
    switch s {
    case SeverityLow:
        return 1.0
    case SeverityMedium:
        return 2.0
    case SeverityHigh:
        return 5.0
    case SeverityCritical:
        return 10.0
    default:
        return 1.0
    }
}

// SubmissionStatus represents the status of an exploit submission
type SubmissionStatus string

const (
    SubmissionStatusPending   SubmissionStatus = "pending"
    SubmissionStatusVerifying SubmissionStatus = "verifying"
    SubmissionStatusVerified  SubmissionStatus = "verified"
    SubmissionStatusRejected  SubmissionStatus = "rejected"
    SubmissionStatusPaid      SubmissionStatus = "paid"
    SubmissionStatusDisputed  SubmissionStatus = "disputed"
)

// NewExploitSubmission creates a new exploit submission
func NewExploitSubmission(
    researcherAddress string,
    targetContract string,
    proofHash []byte,
    description string,
    severity Severity,
) *ExploitSubmission

// ========== API Key ==========

// APIKey represents an API key for authentication
type APIKey struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    KeyHash   []byte     `json:"-" db:"key_hash"` // Never expose
    UserID    uuid.UUID  `json:"user_id" db:"user_id"`
    Tier      APIKeyTier `json:"tier" db:"tier"`
    RateLimit int        `json:"rate_limit" db:"rate_limit"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
    Revoked   bool       `json:"revoked" db:"revoked"`
    LastUsed  *time.Time `json:"last_used,omitempty" db:"last_used"`
}

// APIKeyTier represents the API key tier
type APIKeyTier string

const (
    APIKeyTierFree       APIKeyTier = "free"
    APIKeyTierPaid       APIKeyTier = "paid"
    APIKeyTierEnterprise APIKeyTier = "enterprise"
)

// NewAPIKey creates a new API key (returns key and hash)
func NewAPIKey(userID uuid.UUID, tier APIKeyTier) (*APIKey, string, error)

// Validate checks if key is valid (not expired, not revoked)
func (k *APIKey) Validate() error

// ========== Behavioral Features ==========

// BehavioralFeatures represents extracted wallet behavior features
type BehavioralFeatures struct {
    WalletAddress        string    `json:"wallet_address"`
    ComputedAt           time.Time `json:"computed_at"`
    TxCount              uint32    `json:"tx_count"`
    AvgTxInterval        float64   `json:"avg_tx_interval"` // seconds
    TxIntervalVariance   float64   `json:"tx_interval_variance"`
    GasVariance          float64   `json:"gas_variance"`
    InteractionDiversity uint32    `json:"interaction_diversity"`
    UniqueContracts      uint32    `json:"unique_contracts"`
    LastActivity         time.Time `json:"last_activity"`
    
    // Additional features
    AvgGasUsed           float64   `json:"avg_gas_used"`
    MaxGasUsed           uint64    `json:"max_gas_used"`
    MinGasUsed           uint64    `json:"min_gas_used"`
    TotalValueTransferred string   `json:"total_value_transferred"` // wei as string
    ERC20Interactions    uint32    `json:"erc20_interactions"`
    DeFiInteractions     uint32    `json:"defi_interactions"`
    NFTInteractions      uint32    `json:"nft_interactions"`
    
    // Time-based features
    TimeOfDayDistribution [24]int  `json:"time_of_day_distribution"`
    DayOfWeekDistribution [7]int   `json:"day_of_week_distribution"`
}

// ToFeatureVector converts to a slice for ML inference
func (f *BehavioralFeatures) ToFeatureVector() []float32 {
    return []float32{
        float32(f.TxCount),
        float32(f.AvgTxInterval),
        float32(f.TxIntervalVariance),
        float32(f.GasVariance),
        float32(f.InteractionDiversity),
        float32(f.UniqueContracts),
        float32(f.AvgGasUsed),
        float32(f.ERC20Interactions),
        float32(f.DeFiInteractions),
        float32(f.NFTInteractions),
    }
}

// Normalize normalizes features to 0-1 range
func (f *BehavioralFeatures) Normalize() *BehavioralFeatures
```

---

### 1.3 Events (`backend/internal/models/events.go`)

```go
// events.go
package models

import (
    "time"
    "github.com/google/uuid"
)

// ========== Event Types ==========

// EventType represents the type of domain event
type EventType string

const (
    EventTypeProofVerified       EventType = "proof.verified"
    EventTypeProofRejected       EventType = "proof.rejected"
    EventTypeThreatDetected      EventType = "threat.detected"
    EventTypeThreatPublished     EventType = "threat.published"
    EventTypeGenomeRegistered    EventType = "genome.registered"
    EventTypeExploitSubmitted    EventType = "exploit.submitted"
    EventTypeExploitVerified     EventType = "exploit.verified"
    EventTypeBountyPaid          EventType = "bounty.paid"
    EventTypeUserRegistered      EventType = "user.registered"
    EventTypeAPIKeyCreated       EventType = "apikey.created"
    EventTypeAPIKeyRevoked       EventType = "apikey.revoked"
)

// ========== Base Event ==========

// Event is the base interface for all domain events
type Event interface {
    GetID() uuid.UUID
    GetType() EventType
    GetTimestamp() time.Time
    GetAggregateID() string
    ToJSON() ([]byte, error)
}

// BaseEvent contains common event fields
type BaseEvent struct {
    ID          uuid.UUID `json:"id"`
    Type        EventType `json:"type"`
    Timestamp   time.Time `json:"timestamp"`
    AggregateID string    `json:"aggregate_id"`
    Version     int       `json:"version"`
    TraceID     string    `json:"trace_id,omitempty"`
}

func (e *BaseEvent) GetID() uuid.UUID         { return e.ID }
func (e *BaseEvent) GetType() EventType       { return e.Type }
func (e *BaseEvent) GetTimestamp() time.Time  { return e.Timestamp }
func (e *BaseEvent) GetAggregateID() string   { return e.AggregateID }

// ========== Specific Events ==========

// ProofVerifiedEvent is emitted when a human-proof is verified
type ProofVerifiedEvent struct {
    BaseEvent
    UserID        uuid.UUID `json:"user_id"`
    WalletAddress string    `json:"wallet_address"`
    ProofHash     string    `json:"proof_hash"`
    TxHash        string    `json:"tx_hash"`
    RiskScore     float64   `json:"risk_score"`
    VerifiedAt    time.Time `json:"verified_at"`
}

// NewProofVerifiedEvent creates a new proof verified event
func NewProofVerifiedEvent(proof *HumanProof, user *User, txHash string, riskScore float64) *ProofVerifiedEvent

// ThreatDetectedEvent is emitted when a threat is detected
type ThreatDetectedEvent struct {
    BaseEvent
    EntityAddress string                 `json:"entity_address"`
    SignalType    SignalType             `json:"signal_type"`
    RiskScore     int                    `json:"risk_score"`
    Confidence    float64                `json:"confidence"`
    Source        string                 `json:"source"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NewThreatDetectedEvent creates a new threat detected event
func NewThreatDetectedEvent(signal *ThreatSignal) *ThreatDetectedEvent

// GenomeRegisteredEvent is emitted when a genome is registered
type GenomeRegisteredEvent struct {
    BaseEvent
    GenomeHash      string      `json:"genome_hash"`
    IPFSHash        string      `json:"ipfs_hash"`
    ContractAddress string      `json:"contract_address,omitempty"`
    Label           GenomeLabel `json:"label"`
    TxHash          string      `json:"tx_hash,omitempty"`
}

// NewGenomeRegisteredEvent creates a new genome registered event
func NewGenomeRegisteredEvent(genome *Genome) *GenomeRegisteredEvent

// ExploitSubmittedEvent is emitted when an exploit is submitted
type ExploitSubmittedEvent struct {
    BaseEvent
    SubmissionID      uuid.UUID `json:"submission_id"`
    ResearcherAddress string    `json:"researcher_address"`
    TargetContract    string    `json:"target_contract"`
    Severity          Severity  `json:"severity"`
    EstimatedBounty   uint64    `json:"estimated_bounty"`
}

// NewExploitSubmittedEvent creates a new exploit submitted event
func NewExploitSubmittedEvent(submission *ExploitSubmission) *ExploitSubmittedEvent

// BountyPaidEvent is emitted when a bounty is paid
type BountyPaidEvent struct {
    BaseEvent
    SubmissionID      uuid.UUID `json:"submission_id"`
    ResearcherAddress string    `json:"researcher_address"`
    Amount            uint64    `json:"amount"` // wei
    TxHash            string    `json:"tx_hash"`
}

// NewBountyPaidEvent creates a new bounty paid event
func NewBountyPaidEvent(submission *ExploitSubmission, txHash string) *BountyPaidEvent
```

---

### 1.4 Repository Interfaces (`backend/internal/repository/`)

```go
// repository.go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/vigilum/backend/internal/models"
)

// ========== User Repository ==========

// UserRepository defines user data access operations
type UserRepository interface {
    // Create creates a new user
    Create(ctx context.Context, user *models.User) error
    
    // GetByID retrieves a user by ID
    GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    
    // GetByWalletAddress retrieves a user by wallet address
    GetByWalletAddress(ctx context.Context, address string) (*models.User, error)
    
    // Update updates a user
    Update(ctx context.Context, user *models.User) error
    
    // UpdateRiskScore updates only the risk score
    UpdateRiskScore(ctx context.Context, id uuid.UUID, score float64) error
    
    // Delete soft-deletes a user
    Delete(ctx context.Context, id uuid.UUID) error
    
    // Exists checks if a user exists by wallet address
    Exists(ctx context.Context, address string) (bool, error)
}

// ========== Human Proof Repository ==========

// HumanProofRepository defines proof data access operations
type HumanProofRepository interface {
    // Create creates a new proof
    Create(ctx context.Context, proof *models.HumanProof) error
    
    // GetByID retrieves a proof by ID
    GetByID(ctx context.Context, id uuid.UUID) (*models.HumanProof, error)
    
    // GetByProofHash retrieves a proof by its hash
    GetByProofHash(ctx context.Context, hash []byte) (*models.HumanProof, error)
    
    // GetByUserID retrieves all proofs for a user
    GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.HumanProof, error)
    
    // MarkVerified marks a proof as verified
    MarkVerified(ctx context.Context, id uuid.UUID, verifierAddress, txHash string) error
    
    // CountByUser returns the number of proofs for a user
    CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
    
    // GetLatestByUser returns the latest proof for a user
    GetLatestByUser(ctx context.Context, userID uuid.UUID) (*models.HumanProof, error)
}

// ========== Threat Signal Repository ==========

// ThreatSignalRepository defines threat signal data access operations
type ThreatSignalRepository interface {
    // Create creates a new threat signal
    Create(ctx context.Context, signal *models.ThreatSignal) error
    
    // GetByID retrieves a signal by ID
    GetByID(ctx context.Context, id uuid.UUID) (*models.ThreatSignal, error)
    
    // GetByEntityAddress retrieves all signals for an entity
    GetByEntityAddress(ctx context.Context, address string, limit, offset int) ([]*models.ThreatSignal, error)
    
    // GetUnpublished retrieves signals not yet published on-chain
    GetUnpublished(ctx context.Context, limit int) ([]*models.ThreatSignal, error)
    
    // MarkPublished marks a signal as published
    MarkPublished(ctx context.Context, id uuid.UUID, txHash string) error
    
    // GetLatestByEntity returns the latest signal for an entity
    GetLatestByEntity(ctx context.Context, address string) (*models.ThreatSignal, error)
    
    // GetHighRisk retrieves all high-risk signals
    GetHighRisk(ctx context.Context, minScore int, limit, offset int) ([]*models.ThreatSignal, error)
    
    // AggregateByEntity aggregates signals for an entity
    AggregateByEntity(ctx context.Context, address string) (*AggregatedSignal, error)
}

// AggregatedSignal represents aggregated signals for an entity
type AggregatedSignal struct {
    EntityAddress   string  `json:"entity_address"`
    MaxRiskScore    int     `json:"max_risk_score"`
    AvgRiskScore    float64 `json:"avg_risk_score"`
    SignalCount     int     `json:"signal_count"`
    LatestTimestamp time.Time `json:"latest_timestamp"`
}

// ========== Genome Repository ==========

// GenomeRepository defines genome data access operations
type GenomeRepository interface {
    // Create creates a new genome
    Create(ctx context.Context, genome *models.Genome) error
    
    // GetByID retrieves a genome by ID
    GetByID(ctx context.Context, id uuid.UUID) (*models.Genome, error)
    
    // GetByGenomeHash retrieves a genome by its hash
    GetByGenomeHash(ctx context.Context, hash []byte) (*models.Genome, error)
    
    // GetByContractAddress retrieves genomes for a contract
    GetByContractAddress(ctx context.Context, address string) ([]*models.Genome, error)
    
    // GetByLabel retrieves genomes by label
    GetByLabel(ctx context.Context, label models.GenomeLabel, limit, offset int) ([]*models.Genome, error)
    
    // MarkRegisteredOnChain marks a genome as registered on-chain
    MarkRegisteredOnChain(ctx context.Context, id uuid.UUID, txHash string) error
    
    // Exists checks if a genome hash exists
    Exists(ctx context.Context, hash []byte) (bool, error)
    
    // FindSimilar finds genomes similar to the given hash
    FindSimilar(ctx context.Context, hash []byte, threshold float64, limit int) ([]*GenomeSimilarity, error)
}

// GenomeSimilarity represents a similar genome match
type GenomeSimilarity struct {
    Genome          *models.Genome `json:"genome"`
    SimilarityScore float64        `json:"similarity_score"`
}

// ========== Exploit Submission Repository ==========

// ExploitSubmissionRepository defines exploit submission data access operations
type ExploitSubmissionRepository interface {
    // Create creates a new submission
    Create(ctx context.Context, submission *models.ExploitSubmission) error
    
    // GetByID retrieves a submission by ID
    GetByID(ctx context.Context, id uuid.UUID) (*models.ExploitSubmission, error)
    
    // GetByResearcher retrieves submissions by researcher
    GetByResearcher(ctx context.Context, address string, limit, offset int) ([]*models.ExploitSubmission, error)
    
    // GetByStatus retrieves submissions by status
    GetByStatus(ctx context.Context, status models.SubmissionStatus, limit, offset int) ([]*models.ExploitSubmission, error)
    
    // GetPending retrieves pending submissions for verification
    GetPending(ctx context.Context, limit int) ([]*models.ExploitSubmission, error)
    
    // UpdateStatus updates submission status
    UpdateStatus(ctx context.Context, id uuid.UUID, status models.SubmissionStatus) error
    
    // UpdateVotes updates vote counts
    UpdateVotes(ctx context.Context, id uuid.UUID, votesFor, votesAgainst int) error
    
    // MarkPaid marks a submission as paid
    MarkPaid(ctx context.Context, id uuid.UUID, txHash string) error
}

// ========== API Key Repository ==========

// APIKeyRepository defines API key data access operations
type APIKeyRepository interface {
    // Create creates a new API key
    Create(ctx context.Context, key *models.APIKey) error
    
    // GetByKeyHash retrieves a key by its hash
    GetByKeyHash(ctx context.Context, hash []byte) (*models.APIKey, error)
    
    // GetByUserID retrieves all keys for a user
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error)
    
    // Revoke revokes an API key
    Revoke(ctx context.Context, id uuid.UUID) error
    
    // UpdateLastUsed updates the last used timestamp
    UpdateLastUsed(ctx context.Context, id uuid.UUID) error
    
    // DeleteExpired deletes expired keys
    DeleteExpired(ctx context.Context) (int64, error)
}
```

---

### 1.5 Service Interfaces (`backend/internal/services/`)

```go
// firewall_service.go
package services

import (
    "context"
    "github.com/google/uuid"
    "github.com/vigilum/backend/internal/models"
)

// ========== Identity Firewall Service ==========

// FirewallService defines the identity firewall business logic
type FirewallService interface {
    // VerifyProof verifies a human-behavior ZK proof
    VerifyProof(ctx context.Context, req *VerifyProofRequest) (*VerifyProofResponse, error)
    
    // GenerateChallenge generates a challenge for proof generation
    GenerateChallenge(ctx context.Context) (*ChallengeResponse, error)
    
    // GetRiskScore gets the risk score for an address
    GetRiskScore(ctx context.Context, address string) (*RiskScoreResponse, error)
    
    // GetStats returns service statistics
    GetStats(ctx context.Context) (*FirewallStats, error)
}

// VerifyProofRequest is the request to verify a proof
type VerifyProofRequest struct {
    Proof        []byte            `json:"proof" validate:"required,min=32"`
    PublicInputs map[string]string `json:"public_inputs" validate:"required"`
}

// Validate validates the request
func (r *VerifyProofRequest) Validate() error

// VerifyProofResponse is the response from proof verification
type VerifyProofResponse struct {
    Verified   bool      `json:"verified"`
    ProofHash  string    `json:"proof_hash"`
    TxHash     string    `json:"tx_hash,omitempty"`
    RiskScore  float64   `json:"risk_score"`
    ExpiresAt  time.Time `json:"expires_at"`
    Error      string    `json:"error,omitempty"`
}

// ChallengeResponse is the response containing a challenge
type ChallengeResponse struct {
    ChallengeID string    `json:"challenge_id"`
    Challenge   string    `json:"challenge"`
    ExpiresAt   time.Time `json:"expires_at"`
}

// RiskScoreResponse is the response containing risk information
type RiskScoreResponse struct {
    Address    string              `json:"address"`
    RiskScore  float64             `json:"risk_score"` // 0.0-1.0
    RiskLevel  string              `json:"risk_level"` // low, medium, high, critical
    Signals    []*SignalSummary    `json:"signals"`
    UpdatedAt  time.Time           `json:"updated_at"`
}

// SignalSummary is a summary of a threat signal
type SignalSummary struct {
    Type       string    `json:"type"`
    Source     string    `json:"source"`
    Confidence float64   `json:"confidence"`
    Timestamp  time.Time `json:"timestamp"`
}

// FirewallStats contains service statistics
type FirewallStats struct {
    TotalProofsVerified   int64   `json:"total_proofs_verified"`
    ProofsVerifiedToday   int64   `json:"proofs_verified_today"`
    AvgVerificationTimeMs float64 `json:"avg_verification_time_ms"`
    HighRiskAddresses     int64   `json:"high_risk_addresses"`
}

// ========== Firewall Service Implementation ==========

type firewallService struct {
    userRepo      repository.UserRepository
    proofRepo     repository.HumanProofRepository
    signalRepo    repository.ThreatSignalRepository
    ethClient     ethereum.Client
    zkProver      zkprover.Client
    inferenceService InferenceService
    cache         cache.Cache
    config        *config.Config
    logger        *zap.Logger
}

// NewFirewallService creates a new firewall service
func NewFirewallService(
    userRepo repository.UserRepository,
    proofRepo repository.HumanProofRepository,
    signalRepo repository.ThreatSignalRepository,
    ethClient ethereum.Client,
    zkProver zkprover.Client,
    inferenceService InferenceService,
    cache cache.Cache,
    config *config.Config,
    logger *zap.Logger,
) FirewallService {
    return &firewallService{
        userRepo:         userRepo,
        proofRepo:        proofRepo,
        signalRepo:       signalRepo,
        ethClient:        ethClient,
        zkProver:         zkProver,
        inferenceService: inferenceService,
        cache:            cache,
        config:           config,
        logger:           logger,
    }
}
```

```go
// oracle_service.go
package services

// ========== Threat Oracle Service ==========

// OracleService defines the threat oracle business logic
type OracleService interface {
    // GetSignals gets all signals for an address
    GetSignals(ctx context.Context, address string) (*SignalsResponse, error)
    
    // Subscribe subscribes to threat signals via webhook
    Subscribe(ctx context.Context, req *SubscribeRequest) (*SubscribeResponse, error)
    
    // Unsubscribe removes a subscription
    Unsubscribe(ctx context.Context, subscriptionID string) error
    
    // PublishSignal publishes a signal (internal use)
    PublishSignal(ctx context.Context, signal *models.ThreatSignal) error
    
    // AggregateSignals aggregates signals from all feeds
    AggregateSignals(ctx context.Context) error
    
    // PublishToChain publishes pending signals on-chain
    PublishToChain(ctx context.Context) error
}

// SignalsResponse is the response containing signals
type SignalsResponse struct {
    Address          string                `json:"address"`
    Signals          []*models.ThreatSignal `json:"signals"`
    OnChainRiskScore int                   `json:"on_chain_risk_score"`
}

// SubscribeRequest is the request to subscribe to signals
type SubscribeRequest struct {
    WebhookURL  string   `json:"webhook_url" validate:"required,url"`
    Addresses   []string `json:"addresses" validate:"required,min=1,dive,eth_addr"`
    SignalTypes []string `json:"signal_types,omitempty"`
}

// SubscribeResponse is the response from subscription
type SubscribeResponse struct {
    SubscriptionID string `json:"subscription_id"`
    Active         bool   `json:"active"`
}

// ========== Oracle Service Implementation ==========

type oracleService struct {
    signalRepo    repository.ThreatSignalRepository
    ethClient     ethereum.Client
    feedIngester  FeedIngester
    publisher     SignalPublisher
    webhookSender WebhookSender
    cache         cache.Cache
    config        *config.Config
    logger        *zap.Logger
}
```

```go
// genome_service.go
package services

// ========== Genome Analyzer Service ==========

// GenomeService defines the genome analyzer business logic
type GenomeService interface {
    // AnalyzeContract analyzes a contract and generates a genome
    AnalyzeContract(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error)
    
    // GetAnalysisStatus gets the status of an analysis
    GetAnalysisStatus(ctx context.Context, analysisID string) (*AnalysisStatusResponse, error)
    
    // GetGenome gets a genome by hash
    GetGenome(ctx context.Context, genomeHash string) (*models.Genome, error)
    
    // FindSimilar finds similar genomes
    FindSimilar(ctx context.Context, genomeHash string, threshold float64) ([]*repository.GenomeSimilarity, error)
}

// AnalyzeRequest is the request to analyze a contract
type AnalyzeRequest struct {
    ContractAddress string   `json:"contract_address" validate:"required,eth_addr"`
    Priority        Priority `json:"priority" default:"normal"`
}

// Priority represents analysis priority
type Priority string

const (
    PriorityNormal Priority = "normal"
    PriorityHigh   Priority = "high"
    PriorityUrgent Priority = "urgent"
)

// AnalyzeResponse is the response from analysis request
type AnalyzeResponse struct {
    AnalysisID          string    `json:"analysis_id"`
    Status              string    `json:"status"`
    EstimatedCompletion time.Time `json:"estimated_completion"`
}

// AnalysisStatusResponse is the response for analysis status
type AnalysisStatusResponse struct {
    AnalysisID  string                       `json:"analysis_id"`
    Status      string                       `json:"status"` // queued, processing, completed, failed
    GenomeHash  string                       `json:"genome_hash,omitempty"`
    IPFSHash    string                       `json:"ipfs_hash,omitempty"`
    Label       models.GenomeLabel           `json:"label,omitempty"`
    Similarity  []*repository.GenomeSimilarity `json:"similarity,omitempty"`
    Error       string                       `json:"error,omitempty"`
}

// ========== Genome Service Implementation ==========

type genomeService struct {
    genomeRepo     repository.GenomeRepository
    ethClient      ethereum.Client
    ipfsClient     ipfs.Client
    analyzer       GenomeAnalyzer
    temporalClient temporal.Client
    config         *config.Config
    logger         *zap.Logger
}
```

```go
// redteam_service.go
package services

// ========== Red-Team DAO Service ==========

// RedTeamService defines the Red-Team DAO business logic
type RedTeamService interface {
    // SubmitExploit submits a new exploit proof
    SubmitExploit(ctx context.Context, req *SubmitExploitRequest) (*SubmitExploitResponse, error)
    
    // GetSubmission gets a submission by ID
    GetSubmission(ctx context.Context, submissionID string) (*models.ExploitSubmission, error)
    
    // GetSubmissionsByResearcher gets submissions by researcher
    GetSubmissionsByResearcher(ctx context.Context, address string, limit, offset int) ([]*models.ExploitSubmission, error)
    
    // Vote votes on a submission
    Vote(ctx context.Context, req *VoteRequest) error
    
    // ClaimBounty claims the bounty for a verified submission
    ClaimBounty(ctx context.Context, submissionID string) (*ClaimBountyResponse, error)
    
    // GetLeaderboard gets the researcher leaderboard
    GetLeaderboard(ctx context.Context, limit int) ([]*LeaderboardEntry, error)
}

// SubmitExploitRequest is the request to submit an exploit
type SubmitExploitRequest struct {
    TargetContract string          `json:"target_contract" validate:"required,eth_addr"`
    Proof          []byte          `json:"proof" validate:"required,min=32"`
    Description    string          `json:"description" validate:"required,min=10,max=5000"`
    Severity       models.Severity `json:"severity" validate:"required,oneof=low medium high critical"`
}

// SubmitExploitResponse is the response from exploit submission
type SubmitExploitResponse struct {
    SubmissionID    string `json:"submission_id"`
    Status          string `json:"status"`
    EstimatedBounty uint64 `json:"estimated_bounty"` // wei
}

// VoteRequest is the request to vote on a submission
type VoteRequest struct {
    SubmissionID string `json:"submission_id" validate:"required,uuid"`
    Approve      bool   `json:"approve"`
}

// ClaimBountyResponse is the response from claiming a bounty
type ClaimBountyResponse struct {
    SubmissionID string `json:"submission_id"`
    Amount       uint64 `json:"amount"` // wei
    TxHash       string `json:"tx_hash"`
}

// LeaderboardEntry represents a researcher on the leaderboard
type LeaderboardEntry struct {
    Rank              int     `json:"rank"`
    ResearcherAddress string  `json:"researcher_address"`
    TotalBounties     uint64  `json:"total_bounties"` // wei
    SubmissionCount   int     `json:"submission_count"`
    VerifiedCount     int     `json:"verified_count"`
    ReputationScore   float64 `json:"reputation_score"`
}
```

```go
// inference_service.go
package services

// ========== ML Inference Service ==========

// InferenceService defines ML inference operations
type InferenceService interface {
    // PredictHumanScore predicts the human likelihood score
    PredictHumanScore(ctx context.Context, features *models.BehavioralFeatures) (float64, error)
    
    // PredictAnomalyScore predicts the anomaly score
    PredictAnomalyScore(ctx context.Context, features *models.BehavioralFeatures) (float64, error)
    
    // Reload reloads the ML models
    Reload(ctx context.Context) error
    
    // GetModelInfo returns model metadata
    GetModelInfo(ctx context.Context) (*ModelInfo, error)
}

// ModelInfo contains model metadata
type ModelInfo struct {
    HumanClassifierVersion string    `json:"human_classifier_version"`
    AnomalyModelVersion    string    `json:"anomaly_model_version"`
    LoadedAt               time.Time `json:"loaded_at"`
    Accuracy               float64   `json:"accuracy"`
}

// ========== Inference Service Implementation ==========

type inferenceService struct {
    humanClassifierPath string
    anomalyModelPath    string
    humanSession        *ort.Session
    anomalySession      *ort.Session
    modelInfo           *ModelInfo
    mu                  sync.RWMutex
    logger              *zap.Logger
}

// NewInferenceService creates a new inference service
func NewInferenceService(humanClassifierPath, anomalyModelPath string, logger *zap.Logger) (InferenceService, error)
```

---

### 1.6 HTTP Handlers (`backend/internal/handlers/`)

```go
// firewall_handler.go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/vigilum/backend/internal/services"
)

// FirewallHandler handles identity firewall HTTP requests
type FirewallHandler struct {
    service services.FirewallService
    logger  *zap.Logger
}

// NewFirewallHandler creates a new firewall handler
func NewFirewallHandler(service services.FirewallService, logger *zap.Logger) *FirewallHandler

// RegisterRoutes registers the firewall routes
func (h *FirewallHandler) RegisterRoutes(r *gin.RouterGroup) {
    r.POST("/verify-proof", h.VerifyProof)
    r.GET("/challenge", h.GenerateChallenge)
    r.GET("/risk/:address", h.GetRiskScore)
    r.GET("/stats", h.GetStats)
}

// VerifyProof handles POST /firewall/verify-proof
// @Summary Verify a human-behavior ZK proof
// @Tags Firewall
// @Accept json
// @Produce json
// @Param request body services.VerifyProofRequest true "Proof verification request"
// @Success 200 {object} APIResponse{data=services.VerifyProofResponse}
// @Failure 400 {object} APIResponse{error=APIError}
// @Failure 401 {object} APIResponse{error=APIError}
// @Failure 429 {object} APIResponse{error=APIError}
// @Failure 500 {object} APIResponse{error=APIError}
// @Router /firewall/verify-proof [post]
func (h *FirewallHandler) VerifyProof(c *gin.Context)

// GenerateChallenge handles GET /firewall/challenge
// @Summary Generate a challenge for proof generation
// @Tags Firewall
// @Produce json
// @Success 200 {object} APIResponse{data=services.ChallengeResponse}
// @Failure 500 {object} APIResponse{error=APIError}
// @Router /firewall/challenge [get]
func (h *FirewallHandler) GenerateChallenge(c *gin.Context)

// GetRiskScore handles GET /firewall/risk/:address
// @Summary Get risk score for an address
// @Tags Firewall
// @Produce json
// @Param address path string true "Wallet or contract address"
// @Success 200 {object} APIResponse{data=services.RiskScoreResponse}
// @Failure 400 {object} APIResponse{error=APIError}
// @Failure 404 {object} APIResponse{error=APIError}
// @Failure 500 {object} APIResponse{error=APIError}
// @Router /firewall/risk/{address} [get]
func (h *FirewallHandler) GetRiskScore(c *gin.Context)

// GetStats handles GET /firewall/stats
// @Summary Get service statistics
// @Tags Firewall
// @Produce json
// @Success 200 {object} APIResponse{data=services.FirewallStats}
// @Failure 500 {object} APIResponse{error=APIError}
// @Router /firewall/stats [get]
func (h *FirewallHandler) GetStats(c *gin.Context)
```

```go
// response.go
package handlers

import (
    "time"
)

// APIResponse is the standard API response format
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an API error
type APIError struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Details map[string]string `json:"details,omitempty"`
}

// Error codes
const (
    ErrCodeBadRequest          = "BAD_REQUEST"
    ErrCodeUnauthorized        = "UNAUTHORIZED"
    ErrCodeForbidden           = "FORBIDDEN"
    ErrCodeNotFound            = "NOT_FOUND"
    ErrCodeConflict            = "CONFLICT"
    ErrCodeRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"
    ErrCodeInternalError       = "INTERNAL_ERROR"
    ErrCodeValidationFailed    = "VALIDATION_FAILED"
    ErrCodeProofInvalid        = "PROOF_INVALID"
    ErrCodeContractCallFailed  = "CONTRACT_CALL_FAILED"
    ErrCodeIPFSError           = "IPFS_ERROR"
)

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) *APIResponse {
    return &APIResponse{
        Success:   true,
        Data:      data,
        Timestamp: time.Now().UTC(),
    }
}

// NewErrorResponse creates an error response
func NewErrorResponse(code, message string, details map[string]string) *APIResponse {
    return &APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
            Details: details,
        },
        Timestamp: time.Now().UTC(),
    }
}
```

---

### 1.7 Middleware (`backend/internal/middleware/`)

```go
// auth.go
package middleware

import (
    "github.com/gin-gonic/gin"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
    apiKeyRepo repository.APIKeyRepository
    cache      cache.Cache
    logger     *zap.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(apiKeyRepo repository.APIKeyRepository, cache cache.Cache, logger *zap.Logger) *AuthMiddleware

// Authenticate authenticates the request
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc

// RequireAPIKey requires a valid API key
func (m *AuthMiddleware) RequireAPIKey() gin.HandlerFunc

// OptionalAPIKey checks for API key but doesn't require it
func (m *AuthMiddleware) OptionalAPIKey() gin.HandlerFunc

// extractAPIKey extracts the API key from the request
func (m *AuthMiddleware) extractAPIKey(c *gin.Context) string
```

```go
// rate_limiter.go
package middleware

// RateLimiterMiddleware handles rate limiting
type RateLimiterMiddleware struct {
    redis  *redis.Client
    config *config.RateLimiterConfig
    logger *zap.Logger
}

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(redis *redis.Client, config *config.RateLimiterConfig, logger *zap.Logger) *RateLimiterMiddleware

// RateLimit applies rate limiting based on IP or API key
func (m *RateLimiterMiddleware) RateLimit() gin.HandlerFunc

// getRateLimit returns the rate limit for the request
func (m *RateLimiterMiddleware) getRateLimit(c *gin.Context) int

// getKey returns the rate limit key for the request
func (m *RateLimiterMiddleware) getKey(c *gin.Context) string
```

```go
// logging.go
package middleware

// LoggingMiddleware handles request logging
type LoggingMiddleware struct {
    logger *zap.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware

// RequestLogger logs all requests with timing
func (m *LoggingMiddleware) RequestLogger() gin.HandlerFunc

// logFields returns the log fields for the request
func (m *LoggingMiddleware) logFields(c *gin.Context, latency time.Duration) []zap.Field
```

```go
// errors.go
package middleware

// ErrorMiddleware handles error recovery and formatting
type ErrorMiddleware struct {
    logger *zap.Logger
}

// NewErrorMiddleware creates a new error middleware
func NewErrorMiddleware(logger *zap.Logger) *ErrorMiddleware

// Recovery recovers from panics
func (m *ErrorMiddleware) Recovery() gin.HandlerFunc

// ErrorHandler handles errors from handlers
func (m *ErrorMiddleware) ErrorHandler() gin.HandlerFunc
```

```go
// tracing.go
package middleware

// TracingMiddleware handles distributed tracing
type TracingMiddleware struct {
    tracer opentracing.Tracer
}

// NewTracingMiddleware creates a new tracing middleware
func NewTracingMiddleware(tracer opentracing.Tracer) *TracingMiddleware

// Trace creates spans for all requests
func (m *TracingMiddleware) Trace() gin.HandlerFunc
```

---

### 1.8 Ethereum Integration (`backend/internal/integration/`)

```go
// ethereum.go
package integration

import (
    "context"
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

// EthereumClient is the interface for Ethereum interactions
type EthereumClient interface {
    // Connection
    Connect(ctx context.Context) error
    Close()
    
    // Identity Firewall
    VerifyHumanProof(ctx context.Context, proof []byte) (bool, string, error)
    HasVerifiedProof(ctx context.Context, proofHash [32]byte) (bool, error)
    
    // Threat Oracle
    UpdateRiskScore(ctx context.Context, target common.Address, score *big.Int) (string, error)
    GetRiskScore(ctx context.Context, target common.Address) (*big.Int, error)
    IsHighRisk(ctx context.Context, target common.Address) (bool, error)
    
    // Malware Genome DB
    RegisterGenome(ctx context.Context, genomeHash [32]byte, ipfsHash, label string) (string, error)
    HasGenome(ctx context.Context, genomeHash [32]byte) (bool, error)
    GetGenome(ctx context.Context, genomeHash [32]byte) (string, string, *big.Int, error)
    
    // Red-Team DAO
    SubmitExploit(ctx context.Context, target common.Address, proof []byte, description string) (string, error)
    Vote(ctx context.Context, submissionID [32]byte, approve bool) (string, error)
    ClaimBounty(ctx context.Context, submissionID [32]byte) (string, error)
    
    // Events
    SubscribeToProofVerified(ctx context.Context, ch chan<- *ProofVerifiedEvent) error
    SubscribeToRiskUpdated(ctx context.Context, ch chan<- *RiskUpdatedEvent) error
    
    // Utilities
    GetGasPrice(ctx context.Context) (*big.Int, error)
    WaitForTransaction(ctx context.Context, txHash string) (*TransactionReceipt, error)
}

// ProofVerifiedEvent represents a ProofVerified event from the contract
type ProofVerifiedEvent struct {
    User       common.Address
    ProofHash  [32]byte
    Timestamp  *big.Int
    TxHash     string
    BlockNumber uint64
}

// RiskUpdatedEvent represents a RiskUpdated event from the contract
type RiskUpdatedEvent struct {
    Target      common.Address
    RiskScore   *big.Int
    TxHash      string
    BlockNumber uint64
}

// TransactionReceipt represents a transaction receipt
type TransactionReceipt struct {
    TxHash      string
    BlockNumber uint64
    GasUsed     uint64
    Status      uint64
}

// ========== Ethereum Client Implementation ==========

type ethereumClient struct {
    config              *config.EthereumConfig
    client              *ethclient.Client
    identityFirewall    *contracts.IdentityFirewall
    threatOracle        *contracts.ThreatOracle
    malwareGenomeDB     *contracts.MalwareGenomeDB
    redTeamDAO          *contracts.RedTeamDAO
    privateKey          *ecdsa.PrivateKey
    fromAddress         common.Address
    mu                  sync.RWMutex
    logger              *zap.Logger
}

// NewEthereumClient creates a new Ethereum client
func NewEthereumClient(config *config.EthereumConfig, logger *zap.Logger) (EthereumClient, error)
```

```go
// ipfs.go
package integration

// IPFSClient is the interface for IPFS interactions
type IPFSClient interface {
    // Pin pins content to IPFS
    Pin(ctx context.Context, content []byte) (string, error)
    
    // PinJSON pins JSON content to IPFS
    PinJSON(ctx context.Context, data interface{}) (string, error)
    
    // Get retrieves content from IPFS
    Get(ctx context.Context, cid string) ([]byte, error)
    
    // Unpin unpins content from IPFS
    Unpin(ctx context.Context, cid string) error
    
    // IsPinned checks if content is pinned
    IsPinned(ctx context.Context, cid string) (bool, error)
}

// ========== IPFS Client Implementation ==========

type ipfsClient struct {
    config        *config.IPFSConfig
    kuboClient    *shell.Shell
    infuraClient  *infura.Client
    pinataClient  *pinata.Client
    logger        *zap.Logger
}

// NewIPFSClient creates a new IPFS client
func NewIPFSClient(config *config.IPFSConfig, logger *zap.Logger) (IPFSClient, error)
```

---

## 2. Rust Crypto & SDK Objects

### 2.1 Core Types (`sdk/rust/src/types/`)

```rust
// types/mod.rs
pub mod address;
pub mod proof;
pub mod signal;
pub mod genome;

// types/address.rs
use serde::{Deserialize, Serialize};
use std::fmt;
use std::str::FromStr;

/// Ethereum address (20 bytes)
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct Address([u8; 20]);

impl Address {
    /// Create address from bytes
    pub fn from_bytes(bytes: [u8; 20]) -> Self {
        Self(bytes)
    }
    
    /// Create address from hex string
    pub fn from_hex(hex: &str) -> Result<Self, AddressError> {
        let hex = hex.strip_prefix("0x").unwrap_or(hex);
        if hex.len() != 40 {
            return Err(AddressError::InvalidLength(hex.len()));
        }
        let bytes = hex::decode(hex).map_err(|_| AddressError::InvalidHex)?;
        let mut arr = [0u8; 20];
        arr.copy_from_slice(&bytes);
        Ok(Self(arr))
    }
    
    /// Get bytes
    pub fn as_bytes(&self) -> &[u8; 20] {
        &self.0
    }
    
    /// Convert to checksummed hex string
    pub fn to_checksum(&self) -> String {
        // EIP-55 checksum implementation
        let hex = hex::encode(self.0);
        let hash = keccak256(hex.as_bytes());
        let mut result = String::with_capacity(42);
        result.push_str("0x");
        for (i, c) in hex.chars().enumerate() {
            if c.is_ascii_digit() {
                result.push(c);
            } else {
                let nibble = hash[i / 2] >> (if i % 2 == 0 { 4 } else { 0 }) & 0xf;
                if nibble >= 8 {
                    result.push(c.to_ascii_uppercase());
                } else {
                    result.push(c);
                }
            }
        }
        result
    }
    
    /// Zero address
    pub const ZERO: Self = Self([0u8; 20]);
}

impl fmt::Display for Address {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.to_checksum())
    }
}

impl FromStr for Address {
    type Err = AddressError;
    
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        Self::from_hex(s)
    }
}

#[derive(Debug, thiserror::Error)]
pub enum AddressError {
    #[error("Invalid address length: {0}")]
    InvalidLength(usize),
    #[error("Invalid hex encoding")]
    InvalidHex,
}

// types/proof.rs
use serde::{Deserialize, Serialize};
use std::time::{Duration, SystemTime};

/// Human behavior ZK proof
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HumanProof {
    /// Raw proof bytes
    pub proof: Vec<u8>,
    /// Public inputs for verification
    pub public_inputs: PublicInputs,
    /// Circuit identifier
    pub circuit_id: String,
    /// Prover identifier
    pub prover_id: String,
    /// Timestamp when proof was generated
    pub timestamp: u64,
}

impl HumanProof {
    /// Create new proof
    pub fn new(
        proof: Vec<u8>,
        public_inputs: PublicInputs,
        circuit_id: String,
        prover_id: String,
    ) -> Self {
        Self {
            proof,
            public_inputs,
            circuit_id,
            prover_id,
            timestamp: SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap()
                .as_secs(),
        }
    }
    
    /// Compute proof hash (keccak256)
    pub fn hash(&self) -> [u8; 32] {
        use sha3::{Digest, Keccak256};
        let mut hasher = Keccak256::new();
        hasher.update(&self.proof);
        hasher.update(&self.public_inputs.to_bytes());
        hasher.finalize().into()
    }
    
    /// Check if proof is expired
    pub fn is_expired(&self, ttl: Duration) -> bool {
        let now = SystemTime::now()
            .duration_since(SystemTime::UNIX_EPOCH)
            .unwrap()
            .as_secs();
        now - self.timestamp > ttl.as_secs()
    }
    
    /// Serialize to bytes
    pub fn to_bytes(&self) -> Vec<u8> {
        bincode::serialize(self).expect("Failed to serialize proof")
    }
    
    /// Deserialize from bytes
    pub fn from_bytes(bytes: &[u8]) -> Result<Self, bincode::Error> {
        bincode::deserialize(bytes)
    }
}

/// Public inputs for ZK verification
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PublicInputs {
    /// User's wallet address
    pub wallet_address: Address,
    /// Behavioral features commitment
    pub features_commitment: [u8; 32],
    /// Timestamp
    pub timestamp: u64,
    /// Challenge response
    pub challenge_response: [u8; 32],
}

impl PublicInputs {
    /// Convert to bytes for hashing
    pub fn to_bytes(&self) -> Vec<u8> {
        let mut bytes = Vec::with_capacity(104);
        bytes.extend_from_slice(self.wallet_address.as_bytes());
        bytes.extend_from_slice(&self.features_commitment);
        bytes.extend_from_slice(&self.timestamp.to_le_bytes());
        bytes.extend_from_slice(&self.challenge_response);
        bytes
    }
}

/// Proof verification result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VerificationResult {
    /// Whether proof is valid
    pub valid: bool,
    /// Proof hash
    pub proof_hash: [u8; 32],
    /// Risk score (0.0-1.0)
    pub risk_score: f64,
    /// Transaction hash if submitted on-chain
    pub tx_hash: Option<String>,
    /// Error message if verification failed
    pub error: Option<String>,
}

// types/signal.rs
use serde::{Deserialize, Serialize};

/// Threat signal types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum SignalType {
    ExploitDetected,
    KeyLeaked,
    AnomalyDetected,
    PhishingAttempt,
    MaliciousContract,
    FlashLoanAttack,
    Reentrancy,
    PriceManipulation,
}

impl SignalType {
    /// Get default risk weight
    pub fn default_weight(&self) -> u8 {
        match self {
            Self::ExploitDetected => 90,
            Self::KeyLeaked => 95,
            Self::AnomalyDetected => 60,
            Self::PhishingAttempt => 70,
            Self::MaliciousContract => 85,
            Self::FlashLoanAttack => 80,
            Self::Reentrancy => 85,
            Self::PriceManipulation => 75,
        }
    }
}

/// Threat signal
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ThreatSignal {
    /// Entity address (wallet or contract)
    pub entity_address: Address,
    /// Signal type
    pub signal_type: SignalType,
    /// Risk score (0-100)
    pub risk_score: u8,
    /// Confidence (0.0-1.0)
    pub confidence: f64,
    /// Source of the signal
    pub source: String,
    /// Additional metadata
    pub metadata: Option<serde_json::Value>,
    /// Timestamp
    pub timestamp: u64,
}

impl ThreatSignal {
    /// Check if high risk
    pub fn is_high_risk(&self) -> bool {
        self.risk_score >= 70
    }
    
    /// Check if critical
    pub fn is_critical(&self) -> bool {
        self.risk_score >= 90
    }
    
    /// Compute weighted score
    pub fn weighted_score(&self) -> f64 {
        self.risk_score as f64 * self.confidence
    }
}

// types/genome.rs
use serde::{Deserialize, Serialize};

/// Genome label
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum GenomeLabel {
    KnownExploit,
    Suspicious,
    Benign,
    Unknown,
}

/// Contract genome fingerprint
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Genome {
    /// Genome hash (keccak256 of features)
    pub hash: [u8; 32],
    /// IPFS hash of full analysis
    pub ipfs_hash: String,
    /// Contract address (if applicable)
    pub contract_address: Option<Address>,
    /// Classification label
    pub label: GenomeLabel,
    /// Extracted features
    pub features: GenomeFeatures,
}

/// Genome features for analysis
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GenomeFeatures {
    /// Opcode histogram
    pub opcode_histogram: std::collections::HashMap<String, u32>,
    /// Bytecode length
    pub bytecode_length: usize,
    /// Unique opcodes count
    pub unique_opcodes: usize,
    /// Jump density
    pub jump_density: f64,
    /// External call count
    pub external_call_count: u32,
    /// Storage access count
    pub storage_access_count: u32,
    /// SELFDESTRUCT present
    pub has_selfdestruct: bool,
    /// DELEGATECALL present
    pub has_delegatecall: bool,
    /// CREATE2 present
    pub has_create2: bool,
}

impl GenomeFeatures {
    /// Convert to feature vector for ML
    pub fn to_vector(&self) -> Vec<f32> {
        vec![
            self.bytecode_length as f32,
            self.unique_opcodes as f32,
            self.jump_density as f32,
            self.external_call_count as f32,
            self.storage_access_count as f32,
            self.has_selfdestruct as u8 as f32,
            self.has_delegatecall as u8 as f32,
            self.has_create2 as u8 as f32,
        ]
    }
}
```

---

### 2.2 ZK Prover (`sdk/rust/src/zk/`)

```rust
// zk/mod.rs
pub mod circuit;
pub mod prover;
pub mod verifier;

// zk/circuit.rs
use ark_ff::Field;
use noir_std::hash::pedersen_hash;

/// Human behavior circuit inputs
#[derive(Debug, Clone)]
pub struct HumanCircuitInputs {
    /// Transaction count
    pub tx_count: u32,
    /// Average transaction interval (seconds)
    pub avg_tx_interval: f64,
    /// Transaction interval variance
    pub tx_interval_variance: f64,
    /// Gas variance
    pub gas_variance: f64,
    /// Interaction diversity score
    pub interaction_diversity: u32,
    /// Unique contracts interacted with
    pub unique_contracts: u32,
    /// Challenge from server
    pub challenge: [u8; 32],
    /// User's private key (for signing challenge)
    pub private_key: [u8; 32],
}

impl HumanCircuitInputs {
    /// Compute features commitment
    pub fn compute_commitment(&self) -> [u8; 32] {
        let mut preimage = Vec::new();
        preimage.extend_from_slice(&self.tx_count.to_le_bytes());
        preimage.extend_from_slice(&self.avg_tx_interval.to_le_bytes());
        preimage.extend_from_slice(&self.tx_interval_variance.to_le_bytes());
        preimage.extend_from_slice(&self.gas_variance.to_le_bytes());
        preimage.extend_from_slice(&self.interaction_diversity.to_le_bytes());
        preimage.extend_from_slice(&self.unique_contracts.to_le_bytes());
        pedersen_hash(&preimage)
    }
    
    /// Check if features indicate human behavior
    pub fn is_likely_human(&self) -> bool {
        // Heuristic checks
        self.tx_count >= 10 
            && self.avg_tx_interval > 60.0  // At least 1 minute between txs
            && self.tx_interval_variance > 100.0  // Variable timing
            && self.interaction_diversity >= 3  // Multiple interaction types
            && self.unique_contracts >= 5  // Multiple contracts
    }
}

/// Exploit proof circuit inputs
#[derive(Debug, Clone)]
pub struct ExploitCircuitInputs {
    /// Target contract bytecode
    pub target_bytecode: Vec<u8>,
    /// Exploit calldata
    pub exploit_calldata: Vec<u8>,
    /// Pre-state root
    pub pre_state_root: [u8; 32],
    /// Post-state root
    pub post_state_root: [u8; 32],
    /// Expected vulnerability type
    pub vulnerability_type: VulnerabilityType,
}

/// Vulnerability types
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum VulnerabilityType {
    Reentrancy,
    IntegerOverflow,
    AccessControl,
    FlashLoan,
    PriceManipulation,
    StorageCollision,
    Other,
}

// zk/prover.rs
use std::path::Path;

/// ZK Prover for generating proofs
pub struct Prover {
    /// Circuit ACIR
    circuit: CompiledCircuit,
    /// Proving key
    proving_key: ProvingKey,
    /// Backend (Barretenberg)
    backend: BarretenbergBackend,
}

impl Prover {
    /// Create new prover from circuit file
    pub fn new(circuit_path: &Path) -> Result<Self, ProverError> {
        let circuit = CompiledCircuit::load(circuit_path)?;
        let backend = BarretenbergBackend::new()?;
        let proving_key = backend.generate_proving_key(&circuit)?;
        
        Ok(Self {
            circuit,
            proving_key,
            backend,
        })
    }
    
    /// Generate human behavior proof
    pub fn prove_human(&self, inputs: &HumanCircuitInputs) -> Result<HumanProof, ProverError> {
        // Convert inputs to witness
        let witness = self.generate_witness(inputs)?;
        
        // Generate proof
        let proof_bytes = self.backend.prove(&self.proving_key, &witness)?;
        
        // Build public inputs
        let public_inputs = PublicInputs {
            wallet_address: derive_address(&inputs.private_key),
            features_commitment: inputs.compute_commitment(),
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs(),
            challenge_response: sign_challenge(&inputs.private_key, &inputs.challenge),
        };
        
        Ok(HumanProof::new(
            proof_bytes,
            public_inputs,
            self.circuit.id.clone(),
            "barretenberg".to_string(),
        ))
    }
    
    /// Generate exploit proof
    pub fn prove_exploit(&self, inputs: &ExploitCircuitInputs) -> Result<Vec<u8>, ProverError> {
        let witness = self.generate_exploit_witness(inputs)?;
        self.backend.prove(&self.proving_key, &witness)
    }
    
    /// Generate witness from circuit inputs
    fn generate_witness(&self, inputs: &HumanCircuitInputs) -> Result<Witness, ProverError> {
        // Implementation depends on Noir circuit structure
        todo!()
    }
    
    fn generate_exploit_witness(&self, inputs: &ExploitCircuitInputs) -> Result<Witness, ProverError> {
        todo!()
    }
}

#[derive(Debug, thiserror::Error)]
pub enum ProverError {
    #[error("Failed to load circuit: {0}")]
    CircuitLoad(String),
    #[error("Failed to generate witness: {0}")]
    WitnessGeneration(String),
    #[error("Failed to generate proof: {0}")]
    ProofGeneration(String),
    #[error("Backend error: {0}")]
    Backend(String),
}

// zk/verifier.rs
/// ZK Verifier for verifying proofs
pub struct Verifier {
    /// Verification key
    verification_key: VerificationKey,
    /// Backend
    backend: BarretenbergBackend,
}

impl Verifier {
    /// Create new verifier from verification key
    pub fn new(vk_path: &Path) -> Result<Self, VerifierError> {
        let verification_key = VerificationKey::load(vk_path)?;
        let backend = BarretenbergBackend::new()?;
        
        Ok(Self {
            verification_key,
            backend,
        })
    }
    
    /// Verify a proof
    pub fn verify(&self, proof: &HumanProof) -> Result<bool, VerifierError> {
        let public_inputs = proof.public_inputs.to_bytes();
        self.backend.verify(&self.verification_key, &proof.proof, &public_inputs)
            .map_err(VerifierError::Backend)
    }
    
    /// Batch verify multiple proofs
    pub fn batch_verify(&self, proofs: &[HumanProof]) -> Result<Vec<bool>, VerifierError> {
        proofs.iter().map(|p| self.verify(p)).collect()
    }
}

#[derive(Debug, thiserror::Error)]
pub enum VerifierError {
    #[error("Failed to load verification key: {0}")]
    KeyLoad(String),
    #[error("Backend error: {0}")]
    Backend(String),
}
```

---

### 2.3 SDK Client (`sdk/rust/src/client/`)

```rust
// client/mod.rs
pub mod firewall;
pub mod oracle;
pub mod genome;

// client/firewall.rs
use crate::types::{Address, HumanProof, VerificationResult};
use reqwest::Client;
use std::time::Duration;

/// Identity Firewall client
pub struct FirewallClient {
    /// HTTP client
    client: Client,
    /// Base URL
    base_url: String,
    /// API key
    api_key: Option<String>,
}

impl FirewallClient {
    /// Create new client
    pub fn new(base_url: &str) -> Self {
        Self {
            client: Client::builder()
                .timeout(Duration::from_secs(30))
                .build()
                .unwrap(),
            base_url: base_url.to_string(),
            api_key: None,
        }
    }
    
    /// Set API key
    pub fn with_api_key(mut self, api_key: String) -> Self {
        self.api_key = Some(api_key);
        self
    }
    
    /// Verify a human behavior proof
    pub async fn verify_proof(&self, proof: &HumanProof) -> Result<VerificationResult, ClientError> {
        let url = format!("{}/firewall/verify-proof", self.base_url);
        
        let mut req = self.client.post(&url)
            .json(&VerifyProofRequest {
                proof: proof.proof.clone(),
                public_inputs: proof.public_inputs.clone(),
            });
        
        if let Some(ref key) = self.api_key {
            req = req.header("X-API-Key", key);
        }
        
        let resp = req.send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<VerificationResult> = resp.json().await?;
        Ok(result.data)
    }
    
    /// Get challenge for proof generation
    pub async fn get_challenge(&self) -> Result<Challenge, ClientError> {
        let url = format!("{}/firewall/challenge", self.base_url);
        
        let resp = self.client.get(&url).send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<Challenge> = resp.json().await?;
        Ok(result.data)
    }
    
    /// Get risk score for an address
    pub async fn get_risk_score(&self, address: &Address) -> Result<RiskScore, ClientError> {
        let url = format!("{}/firewall/risk/{}", self.base_url, address);
        
        let mut req = self.client.get(&url);
        
        if let Some(ref key) = self.api_key {
            req = req.header("X-API-Key", key);
        }
        
        let resp = req.send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<RiskScore> = resp.json().await?;
        Ok(result.data)
    }
}

/// Challenge response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Challenge {
    pub challenge_id: String,
    pub challenge: String,
    pub expires_at: u64,
}

/// Risk score response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RiskScore {
    pub address: String,
    pub risk_score: f64,
    pub risk_level: String,
    pub signals: Vec<SignalSummary>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SignalSummary {
    pub signal_type: String,
    pub source: String,
    pub confidence: f64,
    pub timestamp: u64,
}

// client/oracle.rs
/// Threat Oracle client
pub struct OracleClient {
    client: Client,
    base_url: String,
    api_key: Option<String>,
}

impl OracleClient {
    pub fn new(base_url: &str) -> Self {
        Self {
            client: Client::builder()
                .timeout(Duration::from_secs(30))
                .build()
                .unwrap(),
            base_url: base_url.to_string(),
            api_key: None,
        }
    }
    
    pub fn with_api_key(mut self, api_key: String) -> Self {
        self.api_key = Some(api_key);
        self
    }
    
    /// Get all signals for an address
    pub async fn get_signals(&self, address: &Address) -> Result<SignalsResponse, ClientError> {
        let url = format!("{}/oracle/signals/{}", self.base_url, address);
        
        let mut req = self.client.get(&url);
        if let Some(ref key) = self.api_key {
            req = req.header("X-API-Key", key);
        }
        
        let resp = req.send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<SignalsResponse> = resp.json().await?;
        Ok(result.data)
    }
    
    /// Subscribe to signals via webhook
    pub async fn subscribe(
        &self,
        webhook_url: &str,
        addresses: &[Address],
    ) -> Result<Subscription, ClientError> {
        let url = format!("{}/oracle/subscribe", self.base_url);
        
        let mut req = self.client.post(&url)
            .json(&SubscribeRequest {
                webhook_url: webhook_url.to_string(),
                addresses: addresses.iter().map(|a| a.to_string()).collect(),
                signal_types: None,
            });
        
        if let Some(ref key) = self.api_key {
            req = req.header("X-API-Key", key);
        }
        
        let resp = req.send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<Subscription> = resp.json().await?;
        Ok(result.data)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SignalsResponse {
    pub address: String,
    pub signals: Vec<ThreatSignal>,
    pub on_chain_risk_score: u8,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Subscription {
    pub subscription_id: String,
    pub active: bool,
}

// client/genome.rs
/// Genome Analyzer client
pub struct GenomeClient {
    client: Client,
    base_url: String,
    api_key: Option<String>,
}

impl GenomeClient {
    pub fn new(base_url: &str) -> Self {
        Self {
            client: Client::builder()
                .timeout(Duration::from_secs(60))
                .build()
                .unwrap(),
            base_url: base_url.to_string(),
            api_key: None,
        }
    }
    
    /// Analyze a contract
    pub async fn analyze(&self, contract_address: &Address) -> Result<AnalysisJob, ClientError> {
        let url = format!("{}/genome/analyze", self.base_url);
        
        let mut req = self.client.post(&url)
            .json(&AnalyzeRequest {
                contract_address: contract_address.to_string(),
                priority: "normal".to_string(),
            });
        
        if let Some(ref key) = self.api_key {
            req = req.header("X-API-Key", key);
        }
        
        let resp = req.send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<AnalysisJob> = resp.json().await?;
        Ok(result.data)
    }
    
    /// Get analysis status
    pub async fn get_status(&self, analysis_id: &str) -> Result<AnalysisStatus, ClientError> {
        let url = format!("{}/genome/status/{}", self.base_url, analysis_id);
        
        let resp = self.client.get(&url).send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<AnalysisStatus> = resp.json().await?;
        Ok(result.data)
    }
    
    /// Get genome by hash
    pub async fn get_genome(&self, genome_hash: &str) -> Result<Genome, ClientError> {
        let url = format!("{}/genome/{}", self.base_url, genome_hash);
        
        let resp = self.client.get(&url).send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<Genome> = resp.json().await?;
        Ok(result.data)
    }
    
    /// Find similar genomes
    pub async fn find_similar(
        &self,
        genome_hash: &str,
        threshold: f64,
    ) -> Result<Vec<GenomeSimilarity>, ClientError> {
        let url = format!(
            "{}/genome/{}/similar?threshold={}",
            self.base_url, genome_hash, threshold
        );
        
        let resp = self.client.get(&url).send().await?;
        
        if !resp.status().is_success() {
            let error: ApiError = resp.json().await?;
            return Err(ClientError::Api(error));
        }
        
        let result: ApiResponse<Vec<GenomeSimilarity>> = resp.json().await?;
        Ok(result.data)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnalysisJob {
    pub analysis_id: String,
    pub status: String,
    pub estimated_completion: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnalysisStatus {
    pub analysis_id: String,
    pub status: String,
    pub genome_hash: Option<String>,
    pub ipfs_hash: Option<String>,
    pub label: Option<GenomeLabel>,
    pub error: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GenomeSimilarity {
    pub genome: Genome,
    pub similarity_score: f64,
}

// Common types
#[derive(Debug, Clone, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: T,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ApiError {
    code: String,
    message: String,
}

#[derive(Debug, thiserror::Error)]
pub enum ClientError {
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),
    #[error("API error: {0:?}")]
    Api(ApiError),
    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),
}
```

---

### 2.4 WASM Bindings (`sdk/rust/src/wasm/`)

```rust
// wasm/mod.rs
use wasm_bindgen::prelude::*;
use js_sys::{Promise, Uint8Array};

/// Initialize the WASM module
#[wasm_bindgen(start)]
pub fn init() {
    console_error_panic_hook::set_once();
}

/// WASM-exported Prover
#[wasm_bindgen]
pub struct WasmProver {
    inner: crate::zk::prover::Prover,
}

#[wasm_bindgen]
impl WasmProver {
    /// Create new prover from circuit bytes
    #[wasm_bindgen(constructor)]
    pub fn new(circuit_bytes: &[u8]) -> Result<WasmProver, JsValue> {
        let inner = crate::zk::prover::Prover::from_bytes(circuit_bytes)
            .map_err(|e| JsValue::from_str(&e.to_string()))?;
        Ok(Self { inner })
    }
    
    /// Generate human behavior proof
    #[wasm_bindgen(js_name = proveHuman)]
    pub fn prove_human(&self, inputs_json: &str) -> Result<Uint8Array, JsValue> {
        let inputs: crate::zk::circuit::HumanCircuitInputs = 
            serde_json::from_str(inputs_json)
                .map_err(|e| JsValue::from_str(&e.to_string()))?;
        
        let proof = self.inner.prove_human(&inputs)
            .map_err(|e| JsValue::from_str(&e.to_string()))?;
        
        let bytes = proof.to_bytes();
        Ok(Uint8Array::from(&bytes[..]))
    }
}

/// WASM-exported Verifier
#[wasm_bindgen]
pub struct WasmVerifier {
    inner: crate::zk::verifier::Verifier,
}

#[wasm_bindgen]
impl WasmVerifier {
    /// Create new verifier from verification key bytes
    #[wasm_bindgen(constructor)]
    pub fn new(vk_bytes: &[u8]) -> Result<WasmVerifier, JsValue> {
        let inner = crate::zk::verifier::Verifier::from_bytes(vk_bytes)
            .map_err(|e| JsValue::from_str(&e.to_string()))?;
        Ok(Self { inner })
    }
    
    /// Verify a proof
    #[wasm_bindgen]
    pub fn verify(&self, proof_bytes: &[u8]) -> Result<bool, JsValue> {
        let proof = crate::types::HumanProof::from_bytes(proof_bytes)
            .map_err(|e| JsValue::from_str(&e.to_string()))?;
        
        self.inner.verify(&proof)
            .map_err(|e| JsValue::from_str(&e.to_string()))
    }
}

/// WASM-exported client
#[wasm_bindgen]
pub struct WasmClient {
    base_url: String,
    api_key: Option<String>,
}

#[wasm_bindgen]
impl WasmClient {
    #[wasm_bindgen(constructor)]
    pub fn new(base_url: &str, api_key: Option<String>) -> Self {
        Self {
            base_url: base_url.to_string(),
            api_key,
        }
    }
    
    /// Verify proof via API
    #[wasm_bindgen(js_name = verifyProof)]
    pub fn verify_proof(&self, proof_bytes: &[u8]) -> Promise {
        let proof = match crate::types::HumanProof::from_bytes(proof_bytes) {
            Ok(p) => p,
            Err(e) => return Promise::reject(&JsValue::from_str(&e.to_string())),
        };
        
        let base_url = self.base_url.clone();
        let api_key = self.api_key.clone();
        
        wasm_bindgen_futures::future_to_promise(async move {
            let client = crate::client::firewall::FirewallClient::new(&base_url);
            let client = match api_key {
                Some(key) => client.with_api_key(key),
                None => client,
            };
            
            let result = client.verify_proof(&proof).await
                .map_err(|e| JsValue::from_str(&e.to_string()))?;
            
            serde_wasm_bindgen::to_value(&result)
                .map_err(|e| JsValue::from_str(&e.to_string()))
        })
    }
    
    /// Get risk score for address
    #[wasm_bindgen(js_name = getRiskScore)]
    pub fn get_risk_score(&self, address: &str) -> Promise {
        let addr = match crate::types::Address::from_hex(address) {
            Ok(a) => a,
            Err(e) => return Promise::reject(&JsValue::from_str(&e.to_string())),
        };
        
        let base_url = self.base_url.clone();
        let api_key = self.api_key.clone();
        
        wasm_bindgen_futures::future_to_promise(async move {
            let client = crate::client::firewall::FirewallClient::new(&base_url);
            let client = match api_key {
                Some(key) => client.with_api_key(key),
                None => client,
            };
            
            let result = client.get_risk_score(&addr).await
                .map_err(|e| JsValue::from_str(&e.to_string()))?;
            
            serde_wasm_bindgen::to_value(&result)
                .map_err(|e| JsValue::from_str(&e.to_string()))
        })
    }
}

/// Utility: Hash data with keccak256
#[wasm_bindgen(js_name = keccak256)]
pub fn wasm_keccak256(data: &[u8]) -> Uint8Array {
    use sha3::{Digest, Keccak256};
    let hash = Keccak256::digest(data);
    Uint8Array::from(&hash[..])
}

/// Utility: Validate Ethereum address
#[wasm_bindgen(js_name = isValidAddress)]
pub fn wasm_is_valid_address(address: &str) -> bool {
    crate::types::Address::from_hex(address).is_ok()
}
```

---

## 3. Solidity Smart Contract Objects

### 3.1 Identity Firewall (`contracts/src/IdentityFirewall.sol`)

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

/// @title IIdentityFirewall
/// @notice Interface for the Identity Firewall contract
interface IIdentityFirewall {
    /// @notice Emitted when a proof is verified
    event ProofVerified(
        address indexed user,
        bytes32 indexed proofHash,
        uint256 timestamp
    );
    
    /// @notice Emitted when a proof expires
    event ProofExpired(
        address indexed user,
        bytes32 indexed proofHash
    );
    
    /// @notice Emitted when a verifier is added
    event VerifierAdded(address indexed verifier);
    
    /// @notice Emitted when a verifier is removed
    event VerifierRemoved(address indexed verifier);
    
    /// @notice Verify a human behavior proof
    /// @param proof The ZK proof bytes
    /// @return success Whether verification succeeded
    function verifyProof(bytes calldata proof) external returns (bool success);
    
    /// @notice Check if address has verified proof
    /// @param proofHash Hash of the proof
    /// @return exists Whether proof exists and is valid
    function hasVerifiedProof(bytes32 proofHash) external view returns (bool exists);
    
    /// @notice Check if user is verified
    /// @param user User address
    /// @return verified Whether user has valid verification
    function isVerified(address user) external view returns (bool verified);
    
    /// @notice Get verification expiry for user
    /// @param user User address
    /// @return expiresAt Timestamp when verification expires (0 if not verified)
    function getVerificationExpiry(address user) external view returns (uint256 expiresAt);
}

/// @title IdentityFirewall
/// @notice Main contract for human behavior proof verification
contract IdentityFirewall is 
    IIdentityFirewall,
    UUPSUpgradeable,
    AccessControlUpgradeable,
    PausableUpgradeable,
    ReentrancyGuardUpgradeable 
{
    // ============ Constants ============
    
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant VERIFIER_ROLE = keccak256("VERIFIER_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    
    uint256 public constant PROOF_VALIDITY_PERIOD = 7 days;
    uint256 public constant MIN_PROOF_SIZE = 32;
    uint256 public constant MAX_PROOF_SIZE = 4096;
    
    // ============ State Variables ============
    
    /// @notice Mapping from proof hash to verification data
    mapping(bytes32 => ProofData) public proofs;
    
    /// @notice Mapping from user to their latest proof hash
    mapping(address => bytes32) public userProofs;
    
    /// @notice Mapping from user to verification expiry timestamp
    mapping(address => uint256) public verificationExpiry;
    
    /// @notice Total proofs verified
    uint256 public totalProofsVerified;
    
    /// @notice ZK verifier contract address
    address public zkVerifier;
    
    // ============ Structs ============
    
    /// @notice Proof verification data
    struct ProofData {
        address user;
        uint256 timestamp;
        uint256 expiresAt;
        bool valid;
    }
    
    // ============ Errors ============
    
    error InvalidProofSize(uint256 size);
    error ProofAlreadyExists(bytes32 proofHash);
    error ProofVerificationFailed();
    error ZeroAddress();
    error ProofExpiredError(bytes32 proofHash);
    
    // ============ Initialization ============
    
    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }
    
    /// @notice Initialize the contract
    /// @param admin Admin address
    /// @param _zkVerifier ZK verifier contract address
    function initialize(
        address admin,
        address _zkVerifier
    ) external initializer {
        if (admin == address(0) || _zkVerifier == address(0)) {
            revert ZeroAddress();
        }
        
        __UUPSUpgradeable_init();
        __AccessControl_init();
        __Pausable_init();
        __ReentrancyGuard_init();
        
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        
        zkVerifier = _zkVerifier;
    }
    
    // ============ External Functions ============
    
    /// @inheritdoc IIdentityFirewall
    function verifyProof(
        bytes calldata proof
    ) external override nonReentrant whenNotPaused returns (bool) {
        // Validate proof size
        if (proof.length < MIN_PROOF_SIZE || proof.length > MAX_PROOF_SIZE) {
            revert InvalidProofSize(proof.length);
        }
        
        // Compute proof hash
        bytes32 proofHash = keccak256(proof);
        
        // Check if proof already exists
        if (proofs[proofHash].valid) {
            revert ProofAlreadyExists(proofHash);
        }
        
        // Call ZK verifier
        (bool success, bytes memory result) = zkVerifier.staticcall(
            abi.encodeWithSignature("verify(bytes)", proof)
        );
        
        if (!success || result.length == 0 || !abi.decode(result, (bool))) {
            revert ProofVerificationFailed();
        }
        
        // Store proof data
        uint256 expiresAt = block.timestamp + PROOF_VALIDITY_PERIOD;
        proofs[proofHash] = ProofData({
            user: msg.sender,
            timestamp: block.timestamp,
            expiresAt: expiresAt,
            valid: true
        });
        
        // Update user mappings
        userProofs[msg.sender] = proofHash;
        verificationExpiry[msg.sender] = expiresAt;
        
        // Increment counter
        totalProofsVerified++;
        
        emit ProofVerified(msg.sender, proofHash, block.timestamp);
        
        return true;
    }
    
    /// @inheritdoc IIdentityFirewall
    function hasVerifiedProof(
        bytes32 proofHash
    ) external view override returns (bool) {
        ProofData storage data = proofs[proofHash];
        return data.valid && data.expiresAt > block.timestamp;
    }
    
    /// @inheritdoc IIdentityFirewall
    function isVerified(address user) external view override returns (bool) {
        return verificationExpiry[user] > block.timestamp;
    }
    
    /// @inheritdoc IIdentityFirewall
    function getVerificationExpiry(
        address user
    ) external view override returns (uint256) {
        uint256 expiry = verificationExpiry[user];
        return expiry > block.timestamp ? expiry : 0;
    }
    
    // ============ Admin Functions ============
    
    /// @notice Set ZK verifier address
    /// @param _zkVerifier New verifier address
    function setZKVerifier(address _zkVerifier) external onlyRole(ADMIN_ROLE) {
        if (_zkVerifier == address(0)) revert ZeroAddress();
        zkVerifier = _zkVerifier;
    }
    
    /// @notice Pause contract
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }
    
    /// @notice Unpause contract
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }
    
    // ============ Internal Functions ============
    
    /// @notice Authorize upgrade (only UPGRADER_ROLE)
    function _authorizeUpgrade(
        address newImplementation
    ) internal override onlyRole(UPGRADER_ROLE) {}
}
```

---

### 3.2 Threat Oracle (`contracts/src/ThreatOracle.sol`)

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";

/// @title IThreatOracle
/// @notice Interface for the Threat Oracle contract
interface IThreatOracle {
    /// @notice Emitted when risk score is updated
    event RiskUpdated(
        address indexed target,
        uint256 oldScore,
        uint256 newScore,
        address indexed updater
    );
    
    /// @notice Emitted when threshold is changed
    event ThresholdUpdated(uint256 oldThreshold, uint256 newThreshold);
    
    /// @notice Update risk score for an address
    /// @param target Target address
    /// @param score New risk score (0-100)
    function updateRiskScore(address target, uint256 score) external;
    
    /// @notice Batch update risk scores
    /// @param targets Array of target addresses
    /// @param scores Array of risk scores
    function batchUpdateRiskScores(
        address[] calldata targets,
        uint256[] calldata scores
    ) external;
    
    /// @notice Get risk score for an address
    /// @param target Target address
    /// @return score Risk score (0-100)
    function getRiskScore(address target) external view returns (uint256 score);
    
    /// @notice Check if address is high risk
    /// @param target Target address
    /// @return isHighRisk Whether address is high risk
    function isHighRisk(address target) external view returns (bool isHighRisk);
}

/// @title ThreatOracle
/// @notice On-chain threat intelligence oracle
contract ThreatOracle is 
    IThreatOracle,
    UUPSUpgradeable,
    AccessControlUpgradeable,
    PausableUpgradeable 
{
    // ============ Constants ============
    
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    
    uint256 public constant MAX_RISK_SCORE = 100;
    uint256 public constant MAX_BATCH_SIZE = 100;
    
    // ============ State Variables ============
    
    /// @notice Mapping from address to risk score
    mapping(address => uint256) public riskScores;
    
    /// @notice Mapping from address to last update timestamp
    mapping(address => uint256) public lastUpdated;
    
    /// @notice High risk threshold (default 70)
    uint256 public highRiskThreshold;
    
    /// @notice Total addresses tracked
    uint256 public totalTracked;
    
    // ============ Structs ============
    
    /// @notice Risk data for an address
    struct RiskData {
        uint256 score;
        uint256 lastUpdated;
        uint256 updateCount;
    }
    
    /// @notice Extended risk mapping
    mapping(address => RiskData) public riskData;
    
    // ============ Errors ============
    
    error InvalidScore(uint256 score);
    error ArrayLengthMismatch();
    error BatchTooLarge(uint256 size);
    error ZeroAddress();
    
    // ============ Initialization ============
    
    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }
    
    /// @notice Initialize the contract
    /// @param admin Admin address
    /// @param oracle Initial oracle address
    function initialize(
        address admin,
        address oracle
    ) external initializer {
        if (admin == address(0) || oracle == address(0)) {
            revert ZeroAddress();
        }
        
        __UUPSUpgradeable_init();
        __AccessControl_init();
        __Pausable_init();
        
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(ADMIN_ROLE, admin);
        _grantRole(ORACLE_ROLE, oracle);
        _grantRole(UPGRADER_ROLE, admin);
        
        highRiskThreshold = 70;
    }
    
    // ============ External Functions ============
    
    /// @inheritdoc IThreatOracle
    function updateRiskScore(
        address target,
        uint256 score
    ) external override onlyRole(ORACLE_ROLE) whenNotPaused {
        _updateRiskScore(target, score);
    }
    
    /// @inheritdoc IThreatOracle
    function batchUpdateRiskScores(
        address[] calldata targets,
        uint256[] calldata scores
    ) external override onlyRole(ORACLE_ROLE) whenNotPaused {
        if (targets.length != scores.length) {
            revert ArrayLengthMismatch();
        }
        if (targets.length > MAX_BATCH_SIZE) {
            revert BatchTooLarge(targets.length);
        }
        
        for (uint256 i = 0; i < targets.length; i++) {
            _updateRiskScore(targets[i], scores[i]);
        }
    }
    
    /// @inheritdoc IThreatOracle
    function getRiskScore(
        address target
    ) external view override returns (uint256) {
        return riskScores[target];
    }
    
    /// @inheritdoc IThreatOracle
    function isHighRisk(
        address target
    ) external view override returns (bool) {
        return riskScores[target] >= highRiskThreshold;
    }
    
    /// @notice Get full risk data for address
    /// @param target Target address
    /// @return data Risk data struct
    function getRiskData(
        address target
    ) external view returns (RiskData memory data) {
        return riskData[target];
    }
    
    // ============ Admin Functions ============
    
    /// @notice Set high risk threshold
    /// @param threshold New threshold (0-100)
    function setHighRiskThreshold(
        uint256 threshold
    ) external onlyRole(ADMIN_ROLE) {
        if (threshold > MAX_RISK_SCORE) {
            revert InvalidScore(threshold);
        }
        
        uint256 oldThreshold = highRiskThreshold;
        highRiskThreshold = threshold;
        
        emit ThresholdUpdated(oldThreshold, threshold);
    }
    
    /// @notice Add oracle
    /// @param oracle Oracle address to add
    function addOracle(address oracle) external onlyRole(ADMIN_ROLE) {
        if (oracle == address(0)) revert ZeroAddress();
        _grantRole(ORACLE_ROLE, oracle);
    }
    
    /// @notice Remove oracle
    /// @param oracle Oracle address to remove
    function removeOracle(address oracle) external onlyRole(ADMIN_ROLE) {
        _revokeRole(ORACLE_ROLE, oracle);
    }
    
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }
    
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }
    
    // ============ Internal Functions ============
    
    /// @notice Internal risk score update
    function _updateRiskScore(address target, uint256 score) internal {
        if (target == address(0)) revert ZeroAddress();
        if (score > MAX_RISK_SCORE) revert InvalidScore(score);
        
        uint256 oldScore = riskScores[target];
        
        // Check if new address
        if (lastUpdated[target] == 0) {
            totalTracked++;
        }
        
        // Update mappings
        riskScores[target] = score;
        lastUpdated[target] = block.timestamp;
        
        // Update extended data
        riskData[target].score = score;
        riskData[target].lastUpdated = block.timestamp;
        riskData[target].updateCount++;
        
        emit RiskUpdated(target, oldScore, score, msg.sender);
    }
    
    function _authorizeUpgrade(
        address newImplementation
    ) internal override onlyRole(UPGRADER_ROLE) {}
}
```

---

### 3.3 Malware Genome DB (`contracts/src/MalwareGenomeDB.sol`)

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";

/// @title IMalwareGenomeDB
/// @notice Interface for the Malware Genome Database
interface IMalwareGenomeDB {
    /// @notice Genome label enum
    enum GenomeLabel {
        Unknown,
        Benign,
        Suspicious,
        KnownExploit
    }
    
    /// @notice Emitted when genome is registered
    event GenomeRegistered(
        bytes32 indexed genomeHash,
        string ipfsHash,
        GenomeLabel label,
        address indexed registrar
    );
    
    /// @notice Emitted when genome label is updated
    event GenomeLabelUpdated(
        bytes32 indexed genomeHash,
        GenomeLabel oldLabel,
        GenomeLabel newLabel
    );
    
    /// @notice Register a new genome
    /// @param genomeHash Hash of the genome
    /// @param ipfsHash IPFS hash of full analysis
    /// @param label Classification label
    function registerGenome(
        bytes32 genomeHash,
        string calldata ipfsHash,
        GenomeLabel label
    ) external;
    
    /// @notice Check if genome exists
    /// @param genomeHash Genome hash
    /// @return exists Whether genome exists
    function hasGenome(bytes32 genomeHash) external view returns (bool exists);
    
    /// @notice Get genome data
    /// @param genomeHash Genome hash
    /// @return ipfsHash IPFS hash
    /// @return label Classification label
    /// @return timestamp Registration timestamp
    function getGenome(
        bytes32 genomeHash
    ) external view returns (
        string memory ipfsHash,
        GenomeLabel label,
        uint256 timestamp
    );
}

/// @title MalwareGenomeDB
/// @notice On-chain malware genome fingerprint database
contract MalwareGenomeDB is 
    IMalwareGenomeDB,
    UUPSUpgradeable,
    AccessControlUpgradeable,
    PausableUpgradeable 
{
    // ============ Constants ============
    
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant REGISTRAR_ROLE = keccak256("REGISTRAR_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    
    // ============ State Variables ============
    
    /// @notice Genome storage
    mapping(bytes32 => GenomeData) public genomes;
    
    /// @notice Total genomes registered
    uint256 public totalGenomes;
    
    /// @notice Genomes by label count
    mapping(GenomeLabel => uint256) public genomesByLabel;
    
    // ============ Structs ============
    
    /// @notice Genome data struct
    struct GenomeData {
        string ipfsHash;
        GenomeLabel label;
        uint256 timestamp;
        address registrar;
        bool exists;
    }
    
    // ============ Errors ============
    
    error GenomeAlreadyExists(bytes32 genomeHash);
    error GenomeNotFound(bytes32 genomeHash);
    error InvalidIPFSHash();
    error ZeroAddress();
    
    // ============ Initialization ============
    
    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }
    
    function initialize(
        address admin,
        address registrar
    ) external initializer {
        if (admin == address(0) || registrar == address(0)) {
            revert ZeroAddress();
        }
        
        __UUPSUpgradeable_init();
        __AccessControl_init();
        __Pausable_init();
        
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(ADMIN_ROLE, admin);
        _grantRole(REGISTRAR_ROLE, registrar);
        _grantRole(UPGRADER_ROLE, admin);
    }
    
    // ============ External Functions ============
    
    /// @inheritdoc IMalwareGenomeDB
    function registerGenome(
        bytes32 genomeHash,
        string calldata ipfsHash,
        GenomeLabel label
    ) external override onlyRole(REGISTRAR_ROLE) whenNotPaused {
        if (genomes[genomeHash].exists) {
            revert GenomeAlreadyExists(genomeHash);
        }
        if (bytes(ipfsHash).length == 0) {
            revert InvalidIPFSHash();
        }
        
        genomes[genomeHash] = GenomeData({
            ipfsHash: ipfsHash,
            label: label,
            timestamp: block.timestamp,
            registrar: msg.sender,
            exists: true
        });
        
        totalGenomes++;
        genomesByLabel[label]++;
        
        emit GenomeRegistered(genomeHash, ipfsHash, label, msg.sender);
    }
    
    /// @inheritdoc IMalwareGenomeDB
    function hasGenome(
        bytes32 genomeHash
    ) external view override returns (bool) {
        return genomes[genomeHash].exists;
    }
    
    /// @inheritdoc IMalwareGenomeDB
    function getGenome(
        bytes32 genomeHash
    ) external view override returns (
        string memory ipfsHash,
        GenomeLabel label,
        uint256 timestamp
    ) {
        GenomeData storage data = genomes[genomeHash];
        if (!data.exists) {
            revert GenomeNotFound(genomeHash);
        }
        return (data.ipfsHash, data.label, data.timestamp);
    }
    
    /// @notice Update genome label
    /// @param genomeHash Genome hash
    /// @param newLabel New label
    function updateLabel(
        bytes32 genomeHash,
        GenomeLabel newLabel
    ) external onlyRole(ADMIN_ROLE) {
        GenomeData storage data = genomes[genomeHash];
        if (!data.exists) {
            revert GenomeNotFound(genomeHash);
        }
        
        GenomeLabel oldLabel = data.label;
        genomesByLabel[oldLabel]--;
        genomesByLabel[newLabel]++;
        data.label = newLabel;
        
        emit GenomeLabelUpdated(genomeHash, oldLabel, newLabel);
    }
    
    // ============ Admin Functions ============
    
    function addRegistrar(address registrar) external onlyRole(ADMIN_ROLE) {
        if (registrar == address(0)) revert ZeroAddress();
        _grantRole(REGISTRAR_ROLE, registrar);
    }
    
    function removeRegistrar(address registrar) external onlyRole(ADMIN_ROLE) {
        _revokeRole(REGISTRAR_ROLE, registrar);
    }
    
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }
    
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }
    
    function _authorizeUpgrade(
        address newImplementation
    ) internal override onlyRole(UPGRADER_ROLE) {}
}
```

---

### 3.4 Red-Team DAO (`contracts/src/RedTeamDAO.sol`)

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

/// @title IRedTeamDAO
/// @notice Interface for the Red-Team DAO bounty system
interface IRedTeamDAO {
    /// @notice Severity levels
    enum Severity {
        Low,
        Medium,
        High,
        Critical
    }
    
    /// @notice Submission status
    enum Status {
        Pending,
        Verifying,
        Verified,
        Rejected,
        Paid,
        Disputed
    }
    
    /// @notice Emitted when exploit is submitted
    event ExploitSubmitted(
        bytes32 indexed submissionId,
        address indexed researcher,
        address indexed target,
        Severity severity
    );
    
    /// @notice Emitted when exploit is verified
    event ExploitVerified(
        bytes32 indexed submissionId,
        uint256 bountyAmount
    );
    
    /// @notice Emitted when bounty is paid
    event BountyPaid(
        bytes32 indexed submissionId,
        address indexed researcher,
        uint256 amount
    );
    
    /// @notice Submit an exploit proof
    function submitExploit(
        address target,
        bytes calldata proof,
        string calldata description,
        Severity severity
    ) external returns (bytes32 submissionId);
    
    /// @notice Vote on a submission
    function vote(bytes32 submissionId, bool approve) external;
    
    /// @notice Claim bounty for verified submission
    function claimBounty(bytes32 submissionId) external;
}

/// @title RedTeamDAO
/// @notice Decentralized bug bounty platform
contract RedTeamDAO is 
    IRedTeamDAO,
    UUPSUpgradeable,
    AccessControlUpgradeable,
    PausableUpgradeable,
    ReentrancyGuardUpgradeable 
{
    // ============ Constants ============
    
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant VERIFIER_ROLE = keccak256("VERIFIER_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    
    uint256 public constant VOTING_PERIOD = 3 days;
    uint256 public constant MIN_VOTES_REQUIRED = 3;
    uint256 public constant VOTE_THRESHOLD_PERCENT = 66; // 66%
    
    // ============ State Variables ============
    
    /// @notice Submission storage
    mapping(bytes32 => Submission) public submissions;
    
    /// @notice Votes storage
    mapping(bytes32 => mapping(address => bool)) public hasVoted;
    
    /// @notice Bounty multipliers by severity (in basis points, 10000 = 1x)
    mapping(Severity => uint256) public bountyMultipliers;
    
    /// @notice Base bounty amount (in wei)
    uint256 public baseBountyAmount;
    
    /// @notice Treasury balance
    uint256 public treasuryBalance;
    
    /// @notice Total submissions
    uint256 public totalSubmissions;
    
    /// @notice Total bounties paid
    uint256 public totalBountiesPaid;
    
    // ============ Structs ============
    
    /// @notice Submission data
    struct Submission {
        address researcher;
        address target;
        bytes32 proofHash;
        string description;
        Severity severity;
        Status status;
        uint256 timestamp;
        uint256 votingEndsAt;
        uint256 votesFor;
        uint256 votesAgainst;
        uint256 bountyAmount;
        bool exists;
    }
    
    // ============ Errors ============
    
    error SubmissionNotFound(bytes32 submissionId);
    error AlreadyVoted();
    error VotingEnded();
    error VotingNotEnded();
    error NotVerified();
    error AlreadyPaid();
    error InsufficientTreasury();
    error InvalidProof();
    error ZeroAddress();
    
    // ============ Initialization ============
    
    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }
    
    function initialize(
        address admin,
        uint256 _baseBountyAmount
    ) external initializer {
        if (admin == address(0)) revert ZeroAddress();
        
        __UUPSUpgradeable_init();
        __AccessControl_init();
        __Pausable_init();
        __ReentrancyGuard_init();
        
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(ADMIN_ROLE, admin);
        _grantRole(VERIFIER_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        
        baseBountyAmount = _baseBountyAmount;
        
        // Set default multipliers
        bountyMultipliers[Severity.Low] = 10000;      // 1x
        bountyMultipliers[Severity.Medium] = 20000;   // 2x
        bountyMultipliers[Severity.High] = 50000;     // 5x
        bountyMultipliers[Severity.Critical] = 100000; // 10x
    }
    
    // ============ External Functions ============
    
    /// @inheritdoc IRedTeamDAO
    function submitExploit(
        address target,
        bytes calldata proof,
        string calldata description,
        Severity severity
    ) external override whenNotPaused returns (bytes32) {
        if (target == address(0)) revert ZeroAddress();
        if (proof.length < 32) revert InvalidProof();
        
        bytes32 submissionId = keccak256(
            abi.encodePacked(msg.sender, target, proof, block.timestamp)
        );
        
        submissions[submissionId] = Submission({
            researcher: msg.sender,
            target: target,
            proofHash: keccak256(proof),
            description: description,
            severity: severity,
            status: Status.Pending,
            timestamp: block.timestamp,
            votingEndsAt: block.timestamp + VOTING_PERIOD,
            votesFor: 0,
            votesAgainst: 0,
            bountyAmount: _calculateBounty(severity),
            exists: true
        });
        
        totalSubmissions++;
        
        emit ExploitSubmitted(submissionId, msg.sender, target, severity);
        
        return submissionId;
    }
    
    /// @inheritdoc IRedTeamDAO
    function vote(
        bytes32 submissionId,
        bool approve
    ) external override onlyRole(VERIFIER_ROLE) whenNotPaused {
        Submission storage sub = submissions[submissionId];
        if (!sub.exists) revert SubmissionNotFound(submissionId);
        if (hasVoted[submissionId][msg.sender]) revert AlreadyVoted();
        if (block.timestamp > sub.votingEndsAt) revert VotingEnded();
        
        hasVoted[submissionId][msg.sender] = true;
        
        if (approve) {
            sub.votesFor++;
        } else {
            sub.votesAgainst++;
        }
        
        // Check if can be finalized
        uint256 totalVotes = sub.votesFor + sub.votesAgainst;
        if (totalVotes >= MIN_VOTES_REQUIRED) {
            uint256 approvalPercent = (sub.votesFor * 100) / totalVotes;
            if (approvalPercent >= VOTE_THRESHOLD_PERCENT) {
                sub.status = Status.Verified;
                emit ExploitVerified(submissionId, sub.bountyAmount);
            } else if ((100 - approvalPercent) > (100 - VOTE_THRESHOLD_PERCENT)) {
                sub.status = Status.Rejected;
            }
        }
    }
    
    /// @inheritdoc IRedTeamDAO
    function claimBounty(
        bytes32 submissionId
    ) external override nonReentrant whenNotPaused {
        Submission storage sub = submissions[submissionId];
        if (!sub.exists) revert SubmissionNotFound(submissionId);
        if (sub.status != Status.Verified) revert NotVerified();
        if (sub.researcher != msg.sender) revert();
        
        uint256 amount = sub.bountyAmount;
        if (treasuryBalance < amount) revert InsufficientTreasury();
        
        sub.status = Status.Paid;
        treasuryBalance -= amount;
        totalBountiesPaid += amount;
        
        (bool success, ) = payable(msg.sender).call{value: amount}("");
        require(success, "Transfer failed");
        
        emit BountyPaid(submissionId, msg.sender, amount);
    }
    
    // ============ View Functions ============
    
    /// @notice Get submission details
    function getSubmission(
        bytes32 submissionId
    ) external view returns (Submission memory) {
        if (!submissions[submissionId].exists) {
            revert SubmissionNotFound(submissionId);
        }
        return submissions[submissionId];
    }
    
    // ============ Admin Functions ============
    
    /// @notice Fund treasury
    function fundTreasury() external payable onlyRole(ADMIN_ROLE) {
        treasuryBalance += msg.value;
    }
    
    /// @notice Set bounty multiplier
    function setBountyMultiplier(
        Severity severity,
        uint256 multiplier
    ) external onlyRole(ADMIN_ROLE) {
        bountyMultipliers[severity] = multiplier;
    }
    
    /// @notice Set base bounty amount
    function setBaseBountyAmount(
        uint256 amount
    ) external onlyRole(ADMIN_ROLE) {
        baseBountyAmount = amount;
    }
    
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }
    
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }
    
    // ============ Internal Functions ============
    
    function _calculateBounty(Severity severity) internal view returns (uint256) {
        return (baseBountyAmount * bountyMultipliers[severity]) / 10000;
    }
    
    function _authorizeUpgrade(
        address newImplementation
    ) internal override onlyRole(UPGRADER_ROLE) {}
    
    // ============ Receive ============
    
    receive() external payable {
        treasuryBalance += msg.value;
    }
}
```

---

## 4. Python ML Objects

### 4.1 Data Models (`ml/src/models/`)

```python
# models/__init__.py
from .features import BehavioralFeatures, GenomeFeatures
from .predictions import HumanPrediction, AnomalyPrediction
from .training import TrainingConfig, TrainingResult

# models/features.py
from __future__ import annotations
from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional, List, Dict, Any
import numpy as np
from pydantic import BaseModel, Field, validator

class BehavioralFeatures(BaseModel):
    """Extracted behavioral features for a wallet address."""
    
    wallet_address: str = Field(..., regex=r"^0x[a-fA-F0-9]{40}$")
    computed_at: datetime = Field(default_factory=datetime.utcnow)
    
    # Transaction metrics
    tx_count: int = Field(..., ge=0)
    avg_tx_interval: float = Field(..., ge=0)  # seconds
    tx_interval_variance: float = Field(..., ge=0)
    tx_interval_std: float = Field(..., ge=0)
    
    # Gas metrics
    avg_gas_used: float = Field(..., ge=0)
    gas_variance: float = Field(..., ge=0)
    max_gas_used: int = Field(..., ge=0)
    min_gas_used: int = Field(..., ge=0)
    
    # Interaction metrics
    interaction_diversity: int = Field(..., ge=0)
    unique_contracts: int = Field(..., ge=0)
    unique_methods: int = Field(..., ge=0)
    
    # Value metrics
    total_value_transferred: str = Field(...)  # wei as string
    avg_value_per_tx: float = Field(..., ge=0)
    
    # Category interactions
    erc20_interactions: int = Field(default=0, ge=0)
    defi_interactions: int = Field(default=0, ge=0)
    nft_interactions: int = Field(default=0, ge=0)
    
    # Time-based distributions
    time_of_day_distribution: List[int] = Field(default_factory=lambda: [0] * 24)
    day_of_week_distribution: List[int] = Field(default_factory=lambda: [0] * 7)
    
    # Activity metrics
    last_activity: Optional[datetime] = None
    account_age_days: int = Field(default=0, ge=0)
    active_days: int = Field(default=0, ge=0)
    
    @validator("time_of_day_distribution")
    def validate_tod(cls, v):
        if len(v) != 24:
            raise ValueError("time_of_day_distribution must have 24 elements")
        return v
    
    @validator("day_of_week_distribution")
    def validate_dow(cls, v):
        if len(v) != 7:
            raise ValueError("day_of_week_distribution must have 7 elements")
        return v
    
    def to_numpy(self) -> np.ndarray:
        """Convert to numpy array for ML inference."""
        features = [
            self.tx_count,
            self.avg_tx_interval,
            self.tx_interval_variance,
            self.tx_interval_std,
            self.avg_gas_used,
            self.gas_variance,
            self.interaction_diversity,
            self.unique_contracts,
            self.unique_methods,
            self.avg_value_per_tx,
            self.erc20_interactions,
            self.defi_interactions,
            self.nft_interactions,
            self.account_age_days,
            self.active_days,
        ]
        # Add time distributions
        features.extend(self.time_of_day_distribution)
        features.extend(self.day_of_week_distribution)
        return np.array(features, dtype=np.float32)
    
    def normalize(self, scaler: Any) -> np.ndarray:
        """Normalize features using provided scaler."""
        raw = self.to_numpy().reshape(1, -1)
        return scaler.transform(raw).flatten()
    
    @classmethod
    def feature_names(cls) -> List[str]:
        """Return feature names in order."""
        names = [
            "tx_count", "avg_tx_interval", "tx_interval_variance", "tx_interval_std",
            "avg_gas_used", "gas_variance", "interaction_diversity", "unique_contracts",
            "unique_methods", "avg_value_per_tx", "erc20_interactions", "defi_interactions",
            "nft_interactions", "account_age_days", "active_days",
        ]
        names.extend([f"hour_{i}" for i in range(24)])
        names.extend([f"dow_{i}" for i in range(7)])
        return names


class GenomeFeatures(BaseModel):
    """Extracted features from contract bytecode for genome analysis."""
    
    contract_address: Optional[str] = Field(None, regex=r"^0x[a-fA-F0-9]{40}$")
    bytecode_hash: str = Field(...)
    
    # Bytecode metrics
    bytecode_length: int = Field(..., ge=0)
    unique_opcodes: int = Field(..., ge=0)
    total_opcodes: int = Field(..., ge=0)
    
    # Opcode histogram (top 50 opcodes)
    opcode_histogram: Dict[str, int] = Field(default_factory=dict)
    
    # Control flow metrics
    jump_count: int = Field(default=0, ge=0)
    jumpi_count: int = Field(default=0, ge=0)
    jump_density: float = Field(default=0.0, ge=0)
    
    # Call metrics
    call_count: int = Field(default=0, ge=0)
    delegatecall_count: int = Field(default=0, ge=0)
    staticcall_count: int = Field(default=0, ge=0)
    external_call_count: int = Field(default=0, ge=0)
    
    # Storage metrics
    sload_count: int = Field(default=0, ge=0)
    sstore_count: int = Field(default=0, ge=0)
    storage_access_count: int = Field(default=0, ge=0)
    
    # Dangerous opcodes
    has_selfdestruct: bool = Field(default=False)
    has_delegatecall: bool = Field(default=False)
    has_create: bool = Field(default=False)
    has_create2: bool = Field(default=False)
    
    # Gas metrics
    estimated_deploy_gas: int = Field(default=0, ge=0)
    
    # Similarity features (for LSH)
    minhash_signature: Optional[List[int]] = None
    
    def to_numpy(self) -> np.ndarray:
        """Convert to numpy array for ML inference."""
        # Fixed feature vector
        features = [
            self.bytecode_length,
            self.unique_opcodes,
            self.total_opcodes,
            self.jump_count,
            self.jumpi_count,
            self.jump_density,
            self.call_count,
            self.delegatecall_count,
            self.staticcall_count,
            self.external_call_count,
            self.sload_count,
            self.sstore_count,
            self.storage_access_count,
            int(self.has_selfdestruct),
            int(self.has_delegatecall),
            int(self.has_create),
            int(self.has_create2),
            self.estimated_deploy_gas,
        ]
        
        # Add opcode histogram (top 50 padded)
        top_opcodes = [
            "PUSH1", "PUSH2", "DUP1", "DUP2", "SWAP1", "SWAP2", "POP", "MLOAD",
            "MSTORE", "SLOAD", "SSTORE", "JUMP", "JUMPI", "JUMPDEST", "ADD", "SUB",
            "MUL", "DIV", "AND", "OR", "XOR", "NOT", "LT", "GT", "EQ", "ISZERO",
            "CALLDATALOAD", "CALLDATASIZE", "CALLDATACOPY", "CODECOPY", "EXTCODESIZE",
            "RETURNDATASIZE", "RETURNDATACOPY", "CALL", "STATICCALL", "DELEGATECALL",
            "CREATE", "CREATE2", "REVERT", "RETURN", "STOP", "INVALID", "SELFDESTRUCT",
            "LOG0", "LOG1", "LOG2", "LOG3", "LOG4", "SHA3", "ADDRESS",
        ]
        for opcode in top_opcodes:
            features.append(self.opcode_histogram.get(opcode, 0))
        
        return np.array(features, dtype=np.float32)
    
    def compute_genome_hash(self) -> bytes:
        """Compute deterministic hash of features."""
        import hashlib
        data = self.to_numpy().tobytes()
        return hashlib.sha256(data).digest()


# models/predictions.py
from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from enum import Enum

class RiskLevel(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"

class HumanPrediction(BaseModel):
    """Prediction result for human vs bot classification."""
    
    wallet_address: str
    human_probability: float = Field(..., ge=0.0, le=1.0)
    bot_probability: float = Field(..., ge=0.0, le=1.0)
    is_human: bool
    confidence: float = Field(..., ge=0.0, le=1.0)
    risk_score: float = Field(..., ge=0.0, le=1.0)
    risk_level: RiskLevel
    predicted_at: datetime = Field(default_factory=datetime.utcnow)
    model_version: str
    
    # Feature importance (top contributors)
    top_features: List[FeatureImportance] = Field(default_factory=list)
    
    @classmethod
    def from_probabilities(
        cls,
        wallet_address: str,
        probs: tuple[float, float],
        model_version: str,
        threshold: float = 0.5,
    ) -> "HumanPrediction":
        """Create prediction from model output probabilities."""
        human_prob, bot_prob = probs
        is_human = human_prob >= threshold
        confidence = max(human_prob, bot_prob)
        
        # Risk score: higher if bot, lower if human
        risk_score = bot_prob
        
        # Determine risk level
        if risk_score < 0.3:
            risk_level = RiskLevel.LOW
        elif risk_score < 0.6:
            risk_level = RiskLevel.MEDIUM
        elif risk_score < 0.85:
            risk_level = RiskLevel.HIGH
        else:
            risk_level = RiskLevel.CRITICAL
        
        return cls(
            wallet_address=wallet_address,
            human_probability=human_prob,
            bot_probability=bot_prob,
            is_human=is_human,
            confidence=confidence,
            risk_score=risk_score,
            risk_level=risk_level,
            model_version=model_version,
        )


class FeatureImportance(BaseModel):
    """Feature importance in prediction."""
    feature_name: str
    importance: float
    value: float


class AnomalyPrediction(BaseModel):
    """Prediction result for anomaly detection."""
    
    wallet_address: str
    anomaly_score: float = Field(..., ge=0.0, le=1.0)
    is_anomaly: bool
    anomaly_type: Optional[str] = None
    confidence: float = Field(..., ge=0.0, le=1.0)
    predicted_at: datetime = Field(default_factory=datetime.utcnow)
    model_version: str
    
    # Anomaly details
    contributing_factors: List[str] = Field(default_factory=list)
    reconstruction_error: Optional[float] = None
    
    @classmethod
    def from_score(
        cls,
        wallet_address: str,
        score: float,
        model_version: str,
        threshold: float = 0.7,
        reconstruction_error: Optional[float] = None,
    ) -> "AnomalyPrediction":
        """Create prediction from anomaly score."""
        is_anomaly = score >= threshold
        
        return cls(
            wallet_address=wallet_address,
            anomaly_score=score,
            is_anomaly=is_anomaly,
            confidence=abs(score - 0.5) * 2,  # Distance from decision boundary
            model_version=model_version,
            reconstruction_error=reconstruction_error,
        )


class GenomePrediction(BaseModel):
    """Prediction result for genome classification."""
    
    bytecode_hash: str
    contract_address: Optional[str] = None
    label: str  # benign, suspicious, known_exploit, unknown
    label_probabilities: Dict[str, float]
    confidence: float = Field(..., ge=0.0, le=1.0)
    similar_genomes: List[SimilarGenome] = Field(default_factory=list)
    predicted_at: datetime = Field(default_factory=datetime.utcnow)
    model_version: str


class SimilarGenome(BaseModel):
    """Similar genome match."""
    genome_hash: str
    similarity_score: float
    label: str
    ipfs_hash: Optional[str] = None


# models/training.py
from pydantic import BaseModel, Field
from typing import Optional, Dict, Any, List
from datetime import datetime
from enum import Enum

class ModelType(str, Enum):
    HUMAN_CLASSIFIER = "human_classifier"
    ANOMALY_DETECTOR = "anomaly_detector"
    GENOME_CLASSIFIER = "genome_classifier"

class TrainingConfig(BaseModel):
    """Configuration for model training."""
    
    model_type: ModelType
    model_name: str
    version: str
    
    # Data config
    train_data_path: str
    val_data_path: Optional[str] = None
    test_data_path: Optional[str] = None
    
    # Model hyperparameters
    hyperparameters: Dict[str, Any] = Field(default_factory=dict)
    
    # Training settings
    epochs: int = Field(default=100, ge=1)
    batch_size: int = Field(default=32, ge=1)
    learning_rate: float = Field(default=1e-3, gt=0)
    early_stopping_patience: int = Field(default=10, ge=1)
    
    # Regularization
    dropout: float = Field(default=0.3, ge=0, le=1)
    weight_decay: float = Field(default=1e-5, ge=0)
    
    # Class weights for imbalanced data
    class_weights: Optional[Dict[str, float]] = None
    
    # Output
    output_dir: str = Field(default="./models")
    checkpoint_dir: str = Field(default="./checkpoints")
    
    # Experiment tracking
    experiment_name: Optional[str] = None
    tags: List[str] = Field(default_factory=list)
    
    class Config:
        use_enum_values = True


class TrainingResult(BaseModel):
    """Result from model training."""
    
    model_type: ModelType
    model_name: str
    version: str
    
    # Timestamps
    started_at: datetime
    completed_at: datetime
    duration_seconds: float
    
    # Final metrics
    train_loss: float
    val_loss: Optional[float] = None
    test_loss: Optional[float] = None
    
    # Classification metrics
    accuracy: Optional[float] = None
    precision: Optional[float] = None
    recall: Optional[float] = None
    f1_score: Optional[float] = None
    auc_roc: Optional[float] = None
    
    # Per-class metrics
    class_metrics: Optional[Dict[str, Dict[str, float]]] = None
    
    # Training history
    history: Dict[str, List[float]] = Field(default_factory=dict)
    
    # Model artifacts
    model_path: str
    onnx_path: Optional[str] = None
    scaler_path: Optional[str] = None
    
    # Config used
    config: TrainingConfig
    
    class Config:
        use_enum_values = True
```

---

### 4.2 ETL Pipeline (`ml/src/etl/`)

```python
# etl/__init__.py
from .extractors import TransactionExtractor, BytecodeExtractor
from .transformers import FeatureTransformer, Normalizer
from .loaders import ClickHouseLoader, ParquetLoader

# etl/extractors.py
from abc import ABC, abstractmethod
from typing import AsyncIterator, List, Optional, Dict, Any
from datetime import datetime, timedelta
import asyncio
import polars as pl
from pydantic import BaseModel
import httpx

class ExtractorConfig(BaseModel):
    """Configuration for data extractors."""
    rpc_url: str
    batch_size: int = 1000
    max_concurrent: int = 10
    timeout: int = 30
    retry_attempts: int = 3
    retry_delay: float = 1.0

class Extractor(ABC):
    """Base class for data extractors."""
    
    def __init__(self, config: ExtractorConfig):
        self.config = config
        self._client: Optional[httpx.AsyncClient] = None
    
    async def __aenter__(self):
        self._client = httpx.AsyncClient(timeout=self.config.timeout)
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self._client:
            await self._client.aclose()
    
    @abstractmethod
    async def extract(self, **kwargs) -> pl.DataFrame:
        """Extract data and return as Polars DataFrame."""
        pass


class TransactionExtractor(Extractor):
    """Extract transaction data from Ethereum RPC."""
    
    async def extract(
        self,
        addresses: List[str],
        start_block: int,
        end_block: int,
    ) -> pl.DataFrame:
        """Extract transactions for given addresses."""
        all_txs = []
        
        semaphore = asyncio.Semaphore(self.config.max_concurrent)
        
        async def fetch_address_txs(address: str) -> List[Dict]:
            async with semaphore:
                return await self._fetch_transactions(address, start_block, end_block)
        
        tasks = [fetch_address_txs(addr) for addr in addresses]
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        for result in results:
            if isinstance(result, list):
                all_txs.extend(result)
        
        if not all_txs:
            return pl.DataFrame()
        
        return pl.DataFrame(all_txs)
    
    async def _fetch_transactions(
        self,
        address: str,
        start_block: int,
        end_block: int,
    ) -> List[Dict]:
        """Fetch transactions for a single address."""
        txs = []
        
        # Use eth_getLogs for Transfer events + trace_filter for internal txs
        for attempt in range(self.config.retry_attempts):
            try:
                # Fetch outgoing transactions
                response = await self._client.post(
                    self.config.rpc_url,
                    json={
                        "jsonrpc": "2.0",
                        "method": "eth_getLogs",
                        "params": [{
                            "fromBlock": hex(start_block),
                            "toBlock": hex(end_block),
                            "address": address,
                        }],
                        "id": 1,
                    },
                )
                data = response.json()
                
                if "result" in data:
                    for log in data["result"]:
                        txs.append(self._parse_log(log, address))
                
                break
            except Exception as e:
                if attempt == self.config.retry_attempts - 1:
                    raise
                await asyncio.sleep(self.config.retry_delay * (attempt + 1))
        
        return txs
    
    def _parse_log(self, log: Dict, address: str) -> Dict:
        """Parse log entry into transaction record."""
        return {
            "address": address,
            "tx_hash": log.get("transactionHash"),
            "block_number": int(log.get("blockNumber", "0x0"), 16),
            "log_index": int(log.get("logIndex", "0x0"), 16),
            "topics": log.get("topics", []),
            "data": log.get("data"),
        }
    
    async def stream_transactions(
        self,
        addresses: List[str],
        start_block: int,
        end_block: int,
        chunk_size: int = 10000,
    ) -> AsyncIterator[pl.DataFrame]:
        """Stream transactions in chunks for large date ranges."""
        current_block = start_block
        
        while current_block < end_block:
            chunk_end = min(current_block + chunk_size, end_block)
            df = await self.extract(addresses, current_block, chunk_end)
            if not df.is_empty():
                yield df
            current_block = chunk_end


class BytecodeExtractor(Extractor):
    """Extract contract bytecode from Ethereum RPC."""
    
    async def extract(self, addresses: List[str]) -> pl.DataFrame:
        """Extract bytecode for contract addresses."""
        results = []
        
        semaphore = asyncio.Semaphore(self.config.max_concurrent)
        
        async def fetch_bytecode(address: str) -> Dict:
            async with semaphore:
                return await self._fetch_bytecode(address)
        
        tasks = [fetch_bytecode(addr) for addr in addresses]
        bytecodes = await asyncio.gather(*tasks, return_exceptions=True)
        
        for addr, result in zip(addresses, bytecodes):
            if isinstance(result, dict):
                results.append(result)
        
        return pl.DataFrame(results)
    
    async def _fetch_bytecode(self, address: str) -> Dict:
        """Fetch bytecode for a single address."""
        for attempt in range(self.config.retry_attempts):
            try:
                response = await self._client.post(
                    self.config.rpc_url,
                    json={
                        "jsonrpc": "2.0",
                        "method": "eth_getCode",
                        "params": [address, "latest"],
                        "id": 1,
                    },
                )
                data = response.json()
                
                bytecode = data.get("result", "0x")
                
                return {
                    "address": address,
                    "bytecode": bytecode,
                    "bytecode_length": (len(bytecode) - 2) // 2 if bytecode != "0x" else 0,
                    "is_contract": bytecode != "0x",
                    "extracted_at": datetime.utcnow().isoformat(),
                }
            except Exception as e:
                if attempt == self.config.retry_attempts - 1:
                    return {
                        "address": address,
                        "bytecode": None,
                        "error": str(e),
                    }
                await asyncio.sleep(self.config.retry_delay * (attempt + 1))


# etl/transformers.py
from abc import ABC, abstractmethod
from typing import Optional, List, Dict, Any
import polars as pl
import numpy as np
from sklearn.preprocessing import StandardScaler, MinMaxScaler
import joblib

class Transformer(ABC):
    """Base class for data transformers."""
    
    @abstractmethod
    def transform(self, df: pl.DataFrame) -> pl.DataFrame:
        """Transform the DataFrame."""
        pass
    
    @abstractmethod
    def fit(self, df: pl.DataFrame) -> "Transformer":
        """Fit the transformer on data."""
        pass


class FeatureTransformer(Transformer):
    """Transform raw transaction data into behavioral features."""
    
    def fit(self, df: pl.DataFrame) -> "FeatureTransformer":
        """No fitting required for feature extraction."""
        return self
    
    def transform(self, df: pl.DataFrame) -> pl.DataFrame:
        """Transform transactions into behavioral features per address."""
        if df.is_empty():
            return pl.DataFrame()
        
        # Group by address and compute features
        features = df.group_by("address").agg([
            # Transaction count
            pl.count().alias("tx_count"),
            
            # Time intervals
            pl.col("timestamp").diff().mean().alias("avg_tx_interval"),
            pl.col("timestamp").diff().std().alias("tx_interval_std"),
            pl.col("timestamp").diff().var().alias("tx_interval_variance"),
            
            # Gas metrics
            pl.col("gas_used").mean().alias("avg_gas_used"),
            pl.col("gas_used").var().alias("gas_variance"),
            pl.col("gas_used").max().alias("max_gas_used"),
            pl.col("gas_used").min().alias("min_gas_used"),
            
            # Interaction diversity
            pl.col("to_address").n_unique().alias("unique_contracts"),
            pl.col("method_id").n_unique().alias("unique_methods"),
            
            # Value metrics
            pl.col("value").sum().alias("total_value_transferred"),
            pl.col("value").mean().alias("avg_value_per_tx"),
            
            # Activity
            pl.col("timestamp").max().alias("last_activity"),
            pl.col("timestamp").min().alias("first_activity"),
        ])
        
        # Compute derived features
        features = features.with_columns([
            # Account age in days
            ((pl.col("last_activity") - pl.col("first_activity")).dt.total_days())
                .alias("account_age_days"),
            
            # Interaction diversity score
            (pl.col("unique_contracts") + pl.col("unique_methods"))
                .alias("interaction_diversity"),
        ])
        
        return features
    
    def transform_bytecode(self, df: pl.DataFrame) -> pl.DataFrame:
        """Transform bytecode into genome features."""
        features = []
        
        for row in df.iter_rows(named=True):
            bytecode = row.get("bytecode", "0x")
            if not bytecode or bytecode == "0x":
                continue
            
            feature = self._extract_bytecode_features(bytecode, row["address"])
            features.append(feature)
        
        return pl.DataFrame(features)
    
    def _extract_bytecode_features(self, bytecode: str, address: str) -> Dict:
        """Extract features from bytecode."""
        # Remove 0x prefix
        bytecode = bytecode[2:] if bytecode.startswith("0x") else bytecode
        
        # Parse opcodes
        opcodes = self._parse_opcodes(bytes.fromhex(bytecode))
        
        # Compute histogram
        opcode_counts = {}
        for opcode in opcodes:
            opcode_counts[opcode] = opcode_counts.get(opcode, 0) + 1
        
        return {
            "address": address,
            "bytecode_length": len(bytecode) // 2,
            "unique_opcodes": len(opcode_counts),
            "total_opcodes": len(opcodes),
            "jump_count": opcode_counts.get("JUMP", 0),
            "jumpi_count": opcode_counts.get("JUMPI", 0),
            "jump_density": (opcode_counts.get("JUMP", 0) + opcode_counts.get("JUMPI", 0)) / max(len(opcodes), 1),
            "call_count": opcode_counts.get("CALL", 0),
            "delegatecall_count": opcode_counts.get("DELEGATECALL", 0),
            "staticcall_count": opcode_counts.get("STATICCALL", 0),
            "sload_count": opcode_counts.get("SLOAD", 0),
            "sstore_count": opcode_counts.get("SSTORE", 0),
            "has_selfdestruct": "SELFDESTRUCT" in opcode_counts,
            "has_delegatecall": "DELEGATECALL" in opcode_counts,
            "has_create": "CREATE" in opcode_counts,
            "has_create2": "CREATE2" in opcode_counts,
            "opcode_histogram": opcode_counts,
        }
    
    def _parse_opcodes(self, bytecode: bytes) -> List[str]:
        """Parse bytecode into opcode names."""
        # EVM opcode mapping (simplified)
        OPCODES = {
            0x00: "STOP", 0x01: "ADD", 0x02: "MUL", 0x03: "SUB", 0x04: "DIV",
            0x10: "LT", 0x11: "GT", 0x14: "EQ", 0x15: "ISZERO",
            0x16: "AND", 0x17: "OR", 0x18: "XOR", 0x19: "NOT",
            0x20: "SHA3",
            0x30: "ADDRESS", 0x31: "BALANCE", 0x32: "ORIGIN", 0x33: "CALLER",
            0x34: "CALLVALUE", 0x35: "CALLDATALOAD", 0x36: "CALLDATASIZE",
            0x37: "CALLDATACOPY", 0x38: "CODESIZE", 0x39: "CODECOPY",
            0x50: "POP", 0x51: "MLOAD", 0x52: "MSTORE", 0x54: "SLOAD", 0x55: "SSTORE",
            0x56: "JUMP", 0x57: "JUMPI", 0x5b: "JUMPDEST",
            0x60: "PUSH1", 0x7f: "PUSH32",
            0x80: "DUP1", 0x8f: "DUP16",
            0x90: "SWAP1", 0x9f: "SWAP16",
            0xa0: "LOG0", 0xa4: "LOG4",
            0xf0: "CREATE", 0xf1: "CALL", 0xf2: "CALLCODE", 0xf3: "RETURN",
            0xf4: "DELEGATECALL", 0xf5: "CREATE2", 0xfa: "STATICCALL",
            0xfd: "REVERT", 0xfe: "INVALID", 0xff: "SELFDESTRUCT",
        }
        
        opcodes = []
        i = 0
        while i < len(bytecode):
            op = bytecode[i]
            name = OPCODES.get(op, f"UNKNOWN_{hex(op)}")
            
            # Handle PUSH instructions
            if 0x60 <= op <= 0x7f:
                push_size = op - 0x5f
                name = f"PUSH{push_size}"
                i += push_size
            
            opcodes.append(name)
            i += 1
        
        return opcodes


class Normalizer(Transformer):
    """Normalize features for ML training."""
    
    def __init__(self, method: str = "standard"):
        self.method = method
        self._scaler: Optional[Any] = None
        self._feature_columns: List[str] = []
    
    def fit(self, df: pl.DataFrame) -> "Normalizer":
        """Fit the normalizer on training data."""
        # Get numeric columns
        self._feature_columns = [
            col for col in df.columns
            if df[col].dtype in [pl.Float64, pl.Float32, pl.Int64, pl.Int32]
        ]
        
        # Create scaler
        if self.method == "standard":
            self._scaler = StandardScaler()
        elif self.method == "minmax":
            self._scaler = MinMaxScaler()
        else:
            raise ValueError(f"Unknown normalization method: {self.method}")
        
        # Fit scaler
        data = df.select(self._feature_columns).to_numpy()
        self._scaler.fit(data)
        
        return self
    
    def transform(self, df: pl.DataFrame) -> pl.DataFrame:
        """Normalize the DataFrame."""
        if self._scaler is None:
            raise RuntimeError("Normalizer not fitted. Call fit() first.")
        
        data = df.select(self._feature_columns).to_numpy()
        normalized = self._scaler.transform(data)
        
        # Create new DataFrame with normalized values
        normalized_df = pl.DataFrame(
            normalized,
            schema=self._feature_columns,
        )
        
        # Add back non-numeric columns
        for col in df.columns:
            if col not in self._feature_columns:
                normalized_df = normalized_df.with_columns(df[col].alias(col))
        
        return normalized_df
    
    def save(self, path: str):
        """Save normalizer to file."""
        joblib.dump({
            "scaler": self._scaler,
            "feature_columns": self._feature_columns,
            "method": self.method,
        }, path)
    
    @classmethod
    def load(cls, path: str) -> "Normalizer":
        """Load normalizer from file."""
        data = joblib.load(path)
        normalizer = cls(method=data["method"])
        normalizer._scaler = data["scaler"]
        normalizer._feature_columns = data["feature_columns"]
        return normalizer


# etl/loaders.py
from abc import ABC, abstractmethod
from typing import Optional
import polars as pl
from datetime import datetime

class Loader(ABC):
    """Base class for data loaders."""
    
    @abstractmethod
    async def load(self, df: pl.DataFrame, table: str) -> int:
        """Load DataFrame into storage. Returns rows loaded."""
        pass


class ClickHouseLoader(Loader):
    """Load data into ClickHouse."""
    
    def __init__(
        self,
        host: str,
        port: int = 8123,
        database: str = "vigilum",
        user: str = "default",
        password: str = "",
    ):
        self.host = host
        self.port = port
        self.database = database
        self.user = user
        self.password = password
        self._client: Optional[Any] = None
    
    async def connect(self):
        """Connect to ClickHouse."""
        import clickhouse_connect
        self._client = clickhouse_connect.get_client(
            host=self.host,
            port=self.port,
            database=self.database,
            username=self.user,
            password=self.password,
        )
    
    async def load(self, df: pl.DataFrame, table: str) -> int:
        """Load DataFrame into ClickHouse table."""
        if self._client is None:
            await self.connect()
        
        # Convert to pandas for clickhouse-connect
        pdf = df.to_pandas()
        
        # Insert data
        self._client.insert_df(table, pdf)
        
        return len(df)
    
    async def create_table(self, table: str, schema: dict):
        """Create table with given schema."""
        if self._client is None:
            await self.connect()
        
        columns = ", ".join([f"{k} {v}" for k, v in schema.items()])
        query = f"""
        CREATE TABLE IF NOT EXISTS {self.database}.{table} (
            {columns}
        ) ENGINE = MergeTree()
        ORDER BY (address, timestamp)
        """
        self._client.command(query)


class ParquetLoader(Loader):
    """Load data into Parquet files."""
    
    def __init__(self, base_path: str, partition_by: Optional[List[str]] = None):
        self.base_path = base_path
        self.partition_by = partition_by or []
    
    async def load(self, df: pl.DataFrame, table: str) -> int:
        """Load DataFrame into Parquet file."""
        timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
        path = f"{self.base_path}/{table}/{timestamp}.parquet"
        
        # Ensure directory exists
        import os
        os.makedirs(os.path.dirname(path), exist_ok=True)
        
        # Write parquet
        df.write_parquet(path, compression="zstd")
        
        return len(df)
```

---

### 4.3 Training Pipeline (`ml/src/training/`)

```python
# training/__init__.py
from .trainers import HumanClassifierTrainer, AnomalyDetectorTrainer, GenomeClassifierTrainer
from .callbacks import EarlyStopping, ModelCheckpoint, MetricsLogger

# training/trainers.py
from abc import ABC, abstractmethod
from typing import Optional, Dict, Any, Tuple, List
from datetime import datetime
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, Dataset
import numpy as np
import polars as pl
from pathlib import Path
import onnx
import onnxruntime

from ..models import TrainingConfig, TrainingResult, ModelType


class BaseTrainer(ABC):
    """Base trainer class."""
    
    def __init__(self, config: TrainingConfig):
        self.config = config
        self.model: Optional[nn.Module] = None
        self.optimizer: Optional[optim.Optimizer] = None
        self.scheduler: Optional[Any] = None
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        self.history: Dict[str, List[float]] = {
            "train_loss": [],
            "val_loss": [],
            "accuracy": [],
        }
    
    @abstractmethod
    def _build_model(self) -> nn.Module:
        """Build the model architecture."""
        pass
    
    @abstractmethod
    def _create_dataset(self, data_path: str) -> Dataset:
        """Create dataset from data path."""
        pass
    
    def train(self) -> TrainingResult:
        """Execute training pipeline."""
        started_at = datetime.utcnow()
        
        # Build model
        self.model = self._build_model().to(self.device)
        
        # Create optimizer
        self.optimizer = optim.AdamW(
            self.model.parameters(),
            lr=self.config.learning_rate,
            weight_decay=self.config.weight_decay,
        )
        
        # Create datasets and loaders
        train_dataset = self._create_dataset(self.config.train_data_path)
        train_loader = DataLoader(
            train_dataset,
            batch_size=self.config.batch_size,
            shuffle=True,
            num_workers=4,
        )
        
        val_loader = None
        if self.config.val_data_path:
            val_dataset = self._create_dataset(self.config.val_data_path)
            val_loader = DataLoader(
                val_dataset,
                batch_size=self.config.batch_size,
                shuffle=False,
                num_workers=4,
            )
        
        # Training loop
        best_val_loss = float("inf")
        patience_counter = 0
        
        for epoch in range(self.config.epochs):
            # Train epoch
            train_loss = self._train_epoch(train_loader)
            self.history["train_loss"].append(train_loss)
            
            # Validation
            if val_loader:
                val_loss, val_metrics = self._validate(val_loader)
                self.history["val_loss"].append(val_loss)
                self.history["accuracy"].append(val_metrics.get("accuracy", 0))
                
                # Early stopping
                if val_loss < best_val_loss:
                    best_val_loss = val_loss
                    patience_counter = 0
                    self._save_checkpoint(epoch, val_loss)
                else:
                    patience_counter += 1
                    if patience_counter >= self.config.early_stopping_patience:
                        print(f"Early stopping at epoch {epoch}")
                        break
            
            print(f"Epoch {epoch}: train_loss={train_loss:.4f}, val_loss={val_loss:.4f}")
        
        # Save final model
        model_path = self._save_model()
        onnx_path = self._export_onnx()
        
        completed_at = datetime.utcnow()
        
        # Compute final metrics
        final_metrics = {}
        if self.config.test_data_path:
            test_dataset = self._create_dataset(self.config.test_data_path)
            test_loader = DataLoader(test_dataset, batch_size=self.config.batch_size)
            _, final_metrics = self._validate(test_loader)
        
        return TrainingResult(
            model_type=self.config.model_type,
            model_name=self.config.model_name,
            version=self.config.version,
            started_at=started_at,
            completed_at=completed_at,
            duration_seconds=(completed_at - started_at).total_seconds(),
            train_loss=self.history["train_loss"][-1],
            val_loss=self.history["val_loss"][-1] if self.history["val_loss"] else None,
            accuracy=final_metrics.get("accuracy"),
            precision=final_metrics.get("precision"),
            recall=final_metrics.get("recall"),
            f1_score=final_metrics.get("f1"),
            auc_roc=final_metrics.get("auc_roc"),
            history=self.history,
            model_path=model_path,
            onnx_path=onnx_path,
            config=self.config,
        )
    
    def _train_epoch(self, loader: DataLoader) -> float:
        """Train for one epoch."""
        self.model.train()
        total_loss = 0.0
        
        for batch in loader:
            self.optimizer.zero_grad()
            
            inputs, targets = batch
            inputs = inputs.to(self.device)
            targets = targets.to(self.device)
            
            outputs = self.model(inputs)
            loss = self._compute_loss(outputs, targets)
            
            loss.backward()
            self.optimizer.step()
            
            total_loss += loss.item()
        
        return total_loss / len(loader)
    
    def _validate(self, loader: DataLoader) -> Tuple[float, Dict[str, float]]:
        """Validate model."""
        self.model.eval()
        total_loss = 0.0
        all_preds = []
        all_targets = []
        
        with torch.no_grad():
            for batch in loader:
                inputs, targets = batch
                inputs = inputs.to(self.device)
                targets = targets.to(self.device)
                
                outputs = self.model(inputs)
                loss = self._compute_loss(outputs, targets)
                
                total_loss += loss.item()
                all_preds.extend(outputs.argmax(dim=1).cpu().numpy())
                all_targets.extend(targets.cpu().numpy())
        
        # Compute metrics
        from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score
        
        metrics = {
            "accuracy": accuracy_score(all_targets, all_preds),
            "precision": precision_score(all_targets, all_preds, average="weighted", zero_division=0),
            "recall": recall_score(all_targets, all_preds, average="weighted", zero_division=0),
            "f1": f1_score(all_targets, all_preds, average="weighted", zero_division=0),
        }
        
        return total_loss / len(loader), metrics
    
    @abstractmethod
    def _compute_loss(self, outputs: torch.Tensor, targets: torch.Tensor) -> torch.Tensor:
        """Compute loss."""
        pass
    
    def _save_checkpoint(self, epoch: int, val_loss: float):
        """Save training checkpoint."""
        path = Path(self.config.checkpoint_dir) / f"{self.config.model_name}_epoch{epoch}.pt"
        path.parent.mkdir(parents=True, exist_ok=True)
        
        torch.save({
            "epoch": epoch,
            "model_state_dict": self.model.state_dict(),
            "optimizer_state_dict": self.optimizer.state_dict(),
            "val_loss": val_loss,
            "config": self.config.dict(),
        }, path)
    
    def _save_model(self) -> str:
        """Save final model."""
        path = Path(self.config.output_dir) / f"{self.config.model_name}_v{self.config.version}.pt"
        path.parent.mkdir(parents=True, exist_ok=True)
        
        torch.save({
            "model_state_dict": self.model.state_dict(),
            "config": self.config.dict(),
        }, path)
        
        return str(path)
    
    def _export_onnx(self) -> str:
        """Export model to ONNX format."""
        path = Path(self.config.output_dir) / f"{self.config.model_name}_v{self.config.version}.onnx"
        
        # Get input shape from config
        input_shape = self.config.hyperparameters.get("input_shape", (1, 46))
        dummy_input = torch.randn(*input_shape).to(self.device)
        
        torch.onnx.export(
            self.model,
            dummy_input,
            str(path),
            export_params=True,
            opset_version=13,
            do_constant_folding=True,
            input_names=["input"],
            output_names=["output"],
            dynamic_axes={
                "input": {0: "batch_size"},
                "output": {0: "batch_size"},
            },
        )
        
        return str(path)


class HumanClassifierTrainer(BaseTrainer):
    """Trainer for human vs bot classification model."""
    
    def _build_model(self) -> nn.Module:
        """Build human classifier model."""
        input_dim = self.config.hyperparameters.get("input_dim", 46)
        hidden_dims = self.config.hyperparameters.get("hidden_dims", [128, 64, 32])
        
        layers = []
        prev_dim = input_dim
        
        for hidden_dim in hidden_dims:
            layers.extend([
                nn.Linear(prev_dim, hidden_dim),
                nn.BatchNorm1d(hidden_dim),
                nn.ReLU(),
                nn.Dropout(self.config.dropout),
            ])
            prev_dim = hidden_dim
        
        layers.append(nn.Linear(prev_dim, 2))  # Binary classification
        
        return nn.Sequential(*layers)
    
    def _create_dataset(self, data_path: str) -> Dataset:
        """Create dataset from parquet file."""
        df = pl.read_parquet(data_path)
        
        # Get feature columns (exclude label and address)
        feature_cols = [c for c in df.columns if c not in ["label", "address", "wallet_address"]]
        
        features = torch.tensor(df.select(feature_cols).to_numpy(), dtype=torch.float32)
        labels = torch.tensor(df["label"].to_numpy(), dtype=torch.long)
        
        return torch.utils.data.TensorDataset(features, labels)
    
    def _compute_loss(self, outputs: torch.Tensor, targets: torch.Tensor) -> torch.Tensor:
        """Compute cross-entropy loss with optional class weights."""
        weights = None
        if self.config.class_weights:
            weights = torch.tensor(
                [self.config.class_weights.get(str(i), 1.0) for i in range(2)],
                device=self.device,
            )
        
        criterion = nn.CrossEntropyLoss(weight=weights)
        return criterion(outputs, targets)


class AnomalyDetectorTrainer(BaseTrainer):
    """Trainer for anomaly detection model (autoencoder)."""
    
    def _build_model(self) -> nn.Module:
        """Build autoencoder model."""
        input_dim = self.config.hyperparameters.get("input_dim", 46)
        latent_dim = self.config.hyperparameters.get("latent_dim", 8)
        encoder_dims = self.config.hyperparameters.get("encoder_dims", [32, 16])
        
        return Autoencoder(input_dim, latent_dim, encoder_dims, self.config.dropout)
    
    def _create_dataset(self, data_path: str) -> Dataset:
        """Create dataset from parquet file."""
        df = pl.read_parquet(data_path)
        feature_cols = [c for c in df.columns if c not in ["label", "address"]]
        features = torch.tensor(df.select(feature_cols).to_numpy(), dtype=torch.float32)
        
        # For autoencoder, target is same as input
        return torch.utils.data.TensorDataset(features, features)
    
    def _compute_loss(self, outputs: torch.Tensor, targets: torch.Tensor) -> torch.Tensor:
        """Compute reconstruction loss (MSE)."""
        return nn.MSELoss()(outputs, targets)


class Autoencoder(nn.Module):
    """Autoencoder for anomaly detection."""
    
    def __init__(
        self,
        input_dim: int,
        latent_dim: int,
        encoder_dims: List[int],
        dropout: float = 0.3,
    ):
        super().__init__()
        
        # Encoder
        encoder_layers = []
        prev_dim = input_dim
        for dim in encoder_dims:
            encoder_layers.extend([
                nn.Linear(prev_dim, dim),
                nn.ReLU(),
                nn.Dropout(dropout),
            ])
            prev_dim = dim
        encoder_layers.append(nn.Linear(prev_dim, latent_dim))
        self.encoder = nn.Sequential(*encoder_layers)
        
        # Decoder (mirror of encoder)
        decoder_layers = []
        prev_dim = latent_dim
        for dim in reversed(encoder_dims):
            decoder_layers.extend([
                nn.Linear(prev_dim, dim),
                nn.ReLU(),
                nn.Dropout(dropout),
            ])
            prev_dim = dim
        decoder_layers.append(nn.Linear(prev_dim, input_dim))
        self.decoder = nn.Sequential(*decoder_layers)
    
    def forward(self, x: torch.Tensor) -> torch.Tensor:
        latent = self.encoder(x)
        reconstructed = self.decoder(latent)
        return reconstructed
    
    def encode(self, x: torch.Tensor) -> torch.Tensor:
        return self.encoder(x)
    
    def get_anomaly_score(self, x: torch.Tensor) -> torch.Tensor:
        """Compute anomaly score as reconstruction error."""
        reconstructed = self.forward(x)
        return torch.mean((x - reconstructed) ** 2, dim=1)
```

---

### 4.4 Inference Service (`ml/src/inference/`)

```python
# inference/__init__.py
from .service import InferenceService
from .model_loader import ModelLoader

# inference/service.py
from typing import Optional, Dict, Any, List
from datetime import datetime
import numpy as np
import onnxruntime as ort
from pathlib import Path
import asyncio
from concurrent.futures import ThreadPoolExecutor

from ..models import (
    BehavioralFeatures,
    GenomeFeatures,
    HumanPrediction,
    AnomalyPrediction,
    GenomePrediction,
)
from ..etl.transformers import Normalizer


class InferenceService:
    """ML inference service for production."""
    
    def __init__(
        self,
        human_model_path: str,
        anomaly_model_path: str,
        genome_model_path: Optional[str] = None,
        scaler_path: Optional[str] = None,
        max_workers: int = 4,
    ):
        self.human_model_path = human_model_path
        self.anomaly_model_path = anomaly_model_path
        self.genome_model_path = genome_model_path
        self.scaler_path = scaler_path
        
        self._human_session: Optional[ort.InferenceSession] = None
        self._anomaly_session: Optional[ort.InferenceSession] = None
        self._genome_session: Optional[ort.InferenceSession] = None
        self._scaler: Optional[Normalizer] = None
        
        self._executor = ThreadPoolExecutor(max_workers=max_workers)
        self._loaded = False
        self._model_versions: Dict[str, str] = {}
    
    async def load(self):
        """Load all models."""
        loop = asyncio.get_event_loop()
        
        # Load models in parallel
        await asyncio.gather(
            loop.run_in_executor(self._executor, self._load_human_model),
            loop.run_in_executor(self._executor, self._load_anomaly_model),
            loop.run_in_executor(self._executor, self._load_genome_model),
            loop.run_in_executor(self._executor, self._load_scaler),
        )
        
        self._loaded = True
    
    def _load_human_model(self):
        """Load human classifier model."""
        opts = ort.SessionOptions()
        opts.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL
        opts.intra_op_num_threads = 2
        
        self._human_session = ort.InferenceSession(
            self.human_model_path,
            sess_options=opts,
            providers=["CPUExecutionProvider"],
        )
        
        # Extract version from path
        self._model_versions["human"] = Path(self.human_model_path).stem
    
    def _load_anomaly_model(self):
        """Load anomaly detection model."""
        opts = ort.SessionOptions()
        opts.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL
        
        self._anomaly_session = ort.InferenceSession(
            self.anomaly_model_path,
            sess_options=opts,
            providers=["CPUExecutionProvider"],
        )
        
        self._model_versions["anomaly"] = Path(self.anomaly_model_path).stem
    
    def _load_genome_model(self):
        """Load genome classifier model."""
        if not self.genome_model_path:
            return
        
        opts = ort.SessionOptions()
        opts.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL
        
        self._genome_session = ort.InferenceSession(
            self.genome_model_path,
            sess_options=opts,
            providers=["CPUExecutionProvider"],
        )
        
        self._model_versions["genome"] = Path(self.genome_model_path).stem
    
    def _load_scaler(self):
        """Load feature scaler."""
        if self.scaler_path:
            self._scaler = Normalizer.load(self.scaler_path)
    
    async def predict_human(
        self,
        features: BehavioralFeatures,
        threshold: float = 0.5,
    ) -> HumanPrediction:
        """Predict if address is human."""
        if not self._loaded:
            await self.load()
        
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(
            self._executor,
            self._predict_human_sync,
            features,
            threshold,
        )
    
    def _predict_human_sync(
        self,
        features: BehavioralFeatures,
        threshold: float,
    ) -> HumanPrediction:
        """Synchronous human prediction."""
        # Prepare input
        input_array = features.to_numpy().reshape(1, -1)
        
        # Normalize if scaler available
        if self._scaler:
            input_array = self._scaler._scaler.transform(input_array)
        
        # Run inference
        input_name = self._human_session.get_inputs()[0].name
        output_name = self._human_session.get_outputs()[0].name
        
        outputs = self._human_session.run(
            [output_name],
            {input_name: input_array.astype(np.float32)},
        )[0]
        
        # Apply softmax
        exp_outputs = np.exp(outputs - np.max(outputs))
        probs = exp_outputs / exp_outputs.sum()
        
        # Create prediction
        return HumanPrediction.from_probabilities(
            wallet_address=features.wallet_address,
            probs=(probs[0][0], probs[0][1]),  # (human_prob, bot_prob)
            model_version=self._model_versions["human"],
            threshold=threshold,
        )
    
    async def predict_anomaly(
        self,
        features: BehavioralFeatures,
        threshold: float = 0.7,
    ) -> AnomalyPrediction:
        """Predict if behavior is anomalous."""
        if not self._loaded:
            await self.load()
        
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(
            self._executor,
            self._predict_anomaly_sync,
            features,
            threshold,
        )
    
    def _predict_anomaly_sync(
        self,
        features: BehavioralFeatures,
        threshold: float,
    ) -> AnomalyPrediction:
        """Synchronous anomaly prediction."""
        input_array = features.to_numpy().reshape(1, -1)
        
        if self._scaler:
            input_array = self._scaler._scaler.transform(input_array)
        
        input_name = self._anomaly_session.get_inputs()[0].name
        output_name = self._anomaly_session.get_outputs()[0].name
        
        # Get reconstruction
        reconstructed = self._anomaly_session.run(
            [output_name],
            {input_name: input_array.astype(np.float32)},
        )[0]
        
        # Compute reconstruction error
        reconstruction_error = np.mean((input_array - reconstructed) ** 2)
        
        # Normalize to 0-1 score (using sigmoid-like transformation)
        anomaly_score = 1 / (1 + np.exp(-reconstruction_error + 0.5))
        
        return AnomalyPrediction.from_score(
            wallet_address=features.wallet_address,
            score=float(anomaly_score),
            model_version=self._model_versions["anomaly"],
            threshold=threshold,
            reconstruction_error=float(reconstruction_error),
        )
    
    async def predict_genome(
        self,
        features: GenomeFeatures,
    ) -> GenomePrediction:
        """Classify contract genome."""
        if not self._loaded:
            await self.load()
        
        if not self._genome_session:
            raise RuntimeError("Genome model not loaded")
        
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(
            self._executor,
            self._predict_genome_sync,
            features,
        )
    
    def _predict_genome_sync(self, features: GenomeFeatures) -> GenomePrediction:
        """Synchronous genome prediction."""
        input_array = features.to_numpy().reshape(1, -1)
        
        input_name = self._genome_session.get_inputs()[0].name
        output_name = self._genome_session.get_outputs()[0].name
        
        outputs = self._genome_session.run(
            [output_name],
            {input_name: input_array.astype(np.float32)},
        )[0]
        
        # Labels
        labels = ["benign", "suspicious", "known_exploit", "unknown"]
        
        # Apply softmax
        exp_outputs = np.exp(outputs - np.max(outputs))
        probs = exp_outputs / exp_outputs.sum()
        
        # Get prediction
        pred_idx = np.argmax(probs)
        
        return GenomePrediction(
            bytecode_hash=features.bytecode_hash,
            contract_address=features.contract_address,
            label=labels[pred_idx],
            label_probabilities={
                label: float(probs[0][i]) for i, label in enumerate(labels)
            },
            confidence=float(probs[0][pred_idx]),
            model_version=self._model_versions["genome"],
        )
    
    async def batch_predict_human(
        self,
        features_list: List[BehavioralFeatures],
        threshold: float = 0.5,
    ) -> List[HumanPrediction]:
        """Batch predict for multiple addresses."""
        tasks = [self.predict_human(f, threshold) for f in features_list]
        return await asyncio.gather(*tasks)
    
    async def reload(self):
        """Reload all models (for hot reload)."""
        self._loaded = False
        await self.load()
    
    def get_model_info(self) -> Dict[str, Any]:
        """Get model metadata."""
        return {
            "human_classifier_version": self._model_versions.get("human"),
            "anomaly_model_version": self._model_versions.get("anomaly"),
            "genome_model_version": self._model_versions.get("genome"),
            "loaded": self._loaded,
            "loaded_at": datetime.utcnow().isoformat() if self._loaded else None,
        }
    
    async def close(self):
        """Cleanup resources."""
        self._executor.shutdown(wait=False)
```

---

## 5. Protobuf/gRPC Definitions

### 5.1 ZK Prover Service (`proto/zkprover.proto`)

```protobuf
syntax = "proto3";

package vigilum.zkprover.v1;

option go_package = "github.com/vigilum/proto/zkprover/v1;zkproverv1";

// ZKProverService provides ZK proof generation and verification
service ZKProverService {
    // GenerateHumanProof generates a human behavior ZK proof
    rpc GenerateHumanProof(GenerateHumanProofRequest) returns (GenerateHumanProofResponse);
    
    // GenerateExploitProof generates an exploit proof for Red-Team DAO
    rpc GenerateExploitProof(GenerateExploitProofRequest) returns (GenerateExploitProofResponse);
    
    // VerifyProof verifies a ZK proof
    rpc VerifyProof(VerifyProofRequest) returns (VerifyProofResponse);
    
    // BatchVerifyProofs verifies multiple proofs
    rpc BatchVerifyProofs(BatchVerifyProofsRequest) returns (BatchVerifyProofsResponse);
    
    // GetCircuitInfo returns information about available circuits
    rpc GetCircuitInfo(GetCircuitInfoRequest) returns (GetCircuitInfoResponse);
    
    // HealthCheck returns service health status
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

// ========== Human Proof Messages ==========

message GenerateHumanProofRequest {
    // Behavioral features for proof generation
    BehavioralFeatures features = 1;
    
    // Challenge from server
    bytes challenge = 2;
    
    // User's private key (encrypted in transit)
    bytes private_key = 3;
    
    // Circuit ID to use (optional, defaults to latest)
    string circuit_id = 4;
}

message BehavioralFeatures {
    string wallet_address = 1;
    uint32 tx_count = 2;
    double avg_tx_interval = 3;
    double tx_interval_variance = 4;
    double gas_variance = 5;
    uint32 interaction_diversity = 6;
    uint32 unique_contracts = 7;
    uint64 timestamp = 8;
}

message GenerateHumanProofResponse {
    // Generated proof
    HumanProof proof = 1;
    
    // Generation time in milliseconds
    uint64 generation_time_ms = 2;
    
    // Error if generation failed
    Error error = 3;
}

message HumanProof {
    // Raw proof bytes
    bytes proof = 1;
    
    // Public inputs
    PublicInputs public_inputs = 2;
    
    // Circuit ID used
    string circuit_id = 3;
    
    // Prover ID
    string prover_id = 4;
    
    // Timestamp
    uint64 timestamp = 5;
    
    // Proof hash
    bytes proof_hash = 6;
}

message PublicInputs {
    string wallet_address = 1;
    bytes features_commitment = 2;
    uint64 timestamp = 3;
    bytes challenge_response = 4;
}

// ========== Exploit Proof Messages ==========

message GenerateExploitProofRequest {
    // Target contract address
    string target_contract = 1;
    
    // Target contract bytecode
    bytes bytecode = 2;
    
    // Exploit calldata
    bytes exploit_calldata = 3;
    
    // Pre-state root
    bytes pre_state_root = 4;
    
    // Post-state root
    bytes post_state_root = 5;
    
    // Vulnerability type
    VulnerabilityType vulnerability_type = 6;
}

enum VulnerabilityType {
    VULNERABILITY_TYPE_UNSPECIFIED = 0;
    VULNERABILITY_TYPE_REENTRANCY = 1;
    VULNERABILITY_TYPE_INTEGER_OVERFLOW = 2;
    VULNERABILITY_TYPE_ACCESS_CONTROL = 3;
    VULNERABILITY_TYPE_FLASH_LOAN = 4;
    VULNERABILITY_TYPE_PRICE_MANIPULATION = 5;
    VULNERABILITY_TYPE_STORAGE_COLLISION = 6;
    VULNERABILITY_TYPE_OTHER = 7;
}

message GenerateExploitProofResponse {
    // Generated proof
    bytes proof = 1;
    
    // Proof hash
    bytes proof_hash = 2;
    
    // Generation time in milliseconds
    uint64 generation_time_ms = 3;
    
    // Error if generation failed
    Error error = 4;
}

// ========== Verification Messages ==========

message VerifyProofRequest {
    // Proof to verify
    bytes proof = 1;
    
    // Public inputs
    bytes public_inputs = 2;
    
    // Circuit ID (optional)
    string circuit_id = 3;
}

message VerifyProofResponse {
    // Whether proof is valid
    bool valid = 1;
    
    // Verification time in milliseconds
    uint64 verification_time_ms = 2;
    
    // Error if verification failed
    Error error = 3;
}

message BatchVerifyProofsRequest {
    // Proofs to verify
    repeated VerifyProofRequest proofs = 1;
}

message BatchVerifyProofsResponse {
    // Results for each proof
    repeated VerifyProofResponse results = 1;
    
    // Total verification time
    uint64 total_time_ms = 2;
}

// ========== Circuit Info Messages ==========

message GetCircuitInfoRequest {
    // Circuit ID (optional, returns all if empty)
    string circuit_id = 1;
}

message GetCircuitInfoResponse {
    // Available circuits
    repeated CircuitInfo circuits = 1;
}

message CircuitInfo {
    // Circuit ID
    string circuit_id = 1;
    
    // Circuit name
    string name = 2;
    
    // Circuit version
    string version = 3;
    
    // Number of constraints
    uint64 num_constraints = 4;
    
    // Number of public inputs
    uint32 num_public_inputs = 5;
    
    // Proving key size in bytes
    uint64 proving_key_size = 6;
    
    // Verification key size in bytes
    uint64 verification_key_size = 7;
    
    // Backend (e.g., "barretenberg")
    string backend = 8;
}

// ========== Health Check Messages ==========

message HealthCheckRequest {}

message HealthCheckResponse {
    // Service status
    ServiceStatus status = 1;
    
    // Uptime in seconds
    uint64 uptime_seconds = 2;
    
    // Number of proofs generated
    uint64 proofs_generated = 3;
    
    // Number of proofs verified
    uint64 proofs_verified = 4;
    
    // Average generation time (ms)
    double avg_generation_time_ms = 5;
    
    // Average verification time (ms)
    double avg_verification_time_ms = 6;
}

enum ServiceStatus {
    SERVICE_STATUS_UNSPECIFIED = 0;
    SERVICE_STATUS_SERVING = 1;
    SERVICE_STATUS_NOT_SERVING = 2;
}

// ========== Common Messages ==========

message Error {
    // Error code
    ErrorCode code = 1;
    
    // Error message
    string message = 2;
    
    // Additional details
    map<string, string> details = 3;
}

enum ErrorCode {
    ERROR_CODE_UNSPECIFIED = 0;
    ERROR_CODE_INVALID_INPUT = 1;
    ERROR_CODE_PROOF_GENERATION_FAILED = 2;
    ERROR_CODE_VERIFICATION_FAILED = 3;
    ERROR_CODE_CIRCUIT_NOT_FOUND = 4;
    ERROR_CODE_TIMEOUT = 5;
    ERROR_CODE_INTERNAL = 6;
}
```

---

## 6. TypeScript SDK Objects

### 6.1 Types (`sdk/typescript/src/types/`)

```typescript
// types/index.ts
export * from './address';
export * from './proof';
export * from './signal';
export * from './genome';
export * from './api';

// types/address.ts
import { keccak256 } from 'ethers';

/**
 * Ethereum address type
 */
export class Address {
    private readonly bytes: Uint8Array;
    
    constructor(input: string | Uint8Array) {
        if (typeof input === 'string') {
            this.bytes = Address.fromHex(input);
        } else {
            if (input.length !== 20) {
                throw new Error(`Invalid address length: ${input.length}`);
            }
            this.bytes = input;
        }
    }
    
    private static fromHex(hex: string): Uint8Array {
        const cleaned = hex.startsWith('0x') ? hex.slice(2) : hex;
        if (cleaned.length !== 40) {
            throw new Error(`Invalid address length: ${cleaned.length}`);
        }
        if (!/^[0-9a-fA-F]+$/.test(cleaned)) {
            throw new Error('Invalid hex characters');
        }
        return new Uint8Array(
            cleaned.match(/.{2}/g)!.map(byte => parseInt(byte, 16))
        );
    }
    
    /**
     * Get checksummed address string (EIP-55)
     */
    toChecksum(): string {
        const hex = Array.from(this.bytes)
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');
        const hash = keccak256(new TextEncoder().encode(hex)).slice(2);
        
        let result = '0x';
        for (let i = 0; i < 40; i++) {
            const char = hex[i];
            if (parseInt(hash[i], 16) >= 8) {
                result += char.toUpperCase();
            } else {
                result += char.toLowerCase();
            }
        }
        return result;
    }
    
    /**
     * Get lowercase hex string
     */
    toLowercase(): string {
        return '0x' + Array.from(this.bytes)
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');
    }
    
    /**
     * Get raw bytes
     */
    toBytes(): Uint8Array {
        return this.bytes;
    }
    
    /**
     * Check equality
     */
    equals(other: Address): boolean {
        return this.toLowercase() === other.toLowercase();
    }
    
    /**
     * Zero address
     */
    static ZERO = new Address(new Uint8Array(20));
    
    /**
     * Validate address string
     */
    static isValid(address: string): boolean {
        try {
            new Address(address);
            return true;
        } catch {
            return false;
        }
    }
}

// types/proof.ts
export interface HumanProof {
    /** Raw proof bytes (base64 encoded) */
    proof: string;
    /** Public inputs */
    publicInputs: PublicInputs;
    /** Circuit ID */
    circuitId: string;
    /** Prover ID */
    proverId: string;
    /** Timestamp (unix seconds) */
    timestamp: number;
    /** Proof hash (hex) */
    proofHash: string;
}

export interface PublicInputs {
    /** Wallet address */
    walletAddress: string;
    /** Features commitment (hex) */
    featuresCommitment: string;
    /** Timestamp */
    timestamp: number;
    /** Challenge response (hex) */
    challengeResponse: string;
}

export interface VerificationResult {
    /** Whether proof is valid */
    verified: boolean;
    /** Proof hash */
    proofHash: string;
    /** Risk score (0-1) */
    riskScore: number;
    /** Transaction hash if on-chain */
    txHash?: string;
    /** Error message if failed */
    error?: string;
    /** Expiry timestamp */
    expiresAt: Date;
}

export interface Challenge {
    /** Challenge ID */
    challengeId: string;
    /** Challenge data (hex) */
    challenge: string;
    /** Expiry timestamp */
    expiresAt: Date;
}

// types/signal.ts
export enum SignalType {
    ExploitDetected = 'exploit_detected',
    KeyLeaked = 'key_leaked',
    AnomalyDetected = 'anomaly_detected',
    PhishingAttempt = 'phishing_attempt',
    MaliciousContract = 'malicious_contract',
    FlashLoanAttack = 'flash_loan_attack',
    Reentrancy = 'reentrancy',
    PriceManipulation = 'price_manipulation',
}

export enum RiskLevel {
    Low = 'low',
    Medium = 'medium',
    High = 'high',
    Critical = 'critical',
}

export interface ThreatSignal {
    /** Signal ID */
    id: string;
    /** Entity address */
    entityAddress: string;
    /** Signal type */
    signalType: SignalType;
    /** Risk score (0-100) */
    riskScore: number;
    /** Confidence (0-1) */
    confidence: number;
    /** Signal source */
    source: string;
    /** Additional metadata */
    metadata?: Record<string, unknown>;
    /** Timestamp */
    timestamp: Date;
    /** Transaction hash if published */
    txHash?: string;
}

export interface RiskScoreResponse {
    /** Address */
    address: string;
    /** Risk score (0-1) */
    riskScore: number;
    /** Risk level */
    riskLevel: RiskLevel;
    /** Contributing signals */
    signals: SignalSummary[];
    /** Last updated */
    updatedAt: Date;
}

export interface SignalSummary {
    /** Signal type */
    type: SignalType;
    /** Source */
    source: string;
    /** Confidence */
    confidence: number;
    /** Timestamp */
    timestamp: Date;
}

// types/genome.ts
export enum GenomeLabel {
    KnownExploit = 'known_exploit',
    Suspicious = 'suspicious',
    Benign = 'benign',
    Unknown = 'unknown',
}

export interface Genome {
    /** Genome hash (hex) */
    hash: string;
    /** IPFS hash */
    ipfsHash: string;
    /** Contract address */
    contractAddress?: string;
    /** Classification label */
    label: GenomeLabel;
    /** Features */
    features: GenomeFeatures;
    /** Registration timestamp */
    registeredAt?: Date;
    /** On-chain status */
    onChain: boolean;
}

export interface GenomeFeatures {
    /** Bytecode length */
    bytecodeLength: number;
    /** Unique opcodes */
    uniqueOpcodes: number;
    /** Jump density */
    jumpDensity: number;
    /** External call count */
    externalCallCount: number;
    /** Has SELFDESTRUCT */
    hasSelfdestruct: boolean;
    /** Has DELEGATECALL */
    hasDelegatecall: boolean;
    /** Opcode histogram */
    opcodeHistogram: Record<string, number>;
}

export interface GenomeSimilarity {
    /** Similar genome */
    genome: Genome;
    /** Similarity score (0-1) */
    similarityScore: number;
}

export interface AnalysisJob {
    /** Analysis ID */
    analysisId: string;
    /** Status */
    status: 'queued' | 'processing' | 'completed' | 'failed';
    /** Estimated completion */
    estimatedCompletion?: Date;
}

export interface AnalysisResult {
    /** Analysis ID */
    analysisId: string;
    /** Status */
    status: 'completed' | 'failed';
    /** Genome (if completed) */
    genome?: Genome;
    /** Similar genomes */
    similarGenomes?: GenomeSimilarity[];
    /** Error (if failed) */
    error?: string;
}

// types/api.ts
export interface ApiResponse<T> {
    success: boolean;
    data?: T;
    error?: ApiError;
    timestamp: Date;
}

export interface ApiError {
    code: string;
    message: string;
    details?: Record<string, string>;
}

export interface PaginatedResponse<T> {
    items: T[];
    total: number;
    page: number;
    pageSize: number;
    hasNext: boolean;
}

export interface ClientConfig {
    /** Base URL */
    baseUrl: string;
    /** API key (optional) */
    apiKey?: string;
    /** Request timeout in ms */
    timeout?: number;
    /** Retry configuration */
    retry?: RetryConfig;
}

export interface RetryConfig {
    /** Max retries */
    maxRetries: number;
    /** Initial delay in ms */
    initialDelay: number;
    /** Max delay in ms */
    maxDelay: number;
    /** Backoff multiplier */
    multiplier: number;
}
```

---

### 6.2 Client (`sdk/typescript/src/client/`)

```typescript
// client/index.ts
export { VigilumClient } from './vigilum';
export { FirewallClient } from './firewall';
export { OracleClient } from './oracle';
export { GenomeClient } from './genome';

// client/base.ts
import type { ClientConfig, ApiResponse, ApiError, RetryConfig } from '../types';

const DEFAULT_RETRY: RetryConfig = {
    maxRetries: 3,
    initialDelay: 1000,
    maxDelay: 10000,
    multiplier: 2,
};

export abstract class BaseClient {
    protected readonly baseUrl: string;
    protected readonly apiKey?: string;
    protected readonly timeout: number;
    protected readonly retry: RetryConfig;
    
    constructor(config: ClientConfig) {
        this.baseUrl = config.baseUrl.replace(/\/$/, '');
        this.apiKey = config.apiKey;
        this.timeout = config.timeout ?? 30000;
        this.retry = config.retry ?? DEFAULT_RETRY;
    }
    
    protected async request<T>(
        method: string,
        path: string,
        body?: unknown,
    ): Promise<T> {
        let lastError: Error | undefined;
        
        for (let attempt = 0; attempt <= this.retry.maxRetries; attempt++) {
            try {
                const response = await this.doRequest<T>(method, path, body);
                return response;
            } catch (error) {
                lastError = error as Error;
                
                // Don't retry on 4xx errors
                if (error instanceof ApiClientError && error.statusCode < 500) {
                    throw error;
                }
                
                if (attempt < this.retry.maxRetries) {
                    const delay = Math.min(
                        this.retry.initialDelay * Math.pow(this.retry.multiplier, attempt),
                        this.retry.maxDelay,
                    );
                    await this.sleep(delay);
                }
            }
        }
        
        throw lastError;
    }
    
    private async doRequest<T>(
        method: string,
        path: string,
        body?: unknown,
    ): Promise<T> {
        const url = `${this.baseUrl}${path}`;
        const headers: HeadersInit = {
            'Content-Type': 'application/json',
        };
        
        if (this.apiKey) {
            headers['X-API-Key'] = this.apiKey;
        }
        
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);
        
        try {
            const response = await fetch(url, {
                method,
                headers,
                body: body ? JSON.stringify(body) : undefined,
                signal: controller.signal,
            });
            
            const data: ApiResponse<T> = await response.json();
            
            if (!response.ok || !data.success) {
                throw new ApiClientError(
                    data.error?.message ?? 'Request failed',
                    data.error?.code ?? 'UNKNOWN_ERROR',
                    response.status,
                    data.error?.details,
                );
            }
            
            return data.data!;
        } finally {
            clearTimeout(timeoutId);
        }
    }
    
    private sleep(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

export class ApiClientError extends Error {
    constructor(
        message: string,
        public readonly code: string,
        public readonly statusCode: number,
        public readonly details?: Record<string, string>,
    ) {
        super(message);
        this.name = 'ApiClientError';
    }
}

// client/firewall.ts
import { BaseClient } from './base';
import type {
    ClientConfig,
    HumanProof,
    VerificationResult,
    Challenge,
    RiskScoreResponse,
} from '../types';

export class FirewallClient extends BaseClient {
    constructor(config: ClientConfig) {
        super(config);
    }
    
    /**
     * Verify a human behavior ZK proof
     */
    async verifyProof(proof: HumanProof): Promise<VerificationResult> {
        return this.request<VerificationResult>('POST', '/firewall/verify-proof', {
            proof: proof.proof,
            public_inputs: proof.publicInputs,
        });
    }
    
    /**
     * Get a challenge for proof generation
     */
    async getChallenge(): Promise<Challenge> {
        const response = await this.request<{
            challenge_id: string;
            challenge: string;
            expires_at: string;
        }>('GET', '/firewall/challenge');
        
        return {
            challengeId: response.challenge_id,
            challenge: response.challenge,
            expiresAt: new Date(response.expires_at),
        };
    }
    
    /**
     * Get risk score for an address
     */
    async getRiskScore(address: string): Promise<RiskScoreResponse> {
        const response = await this.request<{
            address: string;
            risk_score: number;
            risk_level: string;
            signals: Array<{
                type: string;
                source: string;
                confidence: number;
                timestamp: string;
            }>;
            updated_at: string;
        }>('GET', `/firewall/risk/${address}`);
        
        return {
            address: response.address,
            riskScore: response.risk_score,
            riskLevel: response.risk_level as any,
            signals: response.signals.map(s => ({
                type: s.type as any,
                source: s.source,
                confidence: s.confidence,
                timestamp: new Date(s.timestamp),
            })),
            updatedAt: new Date(response.updated_at),
        };
    }
    
    /**
     * Check if address is verified
     */
    async isVerified(address: string): Promise<boolean> {
        try {
            const score = await this.getRiskScore(address);
            return score.riskScore < 0.5;
        } catch {
            return false;
        }
    }
}

// client/oracle.ts
import { BaseClient } from './base';
import type {
    ClientConfig,
    ThreatSignal,
    SignalType,
} from '../types';

export interface Subscription {
    subscriptionId: string;
    active: boolean;
}

export interface SubscribeOptions {
    webhookUrl: string;
    addresses: string[];
    signalTypes?: SignalType[];
}

export class OracleClient extends BaseClient {
    constructor(config: ClientConfig) {
        super(config);
    }
    
    /**
     * Get all signals for an address
     */
    async getSignals(address: string): Promise<ThreatSignal[]> {
        const response = await this.request<{
            address: string;
            signals: Array<{
                id: string;
                entity_address: string;
                signal_type: string;
                risk_score: number;
                confidence: number;
                source: string;
                metadata?: Record<string, unknown>;
                timestamp: string;
                tx_hash?: string;
            }>;
        }>('GET', `/oracle/signals/${address}`);
        
        return response.signals.map(s => ({
            id: s.id,
            entityAddress: s.entity_address,
            signalType: s.signal_type as SignalType,
            riskScore: s.risk_score,
            confidence: s.confidence,
            source: s.source,
            metadata: s.metadata,
            timestamp: new Date(s.timestamp),
            txHash: s.tx_hash,
        }));
    }
    
    /**
     * Subscribe to signals via webhook
     */
    async subscribe(options: SubscribeOptions): Promise<Subscription> {
        const response = await this.request<{
            subscription_id: string;
            active: boolean;
        }>('POST', '/oracle/subscribe', {
            webhook_url: options.webhookUrl,
            addresses: options.addresses,
            signal_types: options.signalTypes,
        });
        
        return {
            subscriptionId: response.subscription_id,
            active: response.active,
        };
    }
    
    /**
     * Unsubscribe from signals
     */
    async unsubscribe(subscriptionId: string): Promise<void> {
        await this.request('DELETE', `/oracle/subscribe/${subscriptionId}`);
    }
    
    /**
     * Check if address is high risk
     */
    async isHighRisk(address: string, threshold: number = 70): Promise<boolean> {
        const signals = await this.getSignals(address);
        return signals.some(s => s.riskScore >= threshold);
    }
}

// client/genome.ts
import { BaseClient } from './base';
import type {
    ClientConfig,
    Genome,
    GenomeSimilarity,
    AnalysisJob,
    AnalysisResult,
    GenomeLabel,
} from '../types';

export class GenomeClient extends BaseClient {
    constructor(config: ClientConfig) {
        super(config);
    }
    
    /**
     * Analyze a contract
     */
    async analyze(
        contractAddress: string,
        priority: 'normal' | 'high' | 'urgent' = 'normal',
    ): Promise<AnalysisJob> {
        const response = await this.request<{
            analysis_id: string;
            status: string;
            estimated_completion?: string;
        }>('POST', '/genome/analyze', {
            contract_address: contractAddress,
            priority,
        });
        
        return {
            analysisId: response.analysis_id,
            status: response.status as any,
            estimatedCompletion: response.estimated_completion
                ? new Date(response.estimated_completion)
                : undefined,
        };
    }
    
    /**
     * Get analysis status
     */
    async getAnalysisStatus(analysisId: string): Promise<AnalysisResult> {
        const response = await this.request<{
            analysis_id: string;
            status: string;
            genome?: any;
            similar_genomes?: any[];
            error?: string;
        }>('GET', `/genome/status/${analysisId}`);
        
        return {
            analysisId: response.analysis_id,
            status: response.status as any,
            genome: response.genome ? this.parseGenome(response.genome) : undefined,
            similarGenomes: response.similar_genomes?.map(sg => ({
                genome: this.parseGenome(sg.genome),
                similarityScore: sg.similarity_score,
            })),
            error: response.error,
        };
    }
    
    /**
     * Get genome by hash
     */
    async getGenome(genomeHash: string): Promise<Genome> {
        const response = await this.request<any>('GET', `/genome/${genomeHash}`);
        return this.parseGenome(response);
    }
    
    /**
     * Find similar genomes
     */
    async findSimilar(
        genomeHash: string,
        threshold: number = 0.8,
    ): Promise<GenomeSimilarity[]> {
        const response = await this.request<any[]>(
            'GET',
            `/genome/${genomeHash}/similar?threshold=${threshold}`,
        );
        
        return response.map(sg => ({
            genome: this.parseGenome(sg.genome),
            similarityScore: sg.similarity_score,
        }));
    }
    
    /**
     * Wait for analysis to complete
     */
    async waitForAnalysis(
        analysisId: string,
        pollInterval: number = 2000,
        timeout: number = 120000,
    ): Promise<AnalysisResult> {
        const startTime = Date.now();
        
        while (Date.now() - startTime < timeout) {
            const result = await this.getAnalysisStatus(analysisId);
            
            if (result.status === 'completed' || result.status === 'failed') {
                return result;
            }
            
            await new Promise(resolve => setTimeout(resolve, pollInterval));
        }
        
        throw new Error(`Analysis timeout after ${timeout}ms`);
    }
    
    private parseGenome(data: any): Genome {
        return {
            hash: data.hash,
            ipfsHash: data.ipfs_hash,
            contractAddress: data.contract_address,
            label: data.label as GenomeLabel,
            features: {
                bytecodeLength: data.features?.bytecode_length ?? 0,
                uniqueOpcodes: data.features?.unique_opcodes ?? 0,
                jumpDensity: data.features?.jump_density ?? 0,
                externalCallCount: data.features?.external_call_count ?? 0,
                hasSelfdestruct: data.features?.has_selfdestruct ?? false,
                hasDelegatecall: data.features?.has_delegatecall ?? false,
                opcodeHistogram: data.features?.opcode_histogram ?? {},
            },
            registeredAt: data.registered_at ? new Date(data.registered_at) : undefined,
            onChain: data.on_chain ?? false,
        };
    }
}

// client/vigilum.ts
import type { ClientConfig } from '../types';
import { FirewallClient } from './firewall';
import { OracleClient } from './oracle';
import { GenomeClient } from './genome';

/**
 * Main VIGILUM SDK client
 */
export class VigilumClient {
    /** Identity Firewall client */
    public readonly firewall: FirewallClient;
    
    /** Threat Oracle client */
    public readonly oracle: OracleClient;
    
    /** Genome Analyzer client */
    public readonly genome: GenomeClient;
    
    constructor(config: ClientConfig) {
        this.firewall = new FirewallClient(config);
        this.oracle = new OracleClient(config);
        this.genome = new GenomeClient(config);
    }
    
    /**
     * Create client with API key
     */
    static withApiKey(baseUrl: string, apiKey: string): VigilumClient {
        return new VigilumClient({ baseUrl, apiKey });
    }
    
    /**
     * Create client for mainnet
     */
    static mainnet(apiKey?: string): VigilumClient {
        return new VigilumClient({
            baseUrl: 'https://api.vigilum.io/v1',
            apiKey,
        });
    }
    
    /**
     * Create client for testnet
     */
    static testnet(apiKey?: string): VigilumClient {
        return new VigilumClient({
            baseUrl: 'https://testnet-api.vigilum.io/v1',
            apiKey,
        });
    }
}
```

---

## 7. Database Migrations

### 7.1 Postgres Migrations (`backend/migrations/`)

```sql
-- 001_initial_schema.up.sql

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_address VARCHAR(42) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMPTZ,
    risk_score DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    CONSTRAINT valid_address CHECK (wallet_address ~ '^0x[a-fA-F0-9]{40}$')
);

CREATE INDEX idx_users_wallet ON users(wallet_address);
CREATE INDEX idx_users_risk ON users(risk_score) WHERE risk_score >= 0.7;

-- Human proofs table
CREATE TABLE human_proofs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    proof_hash BYTEA NOT NULL UNIQUE,
    proof_data JSONB,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ,
    verifier_address VARCHAR(42),
    tx_hash VARCHAR(66),
    CONSTRAINT valid_tx_hash CHECK (tx_hash IS NULL OR tx_hash ~ '^0x[a-fA-F0-9]{64}$')
);

CREATE INDEX idx_proofs_user ON human_proofs(user_id);
CREATE INDEX idx_proofs_hash ON human_proofs(proof_hash);
CREATE INDEX idx_proofs_verified ON human_proofs(verified) WHERE verified = FALSE;

-- Threat signals table
CREATE TABLE threat_signals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_address VARCHAR(42) NOT NULL,
    signal_type VARCHAR(50) NOT NULL,
    risk_score INTEGER NOT NULL CHECK (risk_score >= 0 AND risk_score <= 100),
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    source VARCHAR(100) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    tx_hash VARCHAR(66),
    CONSTRAINT valid_entity CHECK (entity_address ~ '^0x[a-fA-F0-9]{40}$')
);

CREATE INDEX idx_signals_entity ON threat_signals(entity_address);
CREATE INDEX idx_signals_type ON threat_signals(signal_type);
CREATE INDEX idx_signals_risk ON threat_signals(risk_score) WHERE risk_score >= 70;
CREATE INDEX idx_signals_unpublished ON threat_signals(created_at) WHERE published_at IS NULL;

-- Genomes table
CREATE TABLE genomes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    genome_hash BYTEA NOT NULL UNIQUE,
    ipfs_hash VARCHAR(100) NOT NULL,
    contract_address VARCHAR(42),
    label VARCHAR(20) NOT NULL DEFAULT 'unknown',
    features JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    registered_on_chain BOOLEAN NOT NULL DEFAULT FALSE,
    tx_hash VARCHAR(66),
    CONSTRAINT valid_label CHECK (label IN ('known_exploit', 'suspicious', 'benign', 'unknown'))
);

CREATE INDEX idx_genomes_hash ON genomes(genome_hash);
CREATE INDEX idx_genomes_contract ON genomes(contract_address) WHERE contract_address IS NOT NULL;
CREATE INDEX idx_genomes_label ON genomes(label);

-- Exploit submissions table
CREATE TABLE exploit_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    researcher_address VARCHAR(42) NOT NULL,
    target_contract VARCHAR(42) NOT NULL,
    proof_hash BYTEA NOT NULL,
    genome_id UUID REFERENCES genomes(id),
    description TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL,
    bounty_amount BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    tx_hash VARCHAR(66),
    votes_for INTEGER NOT NULL DEFAULT 0,
    votes_against INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT valid_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'verifying', 'verified', 'rejected', 'paid', 'disputed'))
);

CREATE INDEX idx_submissions_researcher ON exploit_submissions(researcher_address);
CREATE INDEX idx_submissions_status ON exploit_submissions(status);
CREATE INDEX idx_submissions_pending ON exploit_submissions(created_at) WHERE status = 'pending';

-- API keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash BYTEA NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier VARCHAR(20) NOT NULL DEFAULT 'free',
    rate_limit INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    last_used TIMESTAMPTZ,
    CONSTRAINT valid_tier CHECK (tier IN ('free', 'paid', 'enterprise'))
);

CREATE INDEX idx_apikeys_hash ON api_keys(key_hash);
CREATE INDEX idx_apikeys_user ON api_keys(user_id);
CREATE INDEX idx_apikeys_active ON api_keys(created_at) WHERE revoked = FALSE;

-- Events table (for event sourcing)
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(100) NOT NULL,
    version INTEGER NOT NULL,
    data JSONB NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    trace_id VARCHAR(100)
);

CREATE INDEX idx_events_aggregate ON events(aggregate_id, version);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_timestamp ON events(timestamp);
```

---

**[END OF OBJECT DESIGN DOCUMENT]**

**Total Coverage:**
- **Go**: Config, Models, Events, Repositories, Services, Handlers, Middleware, Integration
- **Rust**: Types, ZK Prover/Verifier, SDK Clients, WASM Bindings
- **Solidity**: IdentityFirewall, ThreatOracle, MalwareGenomeDB, RedTeamDAO
- **Python**: Features, Predictions, ETL Pipeline, Training, Inference Service
- **Protobuf**: ZK Prover gRPC Service
- **TypeScript**: Types, Clients (Firewall, Oracle, Genome, Main)
- **SQL**: Initial Postgres schema with indexes

**Lines: ~4000+**
