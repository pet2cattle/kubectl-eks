package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/spf13/cobra"
)

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
	Use:   "use",
	Short: "Switch kubectl context to a different EKS cluster",
	Long: `Switch kubectl context to a different EKS cluster by updating kubeconfig.

Accepts either a cluster ARN or partial cluster name. Automatically updates
your kubeconfig and sets the current context to the specified cluster.

Optionally specify a namespace to set as default, or use a different AWS
profile for authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("Usage: %s use <cluster-arn>\n", cmd.Root().Name())
			return
		}
		clusterArn := strings.TrimSpace(args[0])

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			namespace = ""
		}

		profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			profile = ""
		}

		// check if it is an ARN
		arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
		re := regexp.MustCompile(arnRegex)

		matches := re.FindStringSubmatch(clusterArn)
		if matches == nil {
			fmt.Printf("Invalid cluster ARN: %q\n", clusterArn)
			os.Exit(1)
		}

		SwitchToCluster(clusterArn, namespace, profile)
	},
}

func init() {
	useCmd.Flags().StringP("namespace", "n", "", "Set specific namespace for the context")
	useCmd.Flags().StringP("profile", "p", "", "Set specific AWS profile for the context")

	rootCmd.AddCommand(useCmd)
}
