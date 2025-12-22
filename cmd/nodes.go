package cmd

import (
	"fmt"
	"log"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List EKS nodes",
	Long: `List EKS nodes for one or more clusters.

When cluster filters are provided, queries multiple clusters.
Without filters, queries the current cluster context.`,
	Example: `  # List nodes for current cluster
  kubectl eks nodes

  # List nodes across clusters matching filter
  kubectl eks nodes --name-contains prod

  # List nodes for specific profile
  kubectl eks nodes --profile my-aws-profile

  # List nodes across all clusters in a region
  kubectl eks nodes --region us-west-2`,
	Run: func(cmd *cobra.Command, args []string) {
		noHeaders, _ := cmd.Flags().GetBool("no-headers")

		// Get filter flags
		profile, _ := cmd.Flags().GetString("profile")
		profileContains, _ := cmd.Flags().GetString("profile-contains")
		nameContains, _ := cmd.Flags().GetString("name-contains")
		nameNotContains, _ := cmd.Flags().GetString("name-not-contains")
		region, _ := cmd.Flags().GetString("region")
		version, _ := cmd.Flags().GetString("version")

		// Check if any filter is specified
		hasFilters := profile != "" || profileContains != "" || nameContains != "" ||
			nameNotContains != "" || region != "" || version != ""

		var clusterList []data.ClusterInfo

		if hasFilters {
			// Ensure cache is initialized before LoadClusterList
			loadCacheFromDisk()
			if CachedData == nil {
				CachedData = &data.KubeCtlEksCache{
					ClusterByARN: make(map[string]data.ClusterInfo),
					ClusterList:  make(map[string]map[string][]data.ClusterInfo),
				}
			}
			if CachedData.ClusterList == nil {
				CachedData.ClusterList = make(map[string]map[string][]data.ClusterInfo)
			}

			var err error
			clusterList, err = LoadClusterList([]string{}, profile, profileContains, nameContains, nameNotContains, region, version)
			if err != nil {
				log.Fatalf("Error loading cluster list: %v", err)
			}
			runMultiClusterNodes(clusterList, noHeaders, false)
		} else {
			// No filters - use current context directly
			clusterInfo, err := GetCurrentClusterInfo()
			if err != nil {
				log.Fatalf("Error getting current cluster info: %v", err)
			}
			clusterList = []data.ClusterInfo{clusterInfo}
			runMultiClusterNodes(clusterList, noHeaders, true)
		}
	},
}

func runMultiClusterNodes(clusterList []data.ClusterInfo, noHeaders bool, skipContextSwitch bool) {
	if len(clusterList) == 0 {
		fmt.Println("No clusters found matching the specified filters")
		return
	}

	// Save and restore context only if we're switching clusters
	if !skipContextSwitch {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err := loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		previousContext := config.CurrentContext
		defer func() {
			config.CurrentContext = previousContext
			clientcmd.ModifyConfig(loadingRules, *config, true)
		}()
	}

	allNodes := []data.ClusterNodeInfo{}

	for _, clusterInfo := range clusterList {
		if !skipContextSwitch {
			err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, "")
			if err != nil {
				log.Printf("Warning: Failed to update kubeconfig for cluster %s: %v", clusterInfo.ClusterName, err)
				continue
			}
		}

		// Create fresh client config to pick up the switched context (same pattern as mget)
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)

		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			log.Printf("Warning: Failed to get client config for cluster %s: %v", clusterInfo.ClusterName, err)
			continue
		}

		nodeList, err := k8s.GetNodesWithConfig(restConfig)
		if err != nil {
			log.Printf("Warning: Failed to get nodes from cluster %s: %v", clusterInfo.ClusterName, err)
			continue
		}

		for _, node := range nodeList {
			allNodes = append(allNodes, data.ClusterNodeInfo{
				Profile:     clusterInfo.AWSProfile,
				Region:      clusterInfo.Region,
				ClusterName: clusterInfo.ClusterName,
				Node:        node,
			})
		}
	}

	printutils.PrintMultiClusterNodes(noHeaders, allNodes)

	saveCacheToDisk()
}

func init() {
	nodesCmd.Flags().StringP("profile", "p", "", "AWS profile to use")
	nodesCmd.Flags().StringP("profile-contains", "q", "", "AWS profile contains string")
	nodesCmd.Flags().StringP("name-contains", "c", "", "Cluster name contains string")
	nodesCmd.Flags().StringP("name-not-contains", "x", "", "Cluster name does not contain string")
	nodesCmd.Flags().StringP("region", "r", "", "AWS region to use")
	nodesCmd.Flags().StringP("version", "v", "", "Filter by EKS version")

	rootCmd.AddCommand(nodesCmd)
}
