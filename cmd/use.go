package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jordiprats/kubectl-eks/pkg/data"
	"github.com/jordiprats/kubectl-eks/pkg/eks"
	"github.com/jordiprats/kubectl-eks/pkg/k8s"
	"github.com/jordiprats/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

func resolveClusterForUse(target, profile, profileContains, nameContains, nameNotContains, region, version string, refresh bool) (*data.ClusterInfo, error) {
	arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
	re := regexp.MustCompile(arnRegex)

	target = strings.TrimSpace(target)

	if re.MatchString(target) {
		clusterInfo := loadClusterByArn(target)
		if clusterInfo == nil {
			return nil, fmt.Errorf("cluster not found")
		}

		return clusterInfo, nil
	}

	if strings.HasPrefix(target, "arn:aws:eks:") {
		return nil, fmt.Errorf("invalid cluster ARN: %q", target)
	}

	clusterList, err := LoadClusterList([]string{}, profile, profileContains, nameContains, nameNotContains, region, version, refresh)
	if err != nil {
		return nil, err
	}

	if target == "" {
		if len(clusterList) == 0 {
			return nil, fmt.Errorf("no clusters matched the provided filters")
		}

		if len(clusterList) > 1 {
			printutils.PrintClusters(false, clusterList...)
			return nil, fmt.Errorf("multiple clusters matched the provided filters; refine the filters")
		}

		return &clusterList[0], nil
	}

	matches := make([]data.ClusterInfo, 0, len(clusterList))
	for _, cluster := range clusterList {
		if strings.Contains(cluster.ClusterName, target) {
			matches = append(matches, cluster)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no clusters matched %q", target)
	}

	if len(matches) > 1 {
		printutils.PrintClusters(false, matches...)
		return nil, fmt.Errorf("multiple clusters matched %q; refine the filters", target)
	}

	return &matches[0], nil
}

func SwitchToCluster(clusterArn, namespace, profile string) {
	clusterInfo := loadClusterByArn(clusterArn)

	// clusterInfo := loadClusterByArn(clusterARN)
	if clusterInfo == nil {
		fmt.Println("Cluster not found")
		os.Exit(1)
	}

	if profile != "" {
		clusterInfo.AWSProfile = profile
	}

	err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, "")
	if err != nil {
		fmt.Printf("Failed to update kubeconfig: %s\n", err.Error())
		os.Exit(1)
	}

	if namespace != "" {
		err = k8s.SetNamespace(namespace)
		if err != nil {
			fmt.Printf("Failed to set namespace: %s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("Switched to EKS cluster %q (namespace: %q) in region %q using profile %q\n", clusterInfo.ClusterName, namespace, clusterInfo.Region, clusterInfo.AWSProfile)
		}
	} else {
		fmt.Printf("Switched to EKS cluster %q in region %q using profile %q\n", clusterInfo.ClusterName, clusterInfo.Region, clusterInfo.AWSProfile)
	}
}

var useCmd = &cobra.Command{
	Use:   "use [cluster-name-or-arn]",
	Short: "Switch kubectl context to a different EKS cluster",
	Long: `Switch kubectl context to a different EKS cluster by updating kubeconfig.

Accepts either a cluster ARN or partial cluster name. When using a partial
name, the command applies the same cluster filters as 'list' and switches only
when exactly one cluster matches.

Optionally specify a namespace to set as default, or use a different AWS
profile for authentication.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := ""
		if len(args) == 1 {
			target = strings.TrimSpace(args[0])
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			namespace = ""
		}

		profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			profile = ""
		}

		profileContains, err := cmd.Flags().GetString("profile-contains")
		if err != nil {
			profileContains = ""
		}

		nameContains, err := cmd.Flags().GetString("name-contains")
		if err != nil {
			nameContains = ""
		}

		nameNotContains, err := cmd.Flags().GetString("name-not-contains")
		if err != nil {
			nameNotContains = ""
		}

		region, err := cmd.Flags().GetString("region")
		if err != nil {
			region = ""
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			version = ""
		}

		refresh, err := cmd.Flags().GetBool("refresh")
		if err != nil {
			refresh = false
		}

		clusterInfo, err := resolveClusterForUse(target, profile, profileContains, nameContains, nameNotContains, region, version, refresh)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		SwitchToCluster(clusterInfo.Arn, namespace, profile)
	},
}

func init() {
	useCmd.Flags().BoolP("refresh", "u", false, "Refresh data from AWS")
	useCmd.Flags().StringP("namespace", "n", "", "Set specific namespace for the context")
	useCmd.Flags().StringP("profile", "p", "", "Set specific AWS profile for the context")
	useCmd.Flags().StringP("profile-contains", "q", "", "AWS profile contains string")
	useCmd.Flags().StringP("name-contains", "c", "", "Cluster name contains string")
	useCmd.Flags().StringP("name-not-contains", "x", "", "Cluster name does not contain string")
	useCmd.Flags().StringP("region", "r", "", "AWS region to use")
	useCmd.Flags().StringP("version", "v", "", "Filter by EKS version")

	rootCmd.AddCommand(useCmd)
}
