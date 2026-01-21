package kubemodel

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// @bingen:generate:DeviceUsage
type DeviceUsage struct {
	ContainerUID          uuid.UUID   `json:"containerUid"`              // @bingen:field[version=1]
	DeviceUID             uuid.UUID   `json:"deviceUid"`                 // @bingen:field[version=1]
	UsageSeconds          Measurement `json:"usageSeconds"`              // @bingen:field[version=1]
	UsagePercentageMax    float64     `json:"usagePercentageMax"`        // @bingen:field[version=1]
	MemoryByteSecondsUsed Measurement `json:"memoryByteSecondsUsed"`     // @bingen:field[version=1]
	DeviceType            string      `json:"deviceType,omitempty"`      // @bingen:field[version=1]
	DurationSeconds       Measurement `json:"durationSeconds,omitempty"` // @bingen:field[version=1]
	Start                 time.Time   `json:"start"`                     // @bingen:field[version=1]
	End                   time.Time   `json:"end"`                       // @bingen:field[version=1]
}

func (u *DeviceUsage) Validate() error {
	if u.ContainerUID == uuid.Nil {
		return errors.New("ContainerUID is required")
	}
	if u.DeviceUID == uuid.Nil {
		return errors.New("DeviceUID is required")
	}
	if u.UsagePercentageMax < 0 || u.UsagePercentageMax > 100 {
		return fmt.Errorf("UsagePercentageMax must be 0-100, got %.2f", u.UsagePercentageMax)
	}
	return nil
}

func (u *DeviceUsage) Clone() *DeviceUsage {
	if u == nil {
		return nil
	}

	cloned := &DeviceUsage{
		ContainerUID:          u.ContainerUID,
		DeviceUID:             u.DeviceUID,
		UsageSeconds:          u.UsageSeconds,
		UsagePercentageMax:    u.UsagePercentageMax,
		MemoryByteSecondsUsed: u.MemoryByteSecondsUsed,
		DeviceType:            u.DeviceType,
		DurationSeconds:       u.DurationSeconds,
		Start:                 u.Start,
		End:                   u.End,
	}

	return cloned
}

func (u *DeviceUsage) UsageAverage() Measurement {
	if u.DurationSeconds == 0 {
		return 0
	}
	return (u.UsageSeconds / u.DurationSeconds) * 100
}

func (u *DeviceUsage) MemoryByteUsageAverage() Measurement {
	if u.DurationSeconds == 0 {
		return 0
	}
	return KiBToBytes(u.MemoryByteSecondsUsed) / u.DurationSeconds
}
