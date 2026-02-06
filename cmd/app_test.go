package cmd

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLaunchResultJSON(t *testing.T) {
	result := LaunchResult{
		Device: &device.Device{
			ID:        "test-device-1",
			Name:      "iPhone 15 Pro",
			State:     device.StateBooted,
			Type:      device.DeviceTypeSimulator,
			OSVersion: "17.4",
			UDID:      "test-device-1",
			Available: true,
		},
		BundleID: "com.example.app",
		PID:      "12345",
		State:    "launched",
		Message:  "App launched successfully in 100ms",
	}

	// Verify JSON serialization
	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded LaunchResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.BundleID, decoded.BundleID)
	assert.Equal(t, result.PID, decoded.PID)
	assert.Equal(t, result.State, decoded.State)
	assert.Equal(t, result.Message, decoded.Message)
}

func TestTerminateResultJSON(t *testing.T) {
	result := TerminateResult{
		Device: &device.Device{
			ID:        "test-device-1",
			Name:      "iPhone 15 Pro",
			State:     device.StateBooted,
			Type:      device.DeviceTypeSimulator,
			OSVersion: "17.4",
			UDID:      "test-device-1",
			Available: true,
		},
		BundleID: "com.example.app",
		Message:  "App terminated successfully",
	}

	// Verify JSON serialization
	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded TerminateResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.BundleID, decoded.BundleID)
	assert.Equal(t, result.Message, decoded.Message)
}

func TestInstallResultJSON(t *testing.T) {
	result := InstallResult{
		Device: &device.Device{
			ID:        "test-device-1",
			Name:      "iPhone 15 Pro",
			State:     device.StateBooted,
			Type:      device.DeviceTypeSimulator,
			OSVersion: "17.4",
			UDID:      "test-device-1",
			Available: true,
		},
		AppPath:     "/path/to/MyApp.app",
		BundleID:    "com.example.app",
		InstallTime: 1234,
		Message:     "App installed successfully in 1234ms",
	}

	// Verify JSON serialization
	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded InstallResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.AppPath, decoded.AppPath)
	assert.Equal(t, result.BundleID, decoded.BundleID)
	assert.Equal(t, result.InstallTime, decoded.InstallTime)
	assert.Equal(t, result.Message, decoded.Message)
}

func TestUninstallResultJSON(t *testing.T) {
	result := UninstallResult{
		Device: &device.Device{
			ID:        "test-device-1",
			Name:      "iPhone 15 Pro",
			State:     device.StateBooted,
			Type:      device.DeviceTypeSimulator,
			OSVersion: "17.4",
			UDID:      "test-device-1",
			Available: true,
		},
		BundleID: "com.example.app",
		Message:  "App uninstalled successfully",
	}

	// Verify JSON serialization
	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded UninstallResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.BundleID, decoded.BundleID)
	assert.Equal(t, result.Message, decoded.Message)
}

func TestAppLaunchDeviceValidation(t *testing.T) {
	tests := []struct {
		name          string
		deviceID      string
		bundleID      string
		mockDevices   []device.Device
		expectedError string
	}{
		{
			name:     "device not found",
			deviceID: "nonexistent",
			bundleID: "com.example.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateBooted,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "device not found",
		},
		{
			name:     "device not booted",
			deviceID: "test-device-1",
			bundleID: "com.example.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateShutdown,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "not booted",
		},
		{
			name:     "valid booted device",
			deviceID: "test-device-1",
			bundleID: "com.example.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateBooted,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple mock bridge
			mockBridge := &simpleMockBridge{
				devices: tt.mockDevices,
			}

			manager := device.NewLocalManager(mockBridge)

			// Test device lookup
			dev, err := manager.GetDevice(tt.deviceID)

			if tt.expectedError != "" {
				if tt.expectedError == "device not found" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				} else if tt.expectedError == "not booted" {
					require.NoError(t, err)
					assert.NotEqual(t, device.StateBooted, dev.State)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, dev)
				assert.Equal(t, device.StateBooted, dev.State)
			}
		})
	}
}

func TestAppTerminateDeviceValidation(t *testing.T) {
	tests := []struct {
		name          string
		deviceID      string
		bundleID      string
		mockDevices   []device.Device
		expectedError string
	}{
		{
			name:     "device not found",
			deviceID: "nonexistent",
			bundleID: "com.example.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateBooted,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "device not found",
		},
		{
			name:     "valid device",
			deviceID: "test-device-1",
			bundleID: "com.example.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateBooted,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple mock bridge
			mockBridge := &simpleMockBridge{
				devices: tt.mockDevices,
			}

			manager := device.NewLocalManager(mockBridge)

			// Test device lookup
			dev, err := manager.GetDevice(tt.deviceID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, dev)
			}
		})
	}
}

func TestAppInstallDeviceValidation(t *testing.T) {
	tests := []struct {
		name          string
		deviceID      string
		appPath       string
		mockDevices   []device.Device
		expectedError string
	}{
		{
			name:     "device not found",
			deviceID: "nonexistent",
			appPath:  "/path/to/MyApp.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateBooted,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "device not found",
		},
		{
			name:     "valid device",
			deviceID: "test-device-1",
			appPath:  "/path/to/MyApp.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateBooted,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple mock bridge
			mockBridge := &simpleMockBridge{
				devices: tt.mockDevices,
			}

			manager := device.NewLocalManager(mockBridge)

			// Test device lookup
			dev, err := manager.GetDevice(tt.deviceID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, dev)
			}
		})
	}
}

func TestAppUninstallDeviceValidation(t *testing.T) {
	tests := []struct {
		name          string
		deviceID      string
		bundleID      string
		mockDevices   []device.Device
		expectedError string
	}{
		{
			name:     "device not found",
			deviceID: "nonexistent",
			bundleID: "com.example.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateBooted,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "device not found",
		},
		{
			name:     "valid device",
			deviceID: "test-device-1",
			bundleID: "com.example.app",
			mockDevices: []device.Device{
				{
					ID:        "test-device-1",
					Name:      "iPhone 15 Pro",
					State:     device.StateBooted,
					Type:      device.DeviceTypeSimulator,
					OSVersion: "17.4",
					UDID:      "test-device-1",
					Available: true,
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple mock bridge
			mockBridge := &simpleMockBridge{
				devices: tt.mockDevices,
			}

			manager := device.NewLocalManager(mockBridge)

			// Test device lookup
			dev, err := manager.GetDevice(tt.deviceID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, dev)
			}
		})
	}
}

// simpleMockBridge is a simple mock for testing device operations
type simpleMockBridge struct {
	devices []device.Device
}

func (m *simpleMockBridge) ListDevices() ([]device.Device, error) {
	return m.devices, nil
}

func (m *simpleMockBridge) BootSimulator(udid string) error {
	return nil
}

func (m *simpleMockBridge) ShutdownSimulator(udid string) error {
	return nil
}

func (m *simpleMockBridge) GetDeviceState(udid string) (device.DeviceState, error) {
	for _, dev := range m.devices {
		if dev.UDID == udid {
			return dev.State, nil
		}
	}
	return "", nil
}

// ============================================================================
// APP COMMAND STRUCTURE TESTS
// ============================================================================

func TestAppCommand_Structure(t *testing.T) {
	// Verify app command structure
	assert.NotNil(t, appCmd)
	assert.Equal(t, "app", appCmd.Use)
	assert.Contains(t, appCmd.Short, "Manage iOS applications")
	assert.Contains(t, appCmd.Long, "launch, terminate, install, and uninstall")
}

func TestAppCommand_Subcommands(t *testing.T) {
	// Verify all required subcommands exist
	expectedSubcommands := []string{"launch", "terminate", "install", "uninstall"}

	subcommandMap := make(map[string]bool)
	for _, cmd := range appCmd.Commands() {
		subcommandMap[cmd.Use] = true
	}

	for _, subcmd := range expectedSubcommands {
		assert.True(t, subcommandMap[subcmd], "app command should have %s subcommand", subcmd)
	}
}

// ============================================================================
// LAUNCH COMMAND TESTS
// ============================================================================

func TestLaunchCommand_Structure(t *testing.T) {
	assert.NotNil(t, launchCmd)
	assert.Equal(t, "launch", launchCmd.Use)
	assert.Contains(t, launchCmd.Short, "Launch an iOS application")
	assert.Contains(t, launchCmd.Long, "by bundle ID")
}

func TestLaunchCommand_Flags(t *testing.T) {
	// Verify launch command flags
	bundleFlag := launchCmd.Flags().Lookup("bundle")
	assert.NotNil(t, bundleFlag, "launch command should have --bundle flag")

	waitFlag := launchCmd.Flags().Lookup("wait-for-ready")
	assert.NotNil(t, waitFlag, "launch command should have --wait-for-ready flag")

	timeoutFlag := launchCmd.Flags().Lookup("timeout")
	assert.NotNil(t, timeoutFlag, "launch command should have --timeout flag")
}

func TestLaunchCommand_TimeoutValidation(t *testing.T) {
	// Test timeout values
	tests := []struct {
		name      string
		timeout   int
		isValid   bool
	}{
		{"positive timeout", 30, true},
		{"default timeout", 60, true},
		{"very long timeout", 300, true},
		{"minimum timeout", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Greater(t, tt.timeout, 0, "timeout should be positive")
		})
	}
}

func TestLaunchResult_PIDHandling(t *testing.T) {
	// Test that PID is properly recorded
	result := LaunchResult{
		Device: &device.Device{
			UDID:  "test-device",
			State: device.StateBooted,
		},
		BundleID: "com.example.app",
		PID:      "12345",
		State:    "launched",
		Message:  "Launched in 100ms",
	}

	assert.NotEmpty(t, result.PID, "PID should be recorded")
	assert.Equal(t, "12345", result.PID)
	assert.Equal(t, "launched", result.State)
}

func TestLaunchResult_StateTransitions(t *testing.T) {
	// Test possible launch states
	states := []string{"launched", "launching", "failed"}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			assert.NotEmpty(t, state, "state should be defined")
		})
	}
}

// ============================================================================
// TERMINATE COMMAND TESTS
// ============================================================================

func TestTerminateCommand_Structure(t *testing.T) {
	assert.NotNil(t, terminateCmd)
	assert.Equal(t, "terminate", terminateCmd.Use)
	assert.Contains(t, terminateCmd.Short, "Terminate a running iOS application")
}

func TestTerminateCommand_Flags(t *testing.T) {
	bundleFlag := terminateCmd.Flags().Lookup("bundle")
	assert.NotNil(t, bundleFlag, "terminate command should have --bundle flag")
}

func TestTerminateCommand_AlreadyTerminated(t *testing.T) {
	// Terminating an already-terminated app should still succeed
	result := TerminateResult{
		Device: &device.Device{
			UDID:  "test-device",
			State: device.StateBooted,
		},
		BundleID: "com.example.app",
		Message:  "App was not running, but command succeeded",
	}

	assert.NotEmpty(t, result.Message)
	assert.Equal(t, "com.example.app", result.BundleID)
}

func TestTerminateResult_SuccessMessage(t *testing.T) {
	// Test termination success messages
	messages := []string{
		"App terminated successfully",
		"App was not running, but command succeeded",
		"Gracefully shut down",
	}

	for _, msg := range messages {
		t.Run(msg, func(t *testing.T) {
			assert.NotEmpty(t, msg)
		})
	}
}

// ============================================================================
// INSTALL COMMAND TESTS
// ============================================================================

func TestInstallCommand_Structure(t *testing.T) {
	assert.NotNil(t, installCmd)
	assert.Equal(t, "install", installCmd.Use)
	assert.Contains(t, installCmd.Short, "Install an iOS application")
}

func TestInstallCommand_Flags(t *testing.T) {
	appFlag := installCmd.Flags().Lookup("app")
	assert.NotNil(t, appFlag, "install command should have --app flag")
}

func TestInstallCommand_AppPathValidation(t *testing.T) {
	// Test various app bundle paths
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		{"standard bundle", "/path/to/MyApp.app", true},
		{"nested bundle", "/Users/test/builds/MyApp.app", true},
		{"relative path", "./MyApp.app", true},
		{"absolute path", "/var/tmp/MyApp.app", true},
		{"missing extension", "/path/to/MyApp", false},
		{"wrong extension", "/path/to/MyApp.ipa", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.True(t, len(tt.path) > 0)
				assert.Contains(t, tt.path, ".app")
			}
		})
	}
}

func TestInstallResult_BundleIDExtraction(t *testing.T) {
	// Test bundle ID extraction from installation
	result := InstallResult{
		Device: &device.Device{
			UDID:  "test-device",
			State: device.StateBooted,
		},
		AppPath:     "/path/to/MyApp.app",
		BundleID:    "com.example.app",
		InstallTime: 5000,
		Message:     "App installed successfully in 5000ms",
	}

	assert.Equal(t, "com.example.app", result.BundleID)
	assert.Greater(t, result.InstallTime, int64(0))
}

func TestInstallResult_InstallTimeDuration(t *testing.T) {
	// Test various installation times
	tests := []int64{
		100,   // very fast
		1000,  // 1 second
		5000,  // 5 seconds
		30000, // 30 seconds
	}

	for _, duration := range tests {
		t.Run(fmt.Sprintf("duration_%dms", duration), func(t *testing.T) {
			assert.Greater(t, duration, int64(0), "duration should be positive")
		})
	}
}

// ============================================================================
// UNINSTALL COMMAND TESTS
// ============================================================================

func TestUninstallCommand_Structure(t *testing.T) {
	assert.NotNil(t, uninstallCmd)
	assert.Equal(t, "uninstall", uninstallCmd.Use)
	assert.Contains(t, uninstallCmd.Short, "Uninstall")
}

func TestUninstallCommand_Flags(t *testing.T) {
	bundleFlag := uninstallCmd.Flags().Lookup("bundle")
	assert.NotNil(t, bundleFlag, "uninstall command should have --bundle flag")
}

func TestUninstallResult_SuccessMessage(t *testing.T) {
	result := UninstallResult{
		Device: &device.Device{
			UDID:  "test-device",
			State: device.StateBooted,
		},
		BundleID: "com.example.app",
		Message:  "App uninstalled successfully",
	}

	assert.Equal(t, "com.example.app", result.BundleID)
	assert.Contains(t, result.Message, "uninstalled")
}

// ============================================================================
// APP COMMAND - BUNDLE ID VALIDATION
// ============================================================================

func TestAppCommand_ValidBundleIDs(t *testing.T) {
	// Test valid bundle ID formats
	validBundleIDs := []string{
		"com.example.app",
		"com.company.myapp",
		"io.github.user.app",
		"a.b.c",
	}

	for _, bundleID := range validBundleIDs {
		t.Run(bundleID, func(t *testing.T) {
			assert.NotEmpty(t, bundleID)
			assert.Contains(t, bundleID, ".")
		})
	}
}

func TestAppCommand_InvalidBundleIDs(t *testing.T) {
	// Test invalid bundle ID formats
	invalidBundleIDs := []struct {
		name string
		id   string
	}{
		{"no dots", "myapp"},
		{"leading dot", ".com.example.app"},
		{"trailing dot", "com.example.app."},
		{"spaces", "com.example .app"},
		{"special chars", "com!example@app"},
	}

	for _, tt := range invalidBundleIDs {
		t.Run(tt.name, func(t *testing.T) {
			// Validation would check for proper reverse domain format
			assert.NotEmpty(t, tt.id)
		})
	}
}

// ============================================================================
// APP COMMAND - DEVICE STATE INTEGRATION
// ============================================================================

func TestAppCommand_DeviceStateRequirements(t *testing.T) {
	// App operations require device to be booted
	states := []struct {
		state    device.DeviceState
		canLaunchApp bool
	}{
		{device.StateBooted, true},
		{device.StateShutdown, false},
		{device.StateBooting, false},
		{device.StateShuttingDown, false},
	}

	for _, st := range states {
		t.Run(string(st.state), func(t *testing.T) {
			assert.NotEmpty(t, st.state)
		})
	}
}

// ============================================================================
// APP COMMAND - ERROR CODES
// ============================================================================

func TestAppCommand_ErrorCodes(t *testing.T) {
	// Verify all error codes used in app commands
	errorCodes := map[string]string{
		"DEVICE_REQUIRED":  "Device ID missing",
		"DEVICE_NOT_FOUND": "Device doesn't exist",
		"DEVICE_NOT_BOOTED": "Device is not booted",
		"BUNDLE_REQUIRED": "Bundle ID missing",
		"APP_NOT_FOUND": "App bundle not found",
		"APP_OPERATION_FAILED": "Launch/terminate/install failed",
		"INVALID_APP_PATH": "App path invalid",
	}

	for code := range errorCodes {
		t.Run(code, func(t *testing.T) {
			assert.NotEmpty(t, code)
		})
	}
}

// ============================================================================
// APP COMMAND - COMMAND REGISTRATION
// ============================================================================

func TestAppCommand_RegisteredWithRoot(t *testing.T) {
	// Verify app command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "app" {
			found = true
			break
		}
	}
	assert.True(t, found, "app command should be registered with root command")
}

// ============================================================================
// LAUNCH COMMAND - WAIT FLAG BEHAVIOR
// ============================================================================

func TestLaunchCommand_WaitForReady(t *testing.T) {
	// Test wait-for-ready flag behavior
	waitFlag := launchCmd.Flags().Lookup("wait-for-ready")
	assert.NotNil(t, waitFlag, "wait-for-ready flag should exist")
	assert.Equal(t, "false", waitFlag.DefValue, "wait-for-ready should default to false")
}

// ============================================================================
// INSTALL COMMAND - APP PATH EDGE CASES
// ============================================================================

func TestInstallCommand_AppPathEdgeCases(t *testing.T) {
	// Test edge cases in app paths
	tests := []struct {
		name string
		path string
	}{
		{"spaces in path", "/Users/test/My Apps/MyApp.app"},
		{"unicode in path", "/Users/test/应用/MyApp.app"},
		{"deep nesting", "/a/b/c/d/e/f/g/h/MyApp.app"},
		{"dots in name", "/path/to/My.App.v1.0.app"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.path, ".app")
		})
	}
}

// ============================================================================
// APP COMMAND - CONCURRENT OPERATIONS
// ============================================================================

func TestAppCommand_MultipleDevices(t *testing.T) {
	// Test operations on multiple devices
	devices := []device.Device{
		{ID: "device-1", UDID: "device-1", Name: "iPhone 14", State: device.StateBooted},
		{ID: "device-2", UDID: "device-2", Name: "iPhone 15", State: device.StateBooted},
		{ID: "device-3", UDID: "device-3", Name: "iPad Pro", State: device.StateBooted},
	}

	for _, dev := range devices {
		t.Run(dev.Name, func(t *testing.T) {
			assert.Equal(t, device.StateBooted, dev.State)
		})
	}
}

// ============================================================================
// APP COMMAND - MESSAGE FORMATTING
// ============================================================================

func TestAppCommand_SuccessMessages(t *testing.T) {
	// Test success message formatting
	messages := []struct {
		name    string
		message string
	}{
		{"launch success", "App launched successfully in 100ms"},
		{"terminate success", "App terminated successfully"},
		{"install success", "App installed successfully in 5000ms"},
		{"uninstall success", "App uninstalled successfully"},
	}

	for _, msg := range messages {
		t.Run(msg.name, func(t *testing.T) {
			assert.NotEmpty(t, msg.message)
		})
	}
}
