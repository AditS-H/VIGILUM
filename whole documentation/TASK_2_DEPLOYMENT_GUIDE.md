# Task #2: Deploy VigilumRegistry Smart Contract

## Prerequisites

1. **Foundry Installed** ✅
   - Verified: forge 1.5.1-v1.5.1

2. **Ethereum Wallet with Testnet ETH**
   - For Sepolia testnet deployment
   - Get Sepolia ETH from: https://sepoliafaucet.com/

3. **RPC URL**
   - Infura: https://infura.io/
   - Alchemy: https://www.alchemy.com/
   - Or public Sepolia RPC: https://rpc.sepolia.org

---

## Option 1: Deploy to Sepolia Testnet (Recommended)

### Step 1: Set Environment Variables

Create a `.env` file in `contracts/` directory:

```bash
# Required
PRIVATE_KEY=your_private_key_here_without_0x
RPC_URL=https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY

# Optional (for contract verification on Etherscan)
ETHERSCAN_API_KEY=your_etherscan_api_key
```

**Security Warning:** Never commit `.env` file to git! Already in `.gitignore`.

### Step 2: Compile the Contract

```bash
cd e:\Hacking\VIGILUM\contracts
forge build
```

Expected output:
```
[⠊] Compiling...
[⠃] Compiling 1 files with 0.8.28
[⠒] Solc 0.8.28 finished in X.XXs
Compiler run successful!
```

### Step 3: Deploy to Sepolia

```bash
cd e:\Hacking\VIGILUM\contracts

# Load environment variables (PowerShell)
Get-Content .env | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]+)\s*=\s*(.+)\s*$') {
        [System.Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process')
    }
}

# Deploy
forge script script/DeployVigilumRegistry.s.sol:DeployVigilumRegistry \
  --rpc-url $env:RPC_URL \
  --private-key $env:PRIVATE_KEY \
  --broadcast
```

**With Etherscan Verification:**
```bash
forge script script/DeployVigilumRegistry.s.sol:DeployVigilumRegistry \
  --rpc-url $env:RPC_URL \
  --private-key $env:PRIVATE_KEY \
  --broadcast \
  --verify \
  --etherscan-api-key $env:ETHERSCAN_API_KEY
```

### Step 4: Save Contract Address

From deployment output, find:
```
VigilumRegistry deployed at: 0xABCDEF...
```

**Add to backend config:**

```bash
# In backend/.env or export directly
$env:IDENTITY_FIREWALL_ADDRESS = "0xABCDEF..."
```

Or update `backend/internal/config/config.go` default value.

---

## Option 2: Deploy to Local Anvil (Testing)

### Start Anvil (Local Ethereum Node)

```bash
anvil
```

This starts a local Ethereum node with:
- Chain ID: 31337
- 10 pre-funded accounts
- RPC: http://localhost:8545

### Deploy to Local Node

```bash
cd e:\Hacking\VIGILUM\contracts

forge script script/DeployVigilumRegistry.s.sol:DeployVigilumRegistry \
  --rpc-url http://localhost:8545 \
  --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
  --broadcast
```

**Note:** The private key above is Anvil's default test account #0.

---

## Option 3: Dry Run (Simulation)

Test deployment without broadcasting:

```bash
forge script script/DeployVigilumRegistry.s.sol:DeployVigilumRegistry \
  --rpc-url $env:RPC_URL
```

This simulates the deployment and shows gas estimates without spending real ETH.

---

## Post-Deployment Tasks

### 1. Verify Deployment

Check contract on Etherscan:
```
https://sepolia.etherscan.io/address/YOUR_CONTRACT_ADDRESS
```

### 2. Update Backend Configuration

Edit `backend/cmd/api/main.go` or set environment variable:

```powershell
# PowerShell
$env:ETH_RPC_URL = "https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY"
$env:IDENTITY_FIREWALL_ADDRESS = "0xYOUR_DEPLOYED_ADDRESS"
```

### 3. Test Contract Integration

Start backend and verify Ethereum client connects:

```bash
cd e:\Hacking\VIGILUM\backend
$env:CGO_ENABLED="0"
$env:VIGILUM_ENV="development"
$env:ETH_RPC_URL="https://rpc.sepolia.org"
$env:IDENTITY_FIREWALL_ADDRESS="0xYOUR_ADDRESS"

go run ./cmd/api/main.go
```

Expected log output:
```json
{"level":"INFO","msg":"Ethereum client initialized","chain_id":"11155111","contract":"0x..."}
```

### 4. Authorize Backend as Oracle

The deployer is automatically authorized as an oracle. To authorize the backend wallet:

```bash
# Using cast (Foundry CLI tool)
cast send YOUR_CONTRACT_ADDRESS \
  "authorizeOracle(address)" \
  YOUR_BACKEND_WALLET_ADDRESS \
  --rpc-url $env:RPC_URL \
  --private-key $env:PRIVATE_KEY
```

---

## Troubleshooting

### Error: "Insufficient funds"

**Solution:** Get Sepolia ETH from faucet:
- https://sepoliafaucet.com/
- https://www.alchemy.com/faucets/ethereum-sepolia

### Error: "Failed to connect to RPC"

**Solution:** Check RPC URL is correct and API key is valid.

### Error: "Nonce too low"

**Solution:** Reset nonce:
```bash
forge script ... --broadcast --legacy
```

### Contract Not Verified on Etherscan

**Manual Verification:**
```bash
forge verify-contract \
  YOUR_CONTRACT_ADDRESS \
  src/VigilumRegistry.sol:VigilumRegistry \
  --chain-id 11155111 \
  --etherscan-api-key $env:ETHERSCAN_API_KEY
```

---

## Success Criteria

- [ ] Contract deployed to Sepolia testnet
- [ ] Contract address saved (0x...)
- [ ] Contract verified on Etherscan (optional)
- [ ] Backend ENV variable set (ETH_RPC_URL, IDENTITY_FIREWALL_ADDRESS)
- [ ] Backend logs show "Ethereum client initialized"
- [ ] Backend wallet authorized as oracle on contract

---

## Next Steps

After successful deployment:

→ **Task #4:** Implement Proof Registry on Chain
   - Wire backend to call `registerContract()` and `updateRiskScore()`
   - Create proof verification flow that writes to blockchain

---

## Deployment Record

**Contract:** VigilumRegistry  
**Network:** _______________  
**Address:** _______________  
**Deployer:** _______________  
**Block:** _______________  
**Timestamp:** _______________  
**Gas Used:** _______________  
**Etherscan:** _______________  
