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

	// Button flags
	buttonType string

	// Swipe flags
	swipeStartX   int
	swipeStartY   int
	swipeEndX     int
	swipeEndY     int
	swipeDuration int
)

// ioCmd represents the io parent command
var ioCmd = &cobra.Command{
	Use:   "io",
	Short: "UI interaction commands (tap, text, swipe, button, etc.)",
	Long: `UI interaction commands for iOS simulators.

This command provides subcommands for interacting with the UI:
  - tap: Tap at x,y coordinates
  - text: Type text into the focused field
  - swipe: Swipe from one point to another
  - button: Press hardware buttons (HOME, POWER, etc.)

Examples:
  ios-agent io tap --device <id> --x 100 --y 200
  ios-agent io text --device <id> --text "Hello World"
  ios-agent io swipe --device <id> --start-x 100 --start-y 200 --end-x 100 --end-y 600
  ios-agent io button --device <id> --button HOME`,
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

// swipeCmd implements swipe gesture
var swipeCmd = &cobra.Command{
	Use:   "swipe",
	Short: "Swipe from one point to another",
	Long: `Swipe from start coordinates to end coordinates on the simulator screen.

This command simulates a swipe gesture between two points. You can optionally
specify the duration of the swipe in milliseconds.

Coordinates are relative to the screen size of the device.

Examples:
  ios-agent io swipe --device <id> --start-x 100 --start-y 200 --end-x 100 --end-y 600
  ios-agent io swipe -d <id> --start-x 300 --start-y 400 --end-x 100 --end-y 400 --duration 500
  ios-agent io swipe -d <id> --start-x 200 --start-y 800 --end-x 200 --end-y 100`,
	Run: runSwipeCmd,
}

// buttonCmd implements hardware button press
var buttonCmd = &cobra.Command{
	Use:   "button",
	Short: "Press hardware buttons (HOME, POWER, VOLUME_UP, VOLUME_DOWN)",
	Long: `Press hardware buttons on the simulator.

This command simulates pressing physical hardware buttons like HOME, POWER,
VOLUME_UP, and VOLUME_DOWN.

Supported buttons:
  - HOME: Home button press
  - POWER: Power/lock button
  - VOLUME_UP: Volume up button
  - VOLUME_DOWN: Volume down button

Examples:
  ios-agent io button --device <id> --button HOME
  ios-agent io button -d <id> --button POWER
  ios-agent io button -d <id> --button VOLUME_UP`,
	Run: runButtonCmd,
}

func init() {
	rootCmd.AddCommand(ioCmd)
	ioCmd.AddCommand(tapCmd)
	ioCmd.AddCommand(textCmd)
	ioCmd.AddCommand(swipeCmd)
	ioCmd.AddCommand(buttonCmd)

	// Tap command flags
	tapCmd.Flags().IntVarP(&tapX, "x", "x", 0, "X coordinate for tap")
	tapCmd.Flags().IntVarP(&tapY, "y", "y", 0, "Y coordinate for tap")
	tapCmd.MarkFlagRequired("x")
	tapCmd.MarkFlagRequired("y")

	// Text command flags
	textCmd.Flags().StringVarP(&textInput, "text", "t", "", "Text to type")
	textCmd.MarkFlagRequired("text")

	// Swipe command flags
	swipeCmd.Flags().IntVar(&swipeStartX, "start-x", 0, "Starting X coordinate")
	swipeCmd.Flags().IntVar(&swipeStartY, "start-y", 0, "Starting Y coordinate")
	swipeCmd.Flags().IntVar(&swipeEndX, "end-x", 0, "Ending X coordinate")
	swipeCmd.Flags().IntVar(&swipeEndY, "end-y", 0, "Ending Y coordinate")
	swipeCmd.Flags().IntVar(&swipeDuration, "duration", 300, "Swipe duration in milliseconds (default: 300ms)")
	swipeCmd.MarkFlagRequired("start-x")
	swipeCmd.MarkFlagRequired("start-y")
	swipeCmd.MarkFlagRequired("end-x")
	swipeCmd.MarkFlagRequired("end-y")

	// Button command flags
	buttonCmd.Flags().StringVarP(&buttonType, "button", "b", "", "Button type (HOME, POWER, VOLUME_UP, VOLUME_DOWN)")
	buttonCmd.MarkFlagRequired("button")
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

func runButtonCmd(cmd *cobra.Command, args []string) {
	// Validate device ID is provided
	if deviceID == "" {
		outputError("io.button", "DEVICE_REQUIRED", "device ID is required (use --device flag)", nil)
		return
	}

	// Validate button type is provided
	if buttonType == "" {
		outputError("io.button", "BUTTON_REQUIRED", "button type is required (use --button flag)", nil)
		return
	}

	// Validate button type is supported
	validButtons := map[string]bool{
		"HOME":        true,
		"POWER":       true,
		"VOLUME_UP":   true,
		"VOLUME_DOWN": true,
	}
	if !validButtons[buttonType] {
		outputError("io.button", "INVALID_BUTTON", fmt.Sprintf("invalid button type: %s (must be one of: HOME, POWER, VOLUME_UP, VOLUME_DOWN)", buttonType), nil)
		return
	}

	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Verify device exists and is booted
	dev, err := manager.GetDevice(deviceID)
	if err != nil {
		outputError("io.button", "DEVICE_NOT_FOUND", err.Error(), nil)
		return
	}

	if dev.State != device.StateBooted {
		outputError("io.button", "DEVICE_NOT_BOOTED", fmt.Sprintf("device is not booted: %s (state: %s)", dev.Name, dev.State), nil)
		return
	}

	// Press button
	result, err := bridge.PressButton(dev.UDID, buttonType)
	if err != nil {
		outputError("io.button", "UI_ACTION_FAILED", err.Error(), nil)
		return
	}

	// Output success response
	outputSuccess("io.button", result)
}

func runSwipeCmd(cmd *cobra.Command, args []string) {
	// Validate device ID is provided
	if deviceID == "" {
		outputError("io.swipe", "DEVICE_REQUIRED", "device ID is required (use --device flag)", nil)
		return
	}

	// Validate coordinates are non-negative
	if swipeStartX < 0 || swipeStartY < 0 || swipeEndX < 0 || swipeEndY < 0 {
		outputError("io.swipe", "INVALID_COORDINATES",
			fmt.Sprintf("coordinates must be non-negative: start=(%d, %d), end=(%d, %d)",
				swipeStartX, swipeStartY, swipeEndX, swipeEndY), nil)
		return
	}

	// Validate duration is positive
	if swipeDuration <= 0 {
		outputError("io.swipe", "INVALID_DURATION",
			fmt.Sprintf("duration must be positive: %dms", swipeDuration), nil)
		return
	}

	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Verify device exists and is booted
	dev, err := manager.GetDevice(deviceID)
	if err != nil {
		outputError("io.swipe", "DEVICE_NOT_FOUND", err.Error(), nil)
		return
	}

	if dev.State != device.StateBooted {
		outputError("io.swipe", "DEVICE_NOT_BOOTED",
			fmt.Sprintf("device is not booted: %s (state: %s)", dev.Name, dev.State), nil)
		return
	}

	// Perform swipe
	result, err := bridge.Swipe(dev.UDID, swipeStartX, swipeStartY, swipeEndX, swipeEndY, swipeDuration)
	if err != nil {
		outputError("io.swipe", "UI_ACTION_FAILED", err.Error(), nil)
		return
	}

	// Output success response
	outputSuccess("io.swipe", result)
}
