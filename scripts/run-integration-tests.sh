#!/bin/bash
# Integration test runner for VIGILUM backend repositories
# Starts Docker containers, runs tests, and generates reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
TEST_DB_HOST="${TEST_DB_HOST:-localhost}"
TEST_DB_PORT="${TEST_DB_PORT:-5433}"
TEST_DB_USER="${TEST_DB_USER:-postgres}"
TEST_DB_PASSWORD="${TEST_DB_PASSWORD:-postgres}"
TEST_DB_NAME="${TEST_DB_NAME:-vigilum_test}"
TIMEOUT="${TEST_TIMEOUT:-60}"

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Please install Docker."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose not found. Please install Docker Compose."
        exit 1
    fi
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go not found. Please install Go 1.24+"
        exit 1
    fi
    
    log_info "All prerequisites found ✓"
}

# Start Docker containers
start_containers() {
    log_info "Starting Docker containers..."
    
    export TEST_DB_USER
    export TEST_DB_PASSWORD
    export TEST_DB_NAME
    export TEST_DB_PORT
    
    docker-compose -f docker-compose.test.yml up -d
    
    # Wait for PostgreSQL to be ready
    log_info "Waiting for PostgreSQL to be healthy (max ${TIMEOUT}s)..."
    start_time=$(date +%s)
    
    while true; do
        if docker exec vigilum-postgres-test pg_isready -U "$TEST_DB_USER" -d "$TEST_DB_NAME" &> /dev/null; then
            log_info "PostgreSQL is ready ✓"
            break
        fi
        
        current_time=$(date +%s)
        elapsed=$((current_time - start_time))
        
        if [ $elapsed -gt $TIMEOUT ]; then
            log_error "PostgreSQL failed to start within ${TIMEOUT}s"
            docker-compose -f docker-compose.test.yml logs postgres-test
            exit 1
        fi
        
        echo -n "."
        sleep 2
    done
}

# Stop Docker containers
stop_containers() {
    log_info "Stopping Docker containers..."
    docker-compose -f docker-compose.test.yml down
    log_info "Containers stopped ✓"
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."
    
    cd backend
    
    # Set environment variables for tests
    export TEST_DB_HOST
    export TEST_DB_PORT
    export TEST_DB_USER
    export TEST_DB_PASSWORD
    export TEST_DB_NAME
    
    # Run tests with verbose output and race detection
    if go test -v -race -timeout 300s ./internal/db/repositories; then
        log_info "Integration tests passed ✓"
        return 0
    else
        log_error "Integration tests failed ✗"
        return 1
    fi
}

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."
    
    cd backend
    
    if go test -v ./internal/firewall ./internal/oracle ./internal/genome; then
        log_info "Unit tests passed ✓"
        return 0
    else
        log_error "Unit tests failed ✗"
        return 1
    fi
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    cd backend
    
    # Run tests with coverage
    go test -v -coverprofile=coverage.out ./internal/db/repositories
    go tool cover -html=coverage.out -o coverage.html
    
    log_info "Coverage report generated: coverage.html"
}

# Clean up
cleanup() {
    log_info "Cleaning up..."
    
    # Stop containers
    if [ "$KEEP_CONTAINERS" != "true" ]; then
        stop_containers
    else
        log_warn "Keeping containers running (set KEEP_CONTAINERS=false to clean up)"
    fi
}

# Trap errors and cleanup
trap cleanup EXIT

# Main execution
main() {
    log_info "=== VIGILUM Integration Test Suite ==="
    log_info "Started at $(date)"
    
    check_prerequisites
    start_containers
    
    # Run tests
    if run_unit_tests && run_integration_tests; then
        log_info "All tests passed! ✓"
        
        # Generate coverage report if requested
        if [ "$GENERATE_COVERAGE" = "true" ]; then
            generate_report
        fi
        
        exit 0
    else
        log_error "Tests failed! ✗"
        exit 1
    fi
}

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Integration test runner for VIGILUM backend

Options:
    -h, --help              Show this help message
    -k, --keep-containers   Keep Docker containers running after tests
    -c, --coverage          Generate coverage report
    --db-host HOST          Database host (default: localhost)
    --db-port PORT          Database port (default: 5433)
    --db-user USER          Database user (default: postgres)
    --db-password PASS      Database password (default: postgres)
    --db-name NAME          Database name (default: vigilum_test)
    --timeout SECONDS       Timeout for container startup (default: 60)

Examples:
    # Run all tests
    $0
    
    # Keep containers and generate coverage
    $0 --keep-containers --coverage
    
    # Custom database configuration
    $0 --db-host db.example.com --db-port 5432

EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -k|--keep-containers)
            KEEP_CONTAINERS=true
            shift
            ;;
        -c|--coverage)
            GENERATE_COVERAGE=true
            shift
            ;;
        --db-host)
            TEST_DB_HOST="$2"
            shift 2
            ;;
        --db-port)
            TEST_DB_PORT="$2"
            shift 2
            ;;
        --db-user)
            TEST_DB_USER="$2"
            shift 2
            ;;
        --db-password)
            TEST_DB_PASSWORD="$2"
            shift 2
            ;;
        --db-name)
            TEST_DB_NAME="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

main
