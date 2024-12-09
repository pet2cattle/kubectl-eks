package eks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
)

type EKSNodeGroupInfo struct {
	Name            string
	CapacityType    string
	ReleaseVersion  string
	LaunchTemplate  string
	InstanceType    string
	DesiredCapacity int64
	MaxCapacity     int64
	MinCapacity     int64
	Version         string
	Status          string
}

func GetEKSNodeGroups(profile, region, clusterName string) ([]EKSNodeGroupInfo, error) {
	// Create a new session using the profile and region
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Create an clients
	eksSvc := eks.New(sess)
	ec2Svc := ec2.New(sess)

	input := &eks.ListNodegroupsInput{
		ClusterName: aws.String(clusterName),
	}

	// Get NG names
	result, err := eksSvc.ListNodegroups(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list node groups for cluster %s in region %s: %w", clusterName, region, err)
	}

	ngList := make([]EKSNodeGroupInfo, len(result.Nodegroups))

	for i, ng := range result.Nodegroups {
		if ng != nil {
			// describe NodeGroup
			ngDesc, err := eksSvc.DescribeNodegroup(&eks.DescribeNodegroupInput{
				ClusterName:   aws.String(clusterName),
				NodegroupName: ng,
			})

			if err != nil {
				continue
			}

			if ngDesc.Nodegroup == nil {
				continue
			}

			instanceType := ""

			ltDesc, err := ec2Svc.DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
				LaunchTemplateId: ngDesc.Nodegroup.LaunchTemplate.Id,
				Versions:         []*string{ngDesc.Nodegroup.LaunchTemplate.Version},
			})

			if err == nil && len(ltDesc.LaunchTemplateVersions) > 0 {
				instanceType = *ltDesc.LaunchTemplateVersions[0].LaunchTemplateData.InstanceType
			}

			ngList[i] = EKSNodeGroupInfo{
				Name:            *ng,
				CapacityType:    *ngDesc.Nodegroup.CapacityType,
				ReleaseVersion:  *ngDesc.Nodegroup.ReleaseVersion,
				InstanceType:    instanceType,
				LaunchTemplate:  *ngDesc.Nodegroup.LaunchTemplate.Id,
				DesiredCapacity: *ngDesc.Nodegroup.ScalingConfig.DesiredSize,
				MaxCapacity:     *ngDesc.Nodegroup.ScalingConfig.MaxSize,
				MinCapacity:     *ngDesc.Nodegroup.ScalingConfig.MinSize,
				Version:         *ngDesc.Nodegroup.Version,
				Status:          *ngDesc.Nodegroup.Status,
			}
		}
	}

	return ngList, nil
}
