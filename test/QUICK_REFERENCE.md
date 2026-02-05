# Integration Tests - Quick Reference Card

## Run Tests

```bash
# Run all integration tests (recommended)
make integration-test

# Run all with verbose output
go test -tags=integration ./test/integration_test.go -v -timeout 3m

# Run specific test
go test -tags=integration -run TestIntegration_DeviceDiscovery ./test/integration_test.go -v

# Run with race detector
go test -tags=integration -race ./test/integration_test.go -v
```

## Test Structure

| Test | Scenarios | Duration | Prerequisites |
|------|-----------|----------|---------------|
| DeviceDiscovery | 6 | ~1s | 1+ simulator |
| SimulatorBootShutdownLifecycle | 4 | ~10s | 1+ shutdown simulator |
| ScreenshotCapture | 3 | ~1s | 1+ booted simulator |
| BasicUIInteraction | 4 | ~1s | 1+ booted simulator |
| DeviceStatePolling | 3 | ~10s | 1+ simulator |
| ConcurrentDeviceOperations | 2 | ~2s | 1+ simulator |
| ErrorHandling | 5 | ~1s | None |

**Total:** 7 tests, 26 scenarios, ~25 seconds

## Prerequisites Check

```bash
# Check if simulators are available
xcrun simctl list devices available

# Boot a simulator (if needed)
xcrun simctl boot "iPhone 15 Pro"

# Shutdown a simulator (if needed)
xcrun simctl shutdown "iPhone 15 Pro"

# Check Xcode version
xcodebuild -version
```

## Common Issues

### No Simulators Found
**Error:** "No simulators available, skipping test"
**Fix:** Install Xcode and create a simulator

### Tests Skip Due to Xcode Version
**Message:** "keyboardinput command not available on this Xcode version"
**Status:** Normal - tests skip gracefully on older Xcode versions

### Boot/Shutdown Timeout
**Error:** "Simulator should complete boot within timeout"
**Fix:**
- Increase timeout in code
- Close other apps to free resources
- Restart Xcode/Simulator

## Test Output

### All Pass
```
PASS
ok  	command-line-arguments	22.861s
```

### With Skips (Normal)
```
--- SKIP: TestIntegration_BasicUIInteraction/type_text_into_simulator (0.11s)
    integration_test.go:342: keyboardinput command not available on this Xcode version
PASS
ok  	command-line-arguments	22.861s
```

## CI/CD Integration

```yaml
# GitHub Actions
- name: Install Xcode
  uses: maxim-lobanov/setup-xcode@v1
  with:
    xcode-version: latest-stable

- name: List Simulators
  run: xcrun simctl list devices available

- name: Boot Test Simulator
  run: xcrun simctl boot "iPhone 15 Pro"

- name: Run Integration Tests
  run: make integration-test
  timeout-minutes: 5
```

## File Locations

- **Test File:** `/test/integration_test.go`
- **Documentation:** `/test/INTEGRATION_TESTS.md`
- **Summary:** `/test/IMPLEMENTATION_SUMMARY.md`
- **This Card:** `/test/QUICK_REFERENCE.md`

## Key Metrics

- **Total Tests:** 7
- **Total Scenarios:** 26
- **Code Lines:** ~625
- **Duration:** ~25 seconds
- **Pass Rate:** 100% (pass or graceful skip)

## Build Tags

Tests use `//go:build integration` tag to skip during normal test runs.

```bash
# Skipped (no integration tests run)
go test ./...

# Runs integration tests
go test -tags=integration ./test/...
```

## What's Tested

✅ Device discovery (listing, ID/UDID/name lookup)
✅ Simulator boot/shutdown with state polling
✅ Screenshot capture with file validation
✅ Text input and button presses
✅ Concurrent operations (50+ operations)
✅ Error handling for invalid operations
✅ Resource cleanup
✅ Graceful skipping

## What's Not Tested

❌ App install/launch/terminate (needs .app fixtures)
❌ Tap/swipe gestures (needs mobilecli)
❌ Video recording
❌ Remote devices via Tailscale

## Quick Troubleshooting

| Issue | Solution |
|-------|----------|
| No simulators | Create one in Xcode |
| Boot test skips | Shut down a simulator first |
| Screenshot test skips | Boot a simulator first |
| Tests hang | Kill Simulator.app and retry |
| Xcode errors | Reinstall Xcode CLI tools |

## Contact/Support

For issues with integration tests:
1. Check prerequisites (Xcode, simulators)
2. Review test logs for skip messages
3. Consult `INTEGRATION_TESTS.md` for details
4. Check `features.json` for known issues

---

**Feature ID:** IOS-016
**Status:** Complete ✅
**Last Updated:** 2026-02-04
