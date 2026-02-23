package printutils

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/printers"
)

func PrintWhoAmI(noHeaders bool, awsProfile, region, clusterName, awsArn, awsAccount, awsUserId, k8sUsername, k8sUID string, k8sGroups []string) {
	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "COMPONENT", Type: "string"},
			{Name: "ATTRIBUTE", Type: "string"},
			{Name: "VALUE", Type: "string"},
		},
	}

	// Cluster info
	table.Rows = append(table.Rows, v1.TableRow{
		Cells: []interface{}{"Cluster", "Profile", awsProfile},
	})
	table.Rows = append(table.Rows, v1.TableRow{
		Cells: []interface{}{"Cluster", "Region", region},
	})
	table.Rows = append(table.Rows, v1.TableRow{
		Cells: []interface{}{"Cluster", "Name", clusterName},
	})

	// AWS Identity
	table.Rows = append(table.Rows, v1.TableRow{
		Cells: []interface{}{"AWS", "ARN", awsArn},
	})
	table.Rows = append(table.Rows, v1.TableRow{
		Cells: []interface{}{"AWS", "Account", awsAccount},
	})
	table.Rows = append(table.Rows, v1.TableRow{
		Cells: []interface{}{"AWS", "User ID", awsUserId},
	})

	// Kubernetes Identity
	table.Rows = append(table.Rows, v1.TableRow{
		Cells: []interface{}{"Kubernetes", "Username", k8sUsername},
	})
	if k8sUID != "" {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{"Kubernetes", "UID", k8sUID},
		})
	}
	if len(k8sGroups) > 0 {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{"Kubernetes", "Groups", strings.Join(k8sGroups, ", ")},
		})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

func PrintResourceQuotas(noHeaders bool, quotas ...data.ResourceQuotaInfo) {
	sort.Slice(quotas, func(i, j int) bool {
		if quotas[i].Profile != quotas[j].Profile {
			return quotas[i].Profile < quotas[j].Profile
		}
		if quotas[i].Region != quotas[j].Region {
			return quotas[i].Region < quotas[j].Region
		}
		if quotas[i].ClusterName != quotas[j].ClusterName {
			return quotas[i].ClusterName < quotas[j].ClusterName
		}
		if quotas[i].Namespace != quotas[j].Namespace {
			return quotas[i].Namespace < quotas[j].Namespace
		}
		if quotas[i].QuotaName != quotas[j].QuotaName {
			return quotas[i].QuotaName < quotas[j].QuotaName
		}
		return quotas[i].ResourceName < quotas[j].ResourceName
	})

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER", Type: "string"},
			{Name: "NAMESPACE", Type: "string"},
			{Name: "QUOTA NAME", Type: "string"},
			{Name: "RESOURCE", Type: "string"},
			{Name: "USED", Type: "string"},
			{Name: "HARD", Type: "string"},
		},
	}

	for _, quota := range quotas {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				quota.Profile,
				quota.Region,
				quota.ClusterName,
				quota.Namespace,
				quota.QuotaName,
				quota.ResourceName,
				quota.Used,
				quota.Hard,
			},
		})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

func PrintEvents(noHeaders bool, events ...data.EventInfo) {
	// Sort by timestamp (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].LastSeen.After(events[j].LastSeen)
	})

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "LAST SEEN", Type: "string"},
			{Name: "TYPE", Type: "string"},
			{Name: "REASON", Type: "string"},
			{Name: "OBJECT", Type: "string"},
			{Name: "MESSAGE", Type: "string"},
			{Name: "COUNT", Type: "number"},
		},
	}

	for _, event := range events {
		humanAge := duration.ShortHumanDuration(time.Since(event.LastSeen))

		// Truncate message if too long
		message := event.Message
		if len(message) > 80 {
			message = message[:77] + "..."
		}

		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				humanAge,
				event.Type,
				event.Reason,
				event.Object,
				message,
				event.Count,
			},
		})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}
