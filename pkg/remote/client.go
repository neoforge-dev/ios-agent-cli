package remote

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
)

// RemoteClient executes commands on a remote ios-agent server via SSH
type RemoteClient struct {
	Host string
	Port int
}

// NewRemoteClient creates a new remote client from a host:port string
func NewRemoteClient(hostPort string) (*RemoteClient, error) {
	if hostPort == "" {
		return nil, fmt.Errorf("remote host cannot be empty")
	}

	// Parse host:port
	parts := strings.Split(hostPort, ":")
	host := parts[0]
	port := 22 // Default SSH port

	if len(parts) > 1 {
		_, err := fmt.Sscanf(parts[1], "%d", &port)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", parts[1])
		}
	}

	if host == "" {
		return nil, fmt.Errorf("invalid remote host")
	}

	return &RemoteClient{
		Host: host,
		Port: port,
	}, nil
}

// ListDevices executes 'ios-agent devices' on the remote host
func (c *RemoteClient) ListDevices() ([]device.Device, error) {
	output, err := c.executeRemoteCommand("ios-agent", "devices")
	if err != nil {
		return nil, fmt.Errorf("failed to list remote devices: %w", err)
	}

	// Parse the JSON response
	var response struct {
		Success bool `json:"success"`
		Result  struct {
			Devices []device.Device `json:"devices"`
		} `json:"result"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse remote response: %w", err)
	}

	if !response.Success {
		if response.Error != nil {
			return nil, fmt.Errorf("remote error [%s]: %s", response.Error.Code, response.Error.Message)
		}
		return nil, fmt.Errorf("remote command failed")
	}

	return response.Result.Devices, nil
}

// ExecuteCommand executes an arbitrary ios-agent command on the remote host
func (c *RemoteClient) ExecuteCommand(cmd string, args ...string) ([]byte, error) {
	cmdArgs := append([]string{cmd}, args...)
	return c.executeRemoteCommand("ios-agent", cmdArgs...)
}

// BootSimulator boots a simulator on the remote host
func (c *RemoteClient) BootSimulator(udid string) error {
	output, err := c.executeRemoteCommand("ios-agent", "simulator", "boot", "--device", udid)
	if err != nil {
		return fmt.Errorf("failed to boot remote simulator: %w", err)
	}

	// Parse response to check for errors
	var response struct {
		Success bool `json:"success"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return fmt.Errorf("failed to parse remote response: %w", err)
	}

	if !response.Success {
		if response.Error != nil {
			return fmt.Errorf("remote error [%s]: %s", response.Error.Code, response.Error.Message)
		}
		return fmt.Errorf("failed to boot simulator")
	}

	return nil
}

// ShutdownSimulator shuts down a simulator on the remote host
func (c *RemoteClient) ShutdownSimulator(udid string) error {
	output, err := c.executeRemoteCommand("ios-agent", "simulator", "shutdown", "--device", udid)
	if err != nil {
		return fmt.Errorf("failed to shutdown remote simulator: %w", err)
	}

	// Parse response to check for errors
	var response struct {
		Success bool `json:"success"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return fmt.Errorf("failed to parse remote response: %w", err)
	}

	if !response.Success {
		if response.Error != nil {
			return fmt.Errorf("remote error [%s]: %s", response.Error.Code, response.Error.Message)
		}
		return fmt.Errorf("failed to shutdown simulator")
	}

	return nil
}

// GetDeviceState gets the state of a device on the remote host
func (c *RemoteClient) GetDeviceState(udid string) (device.DeviceState, error) {
	output, err := c.executeRemoteCommand("ios-agent", "devices")
	if err != nil {
		return "", fmt.Errorf("failed to get remote device state: %w", err)
	}

	// Parse the JSON response
	var response struct {
		Success bool `json:"success"`
		Result  struct {
			Devices []device.Device `json:"devices"`
		} `json:"result"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return "", fmt.Errorf("failed to parse remote response: %w", err)
	}

	if !response.Success {
		if response.Error != nil {
			return "", fmt.Errorf("remote error [%s]: %s", response.Error.Code, response.Error.Message)
		}
		return "", fmt.Errorf("remote command failed")
	}

	// Find the device by UDID
	for _, dev := range response.Result.Devices {
		if dev.UDID == udid {
			return dev.State, nil
		}
	}

	return "", fmt.Errorf("device not found: %s", udid)
}

// executeRemoteCommand executes a command on the remote host via SSH
func (c *RemoteClient) executeRemoteCommand(command string, args ...string) ([]byte, error) {
	// Build the remote command
	remoteCmd := command
	if len(args) > 0 {
		// Properly quote arguments for SSH
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			// Escape single quotes in arguments
			escapedArg := strings.ReplaceAll(arg, "'", "'\\''")
			quotedArgs[i] = fmt.Sprintf("'%s'", escapedArg)
		}
		remoteCmd = fmt.Sprintf("%s %s", command, strings.Join(quotedArgs, " "))
	}

	// Build SSH command
	sshArgs := []string{
		"-p", fmt.Sprintf("%d", c.Port),
		c.Host,
		remoteCmd,
	}

	// Execute SSH command
	cmd := exec.Command("ssh", sshArgs...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("ssh command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute ssh: %w", err)
	}

	return output, nil
}
