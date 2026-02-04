package cmd

import (
	"encoding/json"
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
