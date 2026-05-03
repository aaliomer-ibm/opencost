package synthetic

import (
	"fmt"
	"testing"
	"time"

	"github.com/opencost/opencost/modules/collector-source/pkg/metric"
)


func TestBug1_PVCMetricsEmittedAfterPodEnds(t *testing.T) {
	spec := clusterWithOnePVCPod(10 * time.Minute)
	gen := NewGenerator(spec, 1*time.Minute)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tick := start.Add(15 * time.Minute)

	metrics := gen.createMetricsForTimePoint(start, tick, activePodsAtTick(gen, start, tick))

	expectCount(t, metrics, metric.KubePersistentVolumeClaimInfo, 0, "PVC info after pod ended")
	expectCount(t, metrics, metric.KubePersistentVolumeClaimResourceRequestsStorageBytes, 0, "PVC storage after pod ended")
}


func TestBug1b_PVCMetricsDuplicatedForMultiplePods(t *testing.T) {
	spec := clusterWithTwoPodsOnePVC()
	gen := NewGenerator(spec, 1*time.Minute)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tick := start.Add(5 * time.Minute)

	metrics := gen.createMetricsForTimePoint(start, tick, activePodsAtTick(gen, start, tick))

	expectCount(t, metrics, metric.KubePersistentVolumeClaimInfo, 1, "PVC dedup: two pods, one PVC")
}


func TestPreservation_PVMetricsUnconditional(t *testing.T) {
	pvs := []PVSpec{
		{Name: "pv-a", UID: "pva", ProviderID: "v1", StorageClass: "gp3", CapacityBytes: 50e9, CostPerGiBHr: 0.0003},
		{Name: "pv-b", UID: "pvb", ProviderID: "v2", StorageClass: "gp2", CapacityBytes: 100e9, CostPerGiBHr: 0.0005},
	}
	spec := ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{baseNode()},
		Namespaces: []NamespaceSpec{{
			Name: "ns1", UID: "ns1",
			Pods: []PodSpec{{
				Name: "p1", UID: "p1", NodeName: "node-1",
				Containers: []ContainerSpec{baseContainer()},
				Duration:   5 * time.Minute, Labels: map[string]string{},
			}},
		}},
		PVs: pvs,
	}

	gen := NewGenerator(spec, 1*time.Minute)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, offset := range []int{0, 3, 10, 30} {
		tick := start.Add(time.Duration(offset) * time.Minute)
		metrics := gen.createMetricsForTimePoint(start, tick, activePodsAtTick(gen, start, tick))

		for _, pv := range pvs {
			tag := fmt.Sprintf("tick=%dm pv=%s", offset, pv.UID)
			requireMetricForPV(t, metrics, metric.KubecostPVInfo, pv, tag)
			requireMetricForPV(t, metrics, metric.KubePersistentVolumeCapacityBytes, pv, tag)
			requireMetricForPV(t, metrics, metric.PVHourlyCost, pv, tag)
		}
	}
}


func TestPreservation_NodeNamespaceClusterServiceMetrics(t *testing.T) {
	spec := ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{
			{Name: "n1", UID: "n1", ProviderID: "i-1", InstanceType: "m5.large", CPUCores: 4, RAMBytes: 16e9, CPUCostPerHr: 0.1, RAMCostPerGiBHr: 0.01, Labels: map[string]string{}},
			{Name: "n2", UID: "n2", ProviderID: "i-2", InstanceType: "m5.xlarge", CPUCores: 8, RAMBytes: 32e9, CPUCostPerHr: 0.2, RAMCostPerGiBHr: 0.02, Labels: map[string]string{}},
		},
		Namespaces: []NamespaceSpec{
			{Name: "default", UID: "ns1", Labels: map[string]string{}},
			{Name: "system", UID: "ns2", Labels: map[string]string{}},
		},
		Services: []ServiceSpec{
			{Name: "api", UID: "s1", Namespace: "default", Type: "ClusterIP"},
			{Name: "lb", UID: "s2", Namespace: "default", Type: "LoadBalancer", CostPerHr: 0.025, IP: "10.0.0.1"},
		},
	}

	gen := NewGenerator(spec, 1*time.Minute)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, offset := range []int{0, 5, 15} {
		tick := start.Add(time.Duration(offset) * time.Minute)
		metrics := gen.createMetricsForTimePoint(start, tick, activePodsAtTick(gen, start, tick))
		counts := metricNameCounts(metrics)
		tag := fmt.Sprintf("tick=%dm", offset)

		expectMinCount(t, counts, metric.ClusterInfo, 1, tag)
		expectMinCount(t, counts, metric.NodeInfo, 2, tag)
		expectMinCount(t, counts, metric.NamespaceInfo, 2, tag)
		expectMinCount(t, counts, metric.ServiceInfo, 2, tag)
		expectMinCount(t, counts, metric.KubecostLoadBalancerCost, 1, tag)
	}
}


func TestPreservation_PodContainerMetricsForRunningPods(t *testing.T) {
	spec := ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{baseNode()},
		Namespaces: []NamespaceSpec{{
			Name: "prod", UID: "ns1",
			Pods: []PodSpec{{
				Name: "pod-lr", UID: "plr", NodeName: "node-1",
				Containers: []ContainerSpec{
					{Name: "web", CPURequest: 1, RAMRequest: 1e9, CPULimit: 2, RAMLimit: 2e9, CPUUsage: 0.8, RAMUsage: 800e6},
					{Name: "side", CPURequest: 0.25, RAMRequest: 256e6, CPULimit: 0.5, RAMLimit: 512e6, CPUUsage: 0.1, RAMUsage: 100e6},
				},
				Duration: 0, Labels: map[string]string{},
			}},
		}},
	}

	gen := NewGenerator(spec, 1*time.Minute)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tick := start.Add(10 * time.Minute)
	metrics := gen.createMetricsForTimePoint(start, tick, activePodsAtTick(gen, start, tick))
	counts := metricNameCounts(metrics)

	expectCount(t, metrics, metric.PodInfo, 1, "PodInfo")
	expectCount(t, metrics, metric.ContainerCPUAllocation, 2, "ContainerCPUAllocation")
	expectCount(t, metrics, metric.ContainerMemoryAllocationBytes, 2, "ContainerMemoryAllocationBytes")
	expectCount(t, metrics, metric.ContainerNetworkTransmitBytesTotal, 1, "NetworkTx")
	expectCount(t, metrics, metric.ContainerNetworkReceiveBytesTotal, 1, "NetworkRx")
	expectMinCount(t, counts, metric.KubePodContainerResourceRequests, 4, "resource requests")
	expectMinCount(t, counts, metric.KubePodContainerResourceLimits, 4, "resource limits")
}


func baseNode() NodeSpec {
	return NodeSpec{
		Name: "node-1", UID: "n1", ProviderID: "aws:///i-1",
		InstanceType: "m5.large", CPUCores: 4, RAMBytes: 16e9,
		CPUCostPerHr: 0.1, RAMCostPerGiBHr: 0.01, Labels: map[string]string{},
	}
}

func baseContainer() ContainerSpec {
	return ContainerSpec{
		Name: "app", CPURequest: 0.5, RAMRequest: 512e6,
		CPULimit: 1, RAMLimit: 1e9, CPUUsage: 0.3, RAMUsage: 256e6,
	}
}

func clusterWithOnePVCPod(podDuration time.Duration) ClusterSpec {
	return ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{baseNode()},
		Namespaces: []NamespaceSpec{{
			Name: "prod", UID: "ns-uid-prod",
			Pods: []PodSpec{{
				Name: "worker", UID: "pod-1", NodeName: "node-1",
				Containers: []ContainerSpec{baseContainer()},
				PVCRefs:    []PodPVCRef{{ClaimName: "pvc-data", VolumeName: "data-vol"}},
				Duration:   podDuration, Labels: map[string]string{},
			}},
		}},
		PVs: []PVSpec{{
			Name: "pv-data", UID: "pv-uid-1", ProviderID: "vol-1", StorageClass: "gp3",
			CapacityBytes: 100e9, CostPerGiBHr: 0.0003,
		}},
		PVCs: []PVCSpec{{
			Name: "pvc-data", UID: "pvc-uid-1", Namespace: "prod", NamespaceUID: "ns-uid-prod",
			PVName: "pv-data", PVUID: "pv-uid-1", StorageClass: "gp3", CapacityBytes: 100e9,
		}},
	}
}

func clusterWithTwoPodsOnePVC() ClusterSpec {
	return ClusterSpec{
		ClusterUID: "c1", ClusterName: "test", Provider: "AWS", Region: "us-east-1",
		Nodes: []NodeSpec{baseNode()},
		Namespaces: []NamespaceSpec{{
			Name: "prod", UID: "ns-uid-prod",
			Pods: []PodSpec{
				{
					Name: "pod-1", UID: "p1", NodeName: "node-1",
					Containers: []ContainerSpec{baseContainer()},
					PVCRefs:    []PodPVCRef{{ClaimName: "pvc-data", VolumeName: "data-vol"}},
					Duration:   0, Labels: map[string]string{},
				},
				{
					Name: "pod-2", UID: "p2", NodeName: "node-1",
					Containers: []ContainerSpec{{Name: "side", CPURequest: 0.25, RAMRequest: 256e6, CPULimit: 0.5, RAMLimit: 512e6, CPUUsage: 0.1, RAMUsage: 128e6}},
					PVCRefs:    []PodPVCRef{{ClaimName: "pvc-data", VolumeName: "data-vol"}},
					Duration:   0, Labels: map[string]string{},
				},
			},
		}},
		PVs: []PVSpec{{
			Name: "pv-data", UID: "pv-uid-1", ProviderID: "vol-1", StorageClass: "gp3",
			CapacityBytes: 100e9, CostPerGiBHr: 0.0003,
		}},
		PVCs: []PVCSpec{{
			Name: "pvc-data", UID: "pvc-uid-1", Namespace: "prod", NamespaceUID: "ns-uid-prod",
			PVName: "pv-data", PVUID: "pv-uid-1", StorageClass: "gp3", CapacityBytes: 100e9,
		}},
	}
}

func countMetricsByName(metrics []metric.Update, name string) int {
	n := 0
	for _, m := range metrics {
		if m.Name == name {
			n++
		}
	}
	return n
}

func metricNameCounts(metrics []metric.Update) map[string]int {
	counts := make(map[string]int)
	for _, m := range metrics {
		counts[m.Name]++
	}
	return counts
}

func expectCount(t *testing.T, metrics []metric.Update, name string, want int, ctx string) {
	t.Helper()
	got := countMetricsByName(metrics, name)
	if got != want {
		t.Errorf("%s: %s count = %d, want %d", ctx, name, got, want)
	}
}

func expectMinCount(t *testing.T, counts map[string]int, name string, min int, ctx string) {
	t.Helper()
	if counts[name] < min {
		t.Errorf("%s: %s count = %d, want >= %d", ctx, name, counts[name], min)
	}
}

func requireMetricForPV(t *testing.T, metrics []metric.Update, metricName string, pv PVSpec, ctx string) {
	t.Helper()
	for _, m := range metrics {
		if m.Name == metricName && (m.Labels["uid"] == pv.UID || m.Labels["persistentvolume"] == pv.Name) {
			return
		}
	}
	t.Fatalf("%s: missing %s for PV %s", ctx, metricName, pv.UID)
}

func activePodsAtTick(gen *Generator, startTime, tickTime time.Time) []activePod {
	fullMap := gen.buildActivePodMap(startTime, tickTime.Add(gen.interval))
	return fullMap[tickTime]
}
