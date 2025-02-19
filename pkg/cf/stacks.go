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

	// Create a new session using the profile and region
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Create an CF client
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
