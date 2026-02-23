package karpenter

import (
	"fmt"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
)

func GetDriftedResources(profile, region, clusterName string) ([]data.KarpenterDriftInfo, error) {
	nodeClaims, err := GetNodeClaims(profile, region, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get NodeClaims: %w", err)
	}

	var result []data.KarpenterDriftInfo

	for _, nc := range nodeClaims {
		if nc.Drifted {
			info := data.KarpenterDriftInfo{
				Profile:      profile,
				Region:       region,
				ClusterName:  clusterName,
				ResourceType: "NodeClaim",
				Name:         nc.Name,
				NodeName:     nc.NodeName,
				NodePoolName: nc.NodePoolName,
				DriftedSince: nc.Age, // Simplified - would need to check condition timestamp
				Reason:       "Configuration drift detected",
			}
			result = append(result, info)
		}
	}

	return result, nil
}
