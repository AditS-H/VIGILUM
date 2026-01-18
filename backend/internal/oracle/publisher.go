// Package oracle implements threat intelligence aggregation and on-chain publishing.
package oracle

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/vigilum/backend/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════════
// THREAT ORACLE CONTRACT ABI
// ═══════════════════════════════════════════════════════════════════════════════

// ThreatOracleABI is the ABI for the ThreatOracle contract.
const ThreatOracleABI = `[
	{
		"inputs": [{"name": "target", "type": "address"}, {"name": "score", "type": "uint8"}],
		"name": "updateRiskScore",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "targets", "type": "address[]"}, {"name": "scores", "type": "uint8[]"}],
		"name": "batchUpdateRiskScores",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "target", "type": "address"}],
		"name": "getRiskScore",
		"outputs": [{"name": "", "type": "uint8"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "target", "type": "address"}],
		"name": "getLastUpdate",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "target", "type": "address"},
			{"indexed": false, "name": "riskScore", "type": "uint8"},
			{"indexed": false, "name": "timestamp", "type": "uint256"}
		],
		"name": "RiskUpdated",
		"type": "event"
	}
]`

// ═══════════════════════════════════════════════════════════════════════════════
// PUBLISHER
// ═══════════════════════════════════════════════════════════════════════════════

// Publisher handles publishing threat signals to the on-chain oracle.
type Publisher struct {
	mu              sync.RWMutex
	client          *ethclient.Client
	contractAddr    common.Address
	contractABI     abi.ABI
	privateKey      *ecdsa.PrivateKey
	fromAddress     common.Address
	chainID         *big.Int
	
	// Rate limiting
	lastUpdateTime  map[string]time.Time // target -> last update
	minUpdateInterval time.Duration
	
	// Queue
	pending         []SignalUpdate
	maxBatchSize    int
	
	// Stats
	totalPublished  int
	totalFailed     int
}

// PublisherConfig contains configuration for the publisher.
type PublisherConfig struct {
	RPCEndpoint       string
	ContractAddress   string
	PrivateKey        string
	ChainID           int64
	MinUpdateInterval time.Duration
	MaxBatchSize      int
}

// NewPublisher creates a new on-chain signal publisher.
func NewPublisher(cfg PublisherConfig) (*Publisher, error) {
	// Connect to Ethereum
	client, err := ethclient.Dial(cfg.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum: %w", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Derive address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Parse ABI
	contractABI, err := abi.JSON(strings.NewReader(ThreatOracleABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	// Set defaults
	minInterval := cfg.MinUpdateInterval
	if minInterval == 0 {
		minInterval = time.Hour // Default: 1 update per hour per target
	}
	maxBatch := cfg.MaxBatchSize
	if maxBatch == 0 {
		maxBatch = 50 // Default batch size
	}

	return &Publisher{
		client:            client,
		contractAddr:      common.HexToAddress(cfg.ContractAddress),
		contractABI:       contractABI,
		privateKey:        privateKey,
		fromAddress:       fromAddress,
		chainID:           big.NewInt(cfg.ChainID),
		lastUpdateTime:    make(map[string]time.Time),
		minUpdateInterval: minInterval,
		maxBatchSize:      maxBatch,
	}, nil
}

// QueueUpdate adds a signal update to the publishing queue.
func (p *Publisher) QueueUpdate(signal ThreatSignal) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check rate limit
	key := fmt.Sprintf("%d:%s", signal.Target.ChainID, signal.Target.Address)
	if lastUpdate, ok := p.lastUpdateTime[key]; ok {
		if time.Since(lastUpdate) < p.minUpdateInterval {
			return fmt.Errorf("rate limited: last update was %v ago", time.Since(lastUpdate))
		}
	}

	// Add to queue
	p.pending = append(p.pending, SignalUpdate{
		TargetAddress: signal.Target.Address,
		ChainID:       signal.Target.ChainID,
		RiskScore:     signal.RiskScore,
		Reason:        signal.SummaryReason,
		Timestamp:     time.Now(),
		Status:        PublishStatusPending,
	})

	return nil
}

// Publish sends a single signal update to the on-chain oracle.
func (p *Publisher) Publish(ctx context.Context, signal ThreatSignal) (*types.Transaction, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check rate limit
	key := fmt.Sprintf("%d:%s", signal.Target.ChainID, signal.Target.Address)
	if lastUpdate, ok := p.lastUpdateTime[key]; ok {
		if time.Since(lastUpdate) < p.minUpdateInterval {
			return nil, fmt.Errorf("rate limited")
		}
	}

	// Get nonce
	nonce, err := p.client.PendingNonceAt(ctx, p.fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := p.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Encode function call
	targetAddr := common.HexToAddress(string(signal.Target.Address))
	data, err := p.contractABI.Pack("updateRiskScore", targetAddr, signal.RiskScore)
	if err != nil {
		return nil, fmt.Errorf("failed to encode call: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		p.contractAddr,
		big.NewInt(0),
		uint64(100000), // Gas limit
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(p.chainID), p.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := p.client.SendTransaction(ctx, signedTx); err != nil {
		p.totalFailed++
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Update rate limit tracker
	p.lastUpdateTime[key] = time.Now()
	p.totalPublished++

	return signedTx, nil
}

// PublishBatch sends multiple signal updates in a single transaction.
func (p *Publisher) PublishBatch(ctx context.Context, signals []ThreatSignal) (*types.Transaction, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(signals) == 0 {
		return nil, nil
	}

	if len(signals) > p.maxBatchSize {
		signals = signals[:p.maxBatchSize]
	}

	// Filter by rate limit
	var filtered []ThreatSignal
	now := time.Now()
	for _, signal := range signals {
		key := fmt.Sprintf("%d:%s", signal.Target.ChainID, signal.Target.Address)
		if lastUpdate, ok := p.lastUpdateTime[key]; ok {
			if now.Sub(lastUpdate) < p.minUpdateInterval {
				continue
			}
		}
		filtered = append(filtered, signal)
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("all signals rate limited")
	}

	// Prepare batch data
	targets := make([]common.Address, len(filtered))
	scores := make([]uint8, len(filtered))
	for i, signal := range filtered {
		targets[i] = common.HexToAddress(string(signal.Target.Address))
		scores[i] = signal.RiskScore
	}

	// Get nonce
	nonce, err := p.client.PendingNonceAt(ctx, p.fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := p.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Encode batch call
	data, err := p.contractABI.Pack("batchUpdateRiskScores", targets, scores)
	if err != nil {
		return nil, fmt.Errorf("failed to encode batch call: %w", err)
	}

	// Estimate gas (batch is more expensive)
	gasLimit := uint64(50000) + uint64(30000*len(filtered))

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		p.contractAddr,
		big.NewInt(0),
		gasLimit,
		gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(p.chainID), p.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign batch transaction: %w", err)
	}

	// Send transaction
	if err := p.client.SendTransaction(ctx, signedTx); err != nil {
		p.totalFailed += len(filtered)
		return nil, fmt.Errorf("failed to send batch transaction: %w", err)
	}

	// Update rate limit trackers
	for _, signal := range filtered {
		key := fmt.Sprintf("%d:%s", signal.Target.ChainID, signal.Target.Address)
		p.lastUpdateTime[key] = now
	}
	p.totalPublished += len(filtered)

	return signedTx, nil
}

// FlushQueue publishes all pending updates.
func (p *Publisher) FlushQueue(ctx context.Context) (int, error) {
	p.mu.Lock()
	pending := p.pending
	p.pending = nil
	p.mu.Unlock()

	if len(pending) == 0 {
		return 0, nil
	}

	// Convert to signals
	var signals []ThreatSignal
	for _, update := range pending {
		signals = append(signals, ThreatSignal{
			Target: TargetRef{
				Address: update.TargetAddress,
				ChainID: update.ChainID,
			},
			RiskScore: update.RiskScore,
		})
	}

	// Publish in batches
	published := 0
	for i := 0; i < len(signals); i += p.maxBatchSize {
		end := i + p.maxBatchSize
		if end > len(signals) {
			end = len(signals)
		}
		
		batch := signals[i:end]
		_, err := p.PublishBatch(ctx, batch)
		if err != nil {
			// Re-queue failed updates
			p.mu.Lock()
			p.pending = append(p.pending, pending[i:end]...)
			p.mu.Unlock()
			return published, err
		}
		published += len(batch)
	}

	return published, nil
}

// GetOnChainScore reads the current risk score from the on-chain oracle.
func (p *Publisher) GetOnChainScore(ctx context.Context, target domain.Address) (uint8, error) {
	targetAddr := common.HexToAddress(string(target))
	
	data, err := p.contractABI.Pack("getRiskScore", targetAddr)
	if err != nil {
		return 0, err
	}

	result, err := p.client.CallContract(ctx, ethereum.CallMsg{
		To:   &p.contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return 0, err
	}

	var score uint8
	if err := p.contractABI.UnpackIntoInterface(&score, "getRiskScore", result); err != nil {
		return 0, err
	}

	return score, nil
}

// Stats returns publisher statistics.
func (p *Publisher) Stats() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]int{
		"total_published":  p.totalPublished,
		"total_failed":     p.totalFailed,
		"pending_updates":  len(p.pending),
		"tracked_targets":  len(p.lastUpdateTime),
	}
}

// Close closes the publisher and releases resources.
func (p *Publisher) Close() {
	p.client.Close()
}
