# iOS Agent CLI - Implementation Review

**Review Date:** 2026-02-04  
**Reviewer:** Pi Agent  
**Project Status:** ~50% Complete (8/18 features done)

---

## Executive Summary

The ios-agent-cli is a well-architected Go CLI tool for AI-agent-driven iOS automation. The codebase demonstrates clean separation of concerns, idiomatic Go patterns, and thoughtful API design. However, there are opportunities to improve test coverage (currently 8.5%-93.8% across packages), standardize error handling, and streamline the remaining feature implementation.

**Overall Grade: B+**

| Dimension | Score | Notes |
|-----------|-------|-------|
| Code Quality | A- | Clean architecture, good idioms |
| Test Coverage | C+ | Device pkg excellent, xcrun/cmd need work |
| Feature Completeness | B- | Core P0s done, P1s pending |
| Documentation | A | Excellent README, strategy docs |
| Agent-Readiness | A | Consistent JSON output, error codes |

---

## 1. Code Quality Assessment

### 1.1 Strengths ✅

#### Clean Architecture (Layered Design)
```
cmd/           → CLI commands (Cobra)
pkg/device/    → Device manager abstraction
pkg/xcrun/     → xcrun simctl bridge
```

The separation between `cmd/` (CLI concerns) and `pkg/` (core logic) is excellent. This enables:
- Unit testing without CLI bootstrapping
- Potential reuse as a Go library
- Clear dependency direction (cmd → pkg)

#### Interface-Based Design
```go
// pkg/device/manager.go
type DeviceBridge interface {
    ListDevices() ([]Device, error)
    BootSimulator(udid string) error
    ShutdownSimulator(udid string) error
    GetDeviceState(udid string) (DeviceState, error)
}
```

The `DeviceBridge` interface is a key architectural win:
- Enables mocking for unit tests
- Future-proof for remote backend (mobilecli, Tailscale)
- Clean dependency injection

#### Consistent JSON Output
```go
// pkg/cmd/root.go
type Response struct {
    Success   bool        `json:"success"`
    Action    string      `json:"action,omitempty"`
    Result    interface{} `json:"result,omitempty"`
    Error     *ErrorInfo  `json:"error,omitempty"`
    Timestamp string      `json:"timestamp"`
}
```

Every command returns the same JSON envelope, making agent parsing predictable.

#### Idiomatic Go
- Short, focused functions
- Error wrapping with `%w`
- Table-driven tests
- Testify for assertions

### 1.2 Areas for Improvement 🔧

#### 1.2.1 Error Handling Inconsistency

**Issue:** Error codes are scattered across cmd files without centralization.

**Current:**
```go
// cmd/io.go
outputError("io.tap", "DEVICE_REQUIRED", "device ID is required", nil)

// cmd/simulator.go  
outputError("simulator.boot", "DEVICE_NOT_FOUND", err.Error(), nil)
```

**Recommended:** Create centralized error definitions:
```go
// pkg/errors/codes.go
package errors

type ErrorCode string

const (
    ErrDeviceNotFound     ErrorCode = "DEVICE_NOT_FOUND"
    ErrDeviceNotBooted    ErrorCode = "DEVICE_NOT_BOOTED"
    ErrDeviceUnreachable  ErrorCode = "DEVICE_UNREACHABLE"
    ErrAppNotFound        ErrorCode = "APP_NOT_FOUND"
    ErrUIActionFailed     ErrorCode = "UI_ACTION_FAILED"
    ErrSimulatorTimeout   ErrorCode = "SIMULATOR_TIMEOUT"
    ErrInvalidCoordinates ErrorCode = "INVALID_COORDINATES"
)

type AppError struct {
    Code    ErrorCode
    Message string
    Details map[string]string
}

func (e *AppError) Error() string { return string(e.Code) + ": " + e.Message }
```

**Priority:** P0 (IOS-015 in features.json)

#### 1.2.2 Bridge Abstraction Gap

**Issue:** `pkg/xcrun/bridge.go` has UI methods (`Tap`, `TypeText`) that don't belong in xcrun.

**Current:**
```go
// pkg/xcrun/bridge.go - This is simctl wrapper, but has:
func (b *Bridge) Tap(udid string, x, y int) (*TapResult, error) {
    // Uses AppleScript, not simctl!
}
```

**Recommended:** Extract UI interactions to separate package:
```
pkg/
├── xcrun/      # simctl-only operations
├── ui/         # UI interactions (AppleScript, mobilecli)
└── device/     # Device manager using both
```

#### 1.2.3 Global State in cmd/

**Issue:** Package-level variables in `cmd/` files create implicit state.

```go
// cmd/io.go
var (
    tapX int
    tapY int
    textInput string
)
```

**Impact:** Potential test pollution, harder to reason about.

**Recommended:** Use Cobra's local flag binding or command context:
```go
type tapCommand struct {
    x, y     int
    deviceID string
}

func newTapCommand() *cobra.Command {
    tc := &tapCommand{}
    cmd := &cobra.Command{
        Run: tc.run,
    }
    cmd.Flags().IntVar(&tc.x, "x", 0, "X coordinate")
    return cmd
}
```

#### 1.2.4 Missing Context Propagation

**Issue:** No `context.Context` for timeout/cancellation.

**Current:**
```go
func (b *Bridge) BootSimulator(udid string) error {
    cmd := exec.Command("xcrun", "simctl", "boot", udid)
    // No timeout!
}
```

**Recommended:**
```go
func (b *Bridge) BootSimulator(ctx context.Context, udid string) error {
    cmd := exec.CommandContext(ctx, "xcrun", "simctl", "boot", udid)
    // ...
}
```

---

## 2. Remaining Features Analysis

### 2.1 Feature Status Summary

| ID | Feature | Priority | Status | Complexity | Notes |
|----|---------|----------|--------|------------|-------|
| IOS-007 | App install | P1 | Pending | Low | `simctl install` wrapper |
| IOS-008 | App uninstall | P1 | Pending | Low | `simctl uninstall` wrapper |
| IOS-012 | Swipe command | P1 | Pending | Medium | AppleScript or mobilecli |
| IOS-013 | Button press | P1 | Pending | Medium | HOME/POWER buttons |
| IOS-014 | State command | P1 | Pending | Medium | Aggregate device info |
| IOS-015 | Error framework | P0 | Pending | Low | Centralize error codes |
| IOS-016 | Integration tests | P1 | Pending | High | Requires simulator |
| IOS-017 | Remote host | P2 | Pending | Medium | Flag plumbing |
| IOS-018 | Tailscale discovery | P2 | Pending | High | Network scanning |

### 2.2 Complexity Estimates

#### Low Complexity (1-2 hours each)

**IOS-007: App Install**
```go
// pkg/xcrun/bridge.go
func (b *Bridge) InstallApp(udid, ipaPath string) error {
    cmd := exec.Command("xcrun", "simctl", "install", udid, ipaPath)
    return cmd.Run()
}
```

**IOS-008: App Uninstall**
```go
func (b *Bridge) UninstallApp(udid, bundleID string) error {
    cmd := exec.Command("xcrun", "simctl", "uninstall", udid, bundleID)
    return cmd.Run()
}
```

**IOS-015: Error Framework**
- Create `pkg/errors/codes.go`
- Update all `outputError()` calls to use constants
- ~30 minutes refactoring

#### Medium Complexity (2-4 hours each)

**IOS-012: Swipe Command**

Options:
1. **AppleScript** (unreliable for swipe)
2. **mobilecli** (external dependency)
3. **Xcode simctl io** (limited gesture support)

Recommended approach:
```go
// Using simctl's io command with drag support
func (b *Bridge) Swipe(udid string, x1, y1, x2, y2 int, duration float64) error {
    // simctl doesn't have native swipe, options:
    // 1. Multiple rapid taps along path
    // 2. AppleScript drag gesture
    // 3. Require mobilecli for reliable swipe
}
```

**IOS-013: Button Press**
```go
// pkg/xcrun/bridge.go
func (b *Bridge) PressButton(udid, button string) error {
    // simctl supports: home, lock, siri
    switch button {
    case "HOME":
        return exec.Command("xcrun", "simctl", "io", udid, "home").Run()
    case "LOCK":
        return exec.Command("xcrun", "simctl", "io", udid, "lock").Run()
    default:
        return fmt.Errorf("unsupported button: %s", button)
    }
}
```

**IOS-014: State Command**
```go
type DeviceState struct {
    Device       Device       `json:"device"`
    ForegroundApp string      `json:"foreground_app,omitempty"`
    Battery      *BatteryInfo `json:"battery,omitempty"`
    Network      *NetworkInfo `json:"network,omitempty"`
    Screenshot   string       `json:"screenshot_base64,omitempty"`
}
```

#### High Complexity (4+ hours each)

**IOS-016: Integration Tests**
- Requires booting actual simulator
- CI/CD considerations (self-hosted runner with Xcode)
- Test isolation (dedicated test simulator)

**IOS-018: Tailscale Discovery**
- Parse Tailscale status for peers
- Probe for mobilecli server on standard port
- Handle authentication/authorization

---

## 3. Test Coverage Analysis

### 3.1 Current Coverage

| Package | Coverage | Assessment |
|---------|----------|------------|
| `pkg/device` | **93.8%** | Excellent ✅ |
| `cmd/` | **30.0%** | Needs improvement 🔧 |
| `pkg/xcrun` | **8.5%** | Critical gap ❌ |
| Root | **0.0%** | N/A (just main.go) |

### 3.2 Coverage Gaps

#### pkg/xcrun (8.5% → Target: 60%)

**Untested Functions:**
- `ListDevices()` - Requires simctl mock
- `BootSimulator()` - Integration only
- `ShutdownSimulator()` - Integration only
- `CaptureScreenshot()` - File system + simctl
- `Tap()` - AppleScript execution
- `TypeText()` - simctl keyboardinput
- `LaunchApp()` - simctl launch
- `TerminateApp()` - simctl terminate

**Recommended Test Strategy:**
```go
// pkg/xcrun/bridge_test.go

// 1. Create exec mock interface
type CommandRunner interface {
    Run(name string, args ...string) ([]byte, error)
}

// 2. Inject into Bridge
type Bridge struct {
    runner CommandRunner
}

// 3. Mock for testing
type mockRunner struct {
    outputs map[string][]byte
    errors  map[string]error
}

func TestListDevices_ParsesSimctlOutput(t *testing.T) {
    runner := &mockRunner{
        outputs: map[string][]byte{
            "xcrun simctl list devices --json": []byte(`{
                "devices": {
                    "com.apple.CoreSimulator.SimRuntime.iOS-17-4": [
                        {"udid": "ABC123", "name": "iPhone 15", "state": "Booted"}
                    ]
                }
            }`),
        },
    }
    bridge := NewBridgeWithRunner(runner)
    
    devices, err := bridge.ListDevices()
    
    assert.NoError(t, err)
    assert.Len(t, devices, 1)
    assert.Equal(t, "iPhone 15", devices[0].Name)
}
```

#### cmd/ (30% → Target: 70%)

**Untested Command Logic:**
- `runBootCmd` - Polling logic
- `runShutdownCmd` - Error paths
- `runLaunchCmd` - Device state validation
- `runScreenshotCmd` - Path handling
- `runTapCmd` - Coordinate validation
- `runTextCmd` - Empty text handling

**Recommended Approach:** Extract business logic to testable functions:
```go
// Instead of testing cobra commands directly,
// extract the logic:

// cmd/simulator_logic.go
func bootSimulator(manager *device.LocalManager, name, osVersion string, timeout int) (*BootResult, error) {
    // Testable without cobra
}

// cmd/simulator_logic_test.go
func TestBootSimulator_FindsDeviceByNameAndOS(t *testing.T) {
    // Test with mock manager
}
```

### 3.3 Integration Test Plan

```go
// test/integration/simulator_test.go
// +build integration

func TestSimulatorLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // 1. Find available simulator
    devices := listDevices(t)
    require.NotEmpty(t, devices)
    
    // 2. Boot simulator
    bootResult := bootSimulator(t, devices[0].Name)
    require.Equal(t, "Booted", bootResult.Device.State)
    
    // 3. Take screenshot
    screenshotResult := takeScreenshot(t, bootResult.Device.ID)
    require.FileExists(t, screenshotResult.Path)
    
    // 4. Shutdown
    shutdownResult := shutdownSimulator(t, bootResult.Device.ID)
    require.Equal(t, "Shutdown", shutdownResult.Device.State)
}
```

---

## 4. Recommendations

### 4.1 Immediate Actions (This Week)

| Priority | Action | Effort | Impact |
|----------|--------|--------|--------|
| P0 | Implement IOS-015 (Error Framework) | 2h | High - Consistency |
| P0 | Add xcrun bridge unit tests | 4h | High - Reliability |
| P1 | Implement IOS-007/008 (Install/Uninstall) | 2h | Medium - Feature |
| P1 | Extract cmd logic for testability | 3h | High - Quality |

### 4.2 Short-Term (Next 2 Weeks)

| Priority | Action | Effort | Impact |
|----------|--------|--------|--------|
| P1 | Implement IOS-012/013 (Swipe/Button) | 4h | Medium - Feature |
| P1 | Implement IOS-014 (State command) | 3h | High - Agent UX |
| P1 | Set up integration test CI | 4h | High - Reliability |
| P2 | Add context.Context support | 2h | Medium - Robustness |

### 4.3 Architecture Improvements

1. **Extract UI Package**
   ```
   pkg/ui/
   ├── interactions.go  # Tap, swipe, text
   ├── applescript.go   # AppleScript backend
   └── mobilecli.go     # mobilecli backend
   ```

2. **Add Configuration Layer**
   ```go
   // pkg/config/config.go
   type Config struct {
       DefaultTimeout  time.Duration
       MobileCLIPath   string
       RemoteHost      string
       VerboseLogging  bool
   }
   ```

3. **Improve Error Types**
   ```go
   // pkg/errors/errors.go
   type DeviceError struct {
       Code      ErrorCode
       DeviceID  string
       Operation string
       Cause     error
   }
   ```

---

## 5. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| AppleScript tap unreliable | High | Medium | Document limitation, recommend mobilecli |
| simctl API changes | Low | High | Pin Xcode version in CI |
| No swipe without mobilecli | High | Medium | Make mobilecli optional dependency |
| Remote feature complexity | Medium | Low | Defer to Phase 2 |

---

## 6. Conclusion

The ios-agent-cli is on track for a solid MVP. The core architecture is sound, and the remaining P0/P1 features are straightforward to implement. The main gaps are:

1. **Test coverage** in `pkg/xcrun` (critical path)
2. **Error handling** standardization
3. **UI interaction** reliability (swipe especially)

With focused effort on these areas, the project can reach production-ready quality within 2-3 sprints.

**Recommended Next Steps:**
1. ✅ Complete IOS-015 (Error Framework) - Foundational
2. ✅ Add xcrun mock tests - Confidence
3. ✅ Ship IOS-007/008 (Install/Uninstall) - Quick wins
4. 📋 Evaluate mobilecli integration for reliable UI interactions

---

*Review generated by Pi Agent for FORGE Portfolio*
