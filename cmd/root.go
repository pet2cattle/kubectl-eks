package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/pet2cattle/kubectl-eks/pkg/awsconfig"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/pet2cattle/kubectl-eks/pkg/sts"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
)

type ClusterInfo struct {
	ClusterName  string
	Region       string
	AWSProfile   string
	AWSAccountID string
	Status       string
	Version      string
	Arn          string
	CreatedAt    string
}

type KubeCtlEksCache struct {
	ClusterByARN map[string]ClusterInfo
	ClusterList  map[string]map[string][]ClusterInfo
}

var KubernetesConfigFlags *genericclioptions.ConfigFlags

var HomeDir string
var CachedData *KubeCtlEksCache = nil

func loadCacheFromDisk() {
	// Load configuration from file
	configFile := HomeDir + "/.kube/.kubectl-eks-cache"
	configData, err := os.ReadFile(configFile)
	if err == nil {
		// load json data into ConfigData
		CachedData = &KubeCtlEksCache{}
		err = json.Unmarshal(configData, CachedData)
		if err != nil {
			fmt.Println("Error loading configuration file")
			os.Exit(1)
		}
	}
}

func saveCacheToDisk() {
	// Save configuration to file
	configFile := HomeDir + "/.kube/.kubectl-eks-cache"
	configData, err := json.Marshal(CachedData)
	if err != nil {
		fmt.Println("Error saving configuration file")
		os.Exit(1)
	}
	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		fmt.Println("Error saving configuration file")
		os.Exit(1)
	}
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

func loadClusterByArn(clusterArn string) *ClusterInfo {

	clusterInfo := ClusterInfo{}

	// check if it is an ARN
	arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
	re := regexp.MustCompile(arnRegex)

	matches := re.FindStringSubmatch(clusterArn)
	if matches == nil {
		return nil
	}

	// search for an AWS profile that matches the account ID, region and cluster name
	awsProfiles := awsconfig.GetAWSProfilesWithEKSHints()
	foundAwsProfile := ""
	for _, profileDetails := range awsProfiles {
		accountID, err := sts.GetAccountID(profileDetails.Name, matches[1])
		if err != nil {
			continue
		}

		if accountID != matches[2] {
			continue
		}

		for _, hintRegion := range profileDetails.HintEKSRegions {
			if hintRegion != matches[1] {
				continue
			}

			// aws eks list-clusters --region <region> --profile <profile>
			clusters, err := eks.GetClusters(profileDetails.Name, hintRegion)
			if err != nil {
				continue
			}

			for _, cluster := range clusters {
				if cluster != nil {
					if *cluster == matches[3] {
						foundAwsProfile = profileDetails.Name
					}
				}
			}
		}
	}

	if foundAwsProfile == "" {
		foundAwsProfile = "-"
	}

	// create clusterInfo
	clusterInfo = ClusterInfo{ClusterName: matches[3], Region: matches[1], AWSProfile: foundAwsProfile, AWSAccountID: matches[2]}

	clusterDesc, err := eks.DescribeCluster(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName)
	if err != nil || clusterDesc == nil {
		fmt.Fprintf(os.Stderr, "Error describing cluster %s: %v\n", clusterInfo.ClusterName, err.Error())
	} else {
		clusterInfo.Status = *clusterDesc.Status
		clusterInfo.Version = *clusterDesc.Version
		clusterInfo.Arn = *clusterDesc.Arn
		clusterInfo.CreatedAt = clusterDesc.CreatedAt.Format("2006-01-02 15:04:05")
	}

	if CachedData == nil {
		CachedData = &KubeCtlEksCache{}
	}

	if CachedData.ClusterByARN == nil {
		CachedData.ClusterByARN = make(map[string]ClusterInfo)
	}

	// save update loaded configuration
	CachedData.ClusterByARN[clusterArn] = clusterInfo

	return &clusterInfo
}

var rootCmd = &cobra.Command{
	Use:   "kubectl-eks",
	Short: "A kubectl plugin for Amazon EKS",
	Long:  `A kubectl plugin for Amazon EKS`,
	Run: func(cmd *cobra.Command, args []string) {
		refresh, err := cmd.Flags().GetBool("refresh")
		if err != nil {
			refresh = false
		}

		region, err := cmd.Flags().GetString("region")
		if err != nil {
			region = ""
		}

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

		clusterArn := contextDetails.Cluster

		if region != "" {
			// arn:aws:eks:us-east-1:123456789123:cluster/demo

			arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
			re := regexp.MustCompile(arnRegex)
			matches := re.FindStringSubmatch(clusterArn)
			if matches == nil {
				fmt.Printf("Invalid cluster ARN: %q\n", clusterArn)
				os.Exit(1)
			}

			newClusterArn := fmt.Sprintf("arn:aws:eks:%s:%s:cluster/%s", region, matches[2], matches[3])

			currentNamespace, err := k8s.GetCurrentNamespace()
			if err != nil {
				currentNamespace = ""
			}

			SwitchToCluster(newClusterArn, currentNamespace, "")

		} else {
			loadCacheFromDisk()
			if CachedData == nil {
				CachedData = &KubeCtlEksCache{
					ClusterByARN: make(map[string]ClusterInfo),
					ClusterList:  make(map[string]map[string][]ClusterInfo),
				}
			}

			clusterInfo, exists := CachedData.ClusterByARN[clusterArn]
			if !exists || refresh {
				foundClusterInfo := loadClusterByArn(clusterArn)

				if foundClusterInfo == nil {
					fmt.Println("Current cluster is not an EKS cluster")
					os.Exit(1)
				} else {
					clusterInfo = *foundClusterInfo
				}
			}

			// validate cached data, if invalid, refresh
			if clusterInfo.Arn != clusterArn {
				CachedData = &KubeCtlEksCache{
					ClusterByARN: make(map[string]ClusterInfo),
					ClusterList:  make(map[string]map[string][]ClusterInfo),
				}
				foundClusterInfo := loadClusterByArn(clusterArn)
				if foundClusterInfo == nil {
					fmt.Println("Current cluster is not an EKS cluster")
					os.Exit(1)
				} else {
					clusterInfo = *foundClusterInfo
				}
			}

			noHeaders, err := cmd.Flags().GetBool("no-headers")
			if err != nil {
				noHeaders = false
			}

			if clusterInfo.Arn != clusterArn {
				fmt.Printf("%s\n", clusterArn)
			} else {
				PrintClusters(noHeaders, clusterInfo)
			}

			// save data to configuration
			saveCacheToDisk()

			os.Exit(0)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("region", "r", "", "Switch to the same cluster in a different region")
	rootCmd.Flags().BoolP("refresh", "u", false, "Do not use cached data, refresh from AWS")
	rootCmd.PersistentFlags().Bool("no-headers", false, "When using the default or custom-column output format, don't print headers (default print headers)")

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(true)
	KubernetesConfigFlags.AddFlags(rootCmd.PersistentFlags())
}
