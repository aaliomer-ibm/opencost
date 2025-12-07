//nolint:stylecheck // generated code and package conventions
package kubemodel

import "time"

// @bingen:generate:Pod
type Pod struct {
	UID                        string            `json:"uid"`                        // @bingen:field[version=1]
	NamespaceUID               string            `json:"namespaceUid"`               // @bingen:field[version=1]
	OwnerUID                   string            `json:"ownerUid"`                   // @bingen:field[version=1] - Reference to Owner (Deployment, StatefulSet, etc.)
	NodeUID                    string            `json:"nodeUid"`                    // @bingen:field[version=1]
	Name                       string            `json:"name"`                       // @bingen:field[version=1]
	Labels                     map[string]string `json:"labels,omitempty"`           // @bingen:field[version=1]
	Annotations                map[string]string `json:"annotations,omitempty"`      // @bingen:field[version=1]
	DurationSeconds            uint64            `json:"durationSeconds"`            // @bingen:field[version=1]
	CpuMillicoreUsageMax       uint64            `json:"cpuMillicoreUsageMax"`       // @bingen:field[version=1]
	CpuMillicoreUsageSeconds   uint64            `json:"cpuMillicoreUsageSeconds"`   // @bingen:field[version=1]
	CpuMillicoreRequestSeconds uint64            `json:"cpuMillicoreRequestSeconds"` // @bingen:field[version=1]
	RAMByteUsageMax            uint64            `json:"ramByteUsageMax"`            // @bingen:field[version=1]
	RAMKiBUsageSeconds         uint64            `json:"ramKiBUsageSeconds"`         // @bingen:field[version=1]
	RAMKiBRequestSeconds       uint64            `json:"ramKiBRequestSeconds"`       // @bingen:field[version=1]
	NetworkTransferBytes       uint64            `json:"networkTransferBytes"`       // @bingen:field[version=1]
	NetworkReceiveBytes        uint64            `json:"networkReceiveBytes"`        // @bingen:field[version=1]
	// Version 2 fields - Lifecycle tracking
	Start *time.Time `json:"start,omitempty"` // @bingen:field[version=1] - Pod creation/start timestamp
	End   *time.Time `json:"end,omitempty"`   // @bingen:field[version=1] - Pod deletion/end timestamp (nil if still running)
	// Version 2 fields - Network breakdown by destination type
	NetworkInternetEgressBytes uint64 `json:"networkInternetEgressBytes,omitempty"` // @bingen:field[version=1]
	NetworkCrossRegionBytes    uint64 `json:"networkCrossRegionBytes,omitempty"`    // @bingen:field[version=1]
	NetworkSameRegionBytes     uint64 `json:"networkSameRegionBytes,omitempty"`     // @bingen:field[version=1]
	NetworkIntraAZBytes        uint64 `json:"networkIntraAzBytes,omitempty"`        // @bingen:field[version=1]
}

// CpuMillicoreUsageAverage calculates the average CPU usage in millicores over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (p *Pod) CpuMillicoreUsageAverage() uint64 {
	if p.DurationSeconds == 0 {
		return 0
	}
	return p.CpuMillicoreUsageSeconds / p.DurationSeconds
}

// RAMByteUsageAverage calculates the average RAM usage in bytes over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (p *Pod) RAMByteUsageAverage() uint64 {
	if p.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(p.RAMKiBUsageSeconds) / p.DurationSeconds
}

// CpuMillicoreRequestAverage calculates the average CPU request in millicores over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (p *Pod) CpuMillicoreRequestAverage() uint64 {
	if p.DurationSeconds == 0 {
		return 0
	}
	return p.CpuMillicoreRequestSeconds / p.DurationSeconds
}

// RAMByteRequestAverage calculates the average RAM request in bytes over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (p *Pod) RAMByteRequestAverage() uint64 {
	if p.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(p.RAMKiBRequestSeconds) / p.DurationSeconds
}
