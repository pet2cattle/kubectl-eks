package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
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
