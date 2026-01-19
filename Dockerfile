# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/ .

# Build the API server
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -a -installsuffix cgo -o /build/api ./cmd/api

# Build the scanner
RUN go build -a -installsuffix cgo -o /build/scanner ./cmd/scanner

# Build the indexer
RUN go build -a -installsuffix cgo -o /build/indexer ./cmd/indexer

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl postgresql-client

# Create app user
RUN addgroup -g 1000 vigilum && adduser -D -u 1000 -G vigilum vigilum

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/api /app/
COPY --from=builder /build/scanner /app/
COPY --from=builder /build/indexer /app/

# Copy config files if they exist
COPY backend/config/ /app/config/ 2>/dev/null || true

# Change ownership
RUN chown -R vigilum:vigilum /app

USER vigilum

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command
CMD ["/app/api"]
