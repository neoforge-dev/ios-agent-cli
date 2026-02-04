# iOS Agent CLI - Claude Instructions

## Project Context

- **Domain:** neoforge-dev
- **Status:** MVP Development
- **Stack:** Go 1.21+ with Cobra CLI framework
- **Purpose:** AI-agent-friendly iOS automation CLI

## Key Files

- Entry point: `cmd/root.go` (cobra CLI root)
- Device manager: `pkg/device/manager.go`
- mobilecli wrapper: `pkg/mobilecli/client.go`
- xcrun bridge: `pkg/xcrun/bridge.go`
- Output formatter: `pkg/output/formatter.go`

## Development Commands

```bash
# Build
make build

# Test
make test

# Run integration tests
make integration-test

# Install locally
make install

# Lint
make lint
```

## Architecture Layers

1. **CLI Layer** (`cmd/`): Cobra commands, flag parsing, JSON output
2. **Device Manager** (`pkg/device/`): Abstract interface for local/remote devices
3. **Backend Layer** (`pkg/mobilecli/`, `pkg/xcrun/`): Actual device control

## Design Principles

- **Agent-First**: Simple, deterministic CLI commands + JSON output
- **Batteries Included**: Works with local simulators without complex setup
- **Minimal Abstraction**: Delegate to proven tools (mobilecli, xcrun simctl)
- **Progressive Disclosure**: MVP focuses on core happy path

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `xcrun simctl` - Simulator control (built into Xcode)
- `mobilecli` - UI interactions (external tool)

## Error Codes

- `DEVICE_NOT_FOUND` - Device ID doesn't exist
- `DEVICE_UNREACHABLE` - Connection failed
- `APP_NOT_FOUND` - Bundle ID not installed
- `UI_ACTION_FAILED` - Tap/swipe coordinates invalid
- `SIMULATOR_TIMEOUT` - Boot/shutdown exceeded timeout

## Testing Strategy

1. **Unit tests**: JSON serialization, command routing, error mapping
2. **Integration tests**: Real simulator interactions
3. **Mocks**: DeviceManager interface for unit testing

## MVP Scope

**Included:**
- Device discovery (local simulators)
- Simulator boot/shutdown
- App launch, terminate, install, uninstall
- UI interactions (tap, swipe, text, button)
- Screenshot capture
- JSON output format

**Out of Scope (Post-MVP):**
- Remote Tailscale support
- Video recording
- Android support
- Complex gestures

## Harness Integration

This project can be built using the FORGE harness flywheel:

```bash
# Run via harness
forge-harness flywheel run -d neoforge-dev -p ios-agent-cli

# Create feature
forge-harness feature add -p ios-agent-cli -t "Device discovery command"
```

## Related Projects

- `forge-terminal` (codeswiftr-com): First test target for ios-agent
- `agent-browser`: Web equivalent (inspiration)
