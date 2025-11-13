package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/ec2"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

var nodegroupsCmd = &cobra.Command{
	Use:   "nodegroups",
	Short: "List EKS nodegroups",
	Long:  `List EKS nodegroups for a given cluster (or the current one if not specified)`,
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

		// check if it is an ARN
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

		// clusterInfo := loadClusterByArn(clusterARN)
		if clusterInfo == nil {
			fmt.Println("Cluster not found")
			return
		}

		ami, err := cmd.Flags().GetString("ami")
		if err != nil {
			ami = ""
		}

		noHeaders, err := cmd.Flags().GetBool("no-headers")
		if err != nil {
			noHeaders = false
		}

		if ami != "" {
			amiInfo, err := ec2.GetAMIInfo(clusterInfo.AWSProfile, clusterInfo.Region, ami)
			if err != nil {
				fmt.Printf("Error getting AMI info: %s\n", err.Error())
				os.Exit(1)
			}

			printutils.PrintAMIs(noHeaders, *amiInfo)

		} else {
			clusterNGList, err := eks.GetEKSNodeGroups(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName)

			if err != nil {
				fmt.Printf("Error listing nodegroups: %s\n", err.Error())
				return
			}

			printutils.PrintNodeGroup(noHeaders, clusterNGList...)
		}
	},
}

func init() {
	nodegroupsCmd.Flags().StringP("ami", "a", "", "Describe AMI used by the nodegroup")

	rootCmd.AddCommand(nodegroupsCmd)
}
