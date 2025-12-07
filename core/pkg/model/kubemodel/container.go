//nolint:stylecheck // generated code and package conventions
package kubemodel

// @bingen:generate:Container
type Container struct {
	PodUID                     string            `json:"podUid"`                              // @bingen:field[version=1]
	Name                       string            `json:"name"`                                // @bingen:field[version=1]
	DurationSeconds            uint64            `json:"durationSeconds"`                     // @bingen:field[version=1]
	CpuMillicoreSeconds        uint64            `json:"cpuMillicoreSeconds"`                 // @bingen:field[version=1]
	CpuMillicoreUsageMax       uint64            `json:"cpuMillicoreUsageMax"`                // @bingen:field[version=1]
	CpuMillicoreRequestSeconds uint64            `json:"cpuMillicoreRequestSeconds"`          // @bingen:field[version=1]
	RAMKiBSeconds              uint64            `json:"ramKiBSeconds"`                       // @bingen:field[version=1]
	RAMByteUsageMax            uint64            `json:"ramByteUsageMax"`                     // @bingen:field[version=1]
	RAMKiBRequestSeconds       uint64            `json:"ramKiBRequestSeconds"`                // @bingen:field[version=1]
	VolumeStorageKiBSeconds    map[string]uint64 `json:"volumeStorageKiBSeconds,omitempty"`   // @bingen:field[version=1] // volumeUID -> KiB-seconds
	VolumeStorageByteUsageMax  map[string]uint64 `json:"volumeStorageByteUsageMax,omitempty"` // @bingen:field[version=1] // volumeUID -> max bytes
	// Version 2 fields - Resource limits tracking
	// CPU/RAM limits in resource-seconds. Zero value means no limit was set (unlimited).
	CpuMillicoreLimitSeconds uint64 `json:"cpuMillicoreLimitSeconds,omitempty"` // @bingen:field[version=1] - CPU limit in millicore-seconds
	RAMKiBLimitSeconds       uint64 `json:"ramKiBLimitSeconds,omitempty"`       // @bingen:field[version=1] - RAM limit in KiB-seconds
}

// CpuMillicoreUsageAverage calculates the average CPU usage in millicores over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (c *Container) CpuMillicoreUsageAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreSeconds / c.DurationSeconds
}

// RAMByteUsageAverage calculates the average RAM usage in bytes over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (c *Container) RAMByteUsageAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMKiBSeconds) / c.DurationSeconds
}

// TotalStorageKiBSeconds calculates the total storage KiB-seconds across all volumes.
func (c *Container) TotalStorageKiBSeconds() uint64 {
	var total uint64
	for _, kibSeconds := range c.VolumeStorageKiBSeconds {
		total += kibSeconds
	}
	return total
}

// TotalStorageByteUsageMax returns the maximum storage usage across all volumes.
func (c *Container) TotalStorageByteUsageMax() uint64 {
	var max uint64
	for _, usage := range c.VolumeStorageByteUsageMax {
		if usage > max {
			max = usage
		}
	}
	return max
}

// StorageByteUsageAverage calculates the average storage usage in bytes over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (c *Container) StorageByteUsageAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	totalKiBSeconds := c.TotalStorageKiBSeconds()
	return KiBToBytes(totalKiBSeconds) / c.DurationSeconds
}

// CpuMillicoreRequestAverage calculates the average CPU request in millicores over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (c *Container) CpuMillicoreRequestAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreRequestSeconds / c.DurationSeconds
}

// RAMByteRequestAverage calculates the average RAM request in bytes over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (c *Container) RAMByteRequestAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMKiBRequestSeconds) / c.DurationSeconds
}

// CpuMillicoreLimitAverage calculates the average CPU limit in millicores over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (c *Container) CpuMillicoreLimitAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreLimitSeconds / c.DurationSeconds
}

// RAMByteLimitAverage calculates the average RAM limit in bytes over the uptime period.
// Returns 0 if uptime is 0 to avoid division by zero.
func (c *Container) RAMByteLimitAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMKiBLimitSeconds) / c.DurationSeconds
}
