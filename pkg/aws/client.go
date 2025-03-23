package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	corev1 "k8s.io/api/core/v1"
)

// Client provides access to AWS API
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
		return nil, err
	}

	return &Client{
		ec2Client: ec2.NewFromConfig(cfg),
	}, nil
}

// GetSecurityGroupsForPods retrieves security groups for the given pods
func (c *Client) GetSecurityGroupsForPods(ctx context.Context, pods []corev1.Pod) ([]PodSecurityGroupInfo, error) {
	// ここでは実装の概要のみを示す
	// AWS APIを呼び出してポッドのENIとセキュリティグループ情報を取得
	result := make([]PodSecurityGroupInfo, 0, len(pods))

	for _, pod := range pods {
		// Pod IPから ENI を特定
		// ENI に紐づくセキュリティグループを取得
		// 結果を格納
		result = append(result, PodSecurityGroupInfo{
			Pod:            pod,
			SecurityGroups: []types.SecurityGroup{}, // 実際の値を設定
			ENI:            "",                      // 実際の値を設定
		})
	}

	return result, nil
}
