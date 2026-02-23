package data

import (
	"time"
)

type ClusterInfo struct {
	ClusterName  string
	Namespace    string
	Region       string
	AWSProfile   string
	AWSAccountID string
	Status       string
	Version      string
	Arn          string
	CreatedAt    string
}

type ClusterNodeInfo struct {
	Profile     string
	Region      string
	ClusterName string
	Node        NodeInfo
}

type NodeInfo struct {
	Name         string
	InstanceType string
	Compute      string
	ManagedBy    string
	Created      time.Time
	Status       string
}

type KubeCtlEksCache struct {
	ClusterByARN map[string]ClusterInfo
	ClusterList  map[string]map[string][]ClusterInfo
}

type JsonPathResult struct {
	Profile     string
	Region      string
	ClusterName string
	Namespace   string
	Resource    string
	Value       string
	Error       string
}

type ResourceResult struct {
	Profile     string
	Region      string
	ClusterName string
	Namespace   string
	Name        string
	Kind        string
	Data        interface{}
	Error       string
	Status      string
}

type AWSProfile struct {
	Name           string
	DefaultRegion  string
	HintEKSRegions []string
}

type AWSConfig struct {
	Profiles map[string]AWSProfile
}

type AMIInfo struct {
	ID              string
	Name            string
	Architecture    string
	State           string
	DeprecationTime string
}

type ClientStat struct {
	LastRequestTime            time.Time `json:"lastRequestTime"`
	NumberOfRequestsLast30Days int64     `json:"numberOfRequestsLast30Days"`
	UserAgent                  string    `json:"userAgent"`
}

type DeprecationDetail struct {
	ClientStats                    []ClientStat `json:"clientStats"`
	ReplacedWith                   string       `json:"replacedWith"`
	StartServingReplacementVersion string       `json:"startServingReplacementVersion"`
	StopServingVersion             string       `json:"stopServingVersion"`
	Usage                          string       `json:"usage"`
}

type CategorySpecificSummary struct {
	DeprecationDetails []DeprecationDetail `json:"deprecationDetails"`
}

type EKSInsightInfo struct {
	ID             string
	Description    string
	Category       string
	Status         string
	Recommendation string
	Reason         string
	Summary        *CategorySpecificSummary
	AdditionalInfo *map[string]*string
}

type WhoAmIInfo struct {
	AWSProfile  string
	Region      string
	ClusterName string
	AWSArn      string
	AWSAccount  string
	AWSUserId   string
	K8sUsername string
	K8sUID      string
	K8sGroups   []string
}

type ResourceQuotaInfo struct {
	Profile      string
	Region       string
	ClusterName  string
	Namespace    string
	QuotaName    string
	ResourceName string
	Hard         string
	Used         string
}

type EventInfo struct {
	Profile     string
	Region      string
	ClusterName string
	Namespace   string
	LastSeen    time.Time
	Type        string
	Reason      string
	Object      string
	Message     string
	Count       int32
}
