//nolint:stylecheck
package kubemodel

import "time"

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
	VolumeStorageKiBSeconds    map[string]uint64 `json:"volumeStorageKiBSeconds,omitempty"`   // @bingen:field[version=1]
	VolumeStorageByteUsageMax  map[string]uint64 `json:"volumeStorageByteUsageMax,omitempty"` // @bingen:field[version=1]
	CpuMillicoreLimitSeconds   uint64            `json:"cpuMillicoreLimitSeconds,omitempty"`  // @bingen:field[version=1]
	RAMKiBLimitSeconds         uint64            `json:"ramKiBLimitSeconds,omitempty"`        // @bingen:field[version=1]
	Start                      time.Time         `json:"start,omitempty"`                     // @bingen:field[version=1]
	End                        time.Time         `json:"end,omitempty"`                       // @bingen:field[version=1]
}

func (c *Container) CpuMillicoreUsageAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreSeconds / c.DurationSeconds
}

func (c *Container) RAMByteUsageAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMKiBSeconds) / c.DurationSeconds
}

func (c *Container) TotalStorageKiBSeconds() uint64 {
	var total uint64
	for _, kibSeconds := range c.VolumeStorageKiBSeconds {
		total += kibSeconds
	}
	return total
}

func (c *Container) TotalStorageByteUsageMax() uint64 {
	var max uint64
	for _, usage := range c.VolumeStorageByteUsageMax {
		if usage > max {
			max = usage
		}
	}
	return max
}

func (c *Container) StorageByteUsageAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	totalKiBSeconds := c.TotalStorageKiBSeconds()
	return KiBToBytes(totalKiBSeconds) / c.DurationSeconds
}

func (c *Container) CpuMillicoreRequestAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreRequestSeconds / c.DurationSeconds
}

func (c *Container) RAMByteRequestAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMKiBRequestSeconds) / c.DurationSeconds
}

func (c *Container) CpuMillicoreLimitAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreLimitSeconds / c.DurationSeconds
}

func (c *Container) RAMByteLimitAverage() uint64 {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMKiBLimitSeconds) / c.DurationSeconds
}
