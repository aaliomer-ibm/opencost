package synthetic

import (
	"maps"
	"time"

	"github.com/opencost/opencost/core/pkg/log"
	"github.com/opencost/opencost/core/pkg/source"
	"github.com/opencost/opencost/modules/collector-source/pkg/metric"
)

type MetricContext struct {
	Namespace       NamespaceSpec
	Pod             PodSpec
	Container       ContainerSpec
	GenerationStart time.Time
	CurrentTime     time.Time
}

type activePod struct {
	NamespaceIdx int
	PodIdx       int
}

func (g *Generator) Generate(metricsRepo metric.Updater, startTime, endTime time.Time) {
	activePodsByTick := g.buildActivePodMap(startTime, endTime)

	tickCount := 0
	for currentTime := startTime; currentTime.Before(endTime); currentTime = currentTime.Add(g.interval) {
		activeAtTick := activePodsByTick[currentTime]
		metricsAtThisTime := g.createMetricsForTimePoint(startTime, currentTime, activeAtTick)
		metricsRepo.Update(&metric.UpdateSet{
			Timestamp: currentTime,
			Updates:   metricsAtThisTime,
		})
		tickCount++
	}
	log.Infof("Synthetic generator: pushed %d ticks from %s to %s", tickCount, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
}

func (g *Generator) buildActivePodMap(startTime, endTime time.Time) map[time.Time][]activePod {
	result := make(map[time.Time][]activePod)

	for nsIdx, ns := range g.spec.Namespaces {
		for podIdx, pod := range ns.Pods {
			entry := activePod{NamespaceIdx: nsIdx, PodIdx: podIdx}

			if pod.Duration == 0 {
				for t := startTime; t.Before(endTime); t = t.Add(g.interval) {
					result[t] = append(result[t], entry)
				}
				continue
			}

			podStart := startTime.Add(pod.StartOffset)
			podEnd := podStart.Add(pod.Duration)

			firstTick := startTime
			if podStart.After(startTime) {
				elapsed := podStart.Sub(startTime)
				intervals := elapsed / g.interval
				firstTick = startTime.Add(intervals * g.interval)
				if firstTick.Before(podStart) {
					firstTick = firstTick.Add(g.interval)
				}
			}

			for t := firstTick; t.Before(podEnd) && t.Before(endTime); t = t.Add(g.interval) {
				result[t] = append(result[t], entry)
			}
		}
	}

	return result
}

func isPodRunningAt(pod PodSpec, generationStart, checkTime time.Time) bool {
	if pod.Duration == 0 {
		return true
	}
	podStartTime := generationStart.Add(pod.StartOffset)
	podEndTime := podStartTime.Add(pod.Duration)
	return !checkTime.Before(podStartTime) && checkTime.Before(podEndTime)
}

func getPodNodeAt(pod PodSpec, generationStart, checkTime time.Time) string {
	if pod.MigrationTime == nil {
		return pod.NodeName
	}

	migrationTime := generationStart.Add(*pod.MigrationTime)
	if checkTime.Before(migrationTime) {
		return pod.NodeName
	}
	return pod.MigrationNode
}

func (g *Generator) createMetricsForTimePoint(generationStart, currentTime time.Time, activeAtTick []activePod) []metric.Update {
	var metrics []metric.Update

	metrics = append(metrics, g.createClusterMetrics()...)
	metrics = append(metrics, g.createNodeMetrics()...)
	metrics = append(metrics, g.createNamespaceAndPodMetrics(generationStart, currentTime, activeAtTick)...)
	metrics = append(metrics, g.createPersistentVolumeMetrics()...)
	metrics = append(metrics, g.createServiceMetrics()...)

	return metrics
}

func (g *Generator) createNodeMetrics() []metric.Update {
	var metrics []metric.Update
	for _, node := range g.spec.Nodes {
		metrics = append(metrics, g.createNodeMetric(node)...)
	}
	return metrics
}

func (g *Generator) createNamespaceAndPodMetrics(generationStart, currentTime time.Time, activeAtTick []activePod) []metric.Update {
	var metrics []metric.Update
	emittedPVCs := make(map[string]bool)

	for _, namespace := range g.spec.Namespaces {
		metrics = append(metrics, g.createNamespaceMetric(namespace)...)
	}

	for _, ap := range activeAtTick {
		namespace := g.spec.Namespaces[ap.NamespaceIdx]
		pod := namespace.Pods[ap.PodIdx]

		ctx := MetricContext{
			Namespace:       namespace,
			Pod:             pod,
			GenerationStart: generationStart,
			CurrentTime:     currentTime,
		}

		metrics = append(metrics, g.createPodMetric(ctx)...)

		for _, container := range pod.Containers {
			ctx.Container = container
			metrics = append(metrics, g.createContainerMetric(ctx)...)
		}

		for _, ref := range pod.PVCRefs {
			pvc := g.resolvePVCByName(ref.ClaimName, namespace.Name)
			if pvc == nil || emittedPVCs[pvc.UID] {
				continue
			}
			emittedPVCs[pvc.UID] = true
			metrics = append(metrics, g.createPVCMetrics(*pvc)...)
		}
	}

	return metrics
}

func (g *Generator) createPersistentVolumeMetrics() []metric.Update {
	var metrics []metric.Update
	for _, pv := range g.spec.PVs {
		metrics = append(metrics, g.createPVMetrics(pv)...)
	}
	return metrics
}

func (g *Generator) createPVMetrics(pv PVSpec) []metric.Update {
	pvLabels := map[string]string{
		source.UIDLabel:          pv.UID,
		source.PVLabel:           pv.Name,
		source.StorageClassLabel: pv.StorageClass,
		source.ProviderIDLabel:   pv.ProviderID,
	}
	return []metric.Update{
		{Name: metric.KubecostPVInfo, Labels: pvLabels, Value: 0, AdditionalInfo: pvLabels},
		{Name: metric.KubePersistentVolumeCapacityBytes, Labels: pvLabels, Value: pv.CapacityBytes},
		{Name: metric.PVHourlyCost, Labels: merge(pvLabels, map[string]string{source.VolumeNameLabel: pv.Name}), Value: pv.CostPerGiBHr},
	}
}

func (g *Generator) createPVCMetrics(pvc PVCSpec) []metric.Update {
	pvcLabels := map[string]string{
		source.UIDLabel:          pvc.UID,
		source.PVCLabel:          pvc.Name,
		source.NamespaceUIDLabel: pvc.NamespaceUID,
		source.NamespaceLabel:    pvc.Namespace,
		source.VolumeNameLabel:   pvc.PVName,
		source.PVUIDLabel:        pvc.PVUID,
		source.StorageClassLabel: pvc.StorageClass,
	}
	return []metric.Update{
		{Name: metric.KubePersistentVolumeClaimInfo, Labels: pvcLabels, Value: 0, AdditionalInfo: pvcLabels},
		{Name: metric.KubePersistentVolumeClaimResourceRequestsStorageBytes, Labels: pvcLabels, Value: pvc.CapacityBytes},
	}
}

func (g *Generator) createServiceMetrics() []metric.Update {
	var metrics []metric.Update
	for _, service := range g.spec.Services {
		metrics = append(metrics, g.createServiceMetric(service)...)
	}
	return metrics
}

func (g *Generator) createClusterMetrics() []metric.Update {
	clusterLabels := map[string]string{source.UIDLabel: g.spec.ClusterUID}
	return []metric.Update{
		{Name: metric.ClusterInfo, Labels: clusterLabels, Value: 1, AdditionalInfo: clusterLabels},
		{Name: metric.KubecostNetworkZoneEgressCost, Labels: clusterLabels, Value: 0.01},
		{Name: metric.KubecostNetworkRegionEgressCost, Labels: clusterLabels, Value: 0.02},
		{Name: metric.KubecostNetworkInternetEgressCost, Labels: clusterLabels, Value: 0.12},
	}
}

func (g *Generator) createNodeMetric(n NodeSpec) []metric.Update {
	nl := map[string]string{
		source.NodeLabel:         n.Name,
		source.UIDLabel:          n.UID,
		source.ProviderIDLabel:   n.ProviderID,
		source.InstanceTypeLabel: n.InstanceType,
	}
	u := []metric.Update{
		{Name: metric.NodeInfo, Labels: nl, Value: 1, AdditionalInfo: nl},
		{Name: metric.KubeNodeStatusCapacityCPUCores, Labels: nl, Value: n.CPUCores},
		{Name: metric.KubeNodeStatusCapacityMemoryBytes, Labels: nl, Value: n.RAMBytes},
		{Name: metric.KubeNodeStatusAllocatableCPUCores, Labels: nl, Value: n.CPUCores},
		{Name: metric.KubeNodeStatusAllocatableMemoryBytes, Labels: nl, Value: n.RAMBytes},
		{Name: metric.KubeNodeLabels, Labels: nl, Value: 0, AdditionalInfo: safeLabels(n.Labels)},
	}
	for _, res := range []struct {
		r, unit string
		v       float64
	}{
		{"cpu", "core", n.CPUCores}, {"memory", "byte", n.RAMBytes},
	} {
		extra := map[string]string{source.ResourceLabel: res.r, source.UnitLabel: res.unit}
		u = append(u, metric.Update{Name: metric.NodeResourceCapacities, Labels: merge(nl, extra), Value: res.v})
		u = append(u, metric.Update{Name: metric.NodeResourcesAllocatable, Labels: merge(nl, extra), Value: res.v})
	}
	if n.GPUCount > 0 {
		u = append(u, metric.Update{Name: metric.NodeGPUCount, Labels: nl, Value: n.GPUCount})
		u = append(u, metric.Update{Name: metric.NodeGPUHourlyCost, Labels: nl, Value: n.GPUCostPerHr})
	}
	totalCost := n.CPUCostPerHr*n.CPUCores + n.RAMCostPerGiBHr*(n.RAMBytes/1024/1024/1024) + n.GPUCostPerHr*n.GPUCount
	u = append(u,
		metric.Update{Name: metric.NodeTotalHourlyCost, Labels: nl, Value: totalCost},
		metric.Update{Name: metric.NodeCPUHourlyCost, Labels: nl, Value: n.CPUCostPerHr},
		metric.Update{Name: metric.NodeRAMHourlyCost, Labels: nl, Value: n.RAMCostPerGiBHr},
	)
	spotVal := 0.0
	if n.IsSpot {
		spotVal = 1.0
	}
	u = append(u, metric.Update{Name: metric.KubecostNodeIsSpot, Labels: nl, Value: spotVal})
	return u
}

func (g *Generator) createNamespaceMetric(ns NamespaceSpec) []metric.Update {
	l := map[string]string{source.NamespaceLabel: ns.Name, source.UIDLabel: ns.UID}
	return []metric.Update{
		{Name: metric.NamespaceInfo, Labels: l, Value: 0, AdditionalInfo: l},
		{Name: metric.KubeNamespaceLabels, Labels: l, Value: 0, AdditionalInfo: safeLabels(ns.Labels)},
		{Name: metric.KubeNamespaceAnnotations, Labels: l, Value: 0, AdditionalInfo: map[string]string{}},
	}
}

func (g *Generator) createPodMetric(ctx MetricContext) []metric.Update {
	nodeName := getPodNodeAt(ctx.Pod, ctx.GenerationStart, ctx.CurrentTime)
	nodeUID := g.resolveNodeUID(nodeName)
	pl := map[string]string{
		source.UIDLabel:          ctx.Pod.UID,
		source.PodLabel:          ctx.Pod.Name,
		source.NamespaceUIDLabel: ctx.Namespace.UID,
		source.NodeUIDLabel:      nodeUID,
		source.NamespaceLabel:    ctx.Namespace.Name,
		source.NodeLabel:         nodeName,
		source.InstanceLabel:     nodeName,
	}
	u := []metric.Update{
		{Name: metric.PodInfo, Labels: pl, Value: 0, AdditionalInfo: pl},
		{Name: metric.KubePodLabels, Labels: pl, Value: 0, AdditionalInfo: safeLabels(ctx.Pod.Labels)},
		{Name: metric.KubePodAnnotations, Labels: pl, Value: 0, AdditionalInfo: map[string]string{}},
	}
	if ctx.Pod.OwnerKind != "" {
		u = append(u, metric.Update{
			Name: metric.KubePodOwner,
			Labels: merge(pl, map[string]string{
				source.OwnerKindLabel: ctx.Pod.OwnerKind,
				source.OwnerNameLabel: ctx.Pod.OwnerName,
				source.OwnerUIDLabel:  ctx.Pod.OwnerUID,
				source.ContainerLabel: "true",
			}),
		})
	}
	for _, c := range ctx.Pod.Containers {
		cl := merge(pl, map[string]string{source.ContainerLabel: c.Name})
		u = append(u, metric.Update{Name: metric.KubePodContainerStatusRunning, Labels: cl, Value: 0, AdditionalInfo: cl})
	}
	for _, ref := range ctx.Pod.PVCRefs {
		pvcUID := g.resolvePVCUID(ref.ClaimName, ctx.Namespace.Name)
		u = append(u, metric.Update{
			Name: metric.PodPVCVolume,
			Labels: map[string]string{
				source.UIDLabel:           ctx.Pod.UID,
				source.PVCUIDLabel:        pvcUID,
				source.PodVolumeNameLabel: ref.VolumeName,
			},
		})
	}

	ticksElapsed := float64(ctx.CurrentTime.Sub(ctx.GenerationStart)/g.interval) + 1

	u = append(u, metric.Update{
		Name:   metric.ContainerNetworkTransmitBytesTotal,
		Labels: map[string]string{source.NamespaceLabel: ctx.Namespace.Name, source.PodLabel: ctx.Pod.Name, source.UIDLabel: ctx.Pod.UID},
		Value:  1024 * 1024 * ticksElapsed,
	})
	u = append(u, metric.Update{
		Name:   metric.ContainerNetworkReceiveBytesTotal,
		Labels: map[string]string{source.NamespaceLabel: ctx.Namespace.Name, source.PodLabel: ctx.Pod.Name, source.UIDLabel: ctx.Pod.UID},
		Value:  512 * 1024 * ticksElapsed,
	})

	u = append(u, g.createPodNetworkTrafficMetrics(ctx)...)

	return u
}

func (g *Generator) createPodNetworkTrafficMetrics(ctx MetricContext) []metric.Update {
	elapsed := ctx.CurrentTime.Sub(ctx.GenerationStart)
	ticksElapsed := float64(elapsed/g.interval) + 1

	egressPerTick := 1024.0 * 1024.0
	ingressPerTick := 512.0 * 1024.0
	cumulativeEgress := egressPerTick * ticksElapsed
	cumulativeIngress := ingressPerTick * ticksElapsed

	type trafficSplit struct {
		internet   string
		sameRegion string
		sameZone   string
		fraction   float64
	}

	splits := []trafficSplit{
		{internet: "false", sameRegion: "true", sameZone: "true", fraction: 0.70},
		{internet: "false", sameRegion: "true", sameZone: "false", fraction: 0.20},
		{internet: "true", sameRegion: "false", sameZone: "false", fraction: 0.10},
	}

	var u []metric.Update
	for _, s := range splits {
		labels := map[string]string{
			source.UIDLabel:        ctx.Pod.UID,
			source.NamespaceLabel:  ctx.Namespace.Name,
			source.PodNameLabel:    ctx.Pod.Name,
			source.ServiceLabel:    "",
			source.InternetLabel:   s.internet,
			source.SameRegionLabel: s.sameRegion,
			source.SameZoneLabel:   s.sameZone,
			source.NatGatewayLabel: "false",
		}
		u = append(u, metric.Update{
			Name:   metric.KubecostPodNetworkEgressBytesTotal,
			Labels: labels,
			Value:  cumulativeEgress * s.fraction,
		})
		u = append(u, metric.Update{
			Name:   metric.KubecostPodNetworkIngressBytesTotal,
			Labels: labels,
			Value:  cumulativeIngress * s.fraction,
		})
	}
	return u
}

func (g *Generator) createContainerMetric(ctx MetricContext) []metric.Update {
	nodeName := getPodNodeAt(ctx.Pod, ctx.GenerationStart, ctx.CurrentTime)
	cl := map[string]string{
		source.NodeLabel: nodeName, source.InstanceLabel: nodeName,
		source.NamespaceLabel: ctx.Namespace.Name, source.PodLabel: ctx.Pod.Name,
		source.UIDLabel: ctx.Pod.UID, source.ContainerLabel: ctx.Container.Name,
	}
	u := []metric.Update{
		{Name: metric.ContainerCPUAllocation, Labels: cl, Value: ctx.Container.CPURequest},
		{Name: metric.ContainerMemoryAllocationBytes, Labels: cl, Value: ctx.Container.RAMRequest},
		{Name: metric.ContainerCPUUsageSecondsTotal, Labels: cl, Value: ctx.Container.CPUUsage},
		{Name: metric.ContainerMemoryWorkingSetBytes, Labels: cl, Value: ctx.Container.RAMUsage},
	}
	addResourceMetric := func(metricName string, resource, unit string, value float64) {
		if value > 0 {
			u = append(u, metric.Update{
				Name: metricName, Labels: merge(cl, map[string]string{source.ResourceLabel: resource, source.UnitLabel: unit}), Value: value,
			})
		}
	}
	addResourceMetric(metric.KubePodContainerResourceRequests, "cpu", "core", ctx.Container.CPURequest)
	addResourceMetric(metric.KubePodContainerResourceRequests, "memory", "byte", ctx.Container.RAMRequest)
	addResourceMetric(metric.KubePodContainerResourceLimits, "cpu", "core", ctx.Container.CPULimit)
	addResourceMetric(metric.KubePodContainerResourceLimits, "memory", "byte", ctx.Container.RAMLimit)
	if ctx.Container.GPURequest > 0 {
		u = append(u, metric.Update{Name: metric.ContainerGPUAllocation, Labels: cl, Value: ctx.Container.GPURequest})
		addResourceMetric(metric.KubePodContainerResourceRequests, "nvidia_com_gpu", "integer", ctx.Container.GPURequest)
	}
	return u
}

func (g *Generator) createServiceMetric(svc ServiceSpec) []metric.Update {
	sl := map[string]string{
		source.UIDLabel:         svc.UID,
		source.ServiceLabel:     svc.Name,
		source.NamespaceLabel:   svc.Namespace,
		source.ServiceTypeLabel: svc.Type,
	}
	u := []metric.Update{
		{Name: metric.ServiceInfo, Labels: sl, Value: 0, AdditionalInfo: sl},
		{Name: metric.ServiceSelectorLabels, Labels: sl, Value: 0, AdditionalInfo: map[string]string{}},
	}
	if svc.Type == "LoadBalancer" && svc.CostPerHr > 0 {
		lbLabels := map[string]string{
			source.NamespaceLabel:   svc.Namespace,
			source.ServiceNameLabel: svc.Name,
			source.IngressIPLabel:   svc.IP,
			source.UIDLabel:         svc.UID,
		}
		u = append(u, metric.Update{Name: metric.KubecostLoadBalancerCost, Labels: lbLabels, Value: svc.CostPerHr})
	}
	return u
}

func (g *Generator) resolveNodeUID(nodeName string) string {
	for _, n := range g.spec.Nodes {
		if n.Name == nodeName {
			return n.UID
		}
	}
	return ""
}

func (g *Generator) resolvePVCUID(claimName, namespace string) string {
	for _, pvc := range g.spec.PVCs {
		if pvc.Name == claimName && pvc.Namespace == namespace {
			return pvc.UID
		}
	}
	return ""
}

func (g *Generator) resolvePVCByName(claimName, namespace string) *PVCSpec {
	for i := range g.spec.PVCs {
		if g.spec.PVCs[i].Name == claimName && g.spec.PVCs[i].Namespace == namespace {
			return &g.spec.PVCs[i]
		}
	}
	return nil
}

func merge(base, extra map[string]string) map[string]string {
	m := make(map[string]string, len(base)+len(extra))
	maps.Copy(m, base)
	maps.Copy(m, extra)
	return m
}

func safeLabels(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}
