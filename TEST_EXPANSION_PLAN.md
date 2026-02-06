# iOS Agent CLI - Test Coverage Expansion Plan

## Executive Summary

The ios-agent-cli currently has moderate test coverage (0-10% on cmd package) focusing primarily on command structure validation and JSON serialization. This plan identifies critical coverage gaps and provides comprehensive test cases for:

1. **io.go** - UI interaction commands (tap, text, swipe, button)
2. **app.go** - Application lifecycle (launch, terminate, install, uninstall)
3. **screenshot.go** - Screenshot capture with path handling
4. **simulator.go** - Simulator boot/shutdown with state management

## Current Coverage Analysis

### Unit Tests Present
- Command structure validation (Cobra command properties)
- Flag definition and shorthand aliases
- JSON serialization for result types
- Device validation in app operations
- Simulator boot polling logic
- Mock device bridge setup

### Critical Gaps Identified

| Component | Coverage | Gap | Priority |
|-----------|----------|-----|----------|
| io.go tap validation | Structure only | No error path tests | HIGH |
| io.go text validation | Structure only | Empty text, special chars | HIGH |
| io.go swipe validation | Flags only | Duration validation, coordinate bounds | HIGH |
| io.go button validation | Flags only | Invalid button type handling | HIGH |
| app.go launch | Basic device checks | Timeout logic, wait flag behavior | MEDIUM |
| app.go terminate | Basic device checks | Already-terminated app | MEDIUM |
| app.go install | Basic device checks | Invalid path, bundle ID extraction | MEDIUM |
| app.go uninstall | Basic device checks | Not-installed bundle handling | MEDIUM |
| screenshot.go | No tests | Output path creation, format validation | HIGH |
| simulator.go | Boot polling tested | Shutdown flow, name resolution | MEDIUM |

## Test Implementation Strategy

### Phase 1: Input Validation Tests (HIGH PRIORITY)

Focus on the command handler functions (`runTapCmd`, `runTextCmd`, `runSwipeCmd`, `runButtonCmd`) with comprehensive error path testing.

**Files to Update:**
- `cmd/io_test.go` - Add 40+ new test cases

### Phase 2: App Lifecycle Tests (MEDIUM PRIORITY)

Expand app command tests with timeout simulation and edge cases.

**Files to Update:**
- `cmd/app_test.go` - Add 25+ new test cases

### Phase 3: Screenshot Tests (HIGH PRIORITY)

Create comprehensive screenshot tests covering path handling and format validation.

**Files to Create:**
- `cmd/screenshot_test.go` - New file with 20+ test cases

### Phase 4: Simulator State Tests (MEDIUM PRIORITY)

Expand simulator tests with failure scenarios and state consistency.

**Files to Update:**
- `cmd/simulator_test.go` - Add 15+ new test cases

## Test Coverage Metrics

### Current Baseline
```
cmd package: 0% coverage
Total project: ~30% coverage (driven by pkg/ packages)
```

### Target After Expansion
```
cmd package: 60%+ coverage
io.go: 70%+ (command handlers and validation)
app.go: 65%+ (app operations and error paths)
screenshot.go: 75%+ (path handling and formats)
simulator.go: 70%+ (boot/shutdown flows)
Total project: 35%+ coverage
```

## Test Patterns Used

### 1. Table-Driven Tests
```go
tests := []struct {
    name        string
    input       interface{}
    expectError bool
    expectedErr string
}{
    // test cases
}
```

### 2. Mock/Assertion Pattern
```go
bridge := &MockXCRunBridge{}
bridge.On("Tap", "device-1", 100, 200).Return(&TapResult{}, nil)
// Execute and verify
```

### 3. Output Verification
```go
// Capture outputError/outputSuccess calls
// Verify JSON structure and values
```

## Detailed Test Cases by Component

### io.go - Tap Command

**Error Validation:**
- Device ID missing (should output DEVICE_REQUIRED)
- Negative X coordinate (should output INVALID_COORDINATES)
- Negative Y coordinate (should output INVALID_COORDINATES)
- Device not found (should output DEVICE_NOT_FOUND)
- Device not booted (should output DEVICE_NOT_BOOTED)
- Tap operation fails (should output UI_ACTION_FAILED)

**Happy Path:**
- Valid coordinates on booted device
- Zero coordinates (0, 0)
- Large coordinates (1920, 1080)

### io.go - Text Command

**Error Validation:**
- Device ID missing
- Empty text input
- Device not found
- Device not booted
- TypeText operation fails

**Happy Path:**
- Simple ASCII text
- Unicode text with emojis
- Text with special characters (quotes, newlines escaped)
- Long text (>1000 chars)

### io.go - Swipe Command

**Error Validation:**
- Missing start coordinates
- Missing end coordinates
- Negative coordinates
- Invalid duration (0 or negative)
- Device validation failures

**Happy Path:**
- Basic horizontal swipe
- Basic vertical swipe
- Diagonal swipe
- Custom duration
- Same start/end (tap via swipe)

### io.go - Button Command

**Error Validation:**
- Missing button type
- Invalid button type (WRONG_BUTTON)
- Device validation failures

**Happy Path:**
- HOME button press
- POWER button press
- VOLUME_UP press
- VOLUME_DOWN press

### app.go - Launch Command

**Error Validation:**
- Device not found
- Device not booted
- Launch operation fails

**Happy Path:**
- Basic launch
- Launch with wait-for-ready flag
- Launch with timeout
- Successful PID capture

### app.go - Terminate Command

**Error Validation:**
- Device not found
- Terminate operation fails

**Happy Path:**
- Terminate running app
- Terminate already-stopped app
- Verify success message

### app.go - Install Command

**Error Validation:**
- Device not found
- Invalid app path
- Install operation fails
- Extract bundle ID from path fails

**Happy Path:**
- Valid .app bundle installation
- Verify bundle ID extraction
- Verify installation time recording

### app.go - Uninstall Command

**Error Validation:**
- Device not found
- Bundle not installed
- Uninstall operation fails

**Happy Path:**
- Uninstall installed bundle
- Verify success message

### screenshot.go

**Error Validation:**
- Device not found
- Device not booted
- Invalid format (not png/jpeg)
- Output path creation fails (permission denied)
- Screenshot capture fails
- Directory creation fails

**Happy Path:**
- Default output path (timestamped /tmp)
- Custom output path
- PNG format (explicit)
- JPEG format
- Nested directory creation
- File size validation
- Timestamp format in filename

### simulator.go

**Error Validation:**
- No simulator with matching name
- Boot timeout
- Invalid name/OS version combination
- Shutdown invalid device
- Shutdown not-running simulator

**Happy Path:**
- Boot simulator (immediate state)
- Boot with polling (delayed state)
- Shutdown running simulator
- Name + OS version filtering

## JSON Contract Validation

All commands must produce valid JSON output with correct field types:

```go
// Test structure
assert.Equal(t, "io.tap", result.Command)
assert.NotNil(t, result.Success)
assert.True(t, *result.Success)
assert.NotEmpty(t, result.Data)
assert.NotNil(t, result.Timestamp)
```

## Test Execution

### Run All Tests
```bash
cd /Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli
make test
```

### Run Specific Test Suite
```bash
go test -v ./cmd -run TestIoTap
go test -v ./cmd -run TestAppLaunch
go test -v ./cmd -run TestScreenshot
```

### Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Success Criteria

- All error paths tested with expected error codes
- JSON output contract validated for each command
- Edge cases covered (empty input, boundary values, special characters)
- Mock setup demonstrates correct device/bridge interaction
- No flaky tests (deterministic, no time-dependent behavior except explicitly tested)
- Coverage increases from 0% to 65%+ for cmd package

## Timeline

- **Phase 1 (io.go):** 2-3 hours
- **Phase 2 (app.go):** 2 hours
- **Phase 3 (screenshot.go):** 2 hours
- **Phase 4 (simulator.go):** 1.5 hours
- **Total:** ~8-9 hours of test writing + execution

## Files Modified

1. `/Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli/cmd/io_test.go` (expand)
2. `/Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli/cmd/app_test.go` (expand)
3. `/Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli/cmd/screenshot_test.go` (create)
4. `/Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli/cmd/simulator_test.go` (expand)

## References

- Cobra Command Testing: https://github.com/spf13/cobra/wiki/Testing
- Table-Driven Tests: https://github.com/golang/wiki/wiki/TableDrivenTests
- Mock Testing: https://github.com/stretchr/testify
