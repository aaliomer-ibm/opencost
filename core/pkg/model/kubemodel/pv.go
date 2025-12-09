//nolint:stylecheck // generated code and package conventions
package kubemodel

import (
	"time"

	"github.com/google/uuid"
)

// @bingen:generate:PersistentVolume
type PersistentVolume struct {
	// Version 1 fields
	UID          uuid.UUID         `json:"uid"`                   // @bingen:field[version=1]
	ClusterUID   uuid.UUID         `json:"clusterUid"`            // @bingen:field[version=1]
	Name         string            `json:"name"`                  // @bingen:field[version=1]
	Namespace    string            `json:"namespace"`             // @bingen:field[version=1]
	Labels       map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations  map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	StorageClass string            `json:"storageClass"`          // @bingen:field[version=1]
	SizeBytes    uint64            `json:"size"`                  // @bingen:field[version=1]
	// awsElasticBlockStore, azureDisk, gcePersistentDisk, csi, nfs, local, etc.
	Type string `json:"type,omitempty"` // @bingen:field[version=1]
	// ebs.csi.aws.com, disk.csi.azure.com, etc.
	CSIDriver string `json:"csiDriver,omitempty"` // @bingen:field[version=1]
	// Cloud provider's volume identifier
	ProviderVolumeID string `json:"providerVolumeId,omitempty"` // @bingen:field[version=1]
	// ReadWriteOnce, ReadWriteMany, ReadOnlyMany
	AccessModes []string `json:"accessModes,omitempty"` // @bingen:field[version=1]
	// Retain, Delete, Recycle
	ReclaimPolicy string `json:"reclaimPolicy,omitempty"` // @bingen:field[version=1]
	// Cloud region for cross-region cost tracking
	Region string `json:"region,omitempty"` // @bingen:field[version=1]
	// Availability zone for cross-AZ cost tracking
	Zone string `json:"zone,omitempty"` // @bingen:field[version=1]
	// Volume lifecycle timestamps
	Start time.Time `json:"start"`         // @bingen:field[version=1] - Volume creation timestamp
	End   time.Time `json:"end,omitempty"` // @bingen:field[version=1] - Volume deletion timestamp (nil if still active)
	// Duration volume existed within measurement window
	DurationSeconds uint64 `json:"durationSeconds"` // @bingen:field[version=1]
	// JSON-encoded node affinity for local volumes
	NodeAffinity string `json:"nodeAffinity,omitempty"` // @bingen:field[version=1]
	// Storage performance characteristics
	ProvisionedIOPS       uint64 `json:"provisionedIops,omitempty"`       // @bingen:field[version=1] - Provisioned IOPS (AWS io1/io2, Azure Premium)
	ProvisionedThroughput uint64 `json:"provisionedThroughput,omitempty"` // @bingen:field[version=1] - Provisioned throughput in MB/s
	PerformanceMode       string `json:"performanceMode,omitempty"`       // @bingen:field[version=1] - "generalPurpose", "maxIO", "provisioned"
}
