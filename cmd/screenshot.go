package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/spf13/cobra"
)

var (
	screenshotOutput string
	screenshotFormat string
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Capture a screenshot from an iOS device or simulator",
	Long: `Capture a screenshot from an iOS device or simulator.

This command captures the current screen of a device and saves it to a file.
By default, screenshots are saved to /tmp with a timestamp.

Examples:
  ios-agent screenshot --device <id>                     # Save to /tmp
  ios-agent screenshot --device <id> --output shot.png  # Save to custom path
  ios-agent screenshot --device <id> --format jpeg      # Save as JPEG`,
	Run: runScreenshotCmd,
}

func init() {
	rootCmd.AddCommand(screenshotCmd)

	screenshotCmd.Flags().StringVarP(&screenshotOutput, "output", "o", "", "Output file path (default: timestamped file in /tmp)")
	screenshotCmd.Flags().StringVar(&screenshotFormat, "format", "png", "Image format: png or jpeg")
}

func runScreenshotCmd(cmd *cobra.Command, args []string) {
	// Validate device ID is provided
	if deviceID == "" {
		outputError("screenshot.capture", "DEVICE_REQUIRED", "device ID is required (use --device flag)", nil)
		return
	}

	// Validate format
	if screenshotFormat != "png" && screenshotFormat != "jpeg" {
		outputError("screenshot.capture", "INVALID_FORMAT", fmt.Sprintf("invalid format: %s (must be png or jpeg)", screenshotFormat), nil)
		return
	}

	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Verify device exists and is booted
	dev, err := manager.GetDevice(deviceID)
	if err != nil {
		outputError("screenshot.capture", "DEVICE_NOT_FOUND", err.Error(), nil)
		return
	}

	if dev.State != device.StateBooted {
		outputError("screenshot.capture", "DEVICE_NOT_BOOTED", fmt.Sprintf("device is not booted: %s (state: %s)", dev.Name, dev.State), nil)
		return
	}

	// Determine output path
	outputPath := screenshotOutput
	if outputPath == "" {
		// Generate timestamped filename in /tmp
		timestamp := time.Now().Format("20060102-150405")
		ext := screenshotFormat
		if ext == "jpeg" {
			ext = "jpg"
		}
		outputPath = filepath.Join("/tmp", fmt.Sprintf("screenshot-%s.%s", timestamp, ext))
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		outputError("screenshot.capture", "PATH_ERROR", fmt.Sprintf("failed to create output directory: %v", err), nil)
		return
	}

	// Capture screenshot
	result, err := bridge.CaptureScreenshot(dev.UDID, outputPath)
	if err != nil {
		outputError("screenshot.capture", "SCREENSHOT_FAILED", err.Error(), nil)
		return
	}

	// Output success response
	outputSuccess("screenshot.capture", result)
}
