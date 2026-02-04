# Error Handling Framework

This package provides standardized error codes and error handling for the ios-agent-cli.

## Features

- **Standardized error codes**: All commands use consistent error codes (e.g., `DEVICE_NOT_FOUND`, `APP_LAUNCH_FAILED`)
- **Consistent JSON format**: All errors follow the same JSON structure
- **Rich error details**: Errors can include contextual information (device ID, bundle ID, etc.)
- **Type-safe error construction**: Helper functions for common error scenarios

## Error Codes

| Code | Description | Usage |
|------|-------------|-------|
| `DEVICE_NOT_FOUND` | Device ID doesn't exist | Device discovery/lookup fails |
| `DEVICE_UNREACHABLE` | Connection failed | Network/remote device issues |
| `DEVICE_NOT_BOOTED` | Device exists but not running | Operations requiring booted device |
| `DEVICE_REQUIRED` | Device flag not provided | Missing required --device flag |
| `APP_NOT_FOUND` | Bundle ID not installed | App not found on device |
| `APP_LAUNCH_FAILED` | Failed to launch app | App launch operation fails |
| `APP_TERMINATE_FAILED` | Failed to terminate app | App termination fails |
| `UI_ACTION_FAILED` | Tap/swipe failed | UI interaction errors |
| `INVALID_COORDINATES` | X/Y coordinates invalid | Invalid tap/swipe coordinates |
| `TEXT_REQUIRED` | Text input empty | Missing required text input |
| `SIMULATOR_TIMEOUT` | Boot/shutdown exceeded timeout | Simulator operations timeout |
| `BOOT_FAILED` | Simulator boot failed | Boot operation error |
| `SHUTDOWN_FAILED` | Simulator shutdown failed | Shutdown operation error |
| `SCREENSHOT_FAILED` | Screenshot capture failed | Screenshot operation error |
| `INVALID_FORMAT` | Invalid format specified | Invalid image format |
| `PATH_ERROR` | File path error | File system operation error |
| `DEVICE_DISCOVERY_FAILED` | Failed to list devices | Device listing error |
| `INTERNAL_ERROR` | Unexpected internal error | Generic internal errors |

## Usage

### Creating Errors

#### Using Constructor Functions (Recommended)

```go
import "github.com/neoforge-dev/ios-agent-cli/pkg/errors"

// Device not found
err := errors.DeviceNotFoundError("ABC123")

// Device not booted
err := errors.DeviceNotBootedError("ABC123", "Shutdown")

// App launch failed
err := errors.AppLaunchFailedError("ABC123", "com.example.app", "app not installed")

// Invalid coordinates
err := errors.InvalidCoordinatesError(-10, -20)

// Simulator timeout
err := errors.SimulatorTimeoutError("ABC123", 60, 62.5)
```

#### Using Generic Constructors

```go
import "github.com/neoforge-dev/ios-agent-cli/pkg/errors"

// Simple error
err := errors.New(errors.DeviceNotFound, "device not found: ABC123")

// Error with details
err := errors.NewWithDetails(
    errors.AppLaunchFailed,
    "failed to launch app",
    map[string]interface{}{
        "device_id": "ABC123",
        "bundle_id": "com.example.app",
    },
)
```

### Using in Commands

#### New Way (with outputAgentError)

```go
import (
    "github.com/neoforge-dev/ios-agent-cli/pkg/errors"
)

func runSomeCommand(cmd *cobra.Command, args []string) {
    // Validate input
    if deviceID == "" {
        outputAgentError("command.action", errors.DeviceRequiredError())
        return
    }

    // Get device
    dev, err := manager.GetDevice(deviceID)
    if err != nil {
        outputAgentError("command.action", errors.DeviceNotFoundError(deviceID))
        return
    }

    // Check device state
    if dev.State != device.StateBooted {
        outputAgentError("command.action", errors.DeviceNotBootedError(dev.ID, string(dev.State)))
        return
    }

    // Success case
    outputSuccess("command.action", result)
}
```

#### Old Way (still supported)

```go
// Still works but deprecated
outputError("command.action", "DEVICE_NOT_FOUND", "device not found", map[string]string{
    "device_id": deviceID,
})
```

## JSON Output Format

All errors follow this structure:

```json
{
  "success": false,
  "action": "command.action",
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {
      "device_id": "ABC123",
      "additional": "context"
    }
  },
  "timestamp": "2026-02-04T19:00:00Z"
}
```

## Migration Guide

### Step 1: Import the errors package

```go
import "github.com/neoforge-dev/ios-agent-cli/pkg/errors"
```

### Step 2: Replace outputError calls with outputAgentError

**Before:**
```go
outputError("app.launch", "DEVICE_NOT_FOUND", err.Error(), map[string]string{
    "device_id": launchDeviceID,
})
```

**After:**
```go
outputAgentError("app.launch", errors.DeviceNotFoundError(launchDeviceID))
```

### Step 3: Use constructor functions for common errors

Replace error code strings with constructor functions from the errors package.

## Benefits

1. **Type Safety**: Error codes are constants, preventing typos
2. **Consistency**: All errors follow the same structure
3. **Discoverability**: IDE autocomplete shows available error types
4. **Documentation**: Error codes are documented in one place
5. **Testing**: Easier to test error handling with standardized errors
6. **Maintainability**: Centralized error definitions

## Future Enhancements

- Error localization support
- Error severity levels
- Error aggregation and logging
- Retry hints for transient errors
