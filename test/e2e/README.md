# E2E Tests for ios-agent-cli

End-to-end tests for ios-agent-cli using real iOS simulators and apps.

## Test Suites

### forge_terminal_test.sh

Tests ios-agent functionality with the ForgeTerminal app as a target.

**Prerequisites:**
- Xcode Command Line Tools installed (`xcode-select --install`)
- ForgeTerminal.app built (see path in script)
- At least one iOS simulator available
- `jq` installed (optional, for better JSON parsing): `brew install jq`

**Quick Start:**
```bash
# Build ios-agent first
cd /Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli
make build

# Run tests
./test/e2e/forge_terminal_test.sh
```

**What It Tests:**

1. **Device Discovery** - Verifies `ios-agent devices` returns valid JSON with available simulators
2. **Simulator Boot** - Ensures simulator is booted (or boots it if needed)
3. **App Installation** - Installs ForgeTerminal.app using `xcrun simctl install`
4. **App Launch** - Launches the app and verifies it starts
5. **Screenshot Capture** - Tests `ios-agent screenshot` command
6. **App Termination** - Terminates the app cleanly

**Output:**

The script outputs JSON results to stdout and human-readable logs to stderr.

Success example:
```json
{
  "test_suite": "ios-agent-forge-terminal-e2e",
  "timestamp": "2026-02-04T17:30:00Z",
  "summary": {
    "total": 6,
    "passed": 6,
    "failed": 0,
    "success_rate": 100.00
  },
  "tests": [
    {
      "test": "device_discovery",
      "passed": true,
      "message": "Device discovery successful",
      "timestamp": "2026-02-04T17:30:01Z"
    }
    // ... more test results
  ]
}
```

**Test Artifacts:**

On failure, artifacts are archived to `test-artifacts-TIMESTAMP.tar.gz` containing:
- Device list JSON
- Screenshot attempts
- Command responses
- Timestamps

**Environment Variables:**

- `TEST_OUTPUT_DIR` - Override default temp directory (default: `/tmp/ios-agent-test-$$`)

**Exit Codes:**

- `0` - All tests passed
- `1` - One or more tests failed

## Test Fixtures

The `test/fixtures/` directory stores expected results and reference data:

- `expected-device-schema.json` - JSON schema for device discovery
- `reference-screenshot.png` - Reference screenshot for comparison (future)

## Expanding the Test Suite

As more ios-agent commands are implemented, add new test functions:

```bash
# Test template
test_new_feature() {
    log "TEST X: New Feature"

    if [[ -z "${TEST_DEVICE_ID:-}" ]]; then
        error "No device ID from previous test"
        json_result "new_feature" false "No device ID available" "{}"
        return 1
    fi

    # Test implementation
    if "$IOS_AGENT" new-command --device "$TEST_DEVICE_ID"; then
        success "New feature works"
        json_result "new_feature" true "Success" "{}"
        return 0
    else
        error "New feature failed"
        json_result "new_feature" false "Failed" "{}"
        return 1
    fi
}
```

Then add to the `tests` array in `main()`:
```bash
declare -a tests=(
    # ... existing tests
    "test_new_feature"
)
```

## CI Integration

To integrate with CI/CD:

```bash
# GitHub Actions example
- name: Run E2E Tests
  run: |
    make build
    ./test/e2e/forge_terminal_test.sh > test-results.json
    cat test-results.json

- name: Upload artifacts on failure
  if: failure()
  uses: actions/upload-artifact@v3
  with:
    name: test-artifacts
    path: test-artifacts-*.tar.gz
```

## Troubleshooting

**"No devices found"**
- Ensure Xcode is installed: `xcode-select -p`
- List simulators: `xcrun simctl list devices available`
- Create a simulator in Xcode if none exist

**"Simulator boot timeout"**
- Increase `BOOT_TIMEOUT` in the script (default: 120s)
- Check simulator logs: `~/Library/Logs/CoreSimulator/`

**"ForgeTerminal.app not found"**
- Build the app first in Xcode
- Update `FORGE_TERMINAL_APP` path in the script if DerivedData location changed

**Screenshot command fails**
- Verify simulator is booted: `xcrun simctl list | grep Booted`
- Check ios-agent build: `./ios-agent --version`
- Run screenshot manually: `./ios-agent screenshot --device <id> --output test.png`

## Future Enhancements

- [ ] Add UI interaction tests (tap, swipe, text input)
- [ ] Add app state verification tests
- [ ] Add remote device tests (Tailscale)
- [ ] Add performance benchmarks (command execution time)
- [ ] Add image comparison for screenshots
- [ ] Add parallel test execution
- [ ] Add test retry logic for flaky tests
