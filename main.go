package main

import (
	"os"

	"github.com/neoforge-dev/ios-agent-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
