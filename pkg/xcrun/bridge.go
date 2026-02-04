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
