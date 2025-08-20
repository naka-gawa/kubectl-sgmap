package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/naka-gawa/kubectl-sgmap/internal/usecase"
)

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

	cmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "", "output format (json|yaml|table)")
	cmd.Flags().BoolVarP(o.AllNamespaces, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	o.ConfigFlags.AddFlags(cmd.Flags())

	return cmd
}
