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
