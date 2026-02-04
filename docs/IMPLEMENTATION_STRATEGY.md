# iOS Agent CLI - Implementation Strategy

**Document Version**: 1.0
**Date**: February 4, 2026
**Phase**: Phase 1 MVP
**Target Completion**: 2-3 weeks

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Package Design](#2-package-design)
3. [Implementation Order](#3-implementation-order)
4. [xcrun simctl Integration](#4-xcrun-simctl-integration)
5. [mobilecli Integration](#5-mobilecli-integration)
6. [Testing Strategy](#6-testing-strategy)
7. [Error Handling Framework](#7-error-handling-framework)
8. [Development Workflow](#8-development-workflow)
9. [Phase 1 Deliverables](#9-phase-1-deliverables)
10. [Risk Assessment](#10-risk-assessment)

---

## 1. Architecture Overview

### 1.1 Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│ Layer 1: CLI Commands (cmd/)                                     │
│ • Cobra command definitions                                      │
│ • Flag parsing and validation                                    │
│ • JSON output formatting                                         │
│ • Error code mapping                                             │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 2: Device Manager (pkg/device/)                            │
│ • Abstract DeviceManager interface                               │
│ • LocalSimulatorManager implementation                           │
│ • Device discovery and state management                          │
│ • Connection pooling (future: remote devices)                    │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 3: Backend Integration (pkg/xcrun/, pkg/mobilecli/)       │
│ • xcrun simctl wrapper (simulator lifecycle)                     │
│ • mobilecli HTTP client (UI interactions)                        │
│ • JSON parsing and command execution                             │
│ • Process management and timeouts                                │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Key Design Decisions

**Decision 1: Interface-Based Device Manager**
- **Why**: Enables mocking for tests, future remote device support
- **Trade-off**: Slight complexity increase vs. testability and extensibility
- **Implementation**: Single `DeviceManager` interface with `LocalSimulatorManager` impl

**Decision 2: JSON-First Output**
- **Why**: Agent-friendly, language-agnostic, structured
- **Trade-off**: Less human-readable vs. machine parsability
- **Implementation**: All commands use `Response` struct, always JSON

**Decision 3: Delegate to mobilecli for UI**
- **Why**: Avoid reimplementing XCUITest bindings, leverage existing tool
- **Trade-off**: External dependency vs. development speed
- **Implementation**: HTTP client wrapper, graceful fallback if not installed

**Decision 4: Direct xcrun simctl Calls**
- **Why**: Built into Xcode, no external deps, fast, reliable
- **Trade-off**: None (standard approach)
- **Implementation**: Execute binary, parse JSON output

**Decision 5: Synchronous Commands with Polling**
- **Why**: Simple, deterministic, agent-friendly
- **Trade-off**: Potential blocking vs. complexity of async
- **Implementation**: Use timeouts, poll for state changes

---

## 2. Package Design

### 2.1 pkg/device/manager.go

**Purpose**: Abstract interface for device management

```go
package device

import (
	"context"
	"time"
)

// Device represents an iOS device or simulator
type Device struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Platform    string `json:"platform"`    // "ios"
	Type        string `json:"type"`        // "simulator" | "real"
	State       string `json:"state"`       // "booted" | "shutdown" | "unreachable"
	OSVersion   string `json:"os_version"`
	Location    string `json:"location"`    // "local" | "remote"
	RemoteHost  string `json:"remote_host,omitempty"`
	RemotePort  int    `json:"remote_port,omitempty"`
}

// DeviceManager defines the interface for device operations
type DeviceManager interface {
	// Discovery
	ListDevices(ctx context.Context) ([]Device, error)
	GetDevice(ctx context.Context, deviceID string) (*Device, error)

	// Simulator Lifecycle
	BootSimulator(ctx context.Context, name string, osVersion string) (*Device, error)
	ShutdownSimulator(ctx context.Context, deviceID string) error

	// App Management
	LaunchApp(ctx context.Context, deviceID string, bundleID string, waitForReady time.Duration) (*AppLaunchResult, error)
	TerminateApp(ctx context.Context, deviceID string, bundleID string) error
	InstallApp(ctx context.Context, deviceID string, ipaPath string) (*AppInstallResult, error)
	UninstallApp(ctx context.Context, deviceID string, bundleID string) error

	// State
	GetDeviceState(ctx context.Context, deviceID string) (*DeviceState, error)
}

// AppLaunchResult contains app launch details
type AppLaunchResult struct {
	BundleID      string `json:"bundle_id"`
	PID           int    `json:"pid"`
	State         string `json:"state"`
	LaunchTimeMS  int64  `json:"launch_time_ms"`
	Ready         bool   `json:"ready"`
	ReadyTimeMS   int64  `json:"ready_time_ms,omitempty"`
}

// AppInstallResult contains app installation details
type AppInstallResult struct {
	BundleID       string `json:"bundle_id"`
	IPAPath        string `json:"ipa_path"`
	InstallTimeMS  int64  `json:"install_time_ms"`
}

// DeviceState represents the current state of a device
type DeviceState struct {
	DeviceID       string          `json:"device_id"`
	DeviceName     string          `json:"device_name"`
	Screen         *ScreenInfo     `json:"screen,omitempty"`
	ForegroundApp  string          `json:"foreground_app,omitempty"`
	AppState       string          `json:"app_state,omitempty"`
	Battery        int             `json:"battery,omitempty"`
	Connectivity   string          `json:"connectivity,omitempty"`
}

// ScreenInfo contains screen dimensions
type ScreenInfo struct {
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	ScreenshotPath string `json:"screenshot_path,omitempty"`
}
```

**Key Patterns**:
- Use `context.Context` for cancellation and timeouts
- Return pointers for complex results (allows nil on error)
- Use `time.Duration` for time-related parameters
- All public methods return `error` for consistent error handling

---

### 2.2 pkg/device/local.go

**Purpose**: LocalSimulatorManager implementation

```go
package device

import (
	"context"
	"fmt"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/neoforge-dev/ios-agent-cli/pkg/mobilecli"
)

// LocalSimulatorManager manages local iOS simulators
type LocalSimulatorManager struct {
	xcrunBridge  *xcrun.Bridge
	mobileCLI    *mobilecli.Client
	pollInterval time.Duration // For state polling (default: 500ms)
	maxRetries   int           // For transient failures (default: 2)
}

// NewLocalSimulatorManager creates a new local simulator manager
func NewLocalSimulatorManager() *LocalSimulatorManager {
	return &LocalSimulatorManager{
		xcrunBridge:  xcrun.NewBridge(),
		mobileCLI:    mobilecli.NewClient("http://localhost:4723"),
		pollInterval: 500 * time.Millisecond,
		maxRetries:   2,
	}
}

// ListDevices implements DeviceManager.ListDevices
func (m *LocalSimulatorManager) ListDevices(ctx context.Context) ([]Device, error) {
	// Delegate to xcrun bridge
	simctlDevices, err := m.xcrunBridge.ListSimulators(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list simulators: %w", err)
	}

	devices := make([]Device, 0, len(simctlDevices))
	for _, sim := range simctlDevices {
		devices = append(devices, Device{
			ID:        sim.UDID,
			Name:      sim.Name,
			Platform:  "ios",
			Type:      "simulator",
			State:     sim.State,
			OSVersion: sim.Runtime,
			Location:  "local",
		})
	}

	return devices, nil
}

// BootSimulator implements DeviceManager.BootSimulator
func (m *LocalSimulatorManager) BootSimulator(ctx context.Context, name string, osVersion string) (*Device, error) {
	startTime := time.Now()

	// 1. Find simulator by name and OS version
	devices, err := m.ListDevices(ctx)
	if err != nil {
		return nil, err
	}

	var targetDevice *Device
	for i := range devices {
		if devices[i].Name == name && (osVersion == "" || devices[i].OSVersion == osVersion) {
			targetDevice = &devices[i]
			break
		}
	}

	if targetDevice == nil {
		return nil, fmt.Errorf("simulator not found: name=%s, os_version=%s", name, osVersion)
	}

	// 2. If already booted, return immediately
	if targetDevice.State == "booted" {
		return targetDevice, nil
	}

	// 3. Boot simulator
	if err := m.xcrunBridge.BootSimulator(ctx, targetDevice.ID); err != nil {
		return nil, fmt.Errorf("failed to boot simulator: %w", err)
	}

	// 4. Poll until booted (timeout: 60s)
	pollCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			return nil, fmt.Errorf("simulator boot timeout after 60s")
		case <-ticker.C:
			device, err := m.GetDevice(ctx, targetDevice.ID)
			if err != nil {
				continue // Keep polling
			}
			if device.State == "booted" {
				bootTime := time.Since(startTime).Milliseconds()
				// Add boot time to result (extend Device struct or return separately)
				return device, nil
			}
		}
	}
}

// LaunchApp implements DeviceManager.LaunchApp
func (m *LocalSimulatorManager) LaunchApp(ctx context.Context, deviceID string, bundleID string, waitForReady time.Duration) (*AppLaunchResult, error) {
	startTime := time.Now()

	// 1. Verify device is booted
	device, err := m.GetDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if device.State != "booted" {
		return nil, fmt.Errorf("device not booted: state=%s", device.State)
	}

	// 2. Launch app via xcrun simctl
	pid, err := m.xcrunBridge.LaunchApp(ctx, deviceID, bundleID)
	if err != nil {
		return nil, fmt.Errorf("failed to launch app: %w", err)
	}

	result := &AppLaunchResult{
		BundleID:     bundleID,
		PID:          pid,
		State:        "running",
		LaunchTimeMS: time.Since(startTime).Milliseconds(),
		Ready:        false,
	}

	// 3. Optionally wait for app to be ready
	if waitForReady > 0 {
		time.Sleep(waitForReady) // Simple approach: fixed delay
		result.Ready = true
		result.ReadyTimeMS = time.Since(startTime).Milliseconds()
	}

	return result, nil
}

// Additional methods: TerminateApp, InstallApp, UninstallApp, GetDevice, GetDeviceState
// Follow similar patterns with proper error handling
```

**Key Patterns**:
- **Polling Pattern**: Use `time.Ticker` + `context.WithTimeout` for state changes
- **Error Wrapping**: Always wrap errors with context using `fmt.Errorf("%w", err)`
- **Retry Logic**: Implement in `executeWithRetry` helper method (not shown)
- **Validation**: Check device state before operations
- **Timing**: Track operation duration using `time.Since(startTime)`

---

### 2.3 pkg/xcrun/bridge.go

**Purpose**: Wrapper for xcrun simctl commands

```go
package xcrun

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Bridge wraps xcrun simctl commands
type Bridge struct {
	xcrunPath string
}

// NewBridge creates a new xcrun bridge
func NewBridge() *Bridge {
	return &Bridge{
		xcrunPath: "/usr/bin/xcrun", // Standard Xcode CLI tools path
	}
}

// SimctlDevice represents a simulator from simctl list
type SimctlDevice struct {
	UDID         string `json:"udid"`
	Name         string `json:"name"`
	State        string `json:"state"`
	IsAvailable  bool   `json:"isAvailable"`
	Runtime      string `json:"-"` // Parsed from runtime key
}

// SimctlListOutput is the JSON structure from 'xcrun simctl list --json devices'
type SimctlListOutput struct {
	Devices map[string][]SimctlDevice `json:"devices"`
}

// ListSimulators returns all available simulators
func (b *Bridge) ListSimulators(ctx context.Context) ([]SimctlDevice, error) {
	// Execute: xcrun simctl list --json devices
	cmd := exec.CommandContext(ctx, b.xcrunPath, "simctl", "list", "--json", "devices")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("simctl list failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute simctl list: %w", err)
	}

	// Parse JSON output
	var listOutput SimctlListOutput
	if err := json.Unmarshal(output, &listOutput); err != nil {
		return nil, fmt.Errorf("failed to parse simctl output: %w", err)
	}

	// Flatten devices map into slice
	var devices []SimctlDevice
	for runtime, devs := range listOutput.Devices {
		for i := range devs {
			devs[i].Runtime = parseRuntime(runtime)
			if devs[i].IsAvailable {
				devices = append(devices, devs[i])
			}
		}
	}

	return devices, nil
}

// BootSimulator boots a simulator by UDID
func (b *Bridge) BootSimulator(ctx context.Context, udid string) error {
	cmd := exec.CommandContext(ctx, b.xcrunPath, "simctl", "boot", udid)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already booted (not an error)
		if strings.Contains(string(output), "current state: Booted") {
			return nil // Already booted
		}
		return fmt.Errorf("failed to boot simulator %s: %s", udid, string(output))
	}
	return nil
}

// ShutdownSimulator shuts down a simulator by UDID
func (b *Bridge) ShutdownSimulator(ctx context.Context, udid string) error {
	cmd := exec.CommandContext(ctx, b.xcrunPath, "simctl", "shutdown", udid)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already shutdown (not an error)
		if strings.Contains(string(output), "current state: Shutdown") {
			return nil
		}
		return fmt.Errorf("failed to shutdown simulator %s: %s", udid, string(output))
	}
	return nil
}

// LaunchApp launches an app by bundle ID and returns PID
func (b *Bridge) LaunchApp(ctx context.Context, udid string, bundleID string) (int, error) {
	cmd := exec.CommandContext(ctx, b.xcrunPath, "simctl", "launch", udid, bundleID)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return 0, fmt.Errorf("failed to launch app: %s", string(exitErr.Stderr))
		}
		return 0, fmt.Errorf("failed to execute simctl launch: %w", err)
	}

	// Parse PID from output: "com.example.app: 12345"
	pid, err := parsePID(string(output))
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID from output: %w", err)
	}

	return pid, nil
}

// TerminateApp terminates an app by bundle ID
func (b *Bridge) TerminateApp(ctx context.Context, udid string, bundleID string) error {
	cmd := exec.CommandContext(ctx, b.xcrunPath, "simctl", "terminate", udid, bundleID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to terminate app %s: %s", bundleID, string(output))
	}
	return nil
}

// InstallApp installs an IPA/app bundle on simulator
func (b *Bridge) InstallApp(ctx context.Context, udid string, appPath string) error {
	cmd := exec.CommandContext(ctx, b.xcrunPath, "simctl", "install", udid, appPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install app %s: %s", appPath, string(output))
	}
	return nil
}

// UninstallApp uninstalls an app by bundle ID
func (b *Bridge) UninstallApp(ctx context.Context, udid string, bundleID string) error {
	cmd := exec.CommandContext(ctx, b.xcrunPath, "simctl", "uninstall", udid, bundleID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall app %s: %s", bundleID, string(output))
	}
	return nil
}

// TakeScreenshot captures a screenshot and saves to path
func (b *Bridge) TakeScreenshot(ctx context.Context, udid string, outputPath string) error {
	cmd := exec.CommandContext(ctx, b.xcrunPath, "simctl", "io", udid, "screenshot", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %s", string(output))
	}
	return nil
}

// Helper functions

// parseRuntime extracts iOS version from runtime string
// Example: "com.apple.CoreSimulator.SimRuntime.iOS-17-4" -> "17.4"
func parseRuntime(runtime string) string {
	parts := strings.Split(runtime, ".")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		version := strings.ReplaceAll(strings.TrimPrefix(last, "iOS-"), "-", ".")
		return version
	}
	return runtime
}

// parsePID extracts PID from simctl launch output
// Example: "com.example.app: 12345" -> 12345
func parsePID(output string) (int, error) {
	parts := strings.Split(strings.TrimSpace(output), ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("unexpected output format: %s", output)
	}
	var pid int
	_, err := fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &pid)
	return pid, err
}
```

**Key Patterns**:
- **Command Execution**: Use `exec.CommandContext` for cancellation support
- **Error Handling**: Check `ExitError` for stderr output
- **JSON Parsing**: Use `encoding/json` for structured output
- **String Parsing**: Use `strings` package for simple output parsing
- **Idempotency**: Check if operation already done (e.g., "already booted")

---

### 2.4 pkg/output/formatter.go

**Purpose**: Standardized JSON output and error formatting

```go
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Response is the standard JSON response wrapper
type Response struct {
	Success   bool        `json:"success"`
	Action    string      `json:"action,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	ErrDeviceNotFound      ErrorCode = "DEVICE_NOT_FOUND"
	ErrDeviceUnreachable   ErrorCode = "DEVICE_UNREACHABLE"
	ErrAppNotFound         ErrorCode = "APP_NOT_FOUND"
	ErrAppCrash            ErrorCode = "APP_CRASH"
	ErrUIActionFailed      ErrorCode = "UI_ACTION_FAILED"
	ErrSimulatorTimeout    ErrorCode = "SIMULATOR_TIMEOUT"
	ErrInvalidArgs         ErrorCode = "INVALID_ARGS"
	ErrInternalError       ErrorCode = "INTERNAL_ERROR"
	ErrMobileCLINotFound   ErrorCode = "MOBILECLI_NOT_FOUND"
)

// OutputSuccess prints a successful response as JSON
func OutputSuccess(action string, result interface{}) {
	resp := Response{
		Success:   true,
		Action:    action,
		Result:    result,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	printJSON(resp)
}

// OutputError prints an error response as JSON and exits with code 1
func OutputError(action string, code ErrorCode, message string, details interface{}) {
	resp := Response{
		Success:   false,
		Action:    action,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	printJSON(resp)
	os.Exit(1)
}

// OutputErrorWithExit prints error and exits with custom exit code
func OutputErrorWithExit(action string, code ErrorCode, message string, details interface{}, exitCode int) {
	resp := Response{
		Success:   false,
		Action:    action,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	printJSON(resp)
	os.Exit(exitCode)
}

// printJSON prints response as formatted JSON to stdout
func printJSON(resp Response) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

// WrapError converts a Go error to an ErrorCode and message
func WrapError(err error) (ErrorCode, string) {
	if err == nil {
		return "", ""
	}

	msg := err.Error()

	// Pattern matching for common errors
	switch {
	case contains(msg, "not found"):
		return ErrDeviceNotFound, msg
	case contains(msg, "unreachable"), contains(msg, "connection refused"):
		return ErrDeviceUnreachable, msg
	case contains(msg, "timeout"):
		return ErrSimulatorTimeout, msg
	case contains(msg, "app not found"), contains(msg, "bundle not found"):
		return ErrAppNotFound, msg
	case contains(msg, "invalid argument"), contains(msg, "required flag"):
		return ErrInvalidArgs, msg
	default:
		return ErrInternalError, msg
	}
}

// contains checks if string s contains substring substr (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && indexContains(s, substr) >= 0))
}

func indexContains(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
```

**Key Patterns**:
- **Consistent Structure**: All output uses `Response` wrapper
- **Typed Error Codes**: Use `ErrorCode` type (string enum)
- **Error Mapping**: `WrapError` converts Go errors to error codes
- **Timestamp**: Always include RFC3339 timestamp
- **Exit Codes**: Use `os.Exit(1)` for error responses

---

### 2.5 pkg/mobilecli/client.go

**Purpose**: HTTP client wrapper for mobilecli server

```go
package mobilecli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// Client wraps HTTP interactions with mobilecli server
type Client struct {
	baseURL    string
	httpClient *http.Client
	available  bool
}

// NewClient creates a new mobilecli client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		available: false, // Will check on first use
	}
}

// IsAvailable checks if mobilecli is installed and running
func (c *Client) IsAvailable(ctx context.Context) bool {
	// Check if already verified
	if c.available {
		return true
	}

	// Try to ping health endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if mobilecli binary exists
		if _, err := exec.LookPath("mobilecli"); err == nil {
			// Binary exists but server not running
			fmt.Fprintf(os.Stderr, "Warning: mobilecli found but server not running. Start with: mobilecli server --listen 0.0.0.0:4723\n")
		}
		return false
	}
	defer resp.Body.Close()

	c.available = resp.StatusCode == 200
	return c.available
}

// TapRequest represents a tap action request
type TapRequest struct {
	DeviceID string `json:"device_id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

// TapResponse represents a tap action response
type TapResponse struct {
	Success    bool   `json:"success"`
	TapTimeMS  int64  `json:"tap_time_ms"`
	Error      string `json:"error,omitempty"`
}

// Tap performs a tap at x,y coordinates
func (c *Client) Tap(ctx context.Context, deviceID string, x, y int) (*TapResponse, error) {
	if !c.IsAvailable(ctx) {
		return nil, fmt.Errorf("mobilecli not available")
	}

	req := TapRequest{
		DeviceID: deviceID,
		X:        x,
		Y:        y,
	}

	var resp TapResponse
	if err := c.doRequest(ctx, "POST", "/io/tap", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// TypeTextRequest represents a text input request
type TypeTextRequest struct {
	DeviceID string `json:"device_id"`
	Text     string `json:"text"`
}

// TypeTextResponse represents a text input response
type TypeTextResponse struct {
	Success     bool   `json:"success"`
	TextTimeMS  int64  `json:"text_time_ms"`
	Error       string `json:"error,omitempty"`
}

// TypeText types text into the focused field
func (c *Client) TypeText(ctx context.Context, deviceID string, text string) (*TypeTextResponse, error) {
	if !c.IsAvailable(ctx) {
		return nil, fmt.Errorf("mobilecli not available")
	}

	req := TypeTextRequest{
		DeviceID: deviceID,
		Text:     text,
	}

	var resp TypeTextResponse
	if err := c.doRequest(ctx, "POST", "/io/text", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ScreenshotRequest represents a screenshot request
type ScreenshotRequest struct {
	DeviceID   string `json:"device_id"`
	OutputPath string `json:"output_path"`
	Format     string `json:"format"` // "png" or "jpeg"
}

// ScreenshotResponse represents a screenshot response
type ScreenshotResponse struct {
	Success       bool   `json:"success"`
	Path          string `json:"path"`
	SizeBytes     int64  `json:"size_bytes"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	ScreenshotMS  int64  `json:"screenshot_ms"`
	Error         string `json:"error,omitempty"`
}

// Screenshot captures a screenshot of the device
func (c *Client) Screenshot(ctx context.Context, deviceID string, outputPath string, format string) (*ScreenshotResponse, error) {
	if !c.IsAvailable(ctx) {
		return nil, fmt.Errorf("mobilecli not available")
	}

	req := ScreenshotRequest{
		DeviceID:   deviceID,
		OutputPath: outputPath,
		Format:     format,
	}

	var resp ScreenshotResponse
	if err := c.doRequest(ctx, "POST", "/screenshot", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// doRequest performs an HTTP request with JSON marshaling
func (c *Client) doRequest(ctx context.Context, method, path string, reqBody, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("mobilecli request timeout: %w", err)
		}
		return fmt.Errorf("mobilecli request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mobilecli returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
```

**Key Patterns**:
- **Availability Check**: Verify mobilecli is running before requests
- **Graceful Degradation**: Return helpful error messages if not available
- **HTTP Client**: Reuse client with timeout
- **JSON Marshaling**: Type-safe request/response structs
- **Error Context**: Wrap errors with context about what failed

**Fallback Strategy**:
If mobilecli not available:
1. For screenshots: Fall back to `xcrun simctl io screenshot`
2. For UI interactions: Return error with installation instructions
3. Log warning to stderr (not JSON stdout)

---

## 3. Implementation Order

### 3.1 Dependency Graph

```
Foundation (Week 1)
├── IOS-001: Project scaffold ✅ DONE
├── IOS-015: Error handling framework (2h)
│   └── Blocks: ALL commands
└── IOS-002: Device discovery (2h)
    └── Blocks: IOS-003, IOS-005

Simulator Lifecycle (Week 1)
├── IOS-003: Simulator boot (2h)
│   ├── Depends: IOS-002
│   └── Blocks: IOS-004, IOS-005, IOS-009
└── IOS-004: Simulator shutdown (1h)
    └── Depends: IOS-003

App Management (Week 2)
├── IOS-005: App launch (2h)
│   ├── Depends: IOS-003
│   └── Blocks: IOS-006, IOS-008
├── IOS-006: App terminate (1h)
│   └── Depends: IOS-005
├── IOS-007: App install (2h) [P1]
│   └── Depends: IOS-003
└── IOS-008: App uninstall (1h) [P1]
    └── Depends: IOS-005

Observation (Week 2)
├── IOS-009: Screenshot command (2h)
│   ├── Depends: IOS-003
│   └── Blocks: IOS-014, IOS-016
└── IOS-014: State command (3h) [P1]
    └── Depends: IOS-009

UI Interactions (Week 2)
├── IOS-010: Tap interaction (2h)
│   ├── Depends: IOS-003
│   └── Blocks: IOS-011, IOS-012, IOS-013, IOS-016
├── IOS-011: Text input (1h)
│   └── Depends: IOS-010
├── IOS-012: Swipe interaction (2h) [P1]
│   └── Depends: IOS-010
└── IOS-013: Button press (1h) [P1]
    └── Depends: IOS-010

Quality (Week 3)
└── IOS-016: Integration tests (4h)
    └── Depends: IOS-009, IOS-010
```

### 3.2 Sprint Breakdown

**Week 1: Foundation + Simulator Lifecycle**

Day 1-2:
- ✅ IOS-001: Project scaffold (DONE)
- IOS-015: Error handling framework (2h)
  - Implement `pkg/output/formatter.go`
  - Add error code constants
  - Add `WrapError` helper
  - Test JSON output format

Day 3:
- IOS-002: Device discovery (2h)
  - Implement `pkg/xcrun/bridge.go` (ListSimulators)
  - Implement `pkg/device/manager.go` (interface)
  - Implement `pkg/device/local.go` (ListDevices)
  - Add `cmd/devices.go` command
  - Test: `ios-agent devices`

Day 4:
- IOS-003: Simulator boot (2h)
  - Implement `pkg/xcrun/bridge.go` (BootSimulator)
  - Implement `pkg/device/local.go` (BootSimulator with polling)
  - Add `cmd/simulator/boot.go` command
  - Test: Boot iPhone 15 simulator

Day 5:
- IOS-004: Simulator shutdown (1h)
  - Implement `pkg/xcrun/bridge.go` (ShutdownSimulator)
  - Implement `pkg/device/local.go` (ShutdownSimulator)
  - Add `cmd/simulator/shutdown.go` command
  - Test: Shutdown booted simulator

**Week 2: App Management + UI Interactions**

Day 6-7:
- IOS-005: App launch (2h)
  - Implement `pkg/xcrun/bridge.go` (LaunchApp)
  - Implement `pkg/device/local.go` (LaunchApp)
  - Add `cmd/app/launch.go` command
  - Test: Launch Safari/Settings app

Day 8:
- IOS-006: App terminate (1h)
  - Implement xcrun terminate
  - Add `cmd/app/terminate.go` command
- IOS-009: Screenshot command (2h)
  - Implement `pkg/xcrun/bridge.go` (TakeScreenshot)
  - Add `cmd/screenshot.go` command
  - Test: Screenshot booted simulator

Day 9:
- IOS-010: Tap interaction (2h)
  - Implement `pkg/mobilecli/client.go` (basic structure)
  - Implement Tap method
  - Add `cmd/io/tap.go` command
  - Add fallback if mobilecli not available

Day 10:
- IOS-011: Text input (1h)
  - Implement TypeText in mobilecli client
  - Add `cmd/io/text.go` command
- IOS-007: App install (2h) [If time permits]
  - Implement install/uninstall commands

**Week 3: Polish + Testing**

Day 11-12:
- IOS-016: Integration tests (4h)
  - Write test suite in `test/integration_test.go`
  - Test full workflow: boot → launch → screenshot → tap → terminate
  - Add test fixtures
  - Add mocks for DeviceManager

Day 13:
- IOS-014: State command (3h) [If time permits]
  - Aggregate device state from multiple sources
  - Add UI element detection (future)
- IOS-012: Swipe interaction (2h) [If time permits]
- IOS-013: Button press (1h) [If time permits]

Day 14-15:
- Documentation
  - Update README with examples
  - Write ARCHITECTURE.md
  - Write EXAMPLES.md with agent integration examples
  - Add inline code comments
- Bug fixes and polish
- Performance testing

### 3.3 Estimated Effort Summary

| Epic | Features | Total Effort |
|------|----------|--------------|
| Foundation | IOS-001, IOS-015 | 3h (1h done) |
| Device Management | IOS-002, IOS-003, IOS-004 | 5h |
| App Management | IOS-005, IOS-006, IOS-007, IOS-008 | 6h |
| Observation | IOS-009, IOS-014 | 5h |
| UI Interactions | IOS-010, IOS-011, IOS-012, IOS-013 | 6h |
| Quality | IOS-016 | 4h |
| **TOTAL** | **18 features** | **29h (MVP: 22h P0 features)** |

**MVP Timeline**: 2-3 weeks (assuming 8-10h/day focus time)

---

## 4. xcrun simctl Integration

### 4.1 Key simctl Commands

| Command | Purpose | Output Format | Error Handling |
|---------|---------|---------------|----------------|
| `xcrun simctl list --json devices` | List simulators | JSON | Parse JSON, filter by availability |
| `xcrun simctl boot <udid>` | Boot simulator | Text | Check "already booted" |
| `xcrun simctl shutdown <udid>` | Shutdown simulator | Text | Check "already shutdown" |
| `xcrun simctl launch <udid> <bundle>` | Launch app | Text (PID) | Parse PID from output |
| `xcrun simctl terminate <udid> <bundle>` | Terminate app | Text | Direct success/failure |
| `xcrun simctl install <udid> <path>` | Install app | Text | Direct success/failure |
| `xcrun simctl uninstall <udid> <bundle>` | Uninstall app | Text | Direct success/failure |
| `xcrun simctl io <udid> screenshot <path>` | Take screenshot | Text | Check file created |

### 4.2 JSON Parsing Approach

**simctl list output structure**:

```json
{
  "devices": {
    "com.apple.CoreSimulator.SimRuntime.iOS-17-4": [
      {
        "udid": "12345678-1234-5678-1234-567890ABCDEF",
        "name": "iPhone 15 Pro",
        "state": "Booted",
        "isAvailable": true,
        "deviceTypeIdentifier": "com.apple.CoreSimulator.SimDeviceType.iPhone-15-Pro"
      }
    ]
  }
}
```

**Parsing Strategy**:
1. Unmarshal into `SimctlListOutput` struct
2. Iterate over `Devices` map (key = runtime)
3. Filter by `isAvailable == true`
4. Extract OS version from runtime key
5. Flatten into `[]SimctlDevice`

**Error Cases**:
- JSON unmarshal failure → `ErrInternalError`
- Empty devices list → Return empty slice (not error)
- Invalid runtime format → Use full string as version

### 4.3 Error Handling for simctl Failures

**Common Errors**:

| Error Output | Mapping | User Message |
|--------------|---------|--------------|
| `Device not found` | `ErrDeviceNotFound` | "Simulator with ID X not found. Run 'ios-agent devices' to list available simulators." |
| `Unable to boot device in current state: Booted` | Not an error | Return success (idempotent) |
| `Unable to boot device in current state: Booting` | Retry logic | Wait and retry up to 3 times |
| `Unable to shutdown device in current state: Shutdown` | Not an error | Return success (idempotent) |
| `No such file or directory` (install) | `ErrInvalidArgs` | "IPA file not found: <path>" |
| `The application bundle X could not be found` | `ErrAppNotFound` | "App not installed. Install with 'ios-agent app install'." |
| `Operation timed out` | `ErrSimulatorTimeout` | "Simulator operation timed out after 60s" |

**Retry Strategy**:
- Boot/Shutdown: No retry (use polling instead)
- Launch/Terminate: 1 retry after 1s delay
- Install/Uninstall: No retry (deterministic)

**Timeout Strategy**:
- All commands: 30s default timeout (via `context.WithTimeout`)
- Boot polling: 60s total timeout
- Screenshot: 10s timeout (fast operation)

---

## 5. mobilecli Integration

### 5.1 Detection Strategy

**Approach 1: Binary Check**
```go
func isMobileCLIInstalled() bool {
    _, err := exec.LookPath("mobilecli")
    return err == nil
}
```

**Approach 2: Server Health Check**
```go
func isMobileCLIRunning(baseURL string) bool {
    resp, err := http.Get(baseURL + "/health")
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}
```

**Combined Strategy** (Recommended):
1. Check if server is running (health endpoint)
2. If not, check if binary is installed
3. If binary exists but server not running, log warning with instructions

**Warning Message**:
```
Warning: mobilecli found but server not running.
Start with: mobilecli server --listen 0.0.0.0:4723
```

### 5.2 HTTP Client Wrapper Design

**Client Structure**:
```go
type Client struct {
    baseURL    string           // e.g., "http://localhost:4723"
    httpClient *http.Client     // Reusable HTTP client
    available  bool              // Cached availability status
}
```

**Key Methods**:
- `IsAvailable(ctx) bool` - Check if mobilecli is running
- `Tap(ctx, deviceID, x, y) error` - Tap at coordinates
- `TypeText(ctx, deviceID, text) error` - Type text
- `Swipe(ctx, deviceID, startX, startY, endX, endY, duration) error` - Swipe gesture
- `Screenshot(ctx, deviceID, outputPath, format) error` - Take screenshot
- `PressButton(ctx, deviceID, button) error` - Press hardware button

**Request/Response Pattern**:
```go
type TapRequest struct {
    DeviceID string `json:"device_id"`
    X        int    `json:"x"`
    Y        int    `json:"y"`
}

type TapResponse struct {
    Success   bool   `json:"success"`
    TapTimeMS int64  `json:"tap_time_ms"`
    Error     string `json:"error,omitempty"`
}

func (c *Client) Tap(ctx context.Context, deviceID string, x, y int) (*TapResponse, error) {
    req := TapRequest{DeviceID: deviceID, X: x, Y: y}
    var resp TapResponse
    err := c.doRequest(ctx, "POST", "/io/tap", req, &resp)
    return &resp, err
}
```

### 5.3 Fallback Strategy

**If mobilecli Not Available**:

| Command | Fallback | Implementation |
|---------|----------|----------------|
| `screenshot` | `xcrun simctl io screenshot` | Use xcrun bridge |
| `tap` | Error + instructions | Return `ErrMobileCLINotFound` |
| `text` | Error + instructions | Return `ErrMobileCLINotFound` |
| `swipe` | Error + instructions | Return `ErrMobileCLINotFound` |
| `button` | Error + instructions | Return `ErrMobileCLINotFound` |

**Error Response Example**:
```json
{
  "success": false,
  "action": "ui_tap",
  "error": {
    "code": "MOBILECLI_NOT_FOUND",
    "message": "mobilecli is required for UI interactions but not available",
    "details": {
      "installation_url": "https://github.com/mobile-next/mobilecli",
      "install_command": "brew install mobilecli",
      "server_command": "mobilecli server --listen 0.0.0.0:4723"
    }
  },
  "timestamp": "2026-02-04T18:53:00Z"
}
```

**Implementation Pattern**:
```go
func (m *LocalSimulatorManager) PerformTap(ctx context.Context, deviceID string, x, y int) error {
    if !m.mobileCLI.IsAvailable(ctx) {
        return &MobileCLINotAvailableError{
            Message: "mobilecli is required for UI interactions",
            InstallURL: "https://github.com/mobile-next/mobilecli",
        }
    }

    resp, err := m.mobileCLI.Tap(ctx, deviceID, x, y)
    if err != nil {
        return fmt.Errorf("tap failed: %w", err)
    }
    if !resp.Success {
        return fmt.Errorf("tap failed: %s", resp.Error)
    }
    return nil
}
```

---

## 6. Testing Strategy

### 6.1 Unit Test Structure

**File**: `pkg/device/manager_test.go`

```go
package device

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockXCRunBridge mocks the xcrun bridge
type MockXCRunBridge struct {
    mock.Mock
}

func (m *MockXCRunBridge) ListSimulators(ctx context.Context) ([]SimctlDevice, error) {
    args := m.Called(ctx)
    return args.Get(0).([]SimctlDevice), args.Error(1)
}

func (m *MockXCRunBridge) BootSimulator(ctx context.Context, udid string) error {
    args := m.Called(ctx, udid)
    return args.Error(0)
}

// Test: List devices returns correct format
func TestListDevices(t *testing.T) {
    mockBridge := new(MockXCRunBridge)
    mockBridge.On("ListSimulators", mock.Anything).Return([]SimctlDevice{
        {UDID: "test-uuid", Name: "iPhone 15", State: "Booted", Runtime: "17.4"},
    }, nil)

    manager := &LocalSimulatorManager{
        xcrunBridge: mockBridge,
        pollInterval: 100 * time.Millisecond,
    }

    devices, err := manager.ListDevices(context.Background())

    assert.NoError(t, err)
    assert.Len(t, devices, 1)
    assert.Equal(t, "iPhone 15", devices[0].Name)
    assert.Equal(t, "local", devices[0].Location)

    mockBridge.AssertExpectations(t)
}

// Test: Boot simulator with timeout
func TestBootSimulatorTimeout(t *testing.T) {
    mockBridge := new(MockXCRunBridge)
    mockBridge.On("BootSimulator", mock.Anything, "test-uuid").Return(nil)
    mockBridge.On("ListSimulators", mock.Anything).Return([]SimctlDevice{
        {UDID: "test-uuid", State: "Booting"}, // Never becomes "Booted"
    }, nil)

    manager := &LocalSimulatorManager{
        xcrunBridge: mockBridge,
        pollInterval: 100 * time.Millisecond,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
    defer cancel()

    _, err := manager.BootSimulator(ctx, "iPhone 15", "17.4")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}
```

**File**: `pkg/output/formatter_test.go`

```go
package output

import (
    "bytes"
    "encoding/json"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestOutputSuccess(t *testing.T) {
    // Capture stdout
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    OutputSuccess("test_action", map[string]string{"key": "value"})

    w.Close()
    os.Stdout = oldStdout

    var buf bytes.Buffer
    buf.ReadFrom(r)

    var resp Response
    err := json.Unmarshal(buf.Bytes(), &resp)

    assert.NoError(t, err)
    assert.True(t, resp.Success)
    assert.Equal(t, "test_action", resp.Action)
}

func TestWrapError(t *testing.T) {
    tests := []struct {
        name     string
        err      error
        wantCode ErrorCode
    }{
        {"device not found", fmt.Errorf("device not found"), ErrDeviceNotFound},
        {"timeout", fmt.Errorf("operation timeout"), ErrSimulatorTimeout},
        {"generic", fmt.Errorf("unknown error"), ErrInternalError},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            code, _ := WrapError(tt.err)
            assert.Equal(t, tt.wantCode, code)
        })
    }
}
```

### 6.2 Mock Strategy

**DeviceManager Mock** (`test/mocks/device_manager.go`):
```go
package mocks

import (
    "context"
    "time"

    "github.com/neoforge-dev/ios-agent-cli/pkg/device"
    "github.com/stretchr/testify/mock"
)

type MockDeviceManager struct {
    mock.Mock
}

func (m *MockDeviceManager) ListDevices(ctx context.Context) ([]device.Device, error) {
    args := m.Called(ctx)
    return args.Get(0).([]device.Device), args.Error(1)
}

func (m *MockDeviceManager) GetDevice(ctx context.Context, deviceID string) (*device.Device, error) {
    args := m.Called(ctx, deviceID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*device.Device), args.Error(1)
}

func (m *MockDeviceManager) BootSimulator(ctx context.Context, name string, osVersion string) (*device.Device, error) {
    args := m.Called(ctx, name, osVersion)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*device.Device), args.Error(1)
}

// ... implement other interface methods
```

**Usage in Command Tests** (`cmd/devices_test.go`):
```go
package cmd

import (
    "testing"

    "github.com/neoforge-dev/ios-agent-cli/pkg/device"
    "github.com/neoforge-dev/ios-agent-cli/test/mocks"
    "github.com/stretchr/testify/assert"
)

func TestDevicesCommand(t *testing.T) {
    mockManager := new(mocks.MockDeviceManager)
    mockManager.On("ListDevices", mock.Anything).Return([]device.Device{
        {ID: "test-1", Name: "iPhone 15", State: "booted"},
    }, nil)

    // Inject mock into command (requires refactoring cmd to accept manager)
    // Execute command
    // Assert JSON output
}
```

### 6.3 Integration Test Requirements

**File**: `test/integration_test.go`

```go
//go:build integration
// +build integration

package test

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/neoforge-dev/ios-agent-cli/pkg/device"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestFullWorkflow tests complete agent workflow
func TestFullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    ctx := context.Background()
    manager := device.NewLocalSimulatorManager()

    // 1. List devices
    devices, err := manager.ListDevices(ctx)
    require.NoError(t, err)
    require.NotEmpty(t, devices)

    // 2. Boot simulator (if not booted)
    var bootedDevice *device.Device
    for i := range devices {
        if devices[i].State == "booted" {
            bootedDevice = &devices[i]
            break
        }
    }

    if bootedDevice == nil {
        // Boot first available simulator
        bootedDevice, err = manager.BootSimulator(ctx, devices[0].Name, "")
        require.NoError(t, err)
        require.Equal(t, "booted", bootedDevice.State)
    }

    // 3. Launch app (Safari)
    result, err := manager.LaunchApp(ctx, bootedDevice.ID, "com.apple.mobilesafari", 2*time.Second)
    require.NoError(t, err)
    assert.Greater(t, result.PID, 0)

    // 4. Take screenshot
    screenshotPath := "/tmp/ios-agent-test.png"
    // (Implement screenshot method)

    // 5. Terminate app
    err = manager.TerminateApp(ctx, bootedDevice.ID, "com.apple.mobilesafari")
    require.NoError(t, err)

    // 6. Cleanup
    os.Remove(screenshotPath)
}

// TestMobileCLIIntegration tests UI interactions with mobilecli
func TestMobileCLIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Check if mobilecli is available
    // If not, skip test

    // Boot simulator
    // Launch app
    // Perform tap
    // Verify result
}
```

**Running Integration Tests**:
```bash
# Run only unit tests (fast)
go test ./... -short

# Run all tests including integration (slow)
go test ./... -tags integration

# Run specific integration test
go test ./test -tags integration -run TestFullWorkflow -v
```

**Test Fixtures** (`test/fixtures/`):
- `sample_device_list.json` - Mock simctl output
- `sample_screenshot.png` - Test screenshot for validation

**Coverage Target**: 70% minimum
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

---

## 7. Error Handling Framework

### 7.1 Error Code Definitions

**File**: `pkg/output/errors.go`

```go
package output

// ErrorCode represents standardized error codes
type ErrorCode string

const (
    // Device Errors
    ErrDeviceNotFound      ErrorCode = "DEVICE_NOT_FOUND"
    ErrDeviceUnreachable   ErrorCode = "DEVICE_UNREACHABLE"
    ErrSimulatorTimeout    ErrorCode = "SIMULATOR_TIMEOUT"

    // App Errors
    ErrAppNotFound         ErrorCode = "APP_NOT_FOUND"
    ErrAppCrash            ErrorCode = "APP_CRASH"
    ErrAppInstallFailed    ErrorCode = "APP_INSTALL_FAILED"

    // UI Errors
    ErrUIActionFailed      ErrorCode = "UI_ACTION_FAILED"
    ErrCoordinatesInvalid  ErrorCode = "COORDINATES_INVALID"

    // Dependency Errors
    ErrMobileCLINotFound   ErrorCode = "MOBILECLI_NOT_FOUND"
    ErrXcodeNotInstalled   ErrorCode = "XCODE_NOT_INSTALLED"

    // Input Errors
    ErrInvalidArgs         ErrorCode = "INVALID_ARGS"
    ErrMissingRequiredFlag ErrorCode = "MISSING_REQUIRED_FLAG"

    // Internal Errors
    ErrInternalError       ErrorCode = "INTERNAL_ERROR"
    ErrJSONParseFailed     ErrorCode = "JSON_PARSE_FAILED"
)

// ErrorDetails provides additional context for errors
type ErrorDetails struct {
    DeviceID       string `json:"device_id,omitempty"`
    BundleID       string `json:"bundle_id,omitempty"`
    ExpectedState  string `json:"expected_state,omitempty"`
    ActualState    string `json:"actual_state,omitempty"`
    InstallURL     string `json:"installation_url,omitempty"`
    InstallCommand string `json:"install_command,omitempty"`
    ServerCommand  string `json:"server_command,omitempty"`
}
```

### 7.2 Custom Error Types

```go
package output

import "fmt"

// DeviceNotFoundError represents a device not found error
type DeviceNotFoundError struct {
    DeviceID string
}

func (e *DeviceNotFoundError) Error() string {
    return fmt.Sprintf("device not found: %s", e.DeviceID)
}

func (e *DeviceNotFoundError) Code() ErrorCode {
    return ErrDeviceNotFound
}

// MobileCLINotAvailableError represents mobilecli availability error
type MobileCLINotAvailableError struct {
    Message        string
    InstallURL     string
    InstallCommand string
    ServerCommand  string
}

func (e *MobileCLINotAvailableError) Error() string {
    return e.Message
}

func (e *MobileCLINotAvailableError) Code() ErrorCode {
    return ErrMobileCLINotFound
}

func (e *MobileCLINotAvailableError) Details() ErrorDetails {
    return ErrorDetails{
        InstallURL:     e.InstallURL,
        InstallCommand: e.InstallCommand,
        ServerCommand:  e.ServerCommand,
    }
}

// Typed error interface
type TypedError interface {
    error
    Code() ErrorCode
}
```

### 7.3 Error Mapping Strategy

```go
package output

import (
    "strings"
)

// WrapError converts a Go error to ErrorCode and message
func WrapError(err error) (ErrorCode, string, interface{}) {
    if err == nil {
        return "", "", nil
    }

    // Check if it's a typed error
    if typedErr, ok := err.(TypedError); ok {
        code := typedErr.Code()
        msg := err.Error()

        // Extract details if available
        var details interface{}
        if detailedErr, ok := err.(interface{ Details() ErrorDetails }); ok {
            details = detailedErr.Details()
        }

        return code, msg, details
    }

    // Pattern matching for standard errors
    msg := err.Error()

    switch {
    case contains(msg, "not found"):
        return ErrDeviceNotFound, msg, nil
    case contains(msg, "unreachable"), contains(msg, "connection refused"):
        return ErrDeviceUnreachable, msg, nil
    case contains(msg, "timeout"):
        return ErrSimulatorTimeout, msg, nil
    case contains(msg, "app not found"), contains(msg, "bundle not found"):
        return ErrAppNotFound, msg, nil
    case contains(msg, "invalid argument"), contains(msg, "required flag"):
        return ErrInvalidArgs, msg, nil
    case contains(msg, "mobilecli"):
        return ErrMobileCLINotFound, msg, nil
    default:
        return ErrInternalError, msg, nil
    }
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
    return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
```

### 7.4 Command Error Handling Pattern

```go
package cmd

import (
    "github.com/neoforge-dev/ios-agent-cli/pkg/output"
    "github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
    Use:   "devices",
    Short: "List available devices",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()
        manager := getDeviceManager() // Dependency injection

        devices, err := manager.ListDevices(ctx)
        if err != nil {
            code, msg, details := output.WrapError(err)
            output.OutputError("device_discovery", code, msg, details)
            return nil // Error already output
        }

        output.OutputSuccess("device_discovery", map[string]interface{}{
            "devices": devices,
        })
        return nil
    },
}
```

---

## 8. Development Workflow

### 8.1 Local Development Setup

```bash
# 1. Clone and setup
cd /Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli
go mod tidy

# 2. Install dependencies
# (Only standard library + cobra + testify)

# 3. Verify Xcode CLI tools
xcrun simctl list

# 4. Install mobilecli (optional)
# Follow: https://github.com/mobile-next/mobilecli

# 5. Build
make build

# 6. Run
./bin/ios-agent devices
```

### 8.2 Makefile

**File**: `Makefile`

```makefile
.PHONY: build test integration-test lint install clean

# Build binary
build:
	@echo "Building ios-agent..."
	@go build -o bin/ios-agent main.go

# Run unit tests
test:
	@echo "Running unit tests..."
	@go test ./... -short -v

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	@go test ./... -tags integration -v

# Run all tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint code
lint:
	@echo "Linting code..."
	@go vet ./...
	@go fmt ./...

# Install binary to system
install: build
	@echo "Installing to /usr/local/bin..."
	@sudo cp bin/ios-agent /usr/local/bin/

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html

# Run locally
run:
	@go run main.go

# Generate mocks
mocks:
	@echo "Generating mocks..."
	@mockery --all --output test/mocks --case underscore
```

### 8.3 Git Workflow

```bash
# Feature branch workflow
git checkout -b feature/IOS-002-device-discovery

# Make changes
# Run tests
make test

# Commit
git add .
git commit -m "feat(device): implement device discovery (IOS-002)"

# Push and create PR
git push origin feature/IOS-002-device-discovery
```

### 8.4 Code Review Checklist

**Before PR**:
- [ ] All tests pass (`make test`)
- [ ] Code is linted (`make lint`)
- [ ] Integration test written (if applicable)
- [ ] Error handling implemented
- [ ] JSON output tested
- [ ] Documentation updated

**PR Review Focus**:
- [ ] Error codes match spec
- [ ] JSON response format correct
- [ ] Context propagation (cancellation)
- [ ] Timeout handling
- [ ] Idempotency (where applicable)
- [ ] Mock tests cover edge cases

---

## 9. Phase 1 Deliverables

### 9.1 Code Deliverables

**Required Files**:
```
ios-agent-cli/
├── cmd/
│   ├── root.go ✅
│   ├── devices.go
│   ├── screenshot.go
│   ├── simulator/
│   │   ├── boot.go
│   │   └── shutdown.go
│   ├── app/
│   │   ├── launch.go
│   │   ├── terminate.go
│   │   ├── install.go
│   │   └── uninstall.go
│   └── io/
│       ├── tap.go
│       ├── text.go
│       ├── swipe.go
│       └── button.go
├── pkg/
│   ├── device/
│   │   ├── manager.go (interface)
│   │   ├── local.go (implementation)
│   │   └── types.go
│   ├── xcrun/
│   │   ├── bridge.go
│   │   └── parser.go
│   ├── mobilecli/
│   │   ├── client.go
│   │   └── models.go
│   └── output/
│       ├── formatter.go
│       └── errors.go
├── test/
│   ├── integration_test.go
│   ├── mocks/
│   │   └── device_manager.go
│   └── fixtures/
│       └── sample_device_list.json
├── main.go
├── Makefile
├── go.mod
└── go.sum
```

**Binary**:
- `bin/ios-agent` (macOS ARM64 + x86_64)

### 9.2 Documentation Deliverables

**Required Docs**:
- `README.md` - Quick start, installation, basic usage
- `docs/ARCHITECTURE.md` - Design decisions, layer overview
- `docs/EXAMPLES.md` - Common workflows, agent integration examples
- `docs/IMPLEMENTATION_STRATEGY.md` ✅ (This document)

**README Structure**:
```markdown
# iOS Agent CLI

AI-agent-friendly iOS automation CLI for local simulators and remote devices.

## Quick Start

\```bash
# Install
brew install neoforge-dev/tap/ios-agent

# List devices
ios-agent devices

# Boot simulator
ios-agent simulator boot --name "iPhone 15"

# Launch app
ios-agent app launch --device <id> --bundle com.example.app

# Take screenshot
ios-agent screenshot --device <id> --output ./shot.png

# Tap
ios-agent io tap --device <id> --x 100 --y 200
\```

## Installation

### Requirements
- macOS 12+ (Monterey or later)
- Xcode 14+ with Command Line Tools
- (Optional) mobilecli for UI interactions

### Install Xcode CLI Tools
\```bash
xcode-select --install
\```

### Install mobilecli (Optional)
\```bash
# Follow: https://github.com/mobile-next/mobilecli
brew install mobilecli

# Start server
mobilecli server --listen 0.0.0.0:4723
\```

## Usage

All commands return JSON for easy parsing by AI agents.

### Device Discovery
\```bash
ios-agent devices
\```

### Simulator Lifecycle
\```bash
# Boot
ios-agent simulator boot --name "iPhone 15" --os-version 17.4

# Shutdown
ios-agent simulator shutdown --device <device-id>
\```

### App Management
\```bash
# Launch
ios-agent app launch --device <id> --bundle com.example.app --wait-for-ready 5

# Terminate
ios-agent app terminate --device <id> --bundle com.example.app

# Install
ios-agent app install --device <id> --ipa ./MyApp.ipa

# Uninstall
ios-agent app uninstall --device <id> --bundle com.example.app
\```

### UI Interactions
\```bash
# Tap
ios-agent io tap --device <id> --x 100 --y 200

# Type text
ios-agent io text --device <id> "hello world"

# Swipe
ios-agent io swipe --device <id> --start-x 100 --start-y 200 --end-x 100 --end-y 500

# Press button
ios-agent io button --device <id> --button HOME
\```

### Screenshot
\```bash
ios-agent screenshot --device <id> --format png --output ./shot.png
\```

## Agent Integration

See [docs/EXAMPLES.md](docs/EXAMPLES.md) for full agent integration examples.

## Error Handling

All errors return standardized JSON with error codes:

- `DEVICE_NOT_FOUND` - Device ID doesn't exist
- `DEVICE_UNREACHABLE` - Connection failed
- `APP_NOT_FOUND` - Bundle ID not installed
- `UI_ACTION_FAILED` - UI action failed
- `SIMULATOR_TIMEOUT` - Operation timeout
- `MOBILECLI_NOT_FOUND` - mobilecli not available

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

## License

MIT
```

### 9.3 Testing Deliverables

**Required**:
- [ ] Unit tests for all packages (70% coverage minimum)
- [ ] Integration tests for core workflows
- [ ] Mock implementations for DeviceManager
- [ ] Test fixtures (sample JSON, screenshots)

**Test Commands**:
```bash
# Run unit tests
make test

# Run integration tests
make integration-test

# Generate coverage report
make test-coverage
```

---

## 10. Risk Assessment

### 10.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **simctl output format changes** | Low | High | Parse JSON (stable), not text |
| **mobilecli not maintained** | Medium | Medium | Implement fallback to xcrun, document alternative |
| **Simulator boot takes too long** | Medium | Low | Use 60s timeout, provide progress feedback |
| **Screen coordinates vary by device** | High | Medium | Document coordinate system, provide screenshot for context |
| **App crashes during launch** | Medium | Medium | Detect crash via process list, return `APP_CRASH` error |
| **JSON parsing breaks** | Low | High | Use strict schema validation, unit tests |

### 10.2 Dependency Risks

| Dependency | Risk | Mitigation |
|------------|------|------------|
| **Xcode CLI tools** | Low - Standard on macOS | Check at startup, provide installation instructions |
| **mobilecli** | Medium - External, optional | Make optional, provide fallback for screenshot |
| **Go 1.21+** | Low - Standard | Document requirement, use go.mod |
| **cobra** | Low - Mature library | Pinned version in go.mod |

### 10.3 MVP Scope Risks

| Risk | Mitigation |
|------|------------|
| **Scope creep** | Strict P0/P1 prioritization, defer remote support to Phase 2 |
| **Timeline slippage** | Daily progress tracking, cut P1 features if needed |
| **Testing overhead** | Focus on critical path integration tests, mock for edge cases |
| **Documentation lag** | Write docs inline with features, not at end |

### 10.4 Contingency Plans

**If mobilecli unavailable**:
- Use xcrun simctl for screenshots (fallback implemented)
- Document UI interaction limitation
- Recommend users install mobilecli for full functionality

**If boot timeout too short**:
- Increase timeout to 90s
- Add `--timeout` flag for customization
- Log progress messages to stderr (not JSON stdout)

**If integration tests flaky**:
- Add retries for transient failures
- Use environment variables to skip integration tests in CI
- Document known flaky tests

**If timeline at risk**:
- Cut P1 features (swipe, button, state command)
- Focus on P0 MVP: devices, boot, launch, screenshot, tap, text
- Defer documentation polish to post-MVP

---

## Appendix A: Go Code Style Guide

### A.1 Naming Conventions

**Packages**: Short, lowercase, no underscores
- `device`, `xcrun`, `output` ✅
- `device_manager`, `xcrun_bridge` ❌

**Interfaces**: Nouns, no "I" prefix
- `DeviceManager` ✅
- `IDeviceManager` ❌

**Structs**: PascalCase
- `LocalSimulatorManager`, `Device` ✅

**Methods**: camelCase, start with verb
- `ListDevices`, `BootSimulator` ✅
- `Devices`, `Boot` ❌ (missing verb)

**Constants**: PascalCase or SCREAMING_SNAKE_CASE
- `ErrDeviceNotFound` ✅ (error codes)
- `pollInterval` ✅ (unexported)

### A.2 Error Handling

**Always wrap errors**:
```go
// Good
if err != nil {
    return fmt.Errorf("failed to boot simulator: %w", err)
}

// Bad
if err != nil {
    return err
}
```

**Use typed errors for known cases**:
```go
// Good
if device == nil {
    return &DeviceNotFoundError{DeviceID: id}
}

// Bad
if device == nil {
    return fmt.Errorf("device not found")
}
```

### A.3 Context Usage

**Always accept context**:
```go
// Good
func (m *LocalSimulatorManager) ListDevices(ctx context.Context) ([]Device, error)

// Bad
func (m *LocalSimulatorManager) ListDevices() ([]Device, error)
```

**Propagate context**:
```go
// Good
cmd := exec.CommandContext(ctx, "xcrun", "simctl", "list")

// Bad
cmd := exec.Command("xcrun", "simctl", "list")
```

### A.4 JSON Marshaling

**Use struct tags**:
```go
type Device struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    RemoteHost  string `json:"remote_host,omitempty"` // omitempty for optional fields
}
```

**Validate JSON output in tests**:
```go
func TestDeviceJSON(t *testing.T) {
    device := Device{ID: "test", Name: "iPhone"}
    jsonData, _ := json.Marshal(device)

    var decoded Device
    err := json.Unmarshal(jsonData, &decoded)
    assert.NoError(t, err)
    assert.Equal(t, device.ID, decoded.ID)
}
```

---

## Appendix B: Common Pitfalls

### B.1 xcrun simctl Gotchas

**Issue**: "Unable to boot device in current state: Booted"
- **Solution**: Check state before booting, treat as success (idempotent)

**Issue**: Boot command succeeds but device not booted
- **Solution**: Always poll state after boot command

**Issue**: JSON output changes between Xcode versions
- **Solution**: Test with multiple Xcode versions, use defensive parsing

### B.2 Context Cancellation

**Issue**: Operations continue after context cancelled
- **Solution**: Always check `ctx.Done()` in loops

```go
// Good
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-ticker.C:
        // do work
    }
}

// Bad
for {
    // do work (ignores cancellation)
}
```

### B.3 JSON Output

**Issue**: Logging to stdout breaks JSON parsing
- **Solution**: Always log to stderr, JSON to stdout only

```go
// Good
fmt.Fprintf(os.Stderr, "Debug: booting simulator\n")
outputJSON(Response{...})

// Bad
fmt.Println("Debug: booting simulator") // breaks JSON
fmt.Println(jsonString) // agent can't parse
```

### B.4 Error Handling

**Issue**: Errors swallowed by `RunE` return
- **Solution**: Use `output.OutputError` + `return nil` pattern

```go
// Good
RunE: func(cmd *cobra.Command, args []string) error {
    if err != nil {
        code, msg, _ := output.WrapError(err)
        output.OutputError(action, code, msg, nil)
        return nil // Already output error
    }
    return nil
}

// Bad
RunE: func(cmd *cobra.Command, args []string) error {
    if err != nil {
        return err // cobra prints to stderr, not JSON
    }
    return nil
}
```

---

## Summary

This implementation strategy provides a comprehensive roadmap for building the ios-agent-cli Phase 1 MVP. Key takeaways:

1. **Architecture**: Three-layer design (CLI → DeviceManager → Backend) with clear separation of concerns
2. **Implementation Order**: Foundation → Simulator → App Management → UI Interactions → Testing (2-3 weeks)
3. **xcrun Integration**: Direct command execution with JSON parsing, idempotent operations, polling for state changes
4. **mobilecli Integration**: HTTP client wrapper with graceful fallback if not available
5. **Testing**: Interface-based mocking, unit tests (70% coverage), integration tests for critical paths
6. **Error Handling**: Typed error codes, consistent JSON responses, helpful error messages

**Next Steps**:
1. Review this strategy with team
2. Begin Week 1 implementation (Foundation + Simulator Lifecycle)
3. Daily standup to track progress vs. estimated effort
4. Adjust scope if timeline at risk (cut P1 features)

**Success Metrics**:
- All P0 features working end-to-end
- Agent can control simulator and parse responses
- 70%+ code coverage
- Documentation complete

---

**Document Maintained By**: Senior Backend Engineer
**Review Schedule**: Weekly during Phase 1, ad-hoc post-MVP
**Related Docs**: `README.md`, `ARCHITECTURE.md`, `EXAMPLES.md`
