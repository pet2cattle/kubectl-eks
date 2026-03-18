package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractStatus_Pod(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name:     "running pod",
			obj:      map[string]interface{}{"status": map[string]interface{}{"phase": "Running"}},
			expected: "Running",
		},
		{
			name:     "pending pod",
			obj:      map[string]interface{}{"status": map[string]interface{}{"phase": "Pending"}},
			expected: "Pending",
		},
		{
			name:     "missing phase",
			obj:      map[string]interface{}{"status": map[string]interface{}{}},
			expected: "Unknown",
		},
		{
			name:     "no status field",
			obj:      map[string]interface{}{},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "pod")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStatus_Deployment(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "all ready",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"replicas":        int64(3),
					"readyReplicas":   int64(3),
					"updatedReplicas": int64(3),
				},
			},
			expected: "3/3",
		},
		{
			name: "partially ready",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"replicas":        int64(3),
					"readyReplicas":   int64(1),
					"updatedReplicas": int64(2),
				},
			},
			expected: "1/3 (updated: 2)",
		},
		{
			name:     "zero replicas",
			obj:      map[string]interface{}{"status": map[string]interface{}{}},
			expected: "0/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "deployment")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStatus_StatefulSet(t *testing.T) {
	obj := map[string]interface{}{
		"status": map[string]interface{}{
			"replicas":      int64(3),
			"readyReplicas": int64(2),
		},
	}
	assert.Equal(t, "2/3", ExtractStatus(obj, "statefulset"))
}

func TestExtractStatus_DaemonSet(t *testing.T) {
	obj := map[string]interface{}{
		"status": map[string]interface{}{
			"desiredNumberScheduled": int64(5),
			"currentNumberScheduled": int64(5),
			"numberReady":            int64(4),
		},
	}
	assert.Equal(t, "4/5 ready, 5 desired", ExtractStatus(obj, "daemonset"))
}

func TestExtractStatus_Service(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "ClusterIP service",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"type":      "ClusterIP",
					"clusterIP": "10.0.0.1",
				},
			},
			expected: "ClusterIP (10.0.0.1)",
		},
		{
			name: "LoadBalancer with hostname",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"type":      "LoadBalancer",
					"clusterIP": "10.0.0.1",
				},
				"status": map[string]interface{}{
					"loadBalancer": map[string]interface{}{
						"ingress": []interface{}{
							map[string]interface{}{"hostname": "my-lb.elb.amazonaws.com"},
						},
					},
				},
			},
			expected: "LoadBalancer (my-lb.elb.amazonaws.com)",
		},
		{
			name: "LoadBalancer pending",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"type":      "LoadBalancer",
					"clusterIP": "10.0.0.1",
				},
			},
			expected: "LoadBalancer (pending)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "service")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStatus_Node(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "ready node",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{"type": "Ready", "status": "True"},
					},
				},
			},
			expected: "Ready",
		},
		{
			name: "not ready node",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{"type": "Ready", "status": "False"},
					},
				},
			},
			expected: "NotReady",
		},
		{
			name:     "no conditions",
			obj:      map[string]interface{}{},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "node")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStatus_Job(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name:     "completed",
			obj:      map[string]interface{}{"status": map[string]interface{}{"succeeded": int64(1)}},
			expected: "Complete",
		},
		{
			name:     "failed",
			obj:      map[string]interface{}{"status": map[string]interface{}{"failed": int64(2), "active": int64(1)}},
			expected: "Failed (2/3)",
		},
		{
			name:     "running",
			obj:      map[string]interface{}{"status": map[string]interface{}{"active": int64(3)}},
			expected: "Running (3 active)",
		},
		{
			name:     "pending",
			obj:      map[string]interface{}{"status": map[string]interface{}{}},
			expected: "Pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "job")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStatus_PVC(t *testing.T) {
	obj := map[string]interface{}{"status": map[string]interface{}{"phase": "Bound"}}
	assert.Equal(t, "Bound", ExtractStatus(obj, "persistentvolumeclaim"))
}

func TestExtractStatus_PV(t *testing.T) {
	obj := map[string]interface{}{"status": map[string]interface{}{"phase": "Available"}}
	assert.Equal(t, "Available", ExtractStatus(obj, "persistentvolume"))
}

func TestExtractStatus_ReplicaSet(t *testing.T) {
	obj := map[string]interface{}{
		"status": map[string]interface{}{
			"replicas":      int64(3),
			"readyReplicas": int64(3),
		},
	}
	assert.Equal(t, "3/3", ExtractStatus(obj, "replicaset"))
}

func TestExtractStatus_Secret(t *testing.T) {
	obj := map[string]interface{}{
		"type": "Opaque",
		"data": map[string]interface{}{"key1": "val1", "key2": "val2"},
	}
	assert.Equal(t, "Opaque (2)", ExtractStatus(obj, "secret"))
}

func TestExtractStatus_ConfigMap(t *testing.T) {
	obj := map[string]interface{}{
		"data": map[string]interface{}{"setting1": "a", "setting2": "b", "setting3": "c"},
	}
	assert.Equal(t, "3 keys", ExtractStatus(obj, "configmap"))
}

func TestExtractStatus_Namespace(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name:     "active",
			obj:      map[string]interface{}{"status": map[string]interface{}{"phase": "Active"}},
			expected: "Active",
		},
		{
			name:     "terminating",
			obj:      map[string]interface{}{"status": map[string]interface{}{"phase": "Terminating"}},
			expected: "Terminating",
		},
		{
			name:     "no phase defaults to Active",
			obj:      map[string]interface{}{},
			expected: "Active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "namespace")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStatus_Ingress(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "with load balancer IP",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"loadBalancer": map[string]interface{}{
						"ingress": []interface{}{
							map[string]interface{}{"ip": "1.2.3.4"},
						},
					},
				},
			},
			expected: "1.2.3.4",
		},
		{
			name: "with rules, no LB",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"rules": []interface{}{
						map[string]interface{}{"host": "example.com"},
						map[string]interface{}{"host": "api.example.com"},
					},
				},
			},
			expected: "2 rule(s)",
		},
		{
			name:     "pending",
			obj:      map[string]interface{}{},
			expected: "Pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "ingress")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStatus_UnknownKind(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name:     "no status field",
			obj:      map[string]interface{}{},
			expected: "-",
		},
		{
			name: "with phase in status",
			obj: map[string]interface{}{
				"status": map[string]interface{}{"phase": "Running"},
			},
			expected: "Running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "myCustomResource")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStatus_PDB(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "with minAvailable",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"minAvailable": int64(2),
				},
				"status": map[string]interface{}{
					"currentHealthy":     int64(3),
					"disruptionsAllowed": int64(1),
					"expectedPods":       int64(3),
				},
			},
			expected: "3/2 available, 1 allowed",
		},
		{
			name: "with maxUnavailable",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"maxUnavailable": int64(1),
				},
				"status": map[string]interface{}{
					"currentHealthy":     int64(3),
					"disruptionsAllowed": int64(1),
				},
			},
			expected: "3 healthy, max 1 unavailable, 1 allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStatus(tt.obj, "poddisruptionbudget")
			assert.Equal(t, tt.expected, result)
		})
	}
}
