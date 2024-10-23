package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kubectl-sgmap",
		Long: `The 'version' command displays the current version of the kubectl-sgmap plugin.
It helps in tracking the version you're using, especially when updating or debugging.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nRevision: %s\n", version, revision)
		},
	}
}
