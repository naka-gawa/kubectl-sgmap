package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const pluginPrefix = "kubectl-"

// isKubectlPlugin detects if the binary is being executed as a kubectl plugin
// by checking if the executable name starts with 'kubectl-'.
func isKubectlPlugin() bool {
	return strings.HasPrefix(filepath.Base(os.Args[0]), pluginPrefix)
}

// setKubectlPluginName conditionally sets the command name when running as a kubectl plugin
// to avoid showing the full 'kubectl-' prefix in help text.
func setKubectlPluginName(rootCmd *cobra.Command, name string) {
	if isKubectlPlugin() {
		rootCmd.Use = name
	}
}
