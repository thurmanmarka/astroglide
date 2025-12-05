package astroglide_test

import (
	"math"
	"testing"
	"time"

	"github.com/thurmanmarka/astroglide"
)

func TestDaylightHours(t *testing.T) {
	phoenix := astroglide.Coordinates{
		Lat: 33.4484,
		Lon: -112.0740,
	}

	locPHX, _ := time.LoadLocation("America/Phoenix")

	tests := []struct {
		name         string
		date         time.Time
		wantMinHours float64 // minimum expected hours
		wantMaxHours float64 // maximum expected hours
	}{
		{
			name:         "Phoenix Summer Solstice",
			date:         time.Date(2025, time.June, 21, 0, 0, 0, 0, locPHX),
			wantMinHours: 14.0,
			wantMaxHours: 14.5,
		},
		{
			name:         "Phoenix Winter Solstice",
			date:         time.Date(2025, time.December, 21, 0, 0, 0, 0, locPHX),
			wantMinHours: 9.8,
			wantMaxHours: 10.2,
		},
		{
			name:         "Phoenix Spring Equinox",
			date:         time.Date(2025, time.March, 20, 0, 0, 0, 0, locPHX),
			wantMinHours: 11.9,
			wantMaxHours: 12.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hours, err := astroglide.DaylightHours(phoenix, tt.date)
			if err != nil {
				t.Fatalf("DaylightHours() error = %v", err)
			}

			if hours < tt.wantMinHours || hours > tt.wantMaxHours {
				t.Errorf("DaylightHours() = %.2f hours, want between %.2f and %.2f",
					hours, tt.wantMinHours, tt.wantMaxHours)
			}

			t.Logf("%s: %.2f hours of daylight", tt.name, hours)
		})
	}
}

func TestDaylightHours_Equator(t *testing.T) {
	// At the equator, daylight should be ~12 hours year-round
	quito := astroglide.Coordinates{
		Lat: -0.1807,
		Lon: -78.4678,
	}

	locQuito, _ := time.LoadLocation("America/Guayaquil")

	dates := []time.Time{
		time.Date(2025, time.March, 20, 0, 0, 0, 0, locQuito),
		time.Date(2025, time.June, 21, 0, 0, 0, 0, locQuito),
		time.Date(2025, time.September, 22, 0, 0, 0, 0, locQuito),
		time.Date(2025, time.December, 21, 0, 0, 0, 0, locQuito),
	}

	for _, date := range dates {
		hours, err := astroglide.DaylightHours(quito, date)
		if err != nil {
			t.Fatalf("DaylightHours() error = %v for %s", err, date.Format("2006-01-02"))
		}

		// At the equator, expect ~12 hours ± 15 minutes
		if math.Abs(hours-12.0) > 0.25 {
			t.Errorf("Quito %s: got %.2f hours, expected ~12 hours",
				date.Format("2006-01-02"), hours)
		}

		t.Logf("Quito %s: %.2f hours", date.Format("2006-01-02"), hours)
	}
}
