// Code generated by kubectl-plugin-builder; DO NOT EDIT.

/* MIT License
 *
 * Copyright (c) 2024 naka-gawa
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */
package sgmap

import (
	"github.com/spf13/cobra"

	"github.com/naka-gawa/kubectl-sgmap/internal/cmd"
	"github.com/naka-gawa/kubectl-sgmap/internal/cmd/sgmap/pod"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	// sgmapOutputModeFlag provides
	// user-passed option to options.
	sgmapOutputModeFlag string
)

// WARNING: don't rename this function.
func NewCommand(streams *genericclioptions.IOStreams) *cobra.Command {
	c := &cobra.Command{
		Use: "sgmap",

		Aliases: []string{},

		RunE: func(cmd *cobra.Command, args []string) error {
			o := &options{streams: streams}
			if err := o.Complete(cmd, args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	hangChildrenOnCommand(c, streams)
	defineCommandFlags(c)

	return c
}

// hangChildrenOnCommand enumerates command's children and attach them into it.
func hangChildrenOnCommand(c *cobra.Command, streams *genericclioptions.IOStreams) {
	c.AddCommand(pod.NewCommand(streams))
}

// defineCommandFlags declares primitive flags.
func defineCommandFlags(c *cobra.Command) {
	c.Flags().StringVarP(
		&sgmapOutputModeFlag,
		"output",
		"o",
		cmd.OutputModeNormal,
		"the command's output mode",
	)
}
