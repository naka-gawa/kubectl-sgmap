package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/naka-gawa/kubectl-sgmap/internal/usecase"
	aws_client "github.com/naka-gawa/kubectl-sgmap/pkg/aws"
	k8s "github.com/naka-gawa/kubectl-sgmap/pkg/kubernetes"
)

type mockAWSClient struct {
	aws_client.Interface
	DescribeNetworkInterfacesFunc func(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DescribeSecurityGroupsFunc    func(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
}

func (m *mockAWSClient) DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return m.DescribeNetworkInterfacesFunc(ctx, params, optFns...)
}

func (m *mockAWSClient) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return m.DescribeSecurityGroupsFunc(ctx, params, optFns...)
}

func (m *mockAWSClient) FetchSecurityGroupsByPods(ctx context.Context, pods []corev1.Pod) ([]aws_client.PodSecurityGroupInfo, error) {
	// A simplified mock implementation of FetchSecurityGroupsByPods
	var result []aws_client.PodSecurityGroupInfo
	for _, pod := range pods {
		if pod.Status.PodIP == "" {
			continue
		}
		result = append(result, aws_client.PodSecurityGroupInfo{
			Pod: pod,
			ENI: "eni-12345",
			SecurityGroups: []ec2types.SecurityGroup{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
		})
	}
	return result, nil
}

type mockK8sClient struct {
	k8s.Interface
	clientset *fake.Clientset
}

func (m *mockK8sClient) GetPod(ctx context.Context, podName, namespace string) (*corev1.Pod, error) {
	return m.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
}

func (m *mockK8sClient) ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	pods, err := m.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func TestIntegration(t *testing.T) {
	// 1. Setup
	// Create a fake Kubernetes client
	clientset := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			PodIP: "192.168.1.1",
		},
	})
	k8sClient := &mockK8sClient{clientset: clientset}

	// Create a mock EC2 client
	mockClient := &mockAWSClient{}

	// 2. Test Cases
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		options        *usecase.PodOptions
	}{
		{
			name:           "get all pods in default namespace",
			args:           []string{"pod"},
			expectedOutput: "POD NAME  IP ADDRESS   ENI ID     ATTACHMENT  SECURITY GROUPS\ntest-pod  192.168.1.1  eni-12345              sg-12345\n",
			options: &usecase.PodOptions{
				ConfigFlags: genericclioptions.NewConfigFlags(true),
				IOStreams:   &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
				K8sClient:   k8sClient,
				AWSClient:   mockClient,
			},
		},
		{
			name:           "get specific pod by name",
			args:           []string{"pod", "test-pod"},
			expectedOutput: "POD NAME  IP ADDRESS   ENI ID     ATTACHMENT  SECURITY GROUPS\ntest-pod  192.168.1.1  eni-12345              sg-12345\n",
			options: &usecase.PodOptions{
				ConfigFlags: genericclioptions.NewConfigFlags(true),
				IOStreams:   &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
				PodName:     "test-pod",
				K8sClient:   k8sClient,
				AWSClient:   mockClient,
			},
		},
		{
			name:           "get pods in json format",
			args:           []string{"pod", "-o", "json"},
			expectedOutput: "[\n  {\n    \"pod\": {\n      \"metadata\": {\n        \"name\": \"test-pod\",\n        \"namespace\": \"default\",\n        \"creationTimestamp\": null\n      },\n      \"spec\": {\n        \"containers\": null\n      },\n      \"status\": {\n        \"podIP\": \"192.168.1.1\"\n      }\n    },\n    \"securityGroups\": [\n      {\n        \"Description\": null,\n        \"GroupId\": \"sg-12345\",\n        \"GroupName\": null,\n        \"IpPermissions\": null,\n        \"IpPermissionsEgress\": null,\n        \"OwnerId\": null,\n        \"SecurityGroupArn\": null,\n        \"Tags\": null,\n        \"VpcId\": null\n      }\n    ],\n    \"eni\": \"eni-12345\",\n    \"attachmentLevel\": \"\"\n  }\n]\n",
			options: &usecase.PodOptions{
				ConfigFlags:  genericclioptions.NewConfigFlags(true),
				IOStreams:    &genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
				OutputFormat: "json",
				K8sClient:    k8sClient,
				AWSClient:    mockClient,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout to a buffer
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the command
			o := usecase.NewPodOptions(tt.options.IOStreams)
			o.K8sClient = tt.options.K8sClient
			o.AWSClient = tt.options.AWSClient
			o.ConfigFlags = tt.options.ConfigFlags
			o.OutputFormat = tt.options.OutputFormat
			cmd := NewPodCommand(o.IOStreams)
			cmd.SetArgs(tt.args)
			o.IOStreams.Out = w
			err := o.Run(context.Background())
			assert.NoError(t, err)

			// Restore stdout
			w.Close()
			os.Stdout = old
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Assert the output
			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}
