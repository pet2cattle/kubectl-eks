package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/pet2cattle/kubectl-eks/pkg/awsconfig"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
)

type ClusterInfo struct {
	ClusterName  string
	Region       string
	AWSProfile   string
	AWSAccountID string
}

type KubeCtlEksConfig struct {
	Clusters map[string]ClusterInfo
}

var KubernetesConfigFlags *genericclioptions.ConfigFlags

var HomeDir string
var ConfigData *KubeCtlEksConfig = nil

func loadConfig() {
	// Load configuration from file
	configFile := HomeDir + "/.kube/kubectl-eks-config"
	configData, err := os.ReadFile(configFile)
	if err == nil {
		// load json data into ConfigData
		ConfigData = &KubeCtlEksConfig{}
		err = json.Unmarshal(configData, ConfigData)
		if err != nil {
			fmt.Println("Error loading configuration file")
			os.Exit(1)
		}
	}
}

func saveConfig() {
	// Save configuration to file
	configFile := HomeDir + "/.kube/kubectl-eks-config"
	configData, err := json.Marshal(ConfigData)
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

// printResults prints results in a kubectl-style table format
func PrintClusters(clusterInfos ...ClusterInfo) {
	// Create a table printer
	printer := printers.NewTablePrinter(printers.PrintOptions{})

	// Create a Table object
	table := &v1.Table{
		ColumnDefinitions: []v1.TableColumnDefinition{
			{Name: "CLUSTER NAME", Type: "string"},
			{Name: "AWS PROFILE", Type: "string"},
			{Name: "AWS ACCOUNT ID", Type: "string"},
			{Name: "AWS REGION", Type: "string"},
		},
	}

	// Populate rows with data from the variadic ClusterInfo
	for _, clusterInfo := range clusterInfos {
		table.Rows = append(table.Rows, v1.TableRow{
			Cells: []interface{}{
				clusterInfo.ClusterName,
				clusterInfo.AWSProfile,
				clusterInfo.AWSAccountID,
				clusterInfo.Region,
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

var rootCmd = &cobra.Command{
	Use:   "kubectl-eks",
	Short: "A kubectl plugin for Amazon EKS",
	Long:  `A kubectl plugin for Amazon EKS`,
	Run: func(cmd *cobra.Command, args []string) {
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

		clusterName := contextDetails.Cluster
		// fmt.Printf("Cluster name: %s\n", clusterName)

		loadConfig()
		if ConfigData == nil {
			ConfigData = &KubeCtlEksConfig{Clusters: make(map[string]ClusterInfo)}
		}

		clusterInfo, exists := ConfigData.Clusters[clusterName]
		if !exists {
			// check if it is an ARN
			arnRegex := `^arn:aws:eks:([a-z0-9-]+):(\d{12}):cluster/([a-zA-Z0-9-]+)$`
			re := regexp.MustCompile(arnRegex)

			matches := re.FindStringSubmatch(clusterName)
			if matches == nil {
				fmt.Println("Current cluster is not an EKS cluster")
				os.Exit(1)
			}

			// search for an AWS profile that matches the account ID, region and cluster name
			awsProfiles := awsconfig.GetAWSProfilesWithEKSHints()
			foundAwsProfile := ""
			for _, profileDetails := range awsProfiles {
				for _, hintRegion := range profileDetails.HintEKSRegions {
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

			// save update loaded configuration
			ConfigData.Clusters[clusterName] = clusterInfo

			// save data to configuration
			saveConfig()
		}

		PrintClusters(clusterInfo)
		os.Exit(0)
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
	KubernetesConfigFlags = genericclioptions.NewConfigFlags(true)
	KubernetesConfigFlags.AddFlags(rootCmd.PersistentFlags())
}
