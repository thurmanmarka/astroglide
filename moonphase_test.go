package astroglide

import (
	"testing"
	"time"
)

func TestMoonPhaseAt_Debug(t *testing.T) {
	loc, _ := time.LoadLocation("America/Phoenix")

	// Pick a known full moon date/time near mid-2025
	tm := time.Date(2025, 5, 12, 0, 0, 0, 0, loc)

	phase, err := MoonPhaseAt(tm)
	if err != nil {
		t.Fatalf("MoonPhaseAt error: %v", err)
	}

	t.Logf("Time      : %v", phase.Time)
	t.Logf("Fraction  : %.3f", phase.Fraction)
	t.Logf("Elongation: %.2f°", phase.Elongation)
	t.Logf("Waxing    : %v", phase.Waxing)
	t.Logf("Name      : %s", phase.Name)
}
