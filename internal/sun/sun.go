package sun

import (
	"math"
	"time"

	"github.com/thurmanmarka/astroglide/internal/solver"
	"github.com/thurmanmarka/astroglide/internal/timeutil"
)

// StandardZenith is the commonly used zenith angle (in degrees) for sunrise/sunset:
// 90°50' ≈ 90.833°, accounting for refraction + Sun's apparent radius.
const StandardZenith = 90.833

// ApparentHorizonAltitudeSun is the altitude (in degrees) of the Sun's center
// when the apparent upper limb is on the horizon under "standard" conditions.
// Commonly taken as about -0.833 degrees.
const ApparentHorizonAltitudeSun = -0.833

// RiseSetForDate computes sunrise and sunset for the Sun on the given calendar date
// for an observer at lat, lon (degrees). Returned times are in UTC.
// `zenith` is in degrees; for standard sunrise/sunset use StandardZenith.
func RiseSetForDate(lat, lon float64, date time.Time, zenith float64) (sunriseUTC, sunsetUTC time.Time, okRise, okSet bool) {
	// Target altitude: h = 90° - Z.
	targetAlt := 90.0 - zenith
	return eventsForDateAtAltitude(lat, lon, date, targetAlt)
}

// TwilightForDate computes the times when the Sun crosses a given altitude
// (in degrees) during the local calendar day: "dawn" as the upward crossing,
// "dusk" as the downward crossing. Returned times are in UTC.
func TwilightForDate(lat, lon float64, date time.Time, targetAlt float64) (dawnUTC, duskUTC time.Time, okDawn, okDusk bool) {
	return eventsForDateAtAltitude(lat, lon, date, targetAlt)
}

// eventsForDateAtAltitude finds the times when the Sun's apparent altitude crosses
// targetAlt (degrees) during the local calendar day of `date` at (lat, lon).
// It returns the upward crossing (rise-like) and downward crossing (set-like)
// in UTC, along with booleans indicating if each event was found.
func eventsForDateAtAltitude(lat, lon float64, date time.Time, targetAlt float64) (riseUTC, setUTC time.Time, okRise, okSet bool) {
	loc := date.Location()
	year, month, day := date.Date()

	startLocal := time.Date(year, month, day, 0, 0, 0, 0, loc)
	endLocal := startLocal.Add(24 * time.Hour)

	altFunc := func(t time.Time) float64 {
		return apparentAltitude(lat, lon, t)
	}

	const (
		steps = 48 // samples across the day (every 30 minutes)
		tol   = 30 * time.Second
	)

	// Upward crossing (dawn/sunrise-type event)
	riseRes := solver.FindAltitudeEvent(altFunc, startLocal, endLocal, targetAlt, solver.CrossingUp, steps, tol)
	if riseRes.OK {
		riseUTC = riseRes.Time.UTC()
		okRise = true
	}

	// Downward crossing (dusk/sunset-type event)
	setRes := solver.FindAltitudeEvent(altFunc, startLocal, endLocal, targetAlt, solver.CrossingDown, steps, tol)
	if setRes.OK {
		setUTC = setRes.Time.UTC()
		okSet = true
	}

	return riseUTC, setUTC, okRise, okSet
}

// apparentAltitude computes the Sun's approximate geometric altitude (in degrees)
// at geographic location (lat, lon) at time t, using the solar RA/Dec model and
// a simple sidereal time approximation.
func apparentAltitude(lat, lon float64, t time.Time) float64 {
	// Geocentric equatorial coordinates of the Sun
	eq := GeocentricEquatorialApprox(t)

	raRad := timeutil.Deg2Rad(eq.RA)
	decRad := timeutil.Deg2Rad(eq.Dec)
	latRad := timeutil.Deg2Rad(lat)

	// Local sidereal time
	d := timeutil.DaysSinceJ2000(t)
	gmst := 280.46061837 + 360.98564736629*d
	lstDeg := timeutil.Normalize360(gmst + lon)
	lstRad := timeutil.Deg2Rad(lstDeg)

	// Hour angle H = LST - RA, normalized
	H := lstRad - raRad
	for H > math.Pi {
		H -= 2 * math.Pi
	}
	for H < -math.Pi {
		H += 2 * math.Pi
	}

	// Geometric altitude
	sinAlt := math.Sin(latRad)*math.Sin(decRad) + math.Cos(latRad)*math.Cos(decRad)*math.Cos(H)
	altRad := math.Asin(sinAlt)
	geomAlt := timeutil.Rad2Deg(altRad)

	// --- Refraction (experimental) ---
	const applyRefraction = false // flip to true to experiment

	if applyRefraction {
		ref := timeutil.ApproxRefraction(geomAlt)
		return geomAlt + ref
	}

	return geomAlt
}
