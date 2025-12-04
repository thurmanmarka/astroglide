package astroglide

import (
	"testing"
	"time"
)

// local helper (we don't care if there's also one in other *_test.go files).
func diffMinutesDebug(a, b time.Time) float64 {
	d := a.Sub(b)
	if d < 0 {
		d = -d
	}
	return d.Minutes()
}

// TestDebugEphemeris logs rise/set errors vs. hard-coded ephemeris values
// for a handful of locations/dates and for both Sun and Moon.
//
// It is intentionally *non-failing* and meant to be run manually as:
//
//	go test -run TestDebugEphemeris -v
//
// Use the logged errors to tune your models and shrink tolerances in
// the "real" tests.
func TestDebugEphemeris(t *testing.T) {
	// Load timezones we care about.
	locPHX, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		t.Fatalf("failed to load America/Phoenix: %v", err)
	}
	locNY, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load America/New_York: %v", err)
	}

	type ephemCase struct {
		name         string
		body         Body
		coords       Coordinates
		date         time.Time // local date (uses location)
		expectedRise time.Time // in same location as date
		expectedSet  time.Time // in same location as date
	}

	cases := []ephemCase{
		// --- Phoenix, AZ, 2025-11-30 ---

		// Sun reference: sunrise ≈ 07:13, sunset ≈ 17:21 (local, America/Phoenix)
		{
			name: "Sun Phoenix 2025-11-30",
			body: Sun,
			coords: Coordinates{
				Lat: 33.4484,
				Lon: -112.0740,
			},
			date:         time.Date(2025, time.November, 30, 0, 0, 0, 0, locPHX),
			expectedRise: time.Date(2025, time.November, 30, 7, 13, 0, 0, locPHX),
			expectedSet:  time.Date(2025, time.November, 30, 17, 21, 0, 0, locPHX),
		},

		// Moon reference: moonrise ≈ 14:10, moonset ≈ 02:13 (local, America/Phoenix)
		{
			name: "Moon Phoenix 2025-11-30",
			body: Moon,
			coords: Coordinates{
				Lat: 33.4484,
				Lon: -112.0740,
			},
			date:         time.Date(2025, time.November, 30, 0, 0, 0, 0, locPHX),
			expectedRise: time.Date(2025, time.November, 30, 14, 10, 0, 0, locPHX),
			// moonset in the early morning of the same calendar date
			expectedSet: time.Date(2025, time.November, 30, 2, 13, 0, 0, locPHX),
		},

		// --- New York City, NY, 2025-11-30 ---

		// Sun reference: sunrise ≈ 06:59, sunset ≈ 16:31 (local, America/New_York)
		{
			name: "Sun NewYork 2025-11-30",
			body: Sun,
			coords: Coordinates{
				Lat: 40.7128,
				Lon: -74.0060,
			},
			date:         time.Date(2025, time.November, 30, 0, 0, 0, 0, locNY),
			expectedRise: time.Date(2025, time.November, 30, 6, 59, 0, 0, locNY),
			expectedSet:  time.Date(2025, time.November, 30, 16, 31, 0, 0, locNY),
		},

		// Moon reference: moonrise ≈ 13:30, moonset ≈ 01:36 (local, America/New_York)
		{
			name: "Moon NewYork 2025-11-30",
			body: Moon,
			coords: Coordinates{
				Lat: 40.7128,
				Lon: -74.0060,
			},
			date:         time.Date(2025, time.November, 30, 0, 0, 0, 0, locNY),
			expectedRise: time.Date(2025, time.November, 30, 13, 30, 0, 0, locNY),
			expectedSet:  time.Date(2025, time.November, 30, 1, 36, 0, 0, locNY),
		},
	}

	for _, tc := range cases {
		tc := tc // capture

		t.Run(tc.name, func(t *testing.T) {
			rs, err := RiseSetFor(tc.body, tc.coords, tc.date)
			if err != nil {
				t.Logf("[%s] error from RiseSetFor: %v", tc.name, err)
				return
			}

			// Compare in the local timezone of the test date.
			loc := tc.date.Location()
			gotRise := rs.Rise.In(loc)
			gotSet := rs.Set.In(loc)

			riseErr := diffMinutesDebug(gotRise, tc.expectedRise)
			setErr := diffMinutesDebug(gotSet, tc.expectedSet)

			bodyName := "Sun"
			if tc.body == Moon {
				bodyName = "Moon"
			}

			t.Logf("[%s] %s:", tc.name, bodyName)
			t.Logf("  Expected rise: %v", tc.expectedRise)
			t.Logf("  Got      rise: %v", gotRise)
			t.Logf("  Rise error: %.2f minutes", riseErr)

			t.Logf("  Expected set : %v", tc.expectedSet)
			t.Logf("  Got      set : %v", gotSet)
			t.Logf("  Set error : %.2f minutes", setErr)

			// This is intentionally a debug test, so we don't fail on errors.
		})
	}
}
