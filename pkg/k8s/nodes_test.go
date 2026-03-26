package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetNodeStatus(t *testing.T) {
	tests := []struct {
		name string
		node corev1.Node
		want string
	}{
		{
			name: "ready node",
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					}},
				},
			},
			want: "Ready",
		},
		{
			name: "ready unschedulable node",
			node: corev1.Node{
				Spec: corev1.NodeSpec{Unschedulable: true},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					}},
				},
			},
			want: "Ready,SchedulingDisabled",
		},
		{
			name: "not ready unschedulable node",
			node: corev1.Node{
				Spec: corev1.NodeSpec{Unschedulable: true},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionFalse,
					}},
				},
			},
			want: "NotReady,SchedulingDisabled",
		},
		{
			name: "unknown unschedulable node",
			node: corev1.Node{
				Spec: corev1.NodeSpec{Unschedulable: true},
			},
			want: "Unknown,SchedulingDisabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNodeStatus(tt.node)
			if got != tt.want {
				t.Fatalf("getNodeStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetNodeConditionStatus(t *testing.T) {
	node := corev1.Node{
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionFalse},
				{Type: corev1.NodeDiskPressure, Status: corev1.ConditionTrue},
				{Type: corev1.NodePIDPressure, Status: corev1.ConditionUnknown},
			},
		},
	}

	if got := getNodeConditionStatus(node, corev1.NodeMemoryPressure); got != "False" {
		t.Fatalf("memory pressure = %q, want %q", got, "False")
	}
	if got := getNodeConditionStatus(node, corev1.NodeDiskPressure); got != "True" {
		t.Fatalf("disk pressure = %q, want %q", got, "True")
	}
	if got := getNodeConditionStatus(node, corev1.NodePIDPressure); got != "Unknown" {
		t.Fatalf("pid pressure = %q, want %q", got, "Unknown")
	}
	if got := getNodeConditionStatus(node, corev1.NodeNetworkUnavailable); got != "Unknown" {
		t.Fatalf("network unavailable = %q, want %q", got, "Unknown")
	}
}

func TestGetNodeResourceQuantity(t *testing.T) {
	node := corev1.Node{
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("16Gi"),
				corev1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3500m"),
				corev1.ResourceMemory: resource.MustParse("14Gi"),
				corev1.ResourcePods:   resource.MustParse("100"),
			},
		},
	}

	if got := getNodeResourceQuantity(node, true, corev1.ResourceCPU); got != "4" {
		t.Fatalf("cpu capacity = %q, want %q", got, "4")
	}
	if got := getNodeResourceQuantity(node, false, corev1.ResourceCPU); got != "3500m" {
		t.Fatalf("cpu allocatable = %q, want %q", got, "3500m")
	}
	if got := getNodeResourceQuantity(node, true, corev1.ResourceMemory); got != "16Gi" {
		t.Fatalf("memory capacity = %q, want %q", got, "16Gi")
	}
	if got := getNodeResourceQuantity(node, false, corev1.ResourceMemory); got != "14Gi" {
		t.Fatalf("memory allocatable = %q, want %q", got, "14Gi")
	}
	if got := getNodeResourceQuantity(node, true, corev1.ResourcePods); got != "110" {
		t.Fatalf("pods capacity = %q, want %q", got, "110")
	}
	if got := getNodeResourceQuantity(node, false, corev1.ResourcePods); got != "100" {
		t.Fatalf("pods allocatable = %q, want %q", got, "100")
	}
	if got := getNodeResourceQuantity(node, true, corev1.ResourceEphemeralStorage); got != "-" {
		t.Fatalf("missing resource = %q, want %q", got, "-")
	}
}

func TestGetNodeResourceUsed(t *testing.T) {
	node := corev1.Node{
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("16Gi"),
				corev1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3500m"),
				corev1.ResourceMemory: resource.MustParse("14Gi"),
				corev1.ResourcePods:   resource.MustParse("100"),
			},
		},
	}

	if got := getNodeResourceUsed(node, corev1.ResourceCPU); got != "500m" {
		t.Fatalf("cpu used = %q, want %q", got, "500m")
	}
	if got := getNodeResourceUsed(node, corev1.ResourceMemory); got != "2Gi" {
		t.Fatalf("memory used = %q, want %q", got, "2Gi")
	}
	if got := getNodeResourceUsed(node, corev1.ResourcePods); got != "10" {
		t.Fatalf("pods used = %q, want %q", got, "10")
	}
	if got := getNodeResourceUsed(node, corev1.ResourceEphemeralStorage); got != "-" {
		t.Fatalf("missing resource used = %q, want %q", got, "-")
	}
}
