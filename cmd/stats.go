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

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get EKS cluster stats",
	Long:  `Get EKS cluster stats of the current cluster or all the clusters that matches a filter`,
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

		clusterList := []ClusterInfo{}

		// if filters are empty, use current cluster
		if profile == "" && profile_contains == "" && name_contains == "" && region == "" && version == "" {
			clusterArn := ""

			// Load Kubernetes configuration
			config, err := KubernetesConfigFlags.ToRawKubeConfigLoader().RawConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading kubeconfig: %v\n", err.Error())
				os.Exit(1)
			}

			// Get current context
			currentContext := config.CurrentContext
			// fmt.Printf("Current context: %s\n", currentContext)

			// Retrieve cluster information
			contextDetails, exists := config.Contexts[currentContext]
			if !exists {
				fmt.Fprintf(os.Stderr, "Context '%s' not found in kubeconfig\n", currentContext)
				os.Exit(1)
			}

			clusterArn = contextDetails.Cluster

			// check if it is an ARN
			arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
			re := regexp.MustCompile(arnRegex)

			matches := re.FindStringSubmatch(clusterArn)
			if matches == nil {
				if len(args) != 1 {
					fmt.Printf("Current cluster is not an EKS cluster\n")
				} else {
					fmt.Printf("Invalid cluster ARN: %q\n", clusterArn)
				}
				os.Exit(1)
			}

			clusterInfo := loadClusterByArn(clusterArn)

			// clusterInfo := loadClusterByArn(clusterARN)
			if clusterInfo == nil {
				fmt.Println("Cluster not found")
				return
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

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err := loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		previousContext := config.CurrentContext

		// current k8s context
		k8sStatsList := []k8s.K8Sstats{}
		for _, clusterInfo := range clusterList {
			err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, "")
			if err != nil {
				continue
			}

			stats, err := k8s.GetK8sStats(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, clusterInfo.Arn, clusterInfo.Version)
			if err != nil {
				continue
			} else {
				k8sStatsList = append(k8sStatsList, *stats)
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

		PrintK8SStats(k8sStatsList...)

		saveCacheToDisk()
	},
}

func init() {
	statsCmd.Flags().StringP("profile", "p", "", "AWS profile to use")
	statsCmd.Flags().StringP("profile-contains", "q", "", "AWS profile contains string")
	statsCmd.Flags().StringP("name-contains", "c", "", "Cluster name contains string")
	statsCmd.Flags().StringP("region", "r", "", "AWS region to use")
	statsCmd.Flags().StringP("version", "v", "", "Filter by EKS version")

	rootCmd.AddCommand(statsCmd)
}
