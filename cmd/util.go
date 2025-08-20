package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const pluginPrefix = "kubectl-"

func isKubectlPlugin() bool {
	return strings.HasPrefix(filepath.Base(os.Args[0]), pluginPrefix)
}

func setKubectlPluginName(rootCmd *cobra.Command, name string) {
	if isKubectlPlugin() {
		rootCmd.Use = name
	}
}
