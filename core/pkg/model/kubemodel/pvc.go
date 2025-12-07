//nolint:stylecheck // generated code and package conventions
package kubemodel

import "time"

// @bingen:generate:PersistentVolumeClaim
type PersistentVolumeClaim struct {
	// Version 1 fields
	UID               string            `json:"uid"`                   // @bingen:field[version=1]
	NamespaceUID      string            `json:"namespaceUid"`          // @bingen:field[version=1]
	VolumeUID         *string           `json:"volumeUid,omitempty"`   // @bingen:field[version=1]
	PodUID            *string           `json:"podUid,omitempty"`      // @bingen:field[version=1]
	Name              string            `json:"name"`                  // @bingen:field[version=1]
	Labels            map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations       map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	StorageClass      string            `json:"storageClass"`          // @bingen:field[version=1]
	StorageKiBSeconds uint64            `json:"storageKiBSeconds"`     // @bingen:field[version=1]
	RequestedBytes    uint64            `json:"requestedBytes"`        // @bingen:field[version=1]
	Size              uint64            `json:"size"`                  // @bingen:field[version=1] - Size in bytes
	VolumeName        string            `json:"volumeName"`            // @bingen:field[version=1]

	// Version 2 fields - FinOps enhancements
	// ReadWriteOnce, ReadWriteMany, ReadOnlyMany
	AccessModes []string `json:"accessModes,omitempty"` // @bingen:field[version=1]
	// Real usage KiB-seconds from kubelet stats
	ActualUsedKiBSeconds uint64 `json:"actualUsedKiBSeconds,omitempty"` // @bingen:field[version=1]
	// Generic attributes (IOPS, throughput, etc.) from CSI or provider
	VolumeAttributes map[string]string `json:"volumeAttributes,omitempty"` // @bingen:field[version=1]
	// PVC lifecycle timestamps
	Start time.Time  `json:"start"`         // @bingen:field[version=1] - PVC creation timestamp
	End   *time.Time `json:"end,omitempty"` // @bingen:field[version=1] - PVC deletion timestamp (nil if still active)
	// Timestamp when PVC was bound to PV
	BoundAt *time.Time `json:"boundAt,omitempty"` // @bingen:field[version=1]
	// Duration PVC existed within the measurement window (in seconds)
	// Calculated as: min(Finish or now, PV.Finish or now) - Start
	// This ensures PVC duration never exceeds the bound PV's lifetime
	DurationSeconds uint64 `json:"durationSeconds,omitempty"` // @bingen:field[version=1]
}
