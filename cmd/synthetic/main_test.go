package main

import (
	"fmt"
	"testing"
	"time"
)


func TestBug3_ResolutionThresholds(t *testing.T) {
	configs := createResolutionConfigs(30 * time.Minute)
	intervals := make(map[string]bool)
	for _, c := range configs {
		intervals[c.Interval] = true
	}

	for _, want := range []string{"10m", "1h", "1d"} {
		if !intervals[want] {
			t.Errorf("missing %q resolution for 30min duration", want)
		}
	}
}

func TestBug3_ResolutionThresholds_AllDurations(t *testing.T) {
	durations := []time.Duration{
		1 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
		2 * time.Hour,
		6 * time.Hour,
		24 * time.Hour,
		48 * time.Hour,
		72 * time.Hour,
	}

	for _, d := range durations {
		configs := createResolutionConfigs(d)
		intervals := make(map[string]bool)
		for _, c := range configs {
			intervals[c.Interval] = true
		}

		for _, want := range []string{"10m", "1h", "1d"} {
			if !intervals[want] {
				t.Errorf("duration=%v: missing %q resolution", d, want)
			}
		}
	}
}


func TestPreservation_ResolutionRetentionValues(t *testing.T) {
	durations := []time.Duration{
		10 * time.Minute,
		1 * time.Hour,
		3 * time.Hour,
		24 * time.Hour,
		72 * time.Hour,
	}

	for _, d := range durations {
		configs := createResolutionConfigs(d)
		retMap := make(map[string]int)
		for _, c := range configs {
			retMap[c.Interval] = c.Retention
		}

		tag := fmt.Sprintf("duration=%v", d)

		want10m := int(d.Minutes()/10) + 6
		if retMap["10m"] != want10m {
			t.Errorf("%s: 10m retention = %d, want %d", tag, retMap["10m"], want10m)
		}

		want1h := int(d.Hours()) + 24
		if retMap["1h"] != want1h {
			t.Errorf("%s: 1h retention = %d, want %d", tag, retMap["1h"], want1h)
		}

		want1d := int(d.Hours()/24) + 7
		if retMap["1d"] != want1d {
			t.Errorf("%s: 1d retention = %d, want %d", tag, retMap["1d"], want1d)
		}
	}
}
