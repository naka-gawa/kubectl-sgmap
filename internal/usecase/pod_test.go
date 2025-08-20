package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	awsSDK "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
)

type fakeK8sClient struct {
	GetPodsFunc func(ctx context.Context, namespace, podName string) ([]corev1.Pod, error)
}

type fakeAWSClient struct {
	FetchSecurityGroupsByPodsFunc func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error)
}

func (f *fakeK8sClient) GetPods(ctx context.Context, namespace, podName string) ([]corev1.Pod, error) {
	return f.GetPodsFunc(ctx, namespace, podName)
}

func (f *fakeAWSClient) FetchSecurityGroupsByPods(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
	return f.FetchSecurityGroupsByPodsFunc(ctx, pods)
}

func runPodOptionsTest(t *testing.T, o *PodOptions, wantErr bool, checkOutput func(*testing.T, string)) {
	t.Helper()

	err := o.Run(context.Background())
	if wantErr {
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		return
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	got := o.IOStreams.Out.(*bytes.Buffer).String()

	if checkOutput != nil {
		checkOutput(t, got)
	}
}

func newTestConfigFlags(namespace string) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		Namespace: &namespace,
	}
}

func TestPodOptions_Run(t *testing.T) {
	testNamespace := "default"
	test := []struct {
		name         string
		kc           *fakeK8sClient
		ac           *fakeAWSClient
		gotOutput    *bytes.Buffer
		outputFormat string
		checkOutput  func(t *testing.T, output string)
		wantErr      bool
	}{
		{
			name: "no pods found",
			kc: &fakeK8sClient{
				GetPodsFunc: func(ctx context.Context, ns, name string) ([]corev1.Pod, error) {
					return []corev1.Pod{}, nil
				},
			},
			ac: &fakeAWSClient{
				FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
					return nil, nil
				},
			},
			gotOutput: &bytes.Buffer{},
			checkOutput: func(t *testing.T, output string) {
				t.Helper()
				expected := "No resources found in namespace.\n"
				if output != expected {
					t.Errorf("expected output %q, but got %q", expected, output)
				}
			},
			wantErr: false,
		},
		{
			name:         "json output",
			outputFormat: "json",
			kc: &fakeK8sClient{
				GetPodsFunc: func(ctx context.Context, ns, name string) ([]corev1.Pod, error) {
					return []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-pod",
								Namespace: "default",
							},
							Status: corev1.PodStatus{
								PodIP: "10.0.0.1",
								Phase: corev1.PodRunning,
							},
						},
					}, nil
				},
			},
			ac: &fakeAWSClient{
				FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
					return []aws.PodSecurityGroupInfo{
						{
							Pod: pods[0],
							ENI: "eni-1234",
							SecurityGroups: []types.SecurityGroup{
								{
									GroupId:   awsSDK.String("sg-1234"),
									GroupName: awsSDK.String("test-sg"),
								},
							},
						},
					}, nil
				},
			},
			gotOutput: &bytes.Buffer{},
			checkOutput: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, `"eni-1234"`) ||
					!strings.Contains(output, `"sg-1234"`) ||
					!strings.Contains(output, `"test-sg"`) {
					t.Errorf("unexpected json output: %s", output)
				}
			},
		},
		{
			name:         "yaml output",
			outputFormat: "yaml",
			kc: &fakeK8sClient{
				GetPodsFunc: func(ctx context.Context, ns, name string) ([]corev1.Pod, error) {
					return []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-pod",
								Namespace: "default",
							},
							Status: corev1.PodStatus{
								PodIP: "10.0.0.1",
								Phase: corev1.PodRunning,
							},
						},
					}, nil
				},
			},
			ac: &fakeAWSClient{
				FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
					return []aws.PodSecurityGroupInfo{
						{
							Pod: pods[0],
							ENI: "eni-1234",
							SecurityGroups: []types.SecurityGroup{
								{
									GroupId:   awsSDK.String("sg-1234"),
									GroupName: awsSDK.String("test-sg"),
								},
							},
						},
					}, nil
				},
			},
			gotOutput: &bytes.Buffer{},
			checkOutput: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "eni: eni-1234") ||
					!strings.Contains(output, "groupId: sg-1234") ||
					!strings.Contains(output, "groupName: test-sg") {
					t.Errorf("unexpected yaml output: %s", output)
				}
			},
		},
		{
			name:         "structured data validation",
			outputFormat: "json",
			kc: &fakeK8sClient{
				GetPodsFunc: func(ctx context.Context, ns, name string) ([]corev1.Pod, error) {
					return []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-pod",
								Namespace: "default",
							},
							Status: corev1.PodStatus{
								PodIP: "10.0.0.1",
								Phase: corev1.PodRunning,
							},
						},
					}, nil
				},
			},
			ac: &fakeAWSClient{
				FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
					if len(pods) != 1 {
						t.Errorf("expected 1 pod, got %d", len(pods))
					}
					pod := pods[0]
					if pod.Name != "test-pod" || pod.Namespace != "default" || pod.Status.PodIP != "10.0.0.1" {
						t.Errorf("unexpected pod info: %+v", pod)
					}

					return []aws.PodSecurityGroupInfo{
						{
							Pod: pod,
							ENI: "eni-9999",
							SecurityGroups: []types.SecurityGroup{
								{
									GroupId:   awsSDK.String("sg-9999"),
									GroupName: awsSDK.String("zunda-sg"),
								},
							},
						},
					}, nil
				},
			},
			gotOutput: &bytes.Buffer{},
			checkOutput: func(t *testing.T, output string) {
				t.Helper()
				if output == "" {
					return
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			o := &PodOptions{
				ConfigFlags: newTestConfigFlags(testNamespace),
				IOStreams: &genericclioptions.IOStreams{
					Out:    tt.gotOutput,
					ErrOut: io.Discard,
				},
				OutputFormat: tt.outputFormat,
				K8sClient:    tt.kc,
				AWSClient:    tt.ac,
			}
			runPodOptionsTest(t, o, tt.wantErr, tt.checkOutput)
		})
	}
}

func TestPodOptions_K8sClientCreationError(t *testing.T) {
	t.Helper()

	// Provide invalid kubeconfig path to trigger an error
	invalidPath := "/tmp/non-existent-kubeconfig"
	configFlags := &genericclioptions.ConfigFlags{
		KubeConfig: &invalidPath,
	}

	o := &PodOptions{
		ConfigFlags: configFlags,
		IOStreams: &genericclioptions.IOStreams{
			Out:    &bytes.Buffer{},
			ErrOut: io.Discard,
		},
		K8sClient: nil,
		AWSClient: &fakeAWSClient{
			FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
				return nil, nil
			},
		},
	}

	err := o.Run(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to load kubeconfig") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPodOptions_AWSClientError(t *testing.T) {
	testNamespace := "default"
	o := &PodOptions{
		ConfigFlags: newTestConfigFlags(testNamespace),
		IOStreams: &genericclioptions.IOStreams{
			Out:    &bytes.Buffer{},
			ErrOut: io.Discard,
		},
		K8sClient: &fakeK8sClient{
			GetPodsFunc: func(ctx context.Context, ns, name string) ([]corev1.Pod, error) {
				return []corev1.Pod{
					{Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.1"}},
				}, nil
			},
		},
		AWSClient: &fakeAWSClient{
			FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
				return nil, fmt.Errorf("mock AWS error")
			},
		},
	}

	err := o.Run(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get security groups: mock AWS error") {
		t.Errorf("unexpected error message: %v", err)
	}
}
