package kubemodel

import "time"

// @bingen:generate:Window
type Window struct {
	Start time.Time `json:"start"` // @bingen:field[version=1]
	End   time.Time `json:"end"`   // @bingen:field[version=1]
	// Sometimes the duration is smaller than the time between start and end (missing data for example)
	// If the agent is down for 20 minutes in the middle of an hour span, then start and end will be
	// an hour span but the duration will be only 40 minutes to account for the missing 20 minute metric window
	DurationSeconds uint64 `json:"durationSeconds"` // @bingen:field[version=1]
}
