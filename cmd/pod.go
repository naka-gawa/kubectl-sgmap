// Package cmd provides the command-line interface for kubectl-sgmap.
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/naka-gawa/kubectl-sgmap/internal/usecase"
)

var (
	validSortFields = []string{"pod", "ip", "eni", "attachment", "sgids"}
)

func NewPodCommand(streams *genericclioptions.IOStreams) *cobra.Command {
	o := usecase.NewPodOptions(streams)
	cmd := &cobra.Command{
		Use:     "pod [NAME]",
		Aliases: []string{"pods", "po"},
		Short:   "Display security group information for pods",
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if o.SortField != "" {
				isValid := false
				for _, validField := range validSortFields {
					if o.SortField == validField {
						isValid = true
						break
					}
				}
				if !isValid {
					return fmt.Errorf("invalid sort field: %s, valid fields are: %s", o.SortField, strings.Join(validSortFields, ", "))
				}
			}

			if o.OutputFormat != "" {
				validFormats := []string{"json", "yaml", "table", "json-minimal"}
				isValid := false
				for _, format := range validFormats {
					if o.OutputFormat == format {
						isValid = true
						break
					}
				}
				if !isValid {
					return fmt.Errorf("invalid output format: %s, valid formats are: %s", o.OutputFormat, strings.Join(validFormats, ", "))
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				o.PodName = args[0]
			}

			return o.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "", "output format (json|json-minimal|yaml|table)")
	cmd.Flags().StringVar(&o.SortField, "sort", "pod", fmt.Sprintf("Specify the field to sort by (%s)", strings.Join(validSortFields, "|")))
	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	o.ConfigFlags.AddFlags(cmd.Flags())

	return cmd
}
