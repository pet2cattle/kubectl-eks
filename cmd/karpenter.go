package cmd

import (
	"github.com/spf13/cobra"
)

var karpenterCmd = &cobra.Command{
	Use:     "karpenter",
	Aliases: []string{"karp", "kptr", "k"},
	Short:   "Karpenter resource management commands",
	Long: `Manage and inspect Karpenter resources across EKS clusters.

Provides commands to list and inspect Karpenter NodePools, NodeClaims,
AMI usage, and drift status.`,
	Example: `  # List Karpenter NodePools across clusters
  kubectl eks karpenter nodepools
  
  # List active NodeClaims
  kubectl eks karpenter nodeclaims
  
  # Check AMI usage per NodePool
  kubectl eks karpenter ami
  
  # List drifted nodes/nodeclaims
  kubectl eks karpenter drift`,
}

func init() {
	rootCmd.AddCommand(karpenterCmd)
}
