package k8s

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type NodeInfo struct {
	Name         string
	InstanceType string
	Compute      string
	ManagedBy    string
	Created      time.Time
	Status       string
}

func GetNodes(configFlags *genericclioptions.ConfigFlags) ([]NodeInfo, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var nodeList []NodeInfo
	for _, node := range nodes.Items {
		labels := node.Labels

		// Instance type
		instanceType := labels["node.kubernetes.io/instance-type"]
		if instanceType == "" {
			instanceType = labels["beta.kubernetes.io/instance-type"]
		}

		// Compute type (EC2 or Fargate)
		compute := "EC2"
		if labels["eks.amazonaws.com/compute-type"] == "fargate" {
			compute = "Fargate"
		}

		// Managed by (nodegroup, Fargate profile, or Karpenter)
		managedBy := ""
		if ng := labels["eks.amazonaws.com/nodegroup"]; ng != "" {
			managedBy = "Nodegroup: " + ng
		} else if fp := labels["eks.amazonaws.com/fargate-profile"]; fp != "" {
			managedBy = "Fargate: " + fp
		} else if np := labels["karpenter.sh/nodepool"]; np != "" {
			managedBy = "Karpenter: " + np
		} else if labels["karpenter.sh/provisioner-name"] != "" {
			// Older Karpenter versions used provisioner-name
			managedBy = "Karpenter: " + labels["karpenter.sh/provisioner-name"]
		} else {
			managedBy = "AWS"
		}

		// Node status
		status := "Unknown"
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" {
				if cond.Status == "True" {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		nodeList = append(nodeList, NodeInfo{
			Name:         node.Name,
			InstanceType: instanceType,
			Compute:      compute,
			ManagedBy:    managedBy,
			Created:      node.CreationTimestamp.Time,
			Status:       status,
		})
	}

	return nodeList, nil
}
