package usecase

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
	"github.com/naka-gawa/kubectl-sgmap/pkg/output"
)

// PodOptions contains options for the pod command
type PodOptions struct {
	PodName       string
	OutputFormat  string
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
	var clientset kubernetes.Interface
	var err error

	if o.K8sClient != nil {
		clientset = o.K8sClient
	} else {
		config, err := o.ConfigFlags.ToRESTConfig()
		if err != nil {
			return fmt.Errorf("failed to load kubeconfig: %w", err)
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}
	}

	if o.AWSClient == nil {
		o.AWSClient, err = aws.NewClient()
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
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, o.PodName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pod %s in namespace %s: %w", o.PodName, namespace, err)
		}
		pods = []corev1.Pod{*pod}
	} else {
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
		}
		pods = podList.Items
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
