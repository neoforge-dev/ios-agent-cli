# E2E Test Quick Start

## One-Command Test Run

```bash
make e2e-test
```

This will:
1. Build ios-agent binary
2. Run all E2E tests with ForgeTerminal
3. Output JSON results

## Manual Test Run

```bash
# 1. Build ios-agent
make build

# 2. Run tests
./test/e2e/forge_terminal_test.sh
```

## Reading Test Results

The script outputs JSON to stdout and logs to stderr:

```bash
# Capture JSON results
./test/e2e/forge_terminal_test.sh > results.json 2>test.log

# Pretty print results
./test/e2e/forge_terminal_test.sh 2>/dev/null | jq '.'

# Check summary only
./test/e2e/forge_terminal_test.sh 2>/dev/null | jq '.summary'
```

## Test Output Example

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
    // ... more test results
  ]
}
```

## Prerequisites Check

```bash
# Check Xcode CLI tools
xcode-select -p

# Check simulators
xcrun simctl list devices available

# Check ForgeTerminal app
ls -l /Users/bogdan/Library/Developer/Xcode/DerivedData/ForgeTerminal-*/Build/Products/Debug-iphonesimulator/ForgeTerminal.app

# Install jq (optional but recommended)
brew install jq
```

## Troubleshooting

### No devices found
```bash
# List available simulators
xcrun simctl list devices available

# Create a new simulator in Xcode if needed
# Xcode > Window > Devices and Simulators > Simulators > +
```

### ForgeTerminal not found
```bash
# Build ForgeTerminal in Xcode first
# The DerivedData path may change - update FORGE_TERMINAL_APP in the script
```

### Screenshot fails
```bash
# Test screenshot manually
./ios-agent screenshot --device <device-id> --output test.png

# Verify device is booted
xcrun simctl list | grep Booted
```

## Environment Variables

```bash
# Custom output directory
TEST_OUTPUT_DIR=/tmp/my-test-output ./test/e2e/forge_terminal_test.sh
```

## CI/CD Integration

```yaml
# .github/workflows/e2e.yml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run E2E Tests
        run: make e2e-test

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: e2e-results
          path: |
            test-results.json
            test-artifacts-*.tar.gz
```

## Adding New Tests

See [README.md](README.md#expanding-the-test-suite) for test template and integration guide.
