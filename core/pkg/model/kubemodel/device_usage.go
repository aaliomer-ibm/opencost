//nolint:stylecheck,lll // generated code and package conventions
package kubemodel

import (
	"errors"
	"fmt"
)

// @bingen:generate:DeviceUsage
// DeviceUsage represents GPU resources consumed by a container (allocated resource)
// This tracks actual GPU usage by containers for cost analysis
// GPU has two key dimensions: compute and memory
type DeviceUsage struct {
	ContainerUID         string  `json:"containerUid"`         // @bingen:field[version=1] Container consuming GPU resources
	DeviceUID            string  `json:"deviceUid"`            // @bingen:field[version=1] Reference to the GPU device being used
	UsageSeconds         float64 `json:"usageSeconds"`         // @bingen:field[version=1] GPU compute usage in device-seconds consumed
	UsagePercentageMax   float64 `json:"usagePercentageMax"`   // @bingen:field[version=1] GPU compute usage max percentage (0-100)
	MemoryKiBSecondsUsed uint64  `json:"memoryKiBSecondsUsed"` // @bingen:field[version=1] GPU memory usage in KiB-seconds
	// Version 2 fields - Enhanced device tracking
	DeviceType      string `json:"deviceType,omitempty"`      // @bingen:field[version=1] Device type (e.g., "gpu", "tpu", "fpga")
	DurationSeconds uint64 `json:"durationSeconds,omitempty"` // @bingen:field[version=1] Duration of device usage measurement window in seconds
}

// Validate validates the GPUUsage fields
func (u *DeviceUsage) Validate() error {
	if u.ContainerUID == "" {
		return errors.New("ContainerUID is required")
	}
	if u.DeviceUID == "" {
		return errors.New("GpuDeviceUID is required")
	}
	if u.UsagePercentageMax < 0 || u.UsagePercentageMax > 100 {
		return fmt.Errorf("GpuUsagePercentageMax must be 0-100, got %.2f", u.UsagePercentageMax)
	}
	if u.UsageSeconds < 0 {
		return fmt.Errorf("GpuSeconds cannot be negative, got %.2f", u.UsageSeconds)
	}
	return nil
}

// Clone creates a deep copy of the DeviceUsage
func (u *DeviceUsage) Clone() *DeviceUsage {
	if u == nil {
		return nil
	}

	cloned := &DeviceUsage{
		ContainerUID:         u.ContainerUID,
		DeviceUID:            u.DeviceUID,
		UsageSeconds:         u.UsageSeconds,
		UsagePercentageMax:   u.UsagePercentageMax,
		MemoryKiBSecondsUsed: u.MemoryKiBSecondsUsed,
		DeviceType:           u.DeviceType,
		DurationSeconds:      u.DurationSeconds,
	}

	return cloned
}

// UsageAverage calculates the average GPU usage percentage over the duration.
// Returns 0 if duration is 0 to avoid division by zero.
func (u *DeviceUsage) UsageAverage() float64 {
	if u.DurationSeconds == 0 {
		return 0
	}
	// UsageSeconds represents device-seconds, so divide by duration to get average percentage
	return (u.UsageSeconds / float64(u.DurationSeconds)) * 100
}

// MemoryByteUsageAverage calculates the average GPU memory usage in bytes over the duration.
// Returns 0 if duration is 0 to avoid division by zero.
func (u *DeviceUsage) MemoryByteUsageAverage() uint64 {
	if u.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(u.MemoryKiBSecondsUsed) / u.DurationSeconds
}
