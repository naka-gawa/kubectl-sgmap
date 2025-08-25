package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


// MockEC2Client is a mock implementation of the EC2API interface
type MockEC2Client struct {
	mock.Mock
}

func (m *MockEC2Client) DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ec2.DescribeNetworkInterfacesOutput), args.Error(1)
}

func (m *MockEC2Client) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ec2.DescribeSecurityGroupsOutput), args.Error(1)
}

func TestGetENIsByPrivateIPs_Success(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	ips := make([]string, 250)
	for i := 0; i < 250; i++ {
		ips[i] = fmt.Sprintf("10.0.0.%d", i)
	}

	mockClient.On("DescribeNetworkInterfaces", mock.Anything, mock.Anything).Return(
		&ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []types.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-12345678"),
					PrivateIpAddresses: []types.NetworkInterfacePrivateIpAddress{
						{PrivateIpAddress: aws.String("10.0.0.1")},
					},
				},
			},
		}, nil).Twice()

	enis, err := client.GetENIsByPrivateIPs(context.Background(), ips)

	assert.NoError(t, err)
	assert.NotNil(t, enis)
	assert.Len(t, enis, 1)
	mockClient.AssertExpectations(t)
}

func TestGetENIsByPrivateIPs_Error(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	ips := []string{"10.0.0.1"}

	mockClient.On("DescribeNetworkInterfaces", mock.Anything, mock.Anything).Return(
		nil, fmt.Errorf("aws error"),
	)

	_, err := client.GetENIsByPrivateIPs(context.Background(), ips)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestFetchSecurityGroupsByPods_SGError(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.1"},
		},
	}

	mockClient.On("DescribeNetworkInterfaces", mock.Anything, mock.Anything).Return(
		&ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []types.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-1"),
					Groups:             []types.GroupIdentifier{{GroupId: aws.String("sg-1")}},
					PrivateIpAddresses: []types.NetworkInterfacePrivateIpAddress{
						{PrivateIpAddress: aws.String("10.0.0.1")},
					},
				},
			},
		}, nil,
	)

	mockClient.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(
		nil, fmt.Errorf("aws error"),
	)

	_, err := client.FetchSecurityGroupsByPods(context.Background(), pods)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestGetSecurityGroupsParallel_Empty(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	sgMap, err := client.GetSecurityGroupsParallel(context.Background(), []string{})

	assert.NoError(t, err)
	assert.Empty(t, sgMap)
	mockClient.AssertNotCalled(t, "DescribeSecurityGroups")
}

func TestFilterRunningPodsWithIPs(t *testing.T) {
	pods := []corev1.Pod{
		{Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.1"}},
		{Status: corev1.PodStatus{Phase: corev1.PodPending, PodIP: "10.0.0.2"}},
		{Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: ""}},
	}

	ips, ipToPod := filterRunningPodsWithIPs(pods)

	assert.Equal(t, []string{"10.0.0.1"}, ips)
	assert.Len(t, ipToPod, 1)
}

func TestCollectUniqueSGIDs(t *testing.T) {
	eniToSGIDs := map[string][]string{
		"eni-1": {"sg-1", "sg-2"},
		"eni-2": {"sg-2", "sg-3"},
	}

	sgIDs := collectUniqueSGIDs(eniToSGIDs)

	assert.ElementsMatch(t, []string{"sg-1", "sg-2", "sg-3"}, sgIDs)
}

func TestDetermineAttachmentLevel(t *testing.T) {
	testCases := []struct {
		name     string
		eni      types.NetworkInterface
		expected string
	}{
		{"BranchENI", types.NetworkInterface{InterfaceType: "branch"}, "pod"},
		{"NodeENI", types.NetworkInterface{InterfaceType: "interface", Description: aws.String("aws-K8S-i-...")}, "node"},
		{"TrunkENI", types.NetworkInterface{InterfaceType: "interface", Description: aws.String("aws-k8s-trunk-eni")}, "node"},
		{"PodENI", types.NetworkInterface{InterfaceType: "interface", Description: aws.String("pod-eni")}, "pod"},
		{"Default", types.NetworkInterface{InterfaceType: "other"}, "node"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			level := determineAttachmentLevel(tc.eni)
			assert.Equal(t, tc.expected, level)
		})
	}
}

func TestBuildPodSecurityGroupInfo(t *testing.T) {
	ipToPod := map[string]corev1.Pod{
		"10.0.0.1": {ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
	}
	ipToENI := map[string]string{"10.0.0.1": "eni-1"}
	eniToSGIDs := map[string][]string{"eni-1": {"sg-1"}}
	sgMap := map[string]types.SecurityGroup{"sg-1": {GroupId: aws.String("sg-1")}}
	eniMap := map[string]types.NetworkInterface{"eni-1": {NetworkInterfaceId: aws.String("eni-1")}}

	result := buildPodSecurityGroupInfo(ipToPod, ipToENI, eniToSGIDs, sgMap, eniMap)

	assert.Len(t, result, 1)
	assert.Equal(t, "pod1", result[0].Pod.Name)
	assert.Equal(t, "eni-1", result[0].ENI)
	assert.Len(t, result[0].SecurityGroups, 1)
}

func TestFetchSecurityGroupsByPods_Success(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.1"},
		},
	}

	mockClient.On("DescribeNetworkInterfaces", mock.Anything, mock.Anything).Return(
		&ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []types.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-1"),
					Groups:             []types.GroupIdentifier{{GroupId: aws.String("sg-1")}},
					PrivateIpAddresses: []types.NetworkInterfacePrivateIpAddress{
						{PrivateIpAddress: aws.String("10.0.0.1")},
					},
				},
			},
		}, nil,
	)

	mockClient.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(
		&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{GroupId: aws.String("sg-1"), GroupName: aws.String("my-sg")},
			},
		}, nil,
	)

	result, err := client.FetchSecurityGroupsByPods(context.Background(), pods)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "pod1", result[0].Pod.Name)
	assert.Len(t, result[0].SecurityGroups, 1)
	assert.Equal(t, "sg-1", *result[0].SecurityGroups[0].GroupId)
	mockClient.AssertExpectations(t)
}

func TestFetchSecurityGroupsByPods_NoRunningPods(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
			Status:     corev1.PodStatus{Phase: corev1.PodPending},
		},
	}

	result, err := client.FetchSecurityGroupsByPods(context.Background(), pods)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockClient.AssertNotCalled(t, "DescribeNetworkInterfaces")
	mockClient.AssertNotCalled(t, "DescribeSecurityGroups")
}

func TestFetchSecurityGroupsByPods_ENIError(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.1"},
		},
	}

	mockClient.On("DescribeNetworkInterfaces", mock.Anything, mock.Anything).Return(
		nil, fmt.Errorf("aws error"),
	)

	_, err := client.FetchSecurityGroupsByPods(context.Background(), pods)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestGetENIsByPrivateIPs_EmptyIPs(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	_, err := client.GetENIsByPrivateIPs(context.Background(), []string{})

	assert.Error(t, err)
	mockClient.AssertNotCalled(t, "DescribeNetworkInterfaces")
}

func TestGetSecurityGroupsParallel_Success(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	sgIDs := []string{"sg-1", "sg-2", "sg-1", ""}

	mockClient.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(
		&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{GroupId: aws.String("sg-1")},
				{GroupId: aws.String("sg-2")},
			},
		}, nil,
	)

	sgMap, err := client.GetSecurityGroupsParallel(context.Background(), sgIDs)

	assert.NoError(t, err)
	assert.Len(t, sgMap, 2)
	mockClient.AssertExpectations(t)
}

func TestGetSecurityGroupsParallel_Error(t *testing.T) {
	mockClient := new(MockEC2Client)
	client := &Client{ec2Client: mockClient}

	sgIDs := []string{"sg-1"}

	mockClient.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(
		nil, fmt.Errorf("aws error"),
	)

	_, err := client.GetSecurityGroupsParallel(context.Background(), sgIDs)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}
