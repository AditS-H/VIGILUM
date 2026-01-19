// Package repositories provides test utilities for repository testing.
package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// TestDBConfig holds test database configuration.
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// GetTestDBConfig returns the test database configuration from environment.
func GetTestDBConfig() TestDBConfig {
	return TestDBConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5432"),
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		Database: getEnv("TEST_DB_NAME", "vigilum_test"),
		SSLMode:  "disable",
	}
}

// GetConnectionString returns the database connection string.
func (c TestDBConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// SetupTestDB creates and initializes a test database.
func SetupTestDB(t *testing.T) *sql.DB {
	config := GetTestDBConfig()
	connStr := config.GetConnectionString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Run migrations
	if err := runMigrations(t, db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// CleanupTestDB cleans up test database and closes connection.
func CleanupTestDB(t *testing.T, db *sql.DB) {
	if err := rollbackMigrations(t, db); err != nil {
		t.Logf("Warning: Failed to rollback migrations: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Logf("Warning: Failed to close database: %v", err)
	}
}

// runMigrations executes all database migrations.
func runMigrations(t *testing.T, db *sql.DB) error {
	ctx := context.Background()

	// Create users table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			wallet_address VARCHAR(255) NOT NULL UNIQUE,
			risk_score INTEGER DEFAULT 0,
			is_blacklisted BOOLEAN DEFAULT false,
			last_activity TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	// Create human_proofs table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS human_proofs (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL REFERENCES users(id),
			proof_hash VARCHAR(255) NOT NULL,
			proof_data JSONB,
			verifier_address VARCHAR(255),
			verified_at TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create human_proofs table: %w", err)
	}

	// Create threat_signals table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS threat_signals (
			id VARCHAR(255) PRIMARY KEY,
			chain_id BIGINT NOT NULL,
			address VARCHAR(255) NOT NULL,
			signal_type VARCHAR(100) NOT NULL,
			risk_score INTEGER NOT NULL,
			threat_level VARCHAR(50) NOT NULL,
			source_id VARCHAR(255),
			metadata JSONB,
			published_at TIMESTAMP,
			published_tx_hash VARCHAR(255),
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create threat_signals table: %w", err)
	}

	// Create genomes table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS genomes (
			id VARCHAR(255) PRIMARY KEY,
			genome_hash VARCHAR(255) NOT NULL UNIQUE,
			ipfs_hash VARCHAR(255) NOT NULL,
			contract_address VARCHAR(255),
			bytecode_size INTEGER,
			opcode_count INTEGER,
			function_count INTEGER,
			complexity_score DECIMAL(5,2),
			label VARCHAR(100),
			features JSONB,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create genomes table: %w", err)
	}

	// Create exploit_submissions table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS exploit_submissions (
			id VARCHAR(255) PRIMARY KEY,
			researcher_address VARCHAR(255) NOT NULL,
			target_contract VARCHAR(255) NOT NULL,
			chain_id BIGINT NOT NULL,
			proof_hash VARCHAR(255) NOT NULL,
			genome_id VARCHAR(255) REFERENCES genomes(id),
			description TEXT,
			severity VARCHAR(50),
			bounty_amount BIGINT,
			bounty_status VARCHAR(50),
			status VARCHAR(50),
			verified_at TIMESTAMP,
			paid_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create exploit_submissions table: %w", err)
	}

	// Create api_keys table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS api_keys (
			id VARCHAR(255) PRIMARY KEY,
			key_hash BYTEA NOT NULL UNIQUE,
			user_id VARCHAR(255) NOT NULL REFERENCES users(id),
			name VARCHAR(255),
			tier VARCHAR(50),
			rate_limit INTEGER,
			requests_today INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			last_used TIMESTAMP,
			expires_at TIMESTAMP,
			revoked BOOLEAN DEFAULT false
		)
	`)
	if err != nil {
		return fmt.Errorf("create api_keys table: %w", err)
	}

	return nil
}

// rollbackMigrations drops all test tables.
func rollbackMigrations(t *testing.T, db *sql.DB) error {
	ctx := context.Background()

	tables := []string{
		"api_keys",
		"exploit_submissions",
		"genomes",
		"threat_signals",
		"human_proofs",
		"users",
	}

	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("drop table %s: %w", table, err)
		}
	}

	return nil
}

// getEnv returns environment variable or default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TruncateTables clears all data from test tables.
func TruncateTables(t *testing.T, db *sql.DB) error {
	ctx := context.Background()

	tables := []string{
		"api_keys",
		"exploit_submissions",
		"genomes",
		"threat_signals",
		"human_proofs",
		"users",
	}

	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("truncate table %s: %w", table, err)
		}
	}

	return nil
}

// TestTransaction helper for testing transaction scenarios.
func TestTransaction(t *testing.T, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
