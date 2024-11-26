package cmd

import (
	"fmt"

	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "switch to a different EKS cluster",
	Long:  `Update kubeconfig's context to use a different EKS cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("Usage: %s use <cluster-arn>\n", cmd.Root().Name())
			return
		}

		loadCacheFromDisk()

		clusterARN := args[0]

		clusterInfo := loadClusterByArn(clusterARN)
		if clusterInfo == nil {
			fmt.Println("Cluster not found")
			return
		}

		err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Switched to EKS cluster %q in region %q using profile %q\n", clusterInfo.ClusterName, clusterInfo.Region, clusterInfo.AWSProfile)

	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
