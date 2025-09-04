package usecase

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
	"github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
)

type fakeK8sClient struct {
	GetPodFunc   func(ctx context.Context, name, namespace string) (*corev1.Pod, error)
	ListPodsFunc func(ctx context.Context, namespace string) ([]corev1.Pod, error)
}

func (f *fakeK8sClient) GetPod(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
	return f.GetPodFunc(ctx, name, namespace)
}

func (f *fakeK8sClient) ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	return f.ListPodsFunc(ctx, namespace)
}

type fakeAWSClient struct {
	FetchSecurityGroupsByPodsFunc func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error)
}

func (f *fakeAWSClient) FetchSecurityGroupsByPods(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
	return f.FetchSecurityGroupsByPodsFunc(ctx, pods)
}

func TestPodOptions_Run(t *testing.T) {
	testCases := []struct {
		name        string
		k8sClient   kubernetes.Interface
		awsClient   aws.Interface
		podName     string
		output      string
		wantErr     bool
		expectedOut string
	}{
		{
			name: "no pods found",
			k8sClient: &fakeK8sClient{
				ListPodsFunc: func(ctx context.Context, namespace string) ([]corev1.Pod, error) {
					return []corev1.Pod{}, nil
				},
			},
			awsClient:   &fakeAWSClient{},
			expectedOut: "No resources found in namespace.\n",
		},
		{
			name: "get specific pod",
			k8sClient: &fakeK8sClient{
				GetPodFunc: func(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
					return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod"}}, nil
				},
			},
			awsClient: &fakeAWSClient{
				FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
					return []aws.PodSecurityGroupInfo{
						{Pod: pods[0]},
					}, nil
				},
			},
			podName: "test-pod",
		},
		{
			name: "k8s client error",
			k8sClient: &fakeK8sClient{
				ListPodsFunc: func(ctx context.Context, namespace string) ([]corev1.Pod, error) {
					return nil, fmt.Errorf("k8s error")
				},
			},
			awsClient: &fakeAWSClient{},
			wantErr:   true,
		},
		{
			name: "aws client error",
			k8sClient: &fakeK8sClient{
				ListPodsFunc: func(ctx context.Context, namespace string) ([]corev1.Pod, error) {
					return []corev1.Pod{{}}, nil
				},
			},
			awsClient: &fakeAWSClient{
				FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
					return nil, fmt.Errorf("aws error")
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			streams := &genericclioptions.IOStreams{
				Out:    &bytes.Buffer{},
				ErrOut: &bytes.Buffer{},
			}
			o := NewPodOptions(streams)
			o.K8sClient = tc.k8sClient
			o.AWSClient = tc.awsClient
			o.PodName = tc.podName
			o.ConfigFlags.Namespace = stringPointer("default")

			err := o.Run(context.Background())

			if (err != nil) != tc.wantErr {
				t.Fatalf("Run() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.expectedOut != "" {
				if got := streams.Out.(*bytes.Buffer).String(); got != tc.expectedOut {
					t.Errorf("Run() output = %q, want %q", got, tc.expectedOut)
				}
			}
		})
	}
}

func TestPodOptions_getNamespace(t *testing.T) {
	testCases := []struct {
		name              string
		flags             *genericclioptions.ConfigFlags
		allNamespaces     bool
		expectedNamespace string
		wantErr           bool
	}{
		{
			name:              "default namespace",
			flags:             genericclioptions.NewConfigFlags(true),
			expectedNamespace: "default",
		},
		{
			name: "namespace from flag",
			flags: &genericclioptions.ConfigFlags{
				Namespace: stringPointer("my-namespace"),
			},
			expectedNamespace: "my-namespace",
		},
		{
			name:              "all namespaces",
			flags:             genericclioptions.NewConfigFlags(true),
			allNamespaces:     true,
			expectedNamespace: "",
		},
		{
			name: "all namespaces overrides namespace flag",
			flags: &genericclioptions.ConfigFlags{
				Namespace: stringPointer("my-namespace"),
			},
			allNamespaces:     true,
			expectedNamespace: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o := &PodOptions{
				ConfigFlags:   tc.flags,
				AllNamespaces: tc.allNamespaces,
			}

			// a fake kubeconfig is required to prevent error
			o.ConfigFlags.KubeConfig = stringPointer("/dev/null")

			ns, err := o.getNamespace()

			if (err != nil) != tc.wantErr {
				t.Fatalf("getNamespace() error = %v, wantErr %v", err, tc.wantErr)
			}

			if ns != tc.expectedNamespace {
				t.Errorf("getNamespace() = %q, want %q", ns, tc.expectedNamespace)
			}
		})
	}
}

func stringPointer(s string) *string {
	return &s
}

func BenchmarkPodOptions_Run(b *testing.B) {
	k8sClient := &fakeK8sClient{
		ListPodsFunc: func(ctx context.Context, namespace string) ([]corev1.Pod, error) {
			return []corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
			}, nil
		},
	}
	awsClient := &fakeAWSClient{
		FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
			return []aws.PodSecurityGroupInfo{
				{Pod: pods[0]},
				{Pod: pods[1]},
			}, nil
		},
	}
	streams := &genericclioptions.IOStreams{
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	o := NewPodOptions(streams)
	o.K8sClient = k8sClient
	o.AWSClient = awsClient
	o.ConfigFlags.Namespace = stringPointer("default")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = o.Run(context.Background())
	}
}
