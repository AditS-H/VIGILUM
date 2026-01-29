// Package proof provides proof registry integration with on-chain VigilumRegistry contract
package proof

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vigilum/backend/internal/integration"
)

// RegistryClient wraps VigilumRegistry contract interactions
type RegistryClient struct {
	ethClient *integration.EthereumClient
	logger    *slog.Logger
	enabled   bool
}

// NewRegistryClient creates a new registry client
func NewRegistryClient(ethClient *integration.EthereumClient, logger *slog.Logger) *RegistryClient {
	enabled := ethClient != nil
	return &RegistryClient{
		ethClient: ethClient,
		logger:    logger,
		enabled:   enabled,
	}
}

// RegisterProof registers a verified proof on-chain
func (rc *RegistryClient) RegisterProof(ctx context.Context, userAddr, contractAddr common.Address, riskScore uint32) error {
	if !rc.enabled {
		rc.logger.Debug("Registry client not enabled, skipping on-chain registration")
		return nil
	}

	rc.logger.Info("Registering proof on-chain",
		"user", userAddr.Hex(),
		"contract", contractAddr.Hex(),
		"risk_score", riskScore,
	)

	// Convert risk score from 0-100 to basis points (0-10000)
	basisPoints := uint64(riskScore) * 100

	// Generate bytecode hash (for demo, use contract address as hash)
	// In production, fetch actual bytecode and hash it
	bytecodeHash := crypto.Keccak256Hash(contractAddr.Bytes())

	// Call registerContract on VigilumRegistry
	err := rc.ethClient.RegisterContract(ctx, contractAddr, bytecodeHash, big.NewInt(int64(basisPoints)))
	if err != nil {
		rc.logger.Error("Failed to register contract on-chain", "error", err)
		return fmt.Errorf("failed to register contract: %w", err)
	}

	rc.logger.Info("Contract registered successfully",
		"contract", contractAddr.Hex(),
		"bytecode_hash", bytecodeHash.Hex(),
		"risk_score_bp", basisPoints,
	)

	return nil
}

// UpdateRiskScore updates the risk score for a contract on-chain
func (rc *RegistryClient) UpdateRiskScore(ctx context.Context, contractAddr common.Address, newScore uint32, vulnCount uint32) error {
	if !rc.enabled {
		rc.logger.Debug("Registry client not enabled, skipping risk score update")
		return nil
	}

	rc.logger.Info("Updating risk score on-chain",
		"contract", contractAddr.Hex(),
		"new_score", newScore,
		"vuln_count", vulnCount,
	)

	// Convert to basis points
	basisPoints := uint64(newScore) * 100

	// Call updateRiskScore on VigilumRegistry
	err := rc.ethClient.UpdateRiskScore(ctx, contractAddr, big.NewInt(int64(basisPoints)), vulnCount)
	if err != nil {
		rc.logger.Error("Failed to update risk score on-chain", "error", err)
		return fmt.Errorf("failed to update risk score: %w", err)
	}

	rc.logger.Info("Risk score updated successfully",
		"contract", contractAddr.Hex(),
		"risk_score_bp", basisPoints,
		"vuln_count", vulnCount,
	)

	return nil
}

// GetRiskScore retrieves the current risk score from the contract
func (rc *RegistryClient) GetRiskScore(ctx context.Context, contractAddr common.Address) (uint32, error) {
	if !rc.enabled {
		return 0, fmt.Errorf("registry client not enabled")
	}

	score, err := rc.ethClient.GetRiskScore(ctx, contractAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to get risk score: %w", err)
	}

	// Convert from basis points to 0-100 scale
	return uint32(score.Uint64() / 100), nil
}

// IsBlacklisted checks if a contract is blacklisted
func (rc *RegistryClient) IsBlacklisted(ctx context.Context, contractAddr common.Address) (bool, error) {
	if !rc.enabled {
		return false, nil
	}

	return rc.ethClient.IsBlacklisted(ctx, contractAddr)
}

// RegisterContractBatch registers multiple contracts in a single transaction (future optimization)
func (rc *RegistryClient) RegisterContractBatch(ctx context.Context, contracts []common.Address, scores []uint32) error {
	// TODO: Implement batch registration for gas optimization
	// For now, register one by one
	for i, addr := range contracts {
		if i < len(scores) {
			if err := rc.RegisterProof(ctx, common.Address{}, addr, scores[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

// VerificationMetadata holds metadata for a verification event
type VerificationMetadata struct {
	UserAddress     common.Address
	ContractAddress common.Address
	ProofHash       common.Hash
	RiskScore       uint32
	VulnCount       uint32
	Timestamp       uint64
	TxHash          common.Hash
}

// StoreVerification stores verification metadata on-chain and returns tx hash
func (rc *RegistryClient) StoreVerification(ctx context.Context, metadata *VerificationMetadata) (*common.Hash, error) {
	if !rc.enabled {
		rc.logger.Debug("Registry client not enabled, skipping verification storage")
		return nil, nil
	}

	rc.logger.Info("Storing verification on-chain",
		"user", metadata.UserAddress.Hex(),
		"contract", metadata.ContractAddress.Hex(),
		"proof_hash", metadata.ProofHash.Hex(),
		"risk_score", metadata.RiskScore,
	)

	// First check if contract is already registered
	existingScore, err := rc.GetRiskScore(ctx, metadata.ContractAddress)
	if err != nil || existingScore == 0 {
		// Contract not registered, register it first
		err = rc.RegisterProof(ctx, metadata.UserAddress, metadata.ContractAddress, metadata.RiskScore)
		if err != nil {
			return nil, err
		}
	} else {
		// Contract exists, update risk score
		err = rc.UpdateRiskScore(ctx, metadata.ContractAddress, metadata.RiskScore, metadata.VulnCount)
		if err != nil {
			return nil, err
		}
	}

	// TODO: Get actual transaction hash from the contract interaction
	// For now, return the proof hash as placeholder
	txHash := metadata.ProofHash
	return &txHash, nil
}
