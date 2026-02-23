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

func PrintKarpenterNodePools(noHeaders bool, wide bool, nodePools ...data.KarpenterNodePoolInfo) {
	sort.Slice(nodePools, func(i, j int) bool {
		if nodePools[i].Profile != nodePools[j].Profile {
			return nodePools[i].Profile < nodePools[j].Profile
		}
		if nodePools[i].Region != nodePools[j].Region {
			return nodePools[i].Region < nodePools[j].Region
		}
		if nodePools[i].ClusterName != nodePools[j].ClusterName {
			return nodePools[i].ClusterName < nodePools[j].ClusterName
		}
		return nodePools[i].Name < nodePools[j].Name
	})

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	var columns []v1.TableColumnDefinition
	if wide {
		columns = []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "NODEPOOL", Type: "string"},
			{Name: "NODECLASS", Type: "string"},
			{Name: "INSTANCE TYPES", Type: "string"},
			{Name: "CAPACITY TYPES", Type: "string"},
			{Name: "ZONES", Type: "string"},
			{Name: "CPU LIMIT", Type: "string"},
			{Name: "MEMORY LIMIT", Type: "string"},
			{Name: "CONSOLIDATION", Type: "string"},
			{Name: "EXPIRE AFTER", Type: "string"},
			{Name: "WEIGHT", Type: "number"},
		}
	} else {
		columns = []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "NODEPOOL", Type: "string"},
			{Name: "NODECLASS", Type: "string"},
			{Name: "INSTANCE TYPES", Type: "string"},
			{Name: "CAPACITY TYPES", Type: "string"},
		}
	}

	table := &v1.Table{ColumnDefinitions: columns}

	for _, np := range nodePools {
		instanceTypes := strings.Join(np.InstanceTypes, ",")
		if len(instanceTypes) > 30 && !wide {
			instanceTypes = instanceTypes[:27] + "..."
		}

		capacityTypes := strings.Join(np.CapacityTypes, ",")

		var cells []interface{}
		if wide {
			zones := strings.Join(np.Zones, ",")
			cpuLimit := np.CPULimit
			if cpuLimit == "" {
				cpuLimit = "-"
			}
			memLimit := np.MemoryLimit
			if memLimit == "" {
				memLimit = "-"
			}
			consolidation := np.ConsolidationMode
			if consolidation == "" {
				consolidation = "-"
			}
			expireAfter := np.ExpireAfter
			if expireAfter == "" {
				expireAfter = "-"
			}

			cells = []interface{}{
				np.Profile,
				np.Region,
				np.ClusterName,
				np.Name,
				np.NodeClassName,
				instanceTypes,
				capacityTypes,
				zones,
				cpuLimit,
				memLimit,
				consolidation,
				expireAfter,
				np.Weight,
			}
		} else {
			cells = []interface{}{
				np.Profile,
				np.Region,
				np.ClusterName,
				np.Name,
				np.NodeClassName,
				instanceTypes,
				capacityTypes,
			}
		}

		table.Rows = append(table.Rows, v1.TableRow{Cells: cells})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

func PrintKarpenterNodeClaims(noHeaders bool, wide bool, nodeClaims ...data.KarpenterNodeClaimInfo) {
	sort.Slice(nodeClaims, func(i, j int) bool {
		if nodeClaims[i].Profile != nodeClaims[j].Profile {
			return nodeClaims[i].Profile < nodeClaims[j].Profile
		}
		if nodeClaims[i].Region != nodeClaims[j].Region {
			return nodeClaims[i].Region < nodeClaims[j].Region
		}
		if nodeClaims[i].ClusterName != nodeClaims[j].ClusterName {
			return nodeClaims[i].ClusterName < nodeClaims[j].ClusterName
		}
		return nodeClaims[i].Name < nodeClaims[j].Name
	})

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	var columns []v1.TableColumnDefinition
	if wide {
		columns = []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "NODECLAIM", Type: "string"},
			{Name: "NODE", Type: "string"},
			{Name: "NODEPOOL", Type: "string"},
			{Name: "INSTANCE TYPE", Type: "string"},
			{Name: "ZONE", Type: "string"},
			{Name: "CAPACITY TYPE", Type: "string"},
			{Name: "AMI", Type: "string"},
			{Name: "STATUS", Type: "string"},
			{Name: "DRIFTED", Type: "string"},
			{Name: "AGE", Type: "string"},
		}
	} else {
		columns = []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "NODECLAIM", Type: "string"},
			{Name: "NODE", Type: "string"},
			{Name: "NODEPOOL", Type: "string"},
			{Name: "INSTANCE TYPE", Type: "string"},
			{Name: "STATUS", Type: "string"},
			{Name: "AGE", Type: "string"},
		}
	}

	table := &v1.Table{ColumnDefinitions: columns}

	for _, nc := range nodeClaims {
		humanAge := duration.ShortHumanDuration(time.Since(nc.Age))

		var cells []interface{}
		if wide {
			drifted := "No"
			if nc.Drifted {
				drifted = "Yes"
			}

			cells = []interface{}{
				nc.Profile,
				nc.Region,
				nc.ClusterName,
				nc.Name,
				nc.NodeName,
				nc.NodePoolName,
				nc.InstanceType,
				nc.Zone,
				nc.CapacityType,
				nc.AMI,
				nc.Status,
				drifted,
				humanAge,
			}
		} else {
			cells = []interface{}{
				nc.Profile,
				nc.Region,
				nc.ClusterName,
				nc.Name,
				nc.NodeName,
				nc.NodePoolName,
				nc.InstanceType,
				nc.Status,
				humanAge,
			}
		}

		table.Rows = append(table.Rows, v1.TableRow{Cells: cells})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

func PrintKarpenterAMIUsage(noHeaders bool, amiUsage ...data.KarpenterAMIUsageInfo) {
	sort.Slice(amiUsage, func(i, j int) bool {
		if amiUsage[i].Profile != amiUsage[j].Profile {
			return amiUsage[i].Profile < amiUsage[j].Profile
		}
		if amiUsage[i].Region != amiUsage[j].Region {
			return amiUsage[i].Region < amiUsage[j].Region
		}
		if amiUsage[i].ClusterName != amiUsage[j].ClusterName {
			return amiUsage[i].ClusterName < amiUsage[j].ClusterName
		}
		return amiUsage[i].NodePoolName < amiUsage[j].NodePoolName
	})

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "NODEPOOL", Type: "string"},
			{Name: "CURRENT AMI", Type: "string"},
			{Name: "NODE COUNT", Type: "number"},
		},
	}

	for _, ami := range amiUsage {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				ami.Profile,
				ami.Region,
				ami.ClusterName,
				ami.NodePoolName,
				ami.CurrentAMI,
				ami.NodeCount,
			},
		})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}

func PrintKarpenterDrift(noHeaders bool, driftInfo ...data.KarpenterDriftInfo) {
	sort.Slice(driftInfo, func(i, j int) bool {
		if driftInfo[i].Profile != driftInfo[j].Profile {
			return driftInfo[i].Profile < driftInfo[j].Profile
		}
		if driftInfo[i].Region != driftInfo[j].Region {
			return driftInfo[i].Region < driftInfo[j].Region
		}
		if driftInfo[i].ClusterName != driftInfo[j].ClusterName {
			return driftInfo[i].ClusterName < driftInfo[j].ClusterName
		}
		return driftInfo[i].Name < driftInfo[j].Name
	})

	printer := printers.NewTablePrinter(printers.PrintOptions{NoHeaders: noHeaders})

	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "TYPE", Type: "string"},
			{Name: "NAME", Type: "string"},
			{Name: "NODE", Type: "string"},
			{Name: "NODEPOOL", Type: "string"},
			{Name: "DRIFTED SINCE", Type: "string"},
			{Name: "REASON", Type: "string"},
		},
	}

	for _, drift := range driftInfo {
		humanAge := duration.ShortHumanDuration(time.Since(drift.DriftedSince))

		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				drift.Profile,
				drift.Region,
				drift.ClusterName,
				drift.ResourceType,
				drift.Name,
				drift.NodeName,
				drift.NodePoolName,
				humanAge,
				drift.Reason,
			},
		})
	}

	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
		os.Exit(1)
	}
}
