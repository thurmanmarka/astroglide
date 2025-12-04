package sun

import (
	"math"
	"time"

	"github.com/thurmanmarka/astroglide/internal/timeutil"
)

// Equatorial represents equatorial coordinates (right ascension and declination)
// in degrees. RA is in degrees (0–360).
type Equatorial struct {
	RA  float64 // right ascension, degrees
	Dec float64 // declination, degrees
}

// GeocentricEquatorialApprox returns an approximate geocentric RA/Dec for the Sun
// at the given time t.
//
// This is a standard low/medium-precision solar position model, good to
// arcminute-level accuracy in RA/Dec for many applications.
//
// Based on a simplified NOAA / Meeus-style algorithm:
//
//	g  = mean anomaly of the Sun
//	q  = mean longitude of the Sun
//	L  = ecliptic longitude of the Sun
//	eps = obliquity of the ecliptic
func GeocentricEquatorialApprox(t time.Time) Equatorial {
	d := timeutil.DaysSinceJ2000(t)

	// Mean anomaly of the Sun (deg)
	g := timeutil.Deg2Rad(357.529 + 0.98560028*d)

	// Mean longitude of the Sun (deg)
	q := timeutil.Deg2Rad(280.459 + 0.98564736*d)

	// Ecliptic longitude with equation of center
	L := q +
		timeutil.Deg2Rad(1.915)*math.Sin(g) +
		timeutil.Deg2Rad(0.020)*math.Sin(2*g)

	// Obliquity of the ecliptic (deg)
	eps := timeutil.Deg2Rad(23.439 - 0.00000036*d)

	// Convert to equatorial
	x := math.Cos(L)
	y := math.Cos(eps) * math.Sin(L)
	z := math.Sin(eps) * math.Sin(L)

	ra := math.Atan2(y, x)
	if ra < 0 {
		ra += 2 * math.Pi
	}
	dec := math.Asin(z)

	return Equatorial{
		RA:  timeutil.Rad2Deg(ra),
		Dec: timeutil.Rad2Deg(dec),
	}
}
