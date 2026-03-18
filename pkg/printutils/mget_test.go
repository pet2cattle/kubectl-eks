package printutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractAge(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected string
	}{
		{
			name:     "nil data",
			data:     nil,
			expected: "-",
		},
		{
			name:     "not a map",
			data:     "string",
			expected: "-",
		},
		{
			name:     "no metadata",
			data:     map[string]interface{}{},
			expected: "-",
		},
		{
			name: "invalid timestamp",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"creationTimestamp": "not-a-date",
				},
			},
			expected: "-",
		},
		{
			name: "valid timestamp",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"creationTimestamp": "2024-01-15T10:30:00Z",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAge(tt.data)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else {
				// For valid timestamps, just verify it's not "-"
				assert.NotEqual(t, "-", result)
			}
		})
	}
}

func TestExtractAdditionalInfo(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		kind     string
		expected string
	}{
		{
			name:     "nil data",
			data:     nil,
			kind:     "pod",
			expected: "-",
		},
		{
			name:     "unknown kind",
			data:     map[string]interface{}{},
			kind:     "CustomThing",
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAdditionalInfo(tt.data, tt.kind)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPodWideInfo(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "pod with IP and node",
			obj: map[string]interface{}{
				"status": map[string]interface{}{"podIP": "10.0.0.5"},
				"spec":   map[string]interface{}{"nodeName": "node-1"},
			},
			expected: "IP: 10.0.0.5, Node: node-1",
		},
		{
			name: "pod with IP only",
			obj: map[string]interface{}{
				"status": map[string]interface{}{"podIP": "10.0.0.5"},
			},
			expected: "IP: 10.0.0.5",
		},
		{
			name: "pod with node only",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{"nodeName": "node-1"},
			},
			expected: "Node: node-1",
		},
		{
			name:     "empty pod",
			obj:      map[string]interface{}{},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPodWideInfo(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractNodeWideInfo(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "internal and external IPs",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"addresses": []interface{}{
						map[string]interface{}{"type": "InternalIP", "address": "10.0.0.1"},
						map[string]interface{}{"type": "ExternalIP", "address": "54.1.2.3"},
					},
				},
			},
			expected: "Internal: 10.0.0.1, External: 54.1.2.3",
		},
		{
			name: "internal IP only",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"addresses": []interface{}{
						map[string]interface{}{"type": "InternalIP", "address": "10.0.0.1"},
					},
				},
			},
			expected: "Internal: 10.0.0.1",
		},
		{
			name:     "no addresses",
			obj:      map[string]interface{}{},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractNodeWideInfo(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractServiceWideInfo(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "with clusterIP and ports",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"clusterIP": "10.96.0.1",
					"ports": []interface{}{
						map[string]interface{}{"port": int64(80), "protocol": "TCP"},
						map[string]interface{}{"port": int64(443), "protocol": "TCP"},
					},
				},
			},
			expected: "ClusterIP: 10.96.0.1 | Ports: 80/TCP,443/TCP",
		},
		{
			name:     "empty service",
			obj:      map[string]interface{}{},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractServiceWideInfo(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractDeploymentWideInfo(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		contains string
		expected string
	}{
		{
			name: "with selector",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "nginx",
						},
					},
				},
			},
			contains: "app=nginx",
		},
		{
			name:     "no selector",
			obj:      map[string]interface{}{},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDeploymentWideInfo(tt.obj)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

func TestExtractPVCWideInfo(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "full PVC info",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"volumeName":       "pv-001",
					"storageClassName": "gp2",
				},
				"status": map[string]interface{}{
					"capacity": map[string]interface{}{
						"storage": "10Gi",
					},
				},
			},
			expected: "Volume: pv-001 | StorageClass: gp2 | Capacity: 10Gi",
		},
		{
			name:     "empty PVC",
			obj:      map[string]interface{}{},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPVCWideInfo(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPVWideInfo(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"capacity":         map[string]interface{}{"storage": "50Gi"},
			"storageClassName": "gp3",
		},
	}
	assert.Equal(t, "Capacity: 50Gi | StorageClass: gp3", extractPVWideInfo(obj))
}

func TestExtractIngressWideInfo(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]interface{}
		expected string
	}{
		{
			name: "with class and hosts",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"ingressClassName": "nginx",
					"rules": []interface{}{
						map[string]interface{}{"host": "example.com"},
						map[string]interface{}{"host": "api.example.com"},
					},
				},
			},
			expected: "Class: nginx | Hosts: example.com,api.example.com",
		},
		{
			name:     "empty ingress",
			obj:      map[string]interface{}{},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIngressWideInfo(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPriorityClassWideInfo(t *testing.T) {
	obj := map[string]interface{}{
		"value":            int64(1000000),
		"globalDefault":    false,
		"preemptionPolicy": "PreemptLowerPriority",
	}
	result := extractPriorityClassWideInfo(obj)
	assert.Contains(t, result, "Value: 1000000")
	assert.Contains(t, result, "GlobalDefault: false")
	assert.Contains(t, result, "Preemption: PreemptLowerPriority")
}
