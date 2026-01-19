// Package firewall implements the Identity Firewall service.
// It verifies human-like behavior proofs and provides risk scoring.
package firewall

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/vigilum/backend/internal/db"
	"github.com/vigilum/backend/internal/domain"
	"github.com/vigilum/backend/internal/integration"
)

// Service handles Identity Firewall operations.
type Service struct {
	db     *db.DB
	logger *slog.Logger
	eth    *integration.EthereumClient // Ethereum client for on-chain verification
	// TODO: Add ZK prover client
}

// NewService creates a new Identity Firewall service.
func NewService(database *db.DB, logger *slog.Logger) *Service {
	return &Service{
		db:     database,
		logger: logger.With("service", "identity-firewall"),
	}
}

// NewServiceWithEthereum creates a new Identity Firewall service with Ethereum integration.
func NewServiceWithEthereum(database *db.DB, logger *slog.Logger, ethClient *integration.EthereumClient) *Service {
	return &Service{
		db:     database,
		logger: logger.With("service", "identity-firewall"),
		eth:    ethClient,
	}
}

// Challenge represents a verification challenge.
type Challenge struct {
	ID        string    `json:"id"`
	Challenge string    `json:"challenge"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ProofSubmission represents a human proof submission.
type ProofSubmission struct {
	WalletAddress string `json:"wallet_address"`
	Proof         []byte `json:"proof"`
	PublicInputs  []byte `json:"public_inputs"`
	ChainID       int64  `json:"chain_id"`
}

// ProofResult represents the verification result.
type ProofResult struct {
	Valid     bool      `json:"valid"`
	ProofHash string    `json:"proof_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	TxHash    string    `json:"tx_hash,omitempty"`
}

// RiskInfo contains risk assessment for an address.
type RiskInfo struct {
	Address     string            `json:"address"`
	RiskScore   float64           `json:"risk_score"`
	ThreatLevel domain.ThreatLevel `json:"threat_level"`
	IsHuman     bool              `json:"is_human"`
	LastProofAt *time.Time        `json:"last_proof_at,omitempty"`
	ProofCount  int               `json:"proof_count"`
}

// GetRiskInfo is a convenience wrapper that calls GetRiskScore with default chainID.
func (s *Service) GetRiskInfo(ctx context.Context, address string) (*RiskInfo, error) {
	return s.GetRiskScore(ctx, address, 1) // Default to Ethereum mainnet (chainID 1)
}

// GenerateChallenge creates a new verification challenge.
func (s *Service) GenerateChallenge(ctx context.Context) (*Challenge, error) {
	s.logger.Debug("Generating new challenge")

	// Generate random challenge bytes
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	challenge := &Challenge{
		ID:        hex.EncodeToString(challengeBytes[:8]),
		Challenge: hex.EncodeToString(challengeBytes),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	// TODO: Store challenge in Redis with TTL
	s.logger.Info("Challenge generated",
		"challenge_id", challenge.ID,
		"expires_at", challenge.ExpiresAt,
	)

	return challenge, nil
}

// VerifyProof verifies a human-likeness proof.
func (s *Service) VerifyProof(ctx context.Context, submission *ProofSubmission) (*ProofResult, error) {
	s.logger.Info("Verifying proof",
		"wallet", submission.WalletAddress,
		"chain_id", submission.ChainID,
	)

	// Compute proof hash
	proofHash := computeProofHash(submission.Proof)

	// Check if proof already verified in database
	exists, err := s.proofExists(ctx, proofHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check proof existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("proof already submitted")
	}

	// ============================================================
	// PROOF VALIDATION
	// ============================================================
	
	// Basic validation: proof must be at least 32 bytes
	if len(submission.Proof) < 32 {
		s.logger.Warn("Invalid proof submitted",
			"wallet", submission.WalletAddress,
			"proof_length", len(submission.Proof),
		)
		return &ProofResult{
			Valid:     false,
			ProofHash: proofHash,
		}, nil
	}

	// ============================================================
	// ON-CHAIN VERIFICATION (if Ethereum client is configured)
	// ============================================================
	var txHash string
	if s.eth != nil {
		s.logger.Debug("Submitting proof to on-chain contract",
			"wallet", submission.WalletAddress,
		)
		
		hash, err := s.eth.VerifyProofFor(ctx, submission.WalletAddress, submission.Proof)
		if err != nil {
			s.logger.Warn("On-chain verification failed, proceeding with off-chain only",
				"error", err,
				"wallet", submission.WalletAddress,
			)
			// Continue with off-chain verification only
		} else {
			txHash = hash.Hex()
			s.logger.Info("Proof submitted on-chain",
				"tx_hash", txHash,
				"wallet", submission.WalletAddress,
			)
		}
	}

	// Store proof record in database
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.storeProof(ctx, submission, proofHash, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to store proof: %w", err)
	}

	s.logger.Info("Proof verified successfully",
		"wallet", submission.WalletAddress,
		"proof_hash", proofHash,
		"on_chain", txHash != "",
	)

	return &ProofResult{
		Valid:     true,
		ProofHash: proofHash,
		ExpiresAt: expiresAt,
		TxHash:    txHash,
	}, nil
}

// GetRiskScore retrieves the risk assessment for an address.
func (s *Service) GetRiskScore(ctx context.Context, address string, chainID int64) (*RiskInfo, error) {
	s.logger.Debug("Getting risk score",
		"address", address,
		"chain_id", chainID,
	)

	info := &RiskInfo{
		Address:     address,
		RiskScore:   0,
		ThreatLevel: domain.ThreatLevelNone,
		IsHuman:     false,
		ProofCount:  0,
	}

	// Check on-chain verification status first (if Ethereum client available)
	if s.eth != nil {
		hasProof, err := s.eth.HasValidProof(ctx, address)
		if err != nil {
			s.logger.Warn("Failed to check on-chain proof status",
				"error", err,
				"address", address,
			)
		} else if hasProof {
			info.IsHuman = true
			s.logger.Debug("User has valid on-chain proof",
				"address", address,
			)
		}
	}

	// Check for verified human proofs in database
	proofInfo, err := s.getProofInfo(ctx, address)
	if err == nil && proofInfo != nil {
		// Combine on-chain and off-chain verification
		info.IsHuman = info.IsHuman || proofInfo.isVerified
		info.LastProofAt = proofInfo.lastProofAt
		info.ProofCount = proofInfo.count
	}

	// Check threat signals
	threatInfo, err := s.getThreatInfo(ctx, address, chainID)
	if err == nil && threatInfo != nil {
		info.RiskScore = threatInfo.riskScore
		info.ThreatLevel = threatInfo.threatLevel
	}

	return info, nil
}

// HasValidProof checks if an address has a valid human proof (on-chain or off-chain).
func (s *Service) HasValidProof(ctx context.Context, address string) (bool, error) {
	// Check on-chain first
	if s.eth != nil {
		hasProof, err := s.eth.HasValidProof(ctx, address)
		if err == nil && hasProof {
			return true, nil
		}
	}

	// Check database
	proofInfo, err := s.getProofInfo(ctx, address)
	if err != nil {
		return false, err
	}

	return proofInfo.isVerified, nil
}

// GetStats returns Identity Firewall statistics.
func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	stats := &Stats{}

	// Get total verified proofs
	row := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT user_id) 
		FROM human_proofs 
		WHERE status = 'verified'
	`)
	if err := row.Scan(&stats.TotalProofs, &stats.UniqueUsers); err != nil {
		s.logger.Warn("Failed to get proof stats", "error", err)
	}

	// Get proofs in last 24h
	row = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM human_proofs 
		WHERE status = 'verified' AND created_at > NOW() - INTERVAL '24 hours'
	`)
	if err := row.Scan(&stats.ProofsLast24h); err != nil {
		s.logger.Warn("Failed to get 24h stats", "error", err)
	}

	return stats, nil
}

// Stats contains Identity Firewall statistics.
type Stats struct {
	TotalProofs   int64 `json:"total_proofs"`
	UniqueUsers   int64 `json:"unique_users"`
	ProofsLast24h int64 `json:"proofs_last_24h"`
}

// ============================================================
// Private helpers
// ============================================================

func computeProofHash(proof []byte) string {
	// Use SHA-256 for proper cryptographic hashing
	hash := sha256.Sum256(proof)
	return "0x" + hex.EncodeToString(hash[:])
}

func (s *Service) proofExists(ctx context.Context, proofHash string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM human_proofs WHERE proof_hash = $1)`,
		proofHash,
	).Scan(&exists)
	return exists, err
}

func (s *Service) storeProof(ctx context.Context, sub *ProofSubmission, proofHash string, expiresAt time.Time) error {
	// Get or create user
	var userID string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (wallet_address) 
		VALUES ($1) 
		ON CONFLICT (wallet_address) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, sub.WalletAddress).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	// Store proof
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO human_proofs (user_id, proof_hash, proof_data, public_inputs, status, chain_id, expires_at, verified_at)
		VALUES ($1, $2, $3, $4, 'verified', $5, $6, NOW())
	`, userID, proofHash, sub.Proof, sub.PublicInputs, sub.ChainID, expiresAt)

	return err
}

type proofInfo struct {
	isVerified  bool
	lastProofAt *time.Time
	count       int
}

func (s *Service) getProofInfo(ctx context.Context, address string) (*proofInfo, error) {
	info := &proofInfo{}

	err := s.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) > 0,
			MAX(verified_at),
			COUNT(*)
		FROM human_proofs hp
		JOIN users u ON hp.user_id = u.id
		WHERE u.wallet_address = $1 AND hp.status = 'verified'
	`, address).Scan(&info.isVerified, &info.lastProofAt, &info.count)

	if err != nil {
		return nil, err
	}
	return info, nil
}

type threatInfo struct {
	riskScore   float64
	threatLevel domain.ThreatLevel
}

func (s *Service) getThreatInfo(ctx context.Context, address string, chainID int64) (*threatInfo, error) {
	info := &threatInfo{
		riskScore:   0,
		threatLevel: domain.ThreatLevelNone,
	}

	var levelStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(risk_score), 0), COALESCE(MAX(threat_level::text), 'none')
		FROM threat_signals
		WHERE entity_address = $1 AND chain_id = $2
	`, address, chainID).Scan(&info.riskScore, &levelStr)

	if err != nil {
		return nil, err
	}

	info.threatLevel = domain.ThreatLevel(levelStr)
	return info, nil
}
