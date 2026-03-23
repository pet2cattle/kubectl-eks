package cmd

import (
	"testing"

	"github.com/jordiprats/kubectl-eks/pkg/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectClusterByAge_Oldest(t *testing.T) {
	clusters := []data.ClusterInfo{
		{ClusterName: "new", CreatedAt: "2025-06-15 10:00:00"},
		{ClusterName: "old", CreatedAt: "2024-01-01 08:00:00"},
		{ClusterName: "mid", CreatedAt: "2024-09-20 12:00:00"},
	}

	selected, err := selectClusterByAge(clusters, true)
	require.NoError(t, err)
	assert.Equal(t, "old", selected.ClusterName)
}

func TestSelectClusterByAge_Newest(t *testing.T) {
	clusters := []data.ClusterInfo{
		{ClusterName: "new", CreatedAt: "2025-06-15 10:00:00"},
		{ClusterName: "old", CreatedAt: "2024-01-01 08:00:00"},
		{ClusterName: "mid", CreatedAt: "2024-09-20 12:00:00"},
	}

	selected, err := selectClusterByAge(clusters, false)
	require.NoError(t, err)
	assert.Equal(t, "new", selected.ClusterName)
}

func TestSelectClusterByAge_SingleCluster(t *testing.T) {
	clusters := []data.ClusterInfo{
		{ClusterName: "only", CreatedAt: "2025-01-01 00:00:00"},
	}

	oldest, err := selectClusterByAge(clusters, true)
	require.NoError(t, err)
	assert.Equal(t, "only", oldest.ClusterName)

	newest, err := selectClusterByAge(clusters, false)
	require.NoError(t, err)
	assert.Equal(t, "only", newest.ClusterName)
}

func TestSelectClusterByAge_EmptyList(t *testing.T) {
	_, err := selectClusterByAge([]data.ClusterInfo{}, true)
	assert.Error(t, err)
}

func TestSelectClusterByAge_BadDate(t *testing.T) {
	clusters := []data.ClusterInfo{
		{ClusterName: "bad", CreatedAt: "not-a-date"},
	}

	_, err := selectClusterByAge(clusters, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bad")
}

func TestPrintSwitchSuccess_ARN(t *testing.T) {
	// printSwitchSuccess writes to stdout; just verify it doesn't panic
	// with various inputs.
	tests := []struct {
		name      string
		arn       string
		namespace string
		profile   string
	}{
		{"full arn", "arn:aws:eks:us-east-1:123456789012:cluster/demo", "", ""},
		{"with namespace", "arn:aws:eks:eu-west-1:111111111111:cluster/prod", "kube-system", ""},
		{"with profile", "arn:aws:eks:ap-south-1:222222222222:cluster/staging", "", "myprofile"},
		{"all set", "arn:aws:eks:us-west-2:333333333333:cluster/dev", "default", "dev-profile"},
		{"non-arn fallback", "some-cluster-name", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				printSwitchSuccess(tt.arn, tt.namespace, tt.profile)
			})
		})
	}
}
