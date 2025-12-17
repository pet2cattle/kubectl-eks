package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

var fargateProfilesCmd = &cobra.Command{
	Use:     "fargate-profiles",
	Aliases: []string{"fp", "fargate"},
	Short:   "List EKS Fargate profiles",
	Long:    `List EKS Fargate profiles for a given cluster (or the current one if not specified)`,
	Run: func(cmd *cobra.Command, args []string) {
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

		arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
		re := regexp.MustCompile(arnRegex)

		matches := re.FindStringSubmatch(clusterArn)
		if matches == nil {
			if len(args) != 1 {
				fmt.Printf("Current cluster is not an EKS cluster\n")
			} else {
				fmt.Printf("Invalid cluster ARN: %q\n", clusterArn)
			}
			os.Exit(1)
		}

		clusterInfo := loadClusterByArn(clusterArn)

		if clusterInfo == nil {
			fmt.Println("Cluster not found")
			return
		}

		noHeaders, err := cmd.Flags().GetBool("no-headers")
		if err != nil {
			noHeaders = false
		}

		profileList, err := eks.GetEKSFargateProfiles(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName)
		if err != nil {
			fmt.Printf("Error listing Fargate profiles: %s\n", err.Error())
			os.Exit(1)
		}

		printutils.PrintFargateProfiles(noHeaders, profileList...)
	},
}

func init() {
	rootCmd.AddCommand(fargateProfilesCmd)
}
