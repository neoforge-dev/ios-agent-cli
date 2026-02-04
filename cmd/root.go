package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/neoforge-dev/ios-agent-cli/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	deviceID   string
	remoteHost string
	verbose    bool
	format     string
)

// Response is the standard JSON response wrapper
type Response struct {
	Success   bool        `json:"success"`
	Action    string      `json:"action,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "ios-agent",
	Short: "AI-agent-friendly iOS automation CLI",
	Long: `iOS Agent CLI enables AI agents to automate iOS app testing
on local simulators and remote devices over Tailscale.

All commands return JSON for easy parsing by agents.

Examples:
  ios-agent devices                       # List available devices
  ios-agent simulator boot --name "iPhone 15"  # Boot a simulator
  ios-agent app launch --device <id> --bundle com.example.app
  ios-agent screenshot --device <id> --output ./shot.png
  ios-agent io tap --device <id> --x 100 --y 200`,
	Version: "0.1.0",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&deviceID, "device", "d", "", "Device ID to target")
	rootCmd.PersistentFlags().StringVar(&remoteHost, "remote-host", "", "Remote host:port for remote device control")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format (json)")
}

// outputJSON prints the response as JSON
func outputJSON(resp Response) {
	resp.Timestamp = time.Now().UTC().Format(time.RFC3339)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

// outputSuccess outputs a successful response
func outputSuccess(action string, result interface{}) {
	outputJSON(Response{
		Success: true,
		Action:  action,
		Result:  result,
	})
}

// outputError outputs an error response
// Deprecated: Use outputAgentError instead
func outputError(action, code, message string, details interface{}) {
	outputJSON(Response{
		Success: false,
		Action:  action,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
	os.Exit(1)
}

// outputAgentError outputs a standardized error response using AgentError
func outputAgentError(action string, err *errors.AgentError) {
	outputJSON(Response{
		Success: false,
		Action:  action,
		Error: &ErrorInfo{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
		},
	})
	os.Exit(1)
}
