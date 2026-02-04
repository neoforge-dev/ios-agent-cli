package errors

import (
	"testing"
)

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
		want string
	}{
		{"device not found", DeviceNotFound, "DEVICE_NOT_FOUND"},
		{"device unreachable", DeviceUnreachable, "DEVICE_UNREACHABLE"},
		{"device not booted", DeviceNotBooted, "DEVICE_NOT_BOOTED"},
		{"device required", DeviceRequired, "DEVICE_REQUIRED"},
		{"app not found", AppNotFound, "APP_NOT_FOUND"},
		{"app launch failed", AppLaunchFailed, "APP_LAUNCH_FAILED"},
		{"ui action failed", UIActionFailed, "UI_ACTION_FAILED"},
		{"invalid coordinates", InvalidCoordinates, "INVALID_COORDINATES"},
		{"simulator timeout", SimulatorTimeout, "SIMULATOR_TIMEOUT"},
		{"screenshot failed", ScreenshotFailed, "SCREENSHOT_FAILED"},
		{"internal error", InternalError, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.code) != tt.want {
				t.Errorf("ErrorCode = %v, want %v", tt.code, tt.want)
			}
		})
	}
}

func TestAgentError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *AgentError
		wantMsg string
	}{
		{
			name:    "simple error",
			err:     New(DeviceNotFound, "device not found: ABC123"),
			wantMsg: "DEVICE_NOT_FOUND: device not found: ABC123",
		},
		{
			name: "error with details",
			err: NewWithDetails(
				AppLaunchFailed,
				"failed to launch app",
				map[string]interface{}{"device_id": "ABC123", "bundle_id": "com.example.app"},
			),
			wantMsg: "APP_LAUNCH_FAILED: failed to launch app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg && len(tt.err.Details) == 0 {
				t.Errorf("AgentError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestDeviceNotFoundError(t *testing.T) {
	deviceID := "ABC123"
	err := DeviceNotFoundError(deviceID)

	if err.Code != DeviceNotFound {
		t.Errorf("DeviceNotFoundError() code = %v, want %v", err.Code, DeviceNotFound)
	}

	if err.Details["device_id"] != deviceID {
		t.Errorf("DeviceNotFoundError() details.device_id = %v, want %v", err.Details["device_id"], deviceID)
	}
}

func TestDeviceNotBootedError(t *testing.T) {
	deviceID := "ABC123"
	state := "Shutdown"
	err := DeviceNotBootedError(deviceID, state)

	if err.Code != DeviceNotBooted {
		t.Errorf("DeviceNotBootedError() code = %v, want %v", err.Code, DeviceNotBooted)
	}

	if err.Details["device_id"] != deviceID {
		t.Errorf("DeviceNotBootedError() details.device_id = %v, want %v", err.Details["device_id"], deviceID)
	}

	if err.Details["state"] != state {
		t.Errorf("DeviceNotBootedError() details.state = %v, want %v", err.Details["state"], state)
	}
}

func TestAppLaunchFailedError(t *testing.T) {
	deviceID := "ABC123"
	bundleID := "com.example.app"
	reason := "app not installed"
	err := AppLaunchFailedError(deviceID, bundleID, reason)

	if err.Code != AppLaunchFailed {
		t.Errorf("AppLaunchFailedError() code = %v, want %v", err.Code, AppLaunchFailed)
	}

	if err.Details["device_id"] != deviceID {
		t.Errorf("AppLaunchFailedError() details.device_id = %v, want %v", err.Details["device_id"], deviceID)
	}

	if err.Details["bundle_id"] != bundleID {
		t.Errorf("AppLaunchFailedError() details.bundle_id = %v, want %v", err.Details["bundle_id"], bundleID)
	}
}

func TestInvalidCoordinatesError(t *testing.T) {
	x, y := -10, -20
	err := InvalidCoordinatesError(x, y)

	if err.Code != InvalidCoordinates {
		t.Errorf("InvalidCoordinatesError() code = %v, want %v", err.Code, InvalidCoordinates)
	}

	if err.Details["x"] != x {
		t.Errorf("InvalidCoordinatesError() details.x = %v, want %v", err.Details["x"], x)
	}

	if err.Details["y"] != y {
		t.Errorf("InvalidCoordinatesError() details.y = %v, want %v", err.Details["y"], y)
	}
}

func TestSimulatorTimeoutError(t *testing.T) {
	deviceID := "ABC123"
	timeoutSec := 60
	elapsedSec := 62.5
	err := SimulatorTimeoutError(deviceID, timeoutSec, elapsedSec)

	if err.Code != SimulatorTimeout {
		t.Errorf("SimulatorTimeoutError() code = %v, want %v", err.Code, SimulatorTimeout)
	}

	if err.Details["device_id"] != deviceID {
		t.Errorf("SimulatorTimeoutError() details.device_id = %v, want %v", err.Details["device_id"], deviceID)
	}

	if err.Details["timeout_sec"] != timeoutSec {
		t.Errorf("SimulatorTimeoutError() details.timeout_sec = %v, want %v", err.Details["timeout_sec"], timeoutSec)
	}

	if err.Details["elapsed_sec"] != elapsedSec {
		t.Errorf("SimulatorTimeoutError() details.elapsed_sec = %v, want %v", err.Details["elapsed_sec"], elapsedSec)
	}
}
