// Package astroglide provides utilities for computing astronomical events
// like sunrise and sunset for a given location and date.
//
// The public API is designed to remain stable while the internal
// implementations evolve from simple approximate algorithms (Level 1)
// to high-precision ephemeris-grade models (Level 3).
//
// Currently implemented:
//   - Sun rise/set via SlideIntoSunset and RiseSetFor(Sun, ...)
//
// Future plans include:
//   - High-precision solar calculations
//   - Lunar rise/set and additional bodies
//   - Twilight, altitude solvers, and more.
package astroglide

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/thurmanmarka/astroglide/internal/moon"
	"github.com/thurmanmarka/astroglide/internal/sun"
	"github.com/thurmanmarka/astroglide/internal/timeutil"
)

// Body represents a celestial body.
type Body int

// TwilightKind identifies the type of twilight based on the Sun's altitude
// below the horizon.
type TwilightKind int

const (
	// TwilightCivil corresponds to the Sun's center at -6 degrees altitude.
	TwilightCivil TwilightKind = iota

	// TwilightNautical corresponds to the Sun's center at -12 degrees altitude.
	TwilightNautical

	// TwilightAstronomical corresponds to the Sun's center at -18 degrees altitude.
	TwilightAstronomical
)

const (
	Sun Body = iota
	Moon
)

// Coordinates represent an observer's location.
type Coordinates struct {
	Lat       float64 // degrees, north positive
	Lon       float64 // degrees, east positive (west negative, e.g. -105 for 105°W)
	Elevation float64 // meters above sea level (reserved for future use)
}

// RiseSet holds rise and set times of a body on a given date.
type RiseSet struct {
	Rise time.Time
	Set  time.Time
}

// MoonPhase describes the illuminated fraction and qualitative phase
// of the Moon at a given instant.
type MoonPhase struct {
	Time       time.Time // the instant this phase is evaluated at
	Fraction   float64   // illuminated fraction [0..1], 0=new, 1=full
	Elongation float64   // Sun-Moon angular separation in degrees [0..180]
	Waxing     bool      // true if waxing (illumination increasing), false if waning
	Name       string    // e.g. "New Moon", "Waxing Crescent", "First Quarter", ...
}

// PhaseWindow represents a continuous time interval where the Sun's altitude
// stays within a particular range (e.g. golden hour or blue hour).
type PhaseWindow struct {
	Start time.Time
	End   time.Time
}

// DaylightPhases holds the morning and evening windows for a given phase
// (e.g. golden hour or blue hour).
type DaylightPhases struct {
	// Morning is the interval after dawn / sunrise.
	Morning PhaseWindow
	// Evening is the interval before dusk / sunset.
	Evening PhaseWindow

	// HasMorning / HasEvening indicate whether the corresponding window
	// exists on this date at this location (high latitudes can be weird).
	HasMorning bool
	HasEvening bool
}

var (
	// ErrNoRiseNoSet is returned when a body does not rise or set on that date at that location.
	ErrNoRiseNoSet = errors.New("body does not rise or set on this date")

	// ErrNotImplemented is returned when that body isn't supported (yet).
	ErrNotImplemented = errors.New("not implemented for this body yet")
)

// RiseSetFor returns rise and set times for the given body and location on a date.
// For Level 1, only the Sun is implemented with decent accuracy (~±1 minute).
// The date's time zone is used for the returned times.
func RiseSetFor(body Body, loc Coordinates, date time.Time) (RiseSet, error) {
	switch body {
	case Sun:
		return sunRiseSet(loc, date)
	case Moon:
		return moonRiseSet(loc, date)
	default:
		return RiseSet{}, fmt.Errorf("unknown body %v", body)
	}
}

// moonRiseSet wraps the internal/moon implementation and converts UTC to the
// caller's desired time zone (taken from date.Location()).
func moonRiseSet(loc Coordinates, date time.Time) (RiseSet, error) {
	locTZ := date.Location()
	year, month, day := date.Date()

	// internal/moon returns a RiseSet (UTC times) plus ok flags
	rsMoonUTC, okRise, okSet := moon.RiseSetForDate(loc.Lat, loc.Lon, date)

	if !okRise && !okSet {
		return RiseSet{}, ErrNoRiseNoSet
	}

	var rs RiseSet

	if okRise {
		riseLocal := rsMoonUTC.Rise.In(locTZ)
		// Force the local calendar date to the requested one
		riseLocal = withLocalDate(riseLocal, year, month, day)
		rs.Rise = riseLocal
	}

	if okSet {
		setLocal := rsMoonUTC.Set.In(locTZ)
		// Same date-forcing for set
		setLocal = withLocalDate(setLocal, year, month, day)
		rs.Set = setLocal
	}

	return rs, nil
}

// SlideIntoSunset is your glorious convenience helper:
// it returns sunrise and sunset for the Sun at the given location and date.
func SlideIntoSunset(loc Coordinates, date time.Time) (RiseSet, error) {
	return RiseSetFor(Sun, loc, date)
}

// DaylightHours calculates the duration of daylight (time between sunrise and
// sunset) for the Sun at the given location and date. Returns the duration in
// hours as a float64.
//
// If the sun does not rise or set on the given date (e.g., polar regions), it
// returns 0 and ErrNoRiseNoSet. For polar day (sun never sets), you can detect
// this by checking if the sun is always above the horizon separately.
func DaylightHours(loc Coordinates, date time.Time) (float64, error) {
	rs, err := SlideIntoSunset(loc, date)
	if err != nil {
		return 0, err
	}

	duration := rs.Set.Sub(rs.Rise)
	return duration.Hours(), nil
}

// -----------------------------
// Sun wrapper around internal/sun
// -----------------------------

func sunRiseSet(loc Coordinates, date time.Time) (RiseSet, error) {
	locTZ := date.Location()
	year, month, day := date.Date()

	// Delegate to internal/sun which returns UTC times + flags.
	sunriseUTC, sunsetUTC, okRise, okSet := sun.RiseSetForDate(loc.Lat, loc.Lon, date, sun.StandardZenith)

	if !okRise && !okSet {
		return RiseSet{}, ErrNoRiseNoSet
	}

	var rs RiseSet

	if okRise {
		riseLocal := sunriseUTC.In(locTZ)
		// Force the date to match the requested local calendar date.
		riseLocal = withLocalDate(riseLocal, year, month, day)
		rs.Rise = riseLocal
	}

	if okSet {
		setLocal := sunsetUTC.In(locTZ)
		// Same: ensure the local date is the requested date.
		setLocal = withLocalDate(setLocal, year, month, day)
		rs.Set = setLocal
	}

	return rs, nil
}

// withLocalDate returns a copy of t but with its calendar date
// forced to (year, month, day), keeping the same clock time and location.
func withLocalDate(t time.Time, year int, month time.Month, day int) time.Time {
	loc := t.Location()
	return time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

// TwilightFor computes twilight times (dawn and dusk) of the given kind for
// a location and local calendar date. The returned RiseSet uses Rise as the
// "dawn" time (upward crossing of the twilight altitude) and Set as the
// "dusk" time (downward crossing).
//
// For example, TwilightCivil returns civil dawn (Rise) and civil dusk (Set)
// where the Sun's altitude crosses -6 degrees.
func TwilightFor(loc Coordinates, date time.Time, kind TwilightKind) (RiseSet, error) {
	locTZ := date.Location()
	year, month, day := date.Date()

	// Map TwilightKind to target altitude (degrees).
	var targetAlt float64
	switch kind {
	case TwilightCivil:
		targetAlt = -6.0
	case TwilightNautical:
		targetAlt = -12.0
	case TwilightAstronomical:
		targetAlt = -18.0
	default:
		return RiseSet{}, fmt.Errorf("unknown TwilightKind: %d", kind)
	}

	dawnUTC, duskUTC, okDawn, okDusk := sun.TwilightForDate(loc.Lat, loc.Lon, date, targetAlt)
	if !okDawn && !okDusk {
		return RiseSet{}, ErrNoRiseNoSet
	}

	var rs RiseSet

	if okDawn {
		dawnLocal := dawnUTC.In(locTZ)
		// Pin to the requested local calendar date for consistency.
		dawnLocal = withLocalDate(dawnLocal, year, month, day)
		rs.Rise = dawnLocal
	}

	if okDusk {
		duskLocal := duskUTC.In(locTZ)
		duskLocal = withLocalDate(duskLocal, year, month, day)
		rs.Set = duskLocal
	}

	return rs, nil
}

// GoldenHourFor computes the golden hour intervals for the given local
// calendar date and location. Golden hour is (approximately) defined as
// the period when the Sun's center altitude is between -4° and +6°.
//
// It returns DaylightPhases, where Morning is the interval after dawn
// (Sun climbing from -4° up to +6°) and Evening is the interval before
// dusk (Sun descending from +6° down to -4°).
//
// If neither morning nor evening golden hour exists (e.g. extreme
// high-latitude edge cases), ErrNoRiseNoSet is returned.
func GoldenHourFor(loc Coordinates, date time.Time) (DaylightPhases, error) {
	const (
		lowAlt  = -4.0 // degrees
		highAlt = 6.0  // degrees
	)

	locTZ := date.Location()
	year, month, day := date.Date()

	// We can reuse the Sun "Twilight" solver for arbitrary altitudes:
	// it returns the upward crossing (dawn-like) and downward crossing
	// (dusk-like) of targetAlt.
	mLow, eLow, okMLow, okELow := sun.TwilightForDate(loc.Lat, loc.Lon, date, lowAlt)
	mHigh, eHigh, okMHigh, okEHigh := sun.TwilightForDate(loc.Lat, loc.Lon, date, highAlt)

	var phases DaylightPhases

	// Morning golden hour: Sun climbing from lowAlt -> highAlt.
	if okMLow && okMHigh {
		start := mLow.In(locTZ)
		end := mHigh.In(locTZ)
		start = withLocalDate(start, year, month, day)
		end = withLocalDate(end, year, month, day)

		if end.After(start) {
			phases.Morning = PhaseWindow{
				Start: start,
				End:   end,
			}
			phases.HasMorning = true
		}
	}

	// Evening golden hour: Sun descending from highAlt -> lowAlt.
	if okEHigh && okELow {
		start := eHigh.In(locTZ)
		end := eLow.In(locTZ)
		start = withLocalDate(start, year, month, day)
		end = withLocalDate(end, year, month, day)

		if end.After(start) {
			phases.Evening = PhaseWindow{
				Start: start,
				End:   end,
			}
			phases.HasEvening = true
		}
	}

	if !phases.HasMorning && !phases.HasEvening {
		return DaylightPhases{}, ErrNoRiseNoSet
	}

	return phases, nil
}

// BlueHourFor computes the blue hour intervals for the given local calendar
// date and location. Blue hour here is defined as the period when the Sun's
// center altitude is between -6° and -4°.
//
// Morning blue hour is between the -6° and -4° upward crossings; evening
// blue hour is between the -4° and -6° downward crossings.
//
// If neither morning nor evening blue hour exists, ErrNoRiseNoSet is returned.
func BlueHourFor(loc Coordinates, date time.Time) (DaylightPhases, error) {
	const (
		lowAlt  = -6.0 // degrees
		highAlt = -4.0 // degrees
	)

	locTZ := date.Location()
	year, month, day := date.Date()

	mLow, eLow, okMLow, okELow := sun.TwilightForDate(loc.Lat, loc.Lon, date, lowAlt)
	mHigh, eHigh, okMHigh, okEHigh := sun.TwilightForDate(loc.Lat, loc.Lon, date, highAlt)

	var phases DaylightPhases

	// Morning blue hour: Sun climbing from lowAlt (-6°) -> highAlt (-4°).
	if okMLow && okMHigh {
		start := mLow.In(locTZ)
		end := mHigh.In(locTZ)
		start = withLocalDate(start, year, month, day)
		end = withLocalDate(end, year, month, day)

		if end.After(start) {
			phases.Morning = PhaseWindow{
				Start: start,
				End:   end,
			}
			phases.HasMorning = true
		}
	}

	// Evening blue hour: Sun descending from highAlt (-4°) -> lowAlt (-6°).
	if okEHigh && okELow {
		start := eHigh.In(locTZ)
		end := eLow.In(locTZ)
		start = withLocalDate(start, year, month, day)
		end = withLocalDate(end, year, month, day)

		if end.After(start) {
			phases.Evening = PhaseWindow{
				Start: start,
				End:   end,
			}
			phases.HasEvening = true
		}
	}

	if !phases.HasMorning && !phases.HasEvening {
		return DaylightPhases{}, ErrNoRiseNoSet
	}

	return phases, nil
}

// MoonPhaseAt computes the Moon's illuminated fraction and qualitative phase
// at the given time. Phase is a global property (independent of observer
// location), so we work in UTC internally and return the original time.
func MoonPhaseAt(t time.Time) (MoonPhase, error) {
	utc := t.UTC()

	// Moon: geocentric RA/Dec + distance (we only need RA/Dec here).
	mEq := moon.GeocentricEquatorialWithDistanceApprox(utc)

	// Sun: geocentric RA/Dec from the internal sun model.
	sEq := sun.GeocentricEquatorialApprox(utc)

	raSun := timeutil.Deg2Rad(sEq.RA)
	decSun := timeutil.Deg2Rad(sEq.Dec)
	raMoon := timeutil.Deg2Rad(mEq.RA)
	decMoon := timeutil.Deg2Rad(mEq.Dec)

	// Angular separation ψ between Sun and Moon:
	// cos ψ = sin δs sin δm + cos δs cos δm cos(αs - αm)
	dRA := raSun - raMoon
	cosPsi := math.Sin(decSun)*math.Sin(decMoon) +
		math.Cos(decSun)*math.Cos(decMoon)*math.Cos(dRA)

	// Clamp to handle numerical noise
	if cosPsi > 1 {
		cosPsi = 1
	} else if cosPsi < -1 {
		cosPsi = -1
	}

	psi := math.Acos(cosPsi)          // radians
	elongDeg := timeutil.Rad2Deg(psi) // 0..180 degrees

	// Illuminated fraction:
	// k = (1 - cos ψ) / 2
	fraction := 0.5 * (1 - cosPsi)
	if fraction < 0 {
		fraction = 0
	} else if fraction > 1 {
		fraction = 1
	}

	// Waxing vs waning: which side of the Sun is the Moon on?
	// sep = (RA_moon - RA_sun) normalized to [0,360).
	sepDeg := timeutil.Normalize360(mEq.RA - sEq.RA)
	waxing := sepDeg < 180.0

	name := classifyMoonPhaseName(fraction, waxing)

	return MoonPhase{
		Time:       t,
		Fraction:   fraction,
		Elongation: elongDeg,
		Waxing:     waxing,
		Name:       name,
	}, nil
}

func classifyMoonPhaseName(f float64, waxing bool) string {
	const (
		eps        = 0.01 // near 0 or 1
		quarterTol = 0.05 // fraction window around 0.5
	)

	switch {
	case f < eps:
		return "New Moon"
	case f > 1-eps:
		return "Full Moon"
	case math.Abs(f-0.5) < quarterTol:
		if waxing {
			return "First Quarter"
		}
		return "Last Quarter"
	case f < 0.5:
		if waxing {
			return "Waxing Crescent"
		}
		return "Waning Crescent"
	default: // f > 0.5 but not near 1
		if waxing {
			return "Waxing Gibbous"
		}
		return "Waning Gibbous"
	}
}
