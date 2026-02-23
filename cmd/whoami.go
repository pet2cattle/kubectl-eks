package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current IAM identity and Kubernetes RBAC mapping",
	Long: `Show current assumed IAM identity and how it maps to Kubernetes RBAC.

Combines AWS STS GetCallerIdentity with Kubernetes authentication info
to provide a complete view of your current identity and permissions.`,
	Example: `  # Show current identity
  kubectl eks whoami`,
	Run: func(cmd *cobra.Command, args []string) {
		clusterInfo, err := GetCurrentClusterInfo()
		if err != nil {
			log.Fatalf("Error getting current cluster info: %v", err)
		}

		// Get AWS identity
		ctx := context.Background()
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithSharedConfigProfile(clusterInfo.AWSProfile),
			config.WithRegion(clusterInfo.Region),
		)
		if err != nil {
			log.Fatalf("Error loading AWS config: %v", err)
		}

		stsClient := sts.NewFromConfig(cfg)
		callerIdentity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			log.Fatalf("Error getting caller identity: %v", err)
		}

		// Get Kubernetes identity
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
		}

		clientset, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			log.Fatalf("Error creating kubernetes client: %v", err)
		}

		// Try to get self subject review
		selfSubjectReview := &authv1.SelfSubjectReview{}
		result, err := clientset.AuthenticationV1().SelfSubjectReviews().Create(
			ctx,
			selfSubjectReview,
			metav1.CreateOptions{},
		)

		var k8sUser string
		var k8sGroups []string
		var k8sUID string

		if err != nil {
			// Fallback to just showing the error
			k8sUser = fmt.Sprintf("Error: %v", err)
		} else if result != nil && result.Status.UserInfo.Username != "" {
			k8sUser = result.Status.UserInfo.Username
			k8sGroups = result.Status.UserInfo.Groups
			k8sUID = result.Status.UserInfo.UID
		}

		noHeaders, _ := cmd.Flags().GetBool("no-headers")

		printutils.PrintWhoAmI(
			noHeaders,
			clusterInfo.AWSProfile,
			clusterInfo.Region,
			clusterInfo.ClusterName,
			*callerIdentity.Arn,
			*callerIdentity.Account,
			*callerIdentity.UserId,
			k8sUser,
			k8sUID,
			k8sGroups,
		)
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
