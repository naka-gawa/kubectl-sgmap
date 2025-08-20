package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client provides Kubernetes API access
type Client struct {
	clientset *kubernetes.Clientset
}

// Interface defines the methods provided by the AWS EC2 client.
// Used for dependency injection and testing.
type Interface interface {
	GetPods(ctx context.Context, namespace, podName string) ([]corev1.Pod, error)
}

// NewClient creates a new Kubernetes client
func NewClient(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset: clientset,
	}, nil
}

// GetPods retrieves pod information from the Kubernetes API
func (c *Client) GetPods(ctx context.Context, namespace, podName string) ([]corev1.Pod, error) {
	if podName != "" {
		pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod %s in namespace %s: %w", podName, namespace, err)
		}
		return []corev1.Pod{*pod}, nil
	}

	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	return podList.Items, nil
}
