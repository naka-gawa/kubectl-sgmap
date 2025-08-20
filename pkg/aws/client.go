package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/naka-gawa/kubectl-sgmap/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

// Client provides access to AWS EC2 APIs
type Client struct {
	ec2Client *ec2.Client
}

// Interface defines the methods provided by the AWS EC2 client.
// Used for dependency injection and testing.
type Interface interface {
	FetchSecurityGroupsByPods(ctx context.Context, pods []corev1.Pod) ([]PodSecurityGroupInfo, error)
}

// PodSecurityGroupInfo represents the security group information associated with a Pod
type PodSecurityGroupInfo struct {
	Pod            corev1.Pod            `json:"pod" yaml:"pod"`
	SecurityGroups []types.SecurityGroup `json:"securityGroups" yaml:"securityGroups"`
	ENI            string                `json:"eni" yaml:"eni"`
	AttachmentLevel string               `json:"attachmentLevel" yaml:"attachmentLevel"`
}

// NewClient creates a new AWS EC2 client
func NewClient() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	return &Client{
		ec2Client: ec2.NewFromConfig(cfg),
	}, nil
}

// FetchSecurityGroupsByPods fetches security groups associated with pods by resolving their ENIs
func (c *Client) FetchSecurityGroupsByPods(ctx context.Context, pods []corev1.Pod) ([]PodSecurityGroupInfo, error) {
	podIPs, ipToPod := filterRunningPodsWithIPs(pods)
	if len(podIPs) == 0 {
		return nil, nil
	}

	eniMap, ipToENI, eniToSGIDs, err := c.fetchENIAndSGIDs(ctx, podIPs)
	if err != nil {
		return nil, err
	}

	sgMap, err := c.GetSecurityGroupsParallel(ctx, collectUniqueSGIDs(eniToSGIDs))
	if err != nil {
		return nil, err
	}

	return buildPodSecurityGroupInfo(ipToPod, ipToENI, eniToSGIDs, sgMap, eniMap), nil
}

// filterRunningPodsWithIPs filters pods in the Running phase and extracts their IPs
func filterRunningPodsWithIPs(pods []corev1.Pod) ([]string, map[string]corev1.Pod) {
	var ips []string
	ipToPod := make(map[string]corev1.Pod)

	for _, pod := range pods {
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}
		if ip := pod.Status.PodIP; ip != "" {
			ips = append(ips, ip)
			ipToPod[ip] = pod
		}
	}
	return ips, ipToPod
}

// fetchENIAndSGIDs retrieves ENIs and extracts corresponding SG IDs from private IPs
func (c *Client) fetchENIAndSGIDs(ctx context.Context, podIPs []string) (
	map[string]types.NetworkInterface,
	map[string]string,
	map[string][]string,
	error,
) {
	eniMap, err := c.GetENIsByPrivateIPs(ctx, podIPs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to describe ENIs: %w", err)
	}

	ipToENI := make(map[string]string)
	eniToSGIDs := make(map[string][]string)
	for _, eni := range eniMap {
		eniID := aws.ToString(eni.NetworkInterfaceId)
		for _, ip := range eni.PrivateIpAddresses {
			ipToENI[aws.ToString(ip.PrivateIpAddress)] = eniID
		}
		for _, group := range eni.Groups {
			eniToSGIDs[eniID] = append(eniToSGIDs[eniID], aws.ToString(group.GroupId))
		}
	}
	return eniMap, ipToENI, eniToSGIDs, nil
}

// collectUniqueSGIDs deduplicates SG IDs from ENI to SG ID map
func collectUniqueSGIDs(eniToSGIDs map[string][]string) []string {
	set := map[string]struct{}{}
	for _, ids := range eniToSGIDs {
		for _, id := range ids {
			set[id] = struct{}{}
		}
	}
	var result []string
	for id := range set {
		result = append(result, id)
	}
	return result
}

// buildPodSecurityGroupInfo builds the final mapping between Pod and its associated SGs and ENI
func buildPodSecurityGroupInfo(
	ipToPod map[string]corev1.Pod,
	ipToENI map[string]string,
	eniToSGIDs map[string][]string,
	sgMap map[string]types.SecurityGroup,
	eniMap map[string]types.NetworkInterface,
) []PodSecurityGroupInfo {
	var result []PodSecurityGroupInfo

	for ip, pod := range ipToPod {
		eniID, ok := ipToENI[ip]
		if !ok {
			fmt.Printf("Warning: ENI not found for pod %s/%s (IP %s)\n", pod.Namespace, pod.Name, ip)
			continue
		}
		var sgs []types.SecurityGroup
		for _, sgID := range eniToSGIDs[eniID] {
			if sg, ok := sgMap[sgID]; ok {
				sgs = append(sgs, sg)
			}
		}
		
		// Determine attachment level based on ENI characteristics
		attachmentLevel := determineAttachmentLevel(eniMap[eniID])
		
		result = append(result, PodSecurityGroupInfo{
			Pod:             pod,
			ENI:             eniID,
			SecurityGroups:  sgs,
			AttachmentLevel: attachmentLevel,
		})
	}
	return result
}

// determineAttachmentLevel determines where the security group is attached based on ENI characteristics
func determineAttachmentLevel(eni types.NetworkInterface) string {
	// Check ENI type and description to determine attachment level
	interfaceType := string(eni.InterfaceType)
	description := aws.ToString(eni.Description)
	
	// AWS EKS patterns for different attachment levels
	if interfaceType == "branch" {
		return "pod"
	}
	
	if interfaceType == "interface" {
		// Check description for common patterns
		if description != "" {
			// EKS node ENIs typically have descriptions mentioning "EKS" or node group
			if containsAny(description, []string{"eks", "EKS", "node", "Node"}) {
				return "node"
			}
			// VPC CNI trunk interface
			if containsAny(description, []string{"trunk", "Trunk"}) {
				return "node"
			}
		}
		return "node" // Default for interface type
	}
	
	// For other types, try to infer from description or default to node
	if description != "" && containsAny(description, []string{"pod", "Pod"}) {
		return "pod"
	}
	
	return "node" // Default fallback
}

// containsAny checks if the string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substring := range substrings {
		if len(s) >= len(substring) {
			for i := 0; i <= len(s)-len(substring); i++ {
				if s[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}

// GetENIsByPrivateIPs retrieves network interfaces from EC2 based on private IP addresses using batch processing
func (c *Client) GetENIsByPrivateIPs(ctx context.Context, ips []string) (map[string]types.NetworkInterface, error) {
	if len(ips) == 0 {
		return nil, fmt.Errorf("input list of private IPs is empty")
	}

	return utils.RunBatchParallel(ctx, ips, 200, 5, func(ctx context.Context, batch []string) (map[string]types.NetworkInterface, error) {
		input := &ec2.DescribeNetworkInterfacesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("addresses.private-ip-address"),
					Values: batch,
				},
			},
		}

		paginator := ec2.NewDescribeNetworkInterfacesPaginator(c.ec2Client, input)
		result := make(map[string]types.NetworkInterface)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to paginate DescribeNetworkInterfaces: %w", err)
			}
			for _, eni := range page.NetworkInterfaces {
				result[aws.ToString(eni.NetworkInterfaceId)] = eni
			}
		}
		return result, nil
	})
}

// GetSecurityGroupsParallel retrieves security groups by their IDs using parallel processing.
// It deduplicates input IDs, splits them into batches of up to 200 IDs (AWS API limit),
// and processes each batch concurrently using a worker pool.
func (c *Client) GetSecurityGroupsParallel(ctx context.Context, sgIDs []string) (map[string]types.SecurityGroup, error) {
	seen := make(map[string]struct{})
	var deduped []string
	for _, id := range sgIDs {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			deduped = append(deduped, id)
		}
	}

	if len(deduped) == 0 {
		return map[string]types.SecurityGroup{}, nil
	}

	return utils.RunBatchParallel(ctx, deduped, 200, 5, func(ctx context.Context, ids []string) (map[string]types.SecurityGroup, error) {
		input := &ec2.DescribeSecurityGroupsInput{
			GroupIds: ids,
		}

		resp, err := c.ec2Client.DescribeSecurityGroups(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("DescribeSecurityGroups failed: %w", err)
		}

		sgMap := make(map[string]types.SecurityGroup)
		for _, sg := range resp.SecurityGroups {
			if sg.GroupId != nil {
				sgMap[*sg.GroupId] = sg
			}
		}
		return sgMap, nil
	})
}
