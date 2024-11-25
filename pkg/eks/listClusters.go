package eks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

func GetClusters(profile, region string) ([]*string, error) {
	// Create a new session using the profile and region
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Create an EKS client
	svc := eks.New(sess)

	// List EKS clusters in the specified region
	input := &eks.ListClustersInput{}
	result, err := svc.ListClusters(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters for profile %s in region %s: %w", profile, region, err)
	}

	// Return the cluster names
	return result.Clusters, nil
}
