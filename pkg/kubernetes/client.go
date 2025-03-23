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
	// 名前空間の決定
	if allNamespaces {
		namespace = ""
	} else if namespace == "" {
		config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}

		currentContext := config.CurrentContext
		if ctx, exists := config.Contexts[currentContext]; exists && ctx.Namespace != "" {
			namespace = ctx.Namespace
		} else {
			namespace = "default"
		}
	}

	// ポッド名が指定されていれば特定のポッドを取得
	if podName != "" {
		pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod %s: %w", podName, err)
		}
		return []corev1.Pod{*pod}, nil
	}

	// ポッドの一覧を取得
	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return podList.Items, nil
}
