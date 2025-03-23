package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Client provides Kubernetes API access
type Client struct {
	clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset: clientset,
	}, nil
}

// GetPods retrieves pod information from the Kubernetes API
func (c *Client) GetPods(ctx context.Context, namespace, podName string, allNamespaces bool) ([]corev1.Pod, error) {
	if allNamespaces {
		namespace = ""
	} else if namespace == "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

		ns, _, err := kubeConfig.Namespace()
		if err != nil {
			return nil, fmt.Errorf("failed to get current namespace: %w", err)
		}

		namespace = ns
	}

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
