package k8s

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Options is a struct that holds the options for the Get command.
type Options struct {
	genericclioptions.IOStreams
	ConfigFlags *genericclioptions.ConfigFlags
	Namespace   string
}

// GetOptions is a function that returns the options for the Get command.
func GetOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams:   streams,
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		Namespace:   "default",
	}
}

// SetNamespace is a method that sets the namespace for the Options struct.
func (o *Options) SetNamespace() {
	if ns := *o.ConfigFlags.Namespace; ns != "" {
		o.Namespace = ns
	}
}
