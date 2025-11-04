package cmd

import (
	"log"

	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var multiGetPodsCmd = &cobra.Command{
	Use:   "mget-pods",
	Short: "Get a list of pods from multiple clusters",
	Long:  `Get a list of pods from the current cluster or all the clusters that match a filter`,
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

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			namespace = ""
		}

		allNamespaces, err := cmd.Flags().GetBool("all-namespaces")
		if err != nil {
			allNamespaces = false
		}

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err := loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		previousContext := config.CurrentContext

		// current k8s context
		k8SClusterPodList := []k8s.K8SClusterPodList{}
		for _, clusterInfo := range clusterList {
			err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, "")
			if err != nil {
				continue
			}

			k8sPodList, err := k8s.GetPods(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, clusterInfo.Arn, clusterInfo.Version, namespace, allNamespaces)
			if err != nil {
				continue
			} else {
				k8SClusterPodList = append(k8SClusterPodList, *k8sPodList)
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

		PrintMultiGetPods(noHeaders, k8SClusterPodList...)

		saveCacheToDisk()
	},
}

func init() {
	multiGetPodsCmd.Flags().StringP("profile", "p", "", "AWS profile to use")
	multiGetPodsCmd.Flags().StringP("profile-contains", "q", "", "AWS profile contains string")
	multiGetPodsCmd.Flags().StringP("name-contains", "c", "", "Cluster name contains string")
	multiGetPodsCmd.Flags().StringP("name-not-contains", "x", "", "Cluster name does not contain string")
	multiGetPodsCmd.Flags().StringP("region", "r", "", "AWS region to use")
	multiGetPodsCmd.Flags().StringP("version", "v", "", "Filter by EKS version")
	multiGetPodsCmd.Flags().StringP("namespace", "n", "", "Filter by Kubernetes namespace")
	multiGetPodsCmd.Flags().BoolP("all-namespaces", "A", false, "Filter by all Kubernetes namespaces")

	rootCmd.AddCommand(multiGetPodsCmd)
}
