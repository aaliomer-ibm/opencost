package kubemodel

import (
	"time"
)

type Container struct {
	PodUID                     string                 `json:"podUid"`
	Name                       string                 `json:"name"`
	DurationSeconds            Measurement            `json:"durationSeconds"`
	CpuMillicoreSeconds        Measurement            `json:"cpuMillicoreSeconds"`
	CpuMillicoreUsageMax       Measurement            `json:"cpuMillicoreUsageMax"`
	CpuMillicoreRequestSeconds Measurement            `json:"cpuMillicoreRequestSeconds"`
	RAMByteSeconds             Measurement            `json:"ramByteSeconds"`
	RAMByteUsageMax            Measurement            `json:"ramByteUsageMax"`
	RAMKiBRequestSeconds       Measurement            `json:"ramKiBRequestSeconds"`
	VolumeStorageByteSeconds   map[string]Measurement `json:"volumeStorageByteSeconds,omitempty"`
	VolumeStorageByteUsageMax  map[string]Measurement `json:"volumeStorageByteUsageMax,omitempty"`
	CpuMillicoreLimitSeconds   Measurement            `json:"cpuMillicoreLimitSeconds,omitempty"`
	RAMKiBLimitSeconds         Measurement            `json:"ramKiBLimitSeconds,omitempty"`
	Start                      time.Time              `json:"start"`
	End                        time.Time              `json:"end"`
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
