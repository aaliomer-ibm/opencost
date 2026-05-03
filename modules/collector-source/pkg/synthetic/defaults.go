package synthetic

import "time"

const (
	GiB = 1024 * 1024 * 1024
	MiB = 1024 * 1024

	DefaultMinPodLifetime   = 10 * time.Minute
	DefaultMaxPodLifetime   = 1 * time.Hour
	DefaultPVCRatio         = 0.1
	DefaultPodMigrationProb = 0.05
	DefaultGPUProbability   = 0.3
	DefaultMinGPUCount      = 1
	DefaultMaxGPUCount      = 4

	SpotNodeProbability = 0.3

	ContainerGPUAssignmentProb = 0.5
	MinCPULimitMultiplier      = 1.1
	MaxCPULimitMultiplier      = 3.1
	MinRAMLimitMultiplier      = 1.1
	MaxRAMLimitMultiplier      = 2.6
	MinCPUUsageRatio           = 0.2
	MaxCPUUsageRatio           = 0.96
	MinRAMUsageRatio           = 0.3
	MaxRAMUsageRatio           = 0.91
	NodeResourceSafetyMargin   = 0.9
)

func DefaultClusterSpec() ClusterSpec {
	spec := RandomClusterSpec(RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 1, NumNamespaces: 1, NumPVs: 1, NumLoadBalancers: 1,
		LabelsPerPod: 3, LabelsPerNode: 5, LabelsPerNamespace: 2, Seed: 42,
		ScrapeInterval: time.Minute, TotalDuration: 36 * time.Hour,
		MinPodLifetime: DefaultMinPodLifetime, MaxPodLifetime: DefaultMaxPodLifetime,
		PVCRatio: DefaultPVCRatio, PodMigrationProb: DefaultPodMigrationProb,
	})
	for i := range spec.Namespaces {
		for j := range spec.Namespaces[i].Pods {
			spec.Namespaces[i].Pods[j].StartOffset = 0
			spec.Namespaces[i].Pods[j].Duration = 0
		}
	}
	if len(spec.Nodes) > 1 {
		spec.Nodes[1].GPUCount = 1
		spec.Nodes[1].GPUType = "nvidia-tesla-t4"
		spec.Nodes[1].GPUCostPerHr = 0.50
		if len(spec.Namespaces) > 0 && len(spec.Namespaces[0].Pods) > 1 {
			spec.Namespaces[0].Pods[1].NodeName = spec.Nodes[1].Name
			if len(spec.Namespaces[0].Pods[1].Containers) > 0 {
				spec.Namespaces[0].Pods[1].Containers[0].GPURequest = 1.0
			}
		}
	}
	return spec
}

func DefaultRandomConfig() RandomClusterSpecConfig {
	return RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 1, NumNamespaces: 1, NumPVs: 1, NumLoadBalancers: 1,
		LabelsPerPod: 3, LabelsPerNode: 5, LabelsPerNamespace: 2, Seed: 42,
		ScrapeInterval: time.Minute, TotalDuration: 36 * time.Hour,
		MinPodLifetime: DefaultMinPodLifetime, MaxPodLifetime: DefaultMaxPodLifetime,
		PVCRatio: DefaultPVCRatio, PodMigrationProb: DefaultPodMigrationProb,
	}
}
