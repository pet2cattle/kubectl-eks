package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pet2cattle/kubectl-eks/pkg/data"
)

func GetAMIInfo(profile, region, ami string) (*data.AMIInfo, error) {
	// Create a new session using the profile and region
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Create an EC2 client
	svc := ec2.New(sess)

	// Describe the AMI
	input := &ec2.DescribeImagesInput{
		ImageIds: []*string{
			aws.String(ami),
		},
	}

	result, err := svc.DescribeImages(input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe AMI %s for profile %s in region %s: %w", ami, profile, region, err)
	}

	if len(result.Images) == 0 {
		return nil, fmt.Errorf("no AMI found with ID %s", ami)
	}

	if len(result.Images) > 1 {
		return nil, fmt.Errorf("multiple AMIs found with ID %s", ami)
	}

	if *result.Images[0].ImageId != ami {
		return nil, fmt.Errorf("unexpected AMI found with ID %s (searching for %q)", *result.Images[0].ImageId, ami)
	}

	info := data.AMIInfo{
		ID:           *result.Images[0].ImageId,
		Name:         *result.Images[0].Name,
		Architecture: *result.Images[0].Architecture,
		State:        *result.Images[0].State,
	}

	if result.Images[0].DeprecationTime != nil {
		info.DeprecationTime = *result.Images[0].DeprecationTime
	}

	return &info, nil
}
