package kubemodel

import "time"

// @bingen:generate:PersistentVolumeClaim
type PersistentVolumeClaim struct {
	// Version 1 fields
	UID                string            `json:"uid"`                   // @bingen:field[version=1]
	NamespaceUID       string            `json:"namespaceUid"`          // @bingen:field[version=1]
	VolumeUID          *string           `json:"volumeUid,omitempty"`   // @bingen:field[version=1]
	PodUID             *string           `json:"podUid,omitempty"`      // @bingen:field[version=1]
	Name               string            `json:"name"`                  // @bingen:field[version=1]
	Labels             map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations        map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	StorageClass       string            `json:"storageClass"`          // @bingen:field[version=1]
	StorageByteSeconds Measurement       `json:"storageByteSeconds"`    // @bingen:field[version=1]
	RequestedBytes     Measurement       `json:"requestedBytes"`        // @bingen:field[version=1]
	Size               Measurement       `json:"size"`                  // @bingen:field[version=1] - Size in bytes
	VolumeName         string            `json:"volumeName"`            // @bingen:field[version=1]
	// ReadWriteOnce, ReadWriteMany, ReadOnlyMany
	AccessModes           []string    `json:"accessModes,omitempty"`           // @bingen:field[version=1]
	ActualUsedByteSeconds Measurement `json:"actualUsedByteSeconds,omitempty"` // @bingen:field[version=1]
	Start                 time.Time   `json:"start"`                           // @bingen:field[version=1] - PVC creation timestamp
	End                   time.Time   `json:"end,omitempty"`                   // @bingen:field[version=1] - PVC deletion timestamp (nil if still active)
	BoundAt               time.Time   `json:"boundAt,omitempty"`               // @bingen:field[version=1]
	DurationSeconds       Measurement `json:"durationSeconds,omitempty"`       // @bingen:field[version=1]
}
