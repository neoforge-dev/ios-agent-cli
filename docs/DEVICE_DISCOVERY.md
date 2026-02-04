# Device Discovery Documentation

## Overview

The `ios-agent devices` command discovers and lists all available iOS simulators on the local machine using `xcrun simctl`.

## Command

```bash
ios-agent devices
```

## Output Format

Returns a JSON response with the following structure:

```json
{
  "success": true,
  "action": "devices.list",
  "result": {
    "devices": [
      {
        "id": "DEVICE-UUID",
        "name": "Device Name",
        "state": "Booted|Shutdown|...",
        "type": "simulator",
        "os_version": "X.Y",
        "udid": "DEVICE-UUID",
        "available": true
      }
    ]
  },
  "timestamp": "2026-02-04T12:00:00Z"
}
```

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Device unique identifier (same as UDID) |
| `name` | string | Human-readable device name (e.g., "iPhone 14 Pro") |
| `state` | string | Current device state: `Booted`, `Shutdown`, `Creating`, `Booting`, `ShuttingDown` |
| `type` | string | Device type (currently only `simulator`) |
| `os_version` | string | iOS version (e.g., "17.4") |
| `udid` | string | Universal Device Identifier |
| `available` | boolean | Whether the device is available for use |

## Device States

| State | Description |
|-------|-------------|
| `Booted` | Device is running and ready for interaction |
| `Shutdown` | Device is powered off |
| `Creating` | Device is being created (rare) |
| `Booting` | Device is in the process of booting |
| `ShuttingDown` | Device is in the process of shutting down |

## Example Output

### Multiple Devices

```json
{
  "success": true,
  "action": "devices.list",
  "result": {
    "devices": [
      {
        "id": "C160E86C-75C2-4DB0-BE4C-C5181D1B245D",
        "name": "iPhone 17 Pro",
        "state": "Booted",
        "type": "simulator",
        "os_version": "26.2",
        "udid": "C160E86C-75C2-4DB0-BE4C-C5181D1B245D",
        "available": true
      },
      {
        "id": "5C2EE28C-7D98-40D9-91FE-C6B75E94B2EC",
        "name": "iPhone 17 Pro Max",
        "state": "Shutdown",
        "type": "simulator",
        "os_version": "26.2",
        "udid": "5C2EE28C-7D98-40D9-91FE-C6B75E94B2EC",
        "available": true
      },
      {
        "id": "1FE26C08-8069-42E4-B3AB-02931C4E070C",
        "name": "iPad Pro 13-inch (M5)",
        "state": "Shutdown",
        "type": "simulator",
        "os_version": "26.2",
        "udid": "1FE26C08-8069-42E4-B3AB-02931C4E070C",
        "available": true
      }
    ]
  },
  "timestamp": "2026-02-04T17:20:59Z"
}
```

### Empty Device List

If no simulators are available (or none are created), the command returns an empty array:

```json
{
  "success": true,
  "action": "devices.list",
  "result": {
    "devices": []
  },
  "timestamp": "2026-02-04T17:20:59Z"
}
```

Note: This is **not** considered an error. An empty device list is a valid state.

## Error Cases

### xcrun Not Available

If Xcode Command Line Tools are not installed:

```json
{
  "success": false,
  "action": "devices.list",
  "error": {
    "code": "DEVICE_DISCOVERY_FAILED",
    "message": "failed to run xcrun simctl: exec: \"xcrun\": executable file not found in $PATH"
  },
  "timestamp": "2026-02-04T17:20:59Z"
}
```

### xcrun Fails

If xcrun simctl returns an error:

```json
{
  "success": false,
  "action": "devices.list",
  "error": {
    "code": "DEVICE_DISCOVERY_FAILED",
    "message": "xcrun simctl failed: <error details>"
  },
  "timestamp": "2026-02-04T17:20:59Z"
}
```

## Usage in Agents

### Python Example

```python
import subprocess
import json

def get_devices():
    """Get list of available iOS simulators."""
    result = subprocess.run(
        ["ios-agent", "devices"],
        capture_output=True,
        text=True,
        check=True
    )

    response = json.loads(result.stdout)

    if not response["success"]:
        raise Exception(response["error"]["message"])

    return response["result"]["devices"]

def find_device_by_name(name):
    """Find a device by name."""
    devices = get_devices()
    for device in devices:
        if device["name"] == name:
            return device
    return None

# Usage
devices = get_devices()
print(f"Found {len(devices)} devices")

iphone = find_device_by_name("iPhone 17 Pro")
if iphone:
    print(f"iPhone ID: {iphone['id']}")
    print(f"State: {iphone['state']}")
```

### JavaScript/TypeScript Example

```typescript
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

interface Device {
  id: string;
  name: string;
  state: string;
  type: string;
  os_version: string;
  udid: string;
  available: boolean;
}

async function getDevices(): Promise<Device[]> {
  const { stdout } = await execAsync('ios-agent devices');
  const response = JSON.parse(stdout);

  if (!response.success) {
    throw new Error(response.error.message);
  }

  return response.result.devices;
}

async function findBootedDevices(): Promise<Device[]> {
  const devices = await getDevices();
  return devices.filter(d => d.state === 'Booted');
}

// Usage
const devices = await getDevices();
console.log(`Found ${devices.length} devices`);

const bootedDevices = await findBootedDevices();
console.log(`${bootedDevices.length} devices are booted`);
```

## Implementation Details

### Backend

The device discovery is implemented using:

1. **xcrun bridge** (`pkg/xcrun/bridge.go`): Wraps `xcrun simctl list devices --json`
2. **Device manager** (`pkg/device/manager.go`): Provides high-level interface
3. **Type definitions** (`pkg/device/types.go`): Defines device structures

### OS Version Extraction

The OS version is extracted from the simulator runtime string:

```
com.apple.CoreSimulator.SimRuntime.iOS-17-4 → "17.4"
com.apple.CoreSimulator.SimRuntime.iOS-16-0 → "16.0"
```

### Device Filtering

Only **available** devices are included in the output. Unavailable devices (e.g., incompatible runtimes) are automatically filtered out.

## Testing

### Unit Tests

```bash
go test ./pkg/device -v
go test ./pkg/xcrun -v
```

### Integration Tests

```bash
make integration-test
```

Integration tests verify:
- Command returns valid JSON
- JSON structure matches specification
- Device fields are populated correctly
- Empty device list is handled gracefully

## Related Commands

- `ios-agent simulator boot` - Boot a simulator (requires device ID)
- `ios-agent simulator shutdown` - Shutdown a simulator (requires device ID)
- `ios-agent state --device <id>` - Get detailed device state

## Troubleshooting

### No Devices Listed

1. Verify Xcode is installed: `xcode-select -p`
2. Check available simulators: `xcrun simctl list devices`
3. Create a simulator in Xcode if none exist

### Command Not Found

1. Ensure ios-agent is installed: `make install`
2. Verify PATH includes `/usr/local/bin`
3. Try with explicit path: `/usr/local/bin/ios-agent devices`

### Permission Denied

1. Ensure xcrun has proper permissions
2. Run Xcode at least once to accept license
3. Check Xcode Command Line Tools: `xcode-select --install`
