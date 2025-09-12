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

func int32Ptr(i int32) *int32 {
	return &i
}

func TestOutputPodSecurityGroups(t *testing.T) {
	unsortedData := []aws.PodSecurityGroupInfo{
		{
			Pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-c", Namespace: "ns1"},
				Status:     corev1.PodStatus{PodIP: "10.0.0.3"},
			},
			ENI:             "eni-3",
			AttachmentLevel: "pod-eni",
			SecurityGroups:  []awsSDK.SecurityGroup{{GroupId: strPtr("sg-c")}},
		},
		{
			Pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-a", Namespace: "ns1"},
				Status:     corev1.PodStatus{PodIP: "10.0.0.1"},
			},
			ENI:             "eni-1",
			AttachmentLevel: "node-primary-eni",
			SecurityGroups: []awsSDK.SecurityGroup{
				{
					GroupId:   strPtr("sg-a"),
					GroupName: strPtr("sg-name-a"),
					IpPermissions: []awsSDK.IpPermission{
						{
							IpProtocol: strPtr("tcp"),
							FromPort:   int32Ptr(80),
							ToPort:     int32Ptr(80),
							IpRanges:   []awsSDK.IpRange{{CidrIp: strPtr("0.0.0.0/0")}},
						},
					},
					IpPermissionsEgress: []awsSDK.IpPermission{
						{
							IpProtocol: strPtr("-1"),
							IpRanges:   []awsSDK.IpRange{{CidrIp: strPtr("0.0.0.0/0")}},
						},
					},
				},
			},
		},
		{
			Pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-b", Namespace: "ns2"},
				Status:     corev1.PodStatus{PodIP: "10.0.0.2"},
			},
			ENI:             "eni-2",
			AttachmentLevel: "trunk-eni",
			SecurityGroups:  []awsSDK.SecurityGroup{{GroupId: strPtr("sg-b")}},
		},
		{
			Pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-d", Namespace: "ns2"},
				Status:     corev1.PodStatus{PodIP: "10.0.0.10"},
			},
			ENI:             "eni-4",
			AttachmentLevel: "other",
			SecurityGroups:  []awsSDK.SecurityGroup{{GroupId: strPtr("sg-d")}},
		},
	}

	testCases := []struct {
		name      string
		format    string
		sortField string
		data      []aws.PodSecurityGroupInfo
		expected  string
		wantErr   bool
	}{
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
			name:      "sort by pod name",
			format:    "table",
			sortField: "pod",
			data:      unsortedData,
			expected:  "POD NAME  IP ADDRESS  ENI ID  ATTACHMENT        SECURITY GROUPS\npod-a     10.0.0.1    eni-1   node-primary-eni  sg-a (sg-name-a)\npod-b     10.0.0.2    eni-2   trunk-eni         sg-b\npod-c     10.0.0.3    eni-3   pod-eni           sg-c\npod-d     10.0.0.10   eni-4   other             sg-d\n",
		},
		{
			name:      "json output with sorting",
			format:    "json",
			sortField: "pod",
			data:      unsortedData,
			expected:  "[\n  {\n    \"podName\": \"pod-a\",\n    \"namespace\": \"ns1\",\n    \"podIP\": \"10.0.0.1\",\n    \"eni\": \"eni-1\",\n    \"attachmentLevel\": \"node-primary-eni\",\n    \"securityGroups\": [\n      {\n        \"id\": \"sg-a\",\n        \"name\": \"sg-name-a\",\n        \"inboundRules\": [\n          {\n            \"protocol\": \"tcp\",\n            \"fromPort\": 80,\n            \"toPort\": 80,\n            \"sources\": [\n              \"0.0.0.0/0\"\n            ]\n          }\n        ],\n        \"outboundRules\": [\n          {\n            \"protocol\": \"-1\",\n            \"destinations\": [\n              \"0.0.0.0/0\"\n            ]\n          }\n        ]\n      }\n    ]\n  },\n  {\n    \"podName\": \"pod-b\",\n    \"namespace\": \"ns2\",\n    \"podIP\": \"10.0.0.2\",\n    \"eni\": \"eni-2\",\n    \"attachmentLevel\": \"trunk-eni\",\n    \"securityGroups\": [\n      {\n        \"id\": \"sg-b\"\n      }\n    ]\n  },\n  {\n    \"podName\": \"pod-c\",\n    \"namespace\": \"ns1\",\n    \"podIP\": \"10.0.0.3\",\n    \"eni\": \"eni-3\",\n    \"attachmentLevel\": \"pod-eni\",\n    \"securityGroups\": [\n      {\n        \"id\": \"sg-c\"\n      }\n    ]\n  },\n  {\n    \"podName\": \"pod-d\",\n    \"namespace\": \"ns2\",\n    \"podIP\": \"10.0.0.10\",\n    \"eni\": \"eni-4\",\n    \"attachmentLevel\": \"other\",\n    \"securityGroups\": [\n      {\n        \"id\": \"sg-d\"\n      }\n    ]\n  }\n]\n",
		},
		{
			name:      "yaml output with sorting",
			format:    "yaml",
			sortField: "pod",
			data:      unsortedData,
			expected: "- podName: pod-a\n  namespace: ns1\n  eni: eni-1\n  attachmentLevel: node-primary-eni\n  securityGroups:\n    - groupId: sg-a\n      groupName: sg-name-a\n- podName: pod-b\n  namespace: ns2\n  eni: eni-2\n  attachmentLevel: trunk-eni\n  securityGroups:\n    - groupId: sg-b\n      groupName: \"\"\n- podName: pod-c\n  namespace: ns1\n  eni: eni-3\n  attachmentLevel: pod-eni\n  securityGroups:\n    - groupId: sg-c\n      groupName: \"\"\n- podName: pod-d\n  namespace: ns2\n  eni: eni-4\n  attachmentLevel: other\n  securityGroups:\n    - groupId: sg-d\n      groupName: \"\"\n",
		},
		{
			name:   "json output with single entry",
			format: "json",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod1",
							Namespace: "ns1",
						},
						Status: corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{
							GroupId:   strPtr("sg-11111"),
							GroupName: strPtr("sg-name-1"),
							IpPermissions: []awsSDK.IpPermission{
								{
									IpProtocol: strPtr("tcp"),
									FromPort:   int32Ptr(443),
									ToPort:     int32Ptr(443),
									IpRanges:   []awsSDK.IpRange{{CidrIp: strPtr("10.0.0.0/8")}},
								},
							},
						},
					},
				},
			},
			expected: "[\n  {\n    \"podName\": \"pod1\",\n    \"namespace\": \"ns1\",\n    \"podIP\": \"10.0.0.1\",\n    \"eni\": \"eni-12345\",\n    \"attachmentLevel\": \"pod-eni\",\n    \"securityGroups\": [\n      {\n        \"id\": \"sg-11111\",\n        \"name\": \"sg-name-1\",\n        \"inboundRules\": [\n          {\n            \"protocol\": \"tcp\",\n            \"fromPort\": 443,\n            \"toPort\": 443,\n            \"sources\": [\n              \"10.0.0.0/8\"\n            ]\n          }\n        ]\n      }\n    ]\n  }\n]\n",
		},
		{
			name:   "json-minimal output with single entry",
			format: "json-minimal",
			data: []aws.PodSecurityGroupInfo{
				{
					Pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod1",
							Namespace: "ns1",
						},
						Status: corev1.PodStatus{PodIP: "10.0.0.1"},
					},
					ENI:             "eni-12345",
					AttachmentLevel: "pod-eni",
					SecurityGroups: []awsSDK.SecurityGroup{
						{
							GroupId:   strPtr("sg-11111"),
							GroupName: strPtr("sg-name-1"),
							IpPermissions: []awsSDK.IpPermission{
								{
									IpProtocol: strPtr("tcp"),
									FromPort:   int32Ptr(443),
									ToPort:     int32Ptr(443),
									IpRanges:   []awsSDK.IpRange{{CidrIp: strPtr("10.0.0.0/8")}},
								},
							},
						},
					},
				},
			},
			expected: `[{"podName":"pod1","namespace":"ns1","podIP":"10.0.0.1","eni":"eni-12345","attachmentLevel":"pod-eni","securityGroups":[{"id":"sg-11111","name":"sg-name-1","inboundRules":[{"protocol":"tcp","fromPort":443,"toPort":443,"sources":["10.0.0.0/8"]}]}]}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := OutputPodSecurityGroups(&buf, tc.data, tc.format, tc.sortField)

			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected error: %v", err)
			}

			if got := buf.String(); got != tc.expected {
				t.Errorf("unexpected output: got %q, want %q", got, tc.expected)
			}
		})
	}
}
