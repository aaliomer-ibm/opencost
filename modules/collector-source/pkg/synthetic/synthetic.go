package synthetic

import (
	"time"
)

type ClusterSpec struct {
	ClusterUID  string
	ClusterName string
	Provider    string
	Region      string

	Nodes      []NodeSpec
	Namespaces []NamespaceSpec
	PVs        []PVSpec
	PVCs       []PVCSpec
	Services   []ServiceSpec
}

type NodeSpec struct {
	Name         string
	UID          string
	ProviderID   string
	InstanceType string
	CPUCores     float64
	RAMBytes     float64
	GPUCount     float64
	GPUType      string
	IsSpot       bool

	CPUCostPerHr    float64
	RAMCostPerGiBHr float64
	GPUCostPerHr    float64

	Labels map[string]string
}

type NamespaceSpec struct {
	Name   string
	UID    string
	Pods   []PodSpec
	Labels map[string]string
}

type PodSpec struct {
	Name       string
	UID        string
	NodeName   string
	Containers []ContainerSpec

	OwnerKind string
	OwnerName string
	OwnerUID  string

	Labels  map[string]string
	PVCRefs []PodPVCRef

	StartOffset   time.Duration
	Duration      time.Duration
	MigrationTime *time.Duration
	MigrationNode string
}

type PodPVCRef struct {
	ClaimName  string
	VolumeName string
}

type ContainerSpec struct {
	Name       string
	CPURequest float64
	RAMRequest float64
	CPULimit   float64
	RAMLimit   float64
	CPUUsage   float64
	RAMUsage   float64
	GPURequest float64
}

type PVSpec struct {
	Name          string
	UID           string
	ProviderID    string
	StorageClass  string
	CapacityBytes float64
	CostPerGiBHr  float64
}

type PVCSpec struct {
	Name          string
	UID           string
	Namespace     string
	NamespaceUID  string
	PVName        string
	PVUID         string
	StorageClass  string
	CapacityBytes float64
}

type ServiceSpec struct {
	Name      string
	UID       string
	Namespace string
	Type      string
	CostPerHr float64
	IP        string
}

type Generator struct {
	spec     ClusterSpec
	interval time.Duration
}

func NewGenerator(spec ClusterSpec, interval time.Duration) *Generator {
	return &Generator{
		spec:     spec,
		interval: interval,
	}
}
