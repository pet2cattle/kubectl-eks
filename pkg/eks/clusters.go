package eks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

func GetClusters(profile, region string) ([]*string, error) {
	ctx := context.Background()

	// Load the AWS configuration using the profile and region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create an EKS client
	client := eks.NewFromConfig(cfg)

	// List EKS clusters in the specified region
	result, err := client.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters for profile %s in region %s: %w", profile, region, err)
	}

	// Convert []string to []*string for compatibility
	clusters := make([]*string, len(result.Clusters))
	for i, name := range result.Clusters {
		clusters[i] = aws.String(name)
	}

	return clusters, nil
}

func DescribeCluster(profile, region, clusterName string) (*types.Cluster, error) {
	ctx := context.Background()

	// Load the AWS configuration using the profile and region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create an EKS client
	client := eks.NewFromConfig(cfg)

	// Describe the EKS cluster
	result, err := client.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster %s for profile %s in region %s: %w", clusterName, profile, region, err)
	}

	return result.Cluster, nil
}
