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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	awsSDK "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
)

type fakeAWSClient struct {
	FetchSecurityGroupsByPodsFunc func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error)
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

func TestPodOptions_Run(t *testing.T) {
	test := []struct {
		name         string
		clientset    kubernetes.Interface
		ac           *fakeAWSClient
		gotOutput    *bytes.Buffer
		outputFormat string
		checkOutput  func(t *testing.T, output string)
		wantErr      bool
	}{
		{
			name:      "no pods found",
			clientset: fake.NewSimpleClientset(),
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
			clientset: fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					PodIP: "10.0.0.1",
					Phase: corev1.PodRunning,
				},
			}),
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
			clientset: fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					PodIP: "10.0.0.1",
					Phase: corev1.PodRunning,
				},
			}),
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
			clientset: fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					PodIP: "10.0.0.1",
					Phase: corev1.PodRunning,
				},
			}),
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
			o := NewPodOptions(&genericclioptions.IOStreams{
				Out:    tt.gotOutput,
				ErrOut: io.Discard,
			})
			o.OutputFormat = tt.outputFormat
			o.K8sClient = tt.clientset
			o.AWSClient = tt.ac
			// Set a default namespace
			o.ConfigFlags.Namespace = stringPointer("default")
			runPodOptionsTest(t, o, tt.wantErr, tt.checkOutput)
		})
	}
}

func TestPodOptions_AWSClientError(t *testing.T) {
	o := NewPodOptions(&genericclioptions.IOStreams{
		Out:    &bytes.Buffer{},
		ErrOut: io.Discard,
	})
	o.K8sClient = fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.1"},
	})
	o.AWSClient = &fakeAWSClient{
		FetchSecurityGroupsByPodsFunc: func(ctx context.Context, pods []corev1.Pod) ([]aws.PodSecurityGroupInfo, error) {
			return nil, fmt.Errorf("mock AWS error")
		},
	}
	// Set a default namespace
	o.ConfigFlags.Namespace = stringPointer("default")

	err := o.Run(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get security groups: mock AWS error") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func stringPointer(s string) *string {
	return &s
}
