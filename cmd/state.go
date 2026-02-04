package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/spf13/cobra"
)

var (
	includeScreenshot bool
)

// StateResult represents the complete device state snapshot
type StateResult struct {
	Device         *DeviceInfo         `json:"device"`
	ForegroundApp  *ForegroundAppInfo  `json:"foreground_app,omitempty"`
	Screenshot     string              `json:"screenshot,omitempty"`
}

// DeviceInfo represents device information
type DeviceInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	OSVersion string `json:"os_version"`
	Runtime   string `json:"runtime"`
}

// ForegroundAppInfo represents foreground app information
type ForegroundAppInfo struct {
	BundleID string `json:"bundle_id,omitempty"`
	PID      int    `json:"pid,omitempty"`
}

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Get comprehensive device state snapshot",
	Long: `Get comprehensive device state snapshot including device info,
foreground app, and optionally a screenshot.

This command provides a complete snapshot of the device state, useful for
AI agents to understand the current device context before performing actions.

Examples:
  ios-agent state --device <id>                    # Basic state info
  ios-agent state --device <id> --include-screenshot  # Include screenshot`,
	Run: runStateCmd,
}

func init() {
	rootCmd.AddCommand(stateCmd)

	stateCmd.Flags().BoolVar(&includeScreenshot, "include-screenshot", false, "Include screenshot in state snapshot")
}

func runStateCmd(cmd *cobra.Command, args []string) {
	// Validate device ID is provided
	if deviceID == "" {
		outputError("state", "DEVICE_REQUIRED", "device ID is required (use --device flag)", nil)
		return
	}

	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Verify device exists
	dev, err := manager.GetDevice(deviceID)
	if err != nil {
		outputError("state", "DEVICE_NOT_FOUND", err.Error(), nil)
		return
	}

	// Build device info
	deviceInfo := &DeviceInfo{
		ID:        dev.ID,
		Name:      dev.Name,
		State:     string(dev.State),
		OSVersion: dev.OSVersion,
		Runtime:   fmt.Sprintf("iOS %s", dev.OSVersion),
	}

	result := &StateResult{
		Device: deviceInfo,
	}

	// Get foreground app info only if device is booted
	if dev.State == device.StateBooted {
		foregroundApp, err := bridge.GetForegroundApp(dev.UDID)
		if err != nil {
			// Don't fail the command if we can't get foreground app
			// Just log verbosely if enabled
			if verbose {
				fmt.Printf("Warning: Could not determine foreground app: %v\n", err)
			}
		} else if foregroundApp != nil {
			result.ForegroundApp = &ForegroundAppInfo{
				BundleID: foregroundApp.BundleID,
				PID:      foregroundApp.PID,
			}
		}

		// Capture screenshot if requested
		if includeScreenshot {
			// Generate timestamped filename in /tmp
			timestamp := time.Now().Format("20060102-150405")
			screenshotPath := filepath.Join("/tmp", fmt.Sprintf("state-screenshot-%s.png", timestamp))

			screenshotResult, err := bridge.CaptureScreenshot(dev.UDID, screenshotPath)
			if err != nil {
				// Don't fail the command if screenshot capture fails
				if verbose {
					fmt.Printf("Warning: Could not capture screenshot: %v\n", err)
				}
			} else {
				result.Screenshot = screenshotResult.Path
			}
		}
	} else {
		// Device is not booted, can't get foreground app or screenshot
		if includeScreenshot {
			outputError("state", "DEVICE_NOT_BOOTED",
				fmt.Sprintf("device is not booted: %s (state: %s). Cannot capture screenshot.", dev.Name, dev.State), nil)
			return
		}
	}

	// Output success response
	outputSuccess("state", result)
}
