package cmd

import (
	"log"

	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get EKS cluster stats",
	Long:  `Get EKS cluster stats of the current cluster or all the clusters that matches a filter`,
	Run: func(cmd *cobra.Command, args []string) {
		profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			profile = ""
		}

		profile_contains, err := cmd.Flags().GetString("profile-contains")
		if err != nil {
			profile_contains = ""
		}

		name_contains, err := cmd.Flags().GetString("name-contains")
		if err != nil {
			name_contains = ""
		}

		name_not_contains, err := cmd.Flags().GetString("name-not-contains")
		if err != nil {
			name_not_contains = ""
		}

		region, err := cmd.Flags().GetString("region")
		if err != nil {
			region = ""
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			version = ""
		}

		clusterList, err := LoadClusterList(args, profile, profile_contains, name_contains, name_not_contains, region, version)
		if err != nil {
			log.Fatalf("Error loading cluster list: %v", err)
		}

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err := loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		previousContext := config.CurrentContext

		// current k8s context
		k8sStatsList := []k8s.K8Sstats{}
		for _, clusterInfo := range clusterList {
			err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, "")
			if err != nil {
				continue
			}

			stats, err := k8s.GetK8sStats(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, clusterInfo.Arn, clusterInfo.Version)
			if err != nil {
				continue
			} else {
				k8sStatsList = append(k8sStatsList, *stats)
			}
		}

		// Restore the previous context
		loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
		config, err = loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		config.CurrentContext = previousContext
		if err := clientcmd.ModifyConfig(loadingRules, *config, true); err != nil {
			log.Fatalf("Error updating kubeconfig: %v", err)
		}

		noHeaders, err := cmd.Flags().GetBool("no-headers")
		if err != nil {
			noHeaders = false
		}

		PrintK8SStats(noHeaders, k8sStatsList...)

		saveCacheToDisk()
	},
}

func init() {
	statsCmd.Flags().StringP("profile", "p", "", "AWS profile to use")
	statsCmd.Flags().StringP("profile-contains", "q", "", "AWS profile contains string")
	statsCmd.Flags().StringP("name-contains", "c", "", "Cluster name contains string")
	statsCmd.Flags().StringP("name-not-contains", "x", "", "Cluster name does not contain string")
	statsCmd.Flags().StringP("region", "r", "", "AWS region to use")
	statsCmd.Flags().StringP("version", "v", "", "Filter by EKS version")

	rootCmd.AddCommand(statsCmd)
}
