package astroglide

import (
	"testing"
	"time"
)

// helper: absolute difference in minutes
//func diffMinutes(a, b time.Time) float64 {
//	d := a.Sub(b)
//	if d < 0 {
//		d = -d
//	}
//	return d.Minutes()
//}

func TestDebugTwilight_Phoenix_2025_11_28(t *testing.T) {
	loc, err := time.LoadLocation("America/Phoenix")
	if err != nil {
		t.Fatalf("failed to load Phoenix tz: %v", err)
	}

	// Same coordinates you've been using for Phoenix
	coords := Coordinates{
		Lat: 33.4484,
		Lon: -112.0740,
	}

	// Local calendar date
	date := time.Date(2025, time.November, 28, 0, 0, 0, 0, loc)

	// Reference values taken from an online twilight calculator for
	// Phoenix, AZ on 2025-11-28 (local time, America/Phoenix):
	//
	//   Civil dawn:        06:45
	//   Sunrise:           07:11
	//   Sunset:            17:21
	//   Civil dusk:        17:47
	//   Nautical dawn:     06:14
	//   Nautical dusk:     18:18
	//   Astronomical dawn: 05:44
	//   Astronomical dusk: 18:48
	//
	// We only use the twilight times here.
	type twilightCase struct {
		name       string
		kind       TwilightKind
		expectDawn string // HH:MM local
		expectDusk string // HH:MM local
	}

	cases := []twilightCase{
		{
			name:       "Civil",
			kind:       TwilightCivil,
			expectDawn: "06:45",
			expectDusk: "17:47",
		},
		{
			name:       "Nautical",
			kind:       TwilightNautical,
			expectDawn: "06:14",
			expectDusk: "18:18",
		},
		{
			name:       "Astronomical",
			kind:       TwilightAstronomical,
			expectDawn: "05:44",
			expectDusk: "18:48",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			refDawn, err := time.ParseInLocation("15:04", tc.expectDawn, loc)
			if err != nil {
				t.Fatalf("parse ref dawn %q: %v", tc.expectDawn, err)
			}
			refDusk, err := time.ParseInLocation("15:04", tc.expectDusk, loc)
			if err != nil {
				t.Fatalf("parse ref dusk %q: %v", tc.expectDusk, err)
			}
			// Attach the same calendar date
			refDawn = time.Date(date.Year(), date.Month(), date.Day(),
				refDawn.Hour(), refDawn.Minute(), 0, 0, loc)
			refDusk = time.Date(date.Year(), date.Month(), date.Day(),
				refDusk.Hour(), refDusk.Minute(), 0, 0, loc)

			rs, err := TwilightFor(coords, date, tc.kind)
			if err != nil {
				t.Fatalf("TwilightFor(%s) error: %v", tc.name, err)
			}

			dawn := rs.Rise.In(loc)
			dusk := rs.Set.In(loc)

			dawnErr := diffMinutes(dawn, refDawn)
			duskErr := diffMinutes(dusk, refDusk)

			t.Logf("[%s twilight / Phoenix 2025-11-28]", tc.name)
			t.Logf("  Dawn: expected %s, got %s (err=%.2f min)",
				refDawn.Format(time.RFC3339),
				dawn.Format(time.RFC3339),
				dawnErr)
			t.Logf("  Dusk: expected %s, got %s (err=%.2f min)",
				refDusk.Format(time.RFC3339),
				dusk.Format(time.RFC3339),
				duskErr)

			// Optional loose sanity checks; adjust or drop if you want pure "debug"
			const maxAllowedErr = 5.0 // minutes
			if dawnErr > maxAllowedErr || duskErr > maxAllowedErr {
				t.Fatalf("%s twilight error too large (dawn=%.2f, dusk=%.2f minutes)",
					tc.name, dawnErr, duskErr)
			}
		})
	}
}
