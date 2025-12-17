package printutils

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

func PrintFargateProfiles(noHeaders bool, profiles ...eks.FargateProfileInfo) {
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "NAME", Type: "string"},
			{Name: "STATUS", Type: "string"},
			{Name: "NAMESPACE", Type: "string"},
			{Name: "SELECTOR", Type: "string"},
			{Name: "SUBNETS", Type: "number"},
		},
	}

	for _, p := range profiles {
		for _, sel := range p.Selectors {
			table.Rows = append(table.Rows, v1.TableRow{
				Cells: []interface{}{
					p.Name,
					p.Status,
					sel.Namespace,
					formatLabels(sel.Labels),
					len(p.Subnets),
				},
			})
		}
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "<none>"
	}

	pairs := make([]string, 0, len(labels))
	for k, v := range labels {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}
