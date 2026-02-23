package cmd

import (
	"context"
	"log"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

var irsaCmd = &cobra.Command{
	Use:   "irsa",
	Short: "List service accounts with IRSA annotations and their IAM roles",
	Long: `List service accounts with IRSA (IAM Roles for Service Accounts) annotations.

Shows the service account name, namespace, and associated IAM role ARN
from the eks.amazonaws.com/role-arn annotation.`,
	Example: `  # List all service accounts with IRSA
  kubectl eks irsa

  # List IRSA in specific namespace
  kubectl eks irsa -n kube-system

  # List IRSA across all namespaces
  kubectl eks irsa -A`,
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

		serviceAccounts, err := k8s.GetServiceAccountsWithIRSA(context.Background(), namespace)
		if err != nil {
			log.Fatalf("Error getting service accounts: %v", err)
		}

		if len(serviceAccounts) == 0 {
			if namespace == "" {
				log.Println("No service accounts with IRSA annotations found")
			} else {
				log.Printf("No service accounts with IRSA annotations found in namespace: %s\n", namespace)
			}
			return
		}

		irsaInfos := make([]data.IRSAInfo, 0)
		for _, sa := range serviceAccounts {
			roleArn := sa.Annotations["eks.amazonaws.com/role-arn"]
			if roleArn != "" {
				info := data.IRSAInfo{
					Profile:            clusterInfo.AWSProfile,
					Region:             clusterInfo.Region,
					ClusterName:        clusterInfo.ClusterName,
					Namespace:          sa.Namespace,
					ServiceAccountName: sa.Name,
					IAMRoleARN:         roleArn,
				}
				irsaInfos = append(irsaInfos, info)
			}
		}

		if len(irsaInfos) == 0 {
			log.Println("No service accounts with IRSA annotations found")
			return
		}

		printutils.PrintIRSA(noHeaders, irsaInfos...)
	},
}

func init() {
	irsaCmd.Flags().StringP("namespace", "n", "", "Namespace to show IRSA for")
	irsaCmd.Flags().BoolP("all-namespaces", "A", false, "Show IRSA across all namespaces (default)")
	rootCmd.AddCommand(irsaCmd)
}
