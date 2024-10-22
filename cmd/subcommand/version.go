package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kubectl-sg4pod",
		Long: `The 'version' command displays the current version of the kubectl-sg4pod plugin.
It helps in tracking the version you're using, especially when updating or debugging.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nRevision: %s\n", version, revision)
		},
	}
}
