package kubemodel

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// @bingen:generate:Device
type Device struct {
	UID               uuid.UUID   `json:"uid"`               // @bingen:field[version=1] Device UUID (hardware identifier)
	Type              string      `json:"type,omitempty"`    // @bingen:field[version=1] Device type (e.g., "device", "tpu")
	NodeUID           uuid.UUID   `json:"nodeUid"`           // @bingen:field[version=1] Node hosting this device
	DeviceNumber      int32       `json:"deviceNumber"`      // @bingen:field[version=1]
	ModelName         string      `json:"modelName"`         // @bingen:field[version=1]
	IsShared          bool        `json:"isShared"`          // @bingen:field[version=1] Device sharing information
	SharePercentage   float64     `json:"sharePercentage"`   // @bingen:field[version=1]
	UsageSeconds      float64     `json:"usageSeconds"`      // @bingen:field[version=1] Device seconds available
	MemoryByteSeconds Measurement `json:"memoryByteSeconds"` // @bingen:field[version=1] Device memory capacity in KiB-seconds
	PowerWattSeconds  float64     `json:"powerWattSeconds"`  // @bingen:field[version=1] Device power consumption in watt-seconds (Joules)
	PowerWattMax      float64     `json:"powerWattMax"`      // @bingen:field[version=1] Device max power consumption in watts
	// Version 2 fields - Lifecycle tracking
	Start           time.Time   `json:"start,omitempty"` // @bingen:field[version=1] - Device availability start
	End             time.Time   `json:"end,omitempty"`   // @bingen:field[version=1] - Device availability end
	DurationSeconds Measurement `json:"durationSeconds"` // @bingen:field[version=1] - Duration device was available
}

// Validate validates the Device fields
func (d *Device) Validate() error {
	if d.UID == uuid.Nil {
		return errors.New("UID is required")
	}
	if d.NodeUID == uuid.Nil {
		return errors.New("NodeUID is required")
	}
	if d.SharePercentage < 0 || d.SharePercentage > 100 {
		return fmt.Errorf("SharePercentage must be 0-100, got %.2f", d.SharePercentage)
	}
	if d.PowerWattSeconds < 0 {
		return fmt.Errorf("PowerWattSeconds cannot be negative, got %.2f", d.PowerWattSeconds)
	}
	if d.PowerWattMax < 0 {
		return fmt.Errorf("PowerWattMax cannot be negative, got %.2f", d.PowerWattMax)
	}
	return nil
}

// Clone creates a deep copy of the Device
func (d *Device) Clone() *Device {
	if d == nil {
		return nil
	}

	cloned := &Device{
		UID:               d.UID,
		Type:              d.Type,
		NodeUID:           d.NodeUID,
		DeviceNumber:      d.DeviceNumber,
		ModelName:         d.ModelName,
		IsShared:          d.IsShared,
		SharePercentage:   d.SharePercentage,
		UsageSeconds:      d.UsageSeconds,
		MemoryByteSeconds: d.MemoryByteSeconds,
		PowerWattSeconds:  d.PowerWattSeconds,
		PowerWattMax:      d.PowerWattMax,
		DurationSeconds:   d.DurationSeconds,
	}

	cloned.Start = d.Start
	cloned.End = d.End

	return cloned
}
