package cmd

import (
	"testing"

	"github.com/pet2cattle/kubectl-eks/pkg/data"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestCountReadyContainers(t *testing.T) {
	tests := []struct {
		name          string
		pod           corev1.Pod
		expectedReady int
		expectedTotal int
	}{
		{
			name: "all ready",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "a"}, {Name: "b"}},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true},
						{Ready: true},
					},
				},
			},
			expectedReady: 2,
			expectedTotal: 2,
		},
		{
			name: "partial ready",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true},
						{Ready: false},
						{Ready: true},
					},
				},
			},
			expectedReady: 2,
			expectedTotal: 3,
		},
		{
			name: "no containers",
			pod: corev1.Pod{
				Spec:   corev1.PodSpec{},
				Status: corev1.PodStatus{},
			},
			expectedReady: 0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready, total := countReadyContainers(tt.pod)
			assert.Equal(t, tt.expectedReady, ready)
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}

func TestGetPodPendingReason(t *testing.T) {
	tests := []struct {
		name     string
		pod      corev1.Pod
		expected string
	}{
		{
			name: "unschedulable",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:    corev1.PodScheduled,
							Status:  corev1.ConditionFalse,
							Message: "no nodes available",
						},
					},
				},
			},
			expected: "Unschedulable: no nodes available",
		},
		{
			name: "container waiting",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "ImagePullBackOff",
								},
							},
						},
					},
				},
			},
			expected: "Waiting: ImagePullBackOff",
		},
		{
			name:     "generic pending",
			pod:      corev1.Pod{},
			expected: "Pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPodPendingReason(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPodFailedReason(t *testing.T) {
	tests := []struct {
		name     string
		pod      corev1.Pod
		expected string
	}{
		{
			name: "status reason set",
			pod: corev1.Pod{
				Status: corev1.PodStatus{Reason: "Evicted"},
			},
			expected: "Evicted",
		},
		{
			name: "container terminated reason",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									Reason: "OOMKilled",
								},
							},
						},
					},
				},
			},
			expected: "OOMKilled",
		},
		{
			name:     "generic failed",
			pod:      corev1.Pod{},
			expected: "Failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPodFailedReason(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDeploymentConditionMessage(t *testing.T) {
	three := int32(3)

	tests := []struct {
		name     string
		deploy   appsv1.Deployment
		expected string
	}{
		{
			name: "unavailable condition",
			deploy: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{Replicas: &three},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: 1,
					Conditions: []appsv1.DeploymentCondition{
						{
							Type:    appsv1.DeploymentAvailable,
							Status:  corev1.ConditionFalse,
							Message: "Deployment does not have minimum availability",
						},
					},
				},
			},
			expected: "Deployment does not have minimum availability",
		},
		{
			name: "progressing failed",
			deploy: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{Replicas: &three},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: 2,
					Conditions: []appsv1.DeploymentCondition{
						{
							Type:    appsv1.DeploymentProgressing,
							Status:  corev1.ConditionFalse,
							Message: "ReplicaSet has timed out progressing",
						},
					},
				},
			},
			expected: "ReplicaSet has timed out progressing",
		},
		{
			name: "no failing conditions - fallback",
			deploy: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{Replicas: &three},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: 2,
					Conditions:    []appsv1.DeploymentCondition{},
				},
			},
			expected: "Ready 2/3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDeploymentConditionMessage(tt.deploy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSummarizeResults(t *testing.T) {
	cluster := data.ClusterInfo{
		AWSProfile:  "prod",
		Region:      "us-east-1",
		ClusterName: "my-cluster",
	}

	tests := []struct {
		name            string
		results         []data.HealthCheckResult
		expectedStatus  string
		expectedHealthy int
		expectedTotal   int
	}{
		{
			name: "all healthy",
			results: []data.HealthCheckResult{
				{Kind: "Pod", IsHealthy: true},
				{Kind: "Pod", IsHealthy: true},
				{Kind: "Deployment", IsHealthy: true},
			},
			expectedStatus: "Healthy",
		},
		{
			name: "some unhealthy",
			results: []data.HealthCheckResult{
				{Kind: "Pod", IsHealthy: true},
				{Kind: "Pod", IsHealthy: false},
				{Kind: "Deployment", IsHealthy: true},
				{Kind: "Deployment", IsHealthy: false},
				{Kind: "StatefulSet", IsHealthy: false},
			},
			expectedStatus: "3 Unhealthy",
		},
		{
			name:           "empty results",
			results:        []data.HealthCheckResult{},
			expectedStatus: "Healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := summarizeResults(cluster, tt.results)
			assert.Equal(t, "prod", summary.Profile)
			assert.Equal(t, "us-east-1", summary.Region)
			assert.Equal(t, "my-cluster", summary.ClusterName)
			assert.Equal(t, tt.expectedStatus, summary.OverallStatus)
		})
	}
}

func TestSummarizeResults_Counts(t *testing.T) {
	cluster := data.ClusterInfo{ClusterName: "test"}

	results := []data.HealthCheckResult{
		{Kind: "Pod", IsHealthy: true},
		{Kind: "Pod", IsHealthy: false},
		{Kind: "Pod", IsHealthy: true},
		{Kind: "Deployment", IsHealthy: true},
		{Kind: "StatefulSet", IsHealthy: false},
		{Kind: "DaemonSet", IsHealthy: true},
		{Kind: "DaemonSet", IsHealthy: true},
		{Kind: "ReplicaSet", IsHealthy: false},
	}

	summary := summarizeResults(cluster, results)

	assert.Equal(t, 3, summary.TotalPods)
	assert.Equal(t, 2, summary.HealthyPods)
	assert.Equal(t, 1, summary.TotalDeployments)
	assert.Equal(t, 1, summary.HealthyDeployments)
	assert.Equal(t, 1, summary.TotalStatefulSets)
	assert.Equal(t, 0, summary.HealthyStatefulSets)
	assert.Equal(t, 2, summary.TotalDaemonSets)
	assert.Equal(t, 2, summary.HealthyDaemonSets)
	assert.Equal(t, 1, summary.TotalReplicaSets)
	assert.Equal(t, 0, summary.HealthyReplicaSets)
	assert.Equal(t, "3 Unhealthy", summary.OverallStatus)
}
