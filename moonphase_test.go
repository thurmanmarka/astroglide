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

// boolPtr returns a pointer to a bool, useful for optional waxing assertions.
func boolPtr(b bool) *bool { return &b }

// TestMoonPhaseAt_KnownPhases tests each named phase against published lunar
// calendar dates for 2025. Fractions are verified within loose tolerances to
// accommodate the level-1 approximate algorithm.
func TestMoonPhaseAt_KnownPhases(t *testing.T) {
	tests := []struct {
		name       string
		t          time.Time
		wantName   string
		wantWaxing *bool   // nil = don't check
		fracMin    float64
		fracMax    float64
	}{
		// Apr 27 2025 ~19:31 UTC — New Moon
		{
			name:     "New Moon 2025-04-27",
			t:        time.Date(2025, 4, 27, 20, 0, 0, 0, time.UTC),
			wantName: "New Moon",
			fracMin:  0.0,
			fracMax:  0.02,
		},
		// Apr 30 2025 — three days after new moon, halfway to first quarter
		{
			name:       "Waxing Crescent 2025-04-30",
			t:          time.Date(2025, 4, 30, 12, 0, 0, 0, time.UTC),
			wantName:   "Waxing Crescent",
			wantWaxing: boolPtr(true),
			fracMin:    0.05,
			fracMax:    0.45,
		},
		// May 4 2025 ~13:52 UTC — First Quarter
		{
			name:       "First Quarter 2025-05-04",
			t:          time.Date(2025, 5, 4, 14, 0, 0, 0, time.UTC),
			wantName:   "First Quarter",
			wantWaxing: boolPtr(true),
			fracMin:    0.45,
			fracMax:    0.55,
		},
		// May 8 2025 — between first quarter and full, waxing gibbous
		{
			name:       "Waxing Gibbous 2025-05-08",
			t:          time.Date(2025, 5, 8, 12, 0, 0, 0, time.UTC),
			wantName:   "Waxing Gibbous",
			wantWaxing: boolPtr(true),
			fracMin:    0.55,
			fracMax:    0.97,
		},
		// May 12 2025 ~16:56 UTC — Full Moon
		{
			name:     "Full Moon 2025-05-12",
			t:        time.Date(2025, 5, 12, 17, 0, 0, 0, time.UTC),
			wantName: "Full Moon",
			fracMin:  0.98,
			fracMax:  1.0,
		},
		// May 16 2025 — between full moon and last quarter, waning gibbous
		{
			name:       "Waning Gibbous 2025-05-16",
			t:          time.Date(2025, 5, 16, 12, 0, 0, 0, time.UTC),
			wantName:   "Waning Gibbous",
			wantWaxing: boolPtr(false),
			fracMin:    0.55,
			fracMax:    0.97,
		},
		// May 20 2025 ~11:59 UTC — Last Quarter
		{
			name:       "Last Quarter 2025-05-20",
			t:          time.Date(2025, 5, 20, 12, 0, 0, 0, time.UTC),
			wantName:   "Last Quarter",
			wantWaxing: boolPtr(false),
			fracMin:    0.45,
			fracMax:    0.55,
		},
		// May 23 2025 — three days after last quarter, waning crescent
		{
			name:       "Waning Crescent 2025-05-23",
			t:          time.Date(2025, 5, 23, 12, 0, 0, 0, time.UTC),
			wantName:   "Waning Crescent",
			wantWaxing: boolPtr(false),
			fracMin:    0.05,
			fracMax:    0.45,
		},
		// May 26 2025 ~23:02 UTC — New Moon
		{
			name:     "New Moon 2025-05-26",
			t:        time.Date(2025, 5, 26, 23, 0, 0, 0, time.UTC),
			wantName: "New Moon",
			fracMin:  0.0,
			fracMax:  0.02,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase, err := MoonPhaseAt(tt.t)
			if err != nil {
				t.Fatalf("MoonPhaseAt() error: %v", err)
			}

			if phase.Name != tt.wantName {
				t.Errorf("Name = %q, want %q (fraction=%.3f waxing=%v)",
					phase.Name, tt.wantName, phase.Fraction, phase.Waxing)
			}
			if phase.Fraction < tt.fracMin || phase.Fraction > tt.fracMax {
				t.Errorf("Fraction = %.3f, want [%.2f, %.2f]",
					phase.Fraction, tt.fracMin, tt.fracMax)
			}
			if tt.wantWaxing != nil && phase.Waxing != *tt.wantWaxing {
				t.Errorf("Waxing = %v, want %v", phase.Waxing, *tt.wantWaxing)
			}

			t.Logf("fraction=%.3f elongation=%.2f° waxing=%-5v name=%s",
				phase.Fraction, phase.Elongation, phase.Waxing, phase.Name)
		})
	}
}
