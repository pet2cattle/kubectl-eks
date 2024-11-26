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

func DescribeCluster(profile, region, clusterName string) (*eks.Cluster, error) {
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

	// Describe the EKS cluster
	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}
	result, err := svc.DescribeCluster(input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster %s for profile %s in region %s: %w", clusterName, profile, region, err)
	}

	// Return the cluster
	return result.Cluster, nil
}
