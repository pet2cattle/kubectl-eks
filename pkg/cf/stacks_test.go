package cf

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCloudFormationAPI is a mock implementation of CloudFormationAPI
type MockCloudFormationAPI struct {
	mock.Mock
}

func (m *MockCloudFormationAPI) ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cloudformation.ListStacksOutput), args.Error(1)
}

func (m *MockCloudFormationAPI) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cloudformation.DescribeStacksOutput), args.Error(1)
}

func TestGetStacks(t *testing.T) {
	tests := []struct {
		name           string
		searchName     string
		mockResponse   *cloudformation.ListStacksOutput
		mockError      error
		expectedStacks []StackInfo
		expectedError  bool
	}{
		{
			name:       "returns matching stacks",
			searchName: "my-cluster",
			mockResponse: &cloudformation.ListStacksOutput{
				StackSummaries: []types.StackSummary{
					{
						StackName:   aws.String("my-cluster-nodegroup"),
						StackStatus: types.StackStatusCreateComplete,
					},
					{
						StackName:   aws.String("my-cluster-vpc"),
						StackStatus: types.StackStatusCreateComplete,
					},
					{
						StackName:   aws.String("other-stack"),
						StackStatus: types.StackStatusCreateComplete,
					},
				},
			},
			expectedStacks: []StackInfo{
				{Name: "my-cluster-nodegroup", Status: "CREATE_COMPLETE"},
				{Name: "my-cluster-vpc", Status: "CREATE_COMPLETE"},
			},
			expectedError: false,
		},
		{
			name:       "returns empty when no matches",
			searchName: "nonexistent",
			mockResponse: &cloudformation.ListStacksOutput{
				StackSummaries: []types.StackSummary{
					{
						StackName:   aws.String("other-stack"),
						StackStatus: types.StackStatusCreateComplete,
					},
				},
			},
			expectedStacks: nil,
			expectedError:  false,
		},
		{
			name:           "handles API error",
			searchName:     "my-cluster",
			mockResponse:   nil,
			mockError:      errors.New("API error"),
			expectedStacks: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := new(MockCloudFormationAPI)
			mockAPI.On("ListStacks", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)

			client := NewClientWithAPI(mockAPI)
			stacks, err := client.GetStacks(tt.searchName)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStacks, stacks)
			}

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestGetStacksByParameter(t *testing.T) {
	tests := []struct {
		name           string
		paramName      string
		paramValue     string
		mockResponse   *cloudformation.DescribeStacksOutput
		mockError      error
		expectedStacks []StackInfo
		expectedError  bool
	}{
		{
			name:       "returns stacks with matching ClusterName parameter",
			paramName:  "ClusterName",
			paramValue: "my-cluster",
			mockResponse: &cloudformation.DescribeStacksOutput{
				Stacks: []types.Stack{
					{
						StackName:   aws.String("nodegroup-stack"),
						StackStatus: types.StackStatusCreateComplete,
						Parameters: []types.Parameter{
							{ParameterKey: aws.String("ClusterName"), ParameterValue: aws.String("my-cluster")},
						},
					},
					{
						StackName:   aws.String("other-stack"),
						StackStatus: types.StackStatusCreateComplete,
						Parameters: []types.Parameter{
							{ParameterKey: aws.String("ClusterName"), ParameterValue: aws.String("other-cluster")},
						},
					},
				},
			},
			expectedStacks: []StackInfo{
				{Name: "nodegroup-stack", Status: "CREATE_COMPLETE"},
			},
			expectedError: false,
		},
		{
			name:       "skips deleted stacks",
			paramName:  "ClusterName",
			paramValue: "my-cluster",
			mockResponse: &cloudformation.DescribeStacksOutput{
				Stacks: []types.Stack{
					{
						StackName:   aws.String("deleted-stack"),
						StackStatus: types.StackStatusDeleteComplete,
						Parameters: []types.Parameter{
							{ParameterKey: aws.String("ClusterName"), ParameterValue: aws.String("my-cluster")},
						},
					},
					{
						StackName:   aws.String("active-stack"),
						StackStatus: types.StackStatusCreateComplete,
						Parameters: []types.Parameter{
							{ParameterKey: aws.String("ClusterName"), ParameterValue: aws.String("my-cluster")},
						},
					},
				},
			},
			expectedStacks: []StackInfo{
				{Name: "active-stack", Status: "CREATE_COMPLETE"},
			},
			expectedError: false,
		},
		{
			name:       "returns empty when parameter not found",
			paramName:  "ClusterName",
			paramValue: "my-cluster",
			mockResponse: &cloudformation.DescribeStacksOutput{
				Stacks: []types.Stack{
					{
						StackName:   aws.String("stack-without-param"),
						StackStatus: types.StackStatusCreateComplete,
						Parameters:  []types.Parameter{},
					},
				},
			},
			expectedStacks: nil,
			expectedError:  false,
		},
		{
			name:           "handles API error",
			paramName:      "ClusterName",
			paramValue:     "my-cluster",
			mockResponse:   nil,
			mockError:      errors.New("API error"),
			expectedStacks: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := new(MockCloudFormationAPI)
			mockAPI.On("DescribeStacks", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)

			client := NewClientWithAPI(mockAPI)
			stacks, err := client.GetStacksByParameter(tt.paramName, tt.paramValue)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStacks, stacks)
			}

			mockAPI.AssertExpectations(t)
		})
	}
}

func TestGetStacksPagination(t *testing.T) {
	mockAPI := new(MockCloudFormationAPI)

	// First page
	mockAPI.On("ListStacks", mock.Anything, &cloudformation.ListStacksInput{NextToken: (*string)(nil)}).Return(
		&cloudformation.ListStacksOutput{
			StackSummaries: []types.StackSummary{
				{StackName: aws.String("my-cluster-stack1"), StackStatus: types.StackStatusCreateComplete},
			},
			NextToken: aws.String("token1"),
		}, nil).Once()

	// Second page
	mockAPI.On("ListStacks", mock.Anything, &cloudformation.ListStacksInput{NextToken: aws.String("token1")}).Return(
		&cloudformation.ListStacksOutput{
			StackSummaries: []types.StackSummary{
				{StackName: aws.String("my-cluster-stack2"), StackStatus: types.StackStatusCreateComplete},
			},
			NextToken: nil,
		}, nil).Once()

	client := NewClientWithAPI(mockAPI)
	stacks, err := client.GetStacks("my-cluster")

	assert.NoError(t, err)
	assert.Len(t, stacks, 2)
	assert.Equal(t, "my-cluster-stack1", stacks[0].Name)
	assert.Equal(t, "my-cluster-stack2", stacks[1].Name)

	mockAPI.AssertExpectations(t)
}

func TestHasParameterValue(t *testing.T) {
	tests := []struct {
		name     string
		params   []types.Parameter
		key      string
		value    string
		expected bool
	}{
		{
			name: "finds matching parameter",
			params: []types.Parameter{
				{ParameterKey: aws.String("ClusterName"), ParameterValue: aws.String("my-cluster")},
			},
			key:      "ClusterName",
			value:    "my-cluster",
			expected: true,
		},
		{
			name: "returns false for wrong value",
			params: []types.Parameter{
				{ParameterKey: aws.String("ClusterName"), ParameterValue: aws.String("other-cluster")},
			},
			key:      "ClusterName",
			value:    "my-cluster",
			expected: false,
		},
		{
			name: "returns false for missing key",
			params: []types.Parameter{
				{ParameterKey: aws.String("OtherParam"), ParameterValue: aws.String("value")},
			},
			key:      "ClusterName",
			value:    "my-cluster",
			expected: false,
		},
		{
			name:     "handles empty params",
			params:   []types.Parameter{},
			key:      "ClusterName",
			value:    "my-cluster",
			expected: false,
		},
		{
			name: "handles nil ParameterValue",
			params: []types.Parameter{
				{ParameterKey: aws.String("ClusterName"), ParameterValue: nil},
			},
			key:      "ClusterName",
			value:    "my-cluster",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasParameterValue(tt.params, tt.key, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
