package eks

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

type ClientStat struct {
	LastRequestTime            time.Time `json:"lastRequestTime"`
	NumberOfRequestsLast30Days int64     `json:"numberOfRequestsLast30Days"`
	UserAgent                  string    `json:"userAgent"`
}

type DeprecationDetail struct {
	ClientStats                    []ClientStat `json:"clientStats"`
	ReplacedWith                   string       `json:"replacedWith"`
	StartServingReplacementVersion string       `json:"startServingReplacementVersion"`
	StopServingVersion             string       `json:"stopServingVersion"`
	Usage                          string       `json:"usage"`
}

type CategorySpecificSummary struct {
	DeprecationDetails []DeprecationDetail `json:"deprecationDetails"`
}

type EKSInsightInfo struct {
	ID             string
	Description    string
	Category       string
	Status         string
	Recommendation string
	Reason         string
	Summary        *CategorySpecificSummary
	AdditionalInfo *map[string]*string
}

func DescribeEKSInsight(profile, region, clusterName, insightID string) (*EKSInsightInfo, error) {
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

	descInsight, err := eksSvc.DescribeInsight(&eks.DescribeInsightInput{
		ClusterName: aws.String(clusterName),
		Id:          aws.String(insightID),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to describe insight %s for cluster %s in region %s: %w", insightID, clusterName, region, err)
	}

	if descInsight.Insight == nil {
		return nil, fmt.Errorf("insight %s not found for cluster %s in region %s", insightID, clusterName, region)
	}

	var summary CategorySpecificSummary

	for _, eachDeprecationDetail := range descInsight.Insight.CategorySpecificSummary.DeprecationDetails {
		if eachDeprecationDetail == nil {
			continue
		}
		var deprecationDetail DeprecationDetail
		deprecationDetail.ReplacedWith = aws.StringValue(eachDeprecationDetail.ReplacedWith)
		deprecationDetail.StartServingReplacementVersion = aws.StringValue(eachDeprecationDetail.StartServingReplacementVersion)
		deprecationDetail.StopServingVersion = aws.StringValue(eachDeprecationDetail.StopServingVersion)
		deprecationDetail.Usage = aws.StringValue(eachDeprecationDetail.Usage)

		for _, stat := range eachDeprecationDetail.ClientStats {
			if stat == nil {
				continue
			}
			clientStat := ClientStat{}
			if stat.LastRequestTime != nil {
				clientStat.LastRequestTime = *stat.LastRequestTime
			}
			if stat.NumberOfRequestsLast30Days != nil {
				clientStat.NumberOfRequestsLast30Days = *stat.NumberOfRequestsLast30Days
			}
			if stat.UserAgent != nil {
				clientStat.UserAgent = *stat.UserAgent
			}
			deprecationDetail.ClientStats = append(deprecationDetail.ClientStats, clientStat)
		} // end for each ClientStats

		summary.DeprecationDetails = append(summary.DeprecationDetails, deprecationDetail)
	}

	return &EKSInsightInfo{
		ID:             *descInsight.Insight.Id,
		Description:    *descInsight.Insight.Description,
		Category:       *descInsight.Insight.Category,
		Status:         *descInsight.Insight.InsightStatus.Status,
		Reason:         *descInsight.Insight.InsightStatus.Reason,
		Summary:        &summary,
		Recommendation: *descInsight.Insight.Recommendation,
		AdditionalInfo: &descInsight.Insight.AdditionalInfo,
	}, nil
}

func GetEKSInsights(profile, region, clusterName string) ([]EKSInsightInfo, error) {
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

	insightsList, err := eksSvc.ListInsights(&eks.ListInsightsInput{
		ClusterName: aws.String(clusterName),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list insights for cluster %s in region %s: %w", clusterName, region, err)
	}

	insightsInfoList := []EKSInsightInfo{}

	for _, insight := range insightsList.Insights {
		if insight == nil {
			continue
		}

		if insight.Id != nil && insight.InsightStatus.Reason != nil && insight.InsightStatus.Status != nil {
			insightsInfoList = append(insightsInfoList, EKSInsightInfo{
				ID:       *insight.Id,
				Category: *insight.Category,
				Status:   *insight.InsightStatus.Status,
				Reason:   *insight.InsightStatus.Reason,
				Summary:  nil,
			})
		}
	}

	// TODO
	return insightsInfoList, nil
}
