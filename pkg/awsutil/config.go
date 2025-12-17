package awsutil

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

// LoadConfig returns an AWS config for the given profile and region
func LoadConfig(profile, region string) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	return config.LoadDefaultConfig(context.TODO(), opts...)
}

// CloudFormationAPI defines the CloudFormation operations we use
type CloudFormationAPI interface {
	ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error)
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
}

// EKSAPI defines the EKS operations we use
type EKSAPI interface {
	ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error)
	DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
	ListNodegroups(ctx context.Context, params *eks.ListNodegroupsInput, optFns ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error)
	DescribeNodegroup(ctx context.Context, params *eks.DescribeNodegroupInput, optFns ...func(*eks.Options)) (*eks.DescribeNodegroupOutput, error)
	ListFargateProfiles(ctx context.Context, params *eks.ListFargateProfilesInput, optFns ...func(*eks.Options)) (*eks.ListFargateProfilesOutput, error)
	DescribeFargateProfile(ctx context.Context, params *eks.DescribeFargateProfileInput, optFns ...func(*eks.Options)) (*eks.DescribeFargateProfileOutput, error)
}

// Ensure AWS SDK clients implement our interfaces
var _ CloudFormationAPI = (*cloudformation.Client)(nil)
var _ EKSAPI = (*eks.Client)(nil)
