package cf

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type StackInfo struct {
	Name   string
	Status string
}

func GetStacks(name, profile, region string) ([]StackInfo, error) {
	stackList := []StackInfo{}

	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	svc := cloudformation.New(sess)

	var nextToken *string
	for {
		response, err := svc.ListStacks(&cloudformation.ListStacksInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list stacks for profile %s in region %s: %w", profile, region, err)
		}

		for _, stack := range response.StackSummaries {
			if stack != nil {
				if strings.Contains(*stack.StackName, name) {
					stackList = append(stackList, StackInfo{
						Name:   *stack.StackName,
						Status: *stack.StackStatus,
					})
				}
			}
		}

		if response.NextToken == nil {
			break
		}
		nextToken = response.NextToken
	}

	return stackList, nil
}

// GetStacksByParameter returns stacks where the specified parameter matches the given value
func GetStacksByParameter(paramName, paramValue, profile, region string) ([]StackInfo, error) {
	stackList := []StackInfo{}

	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	svc := cloudformation.New(sess)

	var nextToken *string
	for {
		response, err := svc.DescribeStacks(&cloudformation.DescribeStacksInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe stacks for profile %s in region %s: %w", profile, region, err)
		}

		for _, stack := range response.Stacks {
			if stack == nil {
				continue
			}

			// Skip deleted stacks
			if stack.StackStatus != nil && *stack.StackStatus == cloudformation.StackStatusDeleteComplete {
				continue
			}

			if hasParameterValue(stack.Parameters, paramName, paramValue) {
				stackList = append(stackList, StackInfo{
					Name:   *stack.StackName,
					Status: *stack.StackStatus,
				})
			}
		}

		if response.NextToken == nil {
			break
		}
		nextToken = response.NextToken
	}

	return stackList, nil
}

func hasParameterValue(params []*cloudformation.Parameter, key, value string) bool {
	for _, p := range params {
		if p != nil && p.ParameterKey != nil && *p.ParameterKey == key {
			return p.ParameterValue != nil && *p.ParameterValue == value
		}
	}
	return false
}
