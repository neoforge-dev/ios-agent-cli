package cmd

import (
	"fmt"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/spf13/cobra"
)

var (
	// Boot command flags
	simulatorName string
	osVersion     string
	wait          bool
	timeout       int

	// Shutdown command flags
	shutdownDeviceID string
)

// simulatorCmd represents the simulator command group
var simulatorCmd = &cobra.Command{
	Use:   "simulator",
	Short: "Manage iOS simulators",
	Long: `Manage iOS simulators - boot, shutdown, and control lifecycle.

Examples:
  ios-agent simulator boot --name "iPhone 15 Pro"
  ios-agent simulator boot --name "iPhone 14" --os-version "17.4"
  ios-agent simulator shutdown --device <udid>`,
}

// bootCmd represents the boot subcommand
var bootCmd = &cobra.Command{
	Use:   "boot",
	Short: "Boot an iOS simulator",
	Long: `Boot an iOS simulator by name, optionally filtering by OS version.

The command will:
1. Find a simulator matching the given name (and OS version if specified)
2. Boot the simulator using xcrun simctl
3. Poll the simulator state until it is fully booted (or timeout)
4. Return device information and boot time in JSON format

Examples:
  ios-agent simulator boot --name "iPhone 15 Pro"
  ios-agent simulator boot --name "iPhone 14" --os-version "17.4"
  ios-agent simulator boot --name "iPhone 15" --timeout 120
  ios-agent simulator boot --name "iPad Pro" --wait=false`,
	Run: runBootCmd,
}

// shutdownCmd represents the shutdown subcommand
var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Shutdown a running iOS simulator",
	Long: `Shutdown a running iOS simulator by device ID.

The command will:
1. Verify the device exists and is running
2. Shutdown the simulator using xcrun simctl
3. Return success status in JSON format

Examples:
  ios-agent simulator shutdown --device <udid>
  ios-agent simulator shutdown -d <udid>`,
	Run: runShutdownCmd,
}

func init() {
	rootCmd.AddCommand(simulatorCmd)
	simulatorCmd.AddCommand(bootCmd)
	simulatorCmd.AddCommand(shutdownCmd)

	// Boot command flags
	bootCmd.Flags().StringVar(&simulatorName, "name", "", "Simulator name to boot (required)")
	bootCmd.Flags().StringVar(&osVersion, "os-version", "", "Optional OS version filter (e.g., '17.4')")
	bootCmd.Flags().BoolVar(&wait, "wait", true, "Wait for boot to complete")
	bootCmd.Flags().IntVar(&timeout, "timeout", 60, "Boot timeout in seconds")
	bootCmd.MarkFlagRequired("name")

	// Shutdown command flags
	shutdownCmd.Flags().StringVarP(&shutdownDeviceID, "device", "d", "", "Device ID to shutdown (required)")
	shutdownCmd.MarkFlagRequired("device")
}

// BootResult represents the result of a boot operation
type BootResult struct {
	Device     *device.Device `json:"device"`
	BootTimeMs int64          `json:"boot_time_ms"`
}

// ShutdownResult represents the result of a shutdown operation
type ShutdownResult struct {
	Device  *device.Device `json:"device"`
	Message string         `json:"message"`
}

func runBootCmd(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Find device by name
	dev, err := findDeviceByNameAndOS(manager, simulatorName, osVersion)
	if err != nil {
		outputError("simulator.boot", "DEVICE_NOT_FOUND", err.Error(), map[string]string{
			"name":       simulatorName,
			"os_version": osVersion,
		})
		return
	}

	// Check if already booted
	if dev.State == device.StateBooted {
		// Already booted, return success immediately
		result := BootResult{
			Device:     dev,
			BootTimeMs: 0,
		}
		outputSuccess("simulator.boot", result)
		return
	}

	// Boot the simulator
	if err := manager.BootSimulator(dev.ID); err != nil {
		outputError("simulator.boot", "BOOT_FAILED", err.Error(), map[string]string{
			"device_id": dev.ID,
		})
		return
	}

	// If wait is false, return immediately
	if !wait {
		dev.State = device.StateBooting
		result := BootResult{
			Device:     dev,
			BootTimeMs: time.Since(startTime).Milliseconds(),
		}
		outputSuccess("simulator.boot", result)
		return
	}

	// Poll for boot completion
	bootedDev, err := pollForBootCompletion(manager, dev.ID, timeout)
	if err != nil {
		outputError("simulator.boot", "SIMULATOR_TIMEOUT", err.Error(), map[string]string{
			"device_id":     dev.ID,
			"timeout_sec":   fmt.Sprintf("%d", timeout),
			"elapsed_sec":   fmt.Sprintf("%.1f", time.Since(startTime).Seconds()),
		})
		return
	}

	// Calculate boot time
	bootTime := time.Since(startTime).Milliseconds()

	result := BootResult{
		Device:     bootedDev,
		BootTimeMs: bootTime,
	}

	outputSuccess("simulator.boot", result)
}

func runShutdownCmd(cmd *cobra.Command, args []string) {
	// Create device manager with xcrun bridge
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	// Get device to verify it exists
	dev, err := manager.GetDevice(shutdownDeviceID)
	if err != nil {
		outputError("simulator.shutdown", "DEVICE_NOT_FOUND", err.Error(), map[string]string{
			"device_id": shutdownDeviceID,
		})
		return
	}

	// Shutdown the simulator
	if err := manager.ShutdownSimulator(dev.ID); err != nil {
		outputError("simulator.shutdown", "SHUTDOWN_FAILED", err.Error(), map[string]string{
			"device_id": dev.ID,
		})
		return
	}

	// Update device state
	dev.State = device.StateShutdown

	result := ShutdownResult{
		Device:  dev,
		Message: "Simulator shutdown successfully",
	}

	outputSuccess("simulator.shutdown", result)
}

// findDeviceByNameAndOS finds a device matching the name and optional OS version
func findDeviceByNameAndOS(manager *device.LocalManager, name, osVersion string) (*device.Device, error) {
	devices, err := manager.ListDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	var candidates []*device.Device
	for i := range devices {
		dev := &devices[i]
		if dev.Name == name {
			// If OS version is specified, filter by it
			if osVersion != "" && dev.OSVersion != osVersion {
				continue
			}
			candidates = append(candidates, dev)
		}
	}

	if len(candidates) == 0 {
		if osVersion != "" {
			return nil, fmt.Errorf("no device found with name '%s' and OS version '%s'", name, osVersion)
		}
		return nil, fmt.Errorf("no device found with name '%s'", name)
	}

	// Return the first candidate (prefer booted devices)
	for _, dev := range candidates {
		if dev.State == device.StateBooted {
			return dev, nil
		}
	}

	return candidates[0], nil
}

// pollForBootCompletion polls the device state until it is booted or timeout
func pollForBootCompletion(manager *device.LocalManager, deviceID string, timeoutSec int) (*device.Device, error) {
	pollInterval := 500 * time.Millisecond
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)

	for time.Now().Before(deadline) {
		state, err := manager.GetDeviceState(deviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get device state: %w", err)
		}

		if state == device.StateBooted {
			// Device is booted, fetch full device info
			dev, err := manager.GetDevice(deviceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get device info: %w", err)
			}
			return dev, nil
		}

		// Sleep before next poll
		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("simulator boot timed out after %d seconds", timeoutSec)
}
