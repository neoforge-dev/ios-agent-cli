// +build integration

package integration

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Response represents the standard JSON response wrapper
type Response struct {
	Success   bool                   `json:"success"`
	Action    string                 `json:"action,omitempty"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     *ErrorInfo             `json:"error,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Device represents a device in the response
type Device struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	Type      string `json:"type"`
	OSVersion string `json:"os_version"`
	UDID      string `json:"udid"`
	Available bool   `json:"available"`
}

func TestDevicesCommand(t *testing.T) {
	// Run the devices command
	cmd := exec.Command("../../ios-agent", "devices")
	output, err := cmd.Output()
	require.NoError(t, err, "Command should execute successfully")

	// Parse JSON response
	var resp Response
	err = json.Unmarshal(output, &resp)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify response structure
	assert.True(t, resp.Success, "Response should be successful")
	assert.Equal(t, "devices.list", resp.Action, "Action should be 'devices.list'")
	assert.NotEmpty(t, resp.Timestamp, "Timestamp should be present")
	assert.Nil(t, resp.Error, "Error should be nil for successful response")

	// Verify result contains devices array
	assert.Contains(t, resp.Result, "devices", "Result should contain 'devices' key")

	// Parse devices from result
	devicesData, ok := resp.Result["devices"].([]interface{})
	require.True(t, ok, "Devices should be an array")

	// If there are devices, verify their structure
	if len(devicesData) > 0 {
		// Marshal and unmarshal to get proper types
		devicesJSON, err := json.Marshal(devicesData)
		require.NoError(t, err)

		var devices []Device
		err = json.Unmarshal(devicesJSON, &devices)
		require.NoError(t, err)

		// Verify first device has required fields
		device := devices[0]
		assert.NotEmpty(t, device.ID, "Device ID should not be empty")
		assert.NotEmpty(t, device.Name, "Device name should not be empty")
		assert.NotEmpty(t, device.State, "Device state should not be empty")
		assert.Equal(t, "simulator", device.Type, "Device type should be 'simulator'")
		assert.NotEmpty(t, device.OSVersion, "OS version should not be empty")
		assert.NotEmpty(t, device.UDID, "UDID should not be empty")
		assert.True(t, device.Available, "Available should be true for listed devices")
	}
}

func TestDevicesCommandEmptyResult(t *testing.T) {
	// This test verifies that if no simulators are available,
	// the command still returns success with an empty devices array
	// Note: This test may not fail on systems with simulators installed

	cmd := exec.Command("../../ios-agent", "devices")
	output, err := cmd.Output()
	require.NoError(t, err, "Command should execute successfully")

	var resp Response
	err = json.Unmarshal(output, &resp)
	require.NoError(t, err, "Response should be valid JSON")

	// Even with no devices, response should be successful
	assert.True(t, resp.Success, "Response should be successful")
	assert.Equal(t, "devices.list", resp.Action)
	assert.Contains(t, resp.Result, "devices", "Result should contain 'devices' key")

	devicesData, ok := resp.Result["devices"].([]interface{})
	require.True(t, ok, "Devices should be an array")
	assert.NotNil(t, devicesData, "Devices array should not be nil")
}

func TestDevicesCommandJSONFormat(t *testing.T) {
	// Verify that the JSON output is properly formatted and parseable
	cmd := exec.Command("../../ios-agent", "devices")
	output, err := cmd.Output()
	require.NoError(t, err)

	// Verify it's valid JSON
	var data interface{}
	err = json.Unmarshal(output, &data)
	assert.NoError(t, err, "Output should be valid JSON")

	// Verify we can pretty-print it
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	assert.NoError(t, err)
	assert.NotEmpty(t, prettyJSON)
}
