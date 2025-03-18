package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// GetENIForIPAddress retrieves the Elastic Network Interface (ENI) ID and associated security group IDs for a given IP address.
//
// Parameters:
//   - ipAddress: The IP address for which to find the associated ENI.
//
// Returns:
//   - string: The ID of the ENI associated with the given IP address.
//   - []string: A list of security group IDs associated with the ENI.
//   - error: An error object if there was an issue retrieving the ENI information.
//
// The function uses the AWS SDK to load the default configuration and create an EC2 client.
// It then calls the DescribeNetworkInterfaces API to find the ENI associated with the specified IP address.
// If no ENI is found, an error is returned. Otherwise, the ENI ID and associated security group IDs are returned.
func GetENIForIPAddress(ipAddress string) (string, []string, error) {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	// Create EC2 client
	client := ec2.NewFromConfig(cfg)

	// Call EC2 DescribeNetworkInterfaces API to get the ENI associated with the specified IP address
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

	// Return ENI information
	if len(result.NetworkInterfaces) == 0 {
		return "", nil, fmt.Errorf("no ENI found for IP address: %s", ipAddress)
	}

	eni := result.NetworkInterfaces[0]
	eniID := *eni.NetworkInterfaceId

	// Get the list of security group IDs
	var sgIDs []string
	for _, sg := range eni.Groups {
		sgIDs = append(sgIDs, *sg.GroupId)
	}

	return eniID, sgIDs, nil
}
