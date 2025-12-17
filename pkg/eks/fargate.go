package eks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

type FargateProfileInfo struct {
	Name           string
	Status         string
	PodExecRoleArn string
	Subnets        []string
	Selectors      []FargateSelector
}

type FargateSelector struct {
	Namespace string
	Labels    map[string]string
}

func GetEKSFargateProfiles(profile, region, clusterName string) ([]FargateProfileInfo, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	svc := eks.New(sess)

	// List all Fargate profile names
	var profileNames []*string
	var nextToken *string
	for {
		resp, err := svc.ListFargateProfiles(&eks.ListFargateProfilesInput{
			ClusterName: aws.String(clusterName),
			NextToken:   nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list Fargate profiles: %w", err)
		}

		profileNames = append(profileNames, resp.FargateProfileNames...)

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	// Describe each profile
	var profiles []FargateProfileInfo
	for _, name := range profileNames {
		resp, err := svc.DescribeFargateProfile(&eks.DescribeFargateProfileInput{
			ClusterName:        aws.String(clusterName),
			FargateProfileName: name,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe Fargate profile %s: %w", *name, err)
		}

		fp := resp.FargateProfile
		info := FargateProfileInfo{
			Name:           *fp.FargateProfileName,
			Status:         *fp.Status,
			PodExecRoleArn: aws.StringValue(fp.PodExecutionRoleArn),
			Subnets:        aws.StringValueSlice(fp.Subnets),
		}

		for _, sel := range fp.Selectors {
			selector := FargateSelector{
				Namespace: aws.StringValue(sel.Namespace),
				Labels:    aws.StringValueMap(sel.Labels),
			}
			info.Selectors = append(info.Selectors, selector)
		}

		profiles = append(profiles, info)
	}

	return profiles, nil
}
