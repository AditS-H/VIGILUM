# Multi-stage Dockerfile for VIGILUM Backend
# Stage 1: Security Analysis Tools Installation
FROM python:3.12-slim AS security-tools

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install Slither (Static Analysis)
RUN pip install --no-cache-dir slither-analyzer==0.10.0

# Install Mythril (Symbolic Execution) - will be used in Phase 14
RUN pip install --no-cache-dir mythril==0.24.0

# Verify installations
RUN slither --version
RUN myth version

# =============================================================================
# Stage 2: Go Builder
# =============================================================================
FROM golang:1.23-alpine AS go-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all services
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/scanner ./cmd/scanner
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/indexer ./cmd/indexer
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/admin-cli ./cmd/admin-cli

# =============================================================================
# Stage 3: Final Runtime Image
# =============================================================================
FROM python:3.12-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy security tools from stage 1
COPY --from=security-tools /usr/local/bin/slither /usr/local/bin/slither
COPY --from=security-tools /usr/local/bin/myth /usr/local/bin/myth
COPY --from=security-tools /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages

# Copy Go binaries from stage 2
COPY --from=go-builder /bin/api /app/api
COPY --from=go-builder /bin/scanner /app/scanner
COPY --from=go-builder /bin/indexer /app/indexer
COPY --from=go-builder /bin/admin-cli /app/admin-cli

# Create working directories
RUN mkdir -p /tmp/vigilum-slither /tmp/vigilum-mythril

# Set environment variables
ENV SLITHER_PATH=/usr/local/bin/slither
ENV MYTHRIL_PATH=/usr/local/bin/myth
ENV WORK_DIR=/tmp/vigilum-slither

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command (API server)
CMD ["/app/api"]
