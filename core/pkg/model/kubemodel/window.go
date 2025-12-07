//nolint:stylecheck // generated code and package conventions
package kubemodel

import "time"

// @bingen:generate:Window
type Window struct {
	Start time.Time `json:"start"` // @bingen:field[version=1]
	End   time.Time `json:"end"`   // @bingen:field[version=1]
	// Sometimes the duration is smaller than the time between start and end (missing data for example)
	DurationSeconds uint64 `json:"durationSeconds"` // @bingen:field[version=1]
}
