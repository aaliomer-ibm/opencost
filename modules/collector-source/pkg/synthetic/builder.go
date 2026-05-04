package synthetic

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

const permanentClaim = time.Duration(math.MaxInt64)

type ClusterBuilder struct {
	config RandomClusterSpecConfig
	faker  *gofakeit.Faker
}

func NewClusterBuilder(config RandomClusterSpecConfig) *ClusterBuilder {
	applyConfigDefaults(&config)
	return &ClusterBuilder{
		config: config,
		faker:  gofakeit.New(config.Seed),
	}
}

func applyConfigDefaults(c *RandomClusterSpecConfig) {
	if c.ScrapeInterval == 0 {
		c.ScrapeInterval = 10 * time.Minute
	}
	if c.TotalDuration == 0 {
		c.TotalDuration = 36 * time.Hour
	}
	if c.MinPodLifetime == 0 {
		c.MinPodLifetime = DefaultMinPodLifetime
	}
	if c.MaxPodLifetime == 0 {
		c.MaxPodLifetime = DefaultMaxPodLifetime
	}
	if c.PVCRatio == 0 {
		c.PVCRatio = DefaultPVCRatio
	}
	if c.PodMigrationProb == 0 {
		c.PodMigrationProb = DefaultPodMigrationProb
	}
}

func (b *ClusterBuilder) Build() ClusterSpec {
	spec := ClusterSpec{
		ClusterUID:  b.faker.UUID(),
		ClusterName: fmt.Sprintf("cluster-%s-%s", b.faker.Adjective(), b.faker.Noun()),
		Provider:    "AWS",
		Region:      "us-east-1",
	}

	spec.Nodes = b.createNodes()
	spec.Namespaces = b.createNamespaces()
	spec.Services = b.createServices(spec.Namespaces)
	b.createPods(&spec)
	spec.PVs = b.createPVs(b.config.NumPVs)
	b.assignStorage(&spec)

	return spec
}

func (b *ClusterBuilder) createNodes() []NodeSpec {
	instanceType := b.faker.RandomString(instanceTypes)
	cpuCores, ramGiB := instanceTypeResources(instanceType)
	isSpot := b.faker.Float64() < SpotNodeProbability
	cpuCostPerHr := b.faker.Float64Range(0.04, 0.16)
	ramCostPerGiBHr := b.faker.Float64Range(0.006, 0.021)

	var gpuCount float64
	var gpuType string
	var gpuCostPerHr float64
	if b.faker.Float64() < DefaultGPUProbability {
		gpuCount = float64(b.faker.IntRange(DefaultMinGPUCount, DefaultMaxGPUCount))
		gpuType = "nvidia-tesla-t4"
		gpuCostPerHr = b.faker.Float64Range(0.3, 1.0)
	}

	nodes := make([]NodeSpec, b.config.NumNodes)
	for i := range nodes {
		nodes[i] = NodeSpec{
			Name:            fmt.Sprintf("ip-%s.ec2.internal", b.faker.IPv4Address()),
			UID:             b.faker.UUID(),
			ProviderID:      fmt.Sprintf("aws:///us-east-1%s/i-%s", b.faker.RandomString(azSuffixes), b.faker.UUID()[:12]),
			InstanceType:    instanceType,
			CPUCores:        cpuCores,
			RAMBytes:        ramGiB * GiB,
			IsSpot:          isSpot,
			CPUCostPerHr:    cpuCostPerHr,
			RAMCostPerGiBHr: ramCostPerGiBHr,
			GPUCount:        gpuCount,
			GPUType:         gpuType,
			GPUCostPerHr:    gpuCostPerHr,
			Labels:          randLabelMap(b.faker, b.config.LabelsPerNode),
		}
	}
	return nodes
}

func (b *ClusterBuilder) createNamespaces() []NamespaceSpec {
	usedNames := make(map[string]bool)
	namespaces := make([]NamespaceSpec, b.config.NumNamespaces)

	for i := range namespaces {
		name := b.faker.RandomString(namespaceNames)
		for usedNames[name] {
			name = fmt.Sprintf("%s-%d", b.faker.RandomString(namespaceNames), b.faker.IntRange(0, 99))
		}
		usedNames[name] = true

		namespaces[i] = NamespaceSpec{
			Name:   name,
			UID:    b.faker.UUID(),
			Labels: randLabelMap(b.faker, b.config.LabelsPerNamespace),
		}
	}
	return namespaces
}

func (b *ClusterBuilder) createServices(namespaces []NamespaceSpec) []ServiceSpec {
	var services []ServiceSpec

	for i := range b.config.NumLoadBalancers {
		services = append(services, ServiceSpec{
			Name:      fmt.Sprintf("lb-%s-%s", b.faker.RandomString(workloadPrefixes), b.faker.UUID()[:8]),
			UID:       b.faker.UUID(),
			Namespace: namespaces[i%b.config.NumNamespaces].Name,
			Type:      "LoadBalancer",
			CostPerHr: b.faker.Float64Range(0.015, 0.055),
			IP:        b.faker.IPv4Address(),
		})
	}

	for i := range b.config.NumNamespaces {
		services = append(services, ServiceSpec{
			Name:      fmt.Sprintf("svc-%s-%s", b.faker.RandomString(workloadPrefixes), b.faker.UUID()[:8]),
			UID:       b.faker.UUID(),
			Namespace: namespaces[i].Name,
			Type:      "ClusterIP",
		})
	}

	return services
}

func (b *ClusterBuilder) createPods(spec *ClusterSpec) {
	totalPods := b.config.NumNodes * b.config.PodsPerNode
	gpuUsedPerNode := make([]float64, b.config.NumNodes)

	for podIdx := range totalPods {
		nsIdx := podIdx % b.config.NumNamespaces
		nodeIdx := podIdx % b.config.NumNodes
		node := spec.Nodes[nodeIdx]
		gpuRemaining := node.GPUCount - gpuUsedPerNode[nodeIdx]
		pod := b.createPod(node, spec.Nodes, podIdx, gpuRemaining)
		if len(pod.Containers) > 0 && pod.Containers[0].GPURequest > 0 {
			gpuUsedPerNode[nodeIdx] += pod.Containers[0].GPURequest
		}
		spec.Namespaces[nsIdx].Pods = append(spec.Namespaces[nsIdx].Pods, pod)
	}
}

func (b *ClusterBuilder) createPod(node NodeSpec, allNodes []NodeSpec, podIndex int, gpuRemaining float64) PodSpec {
	workload := b.faker.RandomString(workloadPrefixes)
	suffix := fmt.Sprintf("%s-%s", b.faker.Adjective(), b.faker.Noun())

	containerCount := 1
	if podIndex%2 == 0 {
		containerCount = 2
	}

	startOffset, duration := b.randomPodLifecycle()
	migrationTime, migrationNode := b.randomPodMigration(node, allNodes, duration, startOffset)

	return PodSpec{
		Name:          fmt.Sprintf("%s-%s-%s", workload, suffix, b.faker.UUID()[:8]),
		UID:           b.faker.UUID(),
		NodeName:      node.Name,
		Containers:    b.createContainers(node, containerCount, gpuRemaining),
		OwnerKind:     b.faker.RandomString(ownerKinds),
		OwnerName:     fmt.Sprintf("%s-%s", workload, suffix),
		OwnerUID:      b.faker.UUID(),
		Labels:        randLabelMap(b.faker, b.config.LabelsPerPod),
		StartOffset:   startOffset,
		Duration:      duration,
		MigrationTime: migrationTime,
		MigrationNode: migrationNode,
	}
}

func (b *ClusterBuilder) createContainers(node NodeSpec, count int, gpuRemaining float64) []ContainerSpec {
	cpuBudget := (node.CPUCores / float64(b.config.PodsPerNode)) * 1.1
	ramBudget := (node.RAMBytes / float64(b.config.PodsPerNode)) * 1.1

	containers := make([]ContainerSpec, count)
	for idx := range containers {
		containers[idx] = b.createContainer(idx, cpuBudget/float64(count), ramBudget/float64(count), gpuRemaining)
		if containers[idx].GPURequest > 0 {
			gpuRemaining -= containers[idx].GPURequest
		}
	}
	return containers
}

func (b *ClusterBuilder) createContainer(containerIndex int, cpuBudget, ramBudget, gpuRemaining float64) ContainerSpec {
	imageName := b.faker.RandomString(containerImages)
	if containerIndex > 0 {
		imageName = b.faker.RandomString(sidecarImages)
	}

	cpuRequest := b.faker.Float64Range(0.1, cpuBudget)
	ramRequest := b.faker.Float64Range(128*MiB, ramBudget)
	cpuUsage := cpuRequest * b.faker.Float64Range(MinCPUUsageRatio, MaxCPUUsageRatio)
	ramUsage := ramRequest * b.faker.Float64Range(MinRAMUsageRatio, MaxRAMUsageRatio)

	spec := ContainerSpec{
		Name:       imageName,
		CPURequest: cpuRequest,
		RAMRequest: ramRequest,
		CPULimit:   cpuRequest * b.faker.Float64Range(MinCPULimitMultiplier, MaxCPULimitMultiplier),
		RAMLimit:   ramRequest * b.faker.Float64Range(MinRAMLimitMultiplier, MaxRAMLimitMultiplier),
		CPUUsage:   cpuUsage,
		RAMUsage:   ramUsage,
	}

	if containerIndex == 0 && gpuRemaining >= 1.0 && b.faker.Float64() < ContainerGPUAssignmentProb {
		spec.GPURequest = 1.0
	}

	return spec
}

func (b *ClusterBuilder) randomPodLifecycle() (startOffset, duration time.Duration) {
	lifetimeRange := b.config.MaxPodLifetime - b.config.MinPodLifetime
	duration = b.config.MinPodLifetime + time.Duration(b.faker.Float64()*float64(lifetimeRange))

	maxStart := b.config.TotalDuration - duration
	if maxStart > 0 {
		startOffset = time.Duration(b.faker.Float64() * float64(maxStart))
	}
	return startOffset, duration
}

func (b *ClusterBuilder) randomPodMigration(currentNode NodeSpec, allNodes []NodeSpec, lifetime, startOffset time.Duration) (*time.Duration, string) {
	if b.faker.Float64() >= b.config.PodMigrationProb || len(allNodes) <= 1 {
		return nil, ""
	}

	migrationOffset := startOffset + lifetime/4 + time.Duration(b.faker.Float64()*float64(lifetime/2))

	for {
		candidate := allNodes[b.faker.IntRange(0, len(allNodes)-1)]
		if candidate.Name != currentNode.Name {
			return &migrationOffset, candidate.Name
		}
	}
}

func (b *ClusterBuilder) createPVs(count int) []PVSpec {
	if count == 0 {
		return nil
	}

	storageClass := b.faker.RandomString(storageClasses)
	capacityBytes := float64(b.faker.IntRange(10, 200)) * GiB
	costPerGiBHr := b.faker.Float64Range(0.0001, 0.0006)

	pvs := make([]PVSpec, count)
	for i := range pvs {
		pvs[i] = PVSpec{
			Name:          fmt.Sprintf("pv-%s-%s", b.faker.Noun(), b.faker.UUID()[:8]),
			UID:           b.faker.UUID(),
			ProviderID:    fmt.Sprintf("vol-%s", b.faker.UUID()[:12]),
			StorageClass:  storageClass,
			CapacityBytes: capacityBytes,
			CostPerGiBHr:  costPerGiBHr,
		}
	}
	return pvs
}

type podLocation struct {
	NamespaceIdx int
	PodIdx       int
}

type storageCandidate struct {
	loc         podLocation
	startOffset time.Duration
	endOffset   time.Duration
}

func (b *ClusterBuilder) assignStorage(spec *ClusterSpec) {
	if len(spec.PVs) == 0 {
		return
	}

	candidates := b.collectStorageCandidates(spec)
	if len(candidates) == 0 {
		return
	}

	pvFreeAt := make([]time.Duration, len(spec.PVs))

	for _, candidate := range candidates {
		pvIdx := findFreePV(pvFreeAt, candidate.startOffset)
		if pvIdx < 0 {
			continue
		}
		pvFreeAt[pvIdx] = candidate.endOffset

		ns := spec.Namespaces[candidate.loc.NamespaceIdx]
		pvc := b.createPVC(spec.PVs[pvIdx], ns.Name, ns.UID)
		spec.PVCs = append(spec.PVCs, pvc)

		pod := &spec.Namespaces[candidate.loc.NamespaceIdx].Pods[candidate.loc.PodIdx]
		pod.PVCRefs = append(pod.PVCRefs, PodPVCRef{
			ClaimName:  pvc.Name,
			VolumeName: fmt.Sprintf("vol-%s", b.faker.UUID()[:8]),
		})
	}
}

func (b *ClusterBuilder) collectStorageCandidates(spec *ClusterSpec) []storageCandidate {
	var candidates []storageCandidate
	for nsIdx, ns := range spec.Namespaces {
		for podIdx, pod := range ns.Pods {
			if b.faker.Float64() >= b.config.PVCRatio {
				continue
			}
			endOffset := permanentClaim
			if pod.Duration > 0 {
				endOffset = pod.StartOffset + pod.Duration
			}
			candidates = append(candidates, storageCandidate{
				loc:         podLocation{nsIdx, podIdx},
				startOffset: pod.StartOffset,
				endOffset:   endOffset,
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].startOffset < candidates[j].startOffset
	})
	return candidates
}

func findFreePV(pvFreeAt []time.Duration, podStart time.Duration) int {
	for i, freeAt := range pvFreeAt {
		if freeAt <= podStart {
			return i
		}
	}
	return -1
}

func (b *ClusterBuilder) createPVC(pv PVSpec, namespace, namespaceUID string) PVCSpec {
	return PVCSpec{
		Name:          fmt.Sprintf("pvc-%s-%s", b.faker.Noun(), b.faker.UUID()[:8]),
		UID:           b.faker.UUID(),
		Namespace:     namespace,
		NamespaceUID:  namespaceUID,
		PVName:        pv.Name,
		PVUID:         pv.UID,
		StorageClass:  pv.StorageClass,
		CapacityBytes: pv.CapacityBytes,
	}
}
