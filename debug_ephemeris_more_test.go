package astroglide

import (
	"testing"
	"time"
)

// TestDebugEphemerisMoreLocations logs sunrise/sunset and moonrise/moonset
// for a variety of locations and dates. This does NOT compare against
// reference ephemeris; it's just a convenient way to see what astroglide
// predicts, so you can manually compare against an online source.
//
// Run with:
//
//	go test -run TestDebugEphemerisMoreLocations -v
func TestDebugEphemerisMoreLocations(t *testing.T) {
	type locInfo struct {
		name   string
		coords Coordinates
		tzName string
	}

	locations := []locInfo{
		{
			name: "Quito (Equator)",
			coords: Coordinates{
				Lat: -0.1807,
				Lon: -78.4678,
			},
			tzName: "America/Guayaquil",
		},
		{
			name: "Oslo (High North)",
			coords: Coordinates{
				Lat: 59.9139,
				Lon: 10.7522,
			},
			tzName: "Europe/Oslo",
		},
		{
			name: "Sydney (Southern Hemisphere)",
			coords: Coordinates{
				Lat: -33.8688,
				Lon: 151.2093,
			},
			tzName: "Australia/Sydney",
		},
		{
			name: "Reykjavik (Near Arctic)",
			coords: Coordinates{
				Lat: 64.1466,
				Lon: -21.9426,
			},
			tzName: "Atlantic/Reykjavik",
		},
	}

	// Dates we care about: solstices/equinoxes-ish.
	dateSpecs := []struct {
		label string
		year  int
		month time.Month
		day   int
	}{
		{"March Equinox-ish", 2025, time.March, 20},
		{"June Solstice-ish", 2025, time.June, 21},
		{"September Equinox-ish", 2025, time.September, 22},
		{"December Solstice-ish", 2025, time.December, 21},
	}

	for _, locInfo := range locations {
		locInfo := locInfo
		t.Run(locInfo.name, func(t *testing.T) {
			tz, err := time.LoadLocation(locInfo.tzName)
			if err != nil {
				t.Fatalf("failed to load location %q: %v", locInfo.tzName, err)
			}

			for _, ds := range dateSpecs {
				date := time.Date(ds.year, ds.month, ds.day, 0, 0, 0, 0, tz)

				// Sun
				sunRS, err := RiseSetFor(Sun, locInfo.coords, date)
				if err != nil {
					t.Logf("[%s / %s] Sun: error: %v", locInfo.name, ds.label, err)
				} else {
					t.Logf("[%s / %s] Sun:", locInfo.name, ds.label)
					t.Logf("  Sunrise: %v", sunRS.Rise)
					t.Logf("  Sunset : %v", sunRS.Set)
				}

				// Moon
				moonRS, err := RiseSetFor(Moon, locInfo.coords, date)
				if err != nil {
					t.Logf("[%s / %s] Moon: error: %v", locInfo.name, ds.label, err)
				} else {
					t.Logf("[%s / %s] Moon:", locInfo.name, ds.label)
					t.Logf("  Moonrise: %v", moonRS.Rise)
					t.Logf("  Moonset : %v", moonRS.Set)
				}
			}
		})
	}
}
