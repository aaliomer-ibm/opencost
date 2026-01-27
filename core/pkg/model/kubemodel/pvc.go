package kubemodel

import "time"

// @bingen:generate:PersistentVolumeClaim
type PersistentVolumeClaim struct {
	// Version 1 fields
	UID                string            `json:"uid"`
	NamespaceUID       string            `json:"namespaceUid"`
	VolumeUID          *string           `json:"volumeUid,omitempty"`
	PodUID             *string           `json:"podUid,omitempty"`
	Name               string            `json:"name"`
	Labels             map[string]string `json:"labels,omitempty"`
	Annotations        map[string]string `json:"annotations,omitempty"`
	StorageClass       string            `json:"storageClass"`
	StorageByteSeconds Measurement       `json:"storageByteSeconds"`
	RequestedBytes     Measurement       `json:"requestedBytes"`
	Size               Measurement       `json:"size"` // Size in bytes
	VolumeName         string            `json:"volumeName"`
	// ReadWriteOnce, ReadWriteMany, ReadOnlyMany
	AccessModes           []string    `json:"accessModes,omitempty"`
	ActualUsedByteSeconds Measurement `json:"actualUsedByteSeconds,omitempty"`
	Start                 time.Time   `json:"start"`         // PVC creation timestamp
	End                   time.Time   `json:"end,omitempty"` // PVC deletion timestamp (nil if still active)
	BoundAt               time.Time   `json:"boundAt,omitempty"`
	DurationSeconds       Measurement `json:"durationSeconds,omitempty"`
}
