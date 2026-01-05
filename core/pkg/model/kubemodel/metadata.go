package kubemodel

import (
	"time"
)

// @bingen:generate:Metadata
type Metadata struct {
	Start       time.Time           `json:"start"`                 // @bingen:field[version=1] - Processing start timestamp
	End         time.Time           `json:"end"`                   // @bingen:field[version=1] - Processing end timestamp
	ObjectCount int                 `json:"objectCount"`           // @bingen:field[version=1]
	Diagnostics []*DiagnosticResult `json:"diagnostics,omitempty"` // @bingen:field[version=1]
}
