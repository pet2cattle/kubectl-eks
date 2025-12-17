package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/cf"
	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

var stacksCmd = &cobra.Command{
	Use:   "stacks",
	Short: "Get CF stacks",
	Long:  `Get list of possible CF stacks (or the current one if not specified)`,
	Run: func(cmd *cobra.Command, args []string) {
		searchName, err := cmd.Flags().GetString("name")
		if err != nil {
			searchName = ""
		}

		paramFilter, err := cmd.Flags().GetBool("by-parameter")
		if err != nil {
			paramFilter = false
		}

		clusterArn := ""

		if len(args) != 1 {
			config, err := KubernetesConfigFlags.ToRawKubeConfigLoader().RawConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading kubeconfig: %v\n", err.Error())
				os.Exit(1)
			}

			currentContext := config.CurrentContext

			contextDetails, exists := config.Contexts[currentContext]
			if !exists {
				fmt.Fprintf(os.Stderr, "Context '%s' not found in kubeconfig\n", currentContext)
				os.Exit(1)
			}

			clusterArn = contextDetails.Cluster
		} else {
			clusterArn = strings.TrimSpace(args[0])
		}

		loadCacheFromDisk()
		if CachedData == nil {
			CachedData = &data.KubeCtlEksCache{
				ClusterByARN: make(map[string]data.ClusterInfo),
				ClusterList:  make(map[string]map[string][]data.ClusterInfo),
			}
		}

		clusterInfo, exists := CachedData.ClusterByARN[clusterArn]
		if !exists {
			foundClusterInfo := loadClusterByArn(clusterArn)

			if foundClusterInfo == nil {
				fmt.Println("Current cluster is not an EKS cluster")
				os.Exit(1)
			} else {
				clusterInfo = *foundClusterInfo
			}
		}

		if clusterInfo.Arn != clusterArn {
			CachedData = &data.KubeCtlEksCache{
				ClusterByARN: make(map[string]data.ClusterInfo),
				ClusterList:  make(map[string]map[string][]data.ClusterInfo),
			}
			foundClusterInfo := loadClusterByArn(clusterArn)
			if foundClusterInfo == nil {
				fmt.Println("Current cluster is not an EKS cluster")
				os.Exit(1)
			} else {
				clusterInfo = *foundClusterInfo
			}
		}

		if searchName == "" {
			searchName = clusterInfo.ClusterName
		}

		var stackList []cf.StackInfo
		if paramFilter {
			// Filter by ClusterName parameter value
			stackList, err = cf.GetStacksByParameter("ClusterName", searchName, clusterInfo.AWSProfile, clusterInfo.Region)
		} else {
			// Original name-based search
			stackList, err = cf.GetStacks(searchName, clusterInfo.AWSProfile, clusterInfo.Region)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting CF stacks: %v\n", err.Error())
			os.Exit(1)
		}

		noHeaders, err := cmd.Flags().GetBool("no-headers")
		if err != nil {
			noHeaders = false
		}

		printutils.PrintStacks(noHeaders, stackList...)
	},
}

func init() {
	stacksCmd.Flags().String("name", "", "Search for a specific stack name")
	stacksCmd.Flags().BoolP("by-parameter", "p", false, "Filter stacks by ClusterName parameter instead of stack name")
	rootCmd.AddCommand(stacksCmd)
}
