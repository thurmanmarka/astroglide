package moon

import (
	"math"
	"time"

	"github.com/thurmanmarka/astroglide/internal/timeutil" // <-- replace with your module path
)

// Equatorial represents equatorial coordinates (right ascension and declination)
// in degrees. RA is in degrees (0–360) instead of hours to stay consistent with
// internal math helpers.
type Equatorial struct {
	RA  float64 // right ascension, degrees
	Dec float64 // declination, degrees
}

// GeocentricEquatorialApprox returns an approximate geocentric RA/Dec for the Moon
// at the given time t.
//
// This is a medium-precision model using a small set of dominant periodic terms
// in ecliptic longitude and latitude. It's significantly better than the
// ultra-simple model, but still not full ephemeris-grade.
//
// Roughly based on truncated Meeus-style series:
//
//	L'  = mean longitude of the Moon
//	M   = mean anomaly of the Sun
//	Mm  = mean anomaly of the Moon
//	D   = mean elongation of the Moon from the Sun
//	F   = argument of latitude of the Moon
func GeocentricEquatorialApprox(t time.Time) Equatorial {
	d := timeutil.DaysSinceJ2000(t)

	// Convert day count to degrees for the standard fundamental arguments.
	// All linear coefficients here are in deg/day.
	Lprime := 218.3164477 + 13.17639648*d // mean longitude of the Moon
	M := 357.5291092 + 0.98560028*d       // mean anomaly of the Sun
	Mm := 134.9633964 + 13.06499295*d     // mean anomaly of the Moon
	D := 297.8501921 + 12.19074912*d      // mean elongation from the Sun
	F := 93.2720950 + 13.22935024*d       // argument of latitude

	// Normalize to [0, 360)
	Lprime = timeutil.Normalize360(Lprime)
	M = timeutil.Normalize360(M)
	Mm = timeutil.Normalize360(Mm)
	D = timeutil.Normalize360(D)
	F = timeutil.Normalize360(F)

	// Convert to radians for trig.
	Lr := timeutil.Deg2Rad(Lprime)
	Mr := timeutil.Deg2Rad(M)
	Mmr := timeutil.Deg2Rad(Mm)
	Dr := timeutil.Deg2Rad(D)
	Fr := timeutil.Deg2Rad(F)

	// Ecliptic longitude λ (deg), using a handful of main terms.
	// λ ≈ L' + 6.289 sin(Mm) + 1.274 sin(2D − Mm)
	//      + 0.658 sin(2D) + 0.214 sin(2Mm) − 0.186 sin(M)
	//      − 0.114 sin(2F)
	lon := Lr +
		timeutil.Deg2Rad(6.289)*math.Sin(Mmr) +
		timeutil.Deg2Rad(1.274)*math.Sin(2*Dr-Mmr) +
		timeutil.Deg2Rad(0.658)*math.Sin(2*Dr) +
		timeutil.Deg2Rad(0.214)*math.Sin(2*Mmr) -
		timeutil.Deg2Rad(0.186)*math.Sin(Mr) -
		timeutil.Deg2Rad(0.114)*math.Sin(2*Fr)

	// Ecliptic latitude β (deg), similarly truncated:
	// β ≈ 5.128 sin(F) + 0.280 sin(Mm + F)
	//      + 0.277 sin(Mm − F) + 0.173 sin(2D − F)
	lat := timeutil.Deg2Rad(5.128)*math.Sin(Fr) +
		timeutil.Deg2Rad(0.280)*math.Sin(Mmr+Fr) +
		timeutil.Deg2Rad(0.277)*math.Sin(Mmr-Fr) +
		timeutil.Deg2Rad(0.173)*math.Sin(2*Dr-Fr)

	// Mean obliquity of the ecliptic ε (deg) – simple linear model.
	eps := timeutil.Deg2Rad(23.439291 - 0.0000137*d)

	// Convert from ecliptic (lon, lat) to equatorial (RA, Dec).
	x := math.Cos(lat) * math.Cos(lon)
	y := math.Cos(lat) * math.Sin(lon)
	z := math.Sin(lat)

	xEq := x
	yEq := y*math.Cos(eps) - z*math.Sin(eps)
	zEq := y*math.Sin(eps) + z*math.Cos(eps)

	ra := math.Atan2(yEq, xEq)
	if ra < 0 {
		ra += 2 * math.Pi
	}
	dec := math.Asin(zEq)

	return Equatorial{
		RA:  timeutil.Rad2Deg(ra),
		Dec: timeutil.Rad2Deg(dec),
	}
}
