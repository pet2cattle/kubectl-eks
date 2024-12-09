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

var updatesCmd = &cobra.Command{
	Use:   "updates",
	Short: "Check for updates",
	Long:  `Check for updates for a given EKS cluster (or the current one if not specified)`,
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

		updateList, err := eks.GetEKSUpdates(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName)

		if err != nil {
			fmt.Printf("Error listing updates: %s\n", err.Error())
			return
		}

		PrintUpdates(updateList...)

	},
}

func PrintUpdates(updateList ...eks.EKSUpdateInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(updateList, func(i, j int) bool {
		return updateList[i].Type < updateList[j].Type
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "TYPE", Type: "string"},
			{Name: "STATUS", Type: "string"},
		},
	}

	// Populate rows with data from the variadic ClusterInfo
	for _, eachUpdate := range updateList {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				eachUpdate.Type,
				eachUpdate.Status,
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
	rootCmd.AddCommand(updatesCmd)
}
