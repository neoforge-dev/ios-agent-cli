package cmd

import (
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
