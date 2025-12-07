//nolint:stylecheck // generated code and package conventions
package kubemodel

import (
	"time"
)

// @bingen:generate:Node
// Node represents a Kubernetes node with capacity-based resource tracking.
// All resource measures (CPU, RAM) represent node capacity, not requests or limits.
// This aligns with the principle that cost allocation should be based on provisioned capacity.
type Node struct {
	UID                 string            `json:"uid"`                   // @bingen:field[version=1]
	ProviderResourceUID string            `json:"providerResourceUid"`   // @bingen:field[version=1]
	Name                string            `json:"name"`                  // @bingen:field[version=1]
	Labels              map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations         map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	DurationSeconds     uint64            `json:"durationSeconds"`       // @bingen:field[version=1]
	CpuMillicoreSeconds uint64            `json:"cpuMillicoreSeconds"`   // @bingen:field[version=1] - Node CPU capacity in millicore-seconds
	RAMKiBSeconds       uint64            `json:"ramKiBSeconds"`         // @bingen:field[version=1] - Node RAM capacity in KiB-seconds
	// PublicIPSeconds represents the cumulative public IP allocation (count × seconds) for this node.
	// Calculated as: number of ExternalIP addresses from Kubernetes node Status.Addresses × window duration in seconds.
	// Used for cost attribution of public IP addresses associated with the node.
	PublicIPSeconds uint64 `json:"publicIpSeconds"` // @bingen:field[version=1]
	// AttachedVolumes tracks disk storage volumes attached to this node for cost allocation purposes.
	// Currently tracks the root volume (e.g., EBS volume for AWS EC2 nodes) which contains all node storage.
	// Key is volume identifier: "root" for the primary disk volume, or volume UID for additional volumes.
	// Container rootfs/logs usage is tracked separately at the container level and maps to this root volume
	// for proportional cost allocation. Map design allows future extensibility for multi-volume nodes.
	AttachedVolumes      map[string]*NodeVolumeUsage `json:"attachedVolumes,omitempty"` // @bingen:field[version=1]
	CpuMillicoreUsageMax uint64                      `json:"cpuMillicoreUsageMax"`      // @bingen:field[version=1] - Peak CPU usage observed
	RAMByteUsageMax      uint64                      `json:"ramByteUsageMax"`           // @bingen:field[version=1] - Peak RAM usage observed
	// Version 2 fields - Lifecycle tracking
	Start *time.Time `json:"start,omitempty"` // @bingen:field[version=1] - Node creation/start timestamp
	End   *time.Time `json:"end,omitempty"`   // @bingen:field[version=1] - Node deletion/end timestamp (nil if still running)
}

// NodeVolumeUsage tracks storage usage for a disk volume attached to a node.
// Used for cost allocation of cloud storage resources (e.g., AWS EBS volumes).
type NodeVolumeUsage struct {
	VolumeUID       string `json:"volumeUid"`       // @bingen:field[version=1] - "root" for primary disk, or actual volume UID for additional volumes
	CapacityBytes   uint64 `json:"capacityBytes"`   // @bingen:field[version=1] - Total capacity of the volume in bytes
	UsageKiBSeconds uint64 `json:"usageKiBSeconds"` // @bingen:field[version=1] - Cumulative usage (KiB × seconds) over measurement window
	VolumeType      string `json:"volumeType"`      // @bingen:field[version=1] - "root" for primary disk, "persistent" for additional PVs
	ProviderID      string `json:"providerId"`      // @bingen:field[version=1] - Cloud provider volume ID (e.g., "vol-xxxxx" for AWS EBS)
	DurationSeconds uint64 `json:"durationSeconds"` // @bingen:field[version=1] - Duration the volume was attached during measurement window in seconds
}

// CpuMillicoreUsageAverage calculates the average CPU usage in millicores over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (n *Node) CpuMillicoreUsageAverage() uint64 {
	if n.DurationSeconds == 0 {
		return 0
	}
	return n.CpuMillicoreSeconds / n.DurationSeconds
}

// RAMByteUsageAverage calculates the average RAM usage in bytes over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (n *Node) RAMByteUsageAverage() uint64 {
	if n.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(n.RAMKiBSeconds) / n.DurationSeconds
}

// TotalVolumeUsageKiBSeconds returns the sum of all volume usage KiB-seconds across all attached volumes.
func (n *Node) TotalVolumeUsageKiBSeconds() uint64 {
	var total uint64
	for _, volume := range n.AttachedVolumes {
		total += volume.UsageKiBSeconds
	}
	return total
}

// TotalVolumeCapacityBytes returns the sum of all volume capacities across all attached volumes.
func (n *Node) TotalVolumeCapacityBytes() uint64 {
	var total uint64
	for _, volume := range n.AttachedVolumes {
		total += volume.CapacityBytes
	}
	return total
}

// GetVolumeUsageAverage calculates the average storage usage in bytes for a specific volume over the uptime period.
// Returns 0 if uptime is 0 or volume doesn't exist.
func (n *Node) GetVolumeUsageAverage(volumeUID string) uint64 {
	volume, exists := n.AttachedVolumes[volumeUID]
	if !exists || n.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(volume.UsageKiBSeconds) / n.DurationSeconds
}
