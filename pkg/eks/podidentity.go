package eks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/data"
)

// PodIdentityAssociationInfo holds information about a pod identity association
type PodIdentityAssociationInfo struct {
	AssociationID      string
	AssociationArn     string
	ClusterName        string
	Namespace          string
	ServiceAccountName string
	RoleArn            string
}

// GetPodIdentityAssociations retrieves all pod identity associations for a cluster
func GetPodIdentityAssociations(profile, region, clusterName string) ([]data.PodIdentityInfo, error) {
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

	// List pod identity associations
	var allAssociations []data.PodIdentityInfo
	var nextToken *string

	for {
		input := &eks.ListPodIdentityAssociationsInput{
			ClusterName: aws.String(clusterName),
			MaxResults:  aws.Int32(100),
			NextToken:   nextToken,
		}

		result, err := client.ListPodIdentityAssociations(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list pod identity associations for cluster %s: %w", clusterName, err)
		}

		// Describe each association to get details
		for _, assoc := range result.Associations {
			if assoc.AssociationId == nil {
				continue
			}

			describeInput := &eks.DescribePodIdentityAssociationInput{
				ClusterName:   aws.String(clusterName),
				AssociationId: assoc.AssociationId,
			}

			descResult, err := client.DescribePodIdentityAssociation(ctx, describeInput)
			if err != nil {
				// Skip associations that can't be described
				continue
			}

			if descResult.Association != nil {
				assocData := descResult.Association
				info := data.PodIdentityInfo{
					Profile:            profile,
					Region:             region,
					ClusterName:        clusterName,
					Namespace:          aws.ToString(assocData.Namespace),
					ServiceAccountName: aws.ToString(assocData.ServiceAccount),
					IAMRoleARN:         aws.ToString(assocData.RoleArn),
					IdentityType:       "EKS Pod Identity",
				}
				allAssociations = append(allAssociations, info)
			}
		}

		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	return allAssociations, nil
}
