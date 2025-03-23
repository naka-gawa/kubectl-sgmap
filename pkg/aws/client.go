package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	corev1 "k8s.io/api/core/v1"
)

// Client provides AWS API access
type Client struct {
	ec2Client *ec2.Client
}

// PodSecurityGroupInfo represents security group information for a pod
type PodSecurityGroupInfo struct {
	Pod            corev1.Pod
	SecurityGroups []types.SecurityGroup
	ENI            string
}

// NewClient creates a new AWS client
func NewClient() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	return &Client{
		ec2Client: ec2.NewFromConfig(cfg),
	}, nil
}

// GetSecurityGroupsForPods retrieves security group information for the given pods
func (c *Client) GetSecurityGroupsForPods(ctx context.Context, pods []corev1.Pod) ([]PodSecurityGroupInfo, error) {
	result := make([]PodSecurityGroupInfo, 0, len(pods))

	for _, pod := range pods {
		if pod.Status.PodIP == "" {
			continue
		}

		eni, err := c.findENIForPodIP(ctx, pod.Status.PodIP)
		if err != nil {
			fmt.Printf("Warning: failed to find ENI for pod %s/%s: %v\n", pod.Namespace, pod.Name, err)
			continue
		}

		if eni == "" {
			continue
		}

		sgs, err := c.getSecurityGroupsForENI(ctx, eni)
		if err != nil {
			fmt.Printf("Warning: failed to get security groups for pod %s/%s: %v\n", pod.Namespace, pod.Name, err)
			continue
		}

		result = append(result, PodSecurityGroupInfo{
			Pod:            pod,
			SecurityGroups: sgs,
			ENI:            eni,
		})
	}

	return result, nil
}

func (c *Client) findENIForPodIP(ctx context.Context, podIP string) (string, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("private-ip-address"),
				Values: []string{podIP},
			},
		},
	}

	resp, err := c.ec2Client.DescribeNetworkInterfaces(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to describe network interfaces: %w", err)
	}

	if len(resp.NetworkInterfaces) == 0 {
		return "", nil
	}

	return *resp.NetworkInterfaces[0].NetworkInterfaceId, nil
}

// getSecurityGroupsForENI gets security groups attached to an ENI
func (c *Client) getSecurityGroupsForENI(ctx context.Context, eniID string) ([]types.SecurityGroup, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{eniID},
	}

	resp, err := c.ec2Client.DescribeNetworkInterfaces(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe network interface %s: %w", eniID, err)
	}

	if len(resp.NetworkInterfaces) == 0 {
		return nil, fmt.Errorf("network interface %s not found", eniID)
	}

	sgIDs := []string{}
	for _, sg := range resp.NetworkInterfaces[0].Groups {
		sgIDs = append(sgIDs, *sg.GroupId)
	}

	if len(sgIDs) == 0 {
		return []types.SecurityGroup{}, nil
	}

	sgInput := &ec2.DescribeSecurityGroupsInput{
		GroupIds: sgIDs,
	}

	sgResp, err := c.ec2Client.DescribeSecurityGroups(ctx, sgInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe security groups: %w", err)
	}

	return sgResp.SecurityGroups, nil
}
