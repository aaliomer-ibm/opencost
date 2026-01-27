package kubemodel

import (
	"time"
)

type Container struct {
	PodUID                     string                 `json:"podUid"`                              // @bingen:field[version=1]
	Name                       string                 `json:"name"`                                // @bingen:field[version=1]
	DurationSeconds            Measurement            `json:"durationSeconds"`                     // @bingen:field[version=1]
	CpuMillicoreSeconds        Measurement            `json:"cpuMillicoreSeconds"`                 // @bingen:field[version=1]
	CpuMillicoreUsageMax       Measurement            `json:"cpuMillicoreUsageMax"`                // @bingen:field[version=1]
	CpuMillicoreRequestSeconds Measurement            `json:"cpuMillicoreRequestSeconds"`          // @bingen:field[version=1]
	RAMByteSeconds             Measurement            `json:"ramByteSeconds"`                      // @bingen:field[version=1]
	RAMByteUsageMax            Measurement            `json:"ramByteUsageMax"`                     // @bingen:field[version=1]
	RAMKiBRequestSeconds       Measurement            `json:"ramKiBRequestSeconds"`                // @bingen:field[version=1]
	VolumeStorageByteSeconds   map[string]Measurement `json:"volumeStorageByteSeconds,omitempty"`  // @bingen:field[version=1]
	VolumeStorageByteUsageMax  map[string]Measurement `json:"volumeStorageByteUsageMax,omitempty"` // @bingen:field[version=1]
	CpuMillicoreLimitSeconds   Measurement            `json:"cpuMillicoreLimitSeconds,omitempty"`  // @bingen:field[version=1]
	RAMKiBLimitSeconds         Measurement            `json:"ramKiBLimitSeconds,omitempty"`        // @bingen:field[version=1]
	Start                      time.Time              `json:"start"`                               // @bingen:field[version=1]
	End                        time.Time              `json:"end"`                                 // @bingen:field[version=1]
}

func (c *Container) CpuMillicoreUsageAverage() Measurement {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreSeconds / c.DurationSeconds
}

func (c *Container) RAMByteUsageAverage() Measurement {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMByteSeconds / c.DurationSeconds)
}

func (c *Container) TotalStorageByteSeconds() Measurement {
	var total Measurement
	for _, ByteSeconds := range c.VolumeStorageByteSeconds {
		total += ByteSeconds
	}
	return total
}

func (c *Container) TotalStorageByteUsageMax() Measurement {
	var max Measurement
	for _, usage := range c.VolumeStorageByteUsageMax {
		if usage > max {
			max = usage
		}
	}
	return max
}

func (c *Container) StorageByteUsageAverage() Measurement {
	if c.DurationSeconds == 0 {
		return 0
	}
	totalByteSeconds := c.TotalStorageByteSeconds()
	return KiBToBytes(totalByteSeconds) / c.DurationSeconds
}

func (c *Container) CpuMillicoreRequestAverage() Measurement {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreRequestSeconds / c.DurationSeconds
}

func (c *Container) RAMByteRequestAverage() Measurement {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMKiBRequestSeconds / c.DurationSeconds)
}

func (c *Container) CpuMillicoreLimitAverage() Measurement {
	if c.DurationSeconds == 0 {
		return 0
	}
	return c.CpuMillicoreLimitSeconds / c.DurationSeconds
}

func (c *Container) RAMByteLimitAverage() Measurement {
	if c.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(c.RAMKiBLimitSeconds / c.DurationSeconds)
}
