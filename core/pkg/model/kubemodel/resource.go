package kubemodel

// @bingen:generate:ResourceQuantities
// ResourceQuantities represents Kubernetes ResourceQuota quantities with type-safe fields
// for standard resources and flexible maps for dynamic resources (storage classes, extended resources).
//
// Design principles:
// - Standard resources (CPU, Memory, Storage) use explicit fields with units in names for type safety
// - Object counts use explicit fields for compile-time validation
// - Dynamic resources (storage classes, extended resources) use maps for flexibility
// - All quantities use Measurement (never negative)
// - Zero values with omitempty mean quota not set
//
// Units:
// - CPU: millicores (1000m = 1 core)
// - Memory/Storage: bytes
// - Counts: number of objects
// - Extended resources: vendor-specific units (e.g., Device count)
type ResourceQuantities struct {
	// Standard compute resources with explicit units
	CPUMillicores Measurement `json:"cpu,omitempty"`    // @bingen:field[version=1] - CPU quota in millicores (e.g., 1000 = 1 core)
	MemoryBytes   Measurement `json:"memory,omitempty"` // @bingen:field[version=1] - Memory quota in bytes

	// Standard storage resources with explicit units
	StorageBytes          Measurement `json:"storage,omitempty"`          // @bingen:field[version=1] - Total storage quota in bytes
	EphemeralStorageBytes Measurement `json:"ephemeralStorage,omitempty"` // @bingen:field[version=1] - Ephemeral storage quota in bytes

	// Storage per storage class (map because storage classes are cluster-specific)
	// Key: storage class name (e.g., "gp2", "standard", "fast-ssd")
	// Value: storage quota in bytes for that storage class
	StorageByClass map[string]Measurement `json:"storageByClass,omitempty"` // @bingen:field[version=1]

	// Standard Kubernetes object counts
	Pods                   Measurement `json:"pods,omitempty"`                   // @bingen:field[version=1] - Maximum number of pods
	Services               Measurement `json:"services,omitempty"`               // @bingen:field[version=1] - Maximum number of services
	ReplicationControllers Measurement `json:"replicationControllers,omitempty"` // @bingen:field[version=1] - Maximum number of replication controllers
	ResourceQuotas         Measurement `json:"resourceQuotas,omitempty"`         // @bingen:field[version=1] - Maximum number of resource quotas
	Secrets                Measurement `json:"secrets,omitempty"`                // @bingen:field[version=1] - Maximum number of secrets
	ConfigMaps             Measurement `json:"configMaps,omitempty"`             // @bingen:field[version=1] - Maximum number of config maps
	PersistentVolumeClaims Measurement `json:"persistentVolumeClaims,omitempty"` // @bingen:field[version=1] - Maximum number of PVCs

	// PVC count per storage class (map because storage classes are cluster-specific)
	// Key: storage class name (e.g., "gp2", "standard", "fast-ssd")
	// Value: maximum number of PVCs for that storage class
	PVCsByClass map[string]Measurement `json:"pvcsByClass,omitempty"` // @bingen:field[version=1]

	// Extended resources (Devices, FPGAs, custom hardware, etc.)
	// Map because extended resources are cluster-specific and dynamically defined
	// Key: resource name with domain prefix (e.g., "nvidia.com/device", "amd.com/device", "intel.com/fpga")
	// Value: quantity of that extended resource
	// Examples:
	//   - "nvidia.com/device": 4 (4 NVIDIA Devices)
	//   - "amd.com/device": 2 (2 AMD Devices)
	//   - "intel.com/fpga": 1 (1 Intel FPGA)
	//   - "example.com/dongle": 10 (10 custom dongles)
	ExtendedResources map[string]Measurement `json:"extendedResources,omitempty"` // @bingen:field[version=1]
}
