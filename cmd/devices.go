package cmd

import (
	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List all available iOS devices and simulators",
	Long: `List all available iOS devices and simulators.

This command discovers local iOS simulators using xcrun simctl.
Returns JSON output with device ID, name, state, type, and OS version.

Examples:
  ios-agent devices                    # List all devices
  ios-agent devices --format json      # Explicit JSON output`,
	Run: runDevicesCmd,
}

func init() {
	rootCmd.AddCommand(devicesCmd)
}

func runDevicesCmd(cmd *cobra.Command, args []string) {
	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// List all devices
	devices, err := manager.ListDevices()
	if err != nil {
		outputError("devices.list", "DEVICE_DISCOVERY_FAILED", err.Error(), nil)
		return
	}

	// Output success response with device list
	result := device.DeviceList{
		Devices: devices,
	}

	outputSuccess("devices.list", result)
}
