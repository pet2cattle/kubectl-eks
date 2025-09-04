package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/awsconfig"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadClusterList(args []string, profile, profile_contains, name_contains, region, version string) ([]ClusterInfo, error) {
	clusterList := []ClusterInfo{}

	// if filters are empty, use current cluster
	if profile == "" && profile_contains == "" && name_contains == "" && region == "" && version == "" {
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
					if version == "" && name_contains == "" {
						clusterList = append(clusterList, currentClusterList...)
					} else {
						for _, cluster := range currentClusterList {
							// checking filter criteria
							shouldAdd := false

							if version != "" {
								if cluster.Version == version {
									shouldAdd = true
								}
							}

							if name_contains != "" {
								if strings.Contains(cluster.ClusterName, name_contains) {
									if version == "" {
										shouldAdd = true
									}
								} else {
									// resetting shouldAdd to false if name_contains is set and the cluster name does not contain the string
									shouldAdd = false
								}
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

var multiGetPodsCmd = &cobra.Command{
	Use:   "mget-pods",
	Short: "Get a list of pods from multiple clusters",
	Long:  `Get a list of pods from the current cluster or all the clusters that match a filter`,
	Run: func(cmd *cobra.Command, args []string) {
		profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			profile = ""
		}

		profile_contains, err := cmd.Flags().GetString("profile-contains")
		if err != nil {
			profile_contains = ""
		}

		name_contains, err := cmd.Flags().GetString("name-contains")
		if err != nil {
			name_contains = ""
		}

		region, err := cmd.Flags().GetString("region")
		if err != nil {
			region = ""
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			version = ""
		}

		clusterList, err := LoadClusterList(args, profile, profile_contains, name_contains, region, version)
		if err != nil {
			log.Fatalf("Error loading cluster list: %v", err)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			namespace = ""
		}

		allNamespaces, err := cmd.Flags().GetBool("all-namespaces")
		if err != nil {
			allNamespaces = false
		}

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err := loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		previousContext := config.CurrentContext

		// current k8s context
		k8SClusterPodList := []k8s.K8SClusterPodList{}
		for _, clusterInfo := range clusterList {
			err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, "")
			if err != nil {
				continue
			}

			k8sPodList, err := k8s.GetPods(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, clusterInfo.Arn, clusterInfo.Version, namespace, allNamespaces)
			if err != nil {
				continue
			} else {
				k8SClusterPodList = append(k8SClusterPodList, *k8sPodList)
			}
		}

		// Restore the previous context
		loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
		config, err = loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		config.CurrentContext = previousContext
		if err := clientcmd.ModifyConfig(loadingRules, *config, true); err != nil {
			log.Fatalf("Error updating kubeconfig: %v", err)
		}

		PrintMultiGetPods(k8SClusterPodList...)

		saveCacheToDisk()
	},
}

func init() {
	multiGetPodsCmd.Flags().StringP("profile", "p", "", "AWS profile to use")
	multiGetPodsCmd.Flags().StringP("profile-contains", "q", "", "AWS profile contains string")
	multiGetPodsCmd.Flags().StringP("name-contains", "c", "", "Cluster name contains string")
	multiGetPodsCmd.Flags().StringP("region", "r", "", "AWS region to use")
	multiGetPodsCmd.Flags().StringP("version", "v", "", "Filter by EKS version")
	multiGetPodsCmd.Flags().StringP("namespace", "n", "", "Filter by Kubernetes namespace")
	multiGetPodsCmd.Flags().BoolP("all-namespaces", "A", false, "Filter by all Kubernetes namespaces")

	rootCmd.AddCommand(multiGetPodsCmd)
}
