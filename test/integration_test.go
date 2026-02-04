//go:build integration
// +build integration

package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestEnvironment checks if simulators are available and returns a manager
// It skips the test if no simulators are available
func setupTestEnvironment(t *testing.T) (*device.LocalManager, []device.Device) {
	bridge := xcrun.NewBridge()
	manager := device.NewLocalManager(bridge)

	devices, err := manager.ListDevices()
	if err != nil {
		t.Skipf("Cannot list devices, skipping test: %v", err)
		return nil, nil
	}

	if len(devices) == 0 {
		t.Skip("No simulators available, skipping test")
		return nil, nil
	}

	return manager, devices
}

// findShutdownSimulator finds the first available simulator that is shutdown
func findShutdownSimulator(devices []device.Device) *device.Device {
	for _, dev := range devices {
		if dev.State == device.StateShutdown && dev.Available {
			return &dev
		}
	}
	return nil
}

// findBootedSimulator finds the first available simulator that is booted
func findBootedSimulator(devices []device.Device) *device.Device {
	for _, dev := range devices {
		if dev.State == device.StateBooted && dev.Available {
			return &dev
		}
	}
	return nil
}

// TestIntegration_DeviceDiscovery tests that we can discover real simulators
func TestIntegration_DeviceDiscovery(t *testing.T) {
	manager, devices := setupTestEnvironment(t)
	if manager == nil {
		return // Test was skipped
	}

	t.Run("list devices returns valid simulators", func(t *testing.T) {
		assert.NotEmpty(t, devices, "Should find at least one simulator")

		for _, dev := range devices {
			// Verify required fields are populated
			assert.NotEmpty(t, dev.ID, "Device ID should not be empty")
			assert.NotEmpty(t, dev.Name, "Device name should not be empty")
			assert.NotEmpty(t, dev.UDID, "Device UDID should not be empty")
			assert.NotEmpty(t, dev.OSVersion, "OS version should not be empty")
			assert.Equal(t, device.DeviceTypeSimulator, dev.Type, "Device type should be simulator")
			assert.True(t, dev.Available, "Listed devices should be available")

			// Verify state is valid
			validStates := []device.DeviceState{
				device.StateBooted,
				device.StateShutdown,
				device.StateCreating,
				device.StateBooting,
				device.StateShuttingDown,
			}
			assert.Contains(t, validStates, dev.State, "Device state should be valid")
		}
	})

	t.Run("get device by ID returns correct device", func(t *testing.T) {
		targetDevice := devices[0]

		retrievedDev, err := manager.GetDevice(targetDevice.ID)
		require.NoError(t, err, "Should retrieve device by ID")
		assert.NotNil(t, retrievedDev, "Retrieved device should not be nil")
		assert.Equal(t, targetDevice.ID, retrievedDev.ID, "Device ID should match")
		assert.Equal(t, targetDevice.Name, retrievedDev.Name, "Device name should match")
		assert.Equal(t, targetDevice.UDID, retrievedDev.UDID, "Device UDID should match")
	})

	t.Run("get device by UDID returns correct device", func(t *testing.T) {
		targetDevice := devices[0]

		retrievedDev, err := manager.GetDevice(targetDevice.UDID)
		require.NoError(t, err, "Should retrieve device by UDID")
		assert.NotNil(t, retrievedDev, "Retrieved device should not be nil")
		assert.Equal(t, targetDevice.UDID, retrievedDev.UDID, "Device UDID should match")
	})

	t.Run("find device by name returns correct device", func(t *testing.T) {
		targetDevice := devices[0]

		foundDev, err := manager.FindDeviceByName(targetDevice.Name)
		require.NoError(t, err, "Should find device by name")
		assert.NotNil(t, foundDev, "Found device should not be nil")
		assert.Equal(t, targetDevice.Name, foundDev.Name, "Device name should match")
	})

	t.Run("get nonexistent device returns error", func(t *testing.T) {
		dev, err := manager.GetDevice("nonexistent-device-id-12345")
		assert.Error(t, err, "Should return error for nonexistent device")
		assert.Nil(t, dev, "Device should be nil")
		assert.Contains(t, err.Error(), "device not found", "Error should indicate device not found")
	})

	t.Run("find device by nonexistent name returns error", func(t *testing.T) {
		dev, err := manager.FindDeviceByName("Nonexistent iPhone Model XYZ")
		assert.Error(t, err, "Should return error for nonexistent device name")
		assert.Nil(t, dev, "Device should be nil")
		assert.Contains(t, err.Error(), "device not found", "Error should indicate device not found")
	})
}

// TestIntegration_SimulatorBootShutdownLifecycle tests the complete boot/shutdown cycle
func TestIntegration_SimulatorBootShutdownLifecycle(t *testing.T) {
	manager, devices := setupTestEnvironment(t)
	if manager == nil {
		return // Test was skipped
	}

	// Find a shutdown simulator to test with
	shutdownSim := findShutdownSimulator(devices)
	if shutdownSim == nil {
		t.Skip("No shutdown simulators available for boot test")
		return
	}

	// Record the original state to clean up later
	originalState := shutdownSim.State
	deviceID := shutdownSim.ID

	t.Logf("Testing with device: %s (%s) - %s", shutdownSim.Name, shutdownSim.UDID, shutdownSim.OSVersion)

	// Ensure cleanup happens regardless of test outcome
	defer func() {
		// Best-effort cleanup: shut down the simulator if we booted it
		if originalState == device.StateShutdown {
			t.Logf("Cleaning up: shutting down simulator %s", deviceID)
			_ = manager.ShutdownSimulator(deviceID)
			// Wait a bit for shutdown to complete
			time.Sleep(2 * time.Second)
		}
	}()

	t.Run("boot simulator from shutdown state", func(t *testing.T) {
		startTime := time.Now()

		err := manager.BootSimulator(deviceID)
		require.NoError(t, err, "Should boot simulator successfully")

		bootDuration := time.Since(startTime)
		t.Logf("Boot initiated in %v", bootDuration)

		// Wait for boot to complete (with timeout)
		maxWaitTime := 60 * time.Second
		pollInterval := 2 * time.Second
		bootComplete := false

		for elapsed := time.Duration(0); elapsed < maxWaitTime; elapsed += pollInterval {
			state, err := manager.GetDeviceState(deviceID)
			require.NoError(t, err, "Should get device state")

			t.Logf("Device state after %v: %s", elapsed, state)

			if state == device.StateBooted {
				bootComplete = true
				t.Logf("Boot completed in %v", elapsed)
				break
			}

			time.Sleep(pollInterval)
		}

		assert.True(t, bootComplete, "Simulator should complete boot within timeout")

		// Verify device is booted
		dev, err := manager.GetDevice(deviceID)
		require.NoError(t, err, "Should get device after boot")
		assert.Equal(t, device.StateBooted, dev.State, "Device should be in Booted state")
	})

	t.Run("boot already booted simulator returns error", func(t *testing.T) {
		// Device should still be booted from previous test
		err := manager.BootSimulator(deviceID)
		assert.Error(t, err, "Should return error when booting already booted device")
		assert.Contains(t, err.Error(), "already booted", "Error should indicate device is already booted")
	})

	t.Run("shutdown booted simulator", func(t *testing.T) {
		startTime := time.Now()

		err := manager.ShutdownSimulator(deviceID)
		require.NoError(t, err, "Should shutdown simulator successfully")

		shutdownDuration := time.Since(startTime)
		t.Logf("Shutdown initiated in %v", shutdownDuration)

		// Wait for shutdown to complete (with timeout)
		maxWaitTime := 30 * time.Second
		pollInterval := 1 * time.Second
		shutdownComplete := false

		for elapsed := time.Duration(0); elapsed < maxWaitTime; elapsed += pollInterval {
			state, err := manager.GetDeviceState(deviceID)
			require.NoError(t, err, "Should get device state")

			t.Logf("Device state after %v: %s", elapsed, state)

			if state == device.StateShutdown {
				shutdownComplete = true
				t.Logf("Shutdown completed in %v", elapsed)
				break
			}

			time.Sleep(pollInterval)
		}

		assert.True(t, shutdownComplete, "Simulator should complete shutdown within timeout")

		// Verify device is shutdown
		dev, err := manager.GetDevice(deviceID)
		require.NoError(t, err, "Should get device after shutdown")
		assert.Equal(t, device.StateShutdown, dev.State, "Device should be in Shutdown state")
	})

	t.Run("shutdown already shutdown simulator returns error", func(t *testing.T) {
		// Device should be shutdown from previous test
		err := manager.ShutdownSimulator(deviceID)
		assert.Error(t, err, "Should return error when shutting down already shutdown device")
		assert.Contains(t, err.Error(), "already shutdown", "Error should indicate device is already shutdown")
	})
}

// TestIntegration_ScreenshotCapture tests screenshot capture functionality
func TestIntegration_ScreenshotCapture(t *testing.T) {
	_, devices := setupTestEnvironment(t)
	if devices == nil {
		return // Test was skipped
	}

	// Find a booted simulator
	bootedSim := findBootedSimulator(devices)
	if bootedSim == nil {
		t.Skip("No booted simulators available for screenshot test. Please boot a simulator manually.")
		return
	}

	bridge := xcrun.NewBridge()
	t.Logf("Testing screenshot with device: %s (%s)", bootedSim.Name, bootedSim.UDID)

	t.Run("capture screenshot creates file", func(t *testing.T) {
		// Create temp directory for test artifacts
		tempDir := t.TempDir()
		screenshotPath := filepath.Join(tempDir, "test_screenshot.png")

		result, err := bridge.CaptureScreenshot(bootedSim.UDID, screenshotPath)
		require.NoError(t, err, "Should capture screenshot successfully")
		require.NotNil(t, result, "Screenshot result should not be nil")

		// Verify result metadata
		assert.Equal(t, screenshotPath, result.Path, "Result path should match requested path")
		assert.Equal(t, bootedSim.UDID, result.DeviceID, "Result device ID should match")
		assert.Equal(t, "png", result.Format, "Format should be PNG")
		assert.Greater(t, result.SizeBytes, int64(0), "File size should be greater than 0")
		assert.NotEmpty(t, result.Timestamp, "Timestamp should be set")

		// Verify file exists and is readable
		fileInfo, err := os.Stat(screenshotPath)
		require.NoError(t, err, "Screenshot file should exist")
		assert.Greater(t, fileInfo.Size(), int64(1000), "Screenshot file should be larger than 1KB")
		assert.Equal(t, result.SizeBytes, fileInfo.Size(), "File size should match result")

		t.Logf("Screenshot captured: %s (%d bytes)", screenshotPath, fileInfo.Size())
	})

	t.Run("capture screenshot with custom filename", func(t *testing.T) {
		tempDir := t.TempDir()
		timestamp := time.Now().Format("20060102-150405")
		screenshotPath := filepath.Join(tempDir, fmt.Sprintf("simulator_%s_%s.png", bootedSim.Name, timestamp))

		result, err := bridge.CaptureScreenshot(bootedSim.UDID, screenshotPath)
		require.NoError(t, err, "Should capture screenshot with custom filename")
		assert.FileExists(t, screenshotPath, "Screenshot file should exist at custom path")
		assert.Greater(t, result.SizeBytes, int64(0), "File size should be greater than 0")
	})

	t.Run("capture screenshot to invalid path returns error", func(t *testing.T) {
		invalidPath := "/nonexistent/directory/that/does/not/exist/screenshot.png"

		result, err := bridge.CaptureScreenshot(bootedSim.UDID, invalidPath)
		assert.Error(t, err, "Should return error for invalid path")
		assert.Nil(t, result, "Result should be nil on error")
	})
}

// TestIntegration_BasicUIInteraction tests basic UI interactions if simulator is available
func TestIntegration_BasicUIInteraction(t *testing.T) {
	_, devices := setupTestEnvironment(t)
	if devices == nil {
		return // Test was skipped
	}

	// Find a booted simulator
	bootedSim := findBootedSimulator(devices)
	if bootedSim == nil {
		t.Skip("No booted simulators available for UI interaction test. Please boot a simulator manually.")
		return
	}

	bridge := xcrun.NewBridge()
	t.Logf("Testing UI interactions with device: %s (%s)", bootedSim.Name, bootedSim.UDID)

	t.Run("type text into simulator", func(t *testing.T) {
		testText := "Hello from integration test"

		result, err := bridge.TypeText(bootedSim.UDID, testText)

		// Note: keyboardinput command may not be available on all Xcode versions
		// If it's not available, we skip this test rather than fail
		if err != nil && strings.Contains(err.Error(), "Unrecognized subcommand: keyboardinput") {
			t.Skip("keyboardinput command not available on this Xcode version")
			return
		}

		require.NoError(t, err, "Should type text successfully")
		require.NotNil(t, result, "Text input result should not be nil")

		// Verify result metadata
		assert.Equal(t, testText, result.Text, "Result text should match input")
		assert.Equal(t, len(testText), result.Length, "Result length should match input length")
		assert.Equal(t, bootedSim.UDID, result.DeviceID, "Result device ID should match")
		assert.NotEmpty(t, result.Timestamp, "Timestamp should be set")

		t.Logf("Typed text: %s (%d characters)", testText, len(testText))
	})

	t.Run("type special characters", func(t *testing.T) {
		// Test typing various special characters
		testCases := []string{
			"test@example.com",
			"Password123!",
			"Line1\nLine2", // newline is tricky
		}

		for _, testText := range testCases {
			// For newline, xcrun simctl may not handle it well, so we expect potential failures
			result, err := bridge.TypeText(bootedSim.UDID, testText)

			// Skip if keyboardinput is not available
			if err != nil && strings.Contains(err.Error(), "Unrecognized subcommand: keyboardinput") {
				t.Skip("keyboardinput command not available on this Xcode version")
				return
			}

			if strings.Contains(testText, "\n") {
				// Newline handling is simulator/OS dependent, so we're lenient
				t.Logf("Typing text with newline - result: %v, err: %v", result, err)
			} else {
				require.NoError(t, err, "Should type text with special chars: %s", testText)
				assert.Equal(t, testText, result.Text, "Result should match input")
			}
		}
	})

	t.Run("press home button", func(t *testing.T) {
		result, err := bridge.PressButton(bootedSim.UDID, "HOME")

		// Skip if ui click command is not available
		if err != nil && (strings.Contains(err.Error(), "Get or Set UI options") ||
			strings.Contains(err.Error(), "Unrecognized subcommand")) {
			t.Skip("ui click home command not available on this Xcode version")
			return
		}

		require.NoError(t, err, "Should press HOME button successfully")
		require.NotNil(t, result, "Button result should not be nil")

		assert.Equal(t, "HOME", result.Button, "Result button should be HOME")
		assert.Equal(t, bootedSim.UDID, result.DeviceID, "Result device ID should match")
		assert.NotEmpty(t, result.Timestamp, "Timestamp should be set")

		t.Log("HOME button pressed successfully")
	})

	t.Run("press invalid button returns error", func(t *testing.T) {
		result, err := bridge.PressButton(bootedSim.UDID, "INVALID_BUTTON")
		assert.Error(t, err, "Should return error for invalid button")
		assert.Nil(t, result, "Result should be nil on error")
		assert.Contains(t, err.Error(), "unsupported button", "Error should indicate unsupported button")
	})
}

// TestIntegration_DeviceStatePolling tests state polling with various scenarios
func TestIntegration_DeviceStatePolling(t *testing.T) {
	manager, devices := setupTestEnvironment(t)
	if manager == nil {
		return // Test was skipped
	}

	if len(devices) == 0 {
		t.Skip("No devices available for state polling test")
		return
	}

	t.Run("get device state for existing device", func(t *testing.T) {
		targetDevice := devices[0]

		state, err := manager.GetDeviceState(targetDevice.ID)
		require.NoError(t, err, "Should get device state")
		assert.NotEmpty(t, state, "State should not be empty")

		// Verify state is valid
		validStates := []device.DeviceState{
			device.StateBooted,
			device.StateShutdown,
			device.StateCreating,
			device.StateBooting,
			device.StateShuttingDown,
		}
		assert.Contains(t, validStates, state, "State should be valid")

		t.Logf("Device %s state: %s", targetDevice.Name, state)
	})

	t.Run("get device state for nonexistent device", func(t *testing.T) {
		state, err := manager.GetDeviceState("nonexistent-device-id-99999")
		assert.Error(t, err, "Should return error for nonexistent device")
		assert.Empty(t, state, "State should be empty on error")
		assert.Contains(t, err.Error(), "device not found", "Error should indicate device not found")
	})

	t.Run("poll device states for all devices", func(t *testing.T) {
		// Poll states for all devices multiple times to ensure consistency
		iterations := 3
		pollInterval := 500 * time.Millisecond

		for i := 0; i < iterations; i++ {
			t.Logf("Polling iteration %d/%d", i+1, iterations)

			for _, dev := range devices {
				state, err := manager.GetDeviceState(dev.ID)
				require.NoError(t, err, "Should get device state for %s", dev.Name)
				t.Logf("  Device %s: %s", dev.Name, state)
			}

			if i < iterations-1 {
				time.Sleep(pollInterval)
			}
		}
	})
}

// TestIntegration_ConcurrentDeviceOperations tests that device operations are safe with concurrent access
func TestIntegration_ConcurrentDeviceOperations(t *testing.T) {
	manager, devices := setupTestEnvironment(t)
	if manager == nil {
		return // Test was skipped
	}

	if len(devices) == 0 {
		t.Skip("No devices available for concurrent operations test")
		return
	}

	t.Run("concurrent device list operations", func(t *testing.T) {
		// Spawn multiple goroutines that list devices concurrently
		concurrency := 5
		iterations := 10

		done := make(chan bool, concurrency)
		errors := make(chan error, concurrency*iterations)

		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				defer func() {
					done <- true
				}()

				for j := 0; j < iterations; j++ {
					devs, err := manager.ListDevices()
					if err != nil {
						errors <- fmt.Errorf("worker %d iteration %d: %w", workerID, j, err)
						continue
					}

					if len(devs) == 0 {
						errors <- fmt.Errorf("worker %d iteration %d: no devices returned", workerID, j)
					}
				}
			}(i)
		}

		// Wait for all workers to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}
		close(errors)

		// Check if any errors occurred
		var errorList []error
		for err := range errors {
			errorList = append(errorList, err)
		}

		if len(errorList) > 0 {
			t.Logf("Encountered %d errors during concurrent operations:", len(errorList))
			for _, err := range errorList {
				t.Logf("  - %v", err)
			}
			t.Fail()
		} else {
			t.Logf("Successfully completed %d concurrent operations without errors", concurrency*iterations)
		}
	})

	t.Run("concurrent device get operations", func(t *testing.T) {
		if len(devices) == 0 {
			t.Skip("No devices available")
			return
		}

		targetDevice := devices[0]
		concurrency := 10

		done := make(chan bool, concurrency)
		errors := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				defer func() {
					done <- true
				}()

				dev, err := manager.GetDevice(targetDevice.ID)
				if err != nil {
					errors <- fmt.Errorf("worker %d: %w", workerID, err)
					return
				}

				if dev.ID != targetDevice.ID {
					errors <- fmt.Errorf("worker %d: device ID mismatch", workerID)
				}
			}(i)
		}

		// Wait for all workers
		for i := 0; i < concurrency; i++ {
			<-done
		}
		close(errors)

		// Check errors
		var errorList []error
		for err := range errors {
			errorList = append(errorList, err)
		}

		assert.Empty(t, errorList, "Should have no errors during concurrent get operations")
	})
}

// TestIntegration_ErrorHandling tests various error scenarios
func TestIntegration_ErrorHandling(t *testing.T) {
	bridge := xcrun.NewBridge()

	t.Run("boot nonexistent simulator", func(t *testing.T) {
		err := bridge.BootSimulator("00000000-0000-0000-0000-000000000000")
		assert.Error(t, err, "Should return error for nonexistent simulator")
		assert.Contains(t, err.Error(), "failed to boot", "Error should indicate boot failure")
	})

	t.Run("shutdown nonexistent simulator", func(t *testing.T) {
		err := bridge.ShutdownSimulator("00000000-0000-0000-0000-000000000000")
		assert.Error(t, err, "Should return error for nonexistent simulator")
		assert.Contains(t, err.Error(), "failed to shutdown", "Error should indicate shutdown failure")
	})

	t.Run("get state of nonexistent simulator", func(t *testing.T) {
		state, err := bridge.GetDeviceState("00000000-0000-0000-0000-000000000000")
		assert.Error(t, err, "Should return error for nonexistent simulator")
		assert.Empty(t, state, "State should be empty")
		assert.Contains(t, err.Error(), "device not found", "Error should indicate device not found")
	})

	t.Run("screenshot nonexistent simulator", func(t *testing.T) {
		tempDir := t.TempDir()
		screenshotPath := filepath.Join(tempDir, "test.png")

		result, err := bridge.CaptureScreenshot("00000000-0000-0000-0000-000000000000", screenshotPath)
		assert.Error(t, err, "Should return error for nonexistent simulator")
		assert.Nil(t, result, "Result should be nil")
	})

	t.Run("type text to nonexistent simulator", func(t *testing.T) {
		result, err := bridge.TypeText("00000000-0000-0000-0000-000000000000", "test")

		// Skip if keyboardinput is not available
		if err != nil && strings.Contains(err.Error(), "Unrecognized subcommand: keyboardinput") {
			t.Skip("keyboardinput command not available on this Xcode version")
			return
		}

		assert.Error(t, err, "Should return error for nonexistent simulator")
		assert.Nil(t, result, "Result should be nil")
	})
}
