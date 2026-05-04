package synthetic

import (
	"time"
)

type RandomClusterSpecConfig struct {
	NumNodes           int
	PodsPerNode        int
	NumNamespaces      int
	NumPVs             int
	NumLoadBalancers   int
	LabelsPerPod       int
	LabelsPerNode      int
	LabelsPerNamespace int
	Seed               uint64

	ScrapeInterval   time.Duration
	TotalDuration    time.Duration
	MinPodLifetime   time.Duration
	MaxPodLifetime   time.Duration
	PVCRatio         float64
	PodMigrationProb float64
}

func RandomClusterSpec(cfg RandomClusterSpecConfig) ClusterSpec {
	builder := NewClusterBuilder(cfg)
	return builder.Build()
}

var instanceResources = map[string][2]float64{
	"t3.medium": {2, 4}, "t3.large": {2, 8}, "t3.xlarge": {4, 16},
	"m5.large": {2, 8}, "m5.xlarge": {4, 16}, "m5.2xlarge": {8, 32}, "m5.4xlarge": {16, 64},
	"c5.large": {2, 4}, "c5.xlarge": {4, 8}, "c5.2xlarge": {8, 16}, "c5.4xlarge": {16, 32},
	"r5.large": {2, 16}, "r5.xlarge": {4, 32}, "r5.2xlarge": {8, 64}, "r5.4xlarge": {16, 128},
}

func instanceTypeResources(instType string) (float64, float64) {
	if res, ok := instanceResources[instType]; ok {
		return res[0], res[1]
	}
	return 4, 16
}
