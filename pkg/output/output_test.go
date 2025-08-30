package output

import (
	"bytes"
	"testing"

	awsSDK "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/naka-gawa/kubectl-sgmap/pkg/aws"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func strPtr(s string) *string {
	return &s
}

func TestOutputPodSecurityGroups(t *testing.T) {
	testCases := []struct {
		name     string
		format   string
		data     []aws.PodSecurityGroupInfo
		expected string
		wantErr  bool
	}{
		// Test cases will be added here
		{
			name:   "table output with single entry",
			format: "table",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111"), GroupName: strPtr("sg-name-1")},
						{GroupId: strPtr("sg-22222")},
					},
				},
			},
			expected: "POD NAME  IP ADDRESS  ENI ID     ATTACHMENT  SECURITY GROUPS\npod1      10.0.0.1    eni-12345  pod-eni     sg-11111 (sg-name-1), sg-22222\n",
		},
		{
			name:   "table output with multiple entries",
			format: "table",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111"), GroupName: strPtr("sg-name-1")},
					},
				},
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "ns2"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.2"},
					},
					ENI:             "eni-67890",
					AttachmentLevel: "node-primary-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-33333"), GroupName: strPtr("sg-name-3")},
					},
				},
			},
			expected: "POD NAME  IP ADDRESS  ENI ID     ATTACHMENT        SECURITY GROUPS\npod1      10.0.0.1    eni-12345  pod-eni           sg-11111 (sg-name-1)\npod2      10.0.0.2    eni-67890  node-primary-eni  sg-33333 (sg-name-3)\n",
		},
		{
			name:     "table output with empty data",
			format:   "table",
			data:     []aws.PodSecurityGroupInfo{},
			expected: "POD NAME  IP ADDRESS  ENI ID  ATTACHMENT  SECURITY GROUPS\n",
		},
		{
			name:   "table output with missing group name",
			format: "table",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111")},
					},
				},
			},
			expected: "POD NAME  IP ADDRESS  ENI ID     ATTACHMENT  SECURITY GROUPS\npod1      10.0.0.1    eni-12345  pod-eni     sg-11111\n",
		},
		{
			name:   "table output with missing pod ip",
			format: "table",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111"), GroupName: strPtr("sg-name-1")},
					},
				},
			},
			expected: "POD NAME  IP ADDRESS  ENI ID     ATTACHMENT  SECURITY GROUPS\npod1                  eni-12345  pod-eni     sg-11111 (sg-name-1)\n",
		},
		{
			name:   "json output with single entry",
			format: "json",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "pod1",
							Namespace:         "ns1",
							CreationTimestamp: metav1.Time{},
						},
						Status: corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111"), GroupName: strPtr("sg-name-1")},
					},
				},
			},
			expected: "[\n  {\n    \"pod\": {\n      \"metadata\": {\n        \"name\": \"pod1\",\n        \"namespace\": \"ns1\"\n      },\n      \"spec\": {\n        \"containers\": null\n      },\n      \"status\": {\n        \"podIP\": \"10.0.0.1\"\n      }\n    },\n    \"securityGroups\": [\n      {\n        \"Description\": null,\n        \"GroupId\": \"sg-11111\",\n        \"GroupName\": \"sg-name-1\",\n        \"IpPermissions\": null,\n        \"IpPermissionsEgress\": null,\n        \"OwnerId\": null,\n        \"SecurityGroupArn\": null,\n        \"Tags\": null,\n        \"VpcId\": null\n      }\n    ],\n    \"eni\": \"eni-12345\",\n    \"attachmentLevel\": \"pod-eni\"\n  }\n]\n",
		},
		{
			name:   "json output with multiple entries",
			format: "json",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "pod1",
							Namespace:         "ns1",
							CreationTimestamp: metav1.Time{},
						},
						Status: corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111"), GroupName: strPtr("sg-name-1")},
					},
				},
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "pod2",
							Namespace:         "ns2",
							CreationTimestamp: metav1.Time{},
						},
						Status: corev1.PodStatus{PodIP: "10.0.0.2"},
					},
					ENI:             "eni-67890",
					AttachmentLevel: "node-primary-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-33333"), GroupName: strPtr("sg-name-3")},
					},
				},
			},
			expected: "[\n  {\n    \"pod\": {\n      \"metadata\": {\n        \"name\": \"pod1\",\n        \"namespace\": \"ns1\"\n      },\n      \"spec\": {\n        \"containers\": null\n      },\n      \"status\": {\n        \"podIP\": \"10.0.0.1\"\n      }\n    },\n    \"securityGroups\": [\n      {\n        \"Description\": null,\n        \"GroupId\": \"sg-11111\",\n        \"GroupName\": \"sg-name-1\",\n        \"IpPermissions\": null,\n        \"IpPermissionsEgress\": null,\n        \"OwnerId\": null,\n        \"SecurityGroupArn\": null,\n        \"Tags\": null,\n        \"VpcId\": null\n      }\n    ],\n    \"eni\": \"eni-12345\",\n    \"attachmentLevel\": \"pod-eni\"\n  },\n  {\n    \"pod\": {\n      \"metadata\": {\n        \"name\": \"pod2\",\n        \"namespace\": \"ns2\"\n      },\n      \"spec\": {\n        \"containers\": null\n      },\n      \"status\": {\n        \"podIP\": \"10.0.0.2\"\n      }\n    },\n    \"securityGroups\": [\n      {\n        \"Description\": null,\n        \"GroupId\": \"sg-33333\",\n        \"GroupName\": \"sg-name-3\",\n        \"IpPermissions\": null,\n        \"IpPermissionsEgress\": null,\n        \"OwnerId\": null,\n        \"SecurityGroupArn\": null,\n        \"Tags\": null,\n        \"VpcId\": null\n      }\n    ],\n    \"eni\": \"eni-67890\",\n    \"attachmentLevel\": \"node-primary-eni\"\n  }\n]\n",
		},
		{
			name:     "json output with empty data",
			format:   "json",
			data:     []aws.PodSecurityGroupInfo{},
			expected: "[]\n",
		},
		{
			name:   "yaml output with single entry",
			format: "yaml",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111"), GroupName: strPtr("sg-name-1")},
					},
				},
			},
			expected: "- podName: pod1\n  namespace: ns1\n  eni: eni-12345\n  attachmentLevel: pod-eni\n  securityGroups:\n    - groupId: sg-11111\n      groupName: sg-name-1\n",
		},
		{
			name:   "yaml output with multiple entries",
			format: "yaml",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111"), GroupName: strPtr("sg-name-1")},
					},
				},
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "ns2"},
					},
					ENI:             "eni-67890",
					AttachmentLevel: "node-primary-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-33333"), GroupName: strPtr("sg-name-3")},
					},
				},
			},
			expected: "- podName: pod1\n  namespace: ns1\n  eni: eni-12345\n  attachmentLevel: pod-eni\n  securityGroups:\n    - groupId: sg-11111\n      groupName: sg-name-1\n- podName: pod2\n  namespace: ns2\n  eni: eni-67890\n  attachmentLevel: node-primary-eni\n  securityGroups:\n    - groupId: sg-33333\n      groupName: sg-name-3\n",
		},
		{
			name:     "yaml output with empty data",
			format:   "yaml",
			data:     []aws.PodSecurityGroupInfo{},
			expected: "[]\n",
		},
		{
			name:   "invalid format should default to table",
			format: "invalid-format",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
						Status:     corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{GroupId: strPtr("sg-11111"), GroupName: strPtr("sg-name-1")},
					},
				},
			},
			expected: "POD NAME  IP ADDRESS  ENI ID     ATTACHMENT  SECURITY GROUPS\npod1      10.0.0.1    eni-12345  pod-eni     sg-11111 (sg-name-1)\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := OutputPodSecurityGroups(&buf, tc.data, tc.format)

			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected error: %v", err)
			}

			if got := buf.String(); got != tc.expected {
				t.Errorf("unexpected output: got %q, want %q", got, tc.expected)
			}
		})
	}
}
