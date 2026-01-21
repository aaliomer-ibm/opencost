package kubemodel

import (
	"time"

	"github.com/google/uuid"
)

// @bingen:generate:Node
// Node represents a Kubernetes node with capacity-based resource tracking.
// All resource measures (CPU, RAM) represent node capacity, not requests or limits.
// This aligns with the principle that cost allocation should be based on provisioned capacity.
type Node struct {
	UID                  uuid.UUID                      `json:"uid"`                       // @bingen:field[version=1]
	ProviderResourceUID  uuid.UUID                      `json:"providerResourceUid"`       // @bingen:field[version=1]
	Name                 string                         `json:"name"`                      // @bingen:field[version=1]
	Labels               map[string]string              `json:"labels,omitempty"`          // @bingen:field[version=1]
	Annotations          map[string]string              `json:"annotations,omitempty"`     // @bingen:field[version=1]
	DurationSeconds      Measurement                    `json:"durationSeconds"`           // @bingen:field[version=1]
	CpuMillicoreSeconds  Measurement                    `json:"cpuMillicoreSeconds"`       // @bingen:field[version=1] - Node CPU capacity in millicore-seconds
	RAMByteSeconds       Measurement                    `json:"ramByteSeconds"`            // @bingen:field[version=1] - Node RAM capacity in KiB-seconds
	AttachedVolumes      map[uuid.UUID]*NodeVolumeUsage `json:"attachedVolumes,omitempty"` // @bingen:field[version=1]
	CpuMillicoreUsageMax Measurement                    `json:"cpuMillicoreUsageMax"`      // @bingen:field[version=1] - Peak CPU usage observed
	RAMByteUsageMax      Measurement                    `json:"ramByteUsageMax"`           // @bingen:field[version=1] - Peak RAM usage observed
	Start                time.Time                      `json:"start,omitempty"`           // @bingen:field[version=1] - Node creation/start timestamp
	End                  time.Time                      `json:"end,omitempty"`             // @bingen:field[version=1] - Node deletion/end timestamp (nil if still running)
}

// NodeVolumeUsage tracks storage usage for a disk volume attached to a node.
// Used for cost allocation of cloud storage resources (e.g., AWS EBS volumes).
type NodeVolumeUsage struct {
	VolumeUID        uuid.UUID   `json:"volumeUid"`        // @bingen:field[version=1] - "root" for primary disk, or actual volume UID for additional volumes
	CapacityBytes    Measurement `json:"capacityBytes"`    // @bingen:field[version=1] - Total capacity of the volume in bytes
	UsageByteSeconds Measurement `json:"usageByteSeconds"` // @bingen:field[version=1] - Cumulative usage (KiB × seconds) over measurement window
	VolumeType       string      `json:"volumeType"`       // @bingen:field[version=1] - "root" for primary disk, "persistent" for additional PVs
	ProviderID       string      `json:"providerId"`       // @bingen:field[version=1] - Cloud provider volume ID (e.g., "vol-xxxxx" for AWS EBS)
	DurationSeconds  Measurement `json:"durationSeconds"`  // @bingen:field[version=1] - Duration the volume was attached during measurement window in seconds
}

// CpuMillicoreUsageAverage calculates the average CPU usage in millicores over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (n *Node) CpuMillicoreUsageAverage() Measurement {
	if n.DurationSeconds == 0 {
		return 0
	}
	return n.CpuMillicoreSeconds / n.DurationSeconds
}

// RAMByteUsageAverage calculates the average RAM usage in bytes over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (n *Node) RAMByteUsageAverage() Measurement {
	if n.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(n.RAMByteSeconds) / n.DurationSeconds
}

// TotalVolumeUsageByteSeconds returns the sum of all volume usage KiB-seconds across all attached volumes.
func (n *Node) TotalVolumeUsageByteSeconds() Measurement {
	var total Measurement
	for _, volume := range n.AttachedVolumes {
		total += volume.UsageByteSeconds
	}
	return total
}

// TotalVolumeCapacityBytes returns the sum of all volume capacities across all attached volumes.
func (n *Node) TotalVolumeCapacityBytes() Measurement {
	var total Measurement
	for _, volume := range n.AttachedVolumes {
		total += volume.CapacityBytes
	}
	return total
}

// GetVolumeUsageAverage calculates the average storage usage in bytes for a specific volume over the uptime period.
// Returns 0 if uptime is 0 or volume doesn't exist.
func (n *Node) GetVolumeUsageAverage(volumeUID uuid.UUID) Measurement {
	volume, exists := n.AttachedVolumes[volumeUID]
	if !exists || n.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(volume.UsageByteSeconds) / n.DurationSeconds
}
