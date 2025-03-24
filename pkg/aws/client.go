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

// FetchSecurityGroupsByPods retrieves all security groups attached to the specified Elastic Network Interface (ENI).
// It queries the AWS EC2 API to get the network interface details and extracts the associated security groups.
// If the ENI is not found or has no security groups attached, appropriate errors are returned.
// This method performs two AWS API calls: one to retrieve the ENI information and another to get the detailed
// security group information based on the group IDs obtained from the ENI.
func (c *Client) FetchSecurityGroupsByPods(ctx context.Context, pods []corev1.Pod) ([]PodSecurityGroupInfo, error) {
	podIPs, ipToPod := filterRunningPodsWithIPs(pods)
	if len(podIPs) == 0 {
		return nil, nil
	}

	_, ipToENI, eniToSGIDs, err := c.fetchENIAndSGIDs(ctx, podIPs)
	if err != nil {
		return nil, err
	}

	sgMap, err := c.GetSecurityGroupsParallel(ctx, collectUniqueSGIDs(eniToSGIDs))
	if err != nil {
		return nil, err
	}

	return buildPodSecurityGroupInfo(ipToPod, ipToENI, eniToSGIDs, sgMap), nil
}

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

func buildPodSecurityGroupInfo(ipToPod map[string]corev1.Pod, ipToENI map[string]string, eniToSGIDs map[string][]string, sgMap map[string]types.SecurityGroup) []PodSecurityGroupInfo {
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
		result = append(result, PodSecurityGroupInfo{
			Pod:            pod,
			ENI:            eniID,
			SecurityGroups: sgs,
		})
	}
	return result
}

func (c *Client) GetENIsByPrivateIPs(ctx context.Context, ips []string) (map[string]types.NetworkInterface, error) {
	if len(ips) == 0 {
		return nil, nil
	}

	const batchSize = 200
	result := make(map[string]types.NetworkInterface)

	for i := 0; i < len(ips); i += batchSize {
		end := i + batchSize
		if end > len(ips) {
			end = len(ips)
		}
		batch := ips[i:end]

		input := &ec2.DescribeNetworkInterfacesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("addresses.private-ip-address"),
					Values: batch,
				},
			},
		}

		paginator := ec2.NewDescribeNetworkInterfacesPaginator(c.ec2Client, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to paginate DescribeNetworkInterfaces: %w", err)
			}
			for _, eni := range page.NetworkInterfaces {
				result[aws.ToString(eni.NetworkInterfaceId)] = eni
			}
		}
	}

	return result, nil
}

// GetSecurityGroupsParallel retrieves security groups by their IDs using parallel processing.
// It deduplicates input IDs, splits them into batches of up to 200 IDs (AWS API limit),
// and processes each batch concurrently using a worker pool.
func (c *Client) GetSecurityGroupsParallel(ctx context.Context, sgIDs []string) (map[string]types.SecurityGroup, error) {
	// ユニークな SG ID を集める
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

	// バッチ処理ユーティリティを使って並列取得
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
