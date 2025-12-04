package astroglide

import (
	"testing"
	"time"
)

// diffMinutes returns the absolute difference between two times in minutes.
func diffMinutes(a, b time.Time) float64 {
	d := a.Sub(b)
	if d < 0 {
		d = -d
	}
	return d.Minutes()
}

// For now, the lunar model is approximate. We allow a fairly relaxed tolerance.
// Once you refine the RA/Dec model, you can tighten this (e.g. 15–20 minutes).
const moonToleranceMinutes = 45.0

// TestMoonRiseSet_Phoenix_2025_11_30 compares our Moon rise/set against
// online ephemeris values for Phoenix, AZ on 2025-11-30.
//
// Reference (ephemeris tables, Phoenix AZ November 2025):
//
//	Moonrise ≈ 14:10 (2:10 PM)
//	Moonset  ≈ 02:13 (2:13 AM)
//
// Times are local in America/Phoenix.
func TestMoonRiseSet_Phoenix_2025_11_30(t *testing.T) {
	locPHX, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		t.Fatalf("failed to load America/Phoenix: %v", err)
	}

	date := time.Date(2025, time.November, 30, 0, 0, 0, 0, locPHX)

	coords := Coordinates{
		Lat: 33.4484,   // Phoenix latitude
		Lon: -112.0740, // Phoenix longitude
	}

	rs, err := RiseSetFor(Moon, coords, date)
	if err != nil {
		t.Fatalf("RiseSetFor(Moon) returned error: %v", err)
	}

	// Expected reference times in local time zone
	expectedRise := time.Date(2025, time.November, 30, 14, 10, 0, 0, locPHX)
	expectedSet := time.Date(2025, time.November, 30, 2, 13, 0, 0, locPHX)

	if got := diffMinutes(rs.Rise.In(locPHX), expectedRise); got > moonToleranceMinutes {
		t.Errorf("Phoenix moonrise off by %.1f minutes (got %v, want ~%v)",
			got, rs.Rise.In(locPHX), expectedRise)
	}
	if got := diffMinutes(rs.Set.In(locPHX), expectedSet); got > moonToleranceMinutes {
		t.Errorf("Phoenix moonset off by %.1f minutes (got %v, want ~%v)",
			got, rs.Set.In(locPHX), expectedSet)
	}

	// Note: we intentionally do NOT assert rise < set here, because for many dates
	// the Moon sets in the early morning and rises in the afternoon.
}

// TestMoonRiseSet_NewYork_2025_11_30 compares our Moon rise/set against
// online ephemeris values for New York City, NY on 2025-11-30.
//
// Reference (ephemeris tables, NYC November 2025):
//
//	Moonrise ≈ 13:30 (1:30 PM)
//	Moonset  ≈ 01:36 (1:36 AM)
//
// Times are local in America/New_York.
func TestMoonRiseSet_NewYork_2025_11_30(t *testing.T) {
	locNY, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load America/New_York: %v", err)
	}

	date := time.Date(2025, time.November, 30, 0, 0, 0, 0, locNY)

	coords := Coordinates{
		Lat: 40.7128,  // NYC latitude
		Lon: -74.0060, // NYC longitude
	}

	rs, err := RiseSetFor(Moon, coords, date)
	if err != nil {
		t.Fatalf("RiseSetFor(Moon) returned error: %v", err)
	}

	// Expected reference times in local time zone
	expectedRise := time.Date(2025, time.November, 30, 13, 30, 0, 0, locNY)
	expectedSet := time.Date(2025, time.November, 30, 1, 36, 0, 0, locNY)

	if got := diffMinutes(rs.Rise.In(locNY), expectedRise); got > moonToleranceMinutes {
		t.Errorf("NYC moonrise off by %.1f minutes (got %v, want ~%v)",
			got, rs.Rise.In(locNY), expectedRise)
	}
	if got := diffMinutes(rs.Set.In(locNY), expectedSet); got > moonToleranceMinutes {
		t.Errorf("NYC moonset off by %.1f minutes (got %v, want ~%v)",
			got, rs.Set.In(locNY), expectedSet)
	}

	// Same as above: we don't assert ordering.
}
