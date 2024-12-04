package cmd

import (
	"fmt"
	"os"

	"github.com/pet2cattle/kubectl-eks/pkg/awsconfig"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/sts"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List EKS clusters",
	Long:  `List EKS clusters filtering by AWS profile and region`,
	Run: func(cmd *cobra.Command, args []string) {
		refresh, err := cmd.Flags().GetBool("refresh")
		if err != nil {
			refresh = false
		}

		profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			profile = ""
		}

		region, err := cmd.Flags().GetString("region")
		if err != nil {
			region = ""
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			version = ""
		}

		loadCacheFromDisk()
		if CachedData == nil {
			CachedData = &KubeCtlEksCache{
				ClusterByARN: make(map[string]ClusterInfo),
				ClusterList:  make(map[string]map[string][]ClusterInfo),
			}
		}

		if refresh {
			CachedData.ClusterList = make(map[string]map[string][]ClusterInfo)
		}

		clusterList := []ClusterInfo{}

		awsProfiles := awsconfig.GetAWSProfilesWithEKSHints()
		for _, profileDetails := range awsProfiles {
			if profile != "" && profile != profileDetails.Name {
				continue
			}
			for _, hintRegion := range profileDetails.HintEKSRegions {
				if region != "" && region != hintRegion {
					continue
				}

				if refresh {
					_, exists := CachedData.ClusterList[profileDetails.Name]
					if !exists {
						CachedData.ClusterList[profileDetails.Name] = make(map[string][]ClusterInfo)
					}
					_, exists = CachedData.ClusterList[profileDetails.Name][hintRegion]
					if !exists {
						CachedData.ClusterList[profileDetails.Name][hintRegion] = []ClusterInfo{}
					}
					loadClusters(profileDetails.Name, hintRegion)
				} else {
					cachedRegions, exists := CachedData.ClusterList[profileDetails.Name]
					if !exists {
						loadClusters(profileDetails.Name, hintRegion)
					} else {
						_, exists := cachedRegions[hintRegion]
						if !exists {
							loadClusters(profileDetails.Name, hintRegion)
						}
					}
				}

				currentClusterList, exists := CachedData.ClusterList[profileDetails.Name][hintRegion]
				if !exists {
					fmt.Fprintf(os.Stderr, "Unable to load clusters using profile: %s region: %s\n", profileDetails.Name, hintRegion)
				} else {
					if version == "" {
						clusterList = append(clusterList, currentClusterList...)
					} else {
						for _, cluster := range currentClusterList {
							if cluster.Version == version {
								clusterList = append(clusterList, cluster)
							}
						}
					}
				}

			}
		}

		PrintClusters(clusterList...)

		saveCacheToDisk()
	},
}

func loadClusters(profile, region string) {
	// fmt.Printf("Loading clusters using profile: %s region: %s\n", profile, region)

	// Get the list of clusters
	clusters, err := eks.GetClusters(profile, region)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing clusters: %v\n", err.Error())
		os.Exit(1)
	}

	accountID, err := sts.GetAccountID(profile, region)
	if err != nil {
		accountID = "-"
	}

	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}

		clusterData := ClusterInfo{
			ClusterName:  *cluster,
			Region:       region,
			AWSProfile:   profile,
			AWSAccountID: accountID,
		}

		clusterInfo, err := eks.DescribeCluster(profile, region, *cluster)
		if err != nil || clusterInfo == nil {
			fmt.Fprintf(os.Stderr, "Error describing cluster %s: %v\n", *cluster, err.Error())
		} else {
			clusterData.Status = *clusterInfo.Status
			clusterData.Version = *clusterInfo.Version
			clusterData.Arn = *clusterInfo.Arn
		}

		// CachedData.ClusterInfo[clusterName] = clusterInfo

		_, exists := CachedData.ClusterList[profile]
		if !exists {
			CachedData.ClusterList[profile] = make(map[string][]ClusterInfo)
		}

		_, exists = CachedData.ClusterList[profile][region]
		if !exists {
			CachedData.ClusterList[profile][region] = []ClusterInfo{}
		}

		// fmt.Printf("Adding cluster %s to profile %s and region %s\n", clusterData.ClusterName, profile, region)
		CachedData.ClusterList[profile][region] = append(CachedData.ClusterList[profile][region], clusterData)
	}

}

func init() {
	listCmd.Flags().BoolP("refresh", "u", false, "Refresh data from AWS")
	listCmd.Flags().StringP("profile", "p", "", "AWS profile to use")
	listCmd.Flags().StringP("region", "r", "", "AWS region to use")
	listCmd.Flags().StringP("version", "v", "", "Filter by EKS version")

	rootCmd.AddCommand(listCmd)
}
