-- Migration: 001_initial_schema.down.sql
-- Rollback script for initial schema

DROP INDEX IF EXISTS idx_api_keys_revoked;
DROP INDEX IF EXISTS idx_api_keys_tier;
DROP INDEX IF EXISTS idx_api_keys_user;
DROP INDEX IF EXISTS idx_api_keys_hash;
DROP TABLE IF EXISTS api_keys;

DROP INDEX IF EXISTS idx_exploit_target;
DROP INDEX IF EXISTS idx_exploit_bounty_status;
DROP INDEX IF EXISTS idx_exploit_status;
DROP INDEX IF EXISTS idx_exploit_researcher;
DROP TABLE IF EXISTS exploit_submissions;

DROP INDEX IF EXISTS idx_genomes_label;
DROP INDEX IF EXISTS idx_genomes_contract;
DROP INDEX IF EXISTS idx_genomes_hash;
DROP TABLE IF EXISTS genomes;

DROP INDEX IF EXISTS idx_threat_signals_risk;
DROP INDEX IF EXISTS idx_threat_signals_unpublished;
DROP INDEX IF EXISTS idx_threat_signals_entity;
DROP TABLE IF EXISTS threat_signals;

DROP INDEX IF EXISTS idx_human_proofs_verified;
DROP INDEX IF EXISTS idx_human_proofs_hash;
DROP INDEX IF EXISTS idx_human_proofs_user;
DROP TABLE IF EXISTS human_proofs;

DROP INDEX IF EXISTS idx_users_risk_score;
DROP INDEX IF EXISTS idx_users_wallet;
DROP TABLE IF EXISTS users;
