# ios-agent-cli Test Summary

Complete test infrastructure for ios-agent-cli with ForgeTerminal app integration.

## Test Directory Structure

```
test/
├── e2e/                              # End-to-end tests
│   ├── forge_terminal_test.sh        # Main E2E test suite
│   ├── run_single_test.sh            # Single test runner
│   ├── README.md                     # Detailed documentation
│   ├── QUICKSTART.md                 # Quick reference guide
│   └── TEST_MANIFEST.md              # Test coverage matrix
├── fixtures/                         # Test fixtures and reference data
│   └── expected-device-schema.json   # JSON schema for device discovery
├── integration/                      # Integration tests (Go)
└── mocks/                            # Test mocks
```

## Quick Commands

### Run Full E2E Suite
```bash
# Using Makefile (recommended)
make e2e-test

# Direct execution
./test/e2e/forge_terminal_test.sh
```

### Run Single Test
```bash
./test/e2e/run_single_test.sh test_device_discovery
./test/e2e/run_single_test.sh test_screenshot
```

### View Results
```bash
# Pretty print JSON results
./test/e2e/forge_terminal_test.sh 2>/dev/null | jq '.'

# Summary only
./test/e2e/forge_terminal_test.sh 2>/dev/null | jq '.summary'

# Save results
./test/e2e/forge_terminal_test.sh > results.json 2>test.log
```

## Test Coverage

### Currently Implemented (6 tests)

| Test | Command Used | Status |
|------|-------------|--------|
| Device Discovery | `ios-agent devices` | ✅ Complete |
| Simulator Boot | `xcrun simctl boot` | ✅ Complete |
| App Installation | `xcrun simctl install` | ✅ Complete |
| App Launch | `xcrun simctl launch` | ✅ Complete |
| Screenshot Capture | `ios-agent screenshot` | ✅ Complete |
| App Termination | `xcrun simctl terminate` | ✅ Complete |

### Planned (Next Phase)

| Test | Command | Priority |
|------|---------|----------|
| UI Tap | `ios-agent io tap` | P1 |
| Text Input | `ios-agent io text` | P1 |
| Device State | `ios-agent state` | P1 |
| Swipe Gesture | `ios-agent io swipe` | P2 |
| Button Press | `ios-agent io button` | P2 |
| App Uninstall | `ios-agent app uninstall` | P2 |

## Test Features

### Robust Error Handling
- Prerequisites validation (Xcode, simulators, ForgeTerminal)
- Graceful fallbacks for missing commands
- Detailed error messages with context
- Automatic cleanup on failure

### Comprehensive Reporting
- JSON output for programmatic consumption
- Human-readable logs to stderr
- Test artifacts archived on failure
- Execution time tracking

### Agent-Friendly Design
- All commands return structured JSON
- Exit codes follow standard conventions (0=success, 1=failure)
- Idempotent test execution
- No side effects between test runs

### Extensibility
- Easy to add new test scenarios
- Modular test functions
- Shared utility functions
- Clear documentation for contributions

## Prerequisites

1. **Xcode Command Line Tools**
   ```bash
   xcode-select --install
   ```

2. **iOS Simulator**
   - At least one iOS simulator created
   - Check: `xcrun simctl list devices available`

3. **ForgeTerminal App**
   - Built in Xcode
   - Location: See `FORGE_TERMINAL_APP` in test script

4. **jq (Optional)**
   ```bash
   brew install jq
   ```

## Output Format

### JSON Response Structure
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
  "environment": {
    "ios_agent": "/path/to/ios-agent",
    "forge_terminal_app": "/path/to/ForgeTerminal.app",
    "bundle_id": "com.codeswiftr.forge-terminal",
    "test_output_dir": "/tmp/ios-agent-test-12345"
  },
  "tests": [
    {
      "test": "device_discovery",
      "passed": true,
      "message": "Device discovery successful",
      "timestamp": "2026-02-04T17:30:01Z",
      "details": {
        "device_id": "C160E86C-75C2-4DB0-BE4C-C5181D1B245D",
        "device_count": 11
      }
    }
  ]
}
```

### Individual Test Result
```json
{
  "test": "screenshot",
  "passed": true,
  "message": "Screenshot capture successful",
  "timestamp": "2026-02-04T17:30:15Z",
  "details": {
    "path": "/tmp/ios-agent-test-12345/forge-terminal-screenshot.png",
    "size_bytes": 54321,
    "format": "PNG"
  }
}
```

## Integration with Development Workflow

### Local Development
```bash
# Build and test
make build
make e2e-test

# Quick iteration
make build && ./test/e2e/run_single_test.sh test_screenshot
```

### CI/CD Pipeline
```yaml
# GitHub Actions example
- name: Run E2E Tests
  run: make e2e-test

- name: Upload test results
  if: always()
  uses: actions/upload-artifact@v3
  with:
    name: e2e-results
    path: |
      results.json
      test-artifacts-*.tar.gz
```

### Pre-commit Hook
```bash
# .git/hooks/pre-commit
#!/bin/bash
make build && ./test/e2e/forge_terminal_test.sh >/dev/null
```

## Troubleshooting

### Common Issues

**"No devices found"**
- Solution: Create simulator in Xcode or run `xcrun simctl create "iPhone 15" com.apple.CoreSimulator.SimDeviceType.iPhone-15`

**"ForgeTerminal.app not found"**
- Solution: Build ForgeTerminal in Xcode first
- Update `FORGE_TERMINAL_APP` path in test script if DerivedData location changed

**"Screenshot command failed"**
- Solution: Verify simulator is booted and ios-agent built correctly
- Test manually: `./ios-agent screenshot --device <id> --output test.png`

**"Boot timeout"**
- Solution: Increase `BOOT_TIMEOUT` in script (default: 120s)
- Check simulator logs: `~/Library/Logs/CoreSimulator/`

### Debug Mode

```bash
# Enable verbose output
bash -x ./test/e2e/forge_terminal_test.sh

# Run with custom output directory
TEST_OUTPUT_DIR=/tmp/debug ./test/e2e/forge_terminal_test.sh
```

## Performance Benchmarks

Expected execution times on M1/M2 Mac:

| Test | Time (cold) | Time (warm) |
|------|-------------|-------------|
| Device Discovery | 2s | 1s |
| Simulator Boot | 60s | 5s (already booted) |
| App Install | 10s | 5s |
| App Launch | 5s | 3s |
| Screenshot | 3s | 2s |
| App Terminate | 2s | 1s |
| **Total** | ~82s | ~17s |

## Next Steps

1. **Implement UI Interaction Tests**
   - Add test_ui_tap, test_text_input
   - Verify element targeting
   - Test gesture recognition

2. **Add State Verification**
   - Implement test_device_state
   - Validate app state JSON structure
   - Test state transitions

3. **Remote Device Support**
   - Add Tailscale integration tests
   - Test remote device discovery
   - Verify remote command execution

4. **Visual Regression Testing**
   - Add screenshot comparison
   - Detect UI changes
   - Generate diff reports

5. **Performance Testing**
   - Add execution time assertions
   - Benchmark command latency
   - Track performance over time

## Documentation

- [E2E Test README](e2e/README.md) - Detailed guide and test patterns
- [Quick Start](e2e/QUICKSTART.md) - One-command test execution
- [Test Manifest](e2e/TEST_MANIFEST.md) - Complete test coverage matrix
- [Main README](../README.md) - Project overview and setup

## Contributing

To add new tests:

1. Study existing test functions in `forge_terminal_test.sh`
2. Follow the test function pattern
3. Update the `tests` array in main()
4. Add documentation to TEST_MANIFEST.md
5. Test locally before submitting PR

Example test template:
```bash
test_new_feature() {
    log "TEST X: New Feature"

    if [[ -z "${TEST_DEVICE_ID:-}" ]]; then
        error "No device ID from previous test"
        json_result "new_feature" false "No device ID available" "{}"
        return 1
    fi

    # Test implementation
    local output
    output=$("$IOS_AGENT" new-command --device "$TEST_DEVICE_ID" 2>&1) || {
        error "Command failed"
        json_result "new_feature" false "Command execution failed" "{}"
        return 1
    }

    # Validation
    if echo "$output" | jq -e '.success == true' >/dev/null 2>&1; then
        success "New feature works"
        json_result "new_feature" true "Success" "{}"
        return 0
    else
        error "Validation failed"
        json_result "new_feature" false "Failed" "{}"
        return 1
    fi
}
```

## License

MIT License - see [LICENSE](../LICENSE) for details.

---

**Built for FORGE Portfolio** | [NeoForge Dev](https://neoforge.dev)
