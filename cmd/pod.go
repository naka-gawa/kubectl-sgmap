package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/naka-gawa/kubectl-sgmap/internal/usecase"
)

// NewPodCommand creates the pod subcommand
func NewPodCommand(streams *genericclioptions.IOStreams) *cobra.Command {
	o := usecase.NewPodOptions(streams)
	cmd := &cobra.Command{
		Use:     "pod [NAME]",
		Aliases: []string{"pods", "po"},
		Short:   "Display security group information for pods",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				o.PodName = args[0]
			}

			return o.Run(cmd.Context())
		},
	}

	// フラグの設定
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace of the pod")
	cmd.Flags().StringVar(&o.OutputFormat, "output", "", "output format (json|yaml|table)")
	cmd.Flags().BoolVar(&o.AllNamespaces, "all-namespaces", false, "search pods in all namespaces")

	return cmd
}
