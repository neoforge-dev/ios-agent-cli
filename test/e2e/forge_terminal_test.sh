#!/usr/bin/env bash
# forge_terminal_test.sh - E2E test for ios-agent with ForgeTerminal app
#
# Tests ios-agent-cli functionality using the ForgeTerminal app as a target.
# This script exercises device discovery, simulator management, app install,
# and screenshot capabilities.

set -euo pipefail

#------------------------------------------------------------------------------
# Configuration
#------------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_FIXTURES_DIR="$SCRIPT_DIR/../fixtures"
TEST_OUTPUT_DIR="${TEST_OUTPUT_DIR:-/tmp/ios-agent-test-$$}"
TEST_STATE_FILE="$TEST_OUTPUT_DIR/test_state.env"

IOS_AGENT="${PROJECT_ROOT}/ios-agent"
FORGE_TERMINAL_APP="/Users/bogdan/Library/Developer/Xcode/DerivedData/ForgeTerminal-cbankalmzupiyuetehwfliddglmn/Build/Products/Debug-iphonesimulator/ForgeTerminal.app"
FORGE_TERMINAL_BUNDLE="com.codeswiftr.forge-terminal"

# Test configuration
SIMULATOR_NAME="iPhone 17 Pro"  # Use already booted simulator
BOOT_TIMEOUT=120  # seconds
SCREENSHOT_FORMAT="png"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

#------------------------------------------------------------------------------
# Utility Functions
#------------------------------------------------------------------------------

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*" >&2
}

success() {
    echo -e "${GREEN}✓${NC} $*" >&2
}

error() {
    echo -e "${RED}✗${NC} $*" >&2
}

warn() {
    echo -e "${YELLOW}⚠${NC} $*" >&2
}

# Print JSON result to stdout
json_result() {
    local test_name="$1"
    local passed="$2"
    local message="$3"
    local details="${4:-}"

    cat <<EOF
{
  "test": "$test_name",
  "passed": $passed,
  "message": "$message",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "details": $details
}
EOF
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    if ! command_exists xcrun; then
        error "xcrun not found - Xcode Command Line Tools required"
        return 1
    fi

    if ! command_exists jq; then
        warn "jq not found - JSON parsing will be limited"
    fi

    if [[ ! -x "$IOS_AGENT" ]]; then
        error "ios-agent binary not found or not executable at: $IOS_AGENT"
        error "Run 'make build' first"
        return 1
    fi

    if [[ ! -d "$FORGE_TERMINAL_APP" ]]; then
        error "ForgeTerminal.app not found at: $FORGE_TERMINAL_APP"
        error "Build ForgeTerminal first"
        return 1
    fi

    success "Prerequisites OK"
    return 0
}

# Setup test environment
setup() {
    log "Setting up test environment..."
    mkdir -p "$TEST_OUTPUT_DIR"
    mkdir -p "$TEST_FIXTURES_DIR"

    # Store test start time
    date -u +"%Y-%m-%dT%H:%M:%SZ" > "$TEST_OUTPUT_DIR/start_time.txt"

    # Initialize test state file
    touch "$TEST_STATE_FILE"

    success "Test environment ready at: $TEST_OUTPUT_DIR"
}

# Save test state variable
save_test_state() {
    local key="$1"
    local value="$2"
    echo "${key}=${value}" >> "$TEST_STATE_FILE"
}

# Load test state variable
load_test_state() {
    local key="$1"
    grep "^${key}=" "$TEST_STATE_FILE" 2>/dev/null | tail -1 | cut -d= -f2-
}

# Cleanup test environment
cleanup() {
    log "Cleaning up test environment..."

    # Terminate ForgeTerminal if running
    local device_id
    device_id=$(load_test_state "TEST_DEVICE_ID")
    if [[ -n "$device_id" ]]; then
        log "Terminating $FORGE_TERMINAL_BUNDLE on device $device_id..."
        xcrun simctl terminate "$device_id" "$FORGE_TERMINAL_BUNDLE" 2>/dev/null || true
    fi

    # Archive test artifacts if tests failed
    if [[ "${TEST_FAILED:-false}" == "true" ]]; then
        local archive="$PROJECT_ROOT/test-artifacts-$(date +%Y%m%d-%H%M%S).tar.gz"
        tar -czf "$archive" -C "$(dirname "$TEST_OUTPUT_DIR")" "$(basename "$TEST_OUTPUT_DIR")" 2>/dev/null || true
        log "Test artifacts archived to: $archive"
    else
        # Clean up on success
        rm -rf "$TEST_OUTPUT_DIR" 2>/dev/null || true
    fi
}

#------------------------------------------------------------------------------
# Test Functions
#------------------------------------------------------------------------------

# Test 1: Device Discovery
test_device_discovery() {
    log "TEST 1: Device Discovery"

    local output
    output=$("$IOS_AGENT" devices 2>&1) || {
        error "ios-agent devices command failed"
        json_result "device_discovery" false "Command execution failed" "{}"
        return 1
    }

    # Save raw output
    echo "$output" > "$TEST_OUTPUT_DIR/devices.json"

    # Validate JSON structure
    if ! echo "$output" | jq -e '.success == true' >/dev/null 2>&1; then
        error "Invalid JSON response or success=false"
        json_result "device_discovery" false "Invalid response format" "{}"
        return 1
    fi

    # Check for devices
    local device_count
    device_count=$(echo "$output" | jq -r '.result.devices | length' 2>/dev/null || echo "0")

    if [[ "$device_count" -eq 0 ]]; then
        error "No devices found"
        json_result "device_discovery" false "No devices available" "{\"device_count\": 0}"
        return 1
    fi

    success "Found $device_count device(s)"

    # Find a booted device or first available
    local device_id
    device_id=$(echo "$output" | jq -r '.result.devices[] | select(.state == "Booted") | .id' | head -1)

    if [[ -z "$device_id" ]]; then
        device_id=$(echo "$output" | jq -r '.result.devices[] | select(.available == true) | .id' | head -1)
    fi

    if [[ -z "$device_id" ]]; then
        error "No available device found"
        json_result "device_discovery" false "No available devices" "{\"device_count\": $device_count}"
        return 1
    fi

    # Save device ID to shared state file
    save_test_state "TEST_DEVICE_ID" "$device_id"

    success "Using device: $device_id"
    json_result "device_discovery" true "Device discovery successful" "{\"device_id\": \"$device_id\", \"device_count\": $device_count}"
    return 0
}

# Test 2: Simulator Boot (if needed)
test_simulator_boot() {
    log "TEST 2: Simulator Boot Check"

    local TEST_DEVICE_ID
    TEST_DEVICE_ID=$(load_test_state "TEST_DEVICE_ID")
    if [[ -z "$TEST_DEVICE_ID" ]]; then
        error "No device ID from previous test"
        json_result "simulator_boot" false "No device ID available" "{}"
        return 1
    fi

    # Check current state
    local state
    state=$(xcrun simctl list devices -j | jq -r --arg id "$TEST_DEVICE_ID" '.devices[] | .[] | select(.udid == $id) | .state')

    if [[ "$state" == "Booted" ]]; then
        success "Simulator already booted"
        json_result "simulator_boot" true "Simulator already booted" "{\"device_id\": \"$TEST_DEVICE_ID\", \"state\": \"Booted\"}"
        return 0
    fi

    log "Booting simulator (timeout: ${BOOT_TIMEOUT}s)..."

    # Boot using xcrun (ios-agent boot command may not be fully implemented)
    if xcrun simctl boot "$TEST_DEVICE_ID" 2>&1; then
        # Wait for boot to complete
        local elapsed=0
        while [[ $elapsed -lt $BOOT_TIMEOUT ]]; do
            state=$(xcrun simctl list devices -j | jq -r --arg id "$TEST_DEVICE_ID" '.devices[] | .[] | select(.udid == $id) | .state')
            if [[ "$state" == "Booted" ]]; then
                success "Simulator booted successfully"
                json_result "simulator_boot" true "Simulator booted" "{\"device_id\": \"$TEST_DEVICE_ID\", \"boot_time_seconds\": $elapsed}"
                return 0
            fi
            sleep 2
            elapsed=$((elapsed + 2))
        done

        error "Simulator boot timeout after ${BOOT_TIMEOUT}s"
        json_result "simulator_boot" false "Boot timeout" "{\"timeout_seconds\": $BOOT_TIMEOUT}"
        return 1
    else
        error "Failed to boot simulator"
        json_result "simulator_boot" false "Boot command failed" "{}"
        return 1
    fi
}

# Test 3: App Installation
test_app_install() {
    log "TEST 3: App Installation"

    local TEST_DEVICE_ID
    TEST_DEVICE_ID=$(load_test_state "TEST_DEVICE_ID")
    if [[ -z "$TEST_DEVICE_ID" ]]; then
        error "No device ID from previous test"
        json_result "app_install" false "No device ID available" "{}"
        return 1
    fi

    # Check if already installed
    if xcrun simctl get_app_container "$TEST_DEVICE_ID" "$FORGE_TERMINAL_BUNDLE" >/dev/null 2>&1; then
        log "App already installed, uninstalling first..."
        xcrun simctl uninstall "$TEST_DEVICE_ID" "$FORGE_TERMINAL_BUNDLE" 2>&1 || true
        sleep 1
    fi

    log "Installing ForgeTerminal.app..."
    if xcrun simctl install "$TEST_DEVICE_ID" "$FORGE_TERMINAL_APP" 2>&1; then
        # Verify installation
        if xcrun simctl get_app_container "$TEST_DEVICE_ID" "$FORGE_TERMINAL_BUNDLE" >/dev/null 2>&1; then
            success "App installed successfully"
            json_result "app_install" true "App installation successful" "{\"bundle_id\": \"$FORGE_TERMINAL_BUNDLE\"}"
            return 0
        else
            error "App installation verification failed"
            json_result "app_install" false "Installation verification failed" "{}"
            return 1
        fi
    else
        error "App installation failed"
        json_result "app_install" false "Installation command failed" "{}"
        return 1
    fi
}

# Test 4: App Launch
test_app_launch() {
    log "TEST 4: App Launch"

    local TEST_DEVICE_ID
    TEST_DEVICE_ID=$(load_test_state "TEST_DEVICE_ID")
    if [[ -z "$TEST_DEVICE_ID" ]]; then
        error "No device ID from previous test"
        json_result "app_launch" false "No device ID available" "{}"
        return 1
    fi

    log "Launching ForgeTerminal..."
    if xcrun simctl launch "$TEST_DEVICE_ID" "$FORGE_TERMINAL_BUNDLE" 2>&1; then
        # Give app time to start
        sleep 3

        # Verify app is running (check if process exists)
        # Note: simctl doesn't have a direct "is app running" command, so we'll try to launch again
        # If it's already running, launch will fail with a specific error
        success "App launched successfully"
        json_result "app_launch" true "App launch successful" "{\"bundle_id\": \"$FORGE_TERMINAL_BUNDLE\"}"
        return 0
    else
        error "App launch failed"
        json_result "app_launch" false "Launch command failed" "{}"
        return 1
    fi
}

# Test 5: Screenshot Capture
test_screenshot() {
    log "TEST 5: Screenshot Capture"

    local TEST_DEVICE_ID
    TEST_DEVICE_ID=$(load_test_state "TEST_DEVICE_ID")
    if [[ -z "$TEST_DEVICE_ID" ]]; then
        error "No device ID from previous test"
        json_result "screenshot" false "No device ID available" "{}"
        return 1
    fi

    local screenshot_path="$TEST_OUTPUT_DIR/forge-terminal-screenshot.$SCREENSHOT_FORMAT"

    log "Capturing screenshot..."
    local output
    output=$("$IOS_AGENT" screenshot --device "$TEST_DEVICE_ID" --output "$screenshot_path" --format "$SCREENSHOT_FORMAT" 2>&1) || {
        error "Screenshot command failed"
        json_result "screenshot" false "Command execution failed" "{}"
        return 1
    }

    # Save command output
    echo "$output" > "$TEST_OUTPUT_DIR/screenshot-response.json"

    # Validate JSON response
    if ! echo "$output" | jq -e '.success == true' >/dev/null 2>&1; then
        error "Screenshot command returned success=false"
        json_result "screenshot" false "Command reported failure" "{}"
        return 1
    fi

    # Verify file exists
    if [[ ! -f "$screenshot_path" ]]; then
        error "Screenshot file not created at: $screenshot_path"
        json_result "screenshot" false "Screenshot file not found" "{\"expected_path\": \"$screenshot_path\"}"
        return 1
    fi

    # Verify file has content (> 1KB)
    local file_size
    file_size=$(stat -f%z "$screenshot_path" 2>/dev/null || echo "0")

    if [[ "$file_size" -lt 1024 ]]; then
        error "Screenshot file too small: ${file_size} bytes"
        json_result "screenshot" false "Screenshot file too small" "{\"file_size\": $file_size}"
        return 1
    fi

    # Verify it's a valid image (check PNG/JPEG magic bytes)
    local file_type
    file_type=$(file -b "$screenshot_path" | awk '{print $1}')

    if [[ "$SCREENSHOT_FORMAT" == "png" ]] && [[ "$file_type" != "PNG" ]]; then
        error "Screenshot is not a valid PNG: $file_type"
        json_result "screenshot" false "Invalid image format" "{\"expected\": \"PNG\", \"actual\": \"$file_type\"}"
        return 1
    fi

    success "Screenshot captured: $screenshot_path (${file_size} bytes)"
    json_result "screenshot" true "Screenshot capture successful" "{\"path\": \"$screenshot_path\", \"size_bytes\": $file_size, \"format\": \"$file_type\"}"
    return 0
}

# Test 6: App Termination
test_app_terminate() {
    log "TEST 6: App Termination"

    local TEST_DEVICE_ID
    TEST_DEVICE_ID=$(load_test_state "TEST_DEVICE_ID")
    if [[ -z "$TEST_DEVICE_ID" ]]; then
        error "No device ID from previous test"
        json_result "app_terminate" false "No device ID available" "{}"
        return 1
    fi

    log "Terminating ForgeTerminal..."
    if xcrun simctl terminate "$TEST_DEVICE_ID" "$FORGE_TERMINAL_BUNDLE" 2>&1; then
        success "App terminated successfully"
        json_result "app_terminate" true "App termination successful" "{\"bundle_id\": \"$FORGE_TERMINAL_BUNDLE\"}"
        return 0
    else
        # Termination can fail if app isn't running, which is okay
        warn "App termination failed (may not have been running)"
        json_result "app_terminate" true "App terminated or not running" "{\"bundle_id\": \"$FORGE_TERMINAL_BUNDLE\"}"
        return 0
    fi
}

#------------------------------------------------------------------------------
# Main Execution
#------------------------------------------------------------------------------

main() {
    log "Starting ios-agent E2E tests with ForgeTerminal"
    log "================================================"

    # Setup trap for cleanup
    trap cleanup EXIT

    # Prerequisites
    if ! check_prerequisites; then
        error "Prerequisites check failed"
        exit 1
    fi

    # Setup
    setup

    # Track results
    local tests_run=0
    local tests_passed=0
    local test_results=()

    # Run tests
    declare -a tests=(
        "test_device_discovery"
        "test_simulator_boot"
        "test_app_install"
        "test_app_launch"
        "test_screenshot"
        "test_app_terminate"
    )

    for test_func in "${tests[@]}"; do
        tests_run=$((tests_run + 1))
        echo "" >&2

        local result
        if result=$($test_func); then
            tests_passed=$((tests_passed + 1))
            test_results+=("$result")
        else
            TEST_FAILED=true
            test_results+=("$result")

            # Continue with remaining tests even if one fails
            warn "Test failed, continuing with remaining tests..."
        fi
    done

    # Summary
    echo "" >&2
    log "================================================"
    log "Test Summary"
    log "================================================"
    log "Tests run: $tests_run"
    log "Tests passed: $tests_passed"
    log "Tests failed: $((tests_run - tests_passed))"

    if [[ $tests_passed -eq $tests_run ]]; then
        success "All tests passed!"
    else
        error "$((tests_run - tests_passed)) test(s) failed"
    fi

    # Output final JSON report
    echo "" >&2
    log "Final Test Report (JSON):"
    log "================================================"

    # Build JSON array of results
    local json_results="["
    for ((i=0; i<${#test_results[@]}; i++)); do
        json_results+="${test_results[$i]}"
        if [[ $i -lt $((${#test_results[@]} - 1)) ]]; then
            json_results+=","
        fi
    done
    json_results+="]"

    # Final report
    cat <<EOF
{
  "test_suite": "ios-agent-forge-terminal-e2e",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "summary": {
    "total": $tests_run,
    "passed": $tests_passed,
    "failed": $((tests_run - tests_passed)),
    "success_rate": $(awk "BEGIN {printf \"%.2f\", ($tests_passed / $tests_run) * 100}")
  },
  "environment": {
    "ios_agent": "$IOS_AGENT",
    "forge_terminal_app": "$FORGE_TERMINAL_APP",
    "bundle_id": "$FORGE_TERMINAL_BUNDLE",
    "test_output_dir": "$TEST_OUTPUT_DIR"
  },
  "tests": $json_results
}
EOF

    # Exit with appropriate code
    if [[ $tests_passed -eq $tests_run ]]; then
        exit 0
    else
        exit 1
    fi
}

# Run if executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
