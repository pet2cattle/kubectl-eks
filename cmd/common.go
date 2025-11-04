package cmd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pet2cattle/kubectl-eks/pkg/awsconfig"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/printers"
)

func LoadClusterList(args []string, profile, profile_contains, name_contains, name_not_contains, region, version string) ([]ClusterInfo, error) {
	clusterList := []ClusterInfo{}

	// if filters are empty, use current cluster
	if profile == "" && profile_contains == "" && name_contains == "" && name_not_contains == "" && region == "" && version == "" {
		clusterArn := ""

		// Load Kubernetes configuration
		config, err := KubernetesConfigFlags.ToRawKubeConfigLoader().RawConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading kubeconfig: %v", err.Error())
		}

		// Get current context
		currentContext := config.CurrentContext
		// fmt.Printf("Current context: %s\n", currentContext)

		// Retrieve cluster information
		contextDetails, exists := config.Contexts[currentContext]
		if !exists {
			return nil, fmt.Errorf("context '%s' not found in kubeconfig", currentContext)
		}

		clusterArn = contextDetails.Cluster

		// check if it is an ARN
		arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
		re := regexp.MustCompile(arnRegex)

		matches := re.FindStringSubmatch(clusterArn)
		if matches == nil {
			if len(args) != 1 {
				return nil, fmt.Errorf("current cluster is not an EKS cluster")
			} else {
				return nil, fmt.Errorf("invalid cluster ARN: %q", clusterArn)
			}
		}

		clusterInfo := loadClusterByArn(clusterArn)

		// clusterInfo := loadClusterByArn(clusterARN)
		if clusterInfo == nil {
			return nil, fmt.Errorf("cluster not found")
		}

		clusterList = append(clusterList, *clusterInfo)

	} else {
		loadCacheFromDisk()
		if CachedData == nil {
			CachedData = &KubeCtlEksCache{
				ClusterByARN: make(map[string]ClusterInfo),
				ClusterList:  make(map[string]map[string][]ClusterInfo),
			}
		}

		awsProfiles := awsconfig.GetAWSProfilesWithEKSHints()
		for _, profileDetails := range awsProfiles {
			if profile != "" && profile != profileDetails.Name {
				continue
			}
			if profile_contains != "" && !strings.Contains(profileDetails.Name, profile_contains) {
				continue
			}
			for _, hintRegion := range profileDetails.HintEKSRegions {
				if region != "" && region != hintRegion {
					continue
				}

				cachedRegions, exists := CachedData.ClusterList[profileDetails.Name]
				if !exists {
					loadClusters(profileDetails.Name, hintRegion)
				} else {
					_, exists := cachedRegions[hintRegion]
					if !exists {
						loadClusters(profileDetails.Name, hintRegion)
					}
				}

				currentClusterList, exists := CachedData.ClusterList[profileDetails.Name][hintRegion]
				if !exists {
					fmt.Fprintf(os.Stderr, "Unable to load clusters using profile: %s region: %s\n", profileDetails.Name, hintRegion)
				} else {
					if version == "" && name_contains == "" && name_not_contains == "" {
						clusterList = append(clusterList, currentClusterList...)
					} else {
						for _, cluster := range currentClusterList {
							// checking filter criteria
							shouldAdd := true

							// Check version filter
							if version != "" && cluster.Version != version {
								shouldAdd = false
							}

							// Check name_contains filter
							if name_contains != "" && !strings.Contains(cluster.ClusterName, name_contains) {
								shouldAdd = false
							}

							// Check name_not_contains filter
							if name_not_contains != "" && strings.Contains(cluster.ClusterName, name_not_contains) {
								shouldAdd = false
							}

							// only add the cluster if it meets the criteria
							if shouldAdd {
								clusterList = append(clusterList, cluster)
							}
						}
					}
				}

			}
		}
	}

	return clusterList, nil
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
func PrintClusters(noHeaders bool, clusterInfos ...ClusterInfo) {
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
