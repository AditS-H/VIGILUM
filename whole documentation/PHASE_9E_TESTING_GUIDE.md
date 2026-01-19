# Phase 9e: Integration Tests - Testing Guide

## Overview

This guide covers integration testing for the VIGILUM backend repository layer. Integration tests verify that multiple repositories work together correctly to support complete workflows.

## Test Coverage

### 1. User & APIKey Workflow (`TestUserAndAPIKeyWorkflow`)
Tests the complete User → APIKey relationship:
- ✓ Create user with wallet
- ✓ Generate multiple API keys for user
- ✓ Retrieve user's API keys
- ✓ Track API usage (request counting)
- ✓ Reset daily counts
- ✓ Revoke API keys
- ✓ Verify revoked keys are not retrievable

**Expected Outcome:** API keys properly linked to users with full lifecycle management

### 2. User & Human Proof Workflow (`TestUserAndProofWorkflow`)
Tests human proof verification and user risk management:
- ✓ Create user
- ✓ Generate multiple proofs for user
- ✓ Mark proofs as verified
- ✓ Query proofs with pagination
- ✓ Count verified proofs
- ✓ Update user risk score based on proof verification
- ✓ Correlate user reputation with proof status

**Expected Outcome:** Proofs linked to users with verification tracking

### 3. Threat Signal Publishing Workflow (`TestThreatSignalPublishingWorkflow`)
Tests threat aggregation and on-chain publishing:
- ✓ Create threat signals from multiple sources
- ✓ Query unpublished signals
- ✓ Aggregate signals by entity (chain + address)
- ✓ Filter high-risk signals (≥80 risk score)
- ✓ Publish signals to blockchain
- ✓ Verify publication metadata (tx hash, timestamp)
- ✓ Track remaining unpublished signals

**Expected Outcome:** Threats properly aggregated and published with tracking

### 4. Genome Analysis Workflow (`TestGenomeAnalysisWorkflow`)
Tests malware genome clustering and classification:
- ✓ Create reference genome
- ✓ Create similar genomes for clustering
- ✓ Find similar genomes by similarity threshold
- ✓ List genomes by label
- ✓ Get genome distribution statistics
- ✓ Create benign genomes for classification contrast
- ✓ Verify label-based separation

**Expected Outcome:** Genomes clustered with similarity metrics

### 5. Exploit Submission Bounty Workflow (`TestExploitSubmissionBountyWorkflow`)
Tests complete bug bounty lifecycle:
- ✓ Create genome for vulnerability
- ✓ Submit exploits from multiple researchers
- ✓ Query submissions by researcher
- ✓ List pending submissions
- ✓ Verify exploits (auditor action)
- ✓ Pay bounties (treasury action)
- ✓ Track total bounty amounts by status
- ✓ Count submissions per researcher

**Expected Outcome:** Complete bounty workflow from submission to payment

### 6. Multi-Repository Transaction (`TestMultiRepositoryTransaction`)
Tests complex interactions across all repositories:
- ✓ Create complete user ecosystem
- ✓ Generate user API key
- ✓ Create user proof
- ✓ Track API usage
- ✓ Detect threat signals
- ✓ Link all entities by user ID
- ✓ Update user risk based on threats
- ✓ Verify final state consistency

**Expected Outcome:** All repositories stay synchronized

### 7. Error Handling & Edge Cases (`TestErrorHandlingAndEdgeCases`)
Tests error conditions and constraints:
- ✓ NotFound errors for nonexistent records
- ✓ Duplicate prevention (unique constraints)
- ✓ Update nonexistent records fails
- ✓ Delete nonexistent records fails
- ✓ Wallet address uniqueness enforced

**Expected Outcome:** Proper error handling and data integrity

### 8. Performance Benchmarks (`BenchmarkRepositoryOperations`)
Tests performance characteristics:
- ✓ User creation throughput
- ✓ User retrieval latency
- ✓ List operations with filtering
- ✓ Concurrent operation performance

**Expected Outcome:** Operations complete within acceptable latency

## Running Tests

### Prerequisites

1. **Docker & Docker Compose**
   ```bash
   docker --version
   docker-compose --version
   ```

2. **Go 1.24+**
   ```bash
   go version
   ```

3. **PostgreSQL client (optional, for manual inspection)**
   ```bash
   psql --version
   ```

### Quick Start

**Option 1: Using the test runner script**

```bash
# Run all tests
./scripts/run-integration-tests.sh

# Keep containers for inspection
./scripts/run-integration-tests.sh --keep-containers

# Generate coverage report
./scripts/run-integration-tests.sh --coverage

# Custom configuration
./scripts/run-integration-tests.sh \
  --db-host localhost \
  --db-port 5433 \
  --db-user postgres \
  --db-password postgres
```

**Option 2: Manual Docker setup**

```bash
# Start containers
docker-compose -f docker-compose.test.yml up -d

# Wait for PostgreSQL to be ready
docker-compose -f docker-compose.test.yml exec -T postgres-test \
  pg_isready -U postgres -d vigilum_test

# Run tests
cd backend
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_NAME=vigilum_test

go test -v -race ./internal/db/repositories

# Stop containers
docker-compose -f docker-compose.test.yml down
```

**Option 3: Direct PostgreSQL connection**

If you have a PostgreSQL database running:

```bash
cd backend
export TEST_DB_HOST=your-db-host
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=password
export TEST_DB_NAME=vigilum_test

# Run just integration tests
go test -v ./internal/db/repositories -run Integration

# Run specific test
go test -v ./internal/db/repositories -run TestUserAndAPIKeyWorkflow

# Run with coverage
go test -v -cover -coverprofile=coverage.out ./internal/db/repositories
go tool cover -html=coverage.out -o coverage.html
```

## Test Output Interpretation

### Successful Run
```
=== RUN   TestUserAndAPIKeyWorkflow
    integration_test.go:32: ✓ User -> APIKey workflow completed successfully
--- PASS: TestUserAndAPIKeyWorkflow (2.34s)

=== RUN   TestThreatSignalPublishingWorkflow
    integration_test.go:123: ✓ Threat signal publishing workflow completed successfully
--- PASS: TestThreatSignalPublishingWorkflow (1.87s)

PASS
ok      github.com/vigilum/backend/internal/db/repositories   15.23s
```

### Common Failures

**PostgreSQL Connection Error**
```
Failed to open database: dial tcp localhost:5433: connect: connection refused
```
Solution: Ensure Docker containers are running
```bash
docker-compose -f docker-compose.test.yml ps
```

**Database Not Ready**
```
FATAL: database "vigilum_test" does not exist
```
Solution: Give PostgreSQL more time to initialize
```bash
docker-compose -f docker-compose.test.yml logs postgres-test
```

**Duplicate Key Violation**
```
duplicate key value violates unique constraint "users_wallet_address_key"
```
Solution: Clear test data or use fresh container
```bash
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d
```

## Docker Environment

### PostgreSQL Test Container

- **Image:** postgres:17-alpine
- **Host:** localhost (default)
- **Port:** 5433 (default)
- **User:** postgres (default)
- **Password:** postgres (default)
- **Database:** vigilum_test (default)
- **Storage:** `postgres_test_data` volume

### Container Management

```bash
# View containers
docker-compose -f docker-compose.test.yml ps

# View logs
docker-compose -f docker-compose.test.yml logs postgres-test

# Execute PostgreSQL commands
docker-compose -f docker-compose.test.yml exec postgres-test \
  psql -U postgres -d vigilum_test -c "SELECT * FROM users LIMIT 5;"

# Access PostgreSQL shell
docker-compose -f docker-compose.test.yml exec postgres-test \
  psql -U postgres -d vigilum_test

# Stop but keep data
docker-compose -f docker-compose.test.yml stop

# Stop and remove everything
docker-compose -f docker-compose.test.yml down -v
```

## Debugging Tests

### Enable verbose logging
```bash
cd backend
TEST_DEBUG=1 go test -v ./internal/db/repositories
```

### Run specific test
```bash
cd backend
go test -v ./internal/db/repositories -run TestUserAndAPIKeyWorkflow
```

### Run with debugger (Delve)
```bash
cd backend
dlv test ./internal/db/repositories -- -test.v
```

### Inspect database state during test
```bash
# In another terminal while test is running
docker-compose -f docker-compose.test.yml exec postgres-test psql -U postgres -d vigilum_test

# Common queries
\dt                          # List tables
SELECT * FROM users LIMIT 5; # View users
SELECT count(*) FROM api_keys; # Count API keys
```

## Performance Testing

### Run benchmarks
```bash
cd backend
go test -bench=. -benchmem ./internal/db/repositories
```

### Expected Performance

| Operation | Latency | Throughput |
|-----------|---------|-----------|
| UserCreate | < 5ms | 200+ ops/s |
| UserGetByID | < 2ms | 500+ ops/s |
| ListByRiskScore | < 10ms | 100+ ops/s |
| APIKeyCreate | < 5ms | 200+ ops/s |
| ProofMarkVerified | < 3ms | 300+ ops/s |

## Coverage Report

Generate HTML coverage report:

```bash
cd backend
go test -v -coverprofile=coverage.out ./internal/db/repositories
go tool cover -html=coverage.out -o coverage.html
```

View in browser: `open coverage.html` or `xdg-open coverage.html`

## Continuous Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:17-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: vigilum_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run integration tests
        env:
          TEST_DB_HOST: postgres
          TEST_DB_PORT: 5432
          TEST_DB_USER: postgres
          TEST_DB_PASSWORD: postgres
          TEST_DB_NAME: vigilum_test
        run: |
          cd backend
          go test -v -race ./internal/db/repositories
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker-compose -f docker-compose.test.yml logs postgres-test

# Remove and recreate
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d
```

### Port already in use
```bash
# Change port in docker-compose.test.yml or use environment variable
export TEST_DB_PORT=5434
docker-compose -f docker-compose.test.yml up -d
```

### Test timeouts
```bash
# Increase timeout
cd backend
go test -timeout 600s ./internal/db/repositories
```

### Memory issues
```bash
# Limit Docker memory
docker-compose -f docker-compose.test.yml down
docker-compose -f docker-compose.test.yml up -d

# Or modify docker-compose.test.yml
# services:
#   postgres-test:
#     deploy:
#       resources:
#         limits:
#           memory: 512M
```

## Next Steps

After Phase 9e Integration Tests:

1. **Phase 9f:** Performance optimization based on benchmark results
2. **Phase 10:** ZK Prover integration
3. **Phase 11:** API endpoint implementation
4. **Phase 12:** Service layer and business logic

## Resources

- [PostgreSQL Docker Image](https://hub.docker.com/_/postgres)
- [Go Testing Guide](https://golang.org/pkg/testing/)
- [Go Benchmarking](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Docker Compose](https://docs.docker.com/compose/)
