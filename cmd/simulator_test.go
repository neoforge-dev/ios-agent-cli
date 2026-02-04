package cmd

import (
	"testing"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeviceManager is a mock implementation of device manager for testing
type MockDeviceManager struct {
	mock.Mock
}

func (m *MockDeviceManager) ListDevices() ([]device.Device, error) {
	args := m.Called()
	if devices := args.Get(0); devices != nil {
		return devices.([]device.Device), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDeviceManager) GetDevice(id string) (*device.Device, error) {
	args := m.Called(id)
	if dev := args.Get(0); dev != nil {
		return dev.(*device.Device), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDeviceManager) FindDeviceByName(name string) (*device.Device, error) {
	args := m.Called(name)
	if dev := args.Get(0); dev != nil {
		return dev.(*device.Device), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockLocalManager extends LocalManager for testing
type MockLocalManager struct {
	*device.LocalManager
	mockBridge *MockDeviceBridge
}

// MockDeviceBridge is a mock implementation of DeviceBridge
type MockDeviceBridge struct {
	mock.Mock
	devices     []device.Device
	deviceState map[string]device.DeviceState
}

func NewMockDeviceBridge() *MockDeviceBridge {
	return &MockDeviceBridge{
		deviceState: make(map[string]device.DeviceState),
	}
}

func (m *MockDeviceBridge) ListDevices() ([]device.Device, error) {
	args := m.Called()

	// Check if we have a function generator for devices
	if devicesFunc, ok := args.Get(0).(func() []device.Device); ok {
		return devicesFunc(), args.Error(1)
	}

	// Check if devices are directly provided
	if devices := args.Get(0); devices != nil {
		if devs, ok := devices.([]device.Device); ok {
			return devs, args.Error(1)
		}
	}

	// Fall back to internal devices field
	return m.devices, args.Error(1)
}

func (m *MockDeviceBridge) BootSimulator(udid string) error {
	args := m.Called(udid)
	if args.Error(0) == nil {
		// Update state to booted after a short delay
		m.deviceState[udid] = device.StateBooted
	}
	return args.Error(0)
}

func (m *MockDeviceBridge) ShutdownSimulator(udid string) error {
	args := m.Called(udid)
	if args.Error(0) == nil {
		m.deviceState[udid] = device.StateShutdown
	}
	return args.Error(0)
}

func (m *MockDeviceBridge) GetDeviceState(udid string) (device.DeviceState, error) {
	args := m.Called(udid)

	// Check if we have a function generator for the state
	if stateFunc, ok := args.Get(0).(func(string) device.DeviceState); ok {
		return stateFunc(udid), args.Error(1)
	}

	// Check if state is directly in deviceState map
	if state, ok := m.deviceState[udid]; ok {
		return state, args.Error(1)
	}

	// Try to get state from mock args
	if args.Get(0) != nil {
		if state, ok := args.Get(0).(device.DeviceState); ok {
			return state, args.Error(1)
		}
	}

	return "", args.Error(1)
}

func TestFindDeviceByNameAndOS(t *testing.T) {
	tests := []struct {
		name        string
		devices     []device.Device
		searchName  string
		osVersion   string
		expectError bool
		expectedID  string
	}{
		{
			name: "find device by name only",
			devices: []device.Device{
				{ID: "dev1", Name: "iPhone 15 Pro", OSVersion: "17.4", State: device.StateShutdown},
				{ID: "dev2", Name: "iPhone 14", OSVersion: "17.0", State: device.StateShutdown},
			},
			searchName:  "iPhone 15 Pro",
			osVersion:   "",
			expectError: false,
			expectedID:  "dev1",
		},
		{
			name: "find device by name and OS version",
			devices: []device.Device{
				{ID: "dev1", Name: "iPhone 15 Pro", OSVersion: "17.4", State: device.StateShutdown},
				{ID: "dev2", Name: "iPhone 15 Pro", OSVersion: "17.5", State: device.StateShutdown},
			},
			searchName:  "iPhone 15 Pro",
			osVersion:   "17.5",
			expectError: false,
			expectedID:  "dev2",
		},
		{
			name: "prefer booted device",
			devices: []device.Device{
				{ID: "dev1", Name: "iPhone 15 Pro", OSVersion: "17.4", State: device.StateShutdown},
				{ID: "dev2", Name: "iPhone 15 Pro", OSVersion: "17.4", State: device.StateBooted},
			},
			searchName:  "iPhone 15 Pro",
			osVersion:   "",
			expectError: false,
			expectedID:  "dev2",
		},
		{
			name: "device not found",
			devices: []device.Device{
				{ID: "dev1", Name: "iPhone 15 Pro", OSVersion: "17.4", State: device.StateShutdown},
			},
			searchName:  "iPhone 14",
			osVersion:   "",
			expectError: true,
		},
		{
			name: "device not found with OS version filter",
			devices: []device.Device{
				{ID: "dev1", Name: "iPhone 15 Pro", OSVersion: "17.4", State: device.StateShutdown},
			},
			searchName:  "iPhone 15 Pro",
			osVersion:   "17.5",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewMockDeviceBridge()
			bridge.devices = tt.devices
			bridge.On("ListDevices").Return(tt.devices, nil)

			manager := device.NewLocalManager(bridge)

			dev, err := findDeviceByNameAndOS(manager, tt.searchName, tt.osVersion)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, dev)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dev)
				assert.Equal(t, tt.expectedID, dev.ID)
			}
		})
	}
}

func TestPollForBootCompletion(t *testing.T) {
	tests := []struct {
		name        string
		deviceID    string
		timeout     int
		stateSeq    []device.DeviceState // Sequence of states returned
		expectError bool
	}{
		{
			name:        "boot completes immediately",
			deviceID:    "dev1",
			timeout:     5,
			stateSeq:    []device.DeviceState{device.StateBooted},
			expectError: false,
		},
		{
			name:        "boot completes after polling",
			deviceID:    "dev1",
			timeout:     5,
			stateSeq:    []device.DeviceState{device.StateBooting, device.StateBooting, device.StateBooted},
			expectError: false,
		},
		{
			name:        "boot times out",
			deviceID:    "dev1",
			timeout:     1, // Short timeout to test timeout behavior
			stateSeq:    []device.DeviceState{device.StateBooting, device.StateBooting, device.StateBooting},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewMockDeviceBridge()

			// Set up mock to return states in sequence
			callCount := 0
			stateFunc := func(udid string) device.DeviceState {
				if callCount < len(tt.stateSeq) {
					state := tt.stateSeq[callCount]
					callCount++
					return state
				}
				// Continue returning the last state
				return tt.stateSeq[len(tt.stateSeq)-1]
			}

			// Use Run to set up the mock with proper return values
			bridge.On("GetDeviceState", tt.deviceID).Run(func(args mock.Arguments) {}).Return(stateFunc, nil)

			// Mock ListDevices - needed for GetDevice calls
			bridge.On("ListDevices").Return([]device.Device{
				{ID: tt.deviceID, UDID: tt.deviceID, Name: "Test Device", State: device.StateBooted},
			}, nil)

			manager := device.NewLocalManager(bridge)

			dev, err := pollForBootCompletion(manager, tt.deviceID, tt.timeout)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, dev)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dev)
				assert.Equal(t, device.StateBooted, dev.State)
			}
		})
	}
}

func TestBootResult(t *testing.T) {
	dev := &device.Device{
		ID:        "12345",
		Name:      "iPhone 15 Pro",
		State:     device.StateBooted,
		Type:      device.DeviceTypeSimulator,
		OSVersion: "17.4",
		UDID:      "12345",
	}

	result := BootResult{
		Device:     dev,
		BootTimeMs: 4523,
	}

	assert.Equal(t, dev, result.Device)
	assert.Equal(t, int64(4523), result.BootTimeMs)
}

func TestShutdownResult(t *testing.T) {
	dev := &device.Device{
		ID:        "12345",
		Name:      "iPhone 15 Pro",
		State:     device.StateShutdown,
		Type:      device.DeviceTypeSimulator,
		OSVersion: "17.4",
		UDID:      "12345",
	}

	result := ShutdownResult{
		Device:  dev,
		Message: "Simulator shutdown successfully",
	}

	assert.Equal(t, dev, result.Device)
	assert.Equal(t, "Simulator shutdown successfully", result.Message)
}

// Integration-like test that exercises the full flow
func TestBootCommandFlow(t *testing.T) {
	bridge := NewMockDeviceBridge()

	testDevice := device.Device{
		ID:        "test-device-1",
		UDID:      "test-device-1",
		Name:      "iPhone 15 Pro",
		State:     device.StateShutdown,
		Type:      device.DeviceTypeSimulator,
		OSVersion: "17.4",
		Available: true,
	}

	bridge.devices = []device.Device{testDevice}
	bridge.deviceState["test-device-1"] = device.StateShutdown

	// Mock ListDevices - will be called multiple times, need to handle state changes
	listCallCount := 0
	listFunc := func() []device.Device {
		// After the first few calls, return booted device
		if listCallCount >= 2 {
			updatedDevice := testDevice
			updatedDevice.State = device.StateBooted
			return []device.Device{updatedDevice}
		}
		listCallCount++
		return []device.Device{testDevice}
	}
	bridge.On("ListDevices").Run(func(args mock.Arguments) {}).Return(listFunc, nil)

	// Mock BootSimulator
	bridge.On("BootSimulator", "test-device-1").Return(nil)

	// Mock GetDeviceState to simulate boot process
	stateCallCount := 0
	stateFunc := func(udid string) device.DeviceState {
		stateCallCount++
		if stateCallCount >= 2 {
			return device.StateBooted
		}
		return device.StateBooting
	}
	bridge.On("GetDeviceState", "test-device-1").Run(func(args mock.Arguments) {}).Return(stateFunc, nil)

	manager := device.NewLocalManager(bridge)

	// Test boot flow
	dev, err := findDeviceByNameAndOS(manager, "iPhone 15 Pro", "17.4")
	assert.NoError(t, err)
	assert.NotNil(t, dev)

	err = manager.BootSimulator(dev.ID)
	assert.NoError(t, err)

	bootedDev, err := pollForBootCompletion(manager, dev.ID, 5)
	assert.NoError(t, err)
	assert.NotNil(t, bootedDev)
	assert.Equal(t, device.StateBooted, bootedDev.State)

	bridge.AssertExpectations(t)
}

func TestShutdownCommandFlow(t *testing.T) {
	bridge := NewMockDeviceBridge()

	testDevice := device.Device{
		ID:        "test-device-1",
		UDID:      "test-device-1",
		Name:      "iPhone 15 Pro",
		State:     device.StateBooted,
		Type:      device.DeviceTypeSimulator,
		OSVersion: "17.4",
		Available: true,
	}

	bridge.devices = []device.Device{testDevice}

	// Mock ListDevices
	bridge.On("ListDevices").Return(bridge.devices, nil)

	// Mock ShutdownSimulator
	bridge.On("ShutdownSimulator", "test-device-1").Return(nil)

	manager := device.NewLocalManager(bridge)

	// Test shutdown flow
	dev, err := manager.GetDevice("test-device-1")
	assert.NoError(t, err)
	assert.NotNil(t, dev)
	assert.Equal(t, device.StateBooted, dev.State)

	err = manager.ShutdownSimulator(dev.ID)
	assert.NoError(t, err)

	bridge.AssertExpectations(t)
}
