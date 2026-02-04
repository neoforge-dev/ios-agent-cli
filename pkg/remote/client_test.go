package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRemoteClient(t *testing.T) {
	tests := []struct {
		name        string
		hostPort    string
		expectError bool
		expectHost  string
		expectPort  int
	}{
		{
			name:        "valid host only",
			hostPort:    "192.168.1.100",
			expectError: false,
			expectHost:  "192.168.1.100",
			expectPort:  22,
		},
		{
			name:        "valid host with port",
			hostPort:    "192.168.1.100:2222",
			expectError: false,
			expectHost:  "192.168.1.100",
			expectPort:  2222,
		},
		{
			name:        "valid hostname",
			hostPort:    "mac-mini.local",
			expectError: false,
			expectHost:  "mac-mini.local",
			expectPort:  22,
		},
		{
			name:        "valid hostname with port",
			hostPort:    "mac-mini.local:2222",
			expectError: false,
			expectHost:  "mac-mini.local",
			expectPort:  2222,
		},
		{
			name:        "empty host",
			hostPort:    "",
			expectError: true,
		},
		{
			name:        "invalid port",
			hostPort:    "host:invalid",
			expectError: true,
		},
		{
			name:        "colon only",
			hostPort:    ":",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewRemoteClient(tt.hostPort)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.expectHost, client.Host)
				assert.Equal(t, tt.expectPort, client.Port)
			}
		})
	}
}
