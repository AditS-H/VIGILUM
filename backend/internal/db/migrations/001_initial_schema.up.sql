-- Migration: 001_initial_schema.up.sql
-- Creates core tables for VIGILUM backend

-- ============================================================
-- USERS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP,
    risk_score FLOAT DEFAULT 0.0,
    is_blacklisted BOOLEAN DEFAULT FALSE,
    tags TEXT[]
);

CREATE INDEX IF NOT EXISTS idx_users_wallet ON users(wallet_address);
CREATE INDEX IF NOT EXISTS idx_users_risk_score ON users(risk_score DESC);

-- ============================================================
-- HUMAN PROOFS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS human_proofs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    proof_hash BYTEA NOT NULL,
    proof_data JSONB,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP,
    verifier_address VARCHAR(42),
    tx_hash VARCHAR(66),
    expires_at TIMESTAMP,
    UNIQUE(user_id, proof_hash)
);

CREATE INDEX IF NOT EXISTS idx_human_proofs_user ON human_proofs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_human_proofs_hash ON human_proofs(proof_hash);
CREATE INDEX IF NOT EXISTS idx_human_proofs_verified ON human_proofs(verified, created_at DESC);

-- ============================================================
-- THREAT SIGNALS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS threat_signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id BIGINT NOT NULL DEFAULT 1,
    entity_address VARCHAR(42) NOT NULL,
    signal_type VARCHAR(50) NOT NULL,
    risk_score INT NOT NULL CHECK (risk_score BETWEEN 0 AND 100),
    threat_level VARCHAR(20) NOT NULL,
    confidence FLOAT CHECK (confidence BETWEEN 0.0 AND 1.0),
    source VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP,
    UNIQUE(chain_id, entity_address, signal_type, created_at)
);

CREATE INDEX IF NOT EXISTS idx_threat_signals_entity ON threat_signals(chain_id, entity_address, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_threat_signals_unpublished ON threat_signals(published_at) WHERE published_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_threat_signals_risk ON threat_signals(risk_score DESC);

-- ============================================================
-- GENOMES TABLE (Malware fingerprints)
-- ============================================================
CREATE TABLE IF NOT EXISTS genomes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    genome_hash BYTEA UNIQUE NOT NULL,
    ipfs_hash VARCHAR(100) NOT NULL,
    contract_address VARCHAR(42),
    chain_id BIGINT DEFAULT 1,
    label VARCHAR(50),
    bytecode_size INT,
    opcode_count INT,
    function_count INT,
    complexity_score FLOAT,
    features JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(chain_id, contract_address)
);

CREATE INDEX IF NOT EXISTS idx_genomes_hash ON genomes(genome_hash);
CREATE INDEX IF NOT EXISTS idx_genomes_contract ON genomes(chain_id, contract_address);
CREATE INDEX IF NOT EXISTS idx_genomes_label ON genomes(label);

-- ============================================================
-- EXPLOIT SUBMISSIONS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS exploit_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    researcher_address VARCHAR(42) NOT NULL,
    target_contract VARCHAR(42) NOT NULL,
    chain_id BIGINT DEFAULT 1,
    proof_hash BYTEA NOT NULL,
    genome_id UUID REFERENCES genomes(id) ON DELETE SET NULL,
    description TEXT,
    severity VARCHAR(20),
    bounty_amount BIGINT,
    bounty_status VARCHAR(20) DEFAULT 'pending',
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP,
    paid_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_exploit_researcher ON exploit_submissions(researcher_address, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_exploit_status ON exploit_submissions(status);
CREATE INDEX IF NOT EXISTS idx_exploit_bounty_status ON exploit_submissions(bounty_status);
CREATE INDEX IF NOT EXISTS idx_exploit_target ON exploit_submissions(chain_id, target_contract);

-- ============================================================
-- API KEYS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash BYTEA UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255),
    tier VARCHAR(20) DEFAULT 'free',
    rate_limit INT DEFAULT 100,
    requests_today INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used TIMESTAMP,
    expires_at TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_tier ON api_keys(tier);
CREATE INDEX IF NOT EXISTS idx_api_keys_revoked ON api_keys(revoked);
