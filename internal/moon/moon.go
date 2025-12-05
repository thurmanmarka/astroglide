package moon

import (
	"time"

	"math"

	"github.com/thurmanmarka/astroglide/internal/solver"
	"github.com/thurmanmarka/astroglide/internal/timeutil"
)

// ApparentHorizonAltitudeMoon returns the apparent altitude (deg) of the Moon's
// center when we define "rise/set" (upper limb on the horizon), including
// approximate refraction + limb correction and a small dependence on distance.
//
// The base value -0.90° was tuned for Phoenix 2025. We then apply a tiny
// distance-dependent tweak so that when the Moon is closer (larger angular
// size), we allow the center to sit slightly lower, and when it's farther,
// slightly higher.

const moonSetExtraDropDeg = 0.16

func ApparentHorizonAltitudeMoon(distanceKm float64) float64 {
	const (
		meanDistKm  = 384400.0 // average Earth–Moon distance
		baseHorizon = -0.90    // tuned at mean distance
		kScale      = 0.6      // deg per unit fractional distance
	)

	if distanceKm <= 0 {
		// Fallback to base if something weird happens
		return baseHorizon
	}

	// Fractional deviation from mean distance
	frac := (distanceKm - meanDistKm) / meanDistKm
	// When Moon is closer (frac < 0), horizon gets a bit more negative.
	// When farther (frac > 0), horizon gets a bit less negative.
	return baseHorizon - kScale*frac
}

// moonRefractionApprox returns an approximate atmospheric refraction correction
// for the Moon in degrees, to be ADDED to the geometric topocentric altitude.
//
// It uses a standard low-altitude formula (Meeus / Bennett approximation):
//
//	R ≈ 1.02 / tan((h + 10.3 / (h + 5.11))°)   [arcminutes]
//
// We clamp behavior outside the useful range to avoid wild values.
func moonRefractionApprox(altDeg float64) float64 {
	// Above ~90° or well below the horizon, we skip refraction.
	if altDeg > 90 || altDeg < -1 {
		return 0
	}
	// Avoid singularities near -5.11 in the denominator.
	h := altDeg
	if h < -0.5 {
		h = -0.5
	}
	Rarcmin := 1.02 / math.Tan(timeutil.Deg2Rad(h+10.3/(h+5.11)))
	return Rarcmin / 60.0 // convert arcminutes -> degrees
}

// RiseSet holds lunar rise and set times in UTC.
type RiseSet struct {
	Rise time.Time
	Set  time.Time
}

type EquatorialDistance struct {
	RA       float64 // degrees
	Dec      float64 // degrees
	Distance float64 // km
}

// RiseSetForDate computes the Moon's approximate rise and set times for a given
// calendar date and observer location.
//
// lat, lon in degrees (north/east positive, west negative).
// date can be any time on the calendar date you care about (its Location is
// used to define "midnight" for the search window).
//
// Returned Rise and Set are in UTC.
// okRise/okSet indicate whether rise/set events were found in that local date.
func RiseSetForDate(lat, lon float64, date time.Time) (rs RiseSet, okRise, okSet bool) {
	loc := date.Location()

	// Define the search window as the local calendar day: [00:00, 24:00).
	startLocal := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	endLocal := startLocal.Add(24 * time.Hour)

	// Rise altitude function: apparent altitude minus distance-dependent horizon.
	altFuncRise := func(t time.Time) float64 {
		eq := GeocentricEquatorialWithDistanceApprox(t)
		alt := apparentAltitude(lat, lon, t)
		horizon := ApparentHorizonAltitudeMoon(eq.Distance)
		return alt - horizon
	}

	// Set altitude function: same, but with a small extra drop in the horizon
	// so that the Moon "sets" slightly earlier, compensating for the observed
	// ~0.9 minute late bias.
	altFuncSet := func(t time.Time) float64 {
		eq := GeocentricEquatorialWithDistanceApprox(t)
		alt := apparentAltitude(lat, lon, t)
		horizon := ApparentHorizonAltitudeMoon(eq.Distance) + moonSetExtraDropDeg
		return alt - horizon
	}

	// We're solving for zero crossings of altFunc*(t).
	const targetAlt = 0.0

	const (
		steps = 48               // samples across the day
		tol   = 30 * time.Second // bisection tolerance
	)

	// Find rise (crossing upward).
	riseRes := solver.FindAltitudeEvent(
		altFuncRise,
		startLocal,
		endLocal,
		targetAlt,
		solver.CrossingUp,
		steps,
		tol,
	)
	if riseRes.OK {
		rs.Rise = riseRes.Time.UTC()
		okRise = true
	}

	// Find set (crossing downward).
	setRes := solver.FindAltitudeEvent(
		altFuncSet,
		startLocal,
		endLocal,
		targetAlt,
		solver.CrossingDown,
		steps,
		tol,
	)
	if setRes.OK {
		rs.Set = setRes.Time.UTC()
		okSet = true
	}

	return rs, okRise, okSet
}

// apparentAltitude computes the Moon's approximate apparent altitude (in degrees)
// at geographic location (lat, lon) at time t, using a simple geocentric RA/Dec
// model and a basic sidereal time approximation.
func apparentAltitude(lat, lon float64, t time.Time) float64 {
	// Geocentric RA/Dec + distance
	eq := GeocentricEquatorialWithDistanceApprox(t)

	raRad := timeutil.Deg2Rad(eq.RA)
	decRad := timeutil.Deg2Rad(eq.Dec)
	latRad := timeutil.Deg2Rad(lat)

	// Local sidereal time
	d := timeutil.DaysSinceJ2000(t)
	gmst := 280.46061837 + 360.98564736629*d
	lstDeg := timeutil.Normalize360(gmst + lon)
	lstRad := timeutil.Deg2Rad(lstDeg)

	// Geocentric hour angle H
	H := lstRad - raRad
	for H > math.Pi {
		H -= 2 * math.Pi
	}
	for H < -math.Pi {
		H += 2 * math.Pi
	}

	// --- Topocentric correction via horizontal parallax ---
	pi := horizontalParallax(eq.Distance) // radians

	sinφ := math.Sin(latRad)
	cosφ := math.Cos(latRad)

	// Meeus approximate factors for observer at sea level.
	rhoSinφ := 0.99883 * sinφ
	rhoCosφ := 0.99883 * cosφ

	sinδ := math.Sin(decRad)
	cosδ := math.Cos(decRad)
	sinH := math.Sin(H)
	cosH := math.Cos(H)
	sinπ := math.Sin(pi)

	// Δα (correction to RA)
	deltaAlpha := math.Atan2(
		-rhoCosφ*sinπ*sinH,
		cosδ-rhoCosφ*sinπ*cosH,
	)

	// Topocentric RA and Dec
	raTopo := raRad + deltaAlpha
	decTopo := math.Atan2(
		sinδ-rhoSinφ*sinπ,
		cosδ-rhoCosφ*sinπ*cosH,
	)

	// New hour angle with topocentric RA
	Ht := lstRad - raTopo
	for Ht > math.Pi {
		Ht -= 2 * math.Pi
	}
	for Ht < -math.Pi {
		Ht += 2 * math.Pi
	}

	// Topocentric altitude
	sinAlt := sinφ*math.Sin(decTopo) + cosφ*math.Cos(decTopo)*math.Cos(Ht)
	altRad := math.Asin(sinAlt)

	// Convert to degrees
	altDeg := timeutil.Rad2Deg(altRad)

	// Apply Moon-specific atmospheric refraction near the horizon.
	// altDeg += moonRefractionApprox(altDeg)

	return altDeg
}

func horizontalParallax(distanceKm float64) float64 {
	const earthRadiusKm = 6378.14
	if distanceKm <= earthRadiusKm {
		// ridiculously close / invalid, just clamp
		return timeutil.Deg2Rad(1.0) // ~1° in radians as a safe default
	}
	return math.Asin(earthRadiusKm / distanceKm) // radians
}

func GeocentricEquatorialWithDistanceApprox(t time.Time) EquatorialDistance {
	// Use your existing RA/Dec model.
	eq := GeocentricEquatorialApprox(t)

	// Compute only lunar distance Δ (km) with a truncated Meeus-style series.
	T := timeutil.JulianCenturies(t)

	D := timeutil.Normalize360(297.8501921 + 445267.1114034*T)  // mean elongation
	M1 := timeutil.Normalize360(134.9633964 + 477198.8675055*T) // Moon mean anomaly

	Dr := timeutil.Deg2Rad(D)
	M1r := timeutil.Deg2Rad(M1)

	// Approximate Earth–Moon distance in km.
	delta := 385000.56 -
		20905.0*math.Cos(M1r) -
		3699.0*math.Cos(2*Dr-M1r) -
		2956.0*math.Cos(2*Dr) -
		570.0*math.Cos(2*M1r) -
		246.0*math.Cos(2*Dr+M1r)

	return EquatorialDistance{
		RA:       eq.RA,
		Dec:      eq.Dec,
		Distance: delta,
	}
}
