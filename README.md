# iOS Agent CLI

**Modern CLI tool enabling AI agents to automate iOS app testing on local and remote devices/simulators.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Overview

iOS Agent CLI mirrors the capabilities of [Agent Browser](https://github.com/anthropics/agent-browser) for web testing, but for native iOS apps. It provides a simple, deterministic CLI with JSON output that AI agents (Claude Code, Cursor, etc.) can use to automate iOS testing workflows.

### Key Features

- **Device Discovery**: Find local simulators and remote devices over Tailscale
- **Simulator Control**: Boot, shutdown, and manage iOS simulators
- **App Management**: Install, launch, terminate apps
- **UI Interactions**: Tap, swipe, text input, button presses
- **Observation**: Screenshots + structured JSON state for agent decision-making
- **Remote Support**: Test on devices anywhere via Tailscale VPN

## Quick Start

### Prerequisites

- macOS with Xcode Command Line Tools
- Go 1.21+
- [mobilecli](https://github.com/mobile-next/mobilecli) (for UI interactions)

### Installation

```bash
# Build from source
git clone https://github.com/neoforge-dev/ios-agent-cli.git
cd ios-agent-cli
make build
make install  # Installs to /usr/local/bin

# Or with Go
go install github.com/neoforge-dev/ios-agent-cli@latest
```

### Basic Usage

```bash
# Discover devices
ios-agent devices

# Boot a simulator
ios-agent simulator boot --name "iPhone 15" --os-version 17.4

# Launch an app
ios-agent app launch --device <device-id> --bundle com.example.app

# Take a screenshot
ios-agent screenshot --device <device-id> --output ./shot.png

# Interact with UI
ios-agent io tap --device <device-id> --x 100 --y 200
ios-agent io text --device <device-id> "hello world"
```

## Agent Integration Example

```python
# Agent loop pseudo-code
device_id = discover_ios_device()
launch_app(device_id, "com.example.app")

for step in test_steps:
    screenshot = take_screenshot(device_id)
    state = analyze_screenshot(screenshot)  # AI vision

    if state["needs_login"]:
        tap(device_id, state["login_button"])
        type_text(device_id, credentials)
    elif state["on_home_screen"]:
        tap(device_id, state["next_button"])
```

## Command Reference

### Device Management
```bash
ios-agent devices [--include-remote] [--remote-host HOST:PORT]
ios-agent state --device ID [--include-screenshot]
```

### Simulator Control
```bash
ios-agent simulator boot --name NAME [--os-version VERSION]
ios-agent simulator shutdown --device ID
```

### App Management
```bash
ios-agent app launch --device ID --bundle BUNDLE_ID [--wait-for-ready SECONDS]
ios-agent app terminate --device ID --bundle BUNDLE_ID
ios-agent app install --device ID --ipa PATH
ios-agent app uninstall --device ID --bundle BUNDLE_ID
```

### UI Interactions
```bash
ios-agent io tap --device ID --x X --y Y
ios-agent io text --device ID "TEXT"
ios-agent io swipe --device ID --start-x X1 --start-y Y1 --end-x X2 --end-y Y2
ios-agent io button --device ID --button {HOME|POWER|VOLUME_UP|VOLUME_DOWN}
```

### Observation
```bash
ios-agent screenshot --device ID [--format {png|jpeg}] [--output PATH]
```

## Remote Device Support (Tailscale)

Connect to iOS simulators or physical devices on remote Macs over Tailscale.

### Setup (Remote Mac)
```bash
tailscale up --operator=<user>
mobilecli server --listen 0.0.0.0:4723
```

### Usage (Dev Machine)
```bash
ios-agent devices --include-remote
ios-agent screenshot --device remote-mac1-iphone15
```

## JSON Output Format

All commands return structured JSON for agent parsing:

```json
{
  "success": true,
  "action": "screenshot",
  "result": {
    "path": "/tmp/shot.png",
    "width": 1170,
    "height": 2532
  },
  "timestamp": "2026-02-04T18:53:00Z"
}
```

### Error Format
```json
{
  "success": false,
  "error": {
    "code": "DEVICE_NOT_FOUND",
    "message": "No device with ID '123' found"
  }
}
```

## Architecture

```
Agent Interface (CLI)
         |
    JSON Router
         |
   Device Manager
    /          \
Local          Remote
(simctl)    (Tailscale)
    |            |
  mobilecli   mobilecli
```

## Development

```bash
# Run tests
make test

# Build binary
make build

# Run integration tests (requires simulator)
make integration-test

# Lint
make lint
```

## Project Structure

```
ios-agent-cli/
├── cmd/           # CLI commands (cobra)
├── pkg/           # Core packages
│   ├── device/    # Device manager
│   ├── mobilecli/ # mobilecli HTTP client
│   ├── xcrun/     # simctl wrapper
│   ├── tailscale/ # Remote discovery
│   └── output/    # JSON formatting
├── test/          # Integration tests
├── docs/          # Documentation
└── Makefile
```

## Roadmap

### Phase 1 (MVP) - Current
- [x] Project scaffold
- [ ] Device discovery (local)
- [ ] Simulator lifecycle
- [ ] App management
- [ ] Basic UI interactions
- [ ] Screenshot capture
- [ ] Error handling

### Phase 2 (Remote)
- [ ] Tailscale integration
- [ ] Remote device discovery
- [ ] Connection pooling

### Phase 3 (Advanced)
- [ ] Video streaming
- [ ] Network interception
- [ ] Android support

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for development setup and guidelines.

---

**Built for the FORGE portfolio** | [NeoForge Dev](https://neoforge.dev)
