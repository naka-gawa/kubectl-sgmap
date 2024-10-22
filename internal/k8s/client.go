package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// DefaultKubeconfigPath is the default path to the kubeconfig file
var defaultKubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")

// Clientset is a wrapper around the Kubernetes clientset
type Clientset struct {
	Clientset *kubernetes.Clientset
}

// GetClientset is a function that returns a Kubernetes clientset,
// which can be used to interact with a Kubernetes cluster. It handles the kubeconfig
// resolution, allowing the client to connect to the cluster.
//
// GetClientset first checks if the KUBECONFIG environment variable is set.
// If set, it uses the path provided by the environment variable to load the kubeconfig.
// If the environment variable is not set, it defaults to using the standard kubeconfig location
// at $HOME/.kube/config.
//
// The function ensures that the clientset is initialized correctly and returns an error
// if there are any issues with the kubeconfig resolution or clientset creation.
// This provides a safe way for the client to connect to the Kubernetes cluster with expected behavior.
func GetClientset() (*Clientset, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = defaultKubeconfigPath
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return &Clientset{Clientset: clientset}, nil
}
