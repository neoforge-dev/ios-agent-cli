package device

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeviceBridge is a mock implementation of DeviceBridge
type MockDeviceBridge struct {
	mock.Mock
}

func (m *MockDeviceBridge) ListDevices() ([]Device, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Device), args.Error(1)
}

func (m *MockDeviceBridge) BootSimulator(udid string) error {
	args := m.Called(udid)
	return args.Error(0)
}

func (m *MockDeviceBridge) ShutdownSimulator(udid string) error {
	args := m.Called(udid)
	return args.Error(0)
}

func (m *MockDeviceBridge) GetDeviceState(udid string) (DeviceState, error) {
	args := m.Called(udid)
	return args.Get(0).(DeviceState), args.Error(1)
}

// Test fixtures
var testDevices = []Device{
	{
		ID:        "12345678-1234-1234-1234-123456789ABC",
		Name:      "iPhone 14 Pro",
		State:     StateShutdown,
		Type:      DeviceTypeSimulator,
		OSVersion: "17.4",
		UDID:      "12345678-1234-1234-1234-123456789ABC",
		Available: true,
	},
	{
		ID:        "87654321-4321-4321-4321-CBA987654321",
		Name:      "iPhone 15",
		State:     StateBooted,
		Type:      DeviceTypeSimulator,
		OSVersion: "17.4",
		UDID:      "87654321-4321-4321-4321-CBA987654321",
		Available: true,
	},
	{
		ID:        "ABCDEF12-3456-7890-ABCD-EF1234567890",
		Name:      "iPad Pro",
		State:     StateShutdown,
		Type:      DeviceTypeSimulator,
		OSVersion: "17.2",
		UDID:      "ABCDEF12-3456-7890-ABCD-EF1234567890",
		Available: true,
	},
}

func TestLocalManager_ListDevices(t *testing.T) {
	tests := []struct {
		name        string
		mockDevices []Device
		mockError   error
		wantErr     bool
		wantCount   int
	}{
		{
			name:        "successful device list",
			mockDevices: testDevices,
			mockError:   nil,
			wantErr:     false,
			wantCount:   3,
		},
		{
			name:        "empty device list",
			mockDevices: []Device{},
			mockError:   nil,
			wantErr:     false,
			wantCount:   0,
		},
		{
			name:        "bridge error",
			mockDevices: nil,
			mockError:   errors.New("xcrun failed"),
			wantErr:     true,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockBridge := new(MockDeviceBridge)
			mockBridge.On("ListDevices").Return(tt.mockDevices, tt.mockError)

			// Create manager
			manager := NewLocalManager(mockBridge)

			// Execute
			devices, err := manager.ListDevices()

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, devices)
			} else {
				assert.NoError(t, err)
				assert.Len(t, devices, tt.wantCount)
			}

			mockBridge.AssertExpectations(t)
		})
	}
}

func TestLocalManager_GetDevice(t *testing.T) {
	tests := []struct {
		name       string
		deviceID   string
		wantDevice *Device
		wantErr    bool
	}{
		{
			name:       "find device by ID",
			deviceID:   "12345678-1234-1234-1234-123456789ABC",
			wantDevice: &testDevices[0],
			wantErr:    false,
		},
		{
			name:       "find device by UDID",
			deviceID:   "87654321-4321-4321-4321-CBA987654321",
			wantDevice: &testDevices[1],
			wantErr:    false,
		},
		{
			name:       "device not found",
			deviceID:   "nonexistent-id",
			wantDevice: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockBridge := new(MockDeviceBridge)
			mockBridge.On("ListDevices").Return(testDevices, nil)

			// Create manager
			manager := NewLocalManager(mockBridge)

			// Execute
			device, err := manager.GetDevice(tt.deviceID)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, device)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, device)
				assert.Equal(t, tt.wantDevice.ID, device.ID)
				assert.Equal(t, tt.wantDevice.Name, device.Name)
			}

			mockBridge.AssertExpectations(t)
		})
	}
}

func TestLocalManager_FindDeviceByName(t *testing.T) {
	tests := []struct {
		name       string
		deviceName string
		wantDevice *Device
		wantErr    bool
	}{
		{
			name:       "find device by exact name",
			deviceName: "iPhone 14 Pro",
			wantDevice: &testDevices[0],
			wantErr:    false,
		},
		{
			name:       "find iPad by name",
			deviceName: "iPad Pro",
			wantDevice: &testDevices[2],
			wantErr:    false,
		},
		{
			name:       "device name not found",
			deviceName: "iPhone 99",
			wantDevice: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockBridge := new(MockDeviceBridge)
			mockBridge.On("ListDevices").Return(testDevices, nil)

			// Create manager
			manager := NewLocalManager(mockBridge)

			// Execute
			device, err := manager.FindDeviceByName(tt.deviceName)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, device)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, device)
				assert.Equal(t, tt.wantDevice.Name, device.Name)
			}

			mockBridge.AssertExpectations(t)
		})
	}
}

func TestLocalManager_BootSimulator(t *testing.T) {
	tests := []struct {
		name      string
		deviceID  string
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "boot shutdown device",
			deviceID: "12345678-1234-1234-1234-123456789ABC",
			wantErr:  false,
		},
		{
			name:     "boot already booted device",
			deviceID: "87654321-4321-4321-4321-CBA987654321",
			wantErr:  true,
			errMsg:   "device already booted",
		},
		{
			name:     "boot nonexistent device",
			deviceID: "nonexistent",
			wantErr:  true,
			errMsg:   "device not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockBridge := new(MockDeviceBridge)
			mockBridge.On("ListDevices").Return(testDevices, nil)

			// Only expect BootSimulator call if device exists and not already booted
			if tt.name == "boot shutdown device" {
				mockBridge.On("BootSimulator", "12345678-1234-1234-1234-123456789ABC").Return(nil)
			}

			// Create manager
			manager := NewLocalManager(mockBridge)

			// Execute
			err := manager.BootSimulator(tt.deviceID)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockBridge.AssertExpectations(t)
		})
	}
}

func TestLocalManager_ShutdownSimulator(t *testing.T) {
	tests := []struct {
		name      string
		deviceID  string
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "shutdown booted device",
			deviceID: "87654321-4321-4321-4321-CBA987654321",
			wantErr:  false,
		},
		{
			name:     "shutdown already shutdown device",
			deviceID: "12345678-1234-1234-1234-123456789ABC",
			wantErr:  true,
			errMsg:   "device already shutdown",
		},
		{
			name:     "shutdown nonexistent device",
			deviceID: "nonexistent",
			wantErr:  true,
			errMsg:   "device not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockBridge := new(MockDeviceBridge)
			mockBridge.On("ListDevices").Return(testDevices, nil)

			// Only expect ShutdownSimulator call if device exists and booted
			if tt.name == "shutdown booted device" {
				mockBridge.On("ShutdownSimulator", "87654321-4321-4321-4321-CBA987654321").Return(nil)
			}

			// Create manager
			manager := NewLocalManager(mockBridge)

			// Execute
			err := manager.ShutdownSimulator(tt.deviceID)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockBridge.AssertExpectations(t)
		})
	}
}

func TestLocalManager_GetDeviceState(t *testing.T) {
	tests := []struct {
		name       string
		deviceID   string
		wantState  DeviceState
		wantErr    bool
	}{
		{
			name:      "get state of booted device",
			deviceID:  "87654321-4321-4321-4321-CBA987654321",
			wantState: StateBooted,
			wantErr:   false,
		},
		{
			name:      "get state of shutdown device",
			deviceID:  "12345678-1234-1234-1234-123456789ABC",
			wantState: StateShutdown,
			wantErr:   false,
		},
		{
			name:      "get state of nonexistent device",
			deviceID:  "nonexistent",
			wantState: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockBridge := new(MockDeviceBridge)
			mockBridge.On("ListDevices").Return(testDevices, nil)

			// Only expect GetDeviceState call if device exists
			if !tt.wantErr {
				mockBridge.On("GetDeviceState", mock.Anything).Return(tt.wantState, nil)
			}

			// Create manager
			manager := NewLocalManager(mockBridge)

			// Execute
			state, err := manager.GetDeviceState(tt.deviceID)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantState, state)
			}

			mockBridge.AssertExpectations(t)
		})
	}
}
