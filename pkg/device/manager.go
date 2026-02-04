package device

import (
	"fmt"
)

// Manager handles device discovery and management
type Manager interface {
	// ListDevices returns all available devices
	ListDevices() ([]Device, error)

	// GetDevice returns a specific device by ID
	GetDevice(id string) (*Device, error)

	// FindDeviceByName returns a device by name
	FindDeviceByName(name string) (*Device, error)
}

// LocalManager manages local iOS simulators
type LocalManager struct {
	bridge DeviceBridge
}

// DeviceBridge defines the interface for device control backends
type DeviceBridge interface {
	ListDevices() ([]Device, error)
	BootSimulator(udid string) error
	ShutdownSimulator(udid string) error
	GetDeviceState(udid string) (DeviceState, error)
}

// NewLocalManager creates a new local device manager
func NewLocalManager(bridge DeviceBridge) *LocalManager {
	return &LocalManager{
		bridge: bridge,
	}
}

// ListDevices returns all available local simulators
func (m *LocalManager) ListDevices() ([]Device, error) {
	return m.bridge.ListDevices()
}

// GetDevice returns a specific device by ID/UDID
func (m *LocalManager) GetDevice(id string) (*Device, error) {
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

// FindDeviceByName returns the first device matching the given name
func (m *LocalManager) FindDeviceByName(name string) (*Device, error) {
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

// BootSimulator boots a simulator by ID
func (m *LocalManager) BootSimulator(id string) error {
	// First verify the device exists
	dev, err := m.GetDevice(id)
	if err != nil {
		return err
	}

	// Check if already booted
	if dev.State == StateBooted {
		return fmt.Errorf("device already booted: %s", id)
	}

	return m.bridge.BootSimulator(dev.UDID)
}

// ShutdownSimulator shuts down a simulator by ID
func (m *LocalManager) ShutdownSimulator(id string) error {
	// First verify the device exists
	dev, err := m.GetDevice(id)
	if err != nil {
		return err
	}

	// Check if already shutdown
	if dev.State == StateShutdown {
		return fmt.Errorf("device already shutdown: %s", id)
	}

	return m.bridge.ShutdownSimulator(dev.UDID)
}

// GetDeviceState returns the current state of a device
func (m *LocalManager) GetDeviceState(id string) (DeviceState, error) {
	dev, err := m.GetDevice(id)
	if err != nil {
		return "", err
	}

	return m.bridge.GetDeviceState(dev.UDID)
}
