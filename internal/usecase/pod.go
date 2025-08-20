package usecase

import (
	"context"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
	"github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
	"github.com/naka-gawa/kubectl-sgmap/pkg/output"
)

// PodOptions contains options for the pod command
type PodOptions struct {
	PodName       string
	AllNamespaces bool
	OutputFormat  string
	IOStreams     *genericclioptions.IOStreams
	ConfigFlags   *genericclioptions.ConfigFlags
	K8sClient     kubernetes.Interface
	AWSClient     aws.Interface
}

// NewPodOptions creates new PodOptions with default values
func NewPodOptions(streams *genericclioptions.IOStreams, kubeConfigFlags *genericclioptions.ConfigFlags) *PodOptions {
	return &PodOptions{
		IOStreams:   streams,
		ConfigFlags: kubeConfigFlags,
	}
}

// Run executes the pod command business logic
func (o *PodOptions) Run(ctx context.Context) error {
	if o.K8sClient == nil {
		config, err := o.ConfigFlags.ToRESTConfig()
		if err != nil {
			return fmt.Errorf("failed to load kubeconfig: %w", err)
		}
		o.K8sClient, err = kubernetes.NewClient(config)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}
	}

	if o.AWSClient == nil {
		var err error
		o.AWSClient, err = aws.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create aws client: %w", err)
		}
	}

	namespace := *o.ConfigFlags.Namespace
	if o.AllNamespaces {
		namespace = ""
	}

	pods, err := o.K8sClient.GetPods(ctx, namespace, o.PodName)
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}

	if len(pods) == 0 {
		fmt.Fprintf(o.IOStreams.Out, "No resources found in namespace.\n")
		return nil
	}

	result, err := o.AWSClient.FetchSecurityGroupsByPods(ctx, pods)
	if err != nil {
		return fmt.Errorf("failed to get security groups: %w", err)
	}

	if len(result) == 0 {
		fmt.Fprintln(o.IOStreams.Out, "No security group information found for the specified pods")
		return nil
	}

	return output.OutputPodSecurityGroups(o.IOStreams.Out, result, o.OutputFormat)
}
