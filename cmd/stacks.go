package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/cf"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

var stacksCmd = &cobra.Command{
	Use:   "stacks",
	Short: "Get CF stacks",
	Long:  `Get list of possible CF stacks (or the current one if not specified)`,
	Run: func(cmd *cobra.Command, args []string) {
		searchName, err := cmd.Flags().GetString("name")
		if err != nil {
			searchName = ""
		}

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

		loadCacheFromDisk()
		if CachedData == nil {
			CachedData = &KubeCtlEksCache{
				ClusterByARN: make(map[string]ClusterInfo),
				ClusterList:  make(map[string]map[string][]ClusterInfo),
			}
		}

		clusterInfo, exists := CachedData.ClusterByARN[clusterArn]
		if !exists {
			foundClusterInfo := loadClusterByArn(clusterArn)

			if foundClusterInfo == nil {
				fmt.Println("Current cluster is not an EKS cluster")
				os.Exit(1)
			} else {
				clusterInfo = *foundClusterInfo
			}
		}

		// validate cached data, if invalid, refresh
		if clusterInfo.Arn != clusterArn {
			CachedData = &KubeCtlEksCache{
				ClusterByARN: make(map[string]ClusterInfo),
				ClusterList:  make(map[string]map[string][]ClusterInfo),
			}
			foundClusterInfo := loadClusterByArn(clusterArn)
			if foundClusterInfo == nil {
				fmt.Println("Current cluster is not an EKS cluster")
				os.Exit(1)
			} else {
				clusterInfo = *foundClusterInfo
			}
		}

		if searchName == "" {
			searchName = clusterInfo.ClusterName
		}

		// get CF stacks based on cluster name
		stackList, err := cf.GetStacks(searchName, clusterInfo.AWSProfile, clusterInfo.Region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting CF stacks: %v\n", err.Error())
			os.Exit(1)
		}

		noHeaders, err := cmd.Flags().GetBool("no-headers")
		if err != nil {
			noHeaders = false
		}

		PrintStacks(noHeaders, stackList...)

	},
}

func PrintStacks(noHeaders bool, stackList ...cf.StackInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(stackList, func(i, j int) bool {
		return stackList[i].Name < stackList[j].Name
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			// {Name: "AWS ACCOUNT ID", Type: "string"},
			{Name: "NAME", Type: "string"},
			{Name: "STATUS", Type: "string"},
		},
	}

	// Populate rows with data from the variadic ClusterInfo
	for _, stackInfo := range stackList {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				// clusterInfo.AWSAccountID,
				stackInfo.Name,
				stackInfo.Status,
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
	stacksCmd.Flags().String("name", "", "Search for a specific stack name")
	//TODO: filter by status
	rootCmd.AddCommand(stacksCmd)
}
