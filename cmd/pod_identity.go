package cmd

import (
	"log"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

var podIdentityCmd = &cobra.Command{
	Use:     "pod-identity",
	Aliases: []string{"pi"},
	Short:   "List EKS Pod Identity associations from the AWS EKS API",
	Long: `List EKS Pod Identity associations configured via the AWS EKS API.

This command queries the AWS EKS API to show true EKS Pod Identity
associations. These are different from IRSA (IAM Roles for Service Accounts).

EKS Pod Identity is a newer AWS feature that eliminates the need for OIDC providers.`,
	Example: `  # List all Pod Identity associations
  kubectl eks pod-identity

  # List Pod Identity in specific namespace
  kubectl eks pod-identity -n kube-system

  # List Pod Identity across all namespaces
  kubectl eks pod-identity -A`,
	Run: func(cmd *cobra.Command, args []string) {
		clusterInfo, err := GetCurrentClusterInfo()
		if err != nil {
			log.Fatalf("Error getting current cluster info: %v", err)
		}

		namespace, _ := cmd.Flags().GetString("namespace")
		allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")
		noHeaders, _ := cmd.Flags().GetBool("no-headers")

		// Default to all namespaces
		if !allNamespaces && namespace == "" {
			allNamespaces = true
		}

		if allNamespaces {
			namespace = ""
		}

		// Get pod identity associations from EKS API
		associations, err := eks.GetPodIdentityAssociations(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName)
		if err != nil {
			log.Fatalf("Error getting pod identity associations: %v", err)
		}

		// Filter by namespace if specified
		podIdentityInfos := []data.PodIdentityInfo{}
		for _, assoc := range associations {
			if namespace != "" && assoc.Namespace != namespace {
				continue
			}
			podIdentityInfos = append(podIdentityInfos, assoc)
		}

		if len(podIdentityInfos) == 0 {
			if namespace == "" {
				log.Println("No EKS Pod Identity associations found")
			} else {
				log.Printf("No EKS Pod Identity associations found in namespace: %s\n", namespace)
			}
			return
		}

		printutils.PrintPodIdentity(noHeaders, podIdentityInfos...)
	},
}

func init() {
	podIdentityCmd.Flags().StringP("namespace", "n", "", "Namespace to show Pod Identity for")
	podIdentityCmd.Flags().BoolP("all-namespaces", "A", false, "Show Pod Identity across all namespaces (default)")
	rootCmd.AddCommand(podIdentityCmd)
}
