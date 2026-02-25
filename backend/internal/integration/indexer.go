package integration

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// BlockchainIndexer monitors blockchain for contract deployments and triggers scans
type BlockchainIndexer struct {
	logger            *slog.Logger
	currentBlock      uint64
	lastBlockTime     time.Time
	isRunning         bool
	mu                sync.RWMutex
	stopChan          chan bool
	discoveredContracts chan *domain.Contract
	blockCache        map[uint64]*BlockData
	contractStore     ContractStore
}

// ContractStore defines storage for discovered contracts
type ContractStore interface {
	SaveContract(ctx context.Context, contract *domain.Contract) error
	GetContract(ctx context.Context, address domain.Address) (*domain.Contract, error)
	ListContracts(ctx context.Context, limit int) ([]*domain.Contract, error)
}

// BlockData contains information about a block
type BlockData struct {
	Number        uint64
	Hash          string
	Timestamp     time.Time
	Transactions  []TransactionData
	ContractCount int
}

// TransactionData contains transaction info
type TransactionData struct {
	Hash    string
	From    domain.Address
	To      *domain.Address
	Data    string
	IsCreate bool
	Contract *ContractData
}

// ContractData contains contract info
type ContractData struct {
	Address         domain.Address
	Bytecode        []byte
	DeploymentBlock uint64
	DeploymentTx    string
	Source          string
	ABI             string
}

// BlockProcessor processes blocks and extracts contracts
type BlockProcessor interface {
	ProcessBlock(ctx context.Context, blockNumber uint64) (*BlockData, error)
}

// NewBlockchainIndexer creates a new indexer
func NewBlockchainIndexer(logger *slog.Logger, store ContractStore) *BlockchainIndexer {
	return &BlockchainIndexer{
		logger:              logger.With("service", "indexer"),
		currentBlock:        0,
		blockCache:          make(map[uint64]*BlockData),
		stopChan:            make(chan bool),
		discoveredContracts: make(chan *domain.Contract, 100),
		contractStore:       store,
	}
}

// Start begins monitoring the blockchain
func (bi *BlockchainIndexer) Start(ctx context.Context) error {
	bi.mu.Lock()
	if bi.isRunning {
		bi.mu.Unlock()
		return fmt.Errorf("indexer already running")
	}
	bi.isRunning = true
	bi.mu.Unlock()

	bi.logger.Info("starting blockchain indexer")

	// Start in separate goroutine
	go bi.run(ctx)

	return nil
}

// Stop stops the indexer
func (bi *BlockchainIndexer) Stop() error {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	if !bi.isRunning {
		return fmt.Errorf("indexer not running")
	}

	bi.logger.Info("stopping blockchain indexer")
	bi.isRunning = false
	bi.stopChan <- true

	return nil
}

// run is the main indexing loop
func (bi *BlockchainIndexer) run(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second) // Poll every 15 seconds (Ethereum block time)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			bi.logger.Info("indexer context cancelled")
			return
		case <-bi.stopChan:
			bi.logger.Info("indexer stop signal received")
			return
		case <-ticker.C:
			bi.processNextBlock(ctx)
		}
	}
}

// processNextBlock processes the next block in sequence
func (bi *BlockchainIndexer) processNextBlock(ctx context.Context) {
	bi.mu.Lock()
	currentBlock := bi.currentBlock
	bi.mu.Unlock()

	// In production, would fetch from blockchain RPC
	// For now, simulate block processing
	blockData, err := bi.simulateBlockProcessing(ctx, currentBlock)
	if err != nil {
		bi.logger.Warn("block processing failed",
			"block", currentBlock,
			"error", err,
		)
		return
	}

	if blockData == nil {
		return // No new block yet
	}

	// Process contracts in the block
	for _, tx := range blockData.Transactions {
		if tx.IsCreate && tx.Contract != nil {
			contract := bi.contractDataToDomain(tx.Contract, blockData)
			bi.discoveredContracts <- contract

			// Persist to store
			if err := bi.contractStore.SaveContract(ctx, contract); err != nil {
				bi.logger.Warn("failed to save contract",
					"address", contract.Address,
					"error", err,
				)
			}

			bi.logger.Info("discovered contract",
				"address", contract.Address,
				"block", blockData.Number,
				"txHash", tx.Hash,
			)
		}
	}

	// Update current block
	bi.mu.Lock()
	bi.currentBlock = blockData.Number + 1
	bi.lastBlockTime = blockData.Timestamp
	bi.mu.Unlock()

	bi.logger.Debug("processed block",
		"block", blockData.Number,
		"contracts", blockData.ContractCount,
	)
}

// contractDataToDomain converts ContractData to domain.Contract
func (bi *BlockchainIndexer) contractDataToDomain(contractData *ContractData, blockData *BlockData) *domain.Contract {
	return &domain.Contract{
		ID:              domain.ContractID(contractData.Address),
		Address:         contractData.Address,
		ChainID:         1, // Ethereum mainnet - in production get from config
		Bytecode:        contractData.Bytecode,
		SourceCode:      contractData.Source,
		ABI:             contractData.ABI,
		DeployedAt:      blockData.Timestamp,
		DeployTxHash:    domain.Hash(contractData.DeploymentTx),
		IsVerified:      contractData.Source != "", // Has source = verified
		Labels:          []string{"discovered", "indexed"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// simulateBlockProcessing simulates fetching and processing a block
// In production, this would use web3.py or similar for actual blockchain data
func (bi *BlockchainIndexer) simulateBlockProcessing(ctx context.Context, blockNumber uint64) (*BlockData, error) {
	// Check if we're too far ahead (no blocks available)
	if blockNumber > 19500000 { // Ethereum as of now
		return nil, nil
	}

	// Simulate variable block times
	blockData := &BlockData{
		Number:       blockNumber,
		Hash:         fmt.Sprintf("0x%064x", blockNumber),
		Timestamp:    time.Now(),
		Transactions: []TransactionData{},
	}

	// Simulate 1-3 contract deployments per 100 blocks
	contractsInBlock := blockNumber % 100
	if contractsInBlock > 95 { // 5% chance of contract deployment
		contractData := &ContractData{
			Address:         domain.Address(fmt.Sprintf("0x%040x", blockNumber)),
			Bytecode:        []byte(fmt.Sprintf("// Contract %d\npragma solidity ^0.8.0;", blockNumber)),
			DeploymentBlock: blockNumber,
			DeploymentTx:    fmt.Sprintf("0x%064x", blockNumber),
		}

		blockData.Transactions = append(blockData.Transactions, TransactionData{
			Hash:     fmt.Sprintf("0x%064x", blockNumber),
			IsCreate: true,
			Contract: contractData,
		})
		blockData.ContractCount++
	}

	return blockData, nil
}

// GetCurrentBlock returns the current block being indexed
func (bi *BlockchainIndexer) GetCurrentBlock() uint64 {
	bi.mu.RLock()
	defer bi.mu.RUnlock()
	return bi.currentBlock
}

// IsRunning returns whether the indexer is active
func (bi *BlockchainIndexer) IsRunning() bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()
	return bi.isRunning
}

// GetDiscoveredContracts returns the channel of discovered contracts
func (bi *BlockchainIndexer) GetDiscoveredContracts() <-chan *domain.Contract {
	return bi.discoveredContracts
}

// Stats contains indexer statistics
type Stats struct {
	CurrentBlock     uint64
	TotalProcessed   uint64
	ContractsFound   uint64
	LastMineTime     time.Time
	IsRunning        bool
}

// GetStats returns current indexer statistics
func (bi *BlockchainIndexer) GetStats() Stats {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return Stats{
		CurrentBlock:   bi.currentBlock,
		LastMineTime:   bi.lastBlockTime,
		IsRunning:      bi.isRunning,
		ContractsFound: uint64(len(bi.discoveredContracts)),
	}
}
