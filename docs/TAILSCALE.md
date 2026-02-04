# Tailscale Device Discovery

## Overview

IOS-018 adds automatic discovery of iOS devices and other machines on your Tailscale network. This enables distributed testing across multiple machines without manual host configuration.

## Usage

### List Local Devices Only (Default)

```bash
ios-agent devices
```

Returns only local simulators and devices connected to this machine.

### Include Remote Tailscale Machines

```bash
ios-agent devices --include-remote
```

Returns both local devices AND all machines on your Tailscale network.

## Output Format

Devices include a `location` field to distinguish local vs remote:

```json
{
  "success": true,
  "action": "devices.list",
  "result": {
    "devices": [
      {
        "id": "5C2EE28C-7D98-40D9-91FE-C6B75E94B2EC",
        "name": "iPhone 17 Pro Max",
        "state": "Booted",
        "type": "simulator",
        "os_version": "26.2",
        "location": "local"
      },
      {
        "id": "tailscale-code-mb14",
        "name": "code-mb14 (Tailscale)",
        "state": "Unknown",
        "type": "tailscale-machine",
        "os_version": "macOS",
        "location": "remote",
        "remote_host": "100.126.178.4",
        "available": true
      }
    ]
  }
}
```

## Connecting to Remote Machines

To use a remote machine discovered via Tailscale:

```bash
# First, discover machines
ios-agent devices --include-remote

# Then, connect using the remote_host IP
ios-agent devices --remote-host 100.126.178.4:22
ios-agent app launch --remote-host 100.126.178.4:22 --device <id> --bundle com.example.app
```

## Implementation Details

### Discovery Flow

1. Run `tailscale status --json` to get network topology
2. Parse machines from the response (Self + Peer entries)
3. Extract hostname, OS, IP, and online status
4. Add as pseudo-device entries with type `tailscale-machine`

### API

Package: `pkg/tailscale`

```go
// Discover all machines on Tailscale network
machines, err := tailscale.DiscoverMachines()

// Find machine by name
machine, err := tailscale.GetMachineByName("code-mb14")

// Find machine by IP
machine, err := tailscale.GetMachineByIP("100.126.178.4")

// Check if ios-agent is running (MVP: always returns false)
isAvailable := tailscale.ProbeForIOSAgent("100.126.178.4")
```

### Machine Type

```go
type Machine struct {
    Name        string `json:"name"`
    IP          string `json:"ip"`
    Online      bool   `json:"online"`
    OS          string `json:"os"`
    HostName    string `json:"hostname"`
    DNSName     string `json:"dns_name"`
    TailscaleIP string `json:"tailscale_ip"`
}
```

## Requirements

- Tailscale CLI installed (`brew install tailscale`)
- Authenticated to a Tailscale network
- For remote operations: ios-agent running on target machine

## Testing

```bash
# Run tailscale discovery tests
go test ./pkg/tailscale/...

# Run all tests
go test ./...

# Integration test (requires Tailscale connected)
go test -v ./pkg/tailscale -run TestDiscoverMachines
```

## Future Enhancements

### Post-MVP Features (Not Implemented)

1. **Active Probing** - `ProbeForIOSAgent()` currently returns false
   - Could probe port 4723 (WebDriverAgent)
   - Could SSH and check for ios-agent process
   - Could use custom discovery protocol

2. **Automatic Connection** - Select best available remote device
   - Health checks and latency measurement
   - Load balancing across multiple machines
   - Automatic failover

3. **Device Caching** - Cache Tailscale topology
   - Reduce `tailscale status` calls
   - Background refresh with TTL

## Troubleshooting

### No Tailscale Machines Shown

Check Tailscale connection:
```bash
tailscale status
```

If not connected:
```bash
tailscale up
```

### Tailscale Not Installed

```bash
# macOS
brew install tailscale

# Linux
curl -fsSL https://tailscale.com/install.sh | sh
```

### Graceful Fallback

If Tailscale is not available, the `--include-remote` flag silently fails and returns only local devices. Use `--verbose` to see Tailscale errors:

```bash
ios-agent devices --include-remote --verbose
```

## Architecture

```
┌─────────────────┐
│   cmd/devices   │  (CLI layer - adds --include-remote flag)
└────────┬────────┘
         │
         ├─────────────────┐
         │                 │
┌────────▼────────┐ ┌─────▼──────────────┐
│  pkg/device     │ │  pkg/tailscale     │
│  (LocalManager) │ │  (DiscoverMachines)│
└────────┬────────┘ └─────┬──────────────┘
         │                │
         │                │
┌────────▼────────┐ ┌─────▼──────────────┐
│   pkg/xcrun     │ │ tailscale CLI      │
│  (simctl)       │ │ (status --json)    │
└─────────────────┘ └────────────────────┘
```

## Related Features

- **IOS-017**: Remote host support (manual) - Base remote functionality
- **IOS-002**: Device discovery (local) - Base discovery functionality

## Examples

### CI/CD Integration

Use in CI pipeline to discover test devices:

```yaml
- name: Discover iOS devices
  run: |
    DEVICES=$(ios-agent devices --include-remote)
    echo "Available devices: $DEVICES"

    # Extract first available remote device
    REMOTE_HOST=$(echo $DEVICES | jq -r '.result.devices[] | select(.location=="remote" and .available==true) | .remote_host' | head -1)

    # Run tests on remote device
    if [ -n "$REMOTE_HOST" ]; then
      ios-agent app launch --remote-host "$REMOTE_HOST" --bundle com.example.app
    fi
```

### Agent Workflow

```python
import subprocess
import json

# Discover all devices (local + Tailscale)
result = subprocess.run(
    ["ios-agent", "devices", "--include-remote"],
    capture_output=True,
    text=True
)

devices = json.loads(result.stdout)["result"]["devices"]

# Filter for available remote iOS devices
remote_ios = [
    d for d in devices
    if d["location"] == "remote"
    and d["available"]
    and "iOS" in d.get("os_version", "")
]

# Use the first available remote iOS device
if remote_ios:
    host = remote_ios[0]["remote_host"]
    subprocess.run([
        "ios-agent", "app", "launch",
        "--remote-host", host,
        "--bundle", "com.example.app"
    ])
```
