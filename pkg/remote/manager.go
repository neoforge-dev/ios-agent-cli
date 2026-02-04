package remote

import (
	"fmt"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
)

// RemoteManager manages devices on a remote ios-agent server
type RemoteManager struct {
	client *RemoteClient
}

// NewRemoteManager creates a new remote device manager
func NewRemoteManager(client *RemoteClient) *RemoteManager {
	return &RemoteManager{
		client: client,
	}
}

// ListDevices returns all available devices from the remote host
func (m *RemoteManager) ListDevices() ([]device.Device, error) {
	return m.client.ListDevices()
}

// GetDevice returns a specific device by ID from the remote host
func (m *RemoteManager) GetDevice(id string) (*device.Device, error) {
	devices, err := m.ListDevices()
	if err != nil {
		return nil, err
	}

	for _, dev := range devices {
		if dev.ID == id || dev.UDID == id {
			return &dev, nil
		}
	}

	return nil, fmt.Errorf("device not found: %s", id)
}

// FindDeviceByName returns the first device matching the given name from the remote host
func (m *RemoteManager) FindDeviceByName(name string) (*device.Device, error) {
	devices, err := m.ListDevices()
	if err != nil {
		return nil, err
	}

	for _, dev := range devices {
		if dev.Name == name {
			return &dev, nil
		}
	}

	return nil, fmt.Errorf("device not found with name: %s", name)
}

// BootSimulator boots a simulator on the remote host
func (m *RemoteManager) BootSimulator(id string) error {
	// First verify the device exists
	dev, err := m.GetDevice(id)
	if err != nil {
		return err
	}

	// Check if already booted
	if dev.State == device.StateBooted {
		return fmt.Errorf("device already booted: %s", id)
	}

	return m.client.BootSimulator(dev.UDID)
}

// ShutdownSimulator shuts down a simulator on the remote host
func (m *RemoteManager) ShutdownSimulator(id string) error {
	// First verify the device exists
	dev, err := m.GetDevice(id)
	if err != nil {
		return err
	}

	// Check if already shutdown
	if dev.State == device.StateShutdown {
		return fmt.Errorf("device already shutdown: %s", id)
	}

	return m.client.ShutdownSimulator(dev.UDID)
}

// GetDeviceState returns the current state of a device from the remote host
func (m *RemoteManager) GetDeviceState(id string) (device.DeviceState, error) {
	dev, err := m.GetDevice(id)
	if err != nil {
		return "", err
	}

	return m.client.GetDeviceState(dev.UDID)
}
