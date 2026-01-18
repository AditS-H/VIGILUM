# ==========================================================================
# VIGILUM Database Setup Script (PowerShell)
# ==========================================================================
# Usage: .\setup-db.ps1 [command]
# Commands:
#   init      - Initialize database from scratch
#   migrate   - Run pending migrations
#   seed      - Seed test data
#   reset     - Drop and recreate database
#   status    - Show migration status
# ==========================================================================

param(
    [Parameter(Position=0)]
    [ValidateSet("init", "migrate", "seed", "reset", "status")]
    [string]$Command = "init"
)

$ErrorActionPreference = "Stop"

# Load environment variables from .env
$envFile = Join-Path $PSScriptRoot "..\..\..\.env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match "^([^#][^=]+)=(.*)$") {
            [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
        }
    }
}

# Default values
$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "vigilum" }
$DB_PASSWORD = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "vigilum" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "vigilum" }

$env:PGPASSWORD = $DB_PASSWORD

function Write-Info($message) {
    Write-Host "[INFO] $message" -ForegroundColor Green
}

function Write-Warn($message) {
    Write-Host "[WARN] $message" -ForegroundColor Yellow
}

function Write-Err($message) {
    Write-Host "[ERROR] $message" -ForegroundColor Red
}

function Invoke-Psql {
    param([string]$Database, [string]$Query)
    $result = & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $Database -t -c $Query 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "PostgreSQL command failed: $result"
    }
    return $result
}

function Invoke-PsqlFile {
    param([string]$Database, [string]$FilePath)
    & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $Database -f $FilePath
    if ($LASTEXITCODE -ne 0) {
        throw "PostgreSQL file execution failed"
    }
}

function Test-Connection {
    Write-Info "Checking PostgreSQL connection..."
    try {
        Invoke-Psql -Database "postgres" -Query "SELECT 1"
        Write-Info "PostgreSQL connection successful"
        return $true
    } catch {
        Write-Err "Cannot connect to PostgreSQL at ${DB_HOST}:${DB_PORT}"
        return $false
    }
}

function New-Database {
    Write-Info "Creating database '$DB_NAME' if not exists..."
    $exists = Invoke-Psql -Database "postgres" -Query "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'"
    if (-not $exists.Trim()) {
        Invoke-Psql -Database "postgres" -Query "CREATE DATABASE $DB_NAME"
    }
    Write-Info "Database '$DB_NAME' ready"
}

function Invoke-Migrations {
    Write-Info "Running migrations..."
    
    $migrationsDir = Join-Path $PSScriptRoot "..\internal\db\migrations"
    
    # Create migrations tracking table
    Invoke-Psql -Database $DB_NAME -Query @"
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
"@

    # Get migration files
    $migrations = Get-ChildItem -Path $migrationsDir -Filter "*.sql" | Sort-Object Name
    
    foreach ($migration in $migrations) {
        $filename = $migration.Name
        
        # Check if already applied
        $applied = Invoke-Psql -Database $DB_NAME -Query "SELECT COUNT(*) FROM schema_migrations WHERE version = '$filename'"
        
        if ($applied.Trim() -eq "0") {
            Write-Info "Applying migration: $filename"
            Invoke-PsqlFile -Database $DB_NAME -FilePath $migration.FullName
            Invoke-Psql -Database $DB_NAME -Query "INSERT INTO schema_migrations (version) VALUES ('$filename')"
            Write-Info "Migration $filename applied successfully"
        } else {
            Write-Info "Skipping $filename (already applied)"
        }
    }
    
    Write-Info "All migrations completed"
}

function Invoke-Seed {
    Write-Info "Seeding test data..."
    
    $seedSql = @"
-- Insert test users
INSERT INTO users (wallet_address, is_oracle, is_admin) VALUES
    ('0x1234567890123456789012345678901234567890', true, true),
    ('0xabcdefabcdefabcdefabcdefabcdefabcdefabcd', true, false),
    ('0x0000000000000000000000000000000000000001', false, false)
ON CONFLICT (wallet_address) DO NOTHING;

-- Insert test contracts
INSERT INTO contracts (chain_id, address, name, bytecode_hash, risk_score, threat_level) VALUES
    (1, '0xdead000000000000000000000000000000000001', 'Test Token', '0x' || encode(sha256('bytecode1'::bytea), 'hex'), 25.5, 'low'),
    (1, '0xdead000000000000000000000000000000000002', 'Risky DEX', '0x' || encode(sha256('bytecode2'::bytea), 'hex'), 75.0, 'high'),
    (137, '0xdead000000000000000000000000000000000003', 'Safe Vault', '0x' || encode(sha256('bytecode3'::bytea), 'hex'), 10.0, 'info'),
    (1, '0xdead000000000000000000000000000000000004', 'Known Scam', '0x' || encode(sha256('bytecode4'::bytea), 'hex'), 95.0, 'critical')
ON CONFLICT (chain_id, address) DO NOTHING;

-- Mark one as blacklisted
UPDATE contracts SET is_blacklisted = true 
WHERE address = '0xdead000000000000000000000000000000000004';

-- Insert sample threat signals
INSERT INTO threat_signals (chain_id, entity_address, signal_type, risk_score, threat_level, source, confidence)
VALUES
    (1, '0xdead000000000000000000000000000000000004', 'known_scammer', 95.0, 'critical', 'community_report', 0.95),
    (1, '0xdead000000000000000000000000000000000002', 'high_risk_pattern', 70.0, 'high', 'ml_detector', 0.80)
ON CONFLICT DO NOTHING;

-- Insert sample malware genome
INSERT INTO malware_genomes (genome_hash, label, family, is_malicious, severity)
VALUES
    ('0x' || encode(sha256('genome1'::bytea), 'hex'), 'rug_pull_v1', 'rug_pull', true, 'critical'),
    ('0x' || encode(sha256('genome2'::bytea), 'hex'), 'honeypot_standard', 'honeypot', true, 'high')
ON CONFLICT (genome_hash) DO NOTHING;

-- Insert sample alerts
INSERT INTO alerts (type, severity, title, description, chain_id, address)
VALUES
    ('scan_result', 'high', 'High-Risk Contract Detected', 'Reentrancy vulnerability found', 1, '0xdead000000000000000000000000000000000002'),
    ('mempool_threat', 'critical', 'Potential Exploit Attempt', 'Suspicious transaction pattern', 1, '0xdead000000000000000000000000000000000004')
ON CONFLICT DO NOTHING;
"@

    & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c $seedSql
    Write-Info "Test data seeded successfully"
}

function Reset-Database {
    Write-Warn "This will DROP the '$DB_NAME' database. Are you sure? (y/N)"
    $response = Read-Host
    if ($response -eq "y" -or $response -eq "Y") {
        Write-Info "Dropping database '$DB_NAME'..."
        Invoke-Psql -Database "postgres" -Query "DROP DATABASE IF EXISTS $DB_NAME"
        New-Database
        Invoke-Migrations
        Write-Info "Database reset complete"
    } else {
        Write-Info "Reset cancelled"
    }
}

function Get-Status {
    Write-Info "Migration status:"
    & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version, applied_at FROM schema_migrations ORDER BY version"
    
    Write-Info "`nTable counts:"
    & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c @"
        SELECT 
            schemaname,
            relname as table,
            n_live_tup as row_count
        FROM pg_stat_user_tables
        ORDER BY n_live_tup DESC;
"@
}

# Main
if (-not (Test-Connection)) {
    exit 1
}

switch ($Command) {
    "init" {
        New-Database
        Invoke-Migrations
    }
    "migrate" {
        Invoke-Migrations
    }
    "seed" {
        Invoke-Seed
    }
    "reset" {
        Reset-Database
    }
    "status" {
        Get-Status
    }
}
