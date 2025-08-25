// Package kubernetes provides a client for interacting with the Kubernetes API.
package kubernetes

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNewClient_Error(t *testing.T) {
	configFlags := genericclioptions.NewConfigFlags(true)
	configFlags.KubeConfig = stringPointer("/non/existent/path")

	_, err := NewClient(configFlags)
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func stringPointer(s string) *string {
	return &s
}

func TestClient_GetPod(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	})
	client := &Client{clientset: clientset}

	pod, err := client.GetPod(context.Background(), "test-pod", "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pod.Name != "test-pod" {
		t.Errorf("expected pod name to be 'test-pod', got '%s'", pod.Name)
	}
}

func TestClient_GetPod_NotFound(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := &Client{clientset: clientset}

	_, err := client.GetPod(context.Background(), "test-pod", "default")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func TestClient_ListPods(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "default"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-2", Namespace: "default"}},
	)
	client := &Client{clientset: clientset}

	pods, err := client.ListPods(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("expected 2 pods, got %d", len(pods))
	}
}

func TestClient_ListPods_Empty(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := &Client{clientset: clientset}

	pods, err := client.ListPods(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pods) != 0 {
		t.Errorf("expected 0 pods, got %d", len(pods))
	}
}

func TestClient_GetPod_Error(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	clientset.PrependReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("failed to get pod")
	})
	client := &Client{clientset: clientset}

	_, err := client.GetPod(context.Background(), "test-pod", "default")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func TestClient_ListPods_Error(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	clientset.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("failed to list pods")
	})
	client := &Client{clientset: clientset}

	_, err := client.ListPods(context.Background(), "default")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}
