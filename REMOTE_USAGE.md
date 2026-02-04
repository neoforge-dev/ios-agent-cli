# Remote Host Support

IOS-017 implementation: Connect to remote ios-agent servers via SSH.

## Overview

The `--remote-host` flag allows you to execute ios-agent commands on a remote machine that has ios-agent installed.

## Prerequisites

1. **Remote Host Setup**:
   - ios-agent must be installed on the remote machine
   - SSH access must be configured (key-based authentication recommended)
   - Remote machine must have iOS simulators available

2. **SSH Configuration**:
   - Add SSH keys to remote host: `ssh-copy-id user@remote-host`
   - Test SSH access: `ssh user@remote-host ios-agent devices`

## Usage

### Basic Syntax

```bash
ios-agent <command> --remote-host <host>[:port]
```

Default SSH port is 22. Specify custom port with `host:port` format.

### Examples

#### List Devices on Remote Host

```bash
# Using default port 22
ios-agent devices --remote-host 192.168.1.100

# Using custom port
ios-agent devices --remote-host 192.168.1.100:2222

# Using hostname
ios-agent devices --remote-host mac-mini.local
```

#### Boot Simulator on Remote Host

```bash
ios-agent simulator boot --name "iPhone 15 Pro" --remote-host 192.168.1.100
```

#### Shutdown Simulator on Remote Host

```bash
ios-agent simulator shutdown --device <udid> --remote-host 192.168.1.100
```

#### Take Screenshot from Remote Simulator

```bash
ios-agent screenshot --device <udid> --output ./remote-screenshot.png --remote-host 192.168.1.100
```

## Architecture

### Components

1. **RemoteClient** (`pkg/remote/client.go`):
   - Executes SSH commands to remote host
   - Parses JSON responses from remote ios-agent
   - Implements DeviceBridge interface

2. **RemoteManager** (`pkg/remote/manager.go`):
   - Wraps RemoteClient in Manager interface
   - Provides device discovery and lifecycle management
   - Works transparently with local or remote hosts

3. **Command Integration** (`cmd/*.go`):
   - `createDeviceManager()` helper function
   - Checks `--remote-host` flag
   - Returns LocalManager or RemoteManager accordingly

### How It Works

```
Local CLI                  Remote Host
----------                 -----------
ios-agent devices   -->    SSH connection
                    -->    executes: ios-agent devices
                    <--    returns JSON response
Parse JSON
Display results
```

## Limitations

- SSH must be configured with key-based authentication or password
- Remote host must have ios-agent installed at default location
- All output is returned as JSON (same format as local)
- Network latency affects command execution time

## Troubleshooting

### Connection Failed

```bash
# Test SSH connection manually
ssh user@remote-host

# Test remote ios-agent
ssh user@remote-host ios-agent devices
```

### Command Not Found on Remote

Ensure ios-agent is in the remote user's PATH:

```bash
ssh user@remote-host 'which ios-agent'
```

If not found, install ios-agent on the remote host or add to PATH.

### Permission Denied

Check SSH key permissions:

```bash
chmod 600 ~/.ssh/id_rsa
ssh-copy-id user@remote-host
```

## Future Enhancements

- IOS-018: Tailscale device discovery (automatic host detection)
- Connection pooling for faster repeated commands
- Support for SSH agent forwarding
- Multiplexing multiple remote hosts
