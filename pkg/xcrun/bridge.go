package xcrun

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
)

// Bridge wraps xcrun simctl commands
type Bridge struct{}

// NewBridge creates a new xcrun bridge
func NewBridge() *Bridge {
	return &Bridge{}
}

// simctlDevicesResponse represents the response from `xcrun simctl list devices --json`
type simctlDevicesResponse struct {
	Devices map[string][]simctlDevice `json:"devices"`
}

// simctlDevice represents a single device from simctl
type simctlDevice struct {
	State         string `json:"state"`
	IsAvailable   bool   `json:"isAvailable"`
	Name          string `json:"name"`
	UDID          string `json:"udid"`
	DataPath      string `json:"dataPath,omitempty"`
	LogPath       string `json:"logPath,omitempty"`
	AvailabilityError string `json:"availabilityError,omitempty"`
}

// ListDevices lists all available iOS simulators
func (b *Bridge) ListDevices() ([]device.Device, error) {
	// Run xcrun simctl list devices --json
	cmd := exec.Command("xcrun", "simctl", "list", "devices", "--json")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("xcrun simctl failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run xcrun simctl: %w", err)
	}

	// Parse JSON response
	var simctlResp simctlDevicesResponse
	if err := json.Unmarshal(output, &simctlResp); err != nil {
		return nil, fmt.Errorf("failed to parse simctl output: %w", err)
	}

	// Convert simctl devices to our device format
	var devices []device.Device
	for runtime, devList := range simctlResp.Devices {
		// Extract OS version from runtime string
		// Example: "com.apple.CoreSimulator.SimRuntime.iOS-17-4" -> "17.4"
		osVersion := extractOSVersion(runtime)

		for _, simDev := range devList {
			// Only include available devices
			if !simDev.IsAvailable {
				continue
			}

			devices = append(devices, device.Device{
				ID:        simDev.UDID,
				Name:      simDev.Name,
				State:     device.DeviceState(simDev.State),
				Type:      device.DeviceTypeSimulator,
				OSVersion: osVersion,
				UDID:      simDev.UDID,
				Available: simDev.IsAvailable,
			})
		}
	}

	return devices, nil
}

// extractOSVersion extracts the OS version from a runtime string
// Example: "com.apple.CoreSimulator.SimRuntime.iOS-17-4" -> "17.4"
func extractOSVersion(runtime string) string {
	// Look for iOS version pattern
	parts := strings.Split(runtime, ".")
	for _, part := range parts {
		if strings.HasPrefix(part, "iOS-") {
			// Remove "iOS-" prefix and replace remaining dashes with dots
			version := strings.TrimPrefix(part, "iOS-")
			version = strings.ReplaceAll(version, "-", ".")
			return version
		}
	}
	return "unknown"
}

// BootSimulator boots a simulator by UDID
func (b *Bridge) BootSimulator(udid string) error {
	cmd := exec.Command("xcrun", "simctl", "boot", udid)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to boot simulator: %s", string(output))
	}
	return nil
}

// ShutdownSimulator shuts down a simulator by UDID
func (b *Bridge) ShutdownSimulator(udid string) error {
	cmd := exec.Command("xcrun", "simctl", "shutdown", udid)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to shutdown simulator: %s", string(output))
	}
	return nil
}

// GetDeviceState returns the current state of a device
func (b *Bridge) GetDeviceState(udid string) (device.DeviceState, error) {
	devices, err := b.ListDevices()
	if err != nil {
		return "", err
	}

	for _, dev := range devices {
		if dev.UDID == udid {
			return dev.State, nil
		}
	}

	return "", fmt.Errorf("device not found: %s", udid)
}

// ScreenshotResult contains metadata about a captured screenshot
type ScreenshotResult struct {
	Path      string `json:"path"`
	Format    string `json:"format"`
	SizeBytes int64  `json:"size_bytes"`
	DeviceID  string `json:"device_id"`
	Timestamp string `json:"timestamp"`
}

// CaptureScreenshot captures a screenshot from a simulator
func (b *Bridge) CaptureScreenshot(udid, outputPath string) (*ScreenshotResult, error) {
	// Run xcrun simctl io <udid> screenshot <path>
	cmd := exec.Command("xcrun", "simctl", "io", udid, "screenshot", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %s", string(output))
	}

	// Verify file was created and get its size
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("screenshot file not found after capture: %w", err)
	}

	// Determine format from file extension
	format := "png"
	if strings.HasSuffix(strings.ToLower(outputPath), ".jpg") || strings.HasSuffix(strings.ToLower(outputPath), ".jpeg") {
		format = "jpeg"
	}

	return &ScreenshotResult{
		Path:      outputPath,
		Format:    format,
		SizeBytes: fileInfo.Size(),
		DeviceID:  udid,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// TapResult contains metadata about a tap interaction
type TapResult struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	DeviceID  string `json:"device_id"`
	Timestamp string `json:"timestamp"`
}

// Tap simulates a tap at the specified coordinates
// Note: xcrun simctl doesn't support direct tap, so we use AppleScript
func (b *Bridge) Tap(udid string, x, y int) (*TapResult, error) {
	// Use AppleScript to send tap via Simulator.app
	// This is the most reliable method without requiring mobilecli
	script := fmt.Sprintf(`
tell application "System Events"
	tell process "Simulator"
		set frontmost to true
		click at {%d, %d}
	end tell
end tell
`, x, y)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If AppleScript fails, provide a helpful error message
		return nil, fmt.Errorf("failed to tap at (%d, %d): %s. Note: Simulator.app must be running and focused. For more reliable tap support, install mobilecli: https://github.com/meghaphone/mobilecli", x, y, string(output))
	}

	return &TapResult{
		X:         x,
		Y:         y,
		DeviceID:  udid,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// TextInputResult contains metadata about a text input interaction
type TextInputResult struct {
	Text      string `json:"text"`
	Length    int    `json:"length"`
	DeviceID  string `json:"device_id"`
	Timestamp string `json:"timestamp"`
}

// SwipeResult contains metadata about a swipe gesture
type SwipeResult struct {
	StartX     int    `json:"start_x"`
	StartY     int    `json:"start_y"`
	EndX       int    `json:"end_x"`
	EndY       int    `json:"end_y"`
	DurationMs int    `json:"duration_ms"`
	DeviceID   string `json:"device_id"`
	Timestamp  string `json:"timestamp"`
}

// TypeText sends text input to the simulator
func (b *Bridge) TypeText(udid, text string) (*TextInputResult, error) {
	// Use xcrun simctl io <udid> sendkey <text>
	// Note: simctl keyboardinput is more reliable for text input
	cmd := exec.Command("xcrun", "simctl", "keyboardinput", udid, text)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to type text: %s", string(output))
	}

	return &TextInputResult{
		Text:      text,
		Length:    len(text),
		DeviceID:  udid,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}


// ButtonResult contains metadata about a button press interaction
type ButtonResult struct {
	Button    string `json:"button"`
	DeviceID  string `json:"device_id"`
	Timestamp string `json:"timestamp"`
}

// PressButton presses a hardware button on the simulator
func (b *Bridge) PressButton(udid, button string) (*ButtonResult, error) {
	// Map button types to simctl commands
	// For HOME button, use: xcrun simctl ui <udid> click home
	// For other buttons, we may need AppleScript or keyboard shortcuts

	var cmd *exec.Cmd

	switch button {
	case "HOME":
		// Use simctl ui click home
		cmd = exec.Command("xcrun", "simctl", "ui", udid, "click", "home")
	case "POWER":
		// Power button - use keyboard shortcut via AppleScript
		// Cmd+L locks the screen
		script := `
tell application "System Events"
	tell process "Simulator"
		set frontmost to true
		keystroke "l" using {command down}
	end tell
end tell
`
		cmd = exec.Command("osascript", "-e", script)
	case "VOLUME_UP":
		// Volume up - use keyboard shortcut via AppleScript
		script := `
tell application "System Events"
	tell process "Simulator"
		set frontmost to true
		key code 126
	end tell
end tell
`
		cmd = exec.Command("osascript", "-e", script)
	case "VOLUME_DOWN":
		// Volume down - use keyboard shortcut via AppleScript
		script := `
tell application "System Events"
	tell process "Simulator"
		set frontmost to true
		key code 125
	end tell
end tell
`
		cmd = exec.Command("osascript", "-e", script)
	default:
		return nil, fmt.Errorf("unsupported button type: %s", button)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to press %s button: %s", button, string(output))
	}

	return &ButtonResult{
		Button:    button,
		DeviceID:  udid,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
// Swipe simulates a swipe gesture from start point to end point
// Note: xcrun simctl doesn't support direct swipe, so we use AppleScript
func (b *Bridge) Swipe(udid string, startX, startY, endX, endY, durationMs int) (*SwipeResult, error) {
	// Use AppleScript to send swipe gesture via Simulator.app
	// AppleScript doesn't have native swipe support, so we simulate it with drag
	// Duration is converted to approximate delay in AppleScript
	delaySeconds := float64(durationMs) / 1000.0

	script := fmt.Sprintf(`
tell application "System Events"
	tell process "Simulator"
		set frontmost to true
		-- Simulate swipe as mouse drag
		set startPoint to {%d, %d}
		set endPoint to {%d, %d}

		-- Move to start position and hold down mouse
		do shell script "cliclick m:" & %d & "," & %d
		delay 0.05
		do shell script "cliclick dd:" & %d & "," & %d
		delay %f
		do shell script "cliclick du:" & %d & "," & %d
	end tell
end tell
`, startX, startY, endX, endY, startX, startY, startX, startY, delaySeconds, endX, endY)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If AppleScript fails, provide a helpful error message
		return nil, fmt.Errorf("failed to swipe from (%d, %d) to (%d, %d): %s. Note: Simulator.app must be running and focused. This implementation requires cliclick tool: brew install cliclick", startX, startY, endX, endY, string(output))
	}

	return &SwipeResult{
		StartX:     startX,
		StartY:     startY,
		EndX:       endX,
		EndY:       endY,
		DurationMs: durationMs,
		DeviceID:   udid,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// LaunchApp launches an app on a simulator by bundle ID
// Returns the PID of the launched process
func (b *Bridge) LaunchApp(udid, bundleID string) (string, error) {
	// Run xcrun simctl launch <udid> <bundle-id>
	// Output format: "<bundle-id>: <pid>"
	cmd := exec.Command("xcrun", "simctl", "launch", udid, bundleID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to launch app: %s", string(output))
	}

	// Parse PID from output
	// Example output: "com.example.app: 12345"
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Split(outputStr, ":")
	if len(parts) == 2 {
		pid := strings.TrimSpace(parts[1])
		return pid, nil
	}

	// If we can't parse PID, still return success since launch succeeded
	return "", nil
}

// TerminateApp terminates a running app on a simulator by bundle ID
func (b *Bridge) TerminateApp(udid, bundleID string) error {
	// Run xcrun simctl terminate <udid> <bundle-id>
	cmd := exec.Command("xcrun", "simctl", "terminate", udid, bundleID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error is because app is not running
		// xcrun simctl terminate may fail if app is not running
		outputStr := string(output)
		if strings.Contains(outputStr, "No matching processes") {
			// App was not running, consider this success
			return nil
		}
		return fmt.Errorf("failed to terminate app: %s", outputStr)
	}
	return nil
}

// InstallApp installs an app on a simulator
// Returns the bundle ID of the installed app
func (b *Bridge) InstallApp(udid, appPath string) (string, error) {
	// Run xcrun simctl install <udid> <app-path>
	cmd := exec.Command("xcrun", "simctl", "install", udid, appPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to install app: %s", string(output))
	}

	// Extract bundle ID from the app bundle
	// Run plutil to read Info.plist from the app bundle
	infoPlistPath := fmt.Sprintf("%s/Info.plist", appPath)
	plistCmd := exec.Command("plutil", "-extract", "CFBundleIdentifier", "raw", infoPlistPath)
	plistOutput, err := plistCmd.Output()
	if err != nil {
		// If we can't extract bundle ID, return empty string (install still succeeded)
		return "", nil
	}

	bundleID := strings.TrimSpace(string(plistOutput))
	return bundleID, nil
}

// UninstallApp uninstalls an app from a simulator by bundle ID
func (b *Bridge) UninstallApp(udid, bundleID string) error {
	// Run xcrun simctl uninstall <udid> <bundle-id>
	cmd := exec.Command("xcrun", "simctl", "uninstall", udid, bundleID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall app: %s", string(output))
	}
	return nil
}

// ForegroundAppInfo contains info about the foreground app
type ForegroundAppInfo struct {
	BundleID string `json:"bundle_id"`
	PID      int    `json:"pid"`
}

// GetForegroundApp attempts to identify the foreground app on a booted simulator
// Note: This is best-effort as simctl doesn't provide direct foreground app info
func (b *Bridge) GetForegroundApp(udid string) (*ForegroundAppInfo, error) {
	// Use `xcrun simctl spawn` to run `ps` and find the most likely foreground app
	// We look for processes with certain characteristics that indicate they're user apps

	// First, try to get the list of running processes
	cmd := exec.Command("xcrun", "simctl", "spawn", udid, "ps", "-A", "-o", "pid,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get running processes: %w", err)
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	// Parse process list and look for SpringBoard and other system processes
	// The most recent user app launched is typically the foreground app
	// We'll use a heuristic: find the most recent process that looks like an app

	var lastUserAppPID int
	var lastUserAppComm string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "PID") {
			continue
		}

		// Split by whitespace
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pid := fields[0]
		comm := strings.Join(fields[1:], " ")

		// Skip system processes
		if strings.Contains(comm, "launchd") ||
			strings.Contains(comm, "logd") ||
			strings.Contains(comm, "UserEventAgent") ||
			strings.Contains(comm, "configd") ||
			strings.Contains(comm, "nsurlsessiond") ||
			strings.Contains(comm, "SpringBoard") {
			continue
		}

		// Look for processes that look like apps (have .app in path)
		if strings.Contains(comm, ".app") {
			// Extract bundle ID from the path
			// Example: /Applications/Maps.app/Maps -> com.apple.Maps
			lastUserAppComm = comm
			if pidInt, err := fmt.Sscanf(pid, "%d", &lastUserAppPID); err == nil && pidInt == 1 {
				// Successfully parsed PID
			}
		}
	}

	// If we found a potential foreground app, try to get its bundle ID
	if lastUserAppComm != "" {
		// Extract app name from the path
		// Example: /Applications/Maps.app/Maps -> Maps
		parts := strings.Split(lastUserAppComm, "/")
		var appName string
		for _, part := range parts {
			if strings.HasSuffix(part, ".app") {
				appName = strings.TrimSuffix(part, ".app")
				break
			}
		}

		if appName != "" {
			// Try to map app name to bundle ID by listing apps
			// This is a best-effort approach
			bundleID, err := b.findBundleIDByAppName(udid, appName)
			if err == nil && bundleID != "" {
				return &ForegroundAppInfo{
					BundleID: bundleID,
					PID:      lastUserAppPID,
				}, nil
			}
		}
	}

	// If we couldn't determine foreground app, return nil (not an error)
	return nil, nil
}

// findBundleIDByAppName attempts to find a bundle ID by app name
func (b *Bridge) findBundleIDByAppName(udid, appName string) (string, error) {
	// Run xcrun simctl listapps to get all installed apps
	cmd := exec.Command("xcrun", "simctl", "listapps", udid)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list apps: %w", err)
	}

	outputStr := string(output)

	// Parse the plist-style output to find bundle IDs
	// Look for bundle IDs that might match the app name
	lines := strings.Split(outputStr, "\n")

	for i, line := range lines {
		// Look for bundle ID lines (they start with quotes)
		if strings.HasPrefix(strings.TrimSpace(line), "\"com.") {
			bundleID := strings.Trim(strings.TrimSpace(strings.Split(line, "=")[0]), "\"")

			// Check subsequent lines for CFBundleExecutable or CFBundleDisplayName
			// that matches our app name
			for j := i + 1; j < len(lines) && j < i+20; j++ {
				checkLine := lines[j]
				if strings.Contains(checkLine, "CFBundleExecutable") ||
					strings.Contains(checkLine, "CFBundleDisplayName") ||
					strings.Contains(checkLine, "CFBundleName") {
					if strings.Contains(checkLine, appName) {
						return bundleID, nil
					}
				}
				// Stop at the next bundle ID
				if strings.HasPrefix(strings.TrimSpace(checkLine), "\"com.") {
					break
				}
			}
		}
	}

	return "", fmt.Errorf("bundle ID not found for app: %s", appName)
}
