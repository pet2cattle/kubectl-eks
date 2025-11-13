package status

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// extractStatus extracts meaningful status information based on resource kind
func ExtractStatus(obj map[string]interface{}, kind string) string {
	switch strings.ToLower(kind) {
	case "pod":
		return extractPodStatus(obj)
	case "deployment":
		return extractDeploymentStatus(obj)
	case "statefulset":
		return extractStatefulSetStatus(obj)
	case "daemonset":
		return extractDaemonSetStatus(obj)
	case "service":
		return extractServiceStatus(obj)
	case "poddisruptionbudget":
		return extractPDBStatus(obj)
	case "node":
		return extractNodeStatus(obj)
	case "persistentvolumeclaim":
		return extractPVCStatus(obj)
	case "persistentvolume":
		return extractPVStatus(obj)
	case "job":
		return extractJobStatus(obj)
	case "cronjob":
		return extractCronJobStatus(obj)
	case "replicaset":
		return extractReplicaSetStatus(obj)
	case "secret":
		return extractSecretStatus(obj)
	case "configmap":
		return extractConfigMapStatus(obj)
	case "ingress":
		return extractIngressStatus(obj)
	case "namespace":
		return extractNamespaceStatus(obj)
	default:
		// Try generic status extraction for CRDs
		return extractGenericStatus(obj)
	}
}

func extractPodStatus(obj map[string]interface{}) string {
	phase, _, _ := unstructured.NestedString(obj, "status", "phase")
	if phase == "" {
		return "Unknown"
	}
	return phase
}

func extractDeploymentStatus(obj map[string]interface{}) string {
	replicas, _, _ := unstructured.NestedInt64(obj, "status", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(obj, "status", "readyReplicas")
	updatedReplicas, _, _ := unstructured.NestedInt64(obj, "status", "updatedReplicas")

	if readyReplicas == replicas {
		return fmt.Sprintf("%d/%d", readyReplicas, replicas)
	}
	return fmt.Sprintf("%d/%d (updated: %d)", readyReplicas, replicas, updatedReplicas)
}

func extractStatefulSetStatus(obj map[string]interface{}) string {
	replicas, _, _ := unstructured.NestedInt64(obj, "status", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(obj, "status", "readyReplicas")
	return fmt.Sprintf("%d/%d", readyReplicas, replicas)
}

func extractDaemonSetStatus(obj map[string]interface{}) string {
	desired, _, _ := unstructured.NestedInt64(obj, "status", "desiredNumberScheduled")
	current, _, _ := unstructured.NestedInt64(obj, "status", "currentNumberScheduled")
	ready, _, _ := unstructured.NestedInt64(obj, "status", "numberReady")
	return fmt.Sprintf("%d/%d ready, %d desired", ready, current, desired)
}

func extractServiceStatus(obj map[string]interface{}) string {
	svcType, _, _ := unstructured.NestedString(obj, "spec", "type")
	clusterIP, _, _ := unstructured.NestedString(obj, "spec", "clusterIP")

	if svcType == "LoadBalancer" {
		ingress, found, _ := unstructured.NestedSlice(obj, "status", "loadBalancer", "ingress")
		if found && len(ingress) > 0 {
			if ingressMap, ok := ingress[0].(map[string]interface{}); ok {
				if ip, ok := ingressMap["ip"].(string); ok && ip != "" {
					return fmt.Sprintf("%s (%s)", svcType, ip)
				}
				if hostname, ok := ingressMap["hostname"].(string); ok && hostname != "" {
					return fmt.Sprintf("%s (%s)", svcType, hostname)
				}
			}
		}
		return fmt.Sprintf("%s (pending)", svcType)
	}

	if clusterIP != "" {
		return fmt.Sprintf("%s (%s)", svcType, clusterIP)
	}
	return svcType
}

func extractPDBStatus(obj map[string]interface{}) string {
	currentHealthy, _, _ := unstructured.NestedInt64(obj, "status", "currentHealthy")
	desiredHealthy, _, _ := unstructured.NestedInt64(obj, "status", "desiredHealthy")
	disruptionsAllowed, _, _ := unstructured.NestedInt64(obj, "status", "disruptionsAllowed")

	return fmt.Sprintf("%d/%d healthy (allowed: %d)", currentHealthy, desiredHealthy, disruptionsAllowed)
}

func extractNodeStatus(obj map[string]interface{}) string {
	conditions, found, _ := unstructured.NestedSlice(obj, "status", "conditions")
	if !found {
		return "Unknown"
	}

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}
		if condType, ok := condMap["type"].(string); ok && condType == "Ready" {
			if status, ok := condMap["status"].(string); ok {
				if status == "True" {
					return "Ready"
				}
				return "NotReady"
			}
		}
	}
	return "Unknown"
}

func extractPVCStatus(obj map[string]interface{}) string {
	phase, _, _ := unstructured.NestedString(obj, "status", "phase")
	if phase == "" {
		return "Unknown"
	}
	return phase
}

func extractPVStatus(obj map[string]interface{}) string {
	phase, _, _ := unstructured.NestedString(obj, "status", "phase")
	if phase == "" {
		return "Unknown"
	}
	return phase
}

func extractJobStatus(obj map[string]interface{}) string {
	succeeded, _, _ := unstructured.NestedInt64(obj, "status", "succeeded")
	active, _, _ := unstructured.NestedInt64(obj, "status", "active")
	failed, _, _ := unstructured.NestedInt64(obj, "status", "failed")

	if succeeded > 0 {
		return "Complete"
	}
	if failed > 0 {
		return fmt.Sprintf("Failed (%d/%d)", failed, active+failed)
	}
	if active > 0 {
		return fmt.Sprintf("Running (%d active)", active)
	}
	return "Pending"
}

func extractCronJobStatus(obj map[string]interface{}) string {
	active, found, _ := unstructured.NestedSlice(obj, "status", "active")
	if found && len(active) > 0 {
		return fmt.Sprintf("%d active", len(active))
	}

	lastScheduleTime, found, _ := unstructured.NestedString(obj, "status", "lastScheduleTime")
	if found && lastScheduleTime != "" {
		t, err := time.Parse(time.RFC3339, lastScheduleTime)
		if err == nil {
			return fmt.Sprintf("Last: %s", formatAge(t))
		}
	}
	return "No runs"
}

func extractReplicaSetStatus(obj map[string]interface{}) string {
	replicas, _, _ := unstructured.NestedInt64(obj, "status", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(obj, "status", "readyReplicas")
	return fmt.Sprintf("%d/%d", readyReplicas, replicas)
}

func extractSecretStatus(obj map[string]interface{}) string {
	secretType, _, _ := unstructured.NestedString(obj, "type")
	data, found, _ := unstructured.NestedMap(obj, "data")

	if found {
		return fmt.Sprintf("%s (%d)", secretType, len(data))
	}
	return secretType
}

func extractConfigMapStatus(obj map[string]interface{}) string {
	data, found, _ := unstructured.NestedMap(obj, "data")
	if found {
		return fmt.Sprintf("%d keys", len(data))
	}
	return "0 keys"
}

func extractIngressStatus(obj map[string]interface{}) string {
	ingresses, found, _ := unstructured.NestedSlice(obj, "status", "loadBalancer", "ingress")
	if found && len(ingresses) > 0 {
		addresses := []string{}
		for _, ing := range ingresses {
			if ingMap, ok := ing.(map[string]interface{}); ok {
				if ip, ok := ingMap["ip"].(string); ok && ip != "" {
					addresses = append(addresses, ip)
				} else if hostname, ok := ingMap["hostname"].(string); ok && hostname != "" {
					addresses = append(addresses, hostname)
				}
			}
		}
		if len(addresses) > 0 {
			return strings.Join(addresses, ",")
		}
	}

	// Check for rules
	rules, found, _ := unstructured.NestedSlice(obj, "spec", "rules")
	if found {
		return fmt.Sprintf("%d rule(s)", len(rules))
	}

	return "Pending"
}

func extractNamespaceStatus(obj map[string]interface{}) string {
	phase, _, _ := unstructured.NestedString(obj, "status", "phase")
	if phase == "" {
		return "Active"
	}
	return phase
}

// extractGenericStatus tries to extract status from unknown/CRD resources
func extractGenericStatus(obj map[string]interface{}) string {
	// Try common status patterns
	status, found, _ := unstructured.NestedMap(obj, "status")
	if !found {
		return "-"
	}

	// Check for common status fields
	if phase, ok := status["phase"].(string); ok && phase != "" {
		return phase
	}

	if state, ok := status["state"].(string); ok && state != "" {
		return state
	}

	// Check for conditions (common in CRDs)
	if conditions, ok := status["conditions"].([]interface{}); ok && len(conditions) > 0 {
		if condMap, ok := conditions[len(conditions)-1].(map[string]interface{}); ok {
			if condType, ok := condMap["type"].(string); ok {
				if condStatus, ok := condMap["status"].(string); ok {
					if condStatus == "True" {
						return condType
					}
					return fmt.Sprintf("Not%s", condType)
				}
			}
		}
	}

	// Check for ready field
	if ready, ok := status["ready"].(bool); ok {
		if ready {
			return "Ready"
		}
		return "NotReady"
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
