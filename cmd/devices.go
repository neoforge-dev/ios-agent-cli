package cmd

import (
	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/remote"
	"github.com/neoforge-dev/ios-agent-cli/pkg/tailscale"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/spf13/cobra"
)

var (
	includeRemote bool
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List all available iOS devices and simulators",
	Long: `List all available iOS devices and simulators.

This command discovers local iOS simulators using xcrun simctl.
With --include-remote, it also shows available machines on the Tailscale network.
With --remote-host, it connects to a remote ios-agent server.
Returns JSON output with device ID, name, state, type, and OS version.

Examples:
  ios-agent devices                            # List local devices
  ios-agent devices --include-remote           # Include Tailscale machines
  ios-agent devices --remote-host host:port    # List remote devices
  ios-agent devices --format json              # Explicit JSON output`,
	Run: runDevicesCmd,
}

func init() {
	rootCmd.AddCommand(devicesCmd)
	devicesCmd.Flags().BoolVar(&includeRemote, "include-remote", false, "Include remote devices on Tailscale network")
}

func runDevicesCmd(cmd *cobra.Command, args []string) {
	var allDevices []device.Device

	// Get local or remote devices based on --remote-host flag
	if remoteHost == "" {
		// Create local device manager with xcrun bridge
		bridge := xcrun.NewBridge()
		manager := device.NewLocalManager(bridge)

		// List local devices
		localDevices, err := manager.ListDevices()
		if err != nil {
			outputError("devices.list", "DEVICE_DISCOVERY_FAILED", err.Error(), nil)
			return
		}

		// Mark all local devices with location
		for i := range localDevices {
			localDevices[i].Location = device.LocationLocal
		}

		allDevices = localDevices
	} else {
		// Remote host specified - use remote manager
		manager := createDeviceManager()
		devices, err := manager.ListDevices()
		if err != nil {
			outputError("devices.list", "DEVICE_DISCOVERY_FAILED", err.Error(), nil)
			return
		}

		// Mark remote devices
		for i := range devices {
			devices[i].Location = device.LocationRemote
			devices[i].RemoteHost = remoteHost
		}

		allDevices = devices
	}

	// If include-remote flag is set, also discover Tailscale machines
	if includeRemote {
		machines, err := tailscale.DiscoverMachines()
		if err != nil {
			// Don't fail if Tailscale discovery fails, just log if verbose
			if verbose {
				// Note: We can't use outputError here as it calls os.Exit
				// Just continue without Tailscale machines
			}
		} else {
			// Add Tailscale machines as remote "devices"
			// Note: These are machines, not actual iOS devices
			// User needs to specify --remote-host to connect to them
			for _, machine := range machines {
				// Skip if no IP
				if machine.TailscaleIP == "" {
					continue
				}

				// Create a pseudo-device entry for each Tailscale machine
				tsDevice := device.Device{
					ID:         "tailscale-" + machine.Name,
					Name:       machine.Name + " (Tailscale)",
					State:      device.DeviceState("Unknown"),
					Type:       device.DeviceType("tailscale-machine"),
					OSVersion:  machine.OS,
					Location:   device.LocationRemote,
					RemoteHost: machine.TailscaleIP,
					Available:  machine.Online,
				}

				allDevices = append(allDevices, tsDevice)
			}
		}
	}

	// Output success response with device list
	result := device.DeviceList{
		Devices: allDevices,
	}

	outputSuccess("devices.list", result)
}

// createDeviceManager creates the appropriate device manager based on flags
func createDeviceManager() device.Manager {
	if remoteHost != "" {
		// Create remote manager
		client, err := remote.NewRemoteClient(remoteHost)
		if err != nil {
			outputError("manager.init", "REMOTE_CLIENT_FAILED", err.Error(), nil)
			return nil
		}
		return remote.NewRemoteManager(client)
	}

	// Create local manager with xcrun bridge
	bridge := xcrun.NewBridge()
	return device.NewLocalManager(bridge)
}
