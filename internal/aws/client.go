package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// GetENIForIPAddress は、指定された IP アドレスに関連付けられた ENI 情報を取得します。
func GetENIForIPAddress(ipAddress string) (string, []string, error) {
	// AWS SDKの設定を読み込む
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	// EC2 クライアントを作成
	client := ec2.NewFromConfig(cfg)

	// EC2 の DescribeNetworkInterfaces API を呼び出し、指定された IP アドレスに関連する ENI を取得
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("private-ip-address"),
				Values: []string{ipAddress},
			},
		},
	}

	result, err := client.DescribeNetworkInterfaces(context.TODO(), input)
	if err != nil {
		return "", nil, fmt.Errorf("failed to describe network interfaces: %v", err)
	}

	// ENI の情報を返す
	if len(result.NetworkInterfaces) == 0 {
		return "", nil, fmt.Errorf("no ENI found for IP address: %s", ipAddress)
	}

	eni := result.NetworkInterfaces[0]
	eniID := *eni.NetworkInterfaceId

	// セキュリティグループのIDリストを取得
	var sgIDs []string
	for _, sg := range eni.Groups {
		sgIDs = append(sgIDs, *sg.GroupId)
	}

	return eniID, sgIDs, nil
}
