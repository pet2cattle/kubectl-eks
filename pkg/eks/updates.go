package eks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

type EKSUpdateInfo struct {
	Type   string
	Status string
	Errors []string
}

func GetEKSUpdates(profile, region, clusterName string) ([]EKSUpdateInfo, error) {
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

	updatesOut, err := eksSvc.ListUpdates(&eks.ListUpdatesInput{
		Name: aws.String(clusterName),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list updates for cluster %s in region %s: %w", clusterName, region, err)
	}

	updatesList := []EKSUpdateInfo{}

	for _, updateID := range updatesOut.UpdateIds {
		if updateID != nil {
			updateDesc, err := eksSvc.DescribeUpdate(&eks.DescribeUpdateInput{
				Name:     aws.String(clusterName),
				UpdateId: updateID,
			})
			if err != nil {
				continue
			}

			newUpdate := EKSUpdateInfo{
				Type:   *updateDesc.Update.Type,
				Status: *updateDesc.Update.Status,
			}

			if len(updateDesc.Update.Errors) > 0 {
				for _, err := range updateDesc.Update.Errors {
					if err != nil {
						if err.ErrorMessage != nil {
							newUpdate.Errors = append(newUpdate.Errors, *err.ErrorMessage)
						}
					}
				}
			}

			if len(newUpdate.Status) == len(newUpdate.Type) && len(newUpdate.Status) == 0 {
				continue
			}

			updatesList = append(updatesList, newUpdate)
		}
	}

	return updatesList, nil
}
