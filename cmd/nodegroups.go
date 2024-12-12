package cmd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/ec2"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
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

		if ami != "" {
			amiInfo, err := ec2.GetAMIInfo(clusterInfo.AWSProfile, clusterInfo.Region, ami)
			if err != nil {
				fmt.Printf("Error getting AMI info: %s\n", err.Error())
				os.Exit(1)
			}

			PrintAMIs(*amiInfo)

		} else {
			clusterNGList, err := eks.GetEKSNodeGroups(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName)

			if err != nil {
				fmt.Printf("Error listing nodegroups: %s\n", err.Error())
				return
			}

			PrintNodeGroup(clusterNGList...)
		}
	},
}

// printResults prints results in a kubectl-style table format
func PrintAMIs(amiInfos ...ec2.AMIInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(amiInfos, func(i, j int) bool {
		return amiInfos[i].Name < amiInfos[j].Name
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "NAME", Type: "string"},
			{Name: "ARCHITECTURE", Type: "string"},
			{Name: "STATE", Type: "string"},
			{Name: "DEPRECATION TIME", Type: "string"},
		},
	}

	// Populate rows with data from the variadic ClusterInfo
	for _, amiInfo := range amiInfos {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				amiInfo.Name,
				amiInfo.Architecture,
				amiInfo.State,
				amiInfo.DeprecationTime,
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

func PrintNodeGroup(ngInfo ...eks.EKSNodeGroupInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(ngInfo, func(i, j int) bool {
		return ngInfo[i].Name < ngInfo[j].Name
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "NAME", Type: "string"},
			{Name: "CAPACITY TYPE", Type: "string"},
			{Name: "RELEASE VERSION", Type: "string"},
			{Name: "LAUNCH TEMPLATE", Type: "string"},
			{Name: "INSTANCE TYPE", Type: "string"},
			{Name: "DESIRED CAPACITY", Type: "string"},
			{Name: "MAX CAPACITY", Type: "string"},
			{Name: "MIN CAPACITY", Type: "string"},
			{Name: "VERSION", Type: "string"},
			{Name: "STATUS", Type: "string"},
		},
	}

	// Populate rows with data from the variadic ClusterInfo
	for _, eachNG := range ngInfo {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				eachNG.Name,
				eachNG.CapacityType,
				eachNG.ReleaseVersion,
				eachNG.LaunchTemplate,
				eachNG.InstanceType,
				eachNG.DesiredCapacity,
				eachNG.MaxCapacity,
				eachNG.MinCapacity,
				eachNG.Version,
				eachNG.Status,
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
	nodegroupsCmd.Flags().StringP("ami", "a", "", "Describe AMI used by the nodegroup")
	rootCmd.AddCommand(nodegroupsCmd)
}
