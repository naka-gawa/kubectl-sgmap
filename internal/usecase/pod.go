// Package usecase provides business logic for kubectl-sgmap commands.
package usecase

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
	"github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
	"github.com/naka-gawa/kubectl-sgmap/pkg/output"
)

// PodOptions contains options for the pod command
type PodOptions struct {
	PodName       string
	OutputFormat  string
	SortField     string
	AllNamespaces *bool
	ConfigFlags   *genericclioptions.ConfigFlags
	IOStreams     *genericclioptions.IOStreams
	K8sClient     kubernetes.Interface
	AWSClient     aws.Interface
}

// NewPodOptions creates new PodOptions with default values
func NewPodOptions(streams *genericclioptions.IOStreams) *PodOptions {
	return &PodOptions{
		AllNamespaces: new(bool),
		ConfigFlags:   genericclioptions.NewConfigFlags(true),
		IOStreams:     streams,
	}
}

// Run executes the pod command business logic
func (o *PodOptions) Run(ctx context.Context) error {
	k8sClient := o.K8sClient
	if k8sClient == nil {
		var err error
		k8sClient, err = kubernetes.NewClient(o.ConfigFlags)
		if err != nil {
			return err
		}
	}

	if o.AWSClient == nil {
		var err error
		o.AWSClient, err = aws.NewClient(nil)
		if err != nil {
			return fmt.Errorf("failed to create aws client: %w", err)
		}
	}

	namespace, err := o.getNamespace()
	if err != nil {
		return fmt.Errorf("failed to get namespace: %w", err)
	}

	var pods []corev1.Pod
	if o.PodName != "" {
		pod, podErr := k8sClient.GetPod(ctx, o.PodName, namespace)
		if podErr != nil {
			return podErr
		}
		pods = []corev1.Pod{*pod}
	} else {
		var listErr error
		pods, listErr = k8sClient.ListPods(ctx, namespace)
		if listErr != nil {
			return listErr
		}
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

	return output.OutputPodSecurityGroups(o.IOStreams.Out, result, o.OutputFormat, o.SortField)
}

func (o *PodOptions) getNamespace() (string, error) {
	namespace, _, err := o.ConfigFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return "", err
	}

	if o.AllNamespaces != nil && *o.AllNamespaces {
		return "", nil
	}

	return namespace, nil
}
