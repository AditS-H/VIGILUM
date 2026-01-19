# 📚 VIGILUM Learning Resources

A comprehensive guide to all technologies, frameworks, and concepts used in the VIGILUM project. Study these to understand what we've built.

---

## Table of Contents

1. [Go Backend](#1-go-backend)
2. [Solidity Smart Contracts](#2-solidity-smart-contracts)
3. [TypeScript SDK](#3-typescript-sdk)
4. [Python ML Pipeline](#4-python-ml-pipeline)
5. [Noir ZK Circuits](#5-noir-zk-circuits)
6. [Infrastructure](#6-infrastructure)
7. [Core Concepts](#7-core-concepts)
8. [Project-Specific Files](#8-project-specific-files-to-study)

---

## 1. Go Backend

### Language Fundamentals
| Topic | Resource |
|-------|----------|
| **Go Tour** | https://go.dev/tour/ |
| **Effective Go** | https://go.dev/doc/effective_go |
| **Go by Example** | https://gobyexample.com/ |
| **Go Modules** | https://go.dev/blog/using-go-modules |

### Patterns We Used
| Pattern | Resource |
|---------|----------|
| **Project Layout** | https://github.com/golang-standards/project-layout |
| **Clean Architecture** | https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html |
| **Repository Pattern** | https://threedots.tech/post/repository-pattern-in-go/ |
| **Dependency Injection** | https://blog.drewolson.org/dependency-injection-in-go |

### Libraries Used
| Library | Documentation | What We Use It For |
|---------|---------------|-------------------|
| **net/http** | https://pkg.go.dev/net/http | HTTP server & client |
| **httptest** | https://pkg.go.dev/net/http/httptest | Testing HTTP handlers |
| **slog** | https://pkg.go.dev/log/slog | Structured logging |
| **context** | https://pkg.go.dev/context | Request cancellation & timeouts |
| **sync** | https://pkg.go.dev/sync | Mutexes, WaitGroups |
| **time** | https://pkg.go.dev/time | Durations, tickers |
| **encoding/json** | https://pkg.go.dev/encoding/json | JSON serialization |
| **crypto/sha256** | https://pkg.go.dev/crypto/sha256 | Hashing |
| **testing** | https://pkg.go.dev/testing | Unit tests |

### Key Files to Study
```
backend/
├── cmd/api/main.go           # Main API server entry point
├── cmd/api-gateway/main.go   # Reverse proxy gateway
├── cmd/cli/main.go           # CLI tool with cobra-like patterns
├── cmd/scanner/main.go       # Scanner service
├── internal/config/config.go # Configuration management
├── internal/domain/          # Domain entities & repositories
├── internal/middleware/      # HTTP middleware (rate limiting, auth, CORS)
├── internal/scanner/         # Static analysis scanner
```

### Concepts to Master
- [ ] Goroutines and channels
- [ ] Interfaces and dependency injection
- [ ] HTTP middleware chains
- [ ] Context propagation
- [ ] Struct embedding
- [ ] Error handling patterns
- [ ] Table-driven tests

---

## 2. Solidity Smart Contracts

### Language Fundamentals
| Topic | Resource |
|-------|----------|
| **Solidity Docs** | https://docs.soliditylang.org/ |
| **CryptoZombies** | https://cryptozombies.io/ |
| **Solidity by Example** | https://solidity-by-example.org/ |
| **OpenZeppelin Learn** | https://docs.openzeppelin.com/learn/ |

### Foundry (Build Tool)
| Topic | Resource |
|-------|----------|
| **Foundry Book** | https://book.getfoundry.sh/ |
| **Forge Testing** | https://book.getfoundry.sh/forge/tests |
| **Cheatcodes** | https://book.getfoundry.sh/cheatcodes/ |
| **Fuzzing** | https://book.getfoundry.sh/forge/fuzz-testing |

### Security Patterns
| Pattern | Resource |
|---------|----------|
| **Reentrancy Guard** | https://docs.openzeppelin.com/contracts/4.x/api/security#ReentrancyGuard |
| **Access Control** | https://docs.openzeppelin.com/contracts/4.x/access-control |
| **Pausable** | https://docs.openzeppelin.com/contracts/4.x/api/security#Pausable |
| **UUPS Proxy** | https://docs.openzeppelin.com/contracts/4.x/api/proxy#UUPSUpgradeable |

### Key Files to Study
```
contracts/
├── src/
│   ├── IdentityFirewall.sol      # ZK proof verification
│   ├── ThreatOracle.sol          # Risk score oracle
│   ├── VigilumRegistry.sol       # Contract registry
│   └── interfaces/               # Contract interfaces
├── test/
│   ├── IdentityFirewall.t.sol    # Foundry tests
│   └── ThreatOracle.t.sol
```

### Concepts to Master
- [ ] Storage layout and gas optimization
- [ ] Events and indexing
- [ ] Custom errors (gas efficient)
- [ ] Modifiers
- [ ] Access control patterns
- [ ] Proxy patterns (UUPS)
- [ ] Foundry testing with forge

---

## 3. TypeScript SDK

### Language Fundamentals
| Topic | Resource |
|-------|----------|
| **TypeScript Handbook** | https://www.typescriptlang.org/docs/handbook/ |
| **TypeScript Deep Dive** | https://basarat.gitbook.io/typescript/ |
| **Total TypeScript** | https://www.totaltypescript.com/tutorials |

### Libraries Used
| Library | Documentation | What We Use It For |
|---------|---------------|-------------------|
| **viem** | https://viem.sh/ | Ethereum interactions |
| **zod** | https://zod.dev/ | Runtime type validation |
| **vitest** | https://vitest.dev/ | Unit testing |

### Viem Specifics
| Topic | Resource |
|-------|----------|
| **Getting Started** | https://viem.sh/docs/getting-started |
| **Public Client** | https://viem.sh/docs/clients/public |
| **Wallet Client** | https://viem.sh/docs/clients/wallet |
| **Contract Interactions** | https://viem.sh/docs/contract/readContract |
| **ABI Types** | https://viem.sh/docs/glossary/types#abi |

### Key Files to Study
```
sdk/ts-sdk/
├── src/
│   ├── client.ts      # VigilumClient - main entry point
│   ├── analyzer.ts    # ContractAnalyzer - scanning logic
│   ├── guard.ts       # TransactionGuard - pre-tx checks
│   ├── monitor.ts     # ContractMonitor - event watching
│   ├── hooks.ts       # React hooks factory
│   ├── types.ts       # Zod schemas & types
│   └── constants.ts   # Contract addresses & ABIs
├── tests/
│   └── *.test.ts      # Vitest tests
```

### Concepts to Master
- [ ] Generic types and type inference
- [ ] Discriminated unions
- [ ] Zod schema validation
- [ ] Async/await patterns
- [ ] Event emitters
- [ ] Factory patterns
- [ ] Module exports

---

## 4. Python ML Pipeline

### Language Fundamentals
| Topic | Resource |
|-------|----------|
| **Python Tutorial** | https://docs.python.org/3/tutorial/ |
| **Real Python** | https://realpython.com/ |
| **Python Type Hints** | https://docs.python.org/3/library/typing.html |

### ML Libraries Used
| Library | Documentation | What We Use It For |
|---------|---------------|-------------------|
| **PyTorch** | https://pytorch.org/docs/stable/ | Neural networks |
| **NumPy** | https://numpy.org/doc/ | Numerical computing |
| **Polars** | https://pola.rs/ | DataFrame operations |
| **ONNX** | https://onnx.ai/onnx/intro/ | Model export |
| **pytest** | https://docs.pytest.org/ | Testing |

### Deep Learning Concepts
| Topic | Resource |
|-------|----------|
| **Transformers** | https://jalammar.github.io/illustrated-transformer/ |
| **Attention Mechanism** | https://lilianweng.github.io/posts/2018-06-24-attention/ |
| **PyTorch Tutorials** | https://pytorch.org/tutorials/ |
| **ONNX Export** | https://pytorch.org/docs/stable/onnx.html |

### Key Files to Study
```
ml/
├── src/vigilum_ml/
│   ├── model.py       # VulnerabilityDetector neural network
│   ├── models.py      # ModelConfig, TrainingConfig
│   ├── features.py    # Feature extraction from bytecode
│   ├── dataset.py     # PyTorch Dataset class
│   ├── training.py    # Training loop & optimization
│   └── inference/     # ONNX inference
├── tests/
│   └── test_*.py      # pytest tests
```

### Concepts to Master
- [ ] PyTorch nn.Module
- [ ] Transformer architecture
- [ ] Embeddings (byte, positional)
- [ ] Multi-head attention
- [ ] Loss functions (BCE, CrossEntropy)
- [ ] Optimizers (AdamW)
- [ ] Learning rate scheduling
- [ ] ONNX export

---

## 5. Noir ZK Circuits

### Zero-Knowledge Fundamentals
| Topic | Resource |
|-------|----------|
| **ZK Intro** | https://zkintro.com/ |
| **ZK MOOC** | https://zk-learning.org/ |
| **ZK Whiteboard Sessions** | https://www.youtube.com/playlist?list=PLj80z0cJm8QHm_9BdZ1BqcGbgE-BEn-3Y |

### Noir Language
| Topic | Resource |
|-------|----------|
| **Noir Docs** | https://noir-lang.org/docs |
| **Noir by Example** | https://noir-by-example.org/ |
| **Aztec Tutorials** | https://docs.aztec.network/tutorials |

### Key Files to Study
```
circuits/
├── src/
│   ├── lib.nr              # Main library exports
│   ├── human_proof.nr      # Human behavior ZK proof
│   ├── exploit_proof.nr    # Exploit existence proof
│   ├── proof_of_audit.nr   # Audit verification
│   ├── reputation.nr       # Reputation proofs
│   └── gadgets/            # Reusable circuit components
├── Nargo.toml              # Project configuration
```

### Concepts to Master
- [ ] Field elements and arithmetic
- [ ] Constraints and witness generation
- [ ] Public vs private inputs
- [ ] Pedersen hash
- [ ] Range proofs
- [ ] Circuit compilation
- [ ] Proof generation and verification

---

## 6. Infrastructure

### Docker
| Topic | Resource |
|-------|----------|
| **Docker Docs** | https://docs.docker.com/ |
| **Dockerfile Best Practices** | https://docs.docker.com/develop/develop-images/dockerfile_best-practices/ |
| **Docker Compose** | https://docs.docker.com/compose/ |

### Kubernetes
| Topic | Resource |
|-------|----------|
| **K8s Docs** | https://kubernetes.io/docs/home/ |
| **K8s Basics** | https://kubernetes.io/docs/tutorials/kubernetes-basics/ |
| **Helm** | https://helm.sh/docs/ |

### Databases
| Database | Documentation | What We Use It For |
|----------|---------------|-------------------|
| **PostgreSQL** | https://www.postgresql.org/docs/ | Primary data store |
| **ClickHouse** | https://clickhouse.com/docs | Analytics/time-series |
| **Redis** | https://redis.io/docs/ | Caching & rate limiting |
| **Qdrant** | https://qdrant.tech/documentation/ | Vector similarity |

### Key Files to Study
```
infra/
├── k8s/                    # Kubernetes manifests
├── helm/vigilum-chart/     # Helm chart
├── terraform/              # Infrastructure as code
docker-compose.yml          # Local development
```

---

## 7. Core Concepts

### Blockchain Security
| Topic | Resource |
|-------|----------|
| **Smart Contract Security** | https://consensys.github.io/smart-contract-best-practices/ |
| **SWC Registry** | https://swcregistry.io/ |
| **Damn Vulnerable DeFi** | https://www.damnvulnerabledefi.xyz/ |
| **Ethernaut** | https://ethernaut.openzeppelin.com/ |

### Common Vulnerabilities (What We Detect)
| Vulnerability | Resource |
|---------------|----------|
| **Reentrancy** | https://solidity-by-example.org/hacks/re-entrancy/ |
| **Integer Overflow** | https://solidity-by-example.org/hacks/overflow/ |
| **Access Control** | https://solidity-by-example.org/hacks/accessing-private-data/ |
| **Flash Loans** | https://docs.aave.com/developers/guides/flash-loans |
| **Oracle Manipulation** | https://samczsun.com/so-you-want-to-use-a-price-oracle/ |

### API Design
| Topic | Resource |
|-------|----------|
| **REST API Design** | https://restfulapi.net/ |
| **HTTP Status Codes** | https://httpstatuses.com/ |
| **Rate Limiting** | https://cloud.google.com/architecture/rate-limiting-strategies-techniques |

### Testing
| Topic | Resource |
|-------|----------|
| **Test Pyramid** | https://martinfowler.com/articles/practical-test-pyramid.html |
| **Table Driven Tests (Go)** | https://go.dev/wiki/TableDrivenTests |
| **Mocking** | https://martinfowler.com/articles/mocksArentStubs.html |

---

## 8. Project-Specific Files to Study

### Start Here (Priority Order)
1. **[whole documentation/SYSTEM_DESIGN.md](whole%20documentation/SYSTEM_DESIGN.md)** - Architecture overview
2. **[whole documentation/OBJECT_DESIGN.md](whole%20documentation/OBJECT_DESIGN.md)** - Data models & interfaces
3. **[whole documentation/EXECUTION_PLAN.md](whole%20documentation/EXECUTION_PLAN.md)** - Implementation phases
4. **[README.md](README.md)** - Project overview

### Backend Deep Dive
| File | What You'll Learn |
|------|-------------------|
| [backend/internal/middleware/middleware.go](backend/internal/middleware/middleware.go) | HTTP middleware patterns, recovery, logging |
| [backend/internal/middleware/ratelimit.go](backend/internal/middleware/ratelimit.go) | Sliding window rate limiting |
| [backend/internal/domain/entities.go](backend/internal/domain/entities.go) | Domain modeling in Go |
| [backend/internal/scanner/scanner.go](backend/internal/scanner/scanner.go) | Static analysis patterns |
| [backend/cmd/api-gateway/main.go](backend/cmd/api-gateway/main.go) | Reverse proxy implementation |
| [backend/cmd/cli/main.go](backend/cmd/cli/main.go) | CLI application structure |

### Smart Contracts Deep Dive
| File | What You'll Learn |
|------|-------------------|
| [contracts/src/IdentityFirewall.sol](contracts/src/IdentityFirewall.sol) | ZK verification, access control, events |
| [contracts/src/ThreatOracle.sol](contracts/src/ThreatOracle.sol) | Oracle patterns, batch updates |
| [contracts/src/interfaces/](contracts/src/interfaces/) | Interface-first design |
| [contracts/test/IdentityFirewall.t.sol](contracts/test/IdentityFirewall.t.sol) | Foundry testing patterns |

### SDK Deep Dive
| File | What You'll Learn |
|------|-------------------|
| [sdk/ts-sdk/src/client.ts](sdk/ts-sdk/src/client.ts) | Client architecture, retry logic |
| [sdk/ts-sdk/src/analyzer.ts](sdk/ts-sdk/src/analyzer.ts) | Contract analysis, caching |
| [sdk/ts-sdk/src/guard.ts](sdk/ts-sdk/src/guard.ts) | Transaction simulation |
| [sdk/ts-sdk/src/types.ts](sdk/ts-sdk/src/types.ts) | Zod schemas, TypeScript patterns |
| [sdk/ts-sdk/src/hooks.ts](sdk/ts-sdk/src/hooks.ts) | React hooks factory |

### ML Deep Dive
| File | What You'll Learn |
|------|-------------------|
| [ml/src/vigilum_ml/model.py](ml/src/vigilum_ml/model.py) | Transformer architecture |
| [ml/src/vigilum_ml/features.py](ml/src/vigilum_ml/features.py) | Feature engineering |
| [ml/src/vigilum_ml/training.py](ml/src/vigilum_ml/training.py) | Training loops, optimization |

### ZK Circuits Deep Dive
| File | What You'll Learn |
|------|-------------------|
| [circuits/src/human_proof.nr](circuits/src/human_proof.nr) | Behavioral proof constraints |
| [circuits/src/exploit_proof.nr](circuits/src/exploit_proof.nr) | Exploit verification |

---

## 📖 Recommended Learning Path

### Week 1: Foundations
- [ ] Complete Go Tour
- [ ] Read Solidity by Example
- [ ] TypeScript Handbook basics

### Week 2: Project Architecture
- [ ] Read SYSTEM_DESIGN.md thoroughly
- [ ] Study OBJECT_DESIGN.md models
- [ ] Understand the data flow

### Week 3: Smart Contracts
- [ ] Foundry Book chapters 1-5
- [ ] Study IdentityFirewall.sol
- [ ] Run and understand all contract tests

### Week 4: Backend Services
- [ ] Study middleware implementation
- [ ] Understand rate limiting code
- [ ] Review API Gateway proxy logic

### Week 5: SDK Development
- [ ] Learn viem basics
- [ ] Study client.ts patterns
- [ ] Understand Zod validation

### Week 6: ML & ZK
- [ ] PyTorch transformer tutorial
- [ ] Noir documentation
- [ ] Study human_proof.nr circuit

---

## 🎯 Quick Reference

### Run All Tests
```bash
# Smart Contracts (86 tests)
cd contracts && forge test

# ML Pipeline (29 tests)
cd ml && python -m pytest

# TypeScript SDK (37 tests)
cd sdk/ts-sdk && npm test

# Go Backend (45+ tests)
cd backend && go test ./...

# Total: 197+ tests
```

### Key Commands
```bash
# Build contracts
forge build

# Deploy locally
anvil  # Start local chain
forge script script/Deploy.s.sol --fork-url http://localhost:8545

# Run ML training
python -m vigilum_ml.training --config config.yaml

# Build SDK
npm run build

# Run backend
go run ./cmd/api/main.go
```

---

## 📚 Books (Optional Deep Dives)

| Book | Topic |
|------|-------|
| *Mastering Ethereum* | Blockchain fundamentals |
| *The Go Programming Language* | Go mastery |
| *Designing Data-Intensive Applications* | System design |
| *Deep Learning with PyTorch* | ML/DL fundamentals |

---

*Last updated: January 19, 2026*
*Total tests passing: 197*
*Phases completed: 1-7*
