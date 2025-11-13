package printutils

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/pet2cattle/kubectl-eks/pkg/cf"
	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"go.yaml.in/yaml/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/printers"
)

func PrintGenericResults(results []data.ResourceResult, output string, noHeaders bool) {
	if len(results) == 0 {
		return
	}

	// Handle JSON output
	if output == "json" {
		jsonBytes, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(jsonBytes))
		return
	}

	// Handle YAML output
	if output == "yaml" {
		// You'll need to import "gopkg.in/yaml.v3"
		for i, result := range results {
			if result.Error != "" {
				fmt.Printf("# Error for %s/%s in %s: %s\n", result.Namespace, result.Name, result.ClusterName, result.Error)
				continue
			}
			yamlBytes, _ := yaml.Marshal(result.Data)
			fmt.Println("---")
			fmt.Print(string(yamlBytes))
			if i < len(results)-1 {
				fmt.Println()
			}
		}
		return
	}

	// Table output (default and wide)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print headers
	if !noHeaders {
		if output == "wide" {
			fmt.Fprintln(w, "AWS PROFILE\tAWS REGION\tCLUSTER NAME\tNAMESPACE\tKIND\tNAME\tSTATUS\tAGE\tADDITIONAL INFO")
		} else {
			fmt.Fprintln(w, "AWS PROFILE\tAWS REGION\tCLUSTER NAME\tNAMESPACE\tKIND\tNAME\tSTATUS")
		}
	}

	// Print rows
	for _, result := range results {
		if result.Error != "" {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\tERROR: %s\n",
				result.Profile,
				result.Region,
				result.ClusterName,
				result.Namespace,
				result.Kind,
				result.Name,
				result.Error,
			)
			continue
		}

		status := extractStatus(result.Data, result.Kind)

		if output == "wide" {
			age := extractAge(result.Data)
			additionalInfo := extractAdditionalInfo(result.Data, result.Kind)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				result.Profile,
				result.Region,
				result.ClusterName,
				result.Namespace,
				result.Kind,
				result.Name,
				status,
				age,
				additionalInfo,
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				result.Profile,
				result.Region,
				result.ClusterName,
				result.Namespace,
				result.Kind,
				result.Name,
				status,
			)
		}
	}

	w.Flush()
}

func extractStatus(data interface{}, kind string) string {
	if data == nil {
		return "Unknown"
	}

	// Convert to JSON and back to map for easier access
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "Unknown"
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return "Unknown"
	}

	status, ok := obj["status"].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	switch kind {
	case "Pod":
		if phase, ok := status["phase"].(string); ok {
			return phase
		}
	case "Deployment", "StatefulSet", "DaemonSet":
		conditions, ok := status["conditions"].([]interface{})
		if ok && len(conditions) > 0 {
			for _, cond := range conditions {
				condition, ok := cond.(map[string]interface{})
				if !ok {
					continue
				}
				if condType, ok := condition["type"].(string); ok && condType == "Available" {
					if condStatus, ok := condition["status"].(string); ok {
						if condStatus == "True" {
							return "Available"
						}
						return "Unavailable"
					}
				}
			}
		}
		// Fallback to replica status
		if kind == "Deployment" {
			replicas, _ := status["replicas"].(float64)
			readyReplicas, _ := status["readyReplicas"].(float64)
			return fmt.Sprintf("%d/%d", int(readyReplicas), int(replicas))
		}
	case "Service":
		if clusterIP, ok := status["clusterIP"].(string); ok {
			return clusterIP
		}
		return "Active"
	case "ConfigMap", "Secret":
		return "Active"
	}

	return "Unknown"
}

func extractAge(data interface{}) string {
	if data == nil {
		return "Unknown"
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "Unknown"
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return "Unknown"
	}

	metadata, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	creationTimestamp, ok := metadata["creationTimestamp"].(string)
	if !ok {
		return "Unknown"
	}

	// Parse the timestamp
	t, err := time.Parse(time.RFC3339, creationTimestamp)
	if err != nil {
		return "Unknown"
	}

	// Calculate age
	duration := time.Since(t)

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

func extractAdditionalInfo(data interface{}, kind string) string {
	if data == nil {
		return ""
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return ""
	}

	spec, ok := obj["spec"].(map[string]interface{})
	if !ok {
		return ""
	}

	switch kind {
	case "Pod":
		// Show container images
		containers, ok := spec["containers"].([]interface{})
		if ok && len(containers) > 0 {
			images := []string{}
			for _, c := range containers {
				container, ok := c.(map[string]interface{})
				if ok {
					if image, ok := container["image"].(string); ok {
						images = append(images, image)
					}
				}
			}
			return strings.Join(images, ", ")
		}
	case "Deployment", "StatefulSet", "DaemonSet":
		// Show replicas
		if status, ok := obj["status"].(map[string]interface{}); ok {
			replicas, _ := status["replicas"].(float64)
			readyReplicas, _ := status["readyReplicas"].(float64)
			return fmt.Sprintf("Replicas: %d/%d", int(readyReplicas), int(replicas))
		}
	case "Service":
		// Show type and ports
		svcType, _ := spec["type"].(string)
		ports, ok := spec["ports"].([]interface{})
		if ok && len(ports) > 0 {
			portStrs := []string{}
			for _, p := range ports {
				port, ok := p.(map[string]interface{})
				if ok {
					portNum, _ := port["port"].(float64)
					protocol, _ := port["protocol"].(string)
					portStrs = append(portStrs, fmt.Sprintf("%d/%s", int(portNum), protocol))
				}
			}
			return fmt.Sprintf("Type: %s, Ports: %s", svcType, strings.Join(portStrs, ", "))
		}
		return fmt.Sprintf("Type: %s", svcType)
	case "ConfigMap":
		if data, ok := obj["data"].(map[string]interface{}); ok {
			return fmt.Sprintf("Keys: %d", len(data))
		}
	case "Secret":
		if data, ok := obj["data"].(map[string]interface{}); ok {
			return fmt.Sprintf("Keys: %d", len(data))
		}
	}

	return ""
}

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
	table := &v1.Table{
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

	// Populate rows with data from the variadic ClusterInfo
	for _, clusterInfo := range clusterInfos {
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
