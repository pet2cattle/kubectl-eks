package cmd

import (
	"fmt"
	"os"

	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List EKS nodes",
	Long:  `List EKS nodes for the current cluster with AWS console-style information`,
	Run: func(cmd *cobra.Command, args []string) {
		noHeaders, err := cmd.Flags().GetBool("no-headers")
		if err != nil {
			noHeaders = false
		}

		nodeList, err := k8s.GetNodes(KubernetesConfigFlags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing nodes: %v\n", err)
			os.Exit(1)
		}

		printutils.PrintNodes(noHeaders, nodeList...)
	},
}

func init() {
	rootCmd.AddCommand(nodesCmd)
}
