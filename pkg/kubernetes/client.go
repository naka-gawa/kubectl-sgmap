// Package kubernetes provides a client for interacting with the Kubernetes API.
package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// Interface defines the methods for interacting with Kubernetes.
type Interface interface {
	GetPod(ctx context.Context, name, namespace string) (*corev1.Pod, error)
	ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error)
}

// Client is a client for interacting with the Kubernetes API.
type Client struct {
	clientset kubernetes.Interface
}

// NewClient creates a new Kubernetes client from the given config flags.
func NewClient(configFlags *genericclioptions.ConfigFlags) (*Client, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{clientset: clientset}, nil
}

// GetPod gets a pod by name in a namespace.
func (c *Client) GetPod(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s in namespace %s: %w", name, namespace, err)
	}
	return pod, nil
}

// ListPods lists all pods in a namespace.
func (c *Client) ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}
	return podList.Items, nil
}
