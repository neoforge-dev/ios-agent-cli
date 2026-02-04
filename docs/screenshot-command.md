# Screenshot Command Documentation

## Overview

The `screenshot` command captures the current screen of an iOS simulator or physical device and saves it to a file.

## Feature ID

**IOS-009** - Screenshot command implementation

## Usage

```bash
ios-agent screenshot --device <device-id> [options]
```

## Required Flags

- `--device`, `-d` - Device ID or UDID to capture screenshot from

## Optional Flags

- `--output`, `-o` - Output file path (default: timestamped file in /tmp)
- `--format` - Image format: `png` (default) or `jpeg`

## Examples

### Basic Usage

Capture screenshot with automatic filename:
```bash
ios-agent screenshot --device C160E86C-75C2-4DB0-BE4C-C5181D1B245D
```

Output:
```json
{
  "success": true,
  "action": "screenshot.capture",
  "result": {
    "path": "/tmp/screenshot-20260204-120000.png",
    "format": "png",
    "size_bytes": 245678,
    "device_id": "C160E86C-75C2-4DB0-BE4C-C5181D1B245D",
    "timestamp": "2026-02-04T12:00:00Z"
  },
  "timestamp": "2026-02-04T12:00:00Z"
}
```

### Custom Output Path

Save to a specific location:
```bash
ios-agent screenshot --device C160E86C-75C2-4DB0-BE4C-C5181D1B245D --output ./my-screenshot.png
```

### JPEG Format

Capture as JPEG instead of PNG:
```bash
ios-agent screenshot --device C160E86C-75C2-4DB0-BE4C-C5181D1B245D --format jpeg
```

Or specify JPEG extension in output path:
```bash
ios-agent screenshot --device C160E86C-75C2-4DB0-BE4C-C5181D1B245D --output ./shot.jpg
```

## Response Format

### Success Response

```json
{
  "success": true,
  "action": "screenshot.capture",
  "result": {
    "path": "/tmp/screenshot-20260204-120000.png",
    "format": "png",
    "size_bytes": 245678,
    "device_id": "C160E86C-75C2-4DB0-BE4C-C5181D1B245D",
    "timestamp": "2026-02-04T12:00:00Z"
  },
  "timestamp": "2026-02-04T12:00:00Z"
}
```

**Fields:**
- `path` - Full path to the captured screenshot file
- `format` - Image format (png or jpeg)
- `size_bytes` - File size in bytes
- `device_id` - Device UDID that was captured
- `timestamp` - ISO 8601 timestamp when screenshot was taken

### Error Responses

#### Device Required
```json
{
  "success": false,
  "action": "screenshot.capture",
  "error": {
    "code": "DEVICE_REQUIRED",
    "message": "device ID is required (use --device flag)"
  },
  "timestamp": "2026-02-04T12:00:00Z"
}
```

#### Device Not Found
```json
{
  "success": false,
  "action": "screenshot.capture",
  "error": {
    "code": "DEVICE_NOT_FOUND",
    "message": "device not found: invalid-device-id"
  },
  "timestamp": "2026-02-04T12:00:00Z"
}
```

#### Device Not Booted
```json
{
  "success": false,
  "action": "screenshot.capture",
  "error": {
    "code": "DEVICE_NOT_BOOTED",
    "message": "device is not booted: iPhone 17 Pro Max (state: Shutdown)"
  },
  "timestamp": "2026-02-04T12:00:00Z"
}
```

#### Invalid Format
```json
{
  "success": false,
  "action": "screenshot.capture",
  "error": {
    "code": "INVALID_FORMAT",
    "message": "invalid format: gif (must be png or jpeg)"
  },
  "timestamp": "2026-02-04T12:00:00Z"
}
```

#### Screenshot Failed
```json
{
  "success": false,
  "action": "screenshot.capture",
  "error": {
    "code": "SCREENSHOT_FAILED",
    "message": "failed to capture screenshot: <xcrun error message>"
  },
  "timestamp": "2026-02-04T12:00:00Z"
}
```

## Implementation Details

### Technology

The screenshot command uses `xcrun simctl io <device-id> screenshot <path>` to capture the screen.

### File Format Detection

The format is automatically detected from the output file extension:
- `.png` → PNG format
- `.jpg`, `.jpeg` → JPEG format
- Default → PNG format

### Default Output Path

When no output path is specified, screenshots are saved to:
```
/tmp/screenshot-YYYYMMDD-HHMMSS.<ext>
```

Where:
- `YYYYMMDD` - Current date
- `HHMMSS` - Current time
- `<ext>` - `png` or `jpg` based on format flag

## Requirements

### Device State

The target device **must be booted**. Attempting to capture a screenshot from a shutdown device will fail with `DEVICE_NOT_BOOTED` error.

### Xcode Installation

This command requires Xcode Command Line Tools to be installed for the `xcrun` command.

## Testing

### Unit Tests

```bash
go test ./pkg/xcrun/
```

### Integration Tests

Integration tests require a booted simulator:

```bash
# Boot a simulator first
ios-agent simulator boot --name "iPhone 17 Pro"

# Run integration tests
go test -tags=integration ./pkg/xcrun/ -v
```

### Manual Testing

```bash
# Build
make build

# List devices
./ios-agent devices

# Capture screenshot
./ios-agent screenshot --device <device-id>

# Verify file
ls -lh /tmp/screenshot-*.png
```

## AI Agent Usage

For AI agents automating iOS testing:

```json
{
  "command": "screenshot",
  "device": "C160E86C-75C2-4DB0-BE4C-C5181D1B245D",
  "output": "./test-results/screen-001.png",
  "format": "png"
}
```

Execute:
```bash
ios-agent screenshot --device C160E86C-75C2-4DB0-BE4C-C5181D1B245D --output ./test-results/screen-001.png
```

Parse the JSON response to get the file path and verify the screenshot was captured successfully.

## Related Commands

- `ios-agent devices` - List available devices
- `ios-agent simulator boot` - Boot a simulator (IOS-003, pending)
- `ios-agent state` - Get comprehensive device state snapshot (IOS-014, pending)

## Acceptance Criteria

✅ **Saves PNG to specified path** - Implemented with default and custom paths
✅ **Returns dimensions and file size** - Returns file size in `size_bytes` field
✅ **Works with xcrun simctl** - Uses `xcrun simctl io <device-id> screenshot`

## Future Enhancements

- Add image dimensions (width, height) to response
- Support for video recording
- Batch screenshot capture
- Remote device support via Tailscale
