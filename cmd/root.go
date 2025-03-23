package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	version  string
	revision string
	streams  = genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
)

// Execute executes the root command
func Execute() error {
	rootCmd := NewSgmapCommand(&streams)
	rootCmd.AddCommand(newVersionCommand())

	return rootCmd.Execute()
}

// newVersionCommand creates the version command
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kubectl-sgmap",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("kubectl-sgmap version %s, revision %s\n", version, revision)
		},
	}
}
