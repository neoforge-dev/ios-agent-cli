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
