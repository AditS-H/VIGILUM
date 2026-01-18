#!/bin/bash
# ==========================================================================
# VIGILUM Database Setup Script
# ==========================================================================
# Usage: ./setup-db.sh [command]
# Commands:
#   init      - Initialize database from scratch
#   migrate   - Run pending migrations
#   seed      - Seed test data
#   reset     - Drop and recreate database
#   status    - Show migration status
# ==========================================================================

set -e

# Load environment variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-vigilum}
DB_PASSWORD=${DB_PASSWORD:-vigilum}
DB_NAME=${DB_NAME:-vigilum}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Connection string
PGPASSWORD=$DB_PASSWORD
export PGPASSWORD

PSQL_CMD="psql -h $DB_HOST -p $DB_PORT -U $DB_USER"

# Check PostgreSQL connection
check_connection() {
    log_info "Checking PostgreSQL connection..."
    if $PSQL_CMD -d postgres -c "SELECT 1" > /dev/null 2>&1; then
        log_info "PostgreSQL connection successful"
        return 0
    else
        log_error "Cannot connect to PostgreSQL at $DB_HOST:$DB_PORT"
        return 1
    fi
}

# Create database if not exists
create_database() {
    log_info "Creating database '$DB_NAME' if not exists..."
    $PSQL_CMD -d postgres -c "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
        $PSQL_CMD -d postgres -c "CREATE DATABASE $DB_NAME"
    log_info "Database '$DB_NAME' ready"
}

# Run migrations
run_migrations() {
    log_info "Running migrations..."
    
    MIGRATIONS_DIR="internal/db/migrations"
    
    # Create migrations tracking table if not exists
    $PSQL_CMD -d $DB_NAME -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
    "
    
    # Run each migration file
    for migration in $(ls -1 $MIGRATIONS_DIR/*.sql 2>/dev/null | sort); do
        filename=$(basename $migration)
        
        # Check if already applied
        applied=$($PSQL_CMD -d $DB_NAME -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version = '$filename'")
        
        if [ "$applied" -eq "0" ]; then
            log_info "Applying migration: $filename"
            $PSQL_CMD -d $DB_NAME -f $migration
            $PSQL_CMD -d $DB_NAME -c "INSERT INTO schema_migrations (version) VALUES ('$filename')"
            log_info "Migration $filename applied successfully"
        else
            log_info "Skipping $filename (already applied)"
        fi
    done
    
    log_info "All migrations completed"
}

# Seed test data
seed_data() {
    log_info "Seeding test data..."
    
    $PSQL_CMD -d $DB_NAME << 'EOF'
-- Insert test users
INSERT INTO users (wallet_address, is_oracle, is_admin) VALUES
    ('0x1234567890123456789012345678901234567890', true, true),
    ('0xabcdefabcdefabcdefabcdefabcdefabcdefabcd', true, false),
    ('0x0000000000000000000000000000000000000001', false, false)
ON CONFLICT (wallet_address) DO NOTHING;

-- Insert test contracts
INSERT INTO contracts (chain_id, address, name, bytecode_hash, risk_score, threat_level) VALUES
    (1, '0xdead000000000000000000000000000000000001', 'Test Token', '0x' || encode(sha256('bytecode1'), 'hex'), 25.5, 'low'),
    (1, '0xdead000000000000000000000000000000000002', 'Risky DEX', '0x' || encode(sha256('bytecode2'), 'hex'), 75.0, 'high'),
    (137, '0xdead000000000000000000000000000000000003', 'Safe Vault', '0x' || encode(sha256('bytecode3'), 'hex'), 10.0, 'info'),
    (1, '0xdead000000000000000000000000000000000004', 'Known Scam', '0x' || encode(sha256('bytecode4'), 'hex'), 95.0, 'critical')
ON CONFLICT (chain_id, address) DO NOTHING;

-- Mark one as blacklisted
UPDATE contracts SET is_blacklisted = true 
WHERE address = '0xdead000000000000000000000000000000000004';

-- Insert sample vulnerabilities
INSERT INTO vulnerabilities (contract_id, type, severity, title, description, confidence, detected_by)
SELECT 
    c.id,
    'reentrancy',
    'high',
    'Reentrancy Vulnerability',
    'External call before state update in withdraw function',
    0.85,
    'static-analyzer'
FROM contracts c
WHERE c.name = 'Risky DEX'
ON CONFLICT DO NOTHING;

-- Insert sample threat signals
INSERT INTO threat_signals (chain_id, entity_address, signal_type, risk_score, threat_level, source, confidence)
VALUES
    (1, '0xdead000000000000000000000000000000000004', 'known_scammer', 95.0, 'critical', 'community_report', 0.95),
    (1, '0xdead000000000000000000000000000000000002', 'high_risk_pattern', 70.0, 'high', 'ml_detector', 0.80)
ON CONFLICT DO NOTHING;

-- Insert sample malware genome
INSERT INTO malware_genomes (genome_hash, label, family, is_malicious, severity)
VALUES
    ('0x' || encode(sha256('genome1'), 'hex'), 'rug_pull_v1', 'rug_pull', true, 'critical'),
    ('0x' || encode(sha256('genome2'), 'hex'), 'honeypot_standard', 'honeypot', true, 'high')
ON CONFLICT (genome_hash) DO NOTHING;

-- Insert sample alerts
INSERT INTO alerts (type, severity, title, description, chain_id, address)
VALUES
    ('scan_result', 'high', 'High-Risk Contract Detected', 'Reentrancy vulnerability found in withdraw()', 1, '0xdead000000000000000000000000000000000002'),
    ('mempool_threat', 'critical', 'Potential Exploit Attempt', 'Suspicious transaction pattern detected', 1, '0xdead000000000000000000000000000000000004')
ON CONFLICT DO NOTHING;

EOF

    log_info "Test data seeded successfully"
}

# Reset database
reset_database() {
    log_warn "This will DROP the '$DB_NAME' database. Are you sure? (y/N)"
    read -r response
    if [ "$response" = "y" ] || [ "$response" = "Y" ]; then
        log_info "Dropping database '$DB_NAME'..."
        $PSQL_CMD -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME"
        create_database
        run_migrations
        log_info "Database reset complete"
    else
        log_info "Reset cancelled"
    fi
}

# Show migration status
show_status() {
    log_info "Migration status:"
    $PSQL_CMD -d $DB_NAME -c "SELECT version, applied_at FROM schema_migrations ORDER BY version"
    
    log_info "\nTable counts:"
    $PSQL_CMD -d $DB_NAME -c "
        SELECT 
            schemaname,
            relname as table,
            n_live_tup as row_count
        FROM pg_stat_user_tables
        ORDER BY n_live_tup DESC;
    "
}

# Main
case "${1:-init}" in
    init)
        check_connection
        create_database
        run_migrations
        ;;
    migrate)
        check_connection
        run_migrations
        ;;
    seed)
        check_connection
        seed_data
        ;;
    reset)
        check_connection
        reset_database
        ;;
    status)
        check_connection
        show_status
        ;;
    *)
        echo "Usage: $0 {init|migrate|seed|reset|status}"
        exit 1
        ;;
esac
