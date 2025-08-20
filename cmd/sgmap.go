package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewSgmapCommand creates the sgmap command
func NewSgmapCommand(streams *genericclioptions.IOStreams) *cobra.Command {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)

	cmd := &cobra.Command{
		Use:   "sgmap",
		Short: "Display security group information for Kubernetes workloads",
		Long:  `Display security group information for Kubernetes workloads running on AWS`,
	}

	kubeConfigFlags.AddFlags(cmd.PersistentFlags())

	cmd.AddCommand(NewPodCommand(streams, kubeConfigFlags))
	return cmd
}
