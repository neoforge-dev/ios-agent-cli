# Integration Tests for ios-agent-cli

This document describes the integration tests for ios-agent-cli that run against real iOS simulators.

## Overview

The integration tests validate that ios-agent-cli correctly interacts with real iOS simulators using the `xcrun simctl` interface. These tests are designed to be idempotent, gracefully handle missing simulators, and clean up resources after execution.

## Test File

- **Location:** `/test/integration_test.go`
- **Build Tag:** `//go:build integration`

The build tag ensures these tests are skipped during normal `go test` runs and only executed when explicitly requested.

## Running Integration Tests

### Via Makefile (Recommended)

```bash
make integration-test
```

### Via go test

```bash
go test -tags=integration ./test/integration_test.go -v -timeout 3m
```

### Run Specific Test

```bash
go test -tags=integration -run TestIntegration_DeviceDiscovery ./test/integration_test.go -v
```

## Test Scenarios

### 1. Device Discovery Tests (`TestIntegration_DeviceDiscovery`)

Validates device discovery and lookup functionality:

- **list devices returns valid simulators** - Verifies all discovered simulators have required fields
- **get device by ID returns correct device** - Tests lookup by device ID
- **get device by UDID returns correct device** - Tests lookup by UDID
- **find device by name returns correct device** - Tests name-based device search
- **get nonexistent device returns error** - Validates error handling for invalid IDs
- **find device by nonexistent name returns error** - Validates error handling for invalid names

**Duration:** ~1 second
**Requirements:** At least one simulator installed

### 2. Simulator Boot/Shutdown Lifecycle (`TestIntegration_SimulatorBootShutdownLifecycle`)

Tests the complete lifecycle of booting and shutting down a simulator:

- **boot simulator from shutdown state** - Boots a shutdown simulator and polls until booted
- **boot already booted simulator returns error** - Validates error handling
- **shutdown booted simulator** - Shuts down a booted simulator
- **shutdown already shutdown simulator returns error** - Validates error handling

**Duration:** ~10 seconds (includes boot/shutdown time)
**Requirements:** At least one shutdown simulator
**Cleanup:** Automatically restores simulator to original state

### 3. Screenshot Capture (`TestIntegration_ScreenshotCapture`)

Validates screenshot capture functionality:

- **capture screenshot creates file** - Verifies screenshot file is created with correct metadata
- **capture screenshot with custom filename** - Tests custom file naming
- **capture screenshot to invalid path returns error** - Validates error handling

**Duration:** ~1 second
**Requirements:** At least one booted simulator
**Note:** If no booted simulator is available, test is skipped

### 4. Basic UI Interaction (`TestIntegration_BasicUIInteraction`)

Tests basic UI interaction capabilities:

- **type text into simulator** - Tests text input via `simctl keyboardinput` (if available)
- **type special characters** - Tests typing special characters and email addresses
- **press home button** - Tests hardware button simulation (if available)
- **press invalid button returns error** - Validates error handling

**Duration:** ~1 second
**Requirements:** At least one booted simulator
**Note:** Some UI commands may not be available on all Xcode versions. Tests skip gracefully.

### 5. Device State Polling (`TestIntegration_DeviceStatePolling`)

Validates device state polling functionality:

- **get device state for existing device** - Tests state retrieval for a single device
- **get device state for nonexistent device** - Validates error handling
- **poll device states for all devices** - Tests polling multiple devices repeatedly

**Duration:** ~10 seconds (includes polling delays)
**Requirements:** At least one simulator installed

### 6. Concurrent Device Operations (`TestIntegration_ConcurrentDeviceOperations`)

Tests thread-safety and concurrent access:

- **concurrent device list operations** - 50 concurrent ListDevices() calls
- **concurrent device get operations** - 10 concurrent GetDevice() calls

**Duration:** ~2 seconds
**Requirements:** At least one simulator installed
**Purpose:** Ensures no race conditions or crashes under concurrent access

### 7. Error Handling (`TestIntegration_ErrorHandling`)

Validates error handling for various failure scenarios:

- **boot nonexistent simulator** - Tests error for invalid UDID
- **shutdown nonexistent simulator** - Tests error for invalid UDID
- **get state of nonexistent simulator** - Tests error for invalid UDID
- **screenshot nonexistent simulator** - Tests error for invalid UDID
- **type text to nonexistent simulator** - Tests error for invalid UDID (if command available)

**Duration:** ~1 second
**Requirements:** None (intentionally tests failures)

## Test Design Principles

### 1. Graceful Skipping

Tests check for simulator availability and skip gracefully if requirements are not met:

```go
func setupTestEnvironment(t *testing.T) (*device.LocalManager, []device.Device) {
    devices, err := manager.ListDevices()
    if err != nil {
        t.Skipf("Cannot list devices, skipping test: %v", err)
        return nil, nil
    }
    if len(devices) == 0 {
        t.Skip("No simulators available, skipping test")
        return nil, nil
    }
    return manager, devices
}
```

### 2. Idempotent Tests

Tests are designed to be repeatable without side effects:

- Boot/shutdown tests restore original device state in `defer` cleanup
- Screenshot tests use `t.TempDir()` for automatic cleanup
- Tests check current state before attempting operations

### 3. Version Compatibility

Tests detect and handle Xcode version differences:

- **keyboardinput command** - Not available on all Xcode versions
- **ui click home command** - Not available on all Xcode versions

Tests skip gracefully with informative messages when commands are unavailable.

### 4. Resource Cleanup

All tests clean up resources:

```go
defer func() {
    if originalState == device.StateShutdown {
        t.Logf("Cleaning up: shutting down simulator %s", deviceID)
        _ = manager.ShutdownSimulator(deviceID)
    }
}()
```

### 5. Timeouts

Tests use appropriate timeouts:

- Device operations: 1-3 minutes
- Boot operations: 60 seconds max wait
- Shutdown operations: 30 seconds max wait
- Overall test suite: 3 minutes

## Prerequisites

### Required

- macOS with Xcode installed
- At least one iOS simulator configured
- `xcrun` command available in PATH

### Optional (for full test coverage)

- At least one booted simulator (for screenshot and UI interaction tests)
- At least one shutdown simulator (for boot/shutdown lifecycle tests)

## Checking Simulator Availability

```bash
# List all simulators
xcrun simctl list devices available

# Boot a simulator manually (if needed)
xcrun simctl boot <UDID>

# Shutdown a simulator
xcrun simctl shutdown <UDID>
```

## Continuous Integration

For CI environments:

```yaml
# GitHub Actions example
- name: List Available Simulators
  run: xcrun simctl list devices available

- name: Boot Test Simulator
  run: xcrun simctl boot "iPhone 15 Pro"

- name: Run Integration Tests
  run: make integration-test
  timeout-minutes: 5
```

## Test Output

### Successful Run

```
=== RUN   TestIntegration_DeviceDiscovery
--- PASS: TestIntegration_DeviceDiscovery (0.90s)
=== RUN   TestIntegration_SimulatorBootShutdownLifecycle
    integration_test.go:153: Testing with device: iPhone 17 Pro
    integration_test.go:188: Boot completed in 2.8s
    integration_test.go:232: Shutdown completed in 4.8s
--- PASS: TestIntegration_SimulatorBootShutdownLifecycle (10.68s)
...
PASS
ok  	command-line-arguments	26.308s
```

### Skipped Tests

```
=== RUN   TestIntegration_BasicUIInteraction/type_text_into_simulator
    integration_test.go:342: keyboardinput command not available on this Xcode version
--- SKIP: TestIntegration_BasicUIInteraction/type_text_into_simulator (0.11s)
```

## Troubleshooting

### No Simulators Found

**Error:** "No simulators available, skipping test"

**Solution:** Create a simulator in Xcode or via command line:

```bash
xcrun simctl create "iPhone 15 Pro" "com.apple.CoreSimulator.SimDeviceType.iPhone-15-Pro" "com.apple.CoreSimulator.SimRuntime.iOS-17-4"
```

### Boot/Shutdown Tests Timeout

**Error:** "Simulator should complete boot within timeout"

**Possible Causes:**
- System resources constrained
- Simulator hung or crashed
- Xcode not properly installed

**Solution:**
- Increase timeout in test code
- Restart simulators manually
- Reinstall Xcode Command Line Tools

### Screenshot Tests Fail

**Error:** "No booted simulators available for screenshot test"

**Solution:** Boot a simulator before running tests:

```bash
xcrun simctl boot "iPhone 15 Pro"
```

## Test Coverage

Integration tests cover:

- ✅ Device discovery (listing, lookup by ID/UDID/name)
- ✅ Simulator lifecycle (boot, shutdown, state polling)
- ✅ Screenshot capture
- ✅ Basic UI interaction (text input, button presses)
- ✅ Concurrent operations
- ✅ Error handling for invalid operations

Not covered (requires additional setup):

- ❌ App installation/launch/termination (requires .app bundle)
- ❌ Tap/swipe gestures (requires mobilecli or AppleScript)
- ❌ Video recording
- ❌ Remote device access via Tailscale

## Future Enhancements

1. **Add app lifecycle tests** - Create fixture .app bundles for testing install/launch/terminate
2. **Add gesture tests** - Integrate mobilecli for tap/swipe validation
3. **Add performance benchmarks** - Measure operation timing
4. **Add parallel test execution** - Speed up test suite with t.Parallel()
5. **Add test fixtures** - Pre-defined device states for consistent testing

## Related Documentation

- [Test Summary](TEST_SUMMARY.md) - Overview of all test types
- [E2E Tests](e2e/README.md) - End-to-end tests with ForgeTerminal app
- [CLAUDE.md](../CLAUDE.md) - Project-level development guide
