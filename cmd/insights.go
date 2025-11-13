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

var insightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "Get insights about an EKS cluster",
	Long:  `Get insights about an EKS cluster (or the current one if not specified)`,
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

		showID, err := cmd.Flags().GetString("show")
		if err != nil {
			showID = ""
		} else {
			showID = strings.TrimSpace(showID)
		}

		if showID == "" {
			insightsList, err := eks.GetEKSInsights(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName)

			if err != nil {
				fmt.Printf("Error getting insights: %s\n", err.Error())
				return
			}

			noHeaders, err := cmd.Flags().GetBool("no-headers")
			if err != nil {
				noHeaders = false
			}

			printutils.PrintInsights(noHeaders, insightsList...)
		} else {
			insightItem, err := eks.DescribeEKSInsight(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, showID)

			if err != nil {
				fmt.Printf("Error getting insight: %s\n", err.Error())
				return
			}

			// fmt.Printf("ID: %s\n", insightItem.ID)
			fmt.Printf("Category: %s\n", insightItem.Category)
			fmt.Printf("Status: %s\n", insightItem.Status)
			fmt.Printf("Description: %s\n", insightItem.Description)
			fmt.Printf("Recommendation: %s\n", insightItem.Recommendation)
			if insightItem.AdditionalInfo != nil {
				if len(*insightItem.AdditionalInfo) > 0 {
					fmt.Printf("Additional Info:\n")
					for key, value := range *insightItem.AdditionalInfo {
						if value != nil {
							fmt.Printf("  * %s:\n      %s\n", key, *value)
						}
					}
				}
			}
			if len(insightItem.Summary.DeprecationDetails) > 0 {
				fmt.Printf("Deprecation Details:\n")
				for _, deprecation := range insightItem.Summary.DeprecationDetails {
					fmt.Printf("  * %q replaced with %q\n", deprecation.Usage, deprecation.ReplacedWith)
					fmt.Printf("    - Replacement from %s to %s\n", deprecation.StartServingReplacementVersion, deprecation.StopServingVersion)
					if len(deprecation.ClientStats) > 0 {
						fmt.Printf("    - Client Stats:\n")
						for _, clientStat := range deprecation.ClientStats {
							fmt.Printf("      * %s has requested %d in the last 30 days - last requested: %s\n", clientStat.UserAgent, clientStat.NumberOfRequestsLast30Days, clientStat.LastRequestTime.String())
						}
					}
				}
			} else {
				fmt.Printf("No deprecation details found\n")
			}
		}
	},
}

func init() {
	insightsCmd.Flags().String("show", "", "Show details for a specific ID")

	rootCmd.AddCommand(insightsCmd)
}
