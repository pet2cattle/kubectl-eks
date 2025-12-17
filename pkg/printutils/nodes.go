package printutils

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/printers"
)

func PrintNodes(noHeaders bool, nodes ...k8s.NodeInfo) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "NODE NAME", Type: "string"},
			{Name: "INSTANCE TYPE", Type: "string"},
			{Name: "COMPUTE", Type: "string"},
			{Name: "MANAGED BY", Type: "string"},
			{Name: "AGE", Type: "string"},
			{Name: "STATUS", Type: "string"},
		},
	}

	for _, node := range nodes {
		age := duration.ShortHumanDuration(time.Since(node.Created))

		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				node.Name,
				node.InstanceType,
				node.Compute,
				node.ManagedBy,
				age,
				node.Status,
			},
		})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}
