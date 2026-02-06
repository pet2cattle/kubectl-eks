package printutils

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pet2cattle/kubectl-eks/pkg/cf"
	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/printers"
)

func PrintMultiGetPods(noHeaders bool, podList ...k8s.K8SClusterPodList) {
	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "ARN", Type: "string"},
			{Name: "VERSION", Type: "string"},
			{Name: "NAMESPACE", Type: "string"},
			{Name: "POD NAME", Type: "string"},
			{Name: "READY", Type: "string"},
			{Name: "STATUS", Type: "string"},
			{Name: "RESTARTS", Type: "number"},
			{Name: "AGE", Type: "string"},
		},
	}

	// Populate rows with data from the variadic K8Sstats
	for _, clusterList := range podList {
		for _, pod := range clusterList.Pods {
			humanAge := duration.ShortHumanDuration(time.Since(pod.Age.Time))
			table.Rows = append(table.Rows, v1.TableRow{
				Cells: []interface{}{
					clusterList.AWSProfile,
					clusterList.Region,
					clusterList.ClusterName,
					clusterList.Arn,
					clusterList.Version,
					pod.Namespace,
					pod.Name,
					pod.Ready,
					pod.Status,
					pod.Restarts,
					humanAge,
				},
			})
		}
	}

	// Print the table
	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

// print k8s stats in a kubectl-style table format
func PrintK8SStats(noHeaders bool, statsList ...k8s.K8Sstats) {
	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "ARN", Type: "string"},
			{Name: "VERSION", Type: "string"},
			{Name: "NAMESPACES", Type: "number"},
			{Name: "POD COUNT", Type: "number"},
			{Name: "NODE COUNT", Type: "number"},
			{Name: "NODES NOT READY", Type: "number"},
			{Name: "PODS NOT RUNNING", Type: "number"},
			{Name: "PODS WITH RESTARTS", Type: "number"},
		},
	}

	// Populate rows with data from the variadic K8Sstats
	for _, stats := range statsList {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				stats.AWSProfile,
				stats.Region,
				stats.ClusterName,
				stats.Arn,
				stats.Version,
				stats.NamespaceCount,
				stats.PodCount,
				stats.NodeCount,
				stats.NodesNotReady,
				stats.PodsNotRunning,
				stats.PodsWithRestartsCount,
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

// printResults prints results in a kubectl-style table format
func PrintClusters(noHeaders bool, clusterInfos ...data.ClusterInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(clusterInfos, func(i, j int) bool {
		return clusterInfos[i].AWSProfile < clusterInfos[j].AWSProfile
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	// Create a Table object
	var table *v1.Table

	if len(clusterInfos) == 1 && clusterInfos[0].Namespace != "" {
		table = &v1.Table{
			ColumnDefinitions: []v1.TableColumnDefinition{
				// {Name: "AWS ACCOUNT ID", Type: "string"},
				{Name: "AWS PROFILE", Type: "string"},
				{Name: "AWS REGION", Type: "string"},
				{Name: "CLUSTER NAME", Type: "string"},
				{Name: "NAMESPACE", Type: "string"},
				{Name: "STATUS", Type: "string"},
				{Name: "VERSION", Type: "string"},
				{Name: "CREATED", Type: "string"},
				{Name: "ARN", Type: "string"},
			},
		}
	} else {
		table = &v1.Table{
			ColumnDefinitions: []v1.TableColumnDefinition{
				// {Name: "AWS ACCOUNT ID", Type: "string"},
				{Name: "AWS PROFILE", Type: "string"},
				{Name: "AWS REGION", Type: "string"},
				{Name: "CLUSTER NAME", Type: "string"},
				{Name: "STATUS", Type: "string"},
				{Name: "VERSION", Type: "string"},
				{Name: "CREATED", Type: "string"},
				{Name: "ARN", Type: "string"},
			},
		}
	}

	// Populate rows with data from the variadic ClusterInfo
	for _, clusterInfo := range clusterInfos {
		if len(clusterInfos) == 1 && clusterInfo.Namespace != "" {
			table.Rows = append(table.Rows, v1.TableRow{
				Cells: []interface{}{
					// clusterInfo.AWSAccountID,
					clusterInfo.AWSProfile,
					clusterInfo.Region,
					clusterInfo.ClusterName,
					clusterInfo.Namespace,
					clusterInfo.Status,
					clusterInfo.Version,
					clusterInfo.CreatedAt,
					clusterInfo.Arn,
				},
			})
		} else {
			table.Rows = append(table.Rows, v1.TableRow{
				Cells: []interface{}{
					// clusterInfo.AWSAccountID,
					clusterInfo.AWSProfile,
					clusterInfo.Region,
					clusterInfo.ClusterName,
					clusterInfo.Status,
					clusterInfo.Version,
					clusterInfo.CreatedAt,
					clusterInfo.Arn,
				},
			})
		}
	}

	// Print the table
	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

// PrintJsonPathResults prints the results in a kubectl-style table format
func PrintJsonPathResults(noHeaders bool, results []data.JsonPathResult) {
	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "PROFILE", Type: "string"},
			{Name: "REGION", Type: "string"},
			{Name: "CLUSTER", Type: "string"},
			{Name: "NAMESPACE", Type: "string"},
			{Name: "NAME", Type: "string"},
			{Name: "VALUE", Type: "string"},
		},
	}

	// Populate rows with data
	for _, result := range results {
		value := result.Value
		if result.Error != "" {
			value = "ERROR: " + result.Error
		}

		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				result.Profile,
				result.Region,
				result.ClusterName,
				result.Namespace,
				result.Resource,
				value,
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

func PrintInsights(noHeaders bool, insights ...data.EKSInsightInfo) {
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

func PrintUpdates(noHeaders bool, updateList ...eks.EKSUpdateInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(updateList, func(i, j int) bool {
		return updateList[i].Type < updateList[j].Type
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "TYPE", Type: "string"},
			{Name: "STATUS", Type: "string"},
			{Name: "ERRORS", Type: "string"},
		},
	}

	// Populate rows with data from the variadic ClusterInfo
	for _, eachUpdate := range updateList {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				eachUpdate.Type,
				eachUpdate.Status,
				strings.Join(eachUpdate.Errors, ", "),
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

// printResults prints results in a kubectl-style table format
func PrintAMIs(noHeaders bool, amiInfos ...data.AMIInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(amiInfos, func(i, j int) bool {
		return amiInfos[i].Name < amiInfos[j].Name
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

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

func PrintNodeGroup(noHeaders bool, ngInfo ...eks.EKSNodeGroupInfo) {
	// Sort the clusterInfos by ClusterName (you can customize the field for sorting)
	sort.Slice(ngInfo, func(i, j int) bool {
		return ngInfo[i].Name < ngInfo[j].Name
	})

	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

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
