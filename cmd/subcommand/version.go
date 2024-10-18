package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kubectl-view-podsg",
		Long: `The 'version' command displays the current version of the kubectl-view-podsg plugin.
It helps in tracking the version you're using, especially when updating or debugging.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nRevision: %s\n", version, revision)
		},
	}
}
