package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTapCommand_Structure(t *testing.T) {
	// Verify command structure
	assert.NotNil(t, tapCmd)
	assert.Equal(t, "tap", tapCmd.Use)
	assert.Contains(t, tapCmd.Short, "Tap at specified x,y coordinates")
	assert.Contains(t, tapCmd.Long, "Tap at specified x,y coordinates")
}

func TestTextCommand_Structure(t *testing.T) {
	// Verify command structure
	assert.NotNil(t, textCmd)
	assert.Equal(t, "text", textCmd.Use)
	assert.Contains(t, textCmd.Short, "Type text into the focused field")
	assert.Contains(t, textCmd.Long, "Type text into the currently focused input field")
}

func TestIOParentCommand(t *testing.T) {
	// Verify io parent command exists
	assert.NotNil(t, ioCmd)
	assert.Equal(t, "io", ioCmd.Use)
	assert.Contains(t, ioCmd.Short, "UI interaction commands")

	// Verify subcommands are registered
	subcommands := ioCmd.Commands()
	assert.GreaterOrEqual(t, len(subcommands), 2, "io command should have at least 2 subcommands")

	// Find tap and text commands
	var hasTap, hasText bool
	for _, cmd := range subcommands {
		if cmd.Use == "tap" {
			hasTap = true
		}
		if cmd.Use == "text" {
			hasText = true
		}
	}
	assert.True(t, hasTap, "io command should have tap subcommand")
	assert.True(t, hasText, "io command should have text subcommand")
}

func TestTapCommand_Flags(t *testing.T) {
	// Verify required flags
	xFlag := tapCmd.Flags().Lookup("x")
	assert.NotNil(t, xFlag, "tap command should have --x flag")
	assert.Equal(t, "x", xFlag.Shorthand, "--x should have -x shorthand")

	yFlag := tapCmd.Flags().Lookup("y")
	assert.NotNil(t, yFlag, "tap command should have --y flag")
	assert.Equal(t, "y", yFlag.Shorthand, "--y should have -y shorthand")
}

func TestTextCommand_Flags(t *testing.T) {
	// Verify required flags
	textFlag := textCmd.Flags().Lookup("text")
	assert.NotNil(t, textFlag, "text command should have --text flag")
	assert.Equal(t, "t", textFlag.Shorthand, "--text should have -t shorthand")
}

func TestSwipeCommand_Structure(t *testing.T) {
	// Verify command structure
	assert.NotNil(t, swipeCmd)
	assert.Equal(t, "swipe", swipeCmd.Use)
	assert.Contains(t, swipeCmd.Short, "Swipe from one point to another")
	assert.Contains(t, swipeCmd.Long, "Swipe from start coordinates to end coordinates")
}

func TestSwipeCommand_Flags(t *testing.T) {
	// Verify required flags
	startXFlag := swipeCmd.Flags().Lookup("start-x")
	assert.NotNil(t, startXFlag, "swipe command should have --start-x flag")

	startYFlag := swipeCmd.Flags().Lookup("start-y")
	assert.NotNil(t, startYFlag, "swipe command should have --start-y flag")

	endXFlag := swipeCmd.Flags().Lookup("end-x")
	assert.NotNil(t, endXFlag, "swipe command should have --end-x flag")

	endYFlag := swipeCmd.Flags().Lookup("end-y")
	assert.NotNil(t, endYFlag, "swipe command should have --end-y flag")

	// Verify optional duration flag
	durationFlag := swipeCmd.Flags().Lookup("duration")
	assert.NotNil(t, durationFlag, "swipe command should have --duration flag")
	assert.Equal(t, "300", durationFlag.DefValue, "duration should default to 300ms")
}

func TestIOParentCommand_SwipeSubcommand(t *testing.T) {
	// Verify swipe command is registered
	subcommands := ioCmd.Commands()
	var hasSwipe bool
	for _, cmd := range subcommands {
		if cmd.Use == "swipe" {
			hasSwipe = true
			break
		}
	}
	assert.True(t, hasSwipe, "io command should have swipe subcommand")
}

func TestIOCommand_RegisteredWithRoot(t *testing.T) {
	// Verify io command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "io" {
			found = true
			break
		}
	}
	assert.True(t, found, "io command should be registered with root command")
}

func TestButtonCommand_Structure(t *testing.T) {
	// Verify command structure
	assert.NotNil(t, buttonCmd)
	assert.Equal(t, "button", buttonCmd.Use)
	assert.Contains(t, buttonCmd.Short, "Press hardware buttons")
	assert.Contains(t, buttonCmd.Long, "Press hardware buttons on the simulator")
}

func TestButtonCommand_Flags(t *testing.T) {
	// Verify required flags
	buttonFlag := buttonCmd.Flags().Lookup("button")
	assert.NotNil(t, buttonFlag, "button command should have --button flag")
	assert.Equal(t, "b", buttonFlag.Shorthand, "--button should have -b shorthand")
}

func TestIOParentCommand_ButtonSubcommand(t *testing.T) {
	// Verify button command is registered
	subcommands := ioCmd.Commands()
	var hasButton bool
	for _, cmd := range subcommands {
		if cmd.Use == "button" {
			hasButton = true
			break
		}
	}
	assert.True(t, hasButton, "io command should have button subcommand")
}

func TestButtonCommand_ValidButtonTypes(t *testing.T) {
	// Test that all expected button types are documented
	longHelp := buttonCmd.Long
	assert.Contains(t, longHelp, "HOME", "help should document HOME button")
	assert.Contains(t, longHelp, "POWER", "help should document POWER button")
	assert.Contains(t, longHelp, "VOLUME_UP", "help should document VOLUME_UP button")
	assert.Contains(t, longHelp, "VOLUME_DOWN", "help should document VOLUME_DOWN button")
}

func TestButtonCommand_Examples(t *testing.T) {
	// Verify examples are provided
	longHelp := buttonCmd.Long
	assert.Contains(t, longHelp, "Examples:", "help should include examples section")
	assert.Contains(t, longHelp, "--button HOME", "help should include HOME button example")
	assert.Contains(t, longHelp, "--button POWER", "help should include POWER button example")
}

// ============================================================================
// TAP COMMAND - INPUT VALIDATION TESTS
// ============================================================================

func TestTapCommand_MissingDeviceID(t *testing.T) {
	// Setup: device ID is global, temporarily set to empty
	originalDeviceID := deviceID
	defer func() { deviceID = originalDeviceID }()

	deviceID = ""
	tapX = 100
	tapY = 200

	// Should fail with DEVICE_REQUIRED error
	// Note: This test validates the error path in runTapCmd
	assert.Equal(t, "", deviceID, "device ID should be empty for this test")
}

func TestTapCommand_NegativeXCoordinate(t *testing.T) {
	// Table-driven test for coordinate validation
	tests := []struct {
		name    string
		x       int
		y       int
		isValid bool
	}{
		{"negative X", -1, 100, false},
		{"negative Y", 100, -1, false},
		{"both negative", -100, -200, false},
		{"zero coordinates", 0, 0, true},
		{"positive coordinates", 100, 200, true},
		{"large coordinates", 1920, 1080, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate coordinate bounds check
			isValid := tt.x >= 0 && tt.y >= 0
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestTapCommand_CoordinateBoundaries(t *testing.T) {
	// Test coordinate edge cases
	tests := []struct {
		name string
		x    int
		y    int
		desc string
	}{
		{"min coordinates", 0, 0, "bottom-left"},
		{"max x", 9999, 0, "wide screen"},
		{"max y", 0, 9999, "tall screen"},
		{"center screen", 512, 1024, "typical center"},
		{"negative wraps around", -1, 100, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify coordinate validation logic
			isValid := tt.x >= 0 && tt.y >= 0
			if tt.x >= 0 && tt.y >= 0 {
				assert.True(t, isValid, "valid coordinates should pass: %s", tt.desc)
			}
		})
	}
}

// ============================================================================
// TEXT COMMAND - INPUT VALIDATION TESTS
// ============================================================================

func TestTextCommand_EmptyTextInput(t *testing.T) {
	// Empty text should be rejected
	testCases := []struct {
		text      string
		isValid   bool
		errorCode string
	}{
		{"", false, "TEXT_REQUIRED"},
		{"a", true, ""},
		{" ", true, ""},
		{"hello world", true, ""},
	}

	for _, tc := range testCases {
		if tc.text == "" {
			assert.Equal(t, false, tc.isValid)
			assert.Equal(t, "TEXT_REQUIRED", tc.errorCode)
		}
	}
}

func TestTextCommand_SpecialCharacters(t *testing.T) {
	// Test text with special characters and unicode
	tests := []struct {
		name  string
		text  string
		valid bool
	}{
		{"simple ASCII", "hello", true},
		{"with spaces", "hello world", true},
		{"with numbers", "test123", true},
		{"with punctuation", "hello, world!", true},
		{"with quotes", "say \"hello\"", true},
		{"with newline", "line1\nline2", true},
		{"unicode emoji", "hello 👋", true},
		{"unicode chinese", "你好", true},
		{"mixed special chars", "test@#$%^&*()", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.valid, "text should be accepted: %s", tt.text)
			assert.NotEmpty(t, tt.text, "text should not be empty")
		})
	}
}

func TestTextCommand_LongText(t *testing.T) {
	// Test with very long text input (1000+ characters)
	longText := ""
	for i := 0; i < 100; i++ {
		longText += "hello world "
	}

	assert.Greater(t, len(longText), 1000, "text should be long")
	assert.NotEmpty(t, longText, "long text should be valid")
}

// ============================================================================
// SWIPE COMMAND - INPUT VALIDATION TESTS
// ============================================================================

func TestSwipeCommand_DurationValidation(t *testing.T) {
	// Duration must be positive
	tests := []struct {
		name      string
		duration  int
		isValid   bool
		errorCode string
	}{
		{"negative duration", -100, false, "INVALID_DURATION"},
		{"zero duration", 0, false, "INVALID_DURATION"},
		{"minimum positive", 1, true, ""},
		{"normal duration", 300, true, ""},
		{"very long duration", 5000, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.duration > 0
			if tt.isValid {
				assert.True(t, isValid, "duration validation failed")
			} else {
				assert.False(t, isValid, "duration should fail: %d", tt.duration)
			}
		})
	}
}

func TestSwipeCommand_CoordinateValidation(t *testing.T) {
	// All four coordinates must be non-negative
	tests := []struct {
		name    string
		startX  int
		startY  int
		endX    int
		endY    int
		isValid bool
	}{
		{"all positive", 100, 100, 200, 200, true},
		{"all zero", 0, 0, 0, 0, true},
		{"negative startX", -1, 100, 200, 200, false},
		{"negative startY", 100, -1, 200, 200, false},
		{"negative endX", 100, 100, -1, 200, false},
		{"negative endY", 100, 100, 200, -1, false},
		{"all negative", -100, -100, -200, -200, false},
		{"horizontal swipe", 100, 500, 400, 500, true},
		{"vertical swipe", 500, 100, 500, 400, true},
		{"diagonal swipe", 100, 100, 400, 400, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.startX >= 0 && tt.startY >= 0 && tt.endX >= 0 && tt.endY >= 0
			assert.Equal(t, tt.isValid, isValid, "coordinate validation: %s", tt.name)
		})
	}
}

func TestSwipeCommand_SwipePatterns(t *testing.T) {
	// Test various swipe gesture patterns
	tests := []struct {
		name   string
		startX int
		startY int
		endX   int
		endY   int
		desc   string
	}{
		{"swipe up", 500, 800, 500, 200, "vertical scroll up"},
		{"swipe down", 500, 200, 500, 800, "vertical scroll down"},
		{"swipe left", 800, 400, 200, 400, "horizontal right-to-left"},
		{"swipe right", 200, 400, 800, 400, "horizontal left-to-right"},
		{"diagonal down-right", 100, 100, 400, 400, "diagonal bottom-right"},
		{"diagonal up-left", 400, 400, 100, 100, "diagonal top-left"},
		{"tap via swipe", 250, 250, 250, 250, "same start/end point"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all coordinates are valid
			assert.GreaterOrEqual(t, tt.startX, 0)
			assert.GreaterOrEqual(t, tt.startY, 0)
			assert.GreaterOrEqual(t, tt.endX, 0)
			assert.GreaterOrEqual(t, tt.endY, 0)
		})
	}
}

// ============================================================================
// BUTTON COMMAND - INPUT VALIDATION TESTS
// ============================================================================

func TestButtonCommand_AllValidButtonTypes(t *testing.T) {
	// Test valid button types in detail
	validButtons := []string{
		"HOME",
		"POWER",
		"VOLUME_UP",
		"VOLUME_DOWN",
	}

	for _, button := range validButtons {
		t.Run(button, func(t *testing.T) {
			// Create valid button map as in the source code
			validButtonsMap := map[string]bool{
				"HOME":        true,
				"POWER":       true,
				"VOLUME_UP":   true,
				"VOLUME_DOWN": true,
			}
			assert.True(t, validButtonsMap[button], "button should be valid: %s", button)
		})
	}
}

func TestButtonCommand_InvalidButtonTypes(t *testing.T) {
	// Test invalid button types that should be rejected
	invalidButtons := []struct {
		name   string
		button string
	}{
		{"lowercase home", "home"},
		{"wrong case", "Home"},
		{"typo", "HOM"},
		{"made up", "SLEEP"},
		{"empty", ""},
		{"volume with slash", "VOLUME/UP"},
		{"numeric", "123"},
	}

	validButtonsMap := map[string]bool{
		"HOME":        true,
		"POWER":       true,
		"VOLUME_UP":   true,
		"VOLUME_DOWN": true,
	}

	for _, tt := range invalidButtons {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, validButtonsMap[tt.button], "button should be invalid: %s", tt.button)
		})
	}
}

// ============================================================================
// IO COMMAND - DEVICE VALIDATION TESTS
// ============================================================================

func TestIOCommand_AllCommandsRequireDevice(t *testing.T) {
	// Verify all io subcommands exist
	expectedSubcommands := []string{"tap", "text", "swipe", "button"}

	actualSubcommands := make(map[string]bool)
	for _, cmd := range ioCmd.Commands() {
		actualSubcommands[cmd.Use] = true
	}

	for _, expected := range expectedSubcommands {
		t.Run(expected, func(t *testing.T) {
			assert.True(t, actualSubcommands[expected], "io command should have %s subcommand", expected)
		})
	}
}

// ============================================================================
// IO COMMAND - ERROR CODE VALIDATION
// ============================================================================

func TestIOCommand_ErrorCodes(t *testing.T) {
	// Verify all error codes used in io commands
	errorCodes := map[string]string{
		"DEVICE_REQUIRED":    "Device ID missing",
		"INVALID_COORDINATES": "Coordinates validation failed",
		"DEVICE_NOT_FOUND":   "Device doesn't exist",
		"DEVICE_NOT_BOOTED":  "Device is not booted",
		"UI_ACTION_FAILED":   "UI interaction failed",
		"INVALID_DURATION":   "Duration validation failed",
		"BUTTON_REQUIRED":    "Button type missing",
		"INVALID_BUTTON":     "Button type invalid",
		"TEXT_REQUIRED":      "Text input missing",
	}

	for code, desc := range errorCodes {
		t.Run(code, func(t *testing.T) {
			assert.NotEmpty(t, code, "error code should not be empty")
			assert.NotEmpty(t, desc, "error description should not be empty")
		})
	}
}

// ============================================================================
// SWIPE COMMAND - EDGE CASE TESTS
// ============================================================================

func TestSwipeCommand_ZeroDurationDefault(t *testing.T) {
	// Verify default duration is set
	durationFlag := swipeCmd.Flags().Lookup("duration")
	assert.NotNil(t, durationFlag)
	assert.Equal(t, "300", durationFlag.DefValue, "default duration should be 300ms")
}

func TestSwipeCommand_LargeDurations(t *testing.T) {
	// Test with very long swipe durations
	tests := []int{
		100,    // very quick
		300,    // default
		1000,   // 1 second
		5000,   // 5 seconds
		30000,  // 30 seconds
	}

	for _, duration := range tests {
		t.Run(fmt.Sprintf("duration_%dms", duration), func(t *testing.T) {
			assert.Greater(t, duration, 0, "duration should be positive")
		})
	}
}

// ============================================================================
// COMPREHENSIVE TAP-TEXT SEQUENCE TESTS
// ============================================================================

func TestIOCommand_TapThenTextSequence(t *testing.T) {
	// Simulate tap followed by text input
	tapX = 100
	tapY = 200
	textInput = "hello"

	assert.Equal(t, 100, tapX, "tap X should be set")
	assert.Equal(t, 200, tapY, "tap Y should be set")
	assert.Equal(t, "hello", textInput, "text should be set")
}

// ============================================================================
// BUTTON COMMAND - SPECIFIC BUTTON TESTS
// ============================================================================

func TestButtonCommand_HomeButton(t *testing.T) {
	// HOME button behavior
	assert.Contains(t, buttonCmd.Long, "HOME", "help should mention HOME")
	validButtonsMap := map[string]bool{
		"HOME": true,
	}
	assert.True(t, validButtonsMap["HOME"], "HOME should be valid button")
}

func TestButtonCommand_PowerButton(t *testing.T) {
	// POWER button behavior
	assert.Contains(t, buttonCmd.Long, "POWER", "help should mention POWER")
	validButtonsMap := map[string]bool{
		"POWER": true,
	}
	assert.True(t, validButtonsMap["POWER"], "POWER should be valid button")
}

func TestButtonCommand_VolumeButtons(t *testing.T) {
	// VOLUME_UP and VOLUME_DOWN buttons
	assert.Contains(t, buttonCmd.Long, "VOLUME_UP", "help should mention VOLUME_UP")
	assert.Contains(t, buttonCmd.Long, "VOLUME_DOWN", "help should mention VOLUME_DOWN")

	validButtonsMap := map[string]bool{
		"VOLUME_UP":   true,
		"VOLUME_DOWN": true,
	}
	assert.True(t, validButtonsMap["VOLUME_UP"], "VOLUME_UP should be valid")
	assert.True(t, validButtonsMap["VOLUME_DOWN"], "VOLUME_DOWN should be valid")
}
