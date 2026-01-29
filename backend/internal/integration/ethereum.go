// Package integration provides external service integrations (Ethereum, IPFS, etc.)
package integration

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumConfig holds Ethereum client configuration
type EthereumConfig struct {
	RPCURL                string        // Ethereum RPC endpoint (e.g., Infura, Alchemy)
	ChainID               *big.Int      // Chain ID (1 for mainnet, 11155111 for Sepolia)
	IdentityFirewallAddr  string        // Identity Firewall contract address
	PrivateKey            string        // Private key for signing transactions (optional)
	GasLimit              uint64        // Gas limit for transactions
	Timeout               time.Duration // RPC call timeout
	MaxRetries            int           // Max retries for failed calls
	RetryDelay            time.Duration // Delay between retries
}

// DefaultEthereumConfig returns default configuration
func DefaultEthereumConfig() *EthereumConfig {
	return &EthereumConfig{
		RPCURL:       os.Getenv("ETH_RPC_URL"),
		ChainID:      big.NewInt(11155111), // Sepolia
		PrivateKey:   os.Getenv("ETH_PRIVATE_KEY"),
		GasLimit:     300000,
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryDelay:   1 * time.Second,
	}
}

// EthereumClient wraps go-ethereum client with VIGILUM-specific methods
type EthereumClient struct {
	client     *ethclient.Client
	config     *EthereumConfig
	privateKey *ecdsa.PrivateKey
	publicAddr common.Address

	// Contract instances
	identityFirewallABI  abi.ABI
	identityFirewallAddr common.Address

	mu sync.RWMutex
}

// ProofRecord represents an on-chain proof record
type ProofRecord struct {
	User        common.Address
	ProofHash   [32]byte
	VerifiedAt  uint64
	ExpiresAt   uint64
	IsRevoked   bool
	RevokedAt   uint64
	RevokedBy   common.Address
	RevokeReason string
}

// UserStatus represents a user's on-chain status
type UserStatus struct {
	LatestProofHash [32]byte
	TotalProofs     uint32
	FirstVerifiedAt uint64
	LastVerifiedAt  uint64
}

// NewEthereumClient creates a new Ethereum client
func NewEthereumClient(config *EthereumConfig) (*EthereumClient, error) {
	if config == nil {
		config = DefaultEthereumConfig()
	}

	if config.RPCURL == "" {
		return nil, fmt.Errorf("ETH_RPC_URL is required")
	}

	// Connect to Ethereum node
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	client, err := ethclient.DialContext(ctx, config.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Verify connection
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	if config.ChainID != nil && chainID.Cmp(config.ChainID) != 0 {
		return nil, fmt.Errorf("chain ID mismatch: expected %s, got %s", config.ChainID.String(), chainID.String())
	}
	config.ChainID = chainID

	ec := &EthereumClient{
		client: client,
		config: config,
	}

	// Parse private key if provided
	if config.PrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(config.PrivateKey, "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}
		ec.privateKey = privateKey
		ec.publicAddr = crypto.PubkeyToAddress(privateKey.PublicKey)
	}

	// Load Identity Firewall ABI
	if err := ec.loadIdentityFirewallABI(); err != nil {
		return nil, fmt.Errorf("failed to load IdentityFirewall ABI: %w", err)
	}

	// Set contract address
	if config.IdentityFirewallAddr != "" {
		ec.identityFirewallAddr = common.HexToAddress(config.IdentityFirewallAddr)
	}

	log.Printf("Ethereum client connected: chainID=%s, rpc=%s", chainID.String(), config.RPCURL)
	return ec, nil
}

// loadIdentityFirewallABI loads the IdentityFirewall contract ABI
func (ec *EthereumClient) loadIdentityFirewallABI() error {
	// ABI for IdentityFirewall/VigilumRegistry contracts (combined minimal ABI for integration)
	abiJSON := `[
		{
			"name": "verifyHumanProof",
			"type": "function",
			"inputs": [{"name": "proof", "type": "bytes"}],
			"outputs": [{"name": "", "type": "bool"}],
			"stateMutability": "nonpayable"
		},
		{
			"name": "verifyProofFor",
			"type": "function",
			"inputs": [
				{"name": "user", "type": "address"},
				{"name": "proof", "type": "bytes"}
			],
			"outputs": [{"name": "", "type": "bool"}],
			"stateMutability": "nonpayable"
		},
		{
			"name": "hasValidProof",
			"type": "function",
			"inputs": [{"name": "user", "type": "address"}],
			"outputs": [{"name": "", "type": "bool"}],
			"stateMutability": "view"
		},
		{
			"name": "getProofExpiry",
			"type": "function",
			"inputs": [{"name": "proofHash", "type": "bytes32"}],
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view"
		},
		{
			"name": "isProofValid",
			"type": "function",
			"inputs": [{"name": "proofHash", "type": "bytes32"}],
			"outputs": [{"name": "", "type": "bool"}],
			"stateMutability": "view"
		},
		{
			"name": "revokeProof",
			"type": "function",
			"inputs": [
				{"name": "proofHash", "type": "bytes32"},
				{"name": "reason", "type": "string"}
			],
			"outputs": [],
			"stateMutability": "nonpayable"
		},
		{
			"name": "getUserStatus",
			"type": "function",
			"inputs": [{"name": "user", "type": "address"}],
			"outputs": [
				{
					"name": "",
					"type": "tuple",
					"components": [
						{"name": "latestProofHash", "type": "bytes32"},
						{"name": "totalProofs", "type": "uint32"},
						{"name": "firstVerifiedAt", "type": "uint64"},
						{"name": "lastVerifiedAt", "type": "uint64"}
					]
				}
			],
			"stateMutability": "view"
		},
		{
			"name": "registerContract",
			"type": "function",
			"inputs": [
				{"name": "contractAddr", "type": "address"},
				{"name": "bytecodeHash", "type": "bytes32"},
				{"name": "initialRiskScore", "type": "uint256"}
			],
			"outputs": [],
			"stateMutability": "nonpayable"
		},
		{
			"name": "updateRiskScore",
			"type": "function",
			"inputs": [
				{"name": "contractAddr", "type": "address"},
				{"name": "newScore", "type": "uint256"},
				{"name": "vulnCount", "type": "uint32"}
			],
			"outputs": [],
			"stateMutability": "nonpayable"
		},
		{
			"name": "getRiskScore",
			"type": "function",
			"inputs": [{"name": "contractAddr", "type": "address"}],
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view"
		},
		{
			"name": "isBlacklisted",
			"type": "function",
			"inputs": [{"name": "contractAddr", "type": "address"}],
			"outputs": [{"name": "", "type": "bool"}],
			"stateMutability": "view"
		},
		{
			"name": "blacklistContract",
			"type": "function",
			"inputs": [
				{"name": "contractAddr", "type": "address"},
				{"name": "reason", "type": "string"}
			],
			"outputs": [],
			"stateMutability": "nonpayable"
		}
					"name": "",
					"type": "tuple",
					"components": [
						{"name": "latestProofHash", "type": "bytes32"},
						{"name": "totalProofs", "type": "uint32"},
						{"name": "firstVerifiedAt", "type": "uint64"},
						{"name": "lastVerifiedAt", "type": "uint64"}
					]
				}
			],
			"stateMutability": "view"
		},
		{
			"name": "totalVerifiedProofs",
			"type": "function",
			"inputs": [],
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view"
		},
		{
			"name": "totalUniqueUsers",
			"type": "function",
			"inputs": [],
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view"
		},
		{
			"name": "totalRevokedProofs",
			"type": "function",
			"inputs": [],
			"outputs": [{"name": "", "type": "uint256"}],
			"stateMutability": "view"
		},
		{
			"name": "ProofVerified",
			"type": "event",
			"inputs": [
				{"indexed": true, "name": "user", "type": "address"},
				{"indexed": true, "name": "proofHash", "type": "bytes32"},
				{"indexed": false, "name": "timestamp", "type": "uint256"}
			]
		},
		{
			"name": "ProofRevoked",
			"type": "event",
			"inputs": [
				{"indexed": true, "name": "user", "type": "address"},
				{"indexed": true, "name": "proofHash", "type": "bytes32"},
				{"indexed": false, "name": "reason", "type": "string"}
			]
		}
	]`

	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return fmt.Errorf("failed to parse ABI: %w", err)
	}

	ec.identityFirewallABI = parsedABI
	return nil
}

// SetIdentityFirewallAddress sets the Identity Firewall contract address
func (ec *EthereumClient) SetIdentityFirewallAddress(addr string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.identityFirewallAddr = common.HexToAddress(addr)
}

// Close closes the Ethereum client connection
func (ec *EthereumClient) Close() {
	ec.client.Close()
}

// ═══════════════════════════════════════════════════════════════════════════════
// READ METHODS (VIEW CALLS)
// ═══════════════════════════════════════════════════════════════════════════════

// HasValidProof checks if a user has a valid proof on-chain
func (ec *EthereumClient) HasValidProof(ctx context.Context, userAddr string) (bool, error) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return false, fmt.Errorf("IdentityFirewall address not set")
	}

	user := common.HexToAddress(userAddr)
	data, err := ec.identityFirewallABI.Pack("hasValidProof", user)
	if err != nil {
		return false, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := ec.callContract(ctx, data)
	if err != nil {
		return false, fmt.Errorf("contract call failed: %w", err)
	}

	var hasProof bool
	if err := ec.identityFirewallABI.UnpackIntoInterface(&hasProof, "hasValidProof", result); err != nil {
		return false, fmt.Errorf("failed to unpack result: %w", err)
	}

	return hasProof, nil
}

// GetProofExpiry gets the expiry timestamp for a proof hash
func (ec *EthereumClient) GetProofExpiry(ctx context.Context, proofHash [32]byte) (uint64, error) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return 0, fmt.Errorf("IdentityFirewall address not set")
	}

	data, err := ec.identityFirewallABI.Pack("getProofExpiry", proofHash)
	if err != nil {
		return 0, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := ec.callContract(ctx, data)
	if err != nil {
		return 0, fmt.Errorf("contract call failed: %w", err)
	}

	var expiry *big.Int
	if err := ec.identityFirewallABI.UnpackIntoInterface(&expiry, "getProofExpiry", result); err != nil {
		return 0, fmt.Errorf("failed to unpack result: %w", err)
	}

	return expiry.Uint64(), nil
}

// IsProofValid checks if a specific proof hash is valid
func (ec *EthereumClient) IsProofValid(ctx context.Context, proofHash [32]byte) (bool, error) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return false, fmt.Errorf("IdentityFirewall address not set")
	}

	data, err := ec.identityFirewallABI.Pack("isProofValid", proofHash)
	if err != nil {
		return false, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := ec.callContract(ctx, data)
	if err != nil {
		return false, fmt.Errorf("contract call failed: %w", err)
	}

	var isValid bool
	if err := ec.identityFirewallABI.UnpackIntoInterface(&isValid, "isProofValid", result); err != nil {
		return false, fmt.Errorf("failed to unpack result: %w", err)
	}

	return isValid, nil
}

// GetUserStatus gets a user's verification status from the contract
func (ec *EthereumClient) GetUserStatus(ctx context.Context, userAddr string) (*UserStatus, error) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return nil, fmt.Errorf("IdentityFirewall address not set")
	}

	user := common.HexToAddress(userAddr)
	data, err := ec.identityFirewallABI.Pack("getUserStatus", user)
	if err != nil {
		return nil, fmt.Errorf("failed to pack call data: %w", err)
	}

	result, err := ec.callContract(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("contract call failed: %w", err)
	}

	// Unpack the tuple result
	var status struct {
		LatestProofHash [32]byte
		TotalProofs     uint32
		FirstVerifiedAt uint64
		LastVerifiedAt  uint64
	}

	if err := ec.identityFirewallABI.UnpackIntoInterface(&status, "getUserStatus", result); err != nil {
		return nil, fmt.Errorf("failed to unpack result: %w", err)
	}

	return &UserStatus{
		LatestProofHash: status.LatestProofHash,
		TotalProofs:     status.TotalProofs,
		FirstVerifiedAt: status.FirstVerifiedAt,
		LastVerifiedAt:  status.LastVerifiedAt,
	}, nil
}

// GetContractStats gets overall contract statistics
func (ec *EthereumClient) GetContractStats(ctx context.Context) (totalProofs, totalUsers, totalRevoked uint64, err error) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return 0, 0, 0, fmt.Errorf("IdentityFirewall address not set")
	}

	// Get total verified proofs
	data1, _ := ec.identityFirewallABI.Pack("totalVerifiedProofs")
	result1, err := ec.callContract(ctx, data1)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get totalVerifiedProofs: %w", err)
	}
	var total1 *big.Int
	ec.identityFirewallABI.UnpackIntoInterface(&total1, "totalVerifiedProofs", result1)
	totalProofs = total1.Uint64()

	// Get total unique users
	data2, _ := ec.identityFirewallABI.Pack("totalUniqueUsers")
	result2, err := ec.callContract(ctx, data2)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get totalUniqueUsers: %w", err)
	}
	var total2 *big.Int
	ec.identityFirewallABI.UnpackIntoInterface(&total2, "totalUniqueUsers", result2)
	totalUsers = total2.Uint64()

	// Get total revoked proofs
	data3, _ := ec.identityFirewallABI.Pack("totalRevokedProofs")
	result3, err := ec.callContract(ctx, data3)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get totalRevokedProofs: %w", err)
	}
	var total3 *big.Int
	ec.identityFirewallABI.UnpackIntoInterface(&total3, "totalRevokedProofs", result3)
	totalRevoked = total3.Uint64()

	return totalProofs, totalUsers, totalRevoked, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// WRITE METHODS (TRANSACTIONS)
// ═══════════════════════════════════════════════════════════════════════════════

// VerifyHumanProof submits a proof verification transaction
func (ec *EthereumClient) VerifyHumanProof(ctx context.Context, proof []byte) (txHash common.Hash, err error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key not set - cannot send transactions")
	}

	if ec.identityFirewallAddr == (common.Address{}) {
		return common.Hash{}, fmt.Errorf("IdentityFirewall address not set")
	}

	data, err := ec.identityFirewallABI.Pack("verifyHumanProof", proof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack transaction data: %w", err)
	}

	return ec.sendTransaction(ctx, data)
}

// VerifyProofFor submits a proof verification for another user (requires authorization)
func (ec *EthereumClient) VerifyProofFor(ctx context.Context, userAddr string, proof []byte) (txHash common.Hash, err error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key not set - cannot send transactions")
	}

	if ec.identityFirewallAddr == (common.Address{}) {
		return common.Hash{}, fmt.Errorf("IdentityFirewall address not set")
	}

	user := common.HexToAddress(userAddr)
	data, err := ec.identityFirewallABI.Pack("verifyProofFor", user, proof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack transaction data: %w", err)
	}

	return ec.sendTransaction(ctx, data)
}

// RevokeProof submits a proof revocation transaction
func (ec *EthereumClient) RevokeProof(ctx context.Context, proofHash [32]byte, reason string) (txHash common.Hash, err error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.privateKey == nil {
		return common.Hash{}, fmt.Errorf("private key not set - cannot send transactions")
	}

	if ec.identityFirewallAddr == (common.Address{}) {
		return common.Hash{}, fmt.Errorf("IdentityFirewall address not set")
	}

	data, err := ec.identityFirewallABI.Pack("revokeProof", proofHash, reason)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack transaction data: %w", err)
	}

	return ec.sendTransaction(ctx, data)
}

// ═══════════════════════════════════════════════════════════════════════════════
// EVENT WATCHING
// ═══════════════════════════════════════════════════════════════════════════════

// ProofVerifiedEvent represents a decoded ProofVerified event
type ProofVerifiedEvent struct {
	User      common.Address
	ProofHash [32]byte
	Timestamp uint64
	TxHash    common.Hash
	Block     uint64
}

// ProofRevokedEvent represents a decoded ProofRevoked event
type ProofRevokedEvent struct {
	User      common.Address
	ProofHash [32]byte
	Reason    string
	TxHash    common.Hash
	Block     uint64
}

// WatchProofVerified watches for ProofVerified events
func (ec *EthereumClient) WatchProofVerified(ctx context.Context, fromBlock uint64, eventCh chan<- *ProofVerifiedEvent) error {
	ec.mu.RLock()
	contractAddr := ec.identityFirewallAddr
	ec.mu.RUnlock()

	if contractAddr == (common.Address{}) {
		return fmt.Errorf("IdentityFirewall address not set")
	}

	// Get the event signature
	eventSig := ec.identityFirewallABI.Events["ProofVerified"].ID

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{eventSig}},
	}

	logs := make(chan types.Log)
	sub, err := ec.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				log.Printf("Event subscription error: %v", err)
				return
			case vLog := <-logs:
				event, err := ec.parseProofVerifiedEvent(vLog)
				if err != nil {
					log.Printf("Failed to parse ProofVerified event: %v", err)
					continue
				}
				eventCh <- event
			}
		}
	}()

	return nil
}

// parseProofVerifiedEvent parses a ProofVerified event from a log
func (ec *EthereumClient) parseProofVerifiedEvent(vLog types.Log) (*ProofVerifiedEvent, error) {
	event := &ProofVerifiedEvent{
		TxHash: vLog.TxHash,
		Block:  vLog.BlockNumber,
	}

	// Indexed parameters are in topics
	if len(vLog.Topics) < 3 {
		return nil, fmt.Errorf("insufficient topics in log")
	}

	event.User = common.BytesToAddress(vLog.Topics[1].Bytes())
	event.ProofHash = vLog.Topics[2]

	// Non-indexed parameters are in data
	if len(vLog.Data) >= 32 {
		timestamp := new(big.Int).SetBytes(vLog.Data[:32])
		event.Timestamp = timestamp.Uint64()
	}

	return event, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// UTILITY METHODS
// ═══════════════════════════════════════════════════════════════════════════════

// callContract makes a read-only call to the contract
func (ec *EthereumClient) callContract(ctx context.Context, data []byte) ([]byte, error) {
	msg := ethereum.CallMsg{
		To:   &ec.identityFirewallAddr,
		Data: data,
	}

	var result []byte
	var err error

	for i := 0; i < ec.config.MaxRetries; i++ {
		callCtx, cancel := context.WithTimeout(ctx, ec.config.Timeout)
		result, err = ec.client.CallContract(callCtx, msg, nil)
		cancel()

		if err == nil {
			return result, nil
		}

		log.Printf("Contract call failed (attempt %d/%d): %v", i+1, ec.config.MaxRetries, err)
		time.Sleep(ec.config.RetryDelay)
	}

	return nil, fmt.Errorf("contract call failed after %d retries: %w", ec.config.MaxRetries, err)
}

// sendTransaction sends a signed transaction to the contract
func (ec *EthereumClient) sendTransaction(ctx context.Context, data []byte) (common.Hash, error) {
	// Get nonce
	nonce, err := ec.client.PendingNonceAt(ctx, ec.publicAddr)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := ec.client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas
	msg := ethereum.CallMsg{
		From: ec.publicAddr,
		To:   &ec.identityFirewallAddr,
		Data: data,
	}
	gasLimit, err := ec.client.EstimateGas(ctx, msg)
	if err != nil {
		log.Printf("Gas estimation failed, using default: %v", err)
		gasLimit = ec.config.GasLimit
	}

	// Create transaction
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &ec.identityFirewallAddr,
		Value:    big.NewInt(0),
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(ec.config.ChainID), ec.privateKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := ec.client.SendTransaction(ctx, signedTx); err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Printf("Transaction sent: %s", signedTx.Hash().Hex())
	return signedTx.Hash(), nil
}

// WaitForTransaction waits for a transaction to be mined
func (ec *EthereumClient) WaitForTransaction(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return bind.WaitMined(ctx, ec.client, &types.Transaction{})
}

// GetTransactionReceipt gets the receipt for a transaction
func (ec *EthereumClient) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return ec.client.TransactionReceipt(ctx, txHash)
}

// GetBlockNumber gets the current block number
func (ec *EthereumClient) GetBlockNumber(ctx context.Context) (uint64, error) {
	return ec.client.BlockNumber(ctx)
}

// GetBalance gets the ETH balance for an address
func (ec *EthereumClient) GetBalance(ctx context.Context, addr string) (*big.Int, error) {
	address := common.HexToAddress(addr)
	return ec.client.BalanceAt(ctx, address, nil)
}

// PublicAddress returns the public address derived from the private key
func (ec *EthereumClient) PublicAddress() string {
	if ec.publicAddr == (common.Address{}) {
		return ""
	}
	return ec.publicAddr.Hex()
}

// ChainID returns the chain ID
func (ec *EthereumClient) ChainID() *big.Int {
	return ec.config.ChainID
}

// ═══════════════════════════════════════════════════════════════════════════════
// VIGILUM REGISTRY CONTRACT METHODS
// ═══════════════════════════════════════════════════════════════════════════════

// RegisterContract registers a new contract in VigilumRegistry
func (ec *EthereumClient) RegisterContract(ctx context.Context, contractAddr common.Address, bytecodeHash common.Hash, riskScore *big.Int) error {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return fmt.Errorf("VigilumRegistry address not set")
	}

	// Pack function call: registerContract(address,bytes32,uint256)
	data, err := ec.identityFirewallABI.Pack("registerContract", contractAddr, bytecodeHash, riskScore)
	if err != nil {
		return fmt.Errorf("failed to pack registerContract call: %w", err)
	}

	// Send transaction
	_, err = ec.sendTransaction(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to register contract: %w", err)
	}

	return nil
}

// UpdateRiskScore updates the risk score for a registered contract
func (ec *EthereumClient) UpdateRiskScore(ctx context.Context, contractAddr common.Address, newScore *big.Int, vulnCount uint32) error {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return fmt.Errorf("VigilumRegistry address not set")
	}

	// Pack function call: updateRiskScore(address,uint256,uint32)
	data, err := ec.identityFirewallABI.Pack("updateRiskScore", contractAddr, newScore, vulnCount)
	if err != nil {
		return fmt.Errorf("failed to pack updateRiskScore call: %w", err)
	}

	// Send transaction
	_, err = ec.sendTransaction(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to update risk score: %w", err)
	}

	return nil
}

// GetRiskScore retrieves the risk score for a contract
func (ec *EthereumClient) GetRiskScore(ctx context.Context, contractAddr common.Address) (*big.Int, error) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return nil, fmt.Errorf("VigilumRegistry address not set")
	}

	// Pack function call: getRiskScore(address)
	data, err := ec.identityFirewallABI.Pack("getRiskScore", contractAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getRiskScore call: %w", err)
	}

	// Call contract
	result, err := ec.callContract(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk score: %w", err)
	}

	// Unpack result
	var score *big.Int
	err = ec.identityFirewallABI.UnpackIntoInterface(&score, "getRiskScore", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %w", err)
	}

	return score, nil
}

// IsBlacklisted checks if a contract is blacklisted
func (ec *EthereumClient) IsBlacklisted(ctx context.Context, contractAddr common.Address) (bool, error) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.identityFirewallAddr == (common.Address{}) {
		return false, fmt.Errorf("VigilumRegistry address not set")
	}

	// Pack function call: isBlacklisted(address)
	data, err := ec.identityFirewallABI.Pack("isBlacklisted", contractAddr)
	if err != nil {
		return false, fmt.Errorf("failed to pack isBlacklisted call: %w", err)
	}

	// Call contract
	result, err := ec.callContract(ctx, data)
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}

	// Unpack result
	var isBlacklisted bool
	err = ec.identityFirewallABI.UnpackIntoInterface(&isBlacklisted, "isBlacklisted", result)
	if err != nil {
		return false, fmt.Errorf("failed to unpack result: %w", err)
	}

	return isBlacklisted, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// MOCK CLIENT (FOR TESTING)
// ═══════════════════════════════════════════════════════════════════════════════

// MockEthereumClient is a mock implementation for testing
type MockEthereumClient struct {
	proofs       map[string]bool // address -> hasValidProof
	proofHashes  map[[32]byte]ProofRecord
	userStatuses map[string]*UserStatus

	totalVerified uint64
	totalUsers    uint64
	totalRevoked  uint64
}

// NewMockEthereumClient creates a new mock client
func NewMockEthereumClient() *MockEthereumClient {
	return &MockEthereumClient{
		proofs:       make(map[string]bool),
		proofHashes:  make(map[[32]byte]ProofRecord),
		userStatuses: make(map[string]*UserStatus),
	}
}

// HasValidProof mock implementation
func (m *MockEthereumClient) HasValidProof(ctx context.Context, userAddr string) (bool, error) {
	return m.proofs[strings.ToLower(userAddr)], nil
}

// SetProofValid sets a user's proof validity (for testing)
func (m *MockEthereumClient) SetProofValid(userAddr string, valid bool) {
	m.proofs[strings.ToLower(userAddr)] = valid
}

// ═══════════════════════════════════════════════════════════════════════════════
// ABI LOADING FROM FILE
// ═══════════════════════════════════════════════════════════════════════════════

// ContractArtifact represents a compiled contract artifact (from Foundry)
type ContractArtifact struct {
	ABI      json.RawMessage `json:"abi"`
	Bytecode struct {
		Object string `json:"object"`
	} `json:"bytecode"`
}

// LoadContractABI loads a contract ABI from a Foundry artifact file
func LoadContractABI(artifactPath string) (abi.ABI, error) {
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to read artifact file: %w", err)
	}

	var artifact ContractArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse artifact: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(artifact.ABI)))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return parsedABI, nil
}
