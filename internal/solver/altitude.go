package solver

import (
	"time"
)

// AltitudeFunc returns altitude in degrees at time t (topocentric).
type AltitudeFunc func(t time.Time) float64

// EventType describes whether we are looking for a rising or setting event.
type EventType int

const (
	// CrossingUp means altitude is increasing through the target value (rise).
	CrossingUp EventType = iota
	// CrossingDown means altitude is decreasing through the target value (set).
	CrossingDown
)

// Result holds the output of a altitude event search.
type Result struct {
	Time time.Time // approximate time of the event
	OK   bool      // true if an event was found
}

// FindAltitudeEvent searches for a time in [start, end] where the altitude function
// crosses targetDeg in the direction specified by eventType.
// It uses a simple bracket-then-bisect strategy.
//
// This is generic and can be used for Sun, Moon, twilight, etc.
// For Level 1 we don't use it yet; it's here as a building block for Level 2/3.
func FindAltitudeEvent(f AltitudeFunc, start, end time.Time, targetDeg float64, eventType EventType, steps int, tol time.Duration) Result {
	if !start.Before(end) {
		return Result{OK: false}
	}
	if steps < 2 {
		steps = 2
	}

	// Step 1: sample across [start, end] to find a sign change
	// in (altitude - target)
	interval := end.Sub(start) / time.Duration(steps-1)

	var (
		prevT   = start
		prevAlt = f(prevT) - targetDeg
	)

	for i := 1; i < steps; i++ {
		t := start.Add(time.Duration(i) * interval)
		alt := f(t) - targetDeg

		if hasCrossing(prevAlt, alt, eventType) {
			// We have a bracket [prevT, t]
			return bisect(f, prevT, t, targetDeg, eventType, tol)
		}

		prevT, prevAlt = t, alt
	}

	// No crossing found.
	return Result{OK: false}
}

func hasCrossing(a1, a2 float64, eventType EventType) bool {
	switch eventType {
	case CrossingUp:
		// a1 < 0, a2 >= 0
		return a1 < 0 && a2 >= 0
	case CrossingDown:
		// a1 > 0, a2 <= 0
		return a1 > 0 && a2 <= 0
	default:
		// Generic sign change
		return a1*a2 <= 0
	}
}

func bisect(f AltitudeFunc, a, b time.Time, targetDeg float64, eventType EventType, tol time.Duration) Result {
	var (
		altA = f(a) - targetDeg
		altB = f(b) - targetDeg
	)

	// Simple safety check
	if !hasCrossing(altA, altB, eventType) {
		return Result{OK: false}
	}

	for b.Sub(a) > tol {
		mid := a.Add(b.Sub(a) / 2)
		altM := f(mid) - targetDeg

		if hasCrossing(altA, altM, eventType) {
			b = mid
			altB = altM
		} else {
			a = mid
			altA = altM
		}
	}

	return Result{
		Time: a.Add(b.Sub(a) / 2),
		OK:   true,
	}
}
