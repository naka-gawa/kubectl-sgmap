package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewSgmapCommand creates the sgmap command
func NewSgmapCommand(streams *genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sgmap",
		Short: "Display security group information for Kubernetes workloads",
		Long:  `Display security group information for Kubernetes workloads running on AWS`,
	}

	cmd.AddCommand(NewPodCommand(streams))
	return cmd
}
