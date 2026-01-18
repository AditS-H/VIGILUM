.PHONY: all build test clean docker dev

# ═══════════════════════════════════════════════════════════════════════════
# VARIABLES
# ═══════════════════════════════════════════════════════════════════════════

GOBIN := $(shell go env GOPATH)/bin
FORGE := C:/Users/over9/.foundry/bin/forge.exe

# ═══════════════════════════════════════════════════════════════════════════
# MAIN TARGETS
# ═══════════════════════════════════════════════════════════════════════════

all: build

build: build-backend build-contracts build-sdk

test: test-backend test-contracts test-sdk test-ml

clean:
	rm -rf backend/bin
	rm -rf contracts/out contracts/cache
	rm -rf sdk/ts-sdk/dist sdk/ts-sdk/node_modules
	rm -rf ml/dist ml/.venv

# ═══════════════════════════════════════════════════════════════════════════
# BACKEND (GO)
# ═══════════════════════════════════════════════════════════════════════════

build-backend:
	cd backend && go build -o bin/api ./cmd/api
	cd backend && go build -o bin/scanner ./cmd/scanner
	cd backend && go build -o bin/indexer ./cmd/indexer

test-backend:
	cd backend && go test -v -race ./...

lint-backend:
	cd backend && golangci-lint run

run-api:
	cd backend && go run ./cmd/api

run-scanner:
	cd backend && go run ./cmd/scanner

# ═══════════════════════════════════════════════════════════════════════════
# SMART CONTRACTS (FOUNDRY)
# ═══════════════════════════════════════════════════════════════════════════

build-contracts:
	cd contracts && $(FORGE) build

test-contracts:
	cd contracts && $(FORGE) test -vvv

coverage-contracts:
	cd contracts && $(FORGE) coverage

deploy-local:
	cd contracts && $(FORGE) script script/Deploy.s.sol --rpc-url http://localhost:8545 --broadcast

# ═══════════════════════════════════════════════════════════════════════════
# ZK CIRCUITS (NOIR - WSL)
# ═══════════════════════════════════════════════════════════════════════════

build-circuits:
	wsl -e bash -c "cd /mnt/e/Hacking/VIGILUM/circuits && ~/.nargo/bin/nargo compile"

test-circuits:
	wsl -e bash -c "cd /mnt/e/Hacking/VIGILUM/circuits && ~/.nargo/bin/nargo test"

prove-circuits:
	wsl -e bash -c "cd /mnt/e/Hacking/VIGILUM/circuits && ~/.nargo/bin/nargo prove"

# ═══════════════════════════════════════════════════════════════════════════
# SDK (TYPESCRIPT)
# ═══════════════════════════════════════════════════════════════════════════

build-sdk:
	cd sdk/ts-sdk && npm install && npm run build

test-sdk:
	cd sdk/ts-sdk && npm test

lint-sdk:
	cd sdk/ts-sdk && npm run lint

# ═══════════════════════════════════════════════════════════════════════════
# ML PIPELINE (PYTHON)
# ═══════════════════════════════════════════════════════════════════════════

setup-ml:
	cd ml && python -m venv .venv
	cd ml && .venv/Scripts/pip install -e ".[dev]"

test-ml:
	cd ml && .venv/Scripts/pytest tests/ -v

lint-ml:
	cd ml && .venv/Scripts/ruff check src/

train-model:
	cd ml && .venv/Scripts/python -m vigilum_ml.training.pipeline

export-onnx:
	cd ml && .venv/Scripts/python -m vigilum_ml.export.onnx

# ═══════════════════════════════════════════════════════════════════════════
# DOCKER
# ═══════════════════════════════════════════════════════════════════════════

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	docker compose build

# ═══════════════════════════════════════════════════════════════════════════
# DEVELOPMENT
# ═══════════════════════════════════════════════════════════════════════════

dev: docker-up
	@echo "Starting development environment..."
	@echo "Postgres: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "Qdrant: localhost:6333"
	@echo "ClickHouse: localhost:9000"
	@echo "Temporal UI: http://localhost:8080"
	@echo "Grafana: http://localhost:3000"
	@echo "Jaeger: http://localhost:16686"

anvil:
	C:/Users/over9/.foundry/bin/anvil.exe

# ═══════════════════════════════════════════════════════════════════════════
# DATABASE
# ═══════════════════════════════════════════════════════════════════════════

migrate-up:
	cd backend && go run ./cmd/migrate up

migrate-down:
	cd backend && go run ./cmd/migrate down

migrate-create:
	@read -p "Migration name: " name; \
	cd backend && go run ./cmd/migrate create $$name
