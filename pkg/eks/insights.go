package eks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/data"
)

func DescribeEKSInsight(profile, region, clusterName, insightID string) (*data.EKSInsightInfo, error) {
	ctx := context.Background()

	// Load the AWS configuration using the profile and region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create an EKS client
	eksSvc := eks.NewFromConfig(cfg)

	descInsight, err := eksSvc.DescribeInsight(ctx, &eks.DescribeInsightInput{
		ClusterName: aws.String(clusterName),
		Id:          aws.String(insightID),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to describe insight %s for cluster %s in region %s: %w", insightID, clusterName, region, err)
	}

	if descInsight.Insight == nil {
		return nil, fmt.Errorf("insight %s not found for cluster %s in region %s", insightID, clusterName, region)
	}

	var summary data.CategorySpecificSummary

	for _, eachDeprecationDetail := range descInsight.Insight.CategorySpecificSummary.DeprecationDetails {
		var deprecationDetail data.DeprecationDetail
		deprecationDetail.ReplacedWith = aws.ToString(eachDeprecationDetail.ReplacedWith)
		deprecationDetail.StartServingReplacementVersion = aws.ToString(eachDeprecationDetail.StartServingReplacementVersion)
		deprecationDetail.StopServingVersion = aws.ToString(eachDeprecationDetail.StopServingVersion)
		deprecationDetail.Usage = aws.ToString(eachDeprecationDetail.Usage)

		for _, stat := range eachDeprecationDetail.ClientStats {
			clientStat := data.ClientStat{}
			if stat.LastRequestTime != nil {
				clientStat.LastRequestTime = *stat.LastRequestTime
			}
			clientStat.NumberOfRequestsLast30Days = int64(stat.NumberOfRequestsLast30Days)
			if stat.UserAgent != nil {
				clientStat.UserAgent = *stat.UserAgent
			}
			deprecationDetail.ClientStats = append(deprecationDetail.ClientStats, clientStat)
		} // end for each ClientStats

		summary.DeprecationDetails = append(summary.DeprecationDetails, deprecationDetail)
	}

	// Convert map[string]string to map[string]*string for compatibility
	var additionalInfo map[string]*string
	if descInsight.Insight.AdditionalInfo != nil {
		additionalInfo = make(map[string]*string, len(descInsight.Insight.AdditionalInfo))
		for k, v := range descInsight.Insight.AdditionalInfo {
			additionalInfo[k] = aws.String(v)
		}
	}

	return &data.EKSInsightInfo{
		ID:             *descInsight.Insight.Id,
		Description:    *descInsight.Insight.Description,
		Category:       string(descInsight.Insight.Category),
		Status:         string(descInsight.Insight.InsightStatus.Status),
		Reason:         *descInsight.Insight.InsightStatus.Reason,
		Summary:        &summary,
		Recommendation: *descInsight.Insight.Recommendation,
		AdditionalInfo: &additionalInfo,
	}, nil
}

func GetEKSInsights(profile, region, clusterName string) ([]data.EKSInsightInfo, error) {
	ctx := context.Background()

	// Load the AWS configuration using the profile and region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create an EKS client
	eksSvc := eks.NewFromConfig(cfg)

	insightsList, err := eksSvc.ListInsights(ctx, &eks.ListInsightsInput{
		ClusterName: aws.String(clusterName),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list insights for cluster %s in region %s: %w", clusterName, region, err)
	}

	insightsInfoList := []data.EKSInsightInfo{}

	for _, insight := range insightsList.Insights {
		if insight.Id != nil && insight.InsightStatus != nil {
			insightsInfoList = append(insightsInfoList, data.EKSInsightInfo{
				ID:       *insight.Id,
				Category: string(insight.Category),
				Status:   string(insight.InsightStatus.Status),
				Reason:   *insight.InsightStatus.Reason,
				Summary:  nil,
			})
		}
	}

	// TODO
	return insightsInfoList, nil
}
