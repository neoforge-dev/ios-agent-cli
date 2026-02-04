// +build integration

package xcrun

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCaptureScreenshot_Integration tests the screenshot capture functionality
// with a real simulator. Run with: go test -tags=integration ./pkg/xcrun/
func TestCaptureScreenshot_Integration(t *testing.T) {
	bridge := NewBridge()

	// Get list of devices
	devices, err := bridge.ListDevices()
	require.NoError(t, err, "failed to list devices")
	require.NotEmpty(t, devices, "no devices available for testing")

	// Find a booted device
	var bootedDevice string
	for _, dev := range devices {
		if dev.State == "Booted" {
			bootedDevice = dev.UDID
			break
		}
	}

	if bootedDevice == "" {
		t.Skip("No booted simulator available for screenshot test")
	}

	// Test PNG screenshot with default path
	t.Run("capture PNG screenshot", func(t *testing.T) {
		outputPath := filepath.Join(os.TempDir(), "test-screenshot-integration.png")
		defer os.Remove(outputPath)

		result, err := bridge.CaptureScreenshot(bootedDevice, outputPath)
		require.NoError(t, err, "failed to capture screenshot")

		assert.Equal(t, outputPath, result.Path)
		assert.Equal(t, "png", result.Format)
		assert.Greater(t, result.SizeBytes, int64(0), "screenshot file should have non-zero size")
		assert.Equal(t, bootedDevice, result.DeviceID)
		assert.NotEmpty(t, result.Timestamp)

		// Verify file exists
		fileInfo, err := os.Stat(outputPath)
		require.NoError(t, err, "screenshot file should exist")
		assert.Equal(t, result.SizeBytes, fileInfo.Size())
	})

	// Test JPEG screenshot
	t.Run("capture JPEG screenshot", func(t *testing.T) {
		outputPath := filepath.Join(os.TempDir(), "test-screenshot-integration.jpg")
		defer os.Remove(outputPath)

		result, err := bridge.CaptureScreenshot(bootedDevice, outputPath)
		require.NoError(t, err, "failed to capture screenshot")

		assert.Equal(t, outputPath, result.Path)
		assert.Equal(t, "jpeg", result.Format)
		assert.Greater(t, result.SizeBytes, int64(0), "screenshot file should have non-zero size")
	})

	// Test error case with invalid device
	t.Run("error on invalid device", func(t *testing.T) {
		outputPath := filepath.Join(os.TempDir(), "test-screenshot-invalid.png")
		defer os.Remove(outputPath)

		_, err := bridge.CaptureScreenshot("invalid-device-id", outputPath)
		assert.Error(t, err, "should fail with invalid device ID")
	})
}
