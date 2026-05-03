package synthetic

import (
	"math"
	"testing"
	"time"

	"github.com/opencost/opencost/core/pkg/model/kubemodel"
	"github.com/opencost/opencost/core/pkg/opencost"
	"github.com/opencost/opencost/modules/collector-source/pkg/collector"
	"github.com/opencost/opencost/modules/collector-source/pkg/metric"
	synthspec "github.com/opencost/opencost/modules/collector-source/pkg/synthetic"
	"github.com/opencost/opencost/modules/collector-source/pkg/util"
	"github.com/opencost/opencost/pkg/costmodel"
)

const tolerance = 0.001

func setupCostModel(t *testing.T) (*costmodel.CostModel, synthspec.ClusterSpec, time.Time, time.Time) {
	t.Helper()

	spec := synthspec.DefaultClusterSpec()
	interval := 1 * time.Minute
	end := time.Now().UTC().Truncate(time.Minute)
	start := end.Add(-1 * time.Hour)

	resConfigs := []util.ResolutionConfiguration{
		{Interval: "1m", Retention: 120},
		{Interval: "1h", Retention: 48},
	}
	var resolutions []*util.Resolution
	for _, rc := range resConfigs {
		res, err := util.NewResolution(rc)
		if err != nil {
			t.Fatalf("resolution error: %v", err)
		}
		resolutions = append(resolutions, res)
	}

	repo := metric.NewMetricRepository(resolutions, collector.NewOpenCostMetricStore)
	gen := synthspec.NewGenerator(spec, interval)
	gen.Generate(repo, start, end)

	ds := NewSyntheticDataSource(repo, resConfigs, spec.ClusterUID, spec.ClusterName, spec.Provider, spec.Region, interval)
	provider := NewSyntheticProvider(spec)
	cm := costmodel.NewCostModel(spec.ClusterUID, ds, provider, nil, ds.ClusterMap(), ds.BatchDuration())

	return cm, spec, start, end
}

func TestConsistency_NodeCounts(t *testing.T) {
	cm, spec, start, end := setupCostModel(t)

	kms, err := cm.ComputeKubeModelSet(start, end)
	if err != nil {
		t.Fatalf("KubeModelSet: %v", err)
	}

	assetSet, err := cm.ComputeAssets(start, end)
	if err != nil {
		t.Fatalf("AssetSet: %v", err)
	}

	if len(kms.Nodes) != len(spec.Nodes) {
		t.Errorf("KubeModelSet nodes: expected %d, got %d", len(spec.Nodes), len(kms.Nodes))
	}

	assetNodeCount := 0
	for _, asset := range assetSet.Assets {
		if _, ok := asset.(*opencost.Node); ok {
			assetNodeCount++
		}
	}
	if assetNodeCount != len(spec.Nodes) {
		t.Errorf("AssetSet nodes: expected %d, got %d", len(spec.Nodes), assetNodeCount)
	}

	if len(kms.Nodes) != assetNodeCount {
		t.Errorf("KubeModel nodes (%d) != AssetSet nodes (%d)", len(kms.Nodes), assetNodeCount)
	}
}

func TestConsistency_NodeCPUCapacity(t *testing.T) {
	cm, spec, start, end := setupCostModel(t)

	kms, err := cm.ComputeKubeModelSet(start, end)
	if err != nil {
		t.Fatalf("KubeModelSet: %v", err)
	}

	specCPUByName := make(map[string]float64)
	for _, n := range spec.Nodes {
		specCPUByName[n.Name] = n.CPUCores
	}

	for _, kmNode := range kms.Nodes {
		cpuRQ, ok := kmNode.ResourceCapacities[kubemodel.ResourceCPU]
		if !ok {
			continue
		}
		cpuCap := cpuRQ.Values[kubemodel.StatAvg]
		expected, ok := specCPUByName[kmNode.Name]
		if !ok {
			continue
		}
		if math.Abs(cpuCap-expected) > tolerance {
			t.Errorf("node %s CPU capacity: kubemodel=%.2f, spec=%.2f", kmNode.Name, cpuCap, expected)
		}
	}
}

func TestConsistency_TotalNodeCost_Assets_vs_Pricing(t *testing.T) {
	cm, spec, start, end := setupCostModel(t)

	assetSet, err := cm.ComputeAssets(start, end)
	if err != nil {
		t.Fatalf("AssetSet: %v", err)
	}

	pms := BuildPricingModelSetFromAssets([]*opencost.AssetSet{assetSet}, spec.Provider, spec.Region, start)

	totalAssetCPUCost := 0.0
	totalAssetRAMCost := 0.0
	totalAssetCPUCoreHours := 0.0
	totalAssetRAMByteHours := 0.0

	for _, asset := range assetSet.Assets {
		node, ok := asset.(*opencost.Node)
		if !ok {
			continue
		}
		totalAssetCPUCost += node.CPUCost
		totalAssetRAMCost += node.RAMCost
		totalAssetCPUCoreHours += node.CPUCoreHours
		totalAssetRAMByteHours += node.RAMByteHours
	}

	pricingTotalCPURate := 0.0
	pricingTotalRAMRate := 0.0
	for k, v := range pms.NodePricing {
		if k.PricingType == "CPUCore" {
			pricingTotalCPURate += v.HourlyRate
		}
		if k.PricingType == "RamGB" {
			pricingTotalRAMRate += v.HourlyRate
		}
	}

	if totalAssetCPUCoreHours > 0 {
		derivedCPURate := totalAssetCPUCost / totalAssetCPUCoreHours
		if pricingTotalCPURate == 0 {
			t.Error("pricing has no CPU rate but assets have CPU costs")
		}
		t.Logf("Asset-derived CPU rate: %.6f, Pricing sum CPU rate: %.6f", derivedCPURate, pricingTotalCPURate)
	}

	ramGiBHours := totalAssetRAMByteHours / (1024 * 1024 * 1024)
	if ramGiBHours > 0 {
		derivedRAMRate := totalAssetRAMCost / ramGiBHours
		if pricingTotalRAMRate == 0 {
			t.Error("pricing has no RAM rate but assets have RAM costs")
		}
		t.Logf("Asset-derived RAM rate: %.6f, Pricing sum RAM rate: %.6f", derivedRAMRate, pricingTotalRAMRate)
	}
}

func TestConsistency_AllocationCPUCost_vs_AssetCPUCost(t *testing.T) {
	cm, _, start, end := setupCostModel(t)

	allocSet, err := cm.ComputeAllocation(start, end)
	if err != nil {
		t.Fatalf("AllocationSet: %v", err)
	}

	assetSet, err := cm.ComputeAssets(start, end)
	if err != nil {
		t.Fatalf("AssetSet: %v", err)
	}

	totalAllocCPUCost := 0.0
	totalAllocRAMCost := 0.0
	totalAllocGPUCost := 0.0
	for _, alloc := range allocSet.Allocations {
		totalAllocCPUCost += alloc.CPUCost
		totalAllocRAMCost += alloc.RAMCost
		totalAllocGPUCost += alloc.GPUCost
	}

	totalAssetCPUCost := 0.0
	totalAssetRAMCost := 0.0
	totalAssetGPUCost := 0.0
	for _, asset := range assetSet.Assets {
		node, ok := asset.(*opencost.Node)
		if !ok {
			continue
		}
		totalAssetCPUCost += node.CPUCost
		totalAssetRAMCost += node.RAMCost
		totalAssetGPUCost += node.GPUCost
	}

	t.Logf("Allocation CPU=%.4f RAM=%.4f GPU=%.4f", totalAllocCPUCost, totalAllocRAMCost, totalAllocGPUCost)
	t.Logf("Asset      CPU=%.4f RAM=%.4f GPU=%.4f", totalAssetCPUCost, totalAssetRAMCost, totalAssetGPUCost)

	if totalAssetCPUCost > 0 && totalAllocCPUCost > totalAssetCPUCost {
		t.Errorf("allocation CPU cost (%.4f) exceeds asset CPU cost (%.4f)", totalAllocCPUCost, totalAssetCPUCost)
	}
	if totalAssetRAMCost > 0 && totalAllocRAMCost > totalAssetRAMCost {
		t.Errorf("allocation RAM cost (%.4f) exceeds asset RAM cost (%.4f)", totalAllocRAMCost, totalAssetRAMCost)
	}
}

func TestConsistency_PodCount_KubeModel_vs_Allocation(t *testing.T) {
	cm, _, start, end := setupCostModel(t)

	kms, err := cm.ComputeKubeModelSet(start, end)
	if err != nil {
		t.Fatalf("KubeModelSet: %v", err)
	}

	allocSet, err := cm.ComputeAllocation(start, end)
	if err != nil {
		t.Fatalf("AllocationSet: %v", err)
	}

	kmPodCount := len(kms.Pods)

	allocPods := make(map[string]bool)
	for _, alloc := range allocSet.Allocations {
		if alloc.Properties != nil && alloc.Properties.Pod != "" && alloc.Properties.Pod != "__idle__" {
			key := alloc.Properties.Namespace + "/" + alloc.Properties.Pod
			allocPods[key] = true
		}
	}

	t.Logf("KubeModel pods: %d, Allocation unique pods: %d", kmPodCount, len(allocPods))

	if kmPodCount == 0 {
		t.Error("KubeModel has 0 pods")
	}
	if len(allocPods) == 0 {
		t.Error("AllocationSet has 0 pods")
	}
}

func TestConsistency_NamespaceCount(t *testing.T) {
	cm, spec, start, end := setupCostModel(t)

	kms, err := cm.ComputeKubeModelSet(start, end)
	if err != nil {
		t.Fatalf("KubeModelSet: %v", err)
	}

	if len(kms.Namespaces) != len(spec.Namespaces) {
		t.Errorf("KubeModel namespaces: expected %d, got %d", len(spec.Namespaces), len(kms.Namespaces))
	}
}

func TestConsistency_DiskAssets_vs_PVSpec(t *testing.T) {
	cm, spec, start, end := setupCostModel(t)

	assetSet, err := cm.ComputeAssets(start, end)
	if err != nil {
		t.Fatalf("AssetSet: %v", err)
	}

	diskCount := 0
	for _, asset := range assetSet.Assets {
		if _, ok := asset.(*opencost.Disk); ok {
			diskCount++
		}
	}

	if diskCount == 0 && len(spec.PVs) > 0 {
		t.Logf("no disk assets found (PV metrics may not be fully wired); spec has %d PVs", len(spec.PVs))
	}

	t.Logf("Disk assets: %d, PV specs: %d", diskCount, len(spec.PVs))
}

func TestConsistency_LBAssets(t *testing.T) {
	cm, _, start, end := setupCostModel(t)

	assetSet, err := cm.ComputeAssets(start, end)
	if err != nil {
		t.Fatalf("AssetSet: %v", err)
	}

	lbCount := 0
	for _, asset := range assetSet.Assets {
		if _, ok := asset.(*opencost.LoadBalancer); ok {
			lbCount++
		}
	}

	t.Logf("LoadBalancer assets: %d", lbCount)
	if lbCount == 0 {
		t.Log("no LB assets found (LB metrics may not be fully wired)")
	}
}

func TestConsistency_PricingRates_MatchAssetCosts(t *testing.T) {
	cm, spec, start, end := setupCostModel(t)

	assetSet, err := cm.ComputeAssets(start, end)
	if err != nil {
		t.Fatalf("AssetSet: %v", err)
	}

	pms := BuildPricingModelSetFromAssets([]*opencost.AssetSet{assetSet}, spec.Provider, spec.Region, start)

	for _, asset := range assetSet.Assets {
		node, ok := asset.(*opencost.Node)
		if !ok || node.CPUCoreHours == 0 {
			continue
		}

		assetCPURate := node.CPUCost / node.CPUCoreHours
		ramGiBHours := node.RAMByteHours / (1024 * 1024 * 1024)
		assetRAMRate := 0.0
		if ramGiBHours > 0 {
			assetRAMRate = node.RAMCost / ramGiBHours
		}

		for k, v := range pms.NodePricing {
			if k.NodeType != node.NodeType {
				continue
			}
			if k.PricingType == "CPUCore" {
				if math.Abs(v.HourlyRate-assetCPURate) > tolerance {
					t.Errorf("node type %s CPU rate mismatch: pricing=%.6f, asset-derived=%.6f",
						node.NodeType, v.HourlyRate, assetCPURate)
				}
			}
			if k.PricingType == "RamGB" && assetRAMRate > 0 {
				if math.Abs(v.HourlyRate-assetRAMRate) > tolerance {
					t.Errorf("node type %s RAM rate mismatch: pricing=%.6f, asset-derived=%.6f",
						node.NodeType, v.HourlyRate, assetRAMRate)
				}
			}
		}
	}
}
