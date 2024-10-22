package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"
var revision = "none"

var rootCmd = &cobra.Command{
	Use:     "kubectl-sg4pod",
	Version: version + "-" + revision,
	Short:   "A kubectl plugin to view pods and security groups, eni.",
	Long: `kubectl-sg4pod is a kubectl plugin that allows you to view detailed information about pods,
security groups, and ENIs (Elastic Network Interfaces) in your Kubernetes cluster.

This plugin provides various commands to help you inspect and manage your cluster's networking and security configurations.
Ensure you are using the latest version to access new features and bug fixes.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(getCmd())
}
