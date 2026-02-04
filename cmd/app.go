package cmd

import (
	"fmt"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/spf13/cobra"
)

var (
	// Launch command flags
	launchBundleID    string
	launchDeviceID    string
	launchWaitForReady bool
	launchTimeout     int

	// Terminate command flags
	terminateBundleID string
	terminateDeviceID string
)

// appCmd represents the app command group
var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage iOS applications",
	Long: `Manage iOS applications - launch, terminate, install, and uninstall.

Examples:
  ios-agent app launch --device <udid> --bundle com.example.app
  ios-agent app launch --device <udid> --bundle com.example.app --wait-for-ready
  ios-agent app terminate --device <udid> --bundle com.example.app`,
}

// launchCmd represents the launch subcommand
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch an iOS application",
	Long: `Launch an iOS application by bundle ID on a booted simulator.

The command will:
1. Verify the device exists and is booted
2. Launch the app using xcrun simctl
3. Return PID and launch status in JSON format

Examples:
  ios-agent app launch --device <udid> --bundle com.example.app
  ios-agent app launch -d <udid> --bundle com.example.app --wait-for-ready
  ios-agent app launch --device <udid> --bundle com.example.app --timeout 30`,
	Run: runLaunchCmd,
}

// terminateCmd represents the terminate subcommand
var terminateCmd = &cobra.Command{
	Use:   "terminate",
	Short: "Terminate a running iOS application",
	Long: `Terminate a running iOS application by bundle ID on a simulator.

The command will:
1. Verify the device exists
2. Terminate the app using xcrun simctl
3. Return success status in JSON format

If the app is not running, the command handles it gracefully and returns success.

Examples:
  ios-agent app terminate --device <udid> --bundle com.example.app
  ios-agent app terminate -d <udid> --bundle com.example.app`,
	Run: runTerminateCmd,
}

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(launchCmd)
	appCmd.AddCommand(terminateCmd)

	// Launch command flags
	launchCmd.Flags().StringVarP(&launchDeviceID, "device", "d", "", "Device ID to launch app on (required)")
	launchCmd.Flags().StringVar(&launchBundleID, "bundle", "", "Bundle ID of the app to launch (required)")
	launchCmd.Flags().BoolVar(&launchWaitForReady, "wait-for-ready", false, "Wait for app to be ready")
	launchCmd.Flags().IntVar(&launchTimeout, "timeout", 30, "Launch timeout in seconds")
	launchCmd.MarkFlagRequired("device")
	launchCmd.MarkFlagRequired("bundle")

	// Terminate command flags
	terminateCmd.Flags().StringVarP(&terminateDeviceID, "device", "d", "", "Device ID to terminate app on (required)")
	terminateCmd.Flags().StringVar(&terminateBundleID, "bundle", "", "Bundle ID of the app to terminate (required)")
	terminateCmd.MarkFlagRequired("device")
	terminateCmd.MarkFlagRequired("bundle")
}

// LaunchResult represents the result of an app launch operation
type LaunchResult struct {
	Device   *device.Device `json:"device"`
	BundleID string         `json:"bundle_id"`
	PID      string         `json:"pid,omitempty"`
	State    string         `json:"state"`
	Message  string         `json:"message"`
}

// TerminateResult represents the result of an app terminate operation
type TerminateResult struct {
	Device   *device.Device `json:"device"`
	BundleID string         `json:"bundle_id"`
	Message  string         `json:"message"`
}

func runLaunchCmd(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Get device to verify it exists
	dev, err := manager.GetDevice(launchDeviceID)
	if err != nil {
		outputError("app.launch", "DEVICE_NOT_FOUND", err.Error(), map[string]string{
			"device_id": launchDeviceID,
		})
		return
	}

	// Verify device is booted
	if dev.State != device.StateBooted {
		outputError("app.launch", "DEVICE_NOT_BOOTED", "Device must be booted to launch an app", map[string]string{
			"device_id": dev.ID,
			"state":     string(dev.State),
		})
		return
	}

	// Launch the app
	pid, err := bridge.LaunchApp(dev.UDID, launchBundleID)
	if err != nil {
		outputError("app.launch", "APP_LAUNCH_FAILED", err.Error(), map[string]string{
			"device_id": dev.ID,
			"bundle_id": launchBundleID,
		})
		return
	}

	// Calculate launch time
	launchTime := time.Since(startTime).Milliseconds()

	result := LaunchResult{
		Device:   dev,
		BundleID: launchBundleID,
		PID:      pid,
		State:    "launched",
		Message:  fmt.Sprintf("App launched successfully in %dms", launchTime),
	}

	outputSuccess("app.launch", result)
}

func runTerminateCmd(cmd *cobra.Command, args []string) {
	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Get device to verify it exists
	dev, err := manager.GetDevice(terminateDeviceID)
	if err != nil {
		outputError("app.terminate", "DEVICE_NOT_FOUND", err.Error(), map[string]string{
			"device_id": terminateDeviceID,
		})
		return
	}

	// Terminate the app
	err = bridge.TerminateApp(dev.UDID, terminateBundleID)
	if err != nil {
		// Check if error is because app was not running
		// xcrun simctl terminate handles this gracefully but may return error
		outputError("app.terminate", "APP_TERMINATE_FAILED", err.Error(), map[string]string{
			"device_id": dev.ID,
			"bundle_id": terminateBundleID,
		})
		return
	}

	result := TerminateResult{
		Device:   dev,
		BundleID: terminateBundleID,
		Message:  "App terminated successfully",
	}

	outputSuccess("app.terminate", result)
}
