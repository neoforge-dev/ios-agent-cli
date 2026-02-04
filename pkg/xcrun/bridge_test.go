package xcrun

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractOSVersion(t *testing.T) {
	tests := []struct {
		name     string
		runtime  string
		expected string
	}{
		{
			name:     "iOS 17.4 runtime",
			runtime:  "com.apple.CoreSimulator.SimRuntime.iOS-17-4",
			expected: "17.4",
		},
		{
			name:     "iOS 16.0 runtime",
			runtime:  "com.apple.CoreSimulator.SimRuntime.iOS-16-0",
			expected: "16.0",
		},
		{
			name:     "iOS 17.2 runtime",
			runtime:  "com.apple.CoreSimulator.SimRuntime.iOS-17-2",
			expected: "17.2",
		},
		{
			name:     "iOS 15.5 runtime",
			runtime:  "com.apple.CoreSimulator.SimRuntime.iOS-15-5",
			expected: "15.5",
		},
		{
			name:     "watchOS runtime (no iOS)",
			runtime:  "com.apple.CoreSimulator.SimRuntime.watchOS-10-0",
			expected: "unknown",
		},
		{
			name:     "tvOS runtime (no iOS)",
			runtime:  "com.apple.CoreSimulator.SimRuntime.tvOS-17-0",
			expected: "unknown",
		},
		{
			name:     "malformed runtime",
			runtime:  "invalid.runtime.string",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOSVersion(tt.runtime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Note: Integration tests for ListDevices, BootSimulator, etc. should be in
// a separate integration test file that requires Xcode to be installed.
// These would be run with: go test -tags=integration
