package cmd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
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

			PrintInsights(noHeaders, insightsList...)
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

func PrintInsights(noHeaders bool, insights ...eks.EKSInsightInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(insights, func(i, j int) bool {
		return insights[i].ID < insights[j].ID
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "ID", Type: "string"},
			{Name: "CATEGORY", Type: "string"},
			{Name: "STATUS", Type: "string"},
			{Name: "REASON", Type: "string"},
		},
	}

	// Populate rows with data from the variadic ClusterInfo
	for _, eachInsight := range insights {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				eachInsight.ID,
				eachInsight.Category,
				eachInsight.Status,
				eachInsight.Reason,
			},
		})
	}

	// Print the table
	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	insightsCmd.Flags().String("show", "", "Show details for a specific ID")

	rootCmd.AddCommand(insightsCmd)
}
