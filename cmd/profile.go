package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "aws-profile",
	Short: "Get AWS profile",
	Long:  `Get the AWS profile name for the current cluster (or specified cluster ARN)`,
	Run: func(cmd *cobra.Command, args []string) {
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
			CachedData = &KubeCtlEksCache{
				ClusterByARN: make(map[string]ClusterInfo),
				ClusterList:  make(map[string]map[string][]ClusterInfo),
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
			CachedData = &KubeCtlEksCache{
				ClusterByARN: make(map[string]ClusterInfo),
				ClusterList:  make(map[string]map[string][]ClusterInfo),
			}
			foundClusterInfo := loadClusterByArn(clusterArn)
			if foundClusterInfo == nil {
				fmt.Println("Current cluster is not an EKS cluster")
				os.Exit(1)
			} else {
				clusterInfo = *foundClusterInfo
			}
		}

		// Print the AWS profile name
		fmt.Println(clusterInfo.AWSProfile)
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
}
