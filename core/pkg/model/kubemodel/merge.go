//nolint:stylecheck,gocyclo // complex merge logic
package kubemodel

import (
	"fmt"
	"maps"
	"math"
	"slices"
)

// Merge combines two KubeModelSets into a new one, aggregating metrics and handling mutable fields appropriately.
// The resulting KubeModelSet will have:
// - Window: from earliest start to latest end
// - Immutable fields: taken from the first occurrence
// - Mutable billing fields: merged according to their semantics
func Merge(kms1, kms2 *KubeModelSet) (*KubeModelSet, error) {
	if kms1 == nil && kms2 == nil {
		return nil, fmt.Errorf("both KubeModelSets are nil")
	}
	if kms1 == nil {
		return kms2, nil
	}
	if kms2 == nil {
		return kms1, nil
	}

	// Verify same cluster
	if kms1.Cluster != nil && kms2.Cluster != nil && kms1.Cluster.UID != kms2.Cluster.UID {
		return nil, fmt.Errorf(
			"cannot merge KubeModelSets from different clusters: %s vs %s",
			kms1.Cluster.UID, kms2.Cluster.UID)
	}

	// Create merged window (earliest start to latest end)
	windowStart := kms1.Window.Start
	if kms2.Window.Start.Before(windowStart) {
		windowStart = kms2.Window.Start
	}
	windowEnd := kms1.Window.End
	if kms2.Window.End.After(windowEnd) {
		windowEnd = kms2.Window.End
	}

	merged := NewKubeModelSet(windowStart, windowEnd)
	// Set the merged window duration
	if windowEnd.After(windowStart) {
		merged.Window.DurationSeconds = uint64(windowEnd.Sub(windowStart).Seconds())
	}

	// Merge Metadata
	if kms1.Metadata != nil && kms2.Metadata != nil {
		// Start: take earliest
		if kms2.Metadata.Start.Before(kms1.Metadata.Start) {
			merged.Metadata.Start = kms2.Metadata.Start
		} else {
			merged.Metadata.Start = kms1.Metadata.Start
		}
		// End: take latest
		if kms2.Metadata.End.After(kms1.Metadata.End) {
			merged.Metadata.End = kms2.Metadata.End
		} else {
			merged.Metadata.End = kms1.Metadata.End
		}
		// ObjectCount: sum
		merged.Metadata.ObjectCount = kms1.Metadata.ObjectCount + kms2.Metadata.ObjectCount
		// Diagnostics: combine
		merged.Metadata.Diagnostics = append(
			append([]*DiagnosticResult{}, kms1.Metadata.Diagnostics...),
			kms2.Metadata.Diagnostics...,
		)
	} else if kms1.Metadata != nil {
		merged.Metadata.Start = kms1.Metadata.Start
		merged.Metadata.End = kms1.Metadata.End
		merged.Metadata.ObjectCount = kms1.Metadata.ObjectCount
		merged.Metadata.Diagnostics = append([]*DiagnosticResult{}, kms1.Metadata.Diagnostics...)
	} else if kms2.Metadata != nil {
		merged.Metadata.Start = kms2.Metadata.Start
		merged.Metadata.End = kms2.Metadata.End
		merged.Metadata.ObjectCount = kms2.Metadata.ObjectCount
		merged.Metadata.Diagnostics = append([]*DiagnosticResult{}, kms2.Metadata.Diagnostics...)
	}

	merged.Cluster = kms1.Cluster
	if merged.Cluster == nil {
		merged.Cluster = kms2.Cluster
	}

	// Merge all resource types
	mergeNamespaces(merged, kms1, kms2)
	mergeResourceQuotas(merged, kms1, kms2)
	mergeNodes(merged, kms1, kms2)
	mergePods(merged, kms1, kms2)
	mergeContainers(merged, kms1, kms2)
	mergeOwners(merged, kms1, kms2)
	mergeServices(merged, kms1, kms2)
	mergeVolumes(merged, kms1, kms2)
	mergePVCs(merged, kms1, kms2)
	mergeDevices(merged, kms1, kms2)
	mergeDeviceUsages(merged, kms1, kms2)

	return merged, nil
}

func mergeNamespaces(merged, kms1, kms2 *KubeModelSet) {
	for uid, ns := range kms1.Namespaces {
		merged.Namespaces[uid] = copyNamespace(ns)
		merged.idx.namespaceNameToID[ns.Name] = uid
		merged.Metadata.ObjectCount++
	}
	for uid, ns := range kms2.Namespaces {
		if _, exists := merged.Namespaces[uid]; !exists {
			merged.Namespaces[uid] = copyNamespace(ns)
			merged.idx.namespaceNameToID[ns.Name] = uid
			merged.Metadata.ObjectCount++
		}
	}
}

func mergeResourceQuotas(merged, kms1, kms2 *KubeModelSet) {
	for uid, rq := range kms1.ResourceQuotas {
		merged.ResourceQuotas[uid] = copyResourceQuota(rq)
		merged.Metadata.ObjectCount++
	}
	for uid, rq := range kms2.ResourceQuotas {
		if _, exists := merged.ResourceQuotas[uid]; !exists {
			merged.ResourceQuotas[uid] = copyResourceQuota(rq)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergeNodes(merged, kms1, kms2 *KubeModelSet) {
	for uid, node := range kms1.Nodes {
		merged.Nodes[uid] = copyNode(node)
		merged.Metadata.ObjectCount++
	}
	for uid, node2 := range kms2.Nodes {
		if node1, exists := merged.Nodes[uid]; exists {
			// Merge mutable metrics
			node1.CpuMillicoreSeconds += node2.CpuMillicoreSeconds
			node1.RAMKiBSeconds += node2.RAMKiBSeconds
			node1.CpuMillicoreUsageMax = max(node1.CpuMillicoreUsageMax, node2.CpuMillicoreUsageMax)
			node1.RAMByteUsageMax = max(node1.RAMByteUsageMax, node2.RAMByteUsageMax)
			node1.PublicIPSeconds += node2.PublicIPSeconds
			node1.DurationSeconds += node2.DurationSeconds

			// Merge lifecycle fields
			// Start: take earliest
			if node2.Start != nil && (node1.Start == nil || node2.Start.Before(*node1.Start)) {
				node1.Start = node2.Start
			}
			// End: take latest
			if node2.End != nil && (node1.End == nil || node2.End.After(*node1.End)) {
				node1.End = node2.End
			}

			// Merge attached volumes
			for volumeUID, volume2 := range node2.AttachedVolumes {
				if volume1, exists := node1.AttachedVolumes[volumeUID]; exists {
					// Merge volume usage
					volume1.UsageKiBSeconds += volume2.UsageKiBSeconds
					volume1.DurationSeconds += volume2.DurationSeconds
					// Take max capacity (should be the same, but use max to be safe)
					if volume2.CapacityBytes > volume1.CapacityBytes {
						volume1.CapacityBytes = volume2.CapacityBytes
					}
				} else {
					// Copy new volume
					node1.AttachedVolumes[volumeUID] = &NodeVolumeUsage{
						VolumeUID:       volume2.VolumeUID,
						CapacityBytes:   volume2.CapacityBytes,
						UsageKiBSeconds: volume2.UsageKiBSeconds,
						VolumeType:      volume2.VolumeType,
						ProviderID:      volume2.ProviderID,
						DurationSeconds: volume2.DurationSeconds,
					}
				}
			}
		} else {
			merged.Nodes[uid] = copyNode(node2)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergePods(merged, kms1, kms2 *KubeModelSet) {
	for uid, pod := range kms1.Pods {
		merged.Pods[uid] = copyPod(pod)
		merged.Metadata.ObjectCount++
	}
	for uid, pod2 := range kms2.Pods {
		if pod1, exists := merged.Pods[uid]; exists {
			// Merge mutable metrics
			pod1.CpuMillicoreUsageMax = max(pod1.CpuMillicoreUsageMax, pod2.CpuMillicoreUsageMax)
			pod1.RAMByteUsageMax = max(pod1.RAMByteUsageMax, pod2.RAMByteUsageMax)
			pod1.NetworkReceiveBytes += pod2.NetworkReceiveBytes
			pod1.NetworkTransferBytes += pod2.NetworkTransferBytes
			pod1.DurationSeconds += pod2.DurationSeconds
			pod1.CpuMillicoreRequestSeconds += pod2.CpuMillicoreRequestSeconds
			pod1.RAMKiBRequestSeconds += pod2.RAMKiBRequestSeconds

			// Merge lifecycle fields
			// Start: take earliest
			if pod2.Start != nil && (pod1.Start == nil || pod2.Start.Before(*pod1.Start)) {
				pod1.Start = pod2.Start
			}
			// End: take latest
			if pod2.End != nil && (pod1.End == nil || pod2.End.After(*pod1.End)) {
				pod1.End = pod2.End
			}

			// Merge network breakdown fields
			pod1.NetworkInternetEgressBytes += pod2.NetworkInternetEgressBytes
			pod1.NetworkCrossRegionBytes += pod2.NetworkCrossRegionBytes
			pod1.NetworkSameRegionBytes += pod2.NetworkSameRegionBytes
			pod1.NetworkIntraAZBytes += pod2.NetworkIntraAZBytes
		} else {
			merged.Pods[uid] = copyPod(pod2)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergeContainers(merged, kms1, kms2 *KubeModelSet) {
	for uid, container := range kms1.Containers {
		merged.Containers[uid] = copyContainer(container)
		merged.Metadata.ObjectCount++
	}
	for uid, container2 := range kms2.Containers {
		if container1, exists := merged.Containers[uid]; exists {
			// Merge mutable metrics
			container1.CpuMillicoreSeconds += container2.CpuMillicoreSeconds
			container1.RAMKiBSeconds += container2.RAMKiBSeconds
			container1.CpuMillicoreUsageMax = max(container1.CpuMillicoreUsageMax, container2.CpuMillicoreUsageMax)
			container1.RAMByteUsageMax = max(container1.RAMByteUsageMax, container2.RAMByteUsageMax)

			// Merge volume storage maps
			for volumeUID, kibSeconds := range container2.VolumeStorageKiBSeconds {
				container1.VolumeStorageKiBSeconds[volumeUID] += kibSeconds
			}
			for volumeUID, usageMax := range container2.VolumeStorageByteUsageMax {
				if currentMax, exists := container1.VolumeStorageByteUsageMax[volumeUID]; exists {
					container1.VolumeStorageByteUsageMax[volumeUID] = max(currentMax, usageMax)
				} else {
					container1.VolumeStorageByteUsageMax[volumeUID] = usageMax
				}
			}

			// Merge request and limit metrics
			container1.CpuMillicoreRequestSeconds += container2.CpuMillicoreRequestSeconds
			container1.RAMKiBRequestSeconds += container2.RAMKiBRequestSeconds
			container1.CpuMillicoreLimitSeconds += container2.CpuMillicoreLimitSeconds
			container1.RAMKiBLimitSeconds += container2.RAMKiBLimitSeconds

			container1.DurationSeconds += container2.DurationSeconds
		} else {
			merged.Containers[uid] = copyContainer(container2)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergeOwners(merged, kms1, kms2 *KubeModelSet) {
	for uid, owner := range kms1.Owners {
		merged.Owners[uid] = copyOwner(owner)
		merged.Metadata.ObjectCount++
	}
	for uid, owner2 := range kms2.Owners {
		if owner1, exists := merged.Owners[uid]; exists {
			// Merge lifecycle fields
			// Start: take earliest
			if owner2.Start != nil && (owner1.Start == nil || owner2.Start.Before(*owner1.Start)) {
				owner1.Start = owner2.Start
			}
			// End: take latest
			if owner2.End != nil && (owner1.End == nil || owner2.End.After(*owner1.End)) {
				owner1.End = owner2.End
			}
		} else {
			merged.Owners[uid] = copyOwner(owner2)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergeServices(merged, kms1, kms2 *KubeModelSet) {
	for uid, svc := range kms1.Services {
		merged.Services[uid] = copyService(svc)
		merged.Metadata.ObjectCount++
	}
	for uid, svc2 := range kms2.Services {
		if svc1, exists := merged.Services[uid]; exists {
			// Merge mutable metrics
			svc1.NetworkTransferBytes += svc2.NetworkTransferBytes
			svc1.NetworkReceiveBytes += svc2.NetworkReceiveBytes
			svc1.DurationSeconds += svc2.DurationSeconds

			// Merge lifecycle fields
			// Start: take earliest
			if svc2.Start.Before(svc1.Start) {
				svc1.Start = svc2.Start
			}
			// End: take latest
			if svc2.End.After(svc1.End) {
				svc1.End = svc2.End
			}

			// Merge network breakdown fields
			svc1.NetworkInternetEgressBytes += svc2.NetworkInternetEgressBytes
			svc1.NetworkCrossRegionBytes += svc2.NetworkCrossRegionBytes
			svc1.NetworkSameRegionBytes += svc2.NetworkSameRegionBytes
			svc1.NetworkIntraAZBytes += svc2.NetworkIntraAZBytes
		} else {
			merged.Services[uid] = copyService(svc2)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergeVolumes(merged, kms1, kms2 *KubeModelSet) {
	for uid, vol := range kms1.Volumes {
		merged.Volumes[uid] = copyVolume(vol)
		merged.Metadata.ObjectCount++
	}
	for uid, vol2 := range kms2.Volumes {
		if vol1, exists := merged.Volumes[uid]; exists {
			// Merge mutable billing fields
			// Start: take earliest
			if vol2.Start.Before(vol1.Start) {
				vol1.Start = vol2.Start
			}
			// End: take latest
			if vol2.End != nil {
				if vol1.End == nil || vol2.End.After(*vol1.End) {
					vol1.End = vol2.End
				}
			}
			// DurationSeconds: sum across windows
			vol1.DurationSeconds += vol2.DurationSeconds
		} else {
			merged.Volumes[uid] = copyVolume(vol2)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergePVCs(merged, kms1, kms2 *KubeModelSet) {
	for uid, pvc := range kms1.PersistentVolumeClaims {
		merged.PersistentVolumeClaims[uid] = copyPVC(pvc)
		merged.Metadata.ObjectCount++
	}
	for uid, pvc2 := range kms2.PersistentVolumeClaims {
		if pvc1, exists := merged.PersistentVolumeClaims[uid]; exists {
			// Merge mutable billing fields - sum all KiB-seconds and durations
			pvc1.StorageKiBSeconds += pvc2.StorageKiBSeconds
			pvc1.ActualUsedKiBSeconds += pvc2.ActualUsedKiBSeconds
			pvc1.DurationSeconds += pvc2.DurationSeconds

			// Start: take earliest
			if pvc2.Start.Before(pvc1.Start) {
				pvc1.Start = pvc2.Start
			}
			// End: take latest
			if pvc2.End != nil {
				if pvc1.End == nil || pvc2.End.After(*pvc1.End) {
					pvc1.End = pvc2.End
				}
			}
			// BoundAt: take earliest
			if pvc2.BoundAt != nil {
				if pvc1.BoundAt == nil || pvc2.BoundAt.Before(*pvc1.BoundAt) {
					pvc1.BoundAt = pvc2.BoundAt
				}
			}
		} else {
			merged.PersistentVolumeClaims[uid] = copyPVC(pvc2)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergeDevices(merged, kms1, kms2 *KubeModelSet) {
	for uid, dev := range kms1.Devices {
		merged.Devices[uid] = copyDevice(dev)
		merged.Metadata.ObjectCount++
	}
	for uid, dev2 := range kms2.Devices {
		if dev1, exists := merged.Devices[uid]; exists {
			// Merge mutable metrics
			dev1.UsageSeconds += dev2.UsageSeconds
			dev1.MemoryKiBSeconds += dev2.MemoryKiBSeconds
			dev1.PowerWattSeconds += dev2.PowerWattSeconds
			dev1.PowerWattMax = math.Max(dev1.PowerWattMax, dev2.PowerWattMax)
			dev1.DurationSeconds += dev2.DurationSeconds

			// Merge lifecycle fields
			// Start: take earliest
			if dev2.Start != nil && (dev1.Start == nil || dev2.Start.Before(*dev1.Start)) {
				dev1.Start = dev2.Start
			}
			// End: take latest
			if dev2.End != nil && (dev1.End == nil || dev2.End.After(*dev1.End)) {
				dev1.End = dev2.End
			}
		} else {
			merged.Devices[uid] = copyDevice(dev2)
			merged.Metadata.ObjectCount++
		}
	}
}

func mergeDeviceUsages(merged, kms1, kms2 *KubeModelSet) {
	for uid, usage := range kms1.DeviceUsages {
		merged.DeviceUsages[uid] = copyDeviceUsage(usage)
		merged.Metadata.ObjectCount++
	}
	for uid, usage2 := range kms2.DeviceUsages {
		if usage1, exists := merged.DeviceUsages[uid]; exists {
			// Merge mutable metrics
			usage1.UsageSeconds += usage2.UsageSeconds
			usage1.MemoryKiBSecondsUsed += usage2.MemoryKiBSecondsUsed
			usage1.UsagePercentageMax = math.Max(usage1.UsagePercentageMax, usage2.UsagePercentageMax)
		} else {
			merged.DeviceUsages[uid] = copyDeviceUsage(usage2)
			merged.Metadata.ObjectCount++
		}
	}
}

// Copy functions to create deep copies of objects

func copyNamespace(ns *Namespace) *Namespace {
	return &Namespace{
		UID:         ns.UID,
		Name:        ns.Name,
		Labels:      maps.Clone(ns.Labels),
		Annotations: maps.Clone(ns.Annotations),
	}
}

func copyResourceQuota(rq *ResourceQuota) *ResourceQuota {
	copied := &ResourceQuota{
		UID:          rq.UID,
		Name:         rq.Name,
		NamespaceUID: rq.NamespaceUID,
	}
	if rq.Spec != nil {
		copied.Spec = &ResourceQuotaSpec{}
		if rq.Spec.Hard != nil {
			copied.Spec.Hard = &ResourceQuotaSpecHard{
				Requests: copyResourceQuantities(rq.Spec.Hard.Requests),
				Limits:   copyResourceQuantities(rq.Spec.Hard.Limits),
			}
		}
	}
	if rq.Status != nil {
		copied.Status = &ResourceQuotaStatus{}
		if rq.Status.Used != nil {
			copied.Status.Used = &ResourceQuotaStatusUsed{
				Requests: copyResourceQuantities(rq.Status.Used.Requests),
				Limits:   copyResourceQuantities(rq.Status.Used.Limits),
			}
		}
	}
	return copied
}

// copyResourceQuantities creates a deep copy of ResourceQuantities, including all maps
func copyResourceQuantities(rq ResourceQuantities) ResourceQuantities {
	return ResourceQuantities{
		CPUMillicores:          rq.CPUMillicores,
		MemoryBytes:            rq.MemoryBytes,
		StorageBytes:           rq.StorageBytes,
		EphemeralStorageBytes:  rq.EphemeralStorageBytes,
		StorageByClass:         maps.Clone(rq.StorageByClass),
		Pods:                   rq.Pods,
		Services:               rq.Services,
		ReplicationControllers: rq.ReplicationControllers,
		ResourceQuotas:         rq.ResourceQuotas,
		Secrets:                rq.Secrets,
		ConfigMaps:             rq.ConfigMaps,
		PersistentVolumeClaims: rq.PersistentVolumeClaims,
		PVCsByClass:            maps.Clone(rq.PVCsByClass),
		ExtendedResources:      maps.Clone(rq.ExtendedResources),
	}
}

func copyNode(node *Node) *Node {
	copied := &Node{
		UID:                  node.UID,
		ClusterUID:           node.ClusterUID,
		Name:                 node.Name,
		ProviderResourceUID:  node.ProviderResourceUID,
		Labels:               maps.Clone(node.Labels),
		Annotations:          maps.Clone(node.Annotations),
		CpuMillicoreSeconds:  node.CpuMillicoreSeconds,
		RAMKiBSeconds:        node.RAMKiBSeconds,
		CpuMillicoreUsageMax: node.CpuMillicoreUsageMax,
		RAMByteUsageMax:      node.RAMByteUsageMax,
		PublicIPSeconds:      node.PublicIPSeconds,
		DurationSeconds:      node.DurationSeconds,
		AttachedVolumes:      make(map[string]*NodeVolumeUsage),
	}

	// Copy lifecycle fields
	if node.Start != nil {
		start := *node.Start
		copied.Start = &start
	}
	if node.End != nil {
		finish := *node.End
		copied.End = &finish
	}

	// Deep copy attached volumes
	for volumeUID, volume := range node.AttachedVolumes {
		copied.AttachedVolumes[volumeUID] = &NodeVolumeUsage{
			VolumeUID:       volume.VolumeUID,
			CapacityBytes:   volume.CapacityBytes,
			UsageKiBSeconds: volume.UsageKiBSeconds,
			VolumeType:      volume.VolumeType,
			ProviderID:      volume.ProviderID,
			DurationSeconds: volume.DurationSeconds,
		}
	}

	return copied
}

func copyPod(pod *Pod) *Pod {
	copied := &Pod{
		UID:                        pod.UID,
		Name:                       pod.Name,
		NamespaceUID:               pod.NamespaceUID,
		OwnerUID:                   pod.OwnerUID,
		NodeUID:                    pod.NodeUID,
		Labels:                     maps.Clone(pod.Labels),
		Annotations:                maps.Clone(pod.Annotations),
		CpuMillicoreUsageMax:       pod.CpuMillicoreUsageMax,
		RAMByteUsageMax:            pod.RAMByteUsageMax,
		NetworkReceiveBytes:        pod.NetworkReceiveBytes,
		NetworkTransferBytes:       pod.NetworkTransferBytes,
		DurationSeconds:            pod.DurationSeconds,
		CpuMillicoreRequestSeconds: pod.CpuMillicoreRequestSeconds,
		RAMKiBRequestSeconds:       pod.RAMKiBRequestSeconds,
		NetworkInternetEgressBytes: pod.NetworkInternetEgressBytes,
		NetworkCrossRegionBytes:    pod.NetworkCrossRegionBytes,
		NetworkSameRegionBytes:     pod.NetworkSameRegionBytes,
		NetworkIntraAZBytes:        pod.NetworkIntraAZBytes,
	}
	if pod.Start != nil {
		start := *pod.Start
		copied.Start = &start
	}
	if pod.End != nil {
		finish := *pod.End
		copied.End = &finish
	}
	return copied
}

func copyContainer(container *Container) *Container {
	return &Container{
		PodUID:                     container.PodUID,
		Name:                       container.Name,
		CpuMillicoreSeconds:        container.CpuMillicoreSeconds,
		RAMKiBSeconds:              container.RAMKiBSeconds,
		CpuMillicoreUsageMax:       container.CpuMillicoreUsageMax,
		RAMByteUsageMax:            container.RAMByteUsageMax,
		VolumeStorageKiBSeconds:    maps.Clone(container.VolumeStorageKiBSeconds),
		VolumeStorageByteUsageMax:  maps.Clone(container.VolumeStorageByteUsageMax),
		DurationSeconds:            container.DurationSeconds,
		CpuMillicoreRequestSeconds: container.CpuMillicoreRequestSeconds,
		RAMKiBRequestSeconds:       container.RAMKiBRequestSeconds,
		CpuMillicoreLimitSeconds:   container.CpuMillicoreLimitSeconds,
		RAMKiBLimitSeconds:         container.RAMKiBLimitSeconds,
	}
}

func copyOwner(owner *Owner) *Owner {
	copied := &Owner{
		UID:          owner.UID,
		Name:         owner.Name,
		NamespaceUID: owner.NamespaceUID,
		Kind:         owner.Kind,
		Labels:       maps.Clone(owner.Labels),
		Annotations:  maps.Clone(owner.Annotations),
	}
	if owner.Start != nil {
		start := *owner.Start
		copied.Start = &start
	}
	if owner.End != nil {
		finish := *owner.End
		copied.End = &finish
	}
	return copied
}

func copyService(svc *Service) *Service {
	copied := &Service{
		UID:                        svc.UID,
		ClusterUID:                 svc.ClusterUID,
		NamespaceUID:               svc.NamespaceUID,
		Name:                       svc.Name,
		Type:                       svc.Type,
		Hostname:                   svc.Hostname,
		Labels:                     maps.Clone(svc.Labels),
		Annotations:                maps.Clone(svc.Annotations),
		NetworkTransferBytes:       svc.NetworkTransferBytes,
		NetworkReceiveBytes:        svc.NetworkReceiveBytes,
		DurationSeconds:            svc.DurationSeconds,
		NetworkInternetEgressBytes: svc.NetworkInternetEgressBytes,
		NetworkCrossRegionBytes:    svc.NetworkCrossRegionBytes,
		NetworkSameRegionBytes:     svc.NetworkSameRegionBytes,
		NetworkIntraAZBytes:        svc.NetworkIntraAZBytes,
		Selector:                   maps.Clone(svc.Selector),
		Ports:                      slices.Clone(svc.Ports),
	}
	copied.Start = svc.Start
	copied.End = svc.End

	return copied
}

func copyVolume(vol *PersistentVolume) *PersistentVolume {
	copied := &PersistentVolume{
		UID:                   vol.UID,
		ClusterUID:            vol.ClusterUID,
		Name:                  vol.Name,
		Namespace:             vol.Namespace,
		Labels:                maps.Clone(vol.Labels),
		Annotations:           maps.Clone(vol.Annotations),
		StorageClass:          vol.StorageClass,
		SizeBytes:             vol.SizeBytes,
		Type:                  vol.Type,
		CSIDriver:             vol.CSIDriver,
		ProviderVolumeID:      vol.ProviderVolumeID,
		AccessModes:           slices.Clone(vol.AccessModes),
		ReclaimPolicy:         vol.ReclaimPolicy,
		Region:                vol.Region,
		Zone:                  vol.Zone,
		VolumeAttributes:      maps.Clone(vol.VolumeAttributes),
		Start:                 vol.Start,
		DurationSeconds:       vol.DurationSeconds,
		NodeAffinity:          vol.NodeAffinity,
		ProvisionedIOPS:       vol.ProvisionedIOPS,
		ProvisionedThroughput: vol.ProvisionedThroughput,
		PerformanceMode:       vol.PerformanceMode,
	}
	if vol.End != nil {
		finish := *vol.End
		copied.End = &finish
	}
	return copied
}

func copyPVC(pvc *PersistentVolumeClaim) *PersistentVolumeClaim {
	copied := &PersistentVolumeClaim{
		UID:                   pvc.UID,
		NamespaceUID:          pvc.NamespaceUID,
		Name:                  pvc.Name,
		Labels:                maps.Clone(pvc.Labels),
		Annotations:           maps.Clone(pvc.Annotations),
		StorageClass:          pvc.StorageClass,
		StorageKiBSeconds:     pvc.StorageKiBSeconds,
		RequestedBytes:        pvc.RequestedBytes,
		Size:                  pvc.Size,
		VolumeName:            pvc.VolumeName,
		AccessModes:           slices.Clone(pvc.AccessModes),
		VolumeAttributes:      maps.Clone(pvc.VolumeAttributes),
		Start:                 pvc.Start,
		DurationSeconds:       pvc.DurationSeconds,
		ActualUsedKiBSeconds:  pvc.ActualUsedKiBSeconds,
		ProvisionedIOPS:       pvc.ProvisionedIOPS,
		ProvisionedThroughput: pvc.ProvisionedThroughput,
		PerformanceMode:       pvc.PerformanceMode,
	}
	if pvc.VolumeUID != nil {
		volumeUID := *pvc.VolumeUID
		copied.VolumeUID = &volumeUID
	}
	if pvc.PodUID != nil {
		podUID := *pvc.PodUID
		copied.PodUID = &podUID
	}
	if pvc.End != nil {
		finish := *pvc.End
		copied.End = &finish
	}
	if pvc.BoundAt != nil {
		boundAt := *pvc.BoundAt
		copied.BoundAt = &boundAt
	}
	return copied
}

func copyDevice(dev *Device) *Device {
	copied := &Device{
		UID:              dev.UID,
		Type:             dev.Type,
		NodeUID:          dev.NodeUID,
		DeviceNumber:     dev.DeviceNumber,
		ModelName:        dev.ModelName,
		IsShared:         dev.IsShared,
		SharePercentage:  dev.SharePercentage,
		UsageSeconds:     dev.UsageSeconds,
		MemoryKiBSeconds: dev.MemoryKiBSeconds,
		PowerWattSeconds: dev.PowerWattSeconds,
		PowerWattMax:     dev.PowerWattMax,
		DurationSeconds:  dev.DurationSeconds,
	}
	if dev.Start != nil {
		start := *dev.Start
		copied.Start = &start
	}
	if dev.End != nil {
		finish := *dev.End
		copied.End = &finish
	}
	return copied
}

func copyDeviceUsage(usage *DeviceUsage) *DeviceUsage {
	return &DeviceUsage{
		ContainerUID:         usage.ContainerUID,
		DeviceUID:            usage.DeviceUID,
		UsageSeconds:         usage.UsageSeconds,
		UsagePercentageMax:   usage.UsagePercentageMax,
		MemoryKiBSecondsUsed: usage.MemoryKiBSecondsUsed,
	}
}
