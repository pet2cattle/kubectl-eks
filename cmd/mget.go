package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/pet2cattle/kubectl-eks/pkg/eks"
	"github.com/pet2cattle/kubectl-eks/pkg/k8s"
	"github.com/pet2cattle/kubectl-eks/pkg/printutils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/jsonpath"
)

var mGetCmd = &cobra.Command{
	Use:   "mget [resource-type] [resource-name]",
	Short: "Get resources from multiple clusters",
	Long: `Get Kubernetes resources from all clusters that match a filter.
Similar to 'kubectl get' but works across multiple EKS clusters.

Supports output formats:
  -o wide          Additional details
  -o json          JSON output
  -o yaml          YAML output
  -o jsonpath=...  Extract specific fields using JSONPath`,
	Example: `  # List all pods across clusters
  kubectl eks mget pods

  # List pods in specific namespace
  kubectl eks mget pods -n kube-system

  # Get a specific deployment
  kubectl eks mget deployment my-app

  # Extract specific fields with JSONPath
  kubectl eks mget pods -o jsonpath='{.spec.dnsPolicy}'
  
  # List deployments with additional details
  kubectl eks mget deployments -o wide
  
  # Filter clusters and resources
  kubectl eks mget pods --name-contains prod --resource-starts-with nginx`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		resourceType := args[0]
		var resourceName string
		if len(args) > 1 {
			resourceName = args[1]
		}

		// Get flags
		profile, _ := cmd.Flags().GetString("profile")
		profileContains, _ := cmd.Flags().GetString("profile-contains")
		nameContains, _ := cmd.Flags().GetString("name-contains")
		nameNotContains, _ := cmd.Flags().GetString("name-not-contains")
		region, _ := cmd.Flags().GetString("region")
		version, _ := cmd.Flags().GetString("version")
		namespace, _ := cmd.Flags().GetString("namespace")
		allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")
		output, _ := cmd.Flags().GetString("output")
		startsWith, _ := cmd.Flags().GetString("resource-starts-with")
		noHeaders, _ := cmd.Flags().GetBool("no-headers")

		// Load cluster list
		clusterList, err := LoadClusterList([]string{}, profile, profileContains, nameContains, nameNotContains, region, version)
		if err != nil {
			log.Fatalf("Error loading cluster list: %v", err)
		}

		// Save and restore context
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, err := loadingRules.Load()
		if err != nil {
			log.Fatalf("Error loading kubeconfig: %v", err)
		}
		previousContext := config.CurrentContext
		defer func() {
			config.CurrentContext = previousContext
			clientcmd.ModifyConfig(loadingRules, *config, true)
		}()

		// Check if JSONPath output
		if strings.HasPrefix(output, "jsonpath=") {
			jsonpathExpr := strings.TrimPrefix(output, "jsonpath=")
			runJsonPathQuery(clusterList, resourceType, resourceName, jsonpathExpr, namespace, allNamespaces, startsWith, noHeaders)
		} else if resourceType == "pods" || resourceType == "pod" || resourceType == "po" {
			// Use existing pod listing functionality
			runPodListing(clusterList, namespace, allNamespaces, noHeaders)
		} else {
			// Generic resource listing
			runGenericListing(clusterList, resourceType, resourceName, namespace, allNamespaces, startsWith, output, noHeaders)
		}

		saveCacheToDisk()
	},
}

func runPodListing(clusterList []data.ClusterInfo, namespace string, allNamespaces bool, noHeaders bool) {
	k8SClusterPodList := []k8s.K8SClusterPodList{}

	for _, clusterInfo := range clusterList {
		err := eks.UpdateKubeConfig(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, "")
		if err != nil {
			log.Printf("Warning: Failed to update kubeconfig for cluster %s: %v", clusterInfo.ClusterName, err)
			continue
		}

		k8sPodList, err := k8s.GetPods(clusterInfo.AWSProfile, clusterInfo.Region, clusterInfo.ClusterName, clusterInfo.Arn, clusterInfo.Version, namespace, allNamespaces)
		if err != nil {
			log.Printf("Warning: Failed to get pods from cluster %s: %v", clusterInfo.ClusterName, err)
			continue
		}
		k8SClusterPodList = append(k8SClusterPodList, *k8sPodList)
	}

	printutils.PrintMultiGetPods(noHeaders, k8SClusterPodList...)
}

func runGenericListing(clusterList []data.ClusterInfo, resourceType, resourceName, namespace string, allNamespaces bool, startsWith, output string, noHeaders bool) {
	results := []data.ResourceResult{}

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

		namespaces := getNamespaces(clientset, clientConfig, namespace, allNamespaces)

		for _, ns := range namespaces {
			if resourceName != "" {
				// Get single resource
				obj, name, kind, getErr := getResource(clientset, resourceType, ns, resourceName)
				if getErr != nil {
					results = append(results, data.ResourceResult{
						Profile:     clusterInfo.AWSProfile,
						Region:      clusterInfo.Region,
						ClusterName: clusterInfo.ClusterName,
						Namespace:   ns,
						Name:        resourceName,
						Error:       getErr.Error(),
					})
					continue
				}
				results = append(results, data.ResourceResult{
					Profile:     clusterInfo.AWSProfile,
					Region:      clusterInfo.Region,
					ClusterName: clusterInfo.ClusterName,
					Namespace:   ns,
					Name:        name,
					Kind:        kind,
					Data:        obj,
				})
			} else {
				// List resources
				objs, names, kinds, listErr := listResources(clientset, resourceType, ns)
				if listErr != nil {
					results = append(results, data.ResourceResult{
						Profile:     clusterInfo.AWSProfile,
						Region:      clusterInfo.Region,
						ClusterName: clusterInfo.ClusterName,
						Namespace:   ns,
						Error:       listErr.Error(),
					})
					continue
				}

				// Apply startsWith filter
				for i, name := range names {
					if startsWith != "" && !strings.HasPrefix(name, startsWith) {
						continue
					}
					results = append(results, data.ResourceResult{
						Profile:     clusterInfo.AWSProfile,
						Region:      clusterInfo.Region,
						ClusterName: clusterInfo.ClusterName,
						Namespace:   ns,
						Name:        name,
						Kind:        kinds[i],
						Data:        objs[i],
					})
				}
			}
		}
	}

	// Print results based on output format
	printutils.PrintGenericResults(results, output, noHeaders)
}

func runJsonPathQuery(clusterList []data.ClusterInfo, resourceType, resourceName, jsonpathExpr, namespace string, allNamespaces bool, startsWith string, noHeaders bool) {
	// Normalize JSONPath expression
	jsonpathExpr = strings.TrimSpace(jsonpathExpr)
	if strings.HasPrefix(jsonpathExpr, "{") && strings.HasSuffix(jsonpathExpr, "}") {
		jsonpathExpr = strings.TrimPrefix(jsonpathExpr, "{")
		jsonpathExpr = strings.TrimSuffix(jsonpathExpr, "}")
	}

	// Prepare JSONPath parser
	jp := jsonpath.New("jsonpath")
	parseExpr := jsonpathExpr
	if !strings.HasPrefix(parseExpr, "{") {
		parseExpr = "{" + parseExpr + "}"
	}
	if err := jp.Parse(parseExpr); err != nil {
		log.Fatalf("Error parsing JSONPath expression '%s': %v", jsonpathExpr, err)
	}

	results := []data.JsonPathResult{}

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

		namespaces := getNamespaces(clientset, clientConfig, namespace, allNamespaces)

		for _, ns := range namespaces {
			var objects []interface{}
			var resourceNames []string

			if resourceName != "" {
				obj, name, _, getErr := getResource(clientset, resourceType, ns, resourceName)
				if getErr != nil {
					results = append(results, data.JsonPathResult{
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
				objs, names, _, listErr := listResources(clientset, resourceType, ns)
				if listErr != nil {
					results = append(results, data.JsonPathResult{
						Profile:     clusterInfo.AWSProfile,
						Region:      clusterInfo.Region,
						ClusterName: clusterInfo.ClusterName,
						Namespace:   ns,
						Resource:    "all",
						Error:       listErr.Error(),
					})
					continue
				}

				// Filter by starts-with
				for i, name := range names {
					if startsWith == "" || strings.HasPrefix(name, startsWith) {
						objects = append(objects, objs[i])
						resourceNames = append(resourceNames, name)
					}
				}
			}

			// Execute JSONPath on each object
			for i, obj := range objects {
				values, err := jp.FindResults(obj)
				if err != nil {
					results = append(results, data.JsonPathResult{
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
					results = append(results, data.JsonPathResult{
						Profile:     clusterInfo.AWSProfile,
						Region:      clusterInfo.Region,
						ClusterName: clusterInfo.ClusterName,
						Namespace:   ns,
						Resource:    resourceNames[i],
						Value:       "<not found>",
					})
					continue
				}

				val := values[0][0].Interface()
				valueStr := formatValue(val)

				results = append(results, data.JsonPathResult{
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

	printutils.PrintJsonPathResults(noHeaders, results)
}

func getNamespaces(clientset *kubernetes.Clientset, clientConfig clientcmd.ClientConfig, namespace string, allNamespaces bool) []string {
	namespaces := []string{}
	if allNamespaces {
		nsList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, ns := range nsList.Items {
				namespaces = append(namespaces, ns.Name)
			}
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
	return namespaces
}

// Add to getResource function:
func getResource(clientset *kubernetes.Clientset, resourceType, namespace, name string) (interface{}, string, string, error) {
	switch resourceType {
	case "pod", "pods", "po":
		pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
		return pod, pod.Name, "Pod", err
	case "service", "services", "svc":
		svc, err := clientset.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
		return svc, svc.Name, "Service", err
	case "deployment", "deployments", "deploy":
		deploy, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
		return deploy, deploy.Name, "Deployment", err
	case "daemonset", "daemonsets", "ds":
		ds, err := clientset.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		return ds, ds.Name, "DaemonSet", err
	case "statefulset", "statefulsets", "sts":
		sts, err := clientset.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		return sts, sts.Name, "StatefulSet", err
	case "configmap", "configmaps", "cm":
		cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
		return cm, cm.Name, "ConfigMap", err
	case "secret", "secrets":
		secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		return secret, secret.Name, "Secret", err
	case "poddisruptionbudget", "poddisruptionbudgets", "pdb":
		pdb, err := clientset.PolicyV1().PodDisruptionBudgets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		return pdb, pdb.Name, "PodDisruptionBudget", err
	default:
		return nil, "", "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// Add to listResources function:
func listResources(clientset *kubernetes.Clientset, resourceType, namespace string) ([]interface{}, []string, []string, error) {
	var objects []interface{}
	var names []string
	var kinds []string

	switch resourceType {
	case "pod", "pods", "po":
		list, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
			kinds = append(kinds, "Pod")
		}
	case "service", "services", "svc":
		list, err := clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
			kinds = append(kinds, "Service")
		}
	case "deployment", "deployments", "deploy":
		list, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
			kinds = append(kinds, "Deployment")
		}
	case "daemonset", "daemonsets", "ds":
		list, err := clientset.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
			kinds = append(kinds, "DaemonSet")
		}
	case "statefulset", "statefulsets", "sts":
		list, err := clientset.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
			kinds = append(kinds, "StatefulSet")
		}
	case "configmap", "configmaps", "cm":
		list, err := clientset.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
			kinds = append(kinds, "ConfigMap")
		}
	case "secret", "secrets":
		list, err := clientset.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
			kinds = append(kinds, "Secret")
		}
	case "poddisruptionbudget", "poddisruptionbudgets", "pdb":
		list, err := clientset.PolicyV1().PodDisruptionBudgets(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range list.Items {
			objects = append(objects, &list.Items[i])
			names = append(names, list.Items[i].Name)
			kinds = append(kinds, "PodDisruptionBudget")
		}
	default:
		return nil, nil, nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return objects, names, kinds, nil
}

func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int, int32, int64, float32, float64, bool:
		return fmt.Sprintf("%v", v)
	default:
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(jsonBytes)
	}
}

func init() {
	mGetCmd.Flags().StringP("profile", "p", "", "AWS profile to use")
	mGetCmd.Flags().StringP("profile-contains", "q", "", "AWS profile contains string")
	mGetCmd.Flags().StringP("name-contains", "c", "", "Cluster name contains string")
	mGetCmd.Flags().StringP("name-not-contains", "x", "", "Cluster name does not contain string")
	mGetCmd.Flags().StringP("region", "r", "", "AWS region to use")
	mGetCmd.Flags().StringP("version", "v", "", "Filter by EKS version")
	mGetCmd.Flags().StringP("namespace", "n", "", "Kubernetes namespace")
	mGetCmd.Flags().BoolP("all-namespaces", "A", false, "Query all Kubernetes namespaces")
	mGetCmd.Flags().StringP("output", "o", "", "Output format: wide|json|yaml|jsonpath=...")
	mGetCmd.Flags().StringP("resource-starts-with", "w", "", "Filter resources that start with this string")
	mGetCmd.Flags().Bool("no-headers", false, "Don't print headers")

	rootCmd.AddCommand(mGetCmd)
}
