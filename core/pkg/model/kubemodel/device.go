//nolint:stylecheck,lll // generated code and package conventions
package kubemodel

import (
	"errors"
	"fmt"
	"time"
)

// @bingen:generate:Device
// Device represents a physical GPU device with DCGM integration (provisioned resource)
// This tracks available GPU capacity on a node and device-level metrics like power consumption
//
// Virtual GPU Cost Allocation Strategy:
// Virtual GPUs (vGPUs) allow multiple containers to share a single physical GPU. Cost allocation for vGPUs
// is handled through the IsShared and SharePercentage fields:
//
// 1. Physical GPU (IsShared=false, SharePercentage=100):
//   - Full device cost allocated to single container using it
//   - UsageSeconds represents full device availability
//
// 2. Shared/Virtual GPU (IsShared=true, SharePercentage<100):
//   - Device cost is proportionally allocated based on SharePercentage
//   - Example: 4 vGPUs from 1 physical GPU → each has SharePercentage=25
//   - Container costs = (Physical GPU cost) × (SharePercentage/100) × (actual usage ratio)
//   - UsageSeconds and MemoryKiBSeconds represent the vGPU slice capacity
//
// 3. Cost Calculation Flow:
//   - Get physical GPU hourly cost from cloud provider pricing
//   - For shared GPUs: vGPU cost = Physical cost × (SharePercentage/100)
//   - Allocate to containers based on DeviceUsage.UsageSeconds / Device.UsageSeconds ratio
//   - Memory-based allocation uses DeviceUsage.MemoryKiBSecondsUsed / Device.MemoryKiBSeconds
//
// 4. Multi-Instance GPU (MIG) Support:
//   - NVIDIA MIG creates hardware-partitioned GPU instances
//   - Each MIG instance appears as separate Device with unique UID
//   - IsShared=true indicates it's a MIG slice
//   - SharePercentage reflects the MIG partition size (e.g., 1g.5gb = ~14% of A100)
//
// TODO: Power metrics are available in Prometheus from dcgm-exporter (DCGM_FI_DEV_POWER_USAGE, DCGM_FI_DEV_TOTAL_ENERGY_CONSUMPTION)
// but not yet integrated into OpenCost's data pipeline. To fully populate power fields:
// 1. Add DCGM_FI_DEV_POWER_USAGE constant to modules/collector-source/pkg/metric/metrics.go
// 2. Update DCGM scraper metric list in modules/collector-source/pkg/scrape/dcgm.go to include power metrics
// 3. Add power metric queries to modules/prometheus-source/pkg/prom/metricsquerier.go (e.g., QueryGPUPowerUsageAvg, QueryGPUPowerUsageMax)
// 4. Implement power metric hydration in pkg/costmodel/kubemodel.go to populate PowerWattSeconds/PowerWattAverage/PowerWattMax fields
type Device struct {
	UID              string  `json:"uid"`              // @bingen:field[version=1] GPU UUID (hardware identifier)
	Type             string  `json:"type,omitempty"`   // @bingen:field[version=1] Device type (e.g., "gpu", "tpu")
	NodeUID          string  `json:"nodeUid"`          // @bingen:field[version=1] Node hosting this GPU device
	DeviceNumber     int32   `json:"deviceNumber"`     // @bingen:field[version=1]
	ModelName        string  `json:"modelName"`        // @bingen:field[version=1]
	IsShared         bool    `json:"isShared"`         // @bingen:field[version=1] GPU sharing information
	SharePercentage  float64 `json:"sharePercentage"`  // @bingen:field[version=1]
	UsageSeconds     float64 `json:"usageSeconds"`     // @bingen:field[version=1] GPU seconds available
	MemoryKiBSeconds uint64  `json:"memoryKiBSeconds"` // @bingen:field[version=1] GPU memory capacity in KiB-seconds
	PowerWattSeconds float64 `json:"powerWattSeconds"` // @bingen:field[version=1] GPU device power consumption in watt-seconds (Joules)
	PowerWattMax     float64 `json:"powerWattMax"`     // @bingen:field[version=1] GPU device max power consumption in watts
	// Version 2 fields - Lifecycle tracking
	Start           *time.Time `json:"start,omitempty"` // @bingen:field[version=1] - Device availability start
	End             *time.Time `json:"end,omitempty"`   // @bingen:field[version=1] - Device availability end
	DurationSeconds uint64     `json:"durationSeconds"` // @bingen:field[version=1] - Duration device was available
}

// Validate validates the GPUDevice fields
func (d *Device) Validate() error {
	if d.UID == "" {
		return errors.New("UID is required")
	}
	if d.NodeUID == "" {
		return errors.New("NodeUID is required")
	}
	if d.SharePercentage < 0 || d.SharePercentage > 100 {
		return fmt.Errorf("SharePercentage must be 0-100, got %.2f", d.SharePercentage)
	}
	if d.UsageSeconds < 0 {
		return fmt.Errorf("GpuSeconds cannot be negative, got %.2f", d.UsageSeconds)
	}
	if d.PowerWattSeconds < 0 {
		return fmt.Errorf("PowerWattSeconds cannot be negative, got %.2f", d.PowerWattSeconds)
	}
	if d.PowerWattMax < 0 {
		return fmt.Errorf("PowerWattMax cannot be negative, got %.2f", d.PowerWattMax)
	}
	return nil
}

// Clone creates a deep copy of the GPUDevice
func (d *Device) Clone() *Device {
	if d == nil {
		return nil
	}

	cloned := &Device{
		UID:              d.UID,
		Type:             d.Type,
		NodeUID:          d.NodeUID,
		DeviceNumber:     d.DeviceNumber,
		ModelName:        d.ModelName,
		IsShared:         d.IsShared,
		SharePercentage:  d.SharePercentage,
		UsageSeconds:     d.UsageSeconds,
		MemoryKiBSeconds: d.MemoryKiBSeconds,
		PowerWattSeconds: d.PowerWattSeconds,
		PowerWattMax:     d.PowerWattMax,
		DurationSeconds:  d.DurationSeconds,
	}

	if d.Start != nil {
		start := *d.Start
		cloned.Start = &start
	}
	if d.End != nil {
		finish := *d.End
		cloned.End = &finish
	}

	return cloned
}
