package cmd

import (
	"context"
	"log"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

var quotasCmd = &cobra.Command{
	Use:   "quotas",
	Short: "Show resource quotas and current utilization per namespace",
	Long: `Show resource quotas and current utilization across namespaces.

Displays resource quotas (CPU, memory, pods, etc.) along with current
usage to help identify quota limits and available capacity.`,
	Example: `  # Show quotas for current cluster
  kubectl eks quotas

  # Show quotas for specific namespace
  kubectl eks quotas -n production

  # Show quotas across all namespaces
  kubectl eks quotas -A`,
	Run: func(cmd *cobra.Command, args []string) {
		clusterInfo, err := GetCurrentClusterInfo()
		if err != nil {
			log.Fatalf("Error getting current cluster info: %v", err)
		}

		namespace, _ := cmd.Flags().GetString("namespace")
		allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")
		noHeaders, _ := cmd.Flags().GetBool("no-headers")

		if !allNamespaces && namespace == "" {
			currentNs, err := k8s.GetCurrentNamespace()
			if err != nil {
				namespace = "default"
			} else {
				namespace = currentNs
			}
		}

		if allNamespaces {
			namespace = ""
		}

		quotas, err := k8s.GetResourceQuotas(context.Background(), namespace)
		if err != nil {
			log.Fatalf("Error getting resource quotas: %v", err)
		}

		if len(quotas) == 0 {
			if namespace == "" {
				log.Println("No resource quotas found in any namespace")
			} else {
				log.Printf("No resource quotas found in namespace: %s\n", namespace)
			}
			return
		}

		quotaInfos := make([]data.ResourceQuotaInfo, 0)
		for _, quota := range quotas {
			for resourceName, hardLimit := range quota.Status.Hard {
				used := quota.Status.Used[resourceName]
				info := data.ResourceQuotaInfo{
					Profile:      clusterInfo.AWSProfile,
					Region:       clusterInfo.Region,
					ClusterName:  clusterInfo.ClusterName,
					Namespace:    quota.Namespace,
					QuotaName:    quota.Name,
					ResourceName: string(resourceName),
					Hard:         hardLimit.String(),
					Used:         used.String(),
				}
				quotaInfos = append(quotaInfos, info)
			}
		}

		printutils.PrintResourceQuotas(noHeaders, quotaInfos...)
	},
}

func init() {
	quotasCmd.Flags().StringP("namespace", "n", "", "Namespace to show quotas for")
	quotasCmd.Flags().BoolP("all-namespaces", "A", false, "Show quotas across all namespaces")
	rootCmd.AddCommand(quotasCmd)
}
