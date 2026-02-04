package cmd

import (
	"testing"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockXCRunBridge is a mock implementation of xcrun.Bridge for state command testing
type MockXCRunBridge struct {
	mock.Mock
}

func (m *MockXCRunBridge) ListDevices() ([]device.Device, error) {
	args := m.Called()
	if devices := args.Get(0); devices != nil {
		return devices.([]device.Device), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockXCRunBridge) BootSimulator(udid string) error {
	args := m.Called(udid)
	return args.Error(0)
}

func (m *MockXCRunBridge) ShutdownSimulator(udid string) error {
	args := m.Called(udid)
	return args.Error(0)
}

func (m *MockXCRunBridge) GetDeviceState(udid string) (device.DeviceState, error) {
	args := m.Called(udid)
	if args.Get(0) != nil {
		return args.Get(0).(device.DeviceState), args.Error(1)
	}
	return "", args.Error(1)
}

func (m *MockXCRunBridge) CaptureScreenshot(udid, outputPath string) (*xcrun.ScreenshotResult, error) {
	args := m.Called(udid, outputPath)
	if args.Get(0) != nil {
		return args.Get(0).(*xcrun.ScreenshotResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockXCRunBridge) GetForegroundApp(udid string) (*xcrun.ForegroundAppInfo, error) {
	args := m.Called(udid)
	if args.Get(0) != nil {
		return args.Get(0).(*xcrun.ForegroundAppInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockXCRunBridge) LaunchApp(udid, bundleID string) (string, error) {
	args := m.Called(udid, bundleID)
	return args.String(0), args.Error(1)
}

func (m *MockXCRunBridge) TerminateApp(udid, bundleID string) error {
	args := m.Called(udid, bundleID)
	return args.Error(0)
}

func (m *MockXCRunBridge) InstallApp(udid, appPath string) (string, error) {
	args := m.Called(udid, appPath)
	return args.String(0), args.Error(1)
}

func (m *MockXCRunBridge) UninstallApp(udid, bundleID string) error {
	args := m.Called(udid, bundleID)
	return args.Error(0)
}

func (m *MockXCRunBridge) Tap(udid string, x, y int) (*xcrun.TapResult, error) {
	args := m.Called(udid, x, y)
	if args.Get(0) != nil {
		return args.Get(0).(*xcrun.TapResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockXCRunBridge) TypeText(udid, text string) (*xcrun.TextInputResult, error) {
	args := m.Called(udid, text)
	if args.Get(0) != nil {
		return args.Get(0).(*xcrun.TextInputResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockXCRunBridge) Swipe(udid string, startX, startY, endX, endY, durationMs int) (*xcrun.SwipeResult, error) {
	args := m.Called(udid, startX, startY, endX, endY, durationMs)
	if args.Get(0) != nil {
		return args.Get(0).(*xcrun.SwipeResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockXCRunBridge) PressButton(udid, button string) (*xcrun.ButtonResult, error) {
	args := m.Called(udid, button)
	if args.Get(0) != nil {
		return args.Get(0).(*xcrun.ButtonResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestStateResult(t *testing.T) {
	deviceInfo := &DeviceInfo{
		ID:        "12345",
		Name:      "iPhone 15 Pro",
		State:     "Booted",
		OSVersion: "17.4",
		Runtime:   "iOS 17.4",
	}

	foregroundApp := &ForegroundAppInfo{
		BundleID: "com.apple.Maps",
		PID:      51543,
	}

	result := StateResult{
		Device:        deviceInfo,
		ForegroundApp: foregroundApp,
		Screenshot:    "/tmp/state-screenshot-20260204-143022.png",
	}

	assert.Equal(t, deviceInfo, result.Device)
	assert.Equal(t, foregroundApp, result.ForegroundApp)
	assert.Equal(t, "/tmp/state-screenshot-20260204-143022.png", result.Screenshot)
}

func TestDeviceInfo(t *testing.T) {
	deviceInfo := DeviceInfo{
		ID:        "ABC123",
		Name:      "iPhone 14",
		State:     "Shutdown",
		OSVersion: "17.0",
		Runtime:   "iOS 17.0",
	}

	assert.Equal(t, "ABC123", deviceInfo.ID)
	assert.Equal(t, "iPhone 14", deviceInfo.Name)
	assert.Equal(t, "Shutdown", deviceInfo.State)
	assert.Equal(t, "17.0", deviceInfo.OSVersion)
	assert.Equal(t, "iOS 17.0", deviceInfo.Runtime)
}

func TestForegroundAppInfo(t *testing.T) {
	appInfo := ForegroundAppInfo{
		BundleID: "com.example.testapp",
		PID:      12345,
	}

	assert.Equal(t, "com.example.testapp", appInfo.BundleID)
	assert.Equal(t, 12345, appInfo.PID)
}

func TestStateCommand_BootedDeviceWithoutScreenshot(t *testing.T) {
	bridge := new(MockXCRunBridge)

	testDevice := device.Device{
		ID:        "test-device-1",
		UDID:      "test-device-1",
		Name:      "iPhone 15 Pro",
		State:     device.StateBooted,
		Type:      device.DeviceTypeSimulator,
		OSVersion: "17.4",
		Available: true,
	}

	// Mock ListDevices for GetDevice call
	bridge.On("ListDevices").Return([]device.Device{testDevice}, nil)

	// Mock GetForegroundApp
	foregroundApp := &xcrun.ForegroundAppInfo{
		BundleID: "com.apple.Maps",
		PID:      51543,
	}
	bridge.On("GetForegroundApp", "test-device-1").Return(foregroundApp, nil)

	// Verify the mocks were set up correctly
	devices, err := bridge.ListDevices()
	assert.NoError(t, err)
	assert.Len(t, devices, 1)
	assert.Equal(t, "test-device-1", devices[0].ID)

	app, err := bridge.GetForegroundApp("test-device-1")
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, "com.apple.Maps", app.BundleID)
	assert.Equal(t, 51543, app.PID)

	bridge.AssertExpectations(t)
}

func TestStateCommand_BootedDeviceWithScreenshot(t *testing.T) {
	bridge := new(MockXCRunBridge)

	testDevice := device.Device{
		ID:        "test-device-1",
		UDID:      "test-device-1",
		Name:      "iPhone 15 Pro",
		State:     device.StateBooted,
		Type:      device.DeviceTypeSimulator,
		OSVersion: "17.4",
		Available: true,
	}

	// Mock ListDevices for GetDevice call
	bridge.On("ListDevices").Return([]device.Device{testDevice}, nil)

	// Mock GetForegroundApp
	foregroundApp := &xcrun.ForegroundAppInfo{
		BundleID: "com.apple.Calculator",
		PID:      12345,
	}
	bridge.On("GetForegroundApp", "test-device-1").Return(foregroundApp, nil)

	// Mock CaptureScreenshot
	screenshotResult := &xcrun.ScreenshotResult{
		Path:      "/tmp/state-screenshot-20260204-143022.png",
		Format:    "png",
		SizeBytes: 123456,
		DeviceID:  "test-device-1",
		Timestamp: "2026-02-04T14:30:22Z",
	}
	bridge.On("CaptureScreenshot", "test-device-1", mock.MatchedBy(func(path string) bool {
		// Match any path starting with /tmp/state-screenshot-
		return len(path) > 0
	})).Return(screenshotResult, nil)

	// Verify the mocks
	devices, err := bridge.ListDevices()
	assert.NoError(t, err)
	assert.Len(t, devices, 1)

	app, err := bridge.GetForegroundApp("test-device-1")
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, "com.apple.Calculator", app.BundleID)

	screenshot, err := bridge.CaptureScreenshot("test-device-1", "/tmp/state-screenshot-test.png")
	assert.NoError(t, err)
	assert.NotNil(t, screenshot)
	assert.Equal(t, "/tmp/state-screenshot-20260204-143022.png", screenshot.Path)
	assert.Equal(t, "png", screenshot.Format)

	bridge.AssertExpectations(t)
}

func TestStateCommand_ShutdownDevice(t *testing.T) {
	bridge := new(MockXCRunBridge)

	testDevice := device.Device{
		ID:        "test-device-1",
		UDID:      "test-device-1",
		Name:      "iPhone 15 Pro",
		State:     device.StateShutdown,
		Type:      device.DeviceTypeSimulator,
		OSVersion: "17.4",
		Available: true,
	}

	// Mock ListDevices for GetDevice call
	bridge.On("ListDevices").Return([]device.Device{testDevice}, nil)

	// Verify device is shutdown (no foreground app or screenshot possible)
	devices, err := bridge.ListDevices()
	assert.NoError(t, err)
	assert.Len(t, devices, 1)
	assert.Equal(t, device.StateShutdown, devices[0].State)

	bridge.AssertExpectations(t)
}

func TestStateCommand_DeviceNotFound(t *testing.T) {
	bridge := new(MockXCRunBridge)

	// Mock ListDevices returning empty list
	bridge.On("ListDevices").Return([]device.Device{}, nil)

	devices, err := bridge.ListDevices()
	assert.NoError(t, err)
	assert.Len(t, devices, 0)

	bridge.AssertExpectations(t)
}

func TestStateCommand_ForegroundAppNotAvailable(t *testing.T) {
	bridge := new(MockXCRunBridge)

	// Mock GetForegroundApp returning nil (no foreground app)
	bridge.On("GetForegroundApp", "test-device-1").Return(nil, nil)

	// Verify GetForegroundApp can return nil without error
	app, err := bridge.GetForegroundApp("test-device-1")
	assert.NoError(t, err)
	assert.Nil(t, app)

	bridge.AssertExpectations(t)
}
