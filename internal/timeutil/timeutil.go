package timeutil

import (
	"math"
	"time"
)

// DayOfYear returns the 1-based day of year for the given date.
func DayOfYear(year int, month time.Month, day int) int {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return t.YearDay()
}

// FractionalHoursToTime converts fractional hours [0,24) into a UTC time on the given date.
func FractionalHoursToTime(year int, month time.Month, day int, h float64) time.Time {
	// Interpret h as hours *since midnight UTC* on the given date.
	// h can be negative or >24; we let time.Add handle day rollover.
	base := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	// Convert fractional hours to a duration.
	seconds := h * 3600
	// Round to nearest second to avoid crazy nanosecond noise.
	sec := int64(math.Round(seconds))

	return base.Add(time.Duration(sec) * time.Second)
}

// -----------------------------
// Time relative to J2000
// -----------------------------

// j2000 is the J2000.0 epoch: 2000-01-01 12:00:00 UTC.
var j2000 = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)

// DaysSinceJ2000 returns the number of (UTC) days since the J2000.0 epoch.
//
// This is an approximation suitable for low/medium-precision astronomy.
// For high-precision work you might want a true TT-based Julian day, but
// this is fine for our current purposes.
func DaysSinceJ2000(t time.Time) float64 {
	return t.UTC().Sub(j2000).Hours() / 24.0
}

func JulianDay(t time.Time) float64 {
	u := t.UTC()
	year, month, day := u.Date()
	hour := float64(u.Hour()) +
		float64(u.Minute())/60.0 +
		float64(u.Second())/3600.0 +
		float64(u.Nanosecond())/(3600.0*1e9)

	y := year
	m := int(month)

	if m <= 2 {
		y -= 1
		m += 12
	}

	A := y / 100
	B := 2 - A + A/4

	jd := math.Floor(365.25*float64(y+4716)) +
		math.Floor(30.6001*float64(m+1)) +
		float64(day) + float64(B) - 1524.5 +
		hour/24.0

	return jd
}

// JulianCenturies returns centuries since J2000.0.
func JulianCenturies(t time.Time) float64 {
	jd := JulianDay(t)
	return (jd - 2451545.0) / 36525.0
}

// -----------------------------
// Basic degree/radian helpers and trig with degree inputs.
// -----------------------------

func Deg2Rad(d float64) float64 {
	return d * math.Pi / 180.0
}

func Rad2Deg(r float64) float64 {
	return r * 180.0 / math.Pi
}

func SinD(deg float64) float64 {
	return math.Sin(Deg2Rad(deg))
}

func CosD(deg float64) float64 {
	return math.Cos(Deg2Rad(deg))
}

func TanD(deg float64) float64 {
	return math.Tan(Deg2Rad(deg))
}

func Normalize360(d float64) float64 {
	d = math.Mod(d, 360.0)
	if d < 0 {
		d += 360.0
	}
	return d
}

func Normalize24(h float64) float64 {
	h = math.Mod(h, 24.0)
	if h < 0 {
		h += 24.0
	}
	return h
}

// ApproxRefraction returns an approximation of atmospheric refraction (in
// degrees) at a given apparent altitude altDeg (degrees) under standard
// conditions.
//
// Positive return means "add this to the geometric altitude to get apparent
// altitude". This uses a Saemundsson-style formula and is reasonably accurate
// for altitudes near the horizon and above.
//
// Ref: often quoted as:
//
//	R (arcmin) ≈ 1.02 / tan( (alt + 10.3 / (alt + 5.11)) in degrees )
func ApproxRefraction(altDeg float64) float64 {
	// Below -1° we just bail; the formula goes weird and refraction isn't
	// meaningfully defined for "deep below the horizon" in this context.
	if altDeg < -1.0 {
		return 0
	}

	// To avoid division by zero or absurd numbers near -5 to -2 degrees,
	// clamp altDeg a bit when very low.
	alt := altDeg
	if alt < -0.5 {
		alt = -0.5
	}

	// Compute the argument in radians for tan().
	// Note: (alt + 10.3/(alt+5.11)) is in degrees.
	argDeg := alt + 10.3/(alt+5.11)
	argRad := Deg2Rad(argDeg)

	t := math.Tan(argRad)
	if t == 0 {
		return 0
	}

	// Result is in arcminutes; convert to degrees.
	R_arcmin := 1.02 / t
	return R_arcmin / 60.0
}
