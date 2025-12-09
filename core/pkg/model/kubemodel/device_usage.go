//nolint:stylecheck,lll
package kubemodel

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// @bingen:generate:DeviceUsage
type DeviceUsage struct {
	ContainerUID         uuid.UUID `json:"containerUid"`              // @bingen:field[version=1]
	DeviceUID            uuid.UUID `json:"deviceUid"`                 // @bingen:field[version=1]
	UsageSeconds         float64   `json:"usageSeconds"`              // @bingen:field[version=1]
	UsagePercentageMax   float64   `json:"usagePercentageMax"`        // @bingen:field[version=1]
	MemoryKiBSecondsUsed uint64    `json:"memoryKiBSecondsUsed"`      // @bingen:field[version=1]
	DeviceType           string    `json:"deviceType,omitempty"`      // @bingen:field[version=1]
	DurationSeconds      uint64    `json:"durationSeconds,omitempty"` // @bingen:field[version=1]
	Start                time.Time `json:"start,omitempty"`           // @bingen:field[version=1]
	End                  time.Time `json:"end,omitempty"`             // @bingen:field[version=1]
}

func (u *DeviceUsage) Validate() error {
	if u.ContainerUID == uuid.Nil {
		return errors.New("ContainerUID is required")
	}
	if u.DeviceUID == uuid.Nil {
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
		Start:                u.Start,
		End:                  u.End,
	}

	return cloned
}

func (u *DeviceUsage) UsageAverage() float64 {
	if u.DurationSeconds == 0 {
		return 0
	}
	return (u.UsageSeconds / float64(u.DurationSeconds)) * 100
}

func (u *DeviceUsage) MemoryByteUsageAverage() uint64 {
	if u.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(u.MemoryKiBSecondsUsed) / u.DurationSeconds
}
