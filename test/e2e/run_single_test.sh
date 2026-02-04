#!/usr/bin/env bash
# run_single_test.sh - Run a single test from the E2E suite
#
# Usage: ./run_single_test.sh <test_name>
# Example: ./run_single_test.sh test_device_discovery

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_SCRIPT="$SCRIPT_DIR/forge_terminal_test.sh"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*" >&2
}

error() {
    echo -e "${RED}✗${NC} $*" >&2
}

# Check arguments
if [[ $# -eq 0 ]]; then
    error "Usage: $0 <test_name>"
    echo ""
    echo "Available tests:"
    echo "  test_device_discovery"
    echo "  test_simulator_boot"
    echo "  test_app_install"
    echo "  test_app_launch"
    echo "  test_screenshot"
    echo "  test_app_terminate"
    exit 1
fi

TEST_NAME="$1"

# Source the test script to get functions
if [[ ! -f "$SOURCE_SCRIPT" ]]; then
    error "Test script not found: $SOURCE_SCRIPT"
    exit 1
fi

log "Loading test functions from $SOURCE_SCRIPT"

# Source the script (this loads all functions)
# We need to disable errexit temporarily
set +e
source "$SOURCE_SCRIPT"
set -e

# Check if test function exists
if ! declare -f "$TEST_NAME" >/dev/null; then
    error "Test function '$TEST_NAME' not found"
    echo ""
    echo "Available tests:"
    declare -F | grep "^declare -f test_" | sed 's/declare -f /  /'
    exit 1
fi

# Setup environment
log "Setting up test environment..."
check_prerequisites || exit 1
setup

# Run the specific test
log "Running test: $TEST_NAME"
echo ""

if $TEST_NAME; then
    echo ""
    log "Test passed: $TEST_NAME"
    cleanup
    exit 0
else
    echo ""
    error "Test failed: $TEST_NAME"
    TEST_FAILED=true
    cleanup
    exit 1
fi
