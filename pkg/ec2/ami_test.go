package ec2

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pet2cattle/kubectl-eks/pkg/data"
)

// MockEC2Client is a mock implementation of the EC2 client.
type MockEC2Client struct {
	mockDescribeImages func(ctx context.Context, input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)
}

func (m *MockEC2Client) DescribeImages(ctx context.Context, input *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	return m.mockDescribeImages(ctx, input)
}

func TestGetAMIInfo(t *testing.T) {
	tests := []struct {
		name        string
		mockClient  *MockEC2Client
		expectedErr bool
	}{
		{
			name: "Successful AMI retrieval",
			mockClient: &MockEC2Client{
				mockDescribeImages: func(ctx context.Context, input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []ec2types.Image{
							{
								ImageId: aws.String("ami-12345678"),
								Name:    aws.String("my-ami"),
								Architecture: ec2types.ArchitectureValuesX86_64,
								State: &ec2types.ImageState{
									Name: ec2types.ImageStateNameAvailable,
								},
							},
						},
					}, nil
				},
			},
			expectedErr: false,
		},
		{
			name: "Multiple AMIs found",
			mockClient: &MockEC2Client{
				mockDescribeImages: func(ctx context.Context, input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []ec2types.Image{
							{
								ImageId: aws.String("ami-12345678"),
							},
							{
								ImageId: aws.String("ami-87654321"),
							},
						},
					}, nil
				},
			},
			expectedErr: true,
		},
		{
			name: "No AMI found",
			mockClient: &MockEC2Client{
				mockDescribeImages: func(ctx context.Context, input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{}, nil
				},
			},
			expectedErr: true,
		},
		{
			name: "Unexpected AMI found",
			mockClient: &MockEC2Client{
				mockDescribeImages: func(ctx context.Context, input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []ec2types.Image{
							{
								ImageId: aws.String("ami-87654321"),
							},
						},
					}, nil
				},
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(context.Background(),
				config.WithRegion("us-west-2"),
			)
			if err != nil {
				t.Fatalf("failed to load default config: %v", err)
			}

			svc := ec2.NewFromConfig(cfg)
			// Replace the real client with the mock client
			realDescribeImages := svc.DescribeImages
			defer func() { svc.DescribeImages = realDescribeImages }()
			svc.DescribeImages = tt.mockClient.DescribeImages

			_, err = GetAMIInfo("default", "us-west-2", "ami-12345678")
			if (err != nil) != tt.expectedErr {
				t.Errorf("GetAMIInfo() error = %v, wantErr %v", err, tt.expectedErr)
			}
		})
	}
}