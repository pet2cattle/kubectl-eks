package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/jsonpath"
)

var mJsonpathCmd = &cobra.Command{
	Use:   "mget-jsonpath [resource-type] [resource-name] [jsonpath-expression]",
	Short: "Extract JSONPath values from resources across multiple clusters",
	Long:  `Extract values using JSONPath expressions from Kubernetes resources across clusters that match filters. Example: kubectl eks mget-jsonpath pod my-pod '{.spec.dnsPolicy}' or kubectl eks mget-jsonpath pods '{.spec.dnsPolicy}' to query all pods`,
	Args:  cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		var resourceType, resourceName, jsonpathExpr string

		// Parse arguments
		if len(args) == 2 {
			resourceType = args[0]
			jsonpathExpr = args[1]
			resourceName = "" // Query all resources
		} else {
			resourceType = args[0]
			resourceName = args[1]
			jsonpathExpr = args[2]
		}

		// Normalize JSONPath expression (remove outer braces if present)
		jsonpathExpr = strings.TrimSpace(jsonpathExpr)
		if strings.HasPrefix(jsonpathExpr, "{") && strings.HasSuffix(jsonpathExpr, "}") {
			jsonpathExpr = strings.TrimPrefix(jsonpathExpr, "{")
			jsonpathExpr = strings.TrimSuffix(jsonpathExpr, "}")
		}

		profile, _ := cmd.Flags().GetString("profile")
		profile_contains, _ := cmd.Flags().GetString("profile-contains")
		name_contains, _ := cmd.Flags().GetString("name-contains")
		name_not_contains, _ := cmd.Flags().GetString("name-not-contains")
		region, _ := cmd.Flags().GetString("region")
		version, _ := cmd.Flags().GetString("version")

		clusterList, err := LoadClusterList([]string{}, profile, profile_contains, name_contains, name_not_contains, region, version)
		if err != nil {
			log.Fatalf("Error loading cluster list: %v", err)
		}

		namespace, _ := cmd.Flags().GetString("namespace")
		allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err := loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		previousContext := config.CurrentContext

		// Prepare JSONPath parser - must wrap expression in braces
		jp := jsonpath.New("jsonpath")
		parseExpr := jsonpathExpr
		if !strings.HasPrefix(parseExpr, "{") {
			parseExpr = "{" + parseExpr + "}"
		}
		if err := jp.Parse(parseExpr); err != nil {
			log.Fatalf("Error parsing JSONPath expression '%s': %v", jsonpathExpr, err)
		}

		type Result struct {
			Profile     string
			Region      string
			ClusterName string
			Namespace   string
			Resource    string
			Value       string
			Error       string
		}

		results := []Result{}

		for _, clusterInfo := range clusterList {
			err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, "")
			if err != nil {
				continue
			}

			clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				clientcmd.NewDefaultClientConfigLoadingRules(),
				&clientcmd.ConfigOverrides{},
			)

			restConfig, err := clientConfig.ClientConfig()
			if err != nil {
				continue
			}

			clientset, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				continue
			}

			namespaces := []string{}
			if allNamespaces {
				nsList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
				if err != nil {
					continue
				}
				for _, ns := range nsList.Items {
					namespaces = append(namespaces, ns.Name)
				}
			} else if namespace != "" {
				namespaces = append(namespaces, namespace)
			} else {
				ns, _, err := clientConfig.Namespace()
				if err != nil {
					namespaces = append(namespaces, "default")
				} else {
					namespaces = append(namespaces, ns)
				}
			}

			for _, ns := range namespaces {
				var objects []interface{}
				var resourceNames []string

				// If resourceName is provided, get single resource, otherwise list all
				if resourceName != "" {
					// Get single resource
					obj, name, getErr := getResource(clientset, resourceType, ns, resourceName)
					if getErr != nil {
						results = append(results, Result{
							Profile:     clusterInfo.AWSProfile,
							Region:      clusterInfo.Region,
							ClusterName: clusterInfo.ClusterName,
							Namespace:   ns,
							Resource:    resourceName,
							Error:       getErr.Error(),
						})
						continue
					}
					objects = append(objects, obj)
					resourceNames = append(resourceNames, name)
				} else {
					// List all resources
					objs, names, listErr := listResources(clientset, resourceType, ns)
					if listErr != nil {
						results = append(results, Result{
							Profile:     clusterInfo.AWSProfile,
							Region:      clusterInfo.Region,
							ClusterName: clusterInfo.ClusterName,
							Namespace:   ns,
							Resource:    "all",
							Error:       listErr.Error(),
						})
						continue
					}
					objects = objs
					resourceNames = names
				}

				// Process each object
				for i, obj := range objects {
					// Execute JSONPath
					values, err := jp.FindResults(obj)
					if err != nil {
						results = append(results, Result{
							Profile:     clusterInfo.AWSProfile,
							Region:      clusterInfo.Region,
							ClusterName: clusterInfo.ClusterName,
							Namespace:   ns,
							Resource:    resourceNames[i],
							Error:       fmt.Sprintf("JSONPath error: %v", err),
						})
						continue
					}

					if len(values) == 0 || len(values[0]) == 0 {
						results = append(results, Result{
							Profile:     clusterInfo.AWSProfile,
							Region:      clusterInfo.Region,
							ClusterName: clusterInfo.ClusterName,
							Namespace:   ns,
							Resource:    resourceNames[i],
							Value:       "<not found>",
						})
						continue
					}

					// Extract value
					val := values[0][0].Interface()
					var valueStr string

					// If it's a complex object, marshal to JSON
					switch v := val.(type) {
					case string:
						valueStr = v
					case int, int32, int64, float32, float64, bool:
						valueStr = fmt.Sprintf("%v", v)
					default:
						jsonBytes, err := json.Marshal(val)
						if err != nil {
							valueStr = fmt.Sprintf("%v", val)
						} else {
							valueStr = string(jsonBytes)
						}
					}

					results = append(results, Result{
						Profile:     clusterInfo.AWSProfile,
						Region:      clusterInfo.Region,
						ClusterName: clusterInfo.ClusterName,
						Namespace:   ns,
						Resource:    resourceNames[i],
						Value:       valueStr,
					})
				}
			}
		}

		// Restore previous context
		loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
		config, err = loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		config.CurrentContext = previousContext
		if err := clientcmd.ModifyConfig(loadingRules, *config, true); err != nil {
			log.Fatalf("Error updating kubeconfig: %v", err)
		}

		// Print results
		noHeaders, _ := cmd.Flags().GetBool("no-headers")
		if !noHeaders {
			fmt.Printf("%-20s %-15s %-40s %-20s %-30s %s\n", "PROFILE", "REGION", "CLUSTER", "NAMESPACE", "NAME", "VALUE")
		}
		for _, result := range results {
			if result.Error != "" {
				fmt.Printf("%-20s %-15s %-40s %-20s %-30s ERROR: %s\n",
					result.Profile, result.Region, result.ClusterName, result.Namespace, result.Resource, result.Error)
			} else {
				fmt.Printf("%-20s %-15s %-40s %-20s %-30s %s\n",
					result.Profile, result.Region, result.ClusterName, result.Namespace, result.Resource, result.Value)
			}
		}

		saveCacheToDisk()
	},
}

// getResource retrieves a single resource by name
func getResource(clientset *kubernetes.Clientset, resourceType, namespace, name string) (interface{}, string, error) {
	switch resourceType {
	case "pod", "pods", "po":
		pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, "", err
		}
		return pod, pod.Name, nil
	case "service", "services", "svc":
		svc, err := clientset.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, "", err
		}
		return svc, svc.Name, nil
	case "deployment", "deployments", "deploy":
		deploy, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, "", err
		}
		return deploy, deploy.Name, nil
	case "daemonset", "daemonsets", "ds":
		ds, err := clientset.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, "", err
		}
		return ds, ds.Name, nil
	case "statefulset", "statefulsets", "sts":
		sts, err := clientset.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, "", err
		}
		return sts, sts.Name, nil
	case "configmap", "configmaps", "cm":
		cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, "", err
		}
		return cm, cm.Name, nil
	case "secret", "secrets":
		secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, "", err
		}
		return secret, secret.Name, nil
	default:
		return nil, "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// listResources retrieves all resources of a given type
func listResources(clientset *kubernetes.Clientset, resourceType, namespace string) ([]interface{}, []string, error) {
	var objects []interface{}
	var names []string

	switch resourceType {
	case "pod", "pods", "po":
		list, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
		}
	case "service", "services", "svc":
		list, err := clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
		}
	case "deployment", "deployments", "deploy":
		list, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
		}
	case "daemonset", "daemonsets", "ds":
		list, err := clientset.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
		}
	case "statefulset", "statefulsets", "sts":
		list, err := clientset.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
		}
	case "configmap", "configmaps", "cm":
		list, err := clientset.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
		}
	case "secret", "secrets":
		list, err := clientset.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
		}
	default:
		return nil, nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return objects, names, nil
}

func init() {
	mJsonpathCmd.Flags().StringP("profile", "p", "", "AWS profile to use")
	mJsonpathCmd.Flags().StringP("profile-contains", "q", "", "AWS profile contains string")
	mJsonpathCmd.Flags().StringP("name-contains", "c", "", "Cluster name contains string")
	mJsonpathCmd.Flags().StringP("name-not-contains", "x", "", "Cluster name does not contain string")
	mJsonpathCmd.Flags().StringP("region", "r", "", "AWS region to use")
	mJsonpathCmd.Flags().StringP("version", "v", "", "Filter by EKS version")
	mJsonpathCmd.Flags().StringP("namespace", "n", "", "Kubernetes namespace")
	mJsonpathCmd.Flags().BoolP("all-namespaces", "A", false, "Query all Kubernetes namespaces")
	mJsonpathCmd.Flags().Bool("no-headers", false, "Don't print headers")

	rootCmd.AddCommand(mJsonpathCmd)
}
