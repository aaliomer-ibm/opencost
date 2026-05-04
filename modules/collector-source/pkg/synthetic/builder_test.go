package synthetic

import (
	"testing"
	"time"

	"github.com/opencost/opencost/modules/collector-source/pkg/metric"
)


func TestBuilder_ProducesCorrectNodeCount(t *testing.T) {
	for _, numNodes := range []int{1, 3, 5} {
		spec := buildSpec(t, RandomClusterSpecConfig{
			NumNodes: numNodes, PodsPerNode: 1, NumNamespaces: 1,
			NumLoadBalancers: 0, Seed: 1,
		})
		if got := len(spec.Nodes); got != numNodes {
			t.Errorf("NumNodes=%d: got %d nodes", numNodes, got)
		}
	}
}

func TestBuilder_ProducesCorrectNamespaceCount(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 1, NumNamespaces: 4,
		NumLoadBalancers: 0, Seed: 2,
	})
	if got := len(spec.Namespaces); got != 4 {
		t.Errorf("expected 4 namespaces, got %d", got)
	}
}

func TestBuilder_ProducesCorrectPodCount(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 3, PodsPerNode: 4, NumNamespaces: 2,
		NumLoadBalancers: 0, Seed: 3,
	})
	totalPods := countAllPods(spec)
	if totalPods != 12 {
		t.Errorf("expected 12 total pods, got %d", totalPods)
	}
}

func TestBuilder_ServiceCounts(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 1, NumNamespaces: 2,
		NumLoadBalancers: 3, Seed: 7,
	})
	lbCount, clusterIPCount := 0, 0
	for _, svc := range spec.Services {
		switch svc.Type {
		case "LoadBalancer":
			lbCount++
		case "ClusterIP":
			clusterIPCount++
		}
	}
	if lbCount != 3 {
		t.Errorf("expected 3 LoadBalancers, got %d", lbCount)
	}
	if clusterIPCount != 2 {
		t.Errorf("expected 2 ClusterIP services, got %d", clusterIPCount)
	}
}

func TestBuilder_AllNodesHaveRequiredFields(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 5, PodsPerNode: 1, NumNamespaces: 1, Seed: 8,
	})
	for i, n := range spec.Nodes {
		if n.Name == "" || n.UID == "" || n.InstanceType == "" {
			t.Errorf("node[%d] missing required field", i)
		}
		if n.CPUCores <= 0 || n.RAMBytes <= 0 {
			t.Errorf("node[%d] has zero resources", i)
		}
	}
}

func TestBuilder_AllPodsHaveContainers(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 3, NumNamespaces: 2, Seed: 9,
	})
	for _, ns := range spec.Namespaces {
		for _, pod := range ns.Pods {
			if len(pod.Containers) == 0 {
				t.Errorf("pod %q has no containers", pod.Name)
			}
		}
	}
}

func TestBuilder_DeterministicWithSameSeed(t *testing.T) {
	cfg := RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 2, NumNamespaces: 2,
		NumPVs: 2, NumLoadBalancers: 1, Seed: 42, PVCRatio: 0.5,
	}
	a := NewClusterBuilder(cfg).Build()
	b := NewClusterBuilder(cfg).Build()

	if a.ClusterUID != b.ClusterUID {
		t.Error("same seed produced different ClusterUIDs")
	}
	if len(a.PVCs) != len(b.PVCs) {
		t.Errorf("same seed: PVC count %d vs %d", len(a.PVCs), len(b.PVCs))
	}
}

func TestBuilder_DifferentSeedsProduceDifferentSpecs(t *testing.T) {
	cfg := RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 2, NumNamespaces: 1,
		NumPVs: 1, Seed: 1, PVCRatio: 0.5,
	}
	a := NewClusterBuilder(cfg).Build()
	cfg.Seed = 999
	b := NewClusterBuilder(cfg).Build()

	if a.ClusterUID == b.ClusterUID {
		t.Error("different seeds produced identical ClusterUIDs")
	}
}


func TestBuilder_PVPoolSizeMatchesConfig(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 3, NumNamespaces: 1,
		NumPVs: 5, Seed: 10, PVCRatio: 1.0,
	})
	if len(spec.PVs) != 5 {
		t.Errorf("expected 5 PVs, got %d", len(spec.PVs))
	}
}

func TestBuilder_NoPVCsWhenPoolEmpty(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 3, NumNamespaces: 1,
		NumPVs: 0, Seed: 11, PVCRatio: 1.0,
	})
	if len(spec.PVs) != 0 {
		t.Errorf("expected 0 PVs, got %d", len(spec.PVs))
	}
	if len(spec.PVCs) != 0 {
		t.Errorf("expected 0 PVCs with empty pool, got %d", len(spec.PVCs))
	}
}

func TestBuilder_NoPVCsWhenNoPods(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 0, PodsPerNode: 0, NumNamespaces: 1,
		NumPVs: 3, Seed: 12, PVCRatio: 1.0,
	})
	if len(spec.PVs) != 3 {
		t.Errorf("expected 3 PVs in pool, got %d", len(spec.PVs))
	}
	if len(spec.PVCs) != 0 {
		t.Errorf("expected 0 PVCs when no pods, got %d", len(spec.PVCs))
	}
}

func TestBuilder_NoPVCsWhenRatioZero(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 3, NumNamespaces: 1,
		NumPVs: 0, Seed: 13, PVCRatio: 0.5,
	})
	if len(spec.PVs) != 0 {
		t.Errorf("expected 0 PVs, got %d", len(spec.PVs))
	}
	if len(spec.PVCs) != 0 {
		t.Errorf("expected 0 PVCs with no PV pool, got %d", len(spec.PVCs))
	}
}

func TestBuilder_PVCsCappedByPoolSize(t *testing.T) {
	spec := ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{{
			Name: "n1", UID: "n1", ProviderID: "i-1",
			InstanceType: "m5.large", CPUCores: 4, RAMBytes: 16 * GiB,
			CPUCostPerHr: 0.1, RAMCostPerGiBHr: 0.01, Labels: map[string]string{},
		}},
		Namespaces: []NamespaceSpec{{
			Name: "app", UID: "ns1",
			Pods: []PodSpec{
				{Name: "p1", UID: "p1", NodeName: "n1", Duration: 0, Containers: []ContainerSpec{{Name: "c", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}}, Labels: map[string]string{}},
				{Name: "p2", UID: "p2", NodeName: "n1", Duration: 0, Containers: []ContainerSpec{{Name: "c", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}}, Labels: map[string]string{}},
				{Name: "p3", UID: "p3", NodeName: "n1", Duration: 0, Containers: []ContainerSpec{{Name: "c", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}}, Labels: map[string]string{}},
			},
			Labels: map[string]string{},
		}},
		PVs: []PVSpec{{
			Name: "pv-1", UID: "pv1", ProviderID: "vol-1",
			StorageClass: "gp3", CapacityBytes: 100 * GiB, CostPerGiBHr: 0.0003,
		}},
	}

	builder := NewClusterBuilder(RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 3, NumNamespaces: 1,
		NumPVs: 1, Seed: 99, PVCRatio: 1.0,
	})
	builder.assignStorage(&spec)

	if len(spec.PVCs) != 1 {
		t.Errorf("expected 1 PVC (pool capped), got %d", len(spec.PVCs))
	}
	podsWithStorage := 0
	for _, pod := range spec.Namespaces[0].Pods {
		if len(pod.PVCRefs) > 0 {
			podsWithStorage++
		}
	}
	if podsWithStorage != 1 {
		t.Errorf("expected 1 pod with storage, got %d", podsWithStorage)
	}
}

func TestBuilder_PVReusedAfterPodEnds(t *testing.T) {
	spec := ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{{
			Name: "n1", UID: "n1", ProviderID: "i-1",
			InstanceType: "m5.large", CPUCores: 4, RAMBytes: 16 * GiB,
			CPUCostPerHr: 0.1, RAMCostPerGiBHr: 0.01, Labels: map[string]string{},
		}},
		Namespaces: []NamespaceSpec{{
			Name: "app", UID: "ns1",
			Pods: []PodSpec{
				{Name: "p1", UID: "p1", NodeName: "n1", StartOffset: 0, Duration: 10 * time.Minute,
					Containers: []ContainerSpec{{Name: "c", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}},
					Labels:     map[string]string{}},
				{Name: "p2", UID: "p2", NodeName: "n1", StartOffset: 10 * time.Minute, Duration: 10 * time.Minute,
					Containers: []ContainerSpec{{Name: "c", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}},
					Labels:     map[string]string{}},
				{Name: "p3", UID: "p3", NodeName: "n1", StartOffset: 20 * time.Minute, Duration: 10 * time.Minute,
					Containers: []ContainerSpec{{Name: "c", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}},
					Labels:     map[string]string{}},
			},
			Labels: map[string]string{},
		}},
		PVs: []PVSpec{{
			Name: "pv-1", UID: "pv1", ProviderID: "vol-1",
			StorageClass: "gp3", CapacityBytes: 100 * GiB, CostPerGiBHr: 0.0003,
		}},
	}

	builder := NewClusterBuilder(RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 3, NumNamespaces: 1,
		NumPVs: 1, Seed: 99, PVCRatio: 1.0,
	})
	builder.assignStorage(&spec)

	if len(spec.PVCs) != 3 {
		t.Errorf("expected 3 PVCs (PV reused), got %d", len(spec.PVCs))
	}
	for _, pod := range spec.Namespaces[0].Pods {
		if len(pod.PVCRefs) != 1 {
			t.Errorf("pod %s: expected 1 PVCRef, got %d", pod.Name, len(pod.PVCRefs))
		}
	}
}

func TestBuilder_PVNotReusedWhileStillClaimed(t *testing.T) {
	spec := ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{{
			Name: "n1", UID: "n1", ProviderID: "i-1",
			InstanceType: "m5.large", CPUCores: 4, RAMBytes: 16 * GiB,
			CPUCostPerHr: 0.1, RAMCostPerGiBHr: 0.01, Labels: map[string]string{},
		}},
		Namespaces: []NamespaceSpec{{
			Name: "app", UID: "ns1",
			Pods: []PodSpec{
				{Name: "p1", UID: "p1", NodeName: "n1", StartOffset: 0, Duration: 20 * time.Minute,
					Containers: []ContainerSpec{{Name: "c", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}},
					Labels:     map[string]string{}},
				{Name: "p2", UID: "p2", NodeName: "n1", StartOffset: 5 * time.Minute, Duration: 20 * time.Minute,
					Containers: []ContainerSpec{{Name: "c", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}},
					Labels:     map[string]string{}},
			},
			Labels: map[string]string{},
		}},
		PVs: []PVSpec{{
			Name: "pv-1", UID: "pv1", ProviderID: "vol-1",
			StorageClass: "gp3", CapacityBytes: 100 * GiB, CostPerGiBHr: 0.0003,
		}},
	}

	builder := NewClusterBuilder(RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 2, NumNamespaces: 1,
		NumPVs: 1, Seed: 99, PVCRatio: 1.0,
	})
	builder.assignStorage(&spec)

	if len(spec.PVCs) != 1 {
		t.Errorf("expected 1 PVC (overlap blocks second), got %d", len(spec.PVCs))
	}
	if len(spec.Namespaces[0].Pods[0].PVCRefs) != 1 {
		t.Error("pod1 should have gotten the PVC")
	}
	if len(spec.Namespaces[0].Pods[1].PVCRefs) != 0 {
		t.Error("pod2 should NOT have gotten a PVC (PV still in use)")
	}
}

func TestBuilder_PVCsBoundToMatchingPV(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 2, NumNamespaces: 1,
		NumPVs: 2, Seed: 5, PVCRatio: 1.0,
	})
	pvByName := make(map[string]PVSpec)
	for _, pv := range spec.PVs {
		pvByName[pv.Name] = pv
	}
	for _, pvc := range spec.PVCs {
		pv, ok := pvByName[pvc.PVName]
		if !ok {
			t.Fatalf("PVC %q references missing PV %q", pvc.Name, pvc.PVName)
		}
		if pvc.PVUID != pv.UID {
			t.Errorf("PVC %q PVUID mismatch", pvc.Name)
		}
		if pvc.StorageClass != pv.StorageClass {
			t.Errorf("PVC %q StorageClass mismatch", pvc.Name)
		}
		if pvc.CapacityBytes != pv.CapacityBytes {
			t.Errorf("PVC %q CapacityBytes mismatch", pvc.Name)
		}
	}
}

func TestBuilder_PVCsInValidNamespaces(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 2, NumNamespaces: 3,
		NumPVs: 4, Seed: 6, PVCRatio: 1.0,
	})
	nsUIDs := make(map[string]string)
	for _, ns := range spec.Namespaces {
		nsUIDs[ns.Name] = ns.UID
	}
	for _, pvc := range spec.PVCs {
		uid, ok := nsUIDs[pvc.Namespace]
		if !ok {
			t.Fatalf("PVC %q in unknown namespace %q", pvc.Name, pvc.Namespace)
		}
		if pvc.NamespaceUID != uid {
			t.Errorf("PVC %q NamespaceUID mismatch", pvc.Name)
		}
	}
}


func TestBuilderSpec_GeneratesClusterMetric(t *testing.T) {
	spec := buildSpec(t, defaultTestConfig())
	metrics := generateOneTick(t, spec)
	assertMetricCount(t, metrics, metric.ClusterInfo, 1, "ClusterInfo")
}

func TestBuilderSpec_GeneratesNodeMetrics(t *testing.T) {
	spec := buildSpec(t, defaultTestConfig())
	metrics := generateOneTick(t, spec)
	assertMetricCount(t, metrics, metric.NodeInfo, len(spec.Nodes), "NodeInfo")
}

func TestBuilderSpec_GeneratesPodMetricsForActivePods(t *testing.T) {
	spec := buildSpec(t, defaultTestConfig())
	forcePodsAlwaysAlive(&spec)
	metrics := generateOneTick(t, spec)
	assertMetricCount(t, metrics, metric.PodInfo, countAllPods(spec), "PodInfo")
}

func TestBuilderSpec_GeneratesPVMetricsUnconditionally(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 3, NumNamespaces: 1,
		NumPVs: 3, Seed: 20, PVCRatio: 1.0,
	})
	pvCount := len(spec.PVs)
	for i := range spec.Namespaces {
		spec.Namespaces[i].Pods = nil
	}
	metrics := generateOneTick(t, spec)
	assertMetricCount(t, metrics, metric.KubecostPVInfo, pvCount, "PV info without pods")
	assertMetricCount(t, metrics, metric.KubePersistentVolumeClaimInfo, 0, "PVC info without pods")
}

func TestBuilderSpec_PVCMetricsOnlyWhenPodActive(t *testing.T) {
	spec := buildSpecWithPVC(t)
	forcePodsAlwaysAlive(&spec)
	metrics := generateOneTick(t, spec)
	referencedPVCs := countReferencedPVCs(spec)
	assertMetricCount(t, metrics, metric.KubePersistentVolumeClaimInfo, referencedPVCs, "PVC info")
}

func TestBuilderSpec_ServiceMetrics(t *testing.T) {
	spec := buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 1, NumNamespaces: 2,
		NumLoadBalancers: 2, Seed: 30,
	})
	metrics := generateOneTick(t, spec)
	assertMetricCount(t, metrics, metric.ServiceInfo, len(spec.Services), "ServiceInfo")
}


func TestBuilderSpec_PVCAssignedToOnePod(t *testing.T) {
	spec := ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{{
			Name: "node-1", UID: "node-uid-1", ProviderID: "i-1",
			InstanceType: "m5.large", CPUCores: 4, RAMBytes: 16 * GiB,
			CPUCostPerHr: 0.1, RAMCostPerGiBHr: 0.01, Labels: map[string]string{},
		}},
		Namespaces: []NamespaceSpec{{
			Name: "app", UID: "ns-uid-app",
			Pods: []PodSpec{{
				Name: "worker", UID: "pod-uid-1", NodeName: "node-1",
				Containers: []ContainerSpec{{Name: "app", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}},
				PVCRefs:    []PodPVCRef{{ClaimName: "pvc-data", VolumeName: "data-vol"}},
				Labels:     map[string]string{},
			}},
			Labels: map[string]string{},
		}},
		PVs:  []PVSpec{{Name: "pv-data", UID: "pv-uid-1", ProviderID: "vol-abc", StorageClass: "gp3", CapacityBytes: 50 * GiB, CostPerGiBHr: 0.0003}},
		PVCs: []PVCSpec{{Name: "pvc-data", UID: "pvc-uid-1", Namespace: "app", NamespaceUID: "ns-uid-app", PVName: "pv-data", PVUID: "pv-uid-1", StorageClass: "gp3", CapacityBytes: 50 * GiB}},
	}
	metrics := generateOneTick(t, spec)
	assertMetricCount(t, metrics, metric.KubePersistentVolumeClaimInfo, 1, "PVC info")
	assertMetricCount(t, metrics, metric.PodPVCVolume, 1, "PodPVCVolume")
	assertMetricCount(t, metrics, metric.KubecostPVInfo, 1, "PV info")
}

func TestBuilderSpec_PVCNotEmittedWhenPodDead(t *testing.T) {
	spec := ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{{Name: "n1", UID: "n1", ProviderID: "i-1", InstanceType: "m5.large", CPUCores: 4, RAMBytes: 16 * GiB, CPUCostPerHr: 0.1, RAMCostPerGiBHr: 0.01, Labels: map[string]string{}}},
		Namespaces: []NamespaceSpec{{
			Name: "app", UID: "ns1",
			Pods: []PodSpec{{
				Name: "short", UID: "p1", NodeName: "n1", Duration: 5 * time.Minute,
				Containers: []ContainerSpec{{Name: "c1", CPURequest: 0.5, RAMRequest: 512 * MiB, CPULimit: 1, RAMLimit: 1 * GiB, CPUUsage: 0.3, RAMUsage: 256 * MiB}},
				PVCRefs:    []PodPVCRef{{ClaimName: "pvc-data", VolumeName: "vol"}},
				Labels:     map[string]string{},
			}},
			Labels: map[string]string{},
		}},
		PVs:  []PVSpec{{Name: "pv-data", UID: "pv1", ProviderID: "vol-1", StorageClass: "gp3", CapacityBytes: 100 * GiB, CostPerGiBHr: 0.0003}},
		PVCs: []PVCSpec{{Name: "pvc-data", UID: "pvc1", Namespace: "app", NamespaceUID: "ns1", PVName: "pv-data", PVUID: "pv1", StorageClass: "gp3", CapacityBytes: 100 * GiB}},
	}
	gen := NewGenerator(spec, 1*time.Minute)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tick := start.Add(10 * time.Minute)
	metrics := gen.createMetricsForTimePoint(start, tick, activePodsAtTick(gen, start, tick))
	assertMetricCount(t, metrics, metric.KubePersistentVolumeClaimInfo, 0, "PVC after pod dead")
	assertMetricCount(t, metrics, metric.KubecostPVInfo, 1, "PV still present")
}


func checkLabel(t *testing.T, m metric.Update, key, want, ctx string) {
	t.Helper()
	if got := m.Labels[key]; got != want {
		t.Errorf("%s: label %s=%q, want %q", ctx, key, got, want)
	}
}

func buildSpec(t *testing.T, cfg RandomClusterSpecConfig) ClusterSpec {
	t.Helper()
	return NewClusterBuilder(cfg).Build()
}

func defaultTestConfig() RandomClusterSpecConfig {
	return RandomClusterSpecConfig{
		NumNodes: 2, PodsPerNode: 2, NumNamespaces: 2,
		NumPVs: 2, NumLoadBalancers: 1, Seed: 42, PVCRatio: 0.5,
	}
}

func buildSpecWithPVC(t *testing.T) ClusterSpec {
	t.Helper()
	return buildSpec(t, RandomClusterSpecConfig{
		NumNodes: 1, PodsPerNode: 2, NumNamespaces: 1,
		NumPVs: 2, Seed: 100, PVCRatio: 1.0,
	})
}

func forcePodsAlwaysAlive(spec *ClusterSpec) {
	for i := range spec.Namespaces {
		for j := range spec.Namespaces[i].Pods {
			spec.Namespaces[i].Pods[j].StartOffset = 0
			spec.Namespaces[i].Pods[j].Duration = 0
		}
	}
}

func generateOneTick(t *testing.T, spec ClusterSpec) []metric.Update {
	t.Helper()
	gen := NewGenerator(spec, 1*time.Minute)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tick := start.Add(5 * time.Minute)
	return gen.createMetricsForTimePoint(start, tick, activePodsAtTick(gen, start, tick))
}

func countAllPods(spec ClusterSpec) int {
	n := 0
	for _, ns := range spec.Namespaces {
		n += len(ns.Pods)
	}
	return n
}

func countReferencedPVCs(spec ClusterSpec) int {
	seen := make(map[string]bool)
	pvcByName := make(map[string]PVCSpec)
	for _, pvc := range spec.PVCs {
		pvcByName[pvc.Namespace+"/"+pvc.Name] = pvc
	}
	for _, ns := range spec.Namespaces {
		for _, pod := range ns.Pods {
			for _, ref := range pod.PVCRefs {
				if pvc, ok := pvcByName[ns.Name+"/"+ref.ClaimName]; ok {
					seen[pvc.UID] = true
				}
			}
		}
	}
	return len(seen)
}

func assertMetricCount(t *testing.T, metrics []metric.Update, name string, want int, ctx string) {
	t.Helper()
	got := 0
	for _, m := range metrics {
		if m.Name == name {
			got++
		}
	}
	if got != want {
		t.Errorf("%s: %s count = %d, want %d", ctx, name, got, want)
	}
}

func assertMetricLabel(t *testing.T, metrics []metric.Update, name, labelKey, labelVal, ctx string) {
	t.Helper()
	for _, m := range metrics {
		if m.Name == name && m.Labels[labelKey] == labelVal {
			return
		}
	}
	t.Errorf("%s: no %s with %s=%q", ctx, name, labelKey, labelVal)
}

func assertMetricWithLabel(t *testing.T, metrics []metric.Update, name, labelKey, labelVal, ctx string) {
	t.Helper()
	assertMetricLabel(t, metrics, name, labelKey, labelVal, ctx)
}

func findMetricWithLabel(metrics []metric.Update, name, labelKey, labelVal string) *metric.Update {
	for i := range metrics {
		if metrics[i].Name == name && metrics[i].Labels[labelKey] == labelVal {
			return &metrics[i]
		}
	}
	return nil
}

func findAllMetrics(metrics []metric.Update, name string) []metric.Update {
	var result []metric.Update
	for _, m := range metrics {
		if m.Name == name {
			result = append(result, m)
		}
	}
	return result
}
