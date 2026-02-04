package cmd

import (
	"fmt"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/spf13/cobra"
)

var (
	// Tap flags
	tapX int
	tapY int

	// Text flags
	textInput string
)

// ioCmd represents the io parent command
var ioCmd = &cobra.Command{
	Use:   "io",
	Short: "UI interaction commands (tap, text, swipe, etc.)",
	Long: `UI interaction commands for iOS simulators.

This command provides subcommands for interacting with the UI:
  - tap: Tap at x,y coordinates
  - text: Type text into the focused field

Examples:
  ios-agent io tap --device <id> --x 100 --y 200
  ios-agent io text --device <id> --text "Hello World"`,
}

// tapCmd implements the tap interaction
var tapCmd = &cobra.Command{
	Use:   "tap",
	Short: "Tap at specified x,y coordinates",
	Long: `Tap at specified x,y coordinates on the simulator screen.

This command simulates a tap gesture at the given coordinates.
Coordinates are relative to the screen size of the device.

Examples:
  ios-agent io tap --device <id> --x 100 --y 200
  ios-agent io tap -d <id> -x 160 -y 300`,
	Run: runTapCmd,
}

// textCmd implements text input
var textCmd = &cobra.Command{
	Use:   "text",
	Short: "Type text into the focused field",
	Long: `Type text into the currently focused input field.

This command sends text input to the simulator. The target field
must already be focused (e.g., by tapping on it first).

Examples:
  ios-agent io text --device <id> --text "Hello World"
  ios-agent io text -d <id> --text "user@example.com"`,
	Run: runTextCmd,
}

func init() {
	rootCmd.AddCommand(ioCmd)
	ioCmd.AddCommand(tapCmd)
	ioCmd.AddCommand(textCmd)

	// Tap command flags
	tapCmd.Flags().IntVarP(&tapX, "x", "x", 0, "X coordinate for tap")
	tapCmd.Flags().IntVarP(&tapY, "y", "y", 0, "Y coordinate for tap")
	tapCmd.MarkFlagRequired("x")
	tapCmd.MarkFlagRequired("y")

	// Text command flags
	textCmd.Flags().StringVarP(&textInput, "text", "t", "", "Text to type")
	textCmd.MarkFlagRequired("text")
}

func runTapCmd(cmd *cobra.Command, args []string) {
	// Validate device ID is provided
	if deviceID == "" {
		outputError("io.tap", "DEVICE_REQUIRED", "device ID is required (use --device flag)", nil)
		return
	}

	// Validate coordinates are non-negative
	if tapX < 0 || tapY < 0 {
		outputError("io.tap", "INVALID_COORDINATES", fmt.Sprintf("coordinates must be non-negative: x=%d, y=%d", tapX, tapY), nil)
		return
	}

	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Verify device exists and is booted
	dev, err := manager.GetDevice(deviceID)
	if err != nil {
		outputError("io.tap", "DEVICE_NOT_FOUND", err.Error(), nil)
		return
	}

	if dev.State != device.StateBooted {
		outputError("io.tap", "DEVICE_NOT_BOOTED", fmt.Sprintf("device is not booted: %s (state: %s)", dev.Name, dev.State), nil)
		return
	}

	// Perform tap
	result, err := bridge.Tap(dev.UDID, tapX, tapY)
	if err != nil {
		outputError("io.tap", "UI_ACTION_FAILED", err.Error(), nil)
		return
	}

	// Output success response
	outputSuccess("io.tap", result)
}

func runTextCmd(cmd *cobra.Command, args []string) {
	// Validate device ID is provided
	if deviceID == "" {
		outputError("io.text", "DEVICE_REQUIRED", "device ID is required (use --device flag)", nil)
		return
	}

	// Validate text is not empty
	if textInput == "" {
		outputError("io.text", "TEXT_REQUIRED", "text input cannot be empty", nil)
		return
	}

	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Verify device exists and is booted
	dev, err := manager.GetDevice(deviceID)
	if err != nil {
		outputError("io.text", "DEVICE_NOT_FOUND", err.Error(), nil)
		return
	}

	if dev.State != device.StateBooted {
		outputError("io.text", "DEVICE_NOT_BOOTED", fmt.Sprintf("device is not booted: %s (state: %s)", dev.Name, dev.State), nil)
		return
	}

	// Send text input
	result, err := bridge.TypeText(dev.UDID, textInput)
	if err != nil {
		outputError("io.text", "UI_ACTION_FAILED", err.Error(), nil)
		return
	}

	// Output success response
	outputSuccess("io.text", result)
}
