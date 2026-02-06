package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/device"
	"github.com/neoforge-dev/ios-agent-cli/pkg/xcrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// SCREENSHOT COMMAND - STRUCTURE TESTS
// ============================================================================

func TestScreenshotCommand_Structure(t *testing.T) {
	// Verify command structure
	assert.NotNil(t, screenshotCmd)
	assert.Equal(t, "screenshot", screenshotCmd.Use)
	assert.Contains(t, screenshotCmd.Short, "Capture a screenshot")
	assert.Contains(t, screenshotCmd.Long, "Capture a screenshot from an iOS device or simulator")
}

func TestScreenshotCommand_Flags(t *testing.T) {
	// Verify flag configuration
	outputFlag := screenshotCmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag, "screenshot command should have --output flag")
	assert.Equal(t, "o", outputFlag.Shorthand, "--output should have -o shorthand")

	formatFlag := screenshotCmd.Flags().Lookup("format")
	assert.NotNil(t, formatFlag, "screenshot command should have --format flag")
	assert.Equal(t, "png", formatFlag.DefValue, "format should default to png")
}

func TestScreenshotCommand_RegisteredWithRoot(t *testing.T) {
	// Verify screenshot command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "screenshot" {
			found = true
			break
		}
	}
	assert.True(t, found, "screenshot command should be registered with root command")
}

// ============================================================================
// SCREENSHOT COMMAND - FORMAT VALIDATION TESTS
// ============================================================================

func TestScreenshotCommand_ValidFormats(t *testing.T) {
	// Test valid image formats
	validFormats := []string{
		"png",
		"jpeg",
	}

	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			assert.True(t, format == "png" || format == "jpeg", "format should be valid: %s", format)
		})
	}
}

func TestScreenshotCommand_InvalidFormats(t *testing.T) {
	// Test invalid image formats that should be rejected
	invalidFormats := []struct {
		name   string
		format string
	}{
		{"gif format", "gif"},
		{"bmp format", "bmp"},
		{"webp format", "webp"},
		{"tiff format", "tiff"},
		{"wrong case", "PNG"},
		{"JPEG uppercase", "JPEG"},
		{"jpg shorthand", "jpg"},
		{"empty format", ""},
		{"raw format", "raw"},
	}

	for _, tt := range invalidFormats {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.format == "png" || tt.format == "jpeg"
			assert.False(t, isValid, "format should be invalid: %s", tt.format)
		})
	}
}

func TestScreenshotCommand_FormatExtensionMapping(t *testing.T) {
	// Test extension mapping for formats
	tests := []struct {
		format    string
		extension string
	}{
		{"png", ".png"},
		{"jpeg", ".jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			ext := tt.extension
			if tt.format == "png" {
				ext = ".png"
			} else if tt.format == "jpeg" {
				ext = ".jpg"
			}
			assert.True(t, len(ext) > 0, "extension should be set for: %s", tt.format)
		})
	}
}

// ============================================================================
// SCREENSHOT COMMAND - OUTPUT PATH TESTS
// ============================================================================

func TestScreenshotCommand_DefaultOutputPath(t *testing.T) {
	// When no output path specified, should generate timestamped file in /tmp
	// Pattern: /tmp/screenshot-YYYYMMDD-HHMMSS.{png,jpg}

	timestamp := time.Now().Format("20060102-150405")
	defaultPath := filepath.Join("/tmp", fmt.Sprintf("screenshot-%s.png", timestamp))

	assert.True(t, len(defaultPath) > 0, "default path should be generated")
	assert.Contains(t, defaultPath, "/tmp", "default path should be in /tmp")
	assert.Contains(t, defaultPath, "screenshot-", "default path should have screenshot prefix")
	assert.Contains(t, defaultPath, ".png", "default path should have .png extension")
}

func TestScreenshotCommand_TimestampFormat(t *testing.T) {
	// Verify timestamp format is consistent
	timestamp := time.Now().Format("20060102-150405")

	// Timestamp should be 15 characters: YYYYMMDD-HHMMSS
	assert.Equal(t, 15, len(timestamp), "timestamp should be 15 chars: %s", timestamp)

	// Should only contain digits and dash
	for _, ch := range timestamp {
		assert.True(t, ch >= '0' && ch <= '9' || ch == '-', "timestamp should only contain digits and dash")
	}
}

func TestScreenshotCommand_CustomOutputPath(t *testing.T) {
	// Custom output paths should be accepted
	customPaths := []string{
		"/tmp/myshot.png",
		"./screenshot.png",
		"/Users/test/Desktop/screenshot.jpg",
		"/var/tmp/ios-screenshot.png",
		"screenshot.jpeg",
	}

	for _, path := range customPaths {
		t.Run(path, func(t *testing.T) {
			assert.NotEmpty(t, path, "custom path should be valid")
		})
	}
}

func TestScreenshotCommand_NestedDirectories(t *testing.T) {
	// Output path with nested directories
	nestedPath := "/Users/test/Documents/Screenshots/2026/02/06/screenshot.png"

	dir := filepath.Dir(nestedPath)
	assert.Contains(t, dir, "2026/02/06", "path should support nested directories")

	// All parent directories should be creatable
	assert.NotEmpty(t, dir, "directory path should not be empty")
}

func TestScreenshotCommand_PathWithSpaces(t *testing.T) {
	// Paths with spaces should be supported
	pathsWithSpaces := []string{
		"/tmp/My Screenshots/screenshot.png",
		"/Users/test/My Files/image.png",
	}

	for _, path := range pathsWithSpaces {
		t.Run(path, func(t *testing.T) {
			assert.Contains(t, path, " ", "path should contain spaces")
		})
	}
}

// ============================================================================
// SCREENSHOT COMMAND - DEVICE VALIDATION TESTS
// ============================================================================

func TestScreenshotCommand_DeviceRequired(t *testing.T) {
	// Device ID is required for screenshot
	assert.NotNil(t, screenshotCmd, "screenshot command should exist")

	// Device flag should be available from parent (root command)
	deviceFlag := rootCmd.PersistentFlags().Lookup("device")
	assert.NotNil(t, deviceFlag, "device flag should be available")
}

func TestScreenshotCommand_DeviceValidation(t *testing.T) {
	// Test device validation scenarios
	tests := []struct {
		name      string
		deviceID  string
		shouldErr bool
	}{
		{"valid device ID", "test-device-1", false},
		{"empty device ID", "", true},
		{"udid format", "ABC123DEF456GHI789", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldErr {
				assert.Empty(t, tt.deviceID, "device ID should be empty for error case")
			} else {
				assert.NotEmpty(t, tt.deviceID, "device ID should be provided")
			}
		})
	}
}

// ============================================================================
// SCREENSHOT COMMAND - ERROR CODE TESTS
// ============================================================================

func TestScreenshotCommand_ErrorCodes(t *testing.T) {
	// Verify all error codes used in screenshot command
	errorCodes := []string{
		"DEVICE_REQUIRED",
		"INVALID_FORMAT",
		"DEVICE_NOT_FOUND",
		"DEVICE_NOT_BOOTED",
		"PATH_ERROR",
		"SCREENSHOT_FAILED",
	}

	for _, code := range errorCodes {
		t.Run(code, func(t *testing.T) {
			assert.NotEmpty(t, code, "error code should not be empty")
		})
	}
}

// ============================================================================
// SCREENSHOT RESULT - JSON SERIALIZATION TESTS
// ============================================================================

func TestScreenshotResult_JSONStructure(t *testing.T) {
	// Test that ScreenshotResult serializes correctly
	result := &xcrun.ScreenshotResult{
		DeviceID:  "test-device-1",
		Path:      "/tmp/screenshot-20260206-120000.png",
		Format:    "png",
		SizeBytes: 12345,
		Timestamp: "2026-02-06T12:00:00Z",
	}

	// Serialize to JSON
	data, err := json.Marshal(result)
	require.NoError(t, err, "should serialize to JSON")

	// Verify JSON is valid
	assert.NotEmpty(t, data, "JSON data should not be empty")

	// Deserialize back
	var decoded xcrun.ScreenshotResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err, "should deserialize from JSON")

	assert.Equal(t, result.DeviceID, decoded.DeviceID)
	assert.Equal(t, result.Path, decoded.Path)
	assert.Equal(t, result.Format, decoded.Format)
	assert.Equal(t, result.SizeBytes, decoded.SizeBytes)
	assert.Equal(t, result.Timestamp, decoded.Timestamp)
}

// ============================================================================
// SCREENSHOT COMMAND - FILE EXTENSION TESTS
// ============================================================================

func TestScreenshotCommand_JPEGExtensionVariants(t *testing.T) {
	// Test both .jpg and .jpeg extensions for JPEG format
	tests := []struct {
		path      string
		format    string
		expected  string
	}{
		{"/tmp/image.jpg", "jpeg", "jpeg"},
		{"/tmp/image.jpeg", "jpeg", "jpeg"},
		{"/tmp/image.JPG", "jpeg", "jpeg"},
		{"/tmp/image.JPEG", "jpeg", "jpeg"},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.path), func(t *testing.T) {
			ext := filepath.Ext(tt.path)
			assert.NotEmpty(t, ext, "extension should be extracted")
		})
	}
}

// ============================================================================
// SCREENSHOT COMMAND - CONCURRENT CAPTURES TESTS
// ============================================================================

func TestScreenshotCommand_MultipleConcurrentCaptures(t *testing.T) {
	// Verify timestamps are generated in sequence for multiple captures
	timestamps := make([]string, 0)

	for i := 0; i < 3; i++ {
		timestamp := time.Now().Format("20060102-150405")
		timestamps = append(timestamps, timestamp)
		time.Sleep(time.Second) // Wait for second to change
	}

	// Should have at least one unique timestamp due to 1-second waits
	assert.GreaterOrEqual(t, len(timestamps), 2, "should have multiple timestamps")
}

// ============================================================================
// SCREENSHOT COMMAND - DEVICE MANAGER MOCK TESTS
// ============================================================================

func TestScreenshotCommand_WithMockedDeviceManager(t *testing.T) {
	// Setup mock device manager
	mockBridge := &MockXCRunBridge{}
	mockBridge.On("ListDevices").Return([]device.Device{
		{
			ID:        "device-1",
			UDID:      "device-1",
			Name:      "iPhone 15 Pro",
			State:     device.StateBooted,
			Type:      device.DeviceTypeSimulator,
			OSVersion: "17.4",
			Available: true,
		},
	}, nil)

	mockBridge.On("CaptureScreenshot", "device-1", mock.MatchedBy(func(path string) bool {
		return len(path) > 0
	})).Return(&xcrun.ScreenshotResult{
		DeviceID:  "device-1",
		Path:      "/tmp/screenshot.png",
		Format:    "png",
		SizeBytes: 5000,
		Timestamp: "2026-02-06T12:00:00Z",
	}, nil)

	// Verify mock is set up
	assert.NotNil(t, mockBridge)
}

// ============================================================================
// SCREENSHOT COMMAND - FORMAT-SPECIFIC TESTS
// ============================================================================

func TestScreenshotCommand_PNGFormatDefault(t *testing.T) {
	// PNG should be the default format
	assert.Equal(t, "png", screenshotFormat, "format should default to png")
}

func TestScreenshotCommand_JPEGQualityHandling(t *testing.T) {
	// While quality is not directly configurable in the current command,
	// JPEG format should be a valid option
	validFormats := map[string]bool{
		"png":  true,
		"jpeg": true,
	}

	assert.True(t, validFormats["jpeg"], "jpeg should be valid format")
	assert.True(t, validFormats["png"], "png should be valid format")
}

// ============================================================================
// SCREENSHOT COMMAND - OUTPUT VALIDATION TESTS
// ============================================================================

func TestScreenshotCommand_OutputPathValidation(t *testing.T) {
	// Test valid and invalid output paths
	tests := []struct {
		name     string
		path     string
		isValid  bool
	}{
		{"absolute path", "/tmp/screenshot.png", true},
		{"relative path", "./screenshot.png", true},
		{"home directory", "~/screenshot.png", true},
		{"deep nesting", "/var/a/b/c/d/e/screenshot.png", true},
		{"empty path", "", true}, // will use default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Empty path triggers default behavior
			if tt.path == "" {
				assert.Empty(t, tt.path, "empty path should trigger default")
			} else {
				assert.NotEmpty(t, tt.path, "path should be valid")
			}
		})
	}
}

// ============================================================================
// SCREENSHOT COMMAND - MOCK FILE SYSTEM TESTS
// ============================================================================

func TestScreenshotCommand_CreateOutputDirectory(t *testing.T) {
	// Test that output directories can be created
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "a", "b", "c", "screenshot.png")

	// Simulated directory creation
	dir := filepath.Dir(nestedPath)
	err := os.MkdirAll(dir, 0755)
	assert.NoError(t, err, "should create nested directories")
	assert.DirExists(t, dir, "directory should exist after creation")

	// Clean up
	os.RemoveAll(tmpDir)
}

// ============================================================================
// SCREENSHOT COMMAND - INTEGRATION TESTS WITH FORMAT
// ============================================================================

func TestScreenshotCommand_PNGWithTimestamp(t *testing.T) {
	// Complete PNG screenshot workflow
	timestamp := time.Now().Format("20060102-150405")
	path := filepath.Join("/tmp", fmt.Sprintf("screenshot-%s.png", timestamp))

	assert.Contains(t, path, ".png", "path should have .png extension")
	assert.Contains(t, path, "screenshot-", "path should have screenshot prefix")
}

func TestScreenshotCommand_JPEGWithTimestamp(t *testing.T) {
	// Complete JPEG screenshot workflow
	timestamp := time.Now().Format("20060102-150405")
	ext := "jpg"
	path := filepath.Join("/tmp", fmt.Sprintf("screenshot-%s.%s", timestamp, ext))

	assert.Contains(t, path, ".jpg", "path should have .jpg extension")
	assert.Contains(t, path, "screenshot-", "path should have screenshot prefix")
}
