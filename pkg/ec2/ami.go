package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pet2cattle/kubectl-eks/pkg/data"
)

func GetAMIInfo(profile, region, ami string) (*data.AMIInfo, error) {
	ctx := context.Background()

	// Load the AWS configuration using the profile and region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create an EC2 client
	svc := ec2.NewFromConfig(cfg)

	// Describe the AMI
	result, err := svc.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: []string{ami},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe AMI %s for profile %s in region %s: %w", ami, profile, region, err)
	}

	if len(result.Images) == 0 {
		return nil, fmt.Errorf("no AMI found with ID %s", ami)
	}

	if len(result.Images) > 1 {
		return nil, fmt.Errorf("multiple AMIs found with ID %s", ami)
	}

	if aws.ToString(result.Images[0].ImageId) != ami {
		return nil, fmt.Errorf("unexpected AMI found with ID %s (searching for %q)", aws.ToString(result.Images[0].ImageId), ami)
	}

	info := data.AMIInfo{
		ID:           aws.ToString(result.Images[0].ImageId),
		Name:         aws.ToString(result.Images[0].Name),
		Architecture: string(result.Images[0].Architecture),
		State:        string(result.Images[0].State),
	}

	if result.Images[0].DeprecationTime != nil {
		info.DeprecationTime = *result.Images[0].DeprecationTime
	}

	return &info, nil
}
