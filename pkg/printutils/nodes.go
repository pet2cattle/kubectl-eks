package printutils

import (
	"fmt"
	"os"
	"strings"

	"github.com/jordiprats/kubectl-eks/pkg/data"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

func PrintMultiClusterNodes(noHeaders bool, wide bool, nodes []data.ClusterNodeInfo) {
	if len(nodes) == 0 {
		return
	}

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "NAME", Type: "string"},
			{Name: "STATUS", Type: "string"},
			{Name: "INSTANCE TYPE", Type: "string"},
			{Name: "COMPUTE", Type: "string"},
			{Name: "MANAGED BY", Type: "string"},
			{Name: "AGE", Type: "string"},
		},
	}

	if wide {
		table.ColumnDefinitions = append(table.ColumnDefinitions,
			v1.TableColumnDefinition{Name: "CPU", Type: "string"},
			v1.TableColumnDefinition{Name: "MEMORY", Type: "string"},
			v1.TableColumnDefinition{Name: "PODS", Type: "string"},
			v1.TableColumnDefinition{Name: "NODE CONDITIONS", Type: "string"},
		)
	}

	for _, n := range nodes {
		cells := []interface{}{
			n.Profile,
			n.Region,
			n.ClusterName,
			n.Node.Name,
			n.Node.Status,
			n.Node.InstanceType,
			n.Node.Compute,
			n.Node.ManagedBy,
			formatAge(n.Node.Created),
		}

		if wide {
			cells = append(cells,
				formatCapacityAllocatable(n.Node.CPUCapacity, n.Node.CPUAllocatable),
				formatCapacityAllocatable(n.Node.MemoryCapacity, n.Node.MemoryAllocatable),
				formatCapacityAllocatable(n.Node.PodsCapacity, n.Node.PodsAllocatable),
				formatNodeConditions(n.Node),
			)
		}

		table.Rows = append(table.Rows, v1.TableRow{Cells: cells})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

func formatNodeConditions(node data.NodeInfo) string {
	conditions := make([]string, 0, 4)

	if node.MemoryPressure == "True" {
		conditions = append(conditions, "MemoryPressure")
	}
	if node.DiskPressure == "True" {
		conditions = append(conditions, "DiskPressure")
	}
	if node.PIDPressure == "True" {
		conditions = append(conditions, "PIDPressure")
	}
	if node.NetworkUnavailable == "True" {
		conditions = append(conditions, "NetworkUnavailable")
	}

	if len(conditions) == 0 {
		return "-"
	}

	return strings.Join(conditions, ",")
}

func formatCapacityAllocatable(capacity, allocatable string) string {
	return fmt.Sprintf("%s/%s", capacity, allocatable)
}
