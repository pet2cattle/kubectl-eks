package printutils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func PrintGenericResults(results []data.ResourceResult, output string, noHeaders bool) {
	if len(results) == 0 {
		return
	}

	// Handle JSON output
	if output == "json" {
		jsonBytes, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(jsonBytes))
		return
	}

	// Handle YAML output
	if output == "yaml" {
		for i, result := range results {
			if result.Error != "" {
				fmt.Printf("# Error for %s/%s in %s: %s\n", result.Namespace, result.Name, result.ClusterName, result.Error)
				continue
			}
			yamlBytes, _ := yaml.Marshal(result.Data)
			fmt.Println("---")
			fmt.Print(string(yamlBytes))
			if i < len(results)-1 {
				fmt.Println()
			}
		}
		return
	}

	// Table output (default and wide)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print headers
	if !noHeaders {
		if output == "wide" {
			fmt.Fprintln(w, "AWS PROFILE\tAWS REGION\tCLUSTER NAME\tNAMESPACE\tKIND\tNAME\tSTATUS\tAGE\tADDITIONAL INFO")
		} else {
			fmt.Fprintln(w, "AWS PROFILE\tAWS REGION\tCLUSTER NAME\tNAMESPACE\tKIND\tNAME\tSTATUS")
		}
	}

	// Print rows
	for _, result := range results {
		namespace := result.Namespace
		if namespace == "" {
			namespace = "-" // For cluster-scoped resources
		}

		if result.Error != "" {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\tERROR: %s\n",
				result.Profile,
				result.Region,
				result.ClusterName,
				namespace,
				result.Kind,
				result.Name,
				result.Error,
			)
			continue
		}

		// Use the Status field that was already calculated
		status := result.Status
		if status == "" {
			status = "-"
		}

		if output == "wide" {
			age := extractAge(result.Data)
			additionalInfo := extractAdditionalInfo(result.Data, result.Kind)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				result.Profile,
				result.Region,
				result.ClusterName,
				namespace,
				result.Kind,
				result.Name,
				status,
				age,
				additionalInfo,
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				result.Profile,
				result.Region,
				result.ClusterName,
				namespace,
				result.Kind,
				result.Name,
				status,
			)
		}
	}

	w.Flush()
}

// extractAge extracts the age of a resource from its creation timestamp
func extractAge(data interface{}) string {
	obj, ok := data.(map[string]interface{})
	if !ok {
		return "-"
	}

	creationTimestamp, found, _ := unstructured.NestedString(obj, "metadata", "creationTimestamp")
	if !found || creationTimestamp == "" {
		return "-"
	}

	t, err := time.Parse(time.RFC3339, creationTimestamp)
	if err != nil {
		return "-"
	}

	return formatAge(t)
}

// extractAdditionalInfo extracts additional info for wide output based on resource kind
func extractAdditionalInfo(data interface{}, kind string) string {
	obj, ok := data.(map[string]interface{})
	if !ok {
		return "-"
	}

	switch strings.ToLower(kind) {
	case "pod":
		return extractPodWideInfo(obj)
	case "node":
		return extractNodeWideInfo(obj)
	case "service":
		return extractServiceWideInfo(obj)
	case "deployment":
		return extractDeploymentWideInfo(obj)
	case "persistentvolumeclaim":
		return extractPVCWideInfo(obj)
	case "persistentvolume":
		return extractPVWideInfo(obj)
	case "ingress":
		return extractIngressWideInfo(obj)
	default:
		return "-"
	}
}

func extractPodWideInfo(obj map[string]interface{}) string {
	ip, _, _ := unstructured.NestedString(obj, "status", "podIP")
	node, _, _ := unstructured.NestedString(obj, "spec", "nodeName")

	if ip != "" && node != "" {
		return fmt.Sprintf("IP: %s, Node: %s", ip, node)
	}
	if ip != "" {
		return fmt.Sprintf("IP: %s", ip)
	}
	if node != "" {
		return fmt.Sprintf("Node: %s", node)
	}
	return "-"
}

func extractNodeWideInfo(obj map[string]interface{}) string {
	// Get node addresses
	addresses, found, _ := unstructured.NestedSlice(obj, "status", "addresses")
	if !found {
		return "-"
	}

	var internalIP, externalIP string
	for _, addr := range addresses {
		addrMap, ok := addr.(map[string]interface{})
		if !ok {
			continue
		}
		addrType, _ := addrMap["type"].(string)
		addrValue, _ := addrMap["address"].(string)

		if addrType == "InternalIP" {
			internalIP = addrValue
		} else if addrType == "ExternalIP" {
			externalIP = addrValue
		}
	}

	if externalIP != "" && internalIP != "" {
		return fmt.Sprintf("Internal: %s, External: %s", internalIP, externalIP)
	}
	if internalIP != "" {
		return fmt.Sprintf("Internal: %s", internalIP)
	}
	return "-"
}

func extractServiceWideInfo(obj map[string]interface{}) string {
	clusterIP, _, _ := unstructured.NestedString(obj, "spec", "clusterIP")
	externalIPs, _, _ := unstructured.NestedStringSlice(obj, "spec", "externalIPs")
	ports, found, _ := unstructured.NestedSlice(obj, "spec", "ports")

	info := []string{}
	if clusterIP != "" {
		info = append(info, fmt.Sprintf("ClusterIP: %s", clusterIP))
	}
	if len(externalIPs) > 0 {
		info = append(info, fmt.Sprintf("ExternalIP: %s", strings.Join(externalIPs, ",")))
	}
	if found && len(ports) > 0 {
		portStrs := []string{}
		for _, port := range ports {
			portMap, ok := port.(map[string]interface{})
			if !ok {
				continue
			}
			portNum, _ := portMap["port"].(int64)
			protocol, _ := portMap["protocol"].(string)
			if portNum > 0 {
				portStrs = append(portStrs, fmt.Sprintf("%d/%s", portNum, protocol))
			}
		}
		if len(portStrs) > 0 {
			info = append(info, fmt.Sprintf("Ports: %s", strings.Join(portStrs, ",")))
		}
	}

	if len(info) > 0 {
		return strings.Join(info, " | ")
	}
	return "-"
}

func extractDeploymentWideInfo(obj map[string]interface{}) string {
	selector, found, _ := unstructured.NestedMap(obj, "spec", "selector", "matchLabels")
	if !found {
		return "-"
	}

	labels := []string{}
	for k, v := range selector {
		if strVal, ok := v.(string); ok {
			labels = append(labels, fmt.Sprintf("%s=%s", k, strVal))
		}
	}

	if len(labels) > 0 {
		return fmt.Sprintf("Selector: %s", strings.Join(labels, ","))
	}
	return "-"
}

func extractPVCWideInfo(obj map[string]interface{}) string {
	volumeName, _, _ := unstructured.NestedString(obj, "spec", "volumeName")
	storageClass, _, _ := unstructured.NestedString(obj, "spec", "storageClassName")
	capacity, found, _ := unstructured.NestedMap(obj, "status", "capacity")

	info := []string{}
	if volumeName != "" {
		info = append(info, fmt.Sprintf("Volume: %s", volumeName))
	}
	if storageClass != "" {
		info = append(info, fmt.Sprintf("StorageClass: %s", storageClass))
	}
	if found {
		if storage, ok := capacity["storage"].(string); ok {
			info = append(info, fmt.Sprintf("Capacity: %s", storage))
		}
	}

	if len(info) > 0 {
		return strings.Join(info, " | ")
	}
	return "-"
}

func extractPVWideInfo(obj map[string]interface{}) string {
	capacity, found, _ := unstructured.NestedMap(obj, "spec", "capacity")
	storageClass, _, _ := unstructured.NestedString(obj, "spec", "storageClassName")

	info := []string{}
	if found {
		if storage, ok := capacity["storage"].(string); ok {
			info = append(info, fmt.Sprintf("Capacity: %s", storage))
		}
	}
	if storageClass != "" {
		info = append(info, fmt.Sprintf("StorageClass: %s", storageClass))
	}

	if len(info) > 0 {
		return strings.Join(info, " | ")
	}
	return "-"
}

func extractIngressWideInfo(obj map[string]interface{}) string {
	ingressClass, _, _ := unstructured.NestedString(obj, "spec", "ingressClassName")
	rules, found, _ := unstructured.NestedSlice(obj, "spec", "rules")

	info := []string{}
	if ingressClass != "" {
		info = append(info, fmt.Sprintf("Class: %s", ingressClass))
	}

	if found && len(rules) > 0 {
		hosts := []string{}
		for _, rule := range rules {
			ruleMap, ok := rule.(map[string]interface{})
			if !ok {
				continue
			}
			if host, ok := ruleMap["host"].(string); ok && host != "" {
				hosts = append(hosts, host)
			}
		}
		if len(hosts) > 0 {
			info = append(info, fmt.Sprintf("Hosts: %s", strings.Join(hosts, ",")))
		}
	}

	if len(info) > 0 {
		return strings.Join(info, " | ")
	}
	return "-"
}

func formatAge(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	return fmt.Sprintf("%dd", int(duration.Hours()/24))
}
