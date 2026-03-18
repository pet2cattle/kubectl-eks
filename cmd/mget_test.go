package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsClusterScoped(t *testing.T) {
	clusterScoped := []string{
		"nodes",
		"namespaces",
		"persistentvolumes",
		"clusterroles",
		"clusterrolebindings",
		"storageclasses",
		"customresourcedefinitions",
		"priorityclasses",
	}
	for _, resource := range clusterScoped {
		t.Run(resource+" is cluster-scoped", func(t *testing.T) {
			assert.True(t, isClusterScoped(resource))
		})
	}

	namespacedResources := []string{
		"pods",
		"deployments",
		"services",
		"configmaps",
		"secrets",
		"statefulsets",
	}
	for _, resource := range namespacedResources {
		t.Run(resource+" is namespaced", func(t *testing.T) {
			assert.False(t, isClusterScoped(resource))
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string value",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "int value",
			input:    42,
			expected: "42",
		},
		{
			name:     "int64 value",
			input:    int64(100),
			expected: "100",
		},
		{
			name:     "float64 value",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "bool true",
			input:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			input:    false,
			expected: "false",
		},
		{
			name:     "slice value (JSON)",
			input:    []string{"a", "b"},
			expected: `["a","b"]`,
		},
		{
			name:     "map value (JSON)",
			input:    map[string]string{"key": "val"},
			expected: `{"key":"val"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
