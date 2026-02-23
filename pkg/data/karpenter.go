package data

import "time"

type KarpenterNodePoolInfo struct {
	Profile           string
	Region            string
	ClusterName       string
	Name              string
	NodeClassName     string
	InstanceTypes     []string
	CapacityTypes     []string
	Zones             []string
	CPULimit          string
	MemoryLimit       string
	ConsolidationMode string
	ExpireAfter       string
	Weight            int32
}

type KarpenterNodeClaimInfo struct {
	Profile      string
	Region       string
	ClusterName  string
	Name         string
	NodeName     string
	NodePoolName string
	InstanceType string
	Zone         string
	CapacityType string
	AMI          string
	Status       string
	Age          time.Time
	Drifted      bool
}

type KarpenterAMIUsageInfo struct {
	Profile      string
	Region       string
	ClusterName  string
	NodePoolName string
	CurrentAMI   string
	NodeCount    int
}

type KarpenterDriftInfo struct {
	Profile      string
	Region       string
	ClusterName  string
	ResourceType string
	Name         string
	NodeName     string
	NodePoolName string
	DriftedSince time.Time
	Reason       string
}
