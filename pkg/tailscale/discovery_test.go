package tailscale

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiscoverMachines tests the machine discovery functionality
func TestDiscoverMachines(t *testing.T) {
	// Skip if tailscale is not installed
	if !isTailscaleInstalled() {
		t.Skip("Tailscale is not installed, skipping test")
	}

	machines, err := DiscoverMachines()

	// The test should either succeed or fail with a known error
	if err != nil {
		// If tailscale is installed but not connected, that's acceptable
		assert.Contains(t, err.Error(), "tailscale")
		return
	}

	// If we got results, validate them
	require.NoError(t, err)
	assert.NotNil(t, machines)

	// If connected to Tailscale, we should have at least one machine (self)
	if len(machines) > 0 {
		// Validate first machine has required fields
		machine := machines[0]
		assert.NotEmpty(t, machine.Name, "Machine should have a name")
		assert.NotEmpty(t, machine.IP, "Machine should have an IP")
		assert.NotEmpty(t, machine.TailscaleIP, "Machine should have a Tailscale IP")
	}
}

// TestGetMachineByName tests finding a machine by name
func TestGetMachineByName(t *testing.T) {
	if !isTailscaleInstalled() {
		t.Skip("Tailscale is not installed, skipping test")
	}

	machines, err := DiscoverMachines()
	if err != nil || len(machines) == 0 {
		t.Skip("No Tailscale machines available, skipping test")
	}

	// Test with first machine's name
	testName := machines[0].Name
	machine, err := GetMachineByName(testName)

	require.NoError(t, err)
	require.NotNil(t, machine)
	assert.Equal(t, testName, machine.Name)
}

// TestGetMachineByIP tests finding a machine by IP
func TestGetMachineByIP(t *testing.T) {
	if !isTailscaleInstalled() {
		t.Skip("Tailscale is not installed, skipping test")
	}

	machines, err := DiscoverMachines()
	if err != nil || len(machines) == 0 {
		t.Skip("No Tailscale machines available, skipping test")
	}

	// Test with first machine's IP
	testIP := machines[0].TailscaleIP
	machine, err := GetMachineByIP(testIP)

	require.NoError(t, err)
	require.NotNil(t, machine)
	assert.Equal(t, testIP, machine.TailscaleIP)
}

// TestGetMachineByName_NotFound tests error handling for unknown machine
func TestGetMachineByName_NotFound(t *testing.T) {
	if !isTailscaleInstalled() {
		t.Skip("Tailscale is not installed, skipping test")
	}

	machine, err := GetMachineByName("nonexistent-machine-12345")
	assert.Error(t, err)
	assert.Nil(t, machine)
	assert.Contains(t, err.Error(), "machine not found")
}

// TestProbeForIOSAgent tests the probe function
func TestProbeForIOSAgent(t *testing.T) {
	// For MVP, this always returns false
	result := ProbeForIOSAgent("100.64.0.1")
	assert.False(t, result, "ProbeForIOSAgent should return false in MVP")
}

// TestTailscaleStatusParsing tests JSON parsing of tailscale status
func TestTailscaleStatusParsing(t *testing.T) {
	// Mock tailscale status JSON output
	mockJSON := `{
		"Self": {
			"ID": "n12345",
			"PublicKey": "key123",
			"HostName": "test-machine",
			"DNSName": "test-machine.example.ts.net",
			"OS": "macOS",
			"UserID": 1,
			"TailscaleIPs": ["100.64.0.1"],
			"Online": true,
			"Active": true
		},
		"Peer": {
			"peer1": {
				"ID": "n67890",
				"PublicKey": "key456",
				"HostName": "remote-machine",
				"DNSName": "remote-machine.example.ts.net",
				"OS": "linux",
				"UserID": 1,
				"TailscaleIPs": ["100.64.0.2"],
				"Online": true,
				"Active": true
			}
		},
		"User": {
			"1": {
				"ID": 1,
				"LoginName": "user@example.com",
				"DisplayName": "Test User"
			}
		}
	}`

	var status TailscaleStatus
	err := json.Unmarshal([]byte(mockJSON), &status)

	require.NoError(t, err)
	assert.Equal(t, "test-machine", status.Self.HostName)
	assert.Equal(t, "100.64.0.1", status.Self.TailscaleIPs[0])
	assert.Len(t, status.Peer, 1)

	// Validate peer
	for _, peer := range status.Peer {
		assert.Equal(t, "remote-machine", peer.HostName)
		assert.Equal(t, "100.64.0.2", peer.TailscaleIPs[0])
		assert.True(t, peer.Online)
	}
}

// TestIsTailscaleInstalled tests the installation check
func TestIsTailscaleInstalled(t *testing.T) {
	result := isTailscaleInstalled()

	// Check if which command can find tailscale
	cmd := exec.Command("which", "tailscale")
	err := cmd.Run()
	expectedResult := (err == nil)

	assert.Equal(t, expectedResult, result, "isTailscaleInstalled should match which command result")
}

// BenchmarkDiscoverMachines benchmarks the discovery function
func BenchmarkDiscoverMachines(b *testing.B) {
	if !isTailscaleInstalled() {
		b.Skip("Tailscale is not installed, skipping benchmark")
	}

	for i := 0; i < b.N; i++ {
		_, _ = DiscoverMachines()
	}
}
