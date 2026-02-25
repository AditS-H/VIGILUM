# VIGILUM: Next 4 Weeks - Prioritized Implementation Tasks

**Generated:** February 25, 2026  
**Goal:** Achieve MVP state with fully functional multi-engine scanner + basic indexing

---

## 📋 Week 1: Complete Scanner System (CRITICAL)

### Task 1.1: Composite Scanner Orchestrator ⭐ TOP PRIORITY
**File:** `backend/internal/scanner/scanner.go`  
**Lines:** ~200-300 LOC  
**Time:** 4-6 hours

**What to implement:**
```go
// Complete the Orchestrator.ScanAll() method
func (o *Orchestrator) ScanAll(ctx context.Context, contract *domain.Contract, opts *ScanOptions) (*domain.ScanReport, error) {
    // 1. Create background context with timeout
    // 2. Run Slither + Mythril + Static in parallel using errgroup
    // 3. Collect results from all engines
    // 4. Deduplicate findings (same vuln, different detectors)
    // 5. Calculate weighted aggregate risk score
    // 6. Determine overall threat level
    // 7. Return unified ScanReport
}
```

**Acceptance Criteria:**
- [ ] Code compiles
- [ ] All 3+ scanners run in parallel
- [ ] Tests pass (create new test file `scanner_orchestrator_test.go`)
- [ ] Risk score is properly aggregated
- [ ] Error handling for timeouts and failures

---

### Task 1.2: Scanner Result Aggregation Engine
**File:** `backend/internal/scanner/aggregator.go` (NEW)  
**Lines:** ~150 LOC  
**Time:** 2-3 hours

**Implement:**
```go
type ScanAggregator struct {
    findings []domain.Vulnerability
}

// Deduplicate findings from multiple scanners
func (sa *ScanAggregator) Deduplicate() []domain.Vulnerability { }

// Weight results: Slither 40%, Mythril 40%, Static 15%, ML 5%
func (sa *ScanAggregator) CalculateAggregateRiskScore() float64 { }

// Merge metadata from different detectors
func (sa *ScanAggregator) MergeMetadata() map[string]any { }
```

**Acceptance Criteria:**
- [ ] Deduplicates by location + type
- [ ] Weighted scoring algorithm works
- [ ] Confidence scores properly combined
- [ ] Tests verify aggregation logic

---

### Task 1.3: End-to-End Scanner Integration Test
**File:** `backend/internal/scanner/integration_test.go` (NEW)  
**Lines:** ~200 LOC  
**Time:** 2-3 hours

**Test cases:**
- [ ] Single scanner with real contract bytecode
- [ ] Multiple scanners with same contract
- [ ] Timeout handling (scanner exceeds limit)
- [ ] Partial failures (one scanner fails, others continue)
- [ ] Risk score aggregation with multiple vulns
- [ ] Threat level determination thresholds

**Sample Contract for Testing:**
Use a fixture contract with known vulnerabilities:
```solidity
contract TestVuln {
    uint256 x;
    function unsafe() public {
        (bool success,) = msg.sender.call{value: address(this).balance}("");
        // Missing check - unchecked call
        x = 1;
    }
}
```

---

## 📋 Week 2: Fix Tests & Build Indexer Foundation

### Task 2.1: Fix Database Repository Tests 🔴 BLOCKING
**File:** Multiple files in `backend/internal/db/repositories/`  
**Time:** 2-3 hours

**Issues to fix (from TEST_FAILURES_EXPLAINED.md):**
- [ ] Fix return value assignments (Create returns 2 values)
- [ ] Fix ProofHash type: `string` → `[]byte`
- [ ] Fix ProofData dereference: `domain.ProofData` → `&domain.ProofData`
- [ ] Fix field types: int → float64 where needed
- [ ] Re-run tests: `go test ./internal/db/repositories -v`

**Acceptance Criteria:**
- [ ] All repository tests pass
- [ ] No type mismatches remain
- [ ] Test coverage >80% for repositories

---

### Task 2.2: Ethereum Event Listener
**File:** `backend/internal/integration/ethereum.go` (EXTEND/NEW)  
**Lines:** ~300 LOC  
**Time:** 4-6 hours

**Implement:**
```go
type EthereumListener struct {
    client *ethclient.Client
    wsURL  string
}

// Subscribe to new blocks
func (el *EthereumListener) ListenToBlocks(ctx context.Context, ch chan *types.Block) error { }

// Extract contract deployments from transactions
func (el *EthereumListener) ExtractDeployments(block *types.Block) []ContractDeployment { }

// Get bytecode from deployed contract
func (el *EthereumListener) GetBytecode(address common.Address) ([]byte, error) { }
```

**Key Library:** `github.com/ethereum/go-ethereum`

**Acceptance Criteria:**
- [ ] Connects to Ethereum node (Infura/local)
- [ ] Extracts contract deployments from blocks
- [ ] Stores bytecode in database
- [ ] Handles network errors gracefully
- [ ] Performance: <1s per block

---

### Task 2.3: Contract Deployment Indexer Service
**File:** `backend/cmd/indexer/main.go` (EXTEND)  
**Lines:** ~200 LOC  
**Time:** 3-4 hours

**Workflow:**
1. Connect to Ethereum
2. Get current block number
3. Listen for new blocks
4. For each block:
   - Extract deployments
   - Get contract bytecode
   - Store in database
   - Queue for scanning
5. Publish events to NATS

**Acceptance Criteria:**
- [ ] Service starts without errors
- [ ] Tracks current block number
- [ ] Stores contracts in database
- [ ] Publishes scan events
- [ ] Health check endpoint returns OK

---

## 📋 Week 3: ML Model Training & Integration

### Task 3.1: Prepare Training Dataset
**File:** `ml/data/` directory  
**Time:** 1-2 days

**Dataset Collection:**
- [ ] Download 1000+ labeled contracts (vulnerable/safe)
- [ ] Sources: SolidityBench, ContractFuzzer, reentrancy-samples
- [ ] Label distribution: 70% safe, 30% vulnerable
- [ ] Extract bytecode and source code pairs
- [ ] Document dataset in `ml/README.md`

**Acceptance Criteria:**
- [ ] Dataset size: 1000+ contracts
- [ ] Balanced across vulnerability types
- [ ] Reproducible (seeds documented)
- [ ] Stored in `ml/data/processed/`

---

### Task 3.2: Model Training Pipeline
**File:** `ml/src/vigilum_ml/training.py` (COMPLETE)  
**Time:** 4-6 hours

**Implementation:**
```python
def train_model(config: TrainingConfig) -> Tuple[Model, Metrics]:
    # 1. Load dataset
    # 2. Create DataLoaders (train/val/test)
    # 3. Initialize model
    # 4. Train with validation
    # 5. Save best model
    # 6. Generate metrics
    # 7. Return model + metrics
    
trainer = Trainer(model, device='cuda')
history = trainer.train(epochs=100, lr=1e-4)
best_model = trainer.load_best()
metrics = trainer.evaluate()
```

**Acceptance Criteria:**
- [ ] Training completes without errors
- [ ] Validation loss decreases
- [ ] Saves checkpoints every epoch
- [ ] Final accuracy >80% on test set
- [ ] Logs training metrics to file

---

### Task 3.3: ONNX Model Export & Inference
**File:** `ml/scripts/export_onnx.py` (NEW) + `backend/internal/ml/inference_client.go` (NEW)  
**Time:** 3-4 hours

**Python Export:**
```python
def export_to_onnx(model_path: str, output_path: str):
    model = load_model(model_path)
    dummy_input = torch.randn(1, FEATURE_DIM)
    torch.onnx.export(model, dummy_input, output_path)
```

**Go Inference:**
```go
type MLScanner struct {
    model *onnx.Model
}

func (m *MLScanner) Scan(bytecode []byte) (*ScanResult, error) {
    features := ExtractFeatures(bytecode)
    prediction := m.model.Predict(features)
    return prediction, nil
}
```

**Acceptance Criteria:**
- [ ] Model exports to ONNX without errors
- [ ] ONNX file loads in Go
- [ ] Inference produces same results as PyTorch
- [ ] Performance: <100ms per sample

---

## 📋 Week 4: Polish & Prepare for Testnet

### Task 4.1: Temporal Workflow Implementation
**File:** `backend/internal/temporal/workflows.go` (EXTEND)  
**Time:** 2-3 hours

**Implement workflow:**
```go
func ScanContractWorkflow(ctx workflow.Context, contract *domain.Contract) (*domain.ScanReport, error) {
    // 1. Queue contract for scanning
    // 2. Run orchestrated scan (Slither + Mythril + ML)
    // 3. Store results in database
    // 4. Publish analysis events
    // 5. Return report
}
```

**Acceptance Criteria:**
- [ ] Workflow compiles and runs
- [ ] All activities execute in order
- [ ] Error handling with retries
- [ ] Timeout configuration
- [ ] Temporal UI shows workflow history

---

### Task 4.2: Add Missing Test Coverage
**Files:** Multiple test files  
**Time:** 2-3 hours

**Coverage goals:**
- [ ] `oracle/` tests
- [ ] `redteam/` tests
- [ ] `gateway/` tests
- [ ] `integration/` tests
- [ ] Run: `go test ./... -cover`

**Acceptance Criteria:**
- [ ] Overall coverage >70%
- [ ] No untested code paths
- [ ] All critical paths covered

---

### Task 4.3: Performance Tuning & Optimization
**Time:** 1-2 hours

**Optimize:**
- [ ] Scanner parallelization
- [ ] Database query efficiency
- [ ] Cache configuration (Redis)
- [ ] Docker image size
- [ ] Memory usage

**Benchmarks to target:**
- API response: <200ms
- Scan speed: <30s per contract
- Memory: <500MB per scanner

---

### Task 4.4: Documentation & Deployment Guide
**Files:** `docs/DEPLOYMENT.md`, `docs/API_REFERENCE.md`  
**Time:** 2-3 hours

**Document:**
- [ ] How to deploy to testnet
- [ ] Configuration options
- [ ] API endpoint reference
- [ ] Performance tuning guide
- [ ] Troubleshooting guide

---

### Task 4.5: Security Audit Preparation
**Time:** 2-3 hours

**Reviews:**
- [ ] Smart contract security review
- [ ] Go code security scan (gosec)
- [ ] Dependency vulnerability check (safety)
- [ ] Database permissions audit
- [ ] API authentication audit

---

## 📊 Summary of Week-by-Week Deliverables

| Week | Component | Status | LOC  | Impact |
|------|-----------|--------|------|--------|
| 1 | Composite Scanner | 🔴 > ✅ | 700 | CRITICAL |
| 1 | Aggregation Engine | 🔴 > ✅ | 150 | HIGH |
| 1 | Integration Tests | 🔴 > ✅ | 200 | HIGH |
| 2 | Test Fixes | 🔴 > ✅ | 100 | BLOCKING |
| 2 | Event Listener | 🔴 > ✅ | 300 | HIGH |
| 2 | Indexer Service | 🔴 > ✅ | 200 | HIGH |
| 3 | ML Dataset | 🔴 > ✅ | 1000+ | HIGH |
| 3 | Training Pipeline | 🔴 > ✅ | 150 | HIGH |
| 3 | ONNX Export + Go | 🔴 > ✅ | 200 | HIGH |
| 4 | Temporal Workflows | 🔴 > ✅ | 100 | MEDIUM |
| 4 | Test Coverage | 🔴 > ✅ | 200 | MEDIUM |
| 4 | Performance | 🔴 > ✅ | 50 | MEDIUM |
| 4 | Documentation | 🔴 > ✅ | 100 | MEDIUM |
| 4 | Security | ✅    | 0 | MEDIUM |

**Total:** ~3,500 LOC + 1000+ contracts dataset

---

## 🎯 Success Criteria for MVP

- [ ] Composite scanner runs all engines in parallel
- [ ] Scanner aggregates results correctly
- [ ] Indexer monitors blockchain and triggers scans
- [ ] ML model trained and integrated
- [ ] All core tests passing (>90%)
- [ ] Documentation complete
- [ ] Deployed to testnet
- [ ] Demo shows end-to-end vulnerability detection

---

## 📖 Documentation to Read First

Before starting each task, read:

1. **Week 1:** `EXECUTION_PLAN.md` (Scanner section)
2. **Week 2:** `TEST_FAILURES_EXPLAINED.md` + `DOCKER_SETUP.md`
3. **Week 3:** `ml/README.md` + `SYSTEM_DESIGN.md` (ML section)
4. **Week 4:** `architecture.md` + Contract docs

---

## 🚀 How to Start Right Now

```bash
# 1. Read current status
cat PROJECT_COMPLETION_STATUS_2026.md

# 2. Read task list
cat path/to/this/file

# 3. Pick Task 1.1 (Composite Scanner)
cd backend

# 4. Open the file
code internal/scanner/scanner.go

# 5. Find the TODO comments
grep -n "TODO" internal/scanner/scanner.go

# 6. Implement the ScanAll method

# 7. Write tests
code internal/scanner/scanner_orchestrator_test.go

# 8. Run tests
go test ./internal/scanner -v

# 9. Commit and push
git add .
git commit -m "Phase 14: Composite scanner orchestrator"
```

---

## 💯 Estimated Time Breakdown

| Phase | Duration | Start Date | End Date |
|-------|----------|-----------|---------|
| **Week 1: Scanner** | 5 days | Feb 25 | Mar 1 |
| **Week 2: Indexer** | 5 days | Mar 2 | Mar 8 |
| **Week 3: ML** | 5 days | Mar 9 | Mar 15 |
| **Week 4: Polish** | 5 days | Mar 16 | Mar 22 |
| **Buffer** | 5 days | Mar 23 | Mar 29 |
| **Total to MVP** | 25 days | Feb 25 | Mar 22 |

**Testnet Launch:** Early April 2026

