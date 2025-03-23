package usecase

import (
	"context"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
	"github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
	"github.com/naka-gawa/kubectl-sgmap/pkg/utils"
)

// PodOptions contains options for the pod command
type PodOptions struct {
	PodName       string
	Namespace     string
	AllNamespaces bool
	OutputFormat  string
	IOStreams     *genericclioptions.IOStreams
}

// NewPodOptions creates new PodOptions with default values
func NewPodOptions(streams *genericclioptions.IOStreams) *PodOptions {
	return &PodOptions{
		IOStreams: streams,
	}
}

// Run executes the pod command business logic
func (o *PodOptions) Run(ctx context.Context) error {
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	pods, err := k8sClient.GetPods(ctx, o.Namespace, o.PodName, o.AllNamespaces)
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}

	if len(pods) == 0 {
		fmt.Fprintln(o.IOStreams.Out, "No resources found in default namespace.")
		return nil
	}

	awsClient, err := aws.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create aws client: %w", err)
	}

	result, err := awsClient.GetSecurityGroupsForPods(ctx, pods)
	if err != nil {
		return fmt.Errorf("failed to get security groups: %w", err)
	}

	if len(result) == 0 {
		fmt.Fprintln(o.IOStreams.Out, "No security group information found for the specified pods")
		return nil
	}

	// 結果の出力
	return utils.OutputPodSecurityGroups(o.IOStreams.Out, result, o.OutputFormat)
}
