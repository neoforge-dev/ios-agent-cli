package device

// DeviceType represents the type of device
type DeviceType string

const (
	// DeviceTypeSimulator represents an iOS simulator
	DeviceTypeSimulator DeviceType = "simulator"
	// DeviceTypePhysical represents a physical iOS device
	DeviceTypePhysical DeviceType = "physical"
)

// DeviceState represents the state of a device
type DeviceState string

const (
	// StateBooted indicates the device is running
	StateBooted DeviceState = "Booted"
	// StateShutdown indicates the device is shut down
	StateShutdown DeviceState = "Shutdown"
	// StateCreating indicates the device is being created
	StateCreating DeviceState = "Creating"
	// StateBooting indicates the device is booting
	StateBooting DeviceState = "Booting"
	// StateShuttingDown indicates the device is shutting down
	StateShuttingDown DeviceState = "ShuttingDown"
)

// Device represents an iOS device or simulator
type Device struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	State     DeviceState `json:"state"`
	Type      DeviceType  `json:"type"`
	OSVersion string      `json:"os_version"`
	UDID      string      `json:"udid,omitempty"`
	Available bool        `json:"available,omitempty"`
}

// DeviceList represents a list of devices
type DeviceList struct {
	Devices []Device `json:"devices"`
}
