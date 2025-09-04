package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/awsconfig"
)

func LoadClusterList(args []string, profile, profile_contains, name_contains, region, version string) ([]ClusterInfo, error) {
	clusterList := []ClusterInfo{}

	// if filters are empty, use current cluster
	if profile == "" && profile_contains == "" && name_contains == "" && region == "" && version == "" {
		clusterArn := ""

		// Load Kubernetes configuration
		config, err := KubernetesConfigFlags.ToRawKubeConfigLoader().RawConfig()
		if err != nil {
			return nil, fmt.Errorf("Error loading kubeconfig: %v\n", err.Error())
		}

		// Get current context
		currentContext := config.CurrentContext
		// fmt.Printf("Current context: %s\n", currentContext)

		// Retrieve cluster information
		contextDetails, exists := config.Contexts[currentContext]
		if !exists {
			return nil, fmt.Errorf("Context '%s' not found in kubeconfig\n", currentContext)
		}

		clusterArn = contextDetails.Cluster

		// check if it is an ARN
		arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
		re := regexp.MustCompile(arnRegex)

		matches := re.FindStringSubmatch(clusterArn)
		if matches == nil {
			if len(args) != 1 {
				return nil, fmt.Errorf("Current cluster is not an EKS cluster\n")
			} else {
				return nil, fmt.Errorf("Invalid cluster ARN: %q\n", clusterArn)
			}
		}

		clusterInfo := loadClusterByArn(clusterArn)

		// clusterInfo := loadClusterByArn(clusterARN)
		if clusterInfo == nil {
			return nil, fmt.Errorf("Cluster not found")
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
