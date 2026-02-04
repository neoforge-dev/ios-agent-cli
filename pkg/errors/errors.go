package errors

import (
	"fmt"
)

// ErrorCode represents a standardized error code
type ErrorCode string

// Standard error codes for ios-agent-cli
const (
	// Device-related errors
	DeviceNotFound      ErrorCode = "DEVICE_NOT_FOUND"      // Device ID doesn't exist
	DeviceUnreachable   ErrorCode = "DEVICE_UNREACHABLE"    // Connection failed
	DeviceNotBooted     ErrorCode = "DEVICE_NOT_BOOTED"     // Device exists but not running
	DeviceRequired      ErrorCode = "DEVICE_REQUIRED"       // Device flag not provided

	// App-related errors
	AppNotFound         ErrorCode = "APP_NOT_FOUND"         // Bundle ID not installed
	AppLaunchFailed     ErrorCode = "APP_LAUNCH_FAILED"     // Failed to launch app
	AppTerminateFailed  ErrorCode = "APP_TERMINATE_FAILED"  // Failed to terminate app

	// UI interaction errors
	UIActionFailed      ErrorCode = "UI_ACTION_FAILED"      // Tap/swipe failed
	InvalidCoordinates  ErrorCode = "INVALID_COORDINATES"   // X/Y coordinates invalid
	TextRequired        ErrorCode = "TEXT_REQUIRED"         // Text input empty

	// Simulator operation errors
	SimulatorTimeout    ErrorCode = "SIMULATOR_TIMEOUT"     // Boot/shutdown exceeded timeout
	BootFailed          ErrorCode = "BOOT_FAILED"           // Simulator boot operation failed
	ShutdownFailed      ErrorCode = "SHUTDOWN_FAILED"       // Simulator shutdown operation failed

	// Screenshot errors
	ScreenshotFailed    ErrorCode = "SCREENSHOT_FAILED"     // Screenshot capture failed
	InvalidFormat       ErrorCode = "INVALID_FORMAT"        // Invalid image format
	PathError           ErrorCode = "PATH_ERROR"            // File path error

	// Discovery errors
	DeviceDiscoveryFailed ErrorCode = "DEVICE_DISCOVERY_FAILED" // Failed to list devices

	// Generic errors
	InternalError       ErrorCode = "INTERNAL_ERROR"        // Unexpected internal error
)

// AgentError represents a standardized CLI error
type AgentError struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *AgentError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("%s: %s (details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new AgentError
func New(code ErrorCode, message string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewWithDetails creates a new AgentError with details
func NewWithDetails(code ErrorCode, message string, details map[string]interface{}) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WithDetails adds details to an existing error
func (e *AgentError) WithDetails(details map[string]interface{}) *AgentError {
	e.Details = details
	return e
}

// Common error constructors for convenience

// DeviceNotFoundError creates a DEVICE_NOT_FOUND error
func DeviceNotFoundError(deviceID string) *AgentError {
	return NewWithDetails(
		DeviceNotFound,
		fmt.Sprintf("device not found: %s", deviceID),
		map[string]interface{}{"device_id": deviceID},
	)
}

// DeviceNotBootedError creates a DEVICE_NOT_BOOTED error
func DeviceNotBootedError(deviceID, state string) *AgentError {
	return NewWithDetails(
		DeviceNotBooted,
		fmt.Sprintf("device is not booted (state: %s)", state),
		map[string]interface{}{
			"device_id": deviceID,
			"state":     state,
		},
	)
}

// DeviceRequiredError creates a DEVICE_REQUIRED error
func DeviceRequiredError() *AgentError {
	return New(DeviceRequired, "device ID is required (use --device flag)")
}

// AppNotFoundError creates an APP_NOT_FOUND error
func AppNotFoundError(bundleID string) *AgentError {
	return NewWithDetails(
		AppNotFound,
		fmt.Sprintf("app not found: %s", bundleID),
		map[string]interface{}{"bundle_id": bundleID},
	)
}

// AppLaunchFailedError creates an APP_LAUNCH_FAILED error
func AppLaunchFailedError(deviceID, bundleID, reason string) *AgentError {
	return NewWithDetails(
		AppLaunchFailed,
		fmt.Sprintf("failed to launch app: %s", reason),
		map[string]interface{}{
			"device_id": deviceID,
			"bundle_id": bundleID,
		},
	)
}

// AppTerminateFailedError creates an APP_TERMINATE_FAILED error
func AppTerminateFailedError(deviceID, bundleID, reason string) *AgentError {
	return NewWithDetails(
		AppTerminateFailed,
		fmt.Sprintf("failed to terminate app: %s", reason),
		map[string]interface{}{
			"device_id": deviceID,
			"bundle_id": bundleID,
		},
	)
}

// InvalidCoordinatesError creates an INVALID_COORDINATES error
func InvalidCoordinatesError(x, y int) *AgentError {
	return NewWithDetails(
		InvalidCoordinates,
		fmt.Sprintf("invalid coordinates: x=%d, y=%d", x, y),
		map[string]interface{}{
			"x": x,
			"y": y,
		},
	)
}

// TextRequiredError creates a TEXT_REQUIRED error
func TextRequiredError() *AgentError {
	return New(TextRequired, "text input cannot be empty")
}

// SimulatorTimeoutError creates a SIMULATOR_TIMEOUT error
func SimulatorTimeoutError(deviceID string, timeoutSec int, elapsedSec float64) *AgentError {
	return NewWithDetails(
		SimulatorTimeout,
		fmt.Sprintf("simulator operation timed out after %d seconds", timeoutSec),
		map[string]interface{}{
			"device_id":   deviceID,
			"timeout_sec": timeoutSec,
			"elapsed_sec": elapsedSec,
		},
	)
}

// ScreenshotFailedError creates a SCREENSHOT_FAILED error
func ScreenshotFailedError(reason string) *AgentError {
	return New(ScreenshotFailed, fmt.Sprintf("screenshot capture failed: %s", reason))
}

// InternalErrorFromErr wraps a Go error as an INTERNAL_ERROR
func InternalErrorFromErr(err error) *AgentError {
	return New(InternalError, err.Error())
}
