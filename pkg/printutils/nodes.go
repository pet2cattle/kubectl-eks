package printutils

import (
	"fmt"
	"os"

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
			v1.TableColumnDefinition{Name: "MEMORY PRESSURE", Type: "string"},
			v1.TableColumnDefinition{Name: "DISK PRESSURE", Type: "string"},
			v1.TableColumnDefinition{Name: "PID PRESSURE", Type: "string"},
			v1.TableColumnDefinition{Name: "NETWORK UNAVAILABLE", Type: "string"},
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
				n.Node.MemoryPressure,
				n.Node.DiskPressure,
				n.Node.PIDPressure,
				n.Node.NetworkUnavailable,
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
