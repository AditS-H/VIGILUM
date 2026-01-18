-- VIGILUM Database Schema
-- Migration: 001_init.sql
-- Description: Initial database schema for VIGILUM backend
-- ==========================================================================

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==========================================================================
-- ENUMS
-- ==========================================================================

CREATE TYPE threat_level AS ENUM (
    'none',
    'info',
    'low',
    'medium',
    'high',
    'critical'
);

CREATE TYPE scan_type AS ENUM (
    'static',
    'dynamic',
    'ml_inference',
    'symbolic',
    'fuzz',
    'full'
);

CREATE TYPE scan_status AS ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE TYPE vulnerability_type AS ENUM (
    'reentrancy',
    'integer_overflow',
    'integer_underflow',
    'access_control',
    'unchecked_external_call',
    'tx_origin',
    'timestamp_dependency',
    'frontrunning',
    'flash_loan_attack',
    'oracle_manipulation',
    'rug_pull_pattern',
    'honeypot',
    'phishing_signature'
);

CREATE TYPE alert_type AS ENUM (
    'scan_result',
    'realtime_detection',
    'mempool_threat',
    'anomaly_detection',
    'reputation_change'
);

CREATE TYPE proof_status AS ENUM (
    'pending',
    'verified',
    'rejected',
    'expired'
);

-- ==========================================================================
-- USERS & AUTHENTICATION
-- ==========================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_address VARCHAR(42) NOT NULL UNIQUE,
    nonce VARCHAR(64) NOT NULL DEFAULT encode(gen_random_bytes(32), 'hex'),
    is_oracle BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    reputation_score INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_wallet ON users(wallet_address);

-- ==========================================================================
-- HUMAN PROOFS (Identity Firewall)
-- ==========================================================================

CREATE TABLE human_proofs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    proof_hash VARCHAR(66) NOT NULL UNIQUE,
    proof_data BYTEA,
    public_inputs JSONB,
    status proof_status DEFAULT 'pending',
    verified_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    chain_id INTEGER NOT NULL,
    tx_hash VARCHAR(66),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_human_proofs_user ON human_proofs(user_id);
CREATE INDEX idx_human_proofs_status ON human_proofs(status);
CREATE INDEX idx_human_proofs_hash ON human_proofs(proof_hash);

-- ==========================================================================
-- CONTRACTS
-- ==========================================================================

CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id INTEGER NOT NULL,
    address VARCHAR(42) NOT NULL,
    name VARCHAR(255),
    bytecode_hash VARCHAR(66) NOT NULL,
    bytecode BYTEA,
    source_code TEXT,
    abi JSONB,
    compiler_version VARCHAR(50),
    is_verified BOOLEAN DEFAULT FALSE,
    is_blacklisted BOOLEAN DEFAULT FALSE,
    is_proxy BOOLEAN DEFAULT FALSE,
    implementation_address VARCHAR(42),
    deployer_address VARCHAR(42),
    deploy_tx_hash VARCHAR(66),
    deploy_block_number BIGINT,
    deployed_at TIMESTAMP WITH TIME ZONE,
    risk_score DECIMAL(5,2) DEFAULT 0 CHECK (risk_score >= 0 AND risk_score <= 100),
    threat_level threat_level DEFAULT 'none',
    labels TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(chain_id, address)
);

CREATE INDEX idx_contracts_chain_address ON contracts(chain_id, address);
CREATE INDEX idx_contracts_bytecode_hash ON contracts(bytecode_hash);
CREATE INDEX idx_contracts_risk_score ON contracts(risk_score DESC);
CREATE INDEX idx_contracts_threat_level ON contracts(threat_level);
CREATE INDEX idx_contracts_deployer ON contracts(deployer_address);
CREATE INDEX idx_contracts_blacklisted ON contracts(is_blacklisted) WHERE is_blacklisted = TRUE;

-- ==========================================================================
-- VULNERABILITIES
-- ==========================================================================

CREATE TABLE vulnerabilities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID REFERENCES contracts(id) ON DELETE CASCADE,
    type vulnerability_type NOT NULL,
    severity threat_level NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    location_file VARCHAR(255),
    location_start_line INTEGER,
    location_end_line INTEGER,
    location_snippet TEXT,
    remediation TEXT,
    cwe VARCHAR(20),
    confidence DECIMAL(3,2) CHECK (confidence >= 0 AND confidence <= 1),
    detected_by VARCHAR(100) NOT NULL,
    is_confirmed BOOLEAN DEFAULT FALSE,
    is_false_positive BOOLEAN DEFAULT FALSE,
    confirmed_by UUID REFERENCES users(id),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_vulnerabilities_contract ON vulnerabilities(contract_id);
CREATE INDEX idx_vulnerabilities_type ON vulnerabilities(type);
CREATE INDEX idx_vulnerabilities_severity ON vulnerabilities(severity);
CREATE INDEX idx_vulnerabilities_confirmed ON vulnerabilities(is_confirmed);

-- ==========================================================================
-- SCAN REPORTS
-- ==========================================================================

CREATE TABLE scan_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID REFERENCES contracts(id) ON DELETE CASCADE,
    scan_type scan_type NOT NULL,
    status scan_status DEFAULT 'pending',
    risk_score DECIMAL(5,2) CHECK (risk_score >= 0 AND risk_score <= 100),
    threat_level threat_level,
    total_issues INTEGER DEFAULT 0,
    critical_count INTEGER DEFAULT 0,
    high_count INTEGER DEFAULT 0,
    medium_count INTEGER DEFAULT 0,
    low_count INTEGER DEFAULT 0,
    info_count INTEGER DEFAULT 0,
    code_coverage DECIMAL(5,2),
    paths_explored INTEGER,
    instructions_executed BIGINT,
    raw_output BYTEA,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_scan_reports_contract ON scan_reports(contract_id);
CREATE INDEX idx_scan_reports_status ON scan_reports(status);
CREATE INDEX idx_scan_reports_type ON scan_reports(scan_type);
CREATE INDEX idx_scan_reports_started ON scan_reports(started_at DESC);

-- ==========================================================================
-- THREAT SIGNALS (Oracle Layer)
-- ==========================================================================

CREATE TABLE threat_signals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id INTEGER NOT NULL,
    entity_address VARCHAR(42) NOT NULL,
    signal_type VARCHAR(100) NOT NULL,
    risk_score DECIMAL(5,2) CHECK (risk_score >= 0 AND risk_score <= 100),
    threat_level threat_level NOT NULL,
    source VARCHAR(100) NOT NULL,
    confidence DECIMAL(3,2) CHECK (confidence >= 0 AND confidence <= 1),
    evidence JSONB,
    is_published BOOLEAN DEFAULT FALSE,
    published_tx_hash VARCHAR(66),
    published_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_threat_signals_entity ON threat_signals(chain_id, entity_address);
CREATE INDEX idx_threat_signals_type ON threat_signals(signal_type);
CREATE INDEX idx_threat_signals_level ON threat_signals(threat_level);
CREATE INDEX idx_threat_signals_published ON threat_signals(is_published);

-- ==========================================================================
-- MALWARE GENOMES
-- ==========================================================================

CREATE TABLE malware_genomes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    genome_hash VARCHAR(66) NOT NULL UNIQUE,
    ipfs_hash VARCHAR(64),
    arweave_tx VARCHAR(64),
    label VARCHAR(100) NOT NULL,
    family VARCHAR(100),
    opcode_histogram JSONB,
    call_graph JSONB,
    gas_profile JSONB,
    similarity_vector REAL[],
    first_seen_contract_id UUID REFERENCES contracts(id),
    first_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    occurrences INTEGER DEFAULT 1,
    is_malicious BOOLEAN DEFAULT FALSE,
    severity threat_level,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_malware_genomes_hash ON malware_genomes(genome_hash);
CREATE INDEX idx_malware_genomes_label ON malware_genomes(label);
CREATE INDEX idx_malware_genomes_family ON malware_genomes(family);
CREATE INDEX idx_malware_genomes_malicious ON malware_genomes(is_malicious) WHERE is_malicious = TRUE;

-- ==========================================================================
-- ALERTS
-- ==========================================================================

CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type alert_type NOT NULL,
    severity threat_level NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    chain_id INTEGER NOT NULL,
    contract_id UUID REFERENCES contracts(id) ON DELETE SET NULL,
    address VARCHAR(42),
    tx_hash VARCHAR(66),
    metadata JSONB DEFAULT '{}',
    is_acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_alerts_type ON alerts(type);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_chain ON alerts(chain_id);
CREATE INDEX idx_alerts_contract ON alerts(contract_id);
CREATE INDEX idx_alerts_created ON alerts(created_at DESC);
CREATE INDEX idx_alerts_unacked ON alerts(is_acknowledged) WHERE is_acknowledged = FALSE;

-- ==========================================================================
-- EXPLOIT SUBMISSIONS (Red-Team DAO)
-- ==========================================================================

CREATE TABLE exploit_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    submitter_id UUID REFERENCES users(id) ON DELETE SET NULL,
    target_contract_id UUID REFERENCES contracts(id),
    target_chain_id INTEGER NOT NULL,
    target_address VARCHAR(42) NOT NULL,
    proof_hash VARCHAR(66) NOT NULL UNIQUE,
    proof_data BYTEA,
    genome_hash VARCHAR(66) REFERENCES malware_genomes(genome_hash),
    description TEXT,
    impact_level threat_level NOT NULL,
    vulnerability_types vulnerability_type[] DEFAULT '{}',
    is_verified BOOLEAN DEFAULT FALSE,
    is_novel BOOLEAN,
    bounty_amount DECIMAL(20,8),
    bounty_token VARCHAR(42),
    bounty_paid BOOLEAN DEFAULT FALSE,
    bounty_tx_hash VARCHAR(66),
    verified_by UUID REFERENCES users(id),
    verified_at TIMESTAMP WITH TIME ZONE,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_exploit_submissions_submitter ON exploit_submissions(submitter_id);
CREATE INDEX idx_exploit_submissions_target ON exploit_submissions(target_chain_id, target_address);
CREATE INDEX idx_exploit_submissions_verified ON exploit_submissions(is_verified);
CREATE INDEX idx_exploit_submissions_proof ON exploit_submissions(proof_hash);

-- ==========================================================================
-- RESEARCHER REPUTATION
-- ==========================================================================

CREATE TABLE researcher_reputation (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    tier INTEGER DEFAULT 1 CHECK (tier >= 1 AND tier <= 4),
    total_audits INTEGER DEFAULT 0,
    confirmed_vulns INTEGER DEFAULT 0,
    false_positives INTEGER DEFAULT 0,
    total_bounties DECIMAL(20,8) DEFAULT 0,
    reputation_score INTEGER DEFAULT 0,
    stake_amount DECIMAL(20,8) DEFAULT 0,
    slashed_amount DECIMAL(20,8) DEFAULT 0,
    commitment_hash VARCHAR(66),
    last_activity_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_researcher_reputation_user ON researcher_reputation(user_id);
CREATE INDEX idx_researcher_reputation_tier ON researcher_reputation(tier);
CREATE INDEX idx_researcher_reputation_score ON researcher_reputation(reputation_score DESC);

-- ==========================================================================
-- WEBHOOK SUBSCRIPTIONS
-- ==========================================================================

CREATE TABLE webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    webhook_url TEXT NOT NULL,
    secret_hash VARCHAR(66) NOT NULL,
    chain_ids INTEGER[] DEFAULT '{}',
    addresses VARCHAR(42)[] DEFAULT '{}',
    alert_types alert_type[] DEFAULT '{}',
    min_severity threat_level DEFAULT 'low',
    is_active BOOLEAN DEFAULT TRUE,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    failure_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_webhook_subscriptions_user ON webhook_subscriptions(user_id);
CREATE INDEX idx_webhook_subscriptions_active ON webhook_subscriptions(is_active) WHERE is_active = TRUE;

-- ==========================================================================
-- API KEYS
-- ==========================================================================

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(66) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    permissions TEXT[] DEFAULT '{"read"}',
    rate_limit INTEGER DEFAULT 1000,
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_active ON api_keys(is_active) WHERE is_active = TRUE;

-- ==========================================================================
-- AUDIT LOG
-- ==========================================================================

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_log_user ON audit_log(user_id);
CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);

-- ==========================================================================
-- FUNCTIONS
-- ==========================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to tables with updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_contracts_updated_at BEFORE UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_threat_signals_updated_at BEFORE UPDATE ON threat_signals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_researcher_reputation_updated_at BEFORE UPDATE ON researcher_reputation
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhook_subscriptions_updated_at BEFORE UPDATE ON webhook_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==========================================================================
-- VIEWS
-- ==========================================================================

-- Contracts with latest scan summary
CREATE VIEW v_contracts_with_scans AS
SELECT 
    c.*,
    ls.last_scan_at,
    ls.last_scan_status,
    ls.total_scans
FROM contracts c
LEFT JOIN LATERAL (
    SELECT 
        MAX(started_at) as last_scan_at,
        (SELECT status FROM scan_reports WHERE contract_id = c.id ORDER BY started_at DESC LIMIT 1) as last_scan_status,
        COUNT(*) as total_scans
    FROM scan_reports
    WHERE contract_id = c.id
) ls ON TRUE;

-- Active high-risk contracts
CREATE VIEW v_high_risk_contracts AS
SELECT *
FROM contracts
WHERE risk_score >= 70
  AND is_blacklisted = FALSE
ORDER BY risk_score DESC;

-- Recent alerts summary
CREATE VIEW v_recent_alerts AS
SELECT 
    DATE_TRUNC('hour', created_at) as hour,
    type,
    severity,
    COUNT(*) as count
FROM alerts
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY DATE_TRUNC('hour', created_at), type, severity
ORDER BY hour DESC;

-- ==========================================================================
-- COMMENTS
-- ==========================================================================

COMMENT ON TABLE users IS 'Registered users identified by wallet address';
COMMENT ON TABLE human_proofs IS 'ZK proofs of human-like behavior for Identity Firewall';
COMMENT ON TABLE contracts IS 'Smart contracts tracked by VIGILUM';
COMMENT ON TABLE vulnerabilities IS 'Detected vulnerabilities in smart contracts';
COMMENT ON TABLE scan_reports IS 'Security scan results and metrics';
COMMENT ON TABLE threat_signals IS 'Aggregated threat intelligence signals';
COMMENT ON TABLE malware_genomes IS 'Fingerprints of malicious contract patterns';
COMMENT ON TABLE alerts IS 'Security alerts triggered by the system';
COMMENT ON TABLE exploit_submissions IS 'ZK exploit proofs submitted by researchers';
COMMENT ON TABLE researcher_reputation IS 'Reputation scores for security researchers';
