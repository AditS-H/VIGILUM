# VIGILUM Docker Architecture & Complete Guide

## Table of Contents
1. [System Architecture Overview](#system-architecture-overview)
2. [Service Descriptions](#service-descriptions)
3. [Service Connections](#service-connections)
4. [Port Mappings](#port-mappings)
5. [Data Flows](#data-flows)
6. [Learning Resources](#learning-resources)
7. [Quick Start & Commands](#quick-start--commands)
8. [Troubleshooting](#troubleshooting)

---

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    VIGILUM DOCKER ECOSYSTEM                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │          FRONTEND (React/Vite) - Port 3000+              │   │
│  │                                                           │   │
│  │    (Not in docker-compose, runs separately)              │   │
│  └──────────────────────────────────────────────────────────┘   │
│                           ↓                                       │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │    BACKEND API (Go) - Port 8080                          │   │
│  │    - REST API endpoints                                  │   │
│  │    - WebSocket connections                              │   │
│  │    - Proof verification                                 │   │
│  │    - Contract scanning                                  │   │
│  └──────────────────────────────────────────────────────────┘   │
│         ↓              ↓              ↓              ↓            │
│  ┌────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │PostgreSQL│  │  Redis   │  │ Qdrant   │  │  NATS    │           │
│  │ Port 5432│  │ Port 6379│  │ Port 6333│  │ Port 4222│          │
│  └────────┘  └──────────┘  └──────────┘  └──────────┘           │
│         ↓                                                          │
│  ┌──────────────┐                                                 │
│  │  ClickHouse  │                                                 │
│  │  Port 8123   │                                                 │
│  └──────────────┘                                                 │
│         ↓                                                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │     WORKFLOW ORCHESTRATION (Temporal)                    │   │
│  │     - Port 7233 (Backend)                                │   │
│  │     - Port 8081 (UI)                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │     OBSERVABILITY STACK                                  │   │
│  │     ┌──────────┐  ┌──────────┐  ┌──────────┐            │   │
│  │     │ Jaeger   │  │Prometheus│  │ Grafana  │            │   │
│  │     │ 16686    │  │  9090    │  │  3000    │            │   │
│  │     └──────────┘  └──────────┘  └──────────┘            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│              [VIGILUM NETWORK - Bridge Driver]                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Service Descriptions

### 1. PostgreSQL (Port 5432)

**Purpose:** Primary relational database for structured data

**What it stores:**
- User accounts and wallets
- Smart contract metadata
- Vulnerability findings
- Transaction history
- Threat signals
- Bug bounty submissions
- Zero-knowledge proof records

**Docker Image:** `postgres:17-alpine`

**Environment Variables:**
```yaml
POSTGRES_USER: vigilum
POSTGRES_PASSWORD: vigilum
POSTGRES_DB: vigilum
```

**Data Persistence:**
```yaml
volumes:
  - postgres_data:/var/lib/postgresql/data
```

**Connection String (from backend):**
```
Host: postgres
Port: 5432
User: vigilum
Password: vigilum
Database: vigilum
```

---

### 2. Redis (Port 6379)

**Purpose:** In-memory cache & session store

**What it caches:**
- User sessions & authentication tokens
- API response cache (60-300s TTL)
- Rate limiting counters
- Real-time notifications queue
- Proof challenge cache
- Temporary workflow state

**Docker Image:** `redis:8-alpine`

**Key Features:**
- Blazing-fast memory access (microseconds)
- Automatic expiration (TTL support)
- Pub/Sub messaging
- Lua scripting support

**Connection URL (from backend):**
```
redis://redis:6379
```

**Common Commands:**
```bash
# Connect to Redis CLI
docker exec -it vigilum-redis redis-cli

# Check cache size
redis-cli INFO memory

# Flush cache
redis-cli FLUSHALL

# Monitor all commands
redis-cli MONITOR
```

---

### 3. Qdrant (Port 6333 & 6334)

**Purpose:** Vector database for ML embeddings & semantic search

**What it stores:**
- Contract bytecode embeddings (768D vectors)
- Vulnerability pattern embeddings
- Malware genome fingerprints
- Semantic similarity scores

**Docker Image:** `qdrant/qdrant:latest`

**Ports:**
- `6333`: HTTP API
- `6334`: gRPC API

**Data Persistence:**
```yaml
volumes:
  - qdrant_data:/qdrant/storage
```

**Use Cases in VIGILUM:**
1. **Similarity Search** - Find similar contracts to flagged ones
2. **Anomaly Detection** - Identify unusual bytecode patterns
3. **Malware Clustering** - Group similar malicious contracts
4. **Fast Retrieval** - Nearest-neighbor search in microseconds

**Connection URL (from backend):**
```
http://qdrant:6333
```

**API Examples:**
```bash
# Health check
curl http://localhost:6333/health

# List collections
curl http://localhost:6333/collections

# Search similar vectors
curl -X POST http://localhost:6333/collections/contracts/points/search \
  -H "Content-Type: application/json" \
  -d '{"vector": [0.1, 0.2, ...], "limit": 10}'
```

---

### 4. ClickHouse (Port 8123 & 9000)

**Purpose:** Time-series analytics database for large-scale data analysis

**What it stores:**
- Contract scan history (time-indexed)
- Threat metrics over time
- Vulnerability statistics
- Performance metrics
- Audit logs

**Docker Image:** `clickhouse/clickhouse-server:latest`

**Ports:**
- `8123`: HTTP API
- `9000`: Native protocol (faster)

**Data Persistence:**
```yaml
volumes:
  - clickhouse_data:/var/lib/clickhouse
```

**Key Features:**
- Column-oriented storage (100x faster analytics)
- Compression ratio 10:1
- Handles billions of rows
- Real-time aggregations

**Connection URL (from backend):**
```
http://localhost:8123 (HTTP)
localhost:9000 (Native)
```

---

### 5. NATS (Port 4222 & 8222)

**Purpose:** High-performance message broker for event streaming

**What it carries:**
- Proof verification events
- Contract scan events
- Threat detection events
- User notifications
- DAO governance updates

**Docker Image:** `nats:latest`

**Ports:**
- `4222`: NATS protocol
- `8222`: Monitoring endpoint

**Message Topics:**
```
proof.generated      → Proof challenge created
proof.submitted      → Proof submitted for verification
proof.verified       → Proof verification complete

contract.scanned     → Contract scanning complete
contract.flagged     → High-risk contract identified

threat.detected      → Threat signal generated
threat.published     → Signal published on-chain

exploit.submitted    → Bug bounty submission
exploit.verified     → Exploit verified
```

**Connection URL (from backend):**
```
nats://nats:4222
```

**Monitor NATS:**
```bash
# Health check
curl http://localhost:8222/varz

# List subjects
curl http://localhost:8222/subsz
```

---

### 6. Temporal (Port 7233)

**Purpose:** Distributed workflow orchestration engine

**What it manages:**
- Long-running proof verification workflows
- Complex contract scanning pipelines
- Multi-step threat analysis processes
- Scheduled vulnerability audits
- Retry logic with exponential backoff
- State persistence across failures

**Docker Image:** `temporalio/auto-setup:latest`

**Connection Details:**
```
Host: temporal
Port: 7233
Database: PostgreSQL (shared with backend)
```

**Workflow Examples:**
```
ProofVerificationWorkflow:
  Step 1: Generate Challenge (5s timeout)
  Step 2: Wait for Proof Submission (2m timeout)
  Step 3: Verify Proof (10s timeout)
  Step 4: Store Result (5s timeout)
  Retry: Exponential backoff, max 3 attempts

ScanContractWorkflow:
  Step 1: Fetch Contract Bytecode (10s timeout)
  Step 2: Extract Features (20s timeout)
  Step 3: Run ML Models (30s timeout)
  Step 4: Check Vulnerabilities (10s timeout)
  Step 5: Publish Results (5s timeout)
```

**Connection URL (from backend):**
```
temporal:7233
```

---

### 7. Temporal UI (Port 8081)

**Purpose:** Web interface for Temporal workflow monitoring

**What you can do:**
- View running workflows
- Check workflow history
- Inspect payload data
- Retry failed workflows
- Monitor task queues

**Access:** `http://localhost:8081`

**Docker Image:** `temporalio/ui:latest`

**Environment:**
```yaml
TEMPORAL_ADDRESS: temporal:7233
```

---

### 8. Jaeger (Port 16686, 4317, 4318)

**Purpose:** Distributed tracing for request tracking

**What it traces:**
- API request flow through services
- Database query latency
- Cache hit/miss ratios
- External API calls
- Proof verification timing
- ML inference duration

**Docker Image:** `jaegertracing/all-in-one:latest`

**Ports:**
- `16686`: Web UI
- `4317`: gRPC receiver
- `4318`: HTTP receiver

**Trace Examples:**
```
GET /api/contracts/0x123
├── PostgreSQL: SELECT (2ms)
├── Redis check cache (0.5ms)
├── Qdrant similarity search (45ms)
├── ML model inference (120ms)
└── Cache set (1ms)
Total: 168.5ms
```

**Access:** `http://localhost:16686`

**Key Metrics:**
- **P99 Latency** - 99th percentile response time
- **Error Rate** - % of failed requests
- **Service Topology** - How services call each other
- **Span Duration** - Individual operation timing

---

### 9. Prometheus (Port 9090)

**Purpose:** Metrics collection and time-series storage

**What it collects:**
- API request metrics (count, latency, errors)
- Database connection pool stats
- Cache hit/miss rates
- Memory usage
- CPU utilization
- Custom business metrics

**Docker Image:** `prom/prometheus:latest`

**Configuration:** `./config/prometheus.yml`

**Access:** `http://localhost:9090`

**Query Examples:**
```promql
# Request rate (req/sec)
rate(vigilum_requests_total[5m])

# Error rate
rate(vigilum_errors_total[5m])

# API latency (95th percentile)
histogram_quantile(0.95, vigilum_request_duration_ms)

# Cache hit ratio
vigilum_cache_hits / (vigilum_cache_hits + vigilum_cache_misses)
```

---

### 10. Grafana (Port 3000)

**Purpose:** Visualization and dashboarding for metrics

**What it shows:**
- Real-time system health
- Request volume & latency trends
- Error rate alerts
- Database performance
- Cache efficiency
- Proof verification success rate

**Docker Image:** `grafana/grafana:latest`

**Credentials:**
```
Username: admin
Password: admin (Change on first login!)
```

**Access:** `http://localhost:3000`

**Default Dashboards:**
1. **System Overview** - CPU, Memory, Network
2. **API Performance** - Request metrics
3. **Database Health** - Connections, queries
4. **Cache Efficiency** - Hit rates, evictions
5. **Business Metrics** - Proofs verified, contracts scanned

**Data Source:** Prometheus (http://prometheus:9090)

---

## Service Connections

### Backend → PostgreSQL
```
Purpose: CRUD operations for all persistent data
Protocol: TCP (PostgreSQL native)
Connection: backend:5432 → postgres:5432
Typical Operations:
  - Create user accounts
  - Store contract analysis results
  - Save vulnerability records
  - Query historical data
Timeout: 30s (default)
Connection Pool: 20 connections (configurable)
```

### Backend → Redis
```
Purpose: Cache, sessions, rate limiting
Protocol: TCP (Redis)
Connection: backend:6379 → redis:6379
Typical Operations:
  - Get/Set user sessions (TTL: 24h)
  - Cache API responses (TTL: 5m)
  - Store rate limit counters (TTL: 1h)
  - Queue notifications
Timeout: 5s
Key Prefix: vigilum:{context}:{identifier}
```

### Backend → Qdrant
```
Purpose: Semantic search for contract similarity
Protocol: HTTP REST or gRPC
Connection: backend:6333 → qdrant:6333
Typical Operations:
  - Insert contract embeddings
  - Search similar contracts
  - Update malware fingerprints
  - Cluster analysis
Timeout: 30s
Vector Dimensions: 768 (transformer embeddings)
Similarity Metric: Cosine distance
```

### Backend → ClickHouse
```
Purpose: Time-series analytics
Protocol: HTTP or Native TCP
Connection: backend:8123 → clickhouse:8123
Typical Operations:
  - Insert scan events
  - Query historical trends
  - Generate reports
  - Aggregate statistics
Timeout: 60s (data insertion is slower)
Batch Size: 1000 rows per insert
```

### Backend → NATS
```
Purpose: Event-driven messaging
Protocol: NATS (publish-subscribe)
Connection: backend:4222 → nats:4222
Typical Operations:
  - Publish proof events
  - Subscribe to threat alerts
  - Fan-out notifications
  - Cross-service communication
Timeout: 5s per message
Message Format: JSON
QoS: At-least-once delivery (with persistence)
```

### Backend → Temporal
```
Purpose: Workflow orchestration
Protocol: gRPC
Connection: backend:7233 → temporal:7233
Typical Operations:
  - Start proof verification workflow
  - Monitor workflow status
  - Handle retries
  - Persist workflow state
Timeout: Based on workflow definition
Worker Threads: 4 (configurable)
Task Queue: proof-verification, contract-scanning
```

### Temporal → PostgreSQL
```
Purpose: Workflow persistence
Protocol: TCP (PostgreSQL)
Connection: temporal:5432 → postgres:5432
Typical Operations:
  - Store workflow execution history
  - Persist task state
  - Track completion status
  - Enable recovery on failure
Timeout: 30s
```

### Jaeger ← All Services
```
Purpose: Distributed tracing
Protocol: OTLP (OpenTelemetry Protocol)
Receivers: gRPC (4317), HTTP (4318)
Typical Operations:
  - Receive trace spans
  - Correlate across services
  - Store traces
  - Enable debugging
Timeout: 5s
Batch Size: 100 spans
```

### Prometheus ← All Services
```
Purpose: Metrics scraping
Protocol: HTTP
Scrape Interval: 15s (configurable)
Typical Operations:
  - Scrape /metrics endpoint
  - Store time-series data
  - Run alerting rules
  - Enable dashboards
Timeout: 10s
Retention: 15 days (configurable)
```

### Grafana → Prometheus
```
Purpose: Query metrics for visualization
Protocol: HTTP
Connection: grafana:9090 → prometheus:9090
Typical Operations:
  - Query time-series data
  - Plot graphs
  - Set alerts
  - Create dashboards
Timeout: 30s
```

---

## Port Mappings

| Service | Container Port | Host Port | Protocol | Purpose |
|---------|----------------|-----------|----------|---------|
| **Backend** | 8080 | 8080 | HTTP/WS | REST API, WebSocket |
| **PostgreSQL** | 5432 | 5432 | TCP | Database |
| **Redis** | 6379 | 6379 | TCP | Cache |
| **Qdrant HTTP** | 6333 | 6333 | HTTP/REST | Vector Search |
| **Qdrant gRPC** | 6334 | 6334 | gRPC | Vector Search (faster) |
| **ClickHouse HTTP** | 8123 | 8123 | HTTP | Analytics |
| **ClickHouse TCP** | 9000 | 9000 | TCP | Analytics (native) |
| **NATS** | 4222 | 4222 | NATS | Messaging |
| **NATS Monitor** | 8222 | 8222 | HTTP | NATS monitoring |
| **Temporal** | 7233 | 7233 | gRPC | Workflows |
| **Temporal UI** | 8080 | 8081 | HTTP | Workflow UI |
| **Jaeger UI** | 16686 | 16686 | HTTP | Tracing UI |
| **Jaeger gRPC** | 4317 | 4317 | gRPC | Trace receiver |
| **Jaeger HTTP** | 4318 | 4318 | HTTP | Trace receiver |
| **Prometheus** | 9090 | 9090 | HTTP | Metrics |
| **Grafana** | 3000 | 3000 | HTTP | Dashboards |

---

## Data Flows

### 1. Smart Contract Scanning Flow

```
User Request
    ↓
Backend API (8080)
    ├─ → PostgreSQL: Store contract info
    ├─ → Redis: Cache contract metadata
    └─ → NATS: Publish contract.scanned event
         ↓
    Temporal Workflow (7233)
    ├─ Step 1: Fetch bytecode
    ├─ Step 2: Extract features
    ├─ Step 3: Run ML models
    ├─ Step 4: Check against vectors in Qdrant
    ├─ Step 5: Analyze results
    ├─ → PostgreSQL: Store findings
    ├─ → ClickHouse: Store scan history
    └─ → NATS: Publish contract.flagged event
         ↓
    Metrics Collection
    ├─ → Prometheus: scan_duration_ms
    ├─ → Jaeger: trace span
    └─ → Grafana: update dashboard
```

### 2. Zero-Knowledge Proof Verification Flow

```
User Submit Proof
    ↓
Backend API (8080)
    ├─ → PostgreSQL: Store proof record
    ├─ → Redis: Cache challenge
    └─ → NATS: Publish proof.submitted event
         ↓
    Temporal Workflow (7233)
    ├─ Step 1: Retrieve challenge from Redis
    ├─ Step 2: Verify cryptographic proof
    ├─ Step 3: Store verification result
    ├─ → PostgreSQL: Update proof.verified
    ├─ → ClickHouse: Log verification event
    └─ → NATS: Publish proof.verified event
         ↓
    Monitoring
    ├─ → Prometheus: proof_verification_success_rate
    ├─ → Jaeger: end-to-end latency
    └─ → Grafana: update success metric
```

### 3. Threat Signal Publishing Flow

```
Threat Detected
    ↓
Backend API (8080)
    ├─ → PostgreSQL: Store threat signal
    ├─ → Qdrant: Store threat vector embedding
    ├─ → Redis: Cache threat severity
    └─ → NATS: Publish threat.detected event
         ↓
    Analysis Pipeline
    ├─ Aggregation (ClickHouse)
    ├─ Correlation (PostgreSQL)
    ├─ Pattern Detection (Qdrant)
    └─ → NATS: Publish threat.published event
         ↓
    Notification
    ├─ → Redis: Queue user notification
    ├─ → Backend: Send webhook
    └─ Monitoring
         ├─ → Prometheus: threat_count
         ├─ → Jaeger: processing time
         └─ → Grafana: threat dashboard
```

### 4. Monitoring & Observability Flow

```
All Services
    ├─ → Jaeger (4317/4318): Send traces
    ├─ → Prometheus (9090): Export /metrics
    └─ Every 15 seconds:
         Prometheus scrapes all services
         ↓
         Stores time-series data
         ↓
         Runs alerting rules
         ↓
    Grafana
    ├─ Queries Prometheus every 30s
    ├─ Updates dashboard panels
    ├─ Evaluates alert conditions
    └─ Sends notifications if triggered
         ↓
    Jaeger UI (16686)
    ├─ Queries traces
    ├─ Renders service topology
    ├─ Shows span details
    └─ Displays latency breakdown
```

---

## Learning Resources

### PostgreSQL

**Official Documentation:**
- [PostgreSQL Documentation](https://www.postgresql.org/docs/) - Complete reference
- [PostgreSQL Tutorial](https://www.postgresql.org/docs/current/tutorial.html) - Beginner guide

**Online Courses:**
- [PostgreSQL by DataCamp](https://www.datacamp.com/courses/intro-to-sql) - Interactive SQL
- [PostgreSQL Masterclass by Udemy](https://www.udemy.com/course/postgres-12-complete-guide/) - Comprehensive
- [PostgreSQL Performance Tuning](https://www.postgresql.org/docs/current/performance-tips.html) - Official guide

**Best Practices:**
- [PostgreSQL Wiki - Performance](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Indexing Strategies](https://www.postgresql.org/docs/current/sql-createindex.html)
- [Connection Pooling with PgBouncer](https://www.pgbouncer.org/)

**Key Topics for VIGILUM:**
- Transactions (ACID properties)
- Indexes on contract_address, wallet_address
- Partitioning for large tables
- Full-text search for contract analysis

---

### Redis

**Official Documentation:**
- [Redis Documentation](https://redis.io/documentation) - Complete guide
- [Redis Commands Reference](https://redis.io/commands) - All commands explained

**Online Courses:**
- [Redis University - Redis Fundamentals](https://university.redis.com/) - FREE official course
- [Redis by Example](https://redis.io/docs/interact/tutorials/) - Practical examples
- [Udemy Redis Masterclass](https://www.udemy.com/course/redis-the-complete-developer-guide/) - Deep dive

**Best Practices:**
- [Redis Memory Optimization](https://redis.io/docs/management/optimization/memory-optimization/)
- [Cache Invalidation Strategies](https://redis.io/docs/develop/use-cases/caching/)
- [Pub/Sub Patterns](https://redis.io/docs/interact/pubsub/)

**Key Topics for VIGILUM:**
- Session storage (login tokens)
- Cache expiration (TTL)
- Rate limiting with INCR
- Pub/Sub for real-time updates
- Lua scripting for atomic operations

---

### Qdrant

**Official Documentation:**
- [Qdrant Official Docs](https://qdrant.tech/documentation/) - Complete guide
- [Qdrant GitHub](https://github.com/qdrant/qdrant) - Source code & examples

**Online Courses:**
- [Qdrant Vector Search](https://qdrant.tech/documentation/guides/) - Official guides
- [Semantic Search Tutorial](https://qdrant.tech/documentation/tutorials/search-tutorial/) - Hands-on
- [YouTube - Qdrant Basics](https://www.youtube.com/@qdrant) - Video channel

**Architecture:**
- [Vector Database Concepts](https://qdrant.tech/documentation/concepts/) - Theory
- [Similarity Metrics Explained](https://qdrant.tech/documentation/concepts/similarity/) - Distance functions
- [Indexing Strategies](https://qdrant.tech/documentation/indexing/) - Performance

**Key Topics for VIGILUM:**
- Storing contract embeddings (768-dim vectors from transformer)
- Similarity search (cosine distance)
- Vector clustering (malware families)
- Payload filtering (contract metadata)
- HNSW index for fast search

---

### ClickHouse

**Official Documentation:**
- [ClickHouse Documentation](https://clickhouse.com/docs/en/intro) - Complete guide
- [ClickHouse SQL Reference](https://clickhouse.com/docs/en/sql-reference) - All SQL

**Online Courses:**
- [ClickHouse Getting Started](https://clickhouse.com/docs/en/getting-started/) - Quick start
- [ClickHouse Tutorial](https://clickhouse.com/docs/en/tutorial) - Hands-on examples
- [YouTube - ClickHouse Channel](https://www.youtube.com/@ClickHouseDB) - Video tutorials

**Performance Tuning:**
- [Column-Oriented Storage](https://clickhouse.com/docs/en/introduction/distinctive-features) - Why it's fast
- [Compression Codecs](https://clickhouse.com/docs/en/sql-reference/statements/create/table#codecs) - Save space
- [Distributed Queries](https://clickhouse.com/docs/en/engines/table-engines/special/distributed/) - Scale

**Key Topics for VIGILUM:**
- Time-series data storage (scan events, threats)
- ReplacingMergeTree engine (versioned data)
- Aggregating functions (SUM, AVG over time)
- TTL (delete old data automatically)
- Analytical queries (trend analysis)

---

### NATS

**Official Documentation:**
- [NATS Documentation](https://docs.nats.io/) - Complete guide
- [NATS GitHub](https://github.com/nats-io/nats-server) - Source code

**Online Courses:**
- [NATS Getting Started](https://docs.nats.io/getting-started) - Quick start
- [NATS by Example](https://github.com/nats-io/nats.by.example) - Code examples
- [YouTube - NATS Tutorials](https://www.youtube.com/@natsio) - Video channel

**Architecture:**
- [NATS Messaging Patterns](https://docs.nats.io/nats-concepts/core-nats/pubsub) - Pub/Sub basics
- [Request-Reply Pattern](https://docs.nats.io/nats-concepts/core-nats/reqreply) - RPC over messaging
- [Subjects and Wildcards](https://docs.nats.io/nats-concepts/core-nats/pubsub/publish-subscribe) - Topic naming

**Advanced Features:**
- [JetStream](https://docs.nats.io/nats-concepts/jetstream) - Persistence & ordering
- [NATS Clustering](https://docs.nats.io/running-a-nats-service/configuration/clustering) - High availability

**Key Topics for VIGILUM:**
- Event-driven architecture
- Publish-subscribe pattern
- Message subjects (proof.verified, threat.detected)
- Request-reply for sync operations
- JetStream for guaranteed delivery

---

### Temporal

**Official Documentation:**
- [Temporal Documentation](https://docs.temporal.io/) - Complete guide
- [Temporal GitHub](https://github.com/temporalio/temporal) - Source code

**Online Courses:**
- [Temporal Getting Started](https://docs.temporal.io/docs/getting-started) - Quick start
- [Temporal Tutorial](https://learn.temporal.io/) - Interactive learning
- [YouTube - Temporal Channel](https://www.youtube.com/@temporaltech) - Video tutorials

**Concepts:**
- [Workflows vs Activities](https://docs.temporal.io/docs/concepts/activities) - Core concepts
- [Durability & Recovery](https://docs.temporal.io/docs/concepts/durability) - Fault tolerance
- [Task Queues](https://docs.temporal.io/docs/concepts/task-queues) - Worker distribution

**Advanced Topics:**
- [Event Sourcing](https://docs.temporal.io/docs/concepts/durable-execution) - Event-driven workflows
- [Workflow Versioning](https://docs.temporal.io/docs/concepts/versioning) - Evolving workflows
- [Retry Policies](https://docs.temporal.io/docs/concepts/retry-policies) - Resilience

**Key Topics for VIGILUM:**
- Long-running proof verification workflows
- Fault-tolerant contract scanning
- Automatic retries with backoff
- Workflow state persistence
- Parallelizing independent tasks

---

### Jaeger

**Official Documentation:**
- [Jaeger Documentation](https://www.jaegertracing.io/docs/) - Complete guide
- [Jaeger GitHub](https://github.com/jaegertracing/jaeger) - Source code

**Online Courses:**
- [OpenTelemetry Getting Started](https://opentelemetry.io/docs/instrumentation/go/getting-started/) - Foundation
- [Distributed Tracing Concepts](https://www.jaegertracing.io/docs/latest/about/) - Theory
- [YouTube - Jaeger Tutorials](https://www.youtube.com/results?search_query=jaeger+tracing) - Videos

**Integration:**
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/) - For Go backend
- [Trace Context Propagation](https://www.w3.org/TR/trace-context/) - Cross-service tracking

**Key Concepts:**
- [Spans and Traces](https://www.jaegertracing.io/docs/latest/terminology/) - Tracing concepts
- [Sampling Strategies](https://www.jaegertracing.io/docs/latest/sampling/) - Reduce overhead
- [Service Topology](https://www.jaegertracing.io/docs/latest/deployment/) - Visualize dependencies

**Key Topics for VIGILUM:**
- Request tracing through API → DB → Cache
- Identifying performance bottlenecks
- Tracking proof verification latency
- Measuring ML model inference time
- Correlating errors across services

---

### Prometheus

**Official Documentation:**
- [Prometheus Documentation](https://prometheus.io/docs/) - Complete guide
- [Prometheus GitHub](https://github.com/prometheus/prometheus) - Source code

**Online Courses:**
- [Prometheus Getting Started](https://prometheus.io/docs/prometheus/latest/getting_started/) - Quick start
- [Prometheus Querying](https://prometheus.io/docs/prometheus/latest/querying/basics/) - PromQL tutorial
- [YouTube - Prometheus](https://www.youtube.com/results?search_query=prometheus+monitoring) - Video tutorials

**Metrics:**
- [Metric Types](https://prometheus.io/docs/concepts/metric_types/) - Counter, Gauge, Histogram, Summary
- [Best Practices](https://prometheus.io/docs/practices/naming/) - Metric naming & labeling
- [Recording Rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) - Pre-compute metrics

**Querying:**
- [PromQL Tutorial](https://prometheus.io/docs/prometheus/latest/querying/examples/) - Query examples
- [Functions Reference](https://prometheus.io/docs/prometheus/latest/querying/functions/) - All PromQL functions

**Key Topics for VIGILUM:**
- Custom metrics (request count, latency)
- Alert rules (error rate > 5%)
- Time-series aggregation (5m average)
- Histogram percentiles (P99 latency)
- Rate calculations (requests/sec)

---

### Grafana

**Official Documentation:**
- [Grafana Documentation](https://grafana.com/docs/grafana/latest/) - Complete guide
- [Grafana GitHub](https://github.com/grafana/grafana) - Source code

**Online Courses:**
- [Grafana Getting Started](https://grafana.com/grafana/resources/get-started/) - Quick start
- [Building Dashboards](https://grafana.com/docs/grafana/latest/dashboards/) - Dashboard creation
- [YouTube - Grafana Tutorials](https://www.youtube.com/@grafana) - Video channel

**Dashboard Design:**
- [Dashboard Best Practices](https://grafana.com/docs/grafana/latest/dashboards/best-practices/) - Design guidelines
- [Alerting Rules](https://grafana.com/docs/grafana/latest/alerting/) - Setup alerts
- [Dashboard Variables](https://grafana.com/docs/grafana/latest/dashboards/variables/) - Dynamic dashboards

**Key Topics for VIGILUM:**
- Creating system health dashboards
- Real-time metric visualization
- Setting up alerts (Slack, email)
- Dashboard sharing & permissions
- Custom panels & plugins

---

## Quick Start & Commands

### Start All Services

```bash
# Navigate to project root
cd /Hacking/VIGILUM

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes data!)
docker-compose down -v
```

### View Individual Service Logs

```bash
# Backend API
docker-compose logs -f backend

# PostgreSQL
docker-compose logs -f postgres

# Redis
docker-compose logs -f redis

# Qdrant
docker-compose logs -f qdrant

# All services
docker-compose logs -f
```

### Access Service CLIs

```bash
# PostgreSQL psql
docker exec -it vigilum-postgres psql -U vigilum -d vigilum

# Redis CLI
docker exec -it vigilum-redis redis-cli

# NATS CLI
docker exec -it vigilum-nats nats

# Temporal CLI
docker exec -it vigilum-temporal tctl
```

### Service Health Checks

```bash
# Backend API
curl http://localhost:8080/health

# PostgreSQL
docker exec vigilum-postgres pg_isready -U vigilum

# Redis
docker exec vigilum-redis redis-cli ping

# Qdrant
curl http://localhost:6333/health

# NATS
curl http://localhost:8222/varz

# Temporal
curl -X POST http://localhost:7233/health

# Prometheus
curl http://localhost:9090/-/healthy

# Jaeger
curl http://localhost:16686/api/traces
```

### Port Forwarding (If running on remote host)

```bash
# SSH tunnel to access services
ssh -L 8080:localhost:8080 \
    -L 5432:localhost:5432 \
    -L 6379:localhost:6379 \
    -L 6333:localhost:6333 \
    -L 8123:localhost:8123 \
    -L 4222:localhost:4222 \
    -L 7233:localhost:7233 \
    -L 16686:localhost:16686 \
    -L 9090:localhost:9090 \
    -L 3000:localhost:3000 \
    user@remote-host
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs {service_name}

# Check port conflicts
netstat -an | grep {PORT}

# Check disk space
docker system df

# Rebuild container
docker-compose build --no-cache {service_name}
```

### Connectivity Issues

```bash
# Test DNS resolution
docker exec {service_name} nslookup {other_service}

# Test network connectivity
docker exec {service_name} ping {other_service}

# Check network
docker network ls
docker network inspect vigilum-network
```

### Database Issues

```bash
# Check PostgreSQL size
docker exec vigilum-postgres psql -U vigilum -d vigilum -c "\l+"

# Analyze table size
docker exec vigilum-postgres psql -U vigilum -d vigilum -c "\dt+ contracts"

# Vacuum & analyze
docker exec vigilum-postgres psql -U vigilum -d vigilum -c "VACUUM ANALYZE;"
```

### Redis Memory Issues

```bash
# Check memory usage
docker exec vigilum-redis redis-cli INFO memory

# Clear cache
docker exec vigilum-redis redis-cli FLUSHALL

# Eviction policy
docker exec vigilum-redis redis-cli CONFIG GET maxmemory-policy
```

### Performance Slow

```bash
# Check Jaeger traces
# Visit http://localhost:16686

# Check Prometheus metrics
# Visit http://localhost:9090

# Check Grafana dashboards
# Visit http://localhost:3000

# Query slow logs
docker exec vigilum-postgres psql -U vigilum -d vigilum -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
```

---

## Summary

VIGILUM's Docker architecture provides:

✅ **Scalable** - Independent services can scale separately  
✅ **Observable** - Complete tracing, metrics, and dashboards  
✅ **Resilient** - Temporal handles workflow retries & recovery  
✅ **Fast** - Redis cache, Qdrant vector search, ClickHouse analytics  
✅ **Event-Driven** - NATS for async communication  
✅ **Persistent** - PostgreSQL for structured data, volumes for storage  

Every service is documented with official resources for deep learning!

