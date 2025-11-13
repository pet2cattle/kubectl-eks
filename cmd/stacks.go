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

		clusterArn := ""

		if len(args) != 1 {
			// Load Kubernetes configuration
			config, err := KubernetesConfigFlags.ToRawKubeConfigLoader().RawConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading kubeconfig: %v\n", err.Error())
				os.Exit(1)
			}

			// Get current context
			currentContext := config.CurrentContext
			// fmt.Printf("Current context: %s\n", currentContext)

			// Retrieve cluster information
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

		// validate cached data, if invalid, refresh
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

		// get CF stacks based on cluster name
		stackList, err := cf.GetStacks(searchName, clusterInfo.AWSProfile, clusterInfo.Region)
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
	//TODO: filter by status
	rootCmd.AddCommand(stacksCmd)
}
