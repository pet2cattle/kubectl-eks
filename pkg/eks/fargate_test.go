package eks

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEKSAPI is a mock implementation of EKSAPI
type MockEKSAPI struct {
	mock.Mock
}

func (m *MockEKSAPI) ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eks.ListClustersOutput), args.Error(1)
}

func (m *MockEKSAPI) DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eks.DescribeClusterOutput), args.Error(1)
}

func (m *MockEKSAPI) ListNodegroups(ctx context.Context, params *eks.ListNodegroupsInput, optFns ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eks.ListNodegroupsOutput), args.Error(1)
}

func (m *MockEKSAPI) DescribeNodegroup(ctx context.Context, params *eks.DescribeNodegroupInput, optFns ...func(*eks.Options)) (*eks.DescribeNodegroupOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eks.DescribeNodegroupOutput), args.Error(1)
}

func (m *MockEKSAPI) ListFargateProfiles(ctx context.Context, params *eks.ListFargateProfilesInput, optFns ...func(*eks.Options)) (*eks.ListFargateProfilesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eks.ListFargateProfilesOutput), args.Error(1)
}

func (m *MockEKSAPI) DescribeFargateProfile(ctx context.Context, params *eks.DescribeFargateProfileInput, optFns ...func(*eks.Options)) (*eks.DescribeFargateProfileOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eks.DescribeFargateProfileOutput), args.Error(1)
}

func TestGetFargateProfiles(t *testing.T) {
	tests := []struct {
		name             string
		clusterName      string
		listResponse     *eks.ListFargateProfilesOutput
		listError        error
		describeResponse *eks.DescribeFargateProfileOutput
		describeError    error
		expected         []FargateProfileInfo
		expectedError    bool
	}{
		{
			name:        "returns fargate profiles with selectors",
			clusterName: "my-cluster",
			listResponse: &eks.ListFargateProfilesOutput{
				FargateProfileNames: []string{"kube-system-profile"},
			},
			describeResponse: &eks.DescribeFargateProfileOutput{
				FargateProfile: &types.FargateProfile{
					FargateProfileName:  aws.String("kube-system-profile"),
					Status:              types.FargateProfileStatusActive,
					PodExecutionRoleArn: aws.String("arn:aws:iam::123456789012:role/fargate-role"),
					Subnets:             []string{"subnet-1", "subnet-2"},
					Selectors: []types.FargateProfileSelector{
						{
							Namespace: aws.String("kube-system"),
							Labels:    map[string]string{"app": "coredns"},
						},
					},
				},
			},
			expected: []FargateProfileInfo{
				{
					Name:           "kube-system-profile",
					Status:         "ACTIVE",
					PodExecRoleArn: "arn:aws:iam::123456789012:role/fargate-role",
					Subnets:        []string{"subnet-1", "subnet-2"},
					Selectors: []FargateSelector{
						{Namespace: "kube-system", Labels: map[string]string{"app": "coredns"}},
					},
				},
			},
			expectedError: false,
		},
		{
			name:        "returns empty when no profiles",
			clusterName: "my-cluster",
			listResponse: &eks.ListFargateProfilesOutput{
				FargateProfileNames: []string{},
			},
			expected:      nil,
			expectedError: false,
		},
		{
			name:          "handles list error",
			clusterName:   "my-cluster",
			listResponse:  nil,
			listError:     errors.New("API error"),
			expected:      nil,
			expectedError: true,
		},
		{
			name:        "handles describe error",
			clusterName: "my-cluster",
			listResponse: &eks.ListFargateProfilesOutput{
				FargateProfileNames: []string{"failing-profile"},
			},
			describeResponse: nil,
			describeError:    errors.New("describe error"),
			expected:         nil,
			expectedError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := new(MockEKSAPI)

			mockAPI.On("ListFargateProfiles", mock.Anything, mock.MatchedBy(func(input *eks.ListFargateProfilesInput) bool {
				return *input.ClusterName == tt.clusterName
			})).Return(tt.listResponse, tt.listError)

			if tt.listResponse != nil && len(tt.listResponse.FargateProfileNames) > 0 {
				mockAPI.On("DescribeFargateProfile", mock.Anything, mock.Anything).Return(tt.describeResponse, tt.describeError)
			}

			client := NewClientWithAPI(mockAPI)
			profiles, err := client.GetFargateProfiles(tt.clusterName)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, profiles)
			}

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestGetFargateProfilesPagination(t *testing.T) {
	mockAPI := new(MockEKSAPI)

	// First page
	mockAPI.On("ListFargateProfiles", mock.Anything, mock.MatchedBy(func(input *eks.ListFargateProfilesInput) bool {
		return input.NextToken == nil
	})).Return(&eks.ListFargateProfilesOutput{
		FargateProfileNames: []string{"profile-1"},
		NextToken:           aws.String("token1"),
	}, nil).Once()

	// Second page
	mockAPI.On("ListFargateProfiles", mock.Anything, mock.MatchedBy(func(input *eks.ListFargateProfilesInput) bool {
		return input.NextToken != nil && *input.NextToken == "token1"
	})).Return(&eks.ListFargateProfilesOutput{
		FargateProfileNames: []string{"profile-2"},
		NextToken:           nil,
	}, nil).Once()

	// Describe calls for each profile
	mockAPI.On("DescribeFargateProfile", mock.Anything, mock.MatchedBy(func(input *eks.DescribeFargateProfileInput) bool {
		return *input.FargateProfileName == "profile-1"
	})).Return(&eks.DescribeFargateProfileOutput{
		FargateProfile: &types.FargateProfile{
			FargateProfileName: aws.String("profile-1"),
			Status:             types.FargateProfileStatusActive,
			Selectors:          []types.FargateProfileSelector{},
		},
	}, nil).Once()

	mockAPI.On("DescribeFargateProfile", mock.Anything, mock.MatchedBy(func(input *eks.DescribeFargateProfileInput) bool {
		return *input.FargateProfileName == "profile-2"
	})).Return(&eks.DescribeFargateProfileOutput{
		FargateProfile: &types.FargateProfile{
			FargateProfileName: aws.String("profile-2"),
			Status:             types.FargateProfileStatusActive,
			Selectors:          []types.FargateProfileSelector{},
		},
	}, nil).Once()

	client := NewClientWithAPI(mockAPI)
	profiles, err := client.GetFargateProfiles("my-cluster")

	assert.NoError(t, err)
	assert.Len(t, profiles, 2)
	assert.Equal(t, "profile-1", profiles[0].Name)
	assert.Equal(t, "profile-2", profiles[1].Name)

	mockAPI.AssertExpectations(t)
}

func TestGetFargateProfilesMultipleSelectors(t *testing.T) {
	mockAPI := new(MockEKSAPI)

	mockAPI.On("ListFargateProfiles", mock.Anything, mock.Anything).Return(&eks.ListFargateProfilesOutput{
		FargateProfileNames: []string{"multi-selector-profile"},
	}, nil)

	mockAPI.On("DescribeFargateProfile", mock.Anything, mock.Anything).Return(&eks.DescribeFargateProfileOutput{
		FargateProfile: &types.FargateProfile{
			FargateProfileName: aws.String("multi-selector-profile"),
			Status:             types.FargateProfileStatusActive,
			Subnets:            []string{"subnet-1"},
			Selectors: []types.FargateProfileSelector{
				{Namespace: aws.String("kube-system"), Labels: nil},
				{Namespace: aws.String("default"), Labels: map[string]string{"env": "prod"}},
				{Namespace: aws.String("monitoring"), Labels: map[string]string{"app": "prometheus", "team": "infra"}},
			},
		},
	}, nil)

	client := NewClientWithAPI(mockAPI)
	profiles, err := client.GetFargateProfiles("my-cluster")

	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Len(t, profiles[0].Selectors, 3)
	assert.Equal(t, "kube-system", profiles[0].Selectors[0].Namespace)
	assert.Nil(t, profiles[0].Selectors[0].Labels)
	assert.Equal(t, "default", profiles[0].Selectors[1].Namespace)
	assert.Equal(t, map[string]string{"env": "prod"}, profiles[0].Selectors[1].Labels)
	assert.Equal(t, "monitoring", profiles[0].Selectors[2].Namespace)

	mockAPI.AssertExpectations(t)
}
