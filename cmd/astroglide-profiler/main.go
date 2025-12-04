package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/thurmanmarka/astroglide"
)

type stats struct {
	count int
	sum   float64
	min   float64
	max   float64
}

type signedStats struct {
	count int
	sum   float64
	min   float64
	max   float64
}

func (s *signedStats) add(v float64) {
	if math.IsNaN(v) {
		return
	}
	if s.count == 0 {
		s.min, s.max = v, v
	} else {
		if v < s.min {
			s.min = v
		}
		if v > s.max {
			s.max = v
		}
	}
	s.sum += v
	s.count++
}

func (s *stats) add(v float64) {
	if math.IsNaN(v) {
		return
	}
	if s.count == 0 {
		s.min, s.max = v, v
	} else {
		if v < s.min {
			s.min = v
		}
		if v > s.max {
			s.max = v
		}
	}
	s.sum += v
	s.count++
}

func (s *stats) avg() float64 {
	if s.count == 0 {
		return math.NaN()
	}
	return s.sum / float64(s.count)
}

func diffMinutes(a, b time.Time) float64 {
	// If either time is zero, treat as "no data".
	if a.IsZero() || b.IsZero() {
		return math.NaN()
	}

	d := a.Sub(b)
	if d < 0 {
		d = -d
	}
	return d.Minutes()
}

func diffMinutesSigned(a, b time.Time) float64 {
	if a.IsZero() || b.IsZero() {
		return math.NaN()
	}
	return a.Sub(b).Minutes() // can be negative or positive
}

func (s *signedStats) mean() float64 {
	if s.count == 0 {
		return math.NaN()
	}
	return s.sum / float64(s.count)
}

// CSV format:
//
// date,rise,set
// 2025-01-01,07:32,17:12
// 2025-01-02,07:32,17:13
//
// - date is YYYY-MM-DD
// - rise/set are local times in HH:MM (24-hour clock)
// - All times are assumed to be in the timezone given by -tz.
func main() {
	var (
		lat      = flag.Float64("lat", 0, "latitude in degrees (north positive)")
		lon      = flag.Float64("lon", 0, "longitude in degrees (east positive, west negative)")
		tzName   = flag.String("tz", "UTC", "IANA time zone name (e.g. America/Phoenix)")
		bodyS    = flag.String("body", "sun", "celestial body: sun or moon")
		year     = flag.Int("year", 0, "year of the ephemeris data (optional, used for sanity checks)")
		refCSV   = flag.String("refcsv", "", "path to reference ephemeris CSV file (date,rise,set)")
		verbose  = flag.Bool("verbose", false, "log per-day errors instead of only summary")
		twilight = flag.String("twilight", "", "twilight kind: civil, nautical, astronomical (Sun only)")
		outCSV   = flag.String("outcsv", "", "optional path to write per-row error CSV")
	)

	flag.Parse()

	if *refCSV == "" {
		log.Fatalf("missing -refcsv (path to reference CSV)")
	}

	loc, err := time.LoadLocation(*tzName)
	if err != nil {
		log.Fatalf("failed to load timezone %q: %v", *tzName, err)
	}

	var body astroglide.Body
	switch strings.ToLower(*bodyS) {
	case "sun":
		body = astroglide.Sun
	case "moon":
		body = astroglide.Moon
	default:
		log.Fatalf("unsupported body %q (use sun or moon)", *bodyS)
	}

	useTwilight := false
	var twilightKind astroglide.TwilightKind

	if *twilight != "" {
		useTwilight = true

		if strings.ToLower(*bodyS) != "sun" {
			log.Fatalf("twilight mode only supported for -body sun")
		}

		switch strings.ToLower(*twilight) {
		case "civil":
			twilightKind = astroglide.TwilightCivil
		case "nautical":
			twilightKind = astroglide.TwilightNautical
		case "astronomical":
			twilightKind = astroglide.TwilightAstronomical
		default:
			log.Fatalf("unknown twilight kind %q (use civil, nautical, or astronomical)", *twilight)
		}
	}

	// Build mode description once
	modeDesc := strings.ToUpper(*bodyS)
	if useTwilight {
		switch twilightKind {
		case astroglide.TwilightCivil:
			modeDesc = "SUN (CIVIL TWILIGHT)"
		case astroglide.TwilightNautical:
			modeDesc = "SUN (NAUTICAL TWILIGHT)"
		case astroglide.TwilightAstronomical:
			modeDesc = "SUN (ASTRONOMICAL TWILIGHT)"
		default:
			modeDesc = "SUN (UNKNOWN TWILIGHT)"
		}
	}

	var outWriter *csv.Writer

	if *outCSV != "" {
		outFile, err := os.Create(*outCSV)
		if err != nil {
			log.Fatalf("failed to create outcsv %q: %v", *outCSV, err)
		}
		defer outFile.Close()

		outWriter = csv.NewWriter(outFile)
		defer outWriter.Flush()

		// Header row
		if err := outWriter.Write([]string{
			"date",
			"body",
			"mode",
			"rise_err",
			"set_err",
			"rise_signed",
			"set_signed",
			"phase_fraction",
			"phase_name",
			"phase_elongation",
			"phase_waxing",
		}); err != nil {
			log.Fatalf("failed to write outcsv header: %v", err)
		}
	}

	if *lat == 0 && *lon == 0 {
		log.Println("warning: lat=0 lon=0 (Gulf of Guinea). Did you mean to set -lat/-lon?")
	}

	f, err := os.Open(*refCSV)
	if err != nil {
		log.Fatalf("failed to open refcsv %q: %v", *refCSV, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1 // allow variable, we validate

	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("failed to read CSV: %v", err)
	}

	if len(records) == 0 {
		log.Fatalf("empty CSV file")
	}

	// If first row looks like a header, skip it.
	startIdx := 0
	if len(records[0]) >= 1 && strings.EqualFold(records[0][0], "date") {
		startIdx = 1
	}

	var (
		riseStats       stats
		setStats        stats
		riseSignedStats signedStats
		setSignedStats  signedStats
		skipped         int
		totalRows       int
	)

	coords := astroglide.Coordinates{
		Lat: *lat,
		Lon: *lon,
	}

	for i := startIdx; i < len(records); i++ {
		row := records[i]
		totalRows++

		if len(row) < 3 {
			log.Printf("row %d: expected at least 3 columns (date,rise,set), got %d, skipping", i+1, len(row))
			skipped++
			continue
		}
		dateStr := strings.TrimSpace(row[0])
		riseStr := strings.TrimSpace(row[1])
		setStr := strings.TrimSpace(row[2])

		// Parse the date.
		date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			log.Printf("row %d: invalid date %q: %v, skipping", i+1, dateStr, err)
			skipped++
			continue
		}

		if *year != 0 && date.Year() != *year {
			// Just warn; don't skip.
			log.Printf("row %d: warning: date %s not in year %d", i+1, dateStr, *year)
		}

		// Parse expected rise.
		refRise, err := parseLocalTime(date, riseStr, loc)
		if err != nil {
			log.Printf("row %d: invalid rise time %q: %v, skipping", i+1, riseStr, err)
			skipped++
			continue
		}

		// Parse expected set.
		refSet, err := parseLocalTime(date, setStr, loc)
		if err != nil {
			log.Printf("row %d: invalid set time %q: %v, skipping", i+1, setStr, err)
			skipped++
			continue
		}

		// Compute astroglide rise/set.
		var rs astroglide.RiseSet

		if useTwilight {
			// In twilight mode, interpret CSV "rise" as dawn and "set" as dusk.
			rs, err = astroglide.TwilightFor(coords, date, twilightKind)
		} else {
			rs, err = astroglide.RiseSetFor(body, coords, date)
		}

		if err != nil {
			log.Printf("row %d: astroglide error: %v, skipping", i+1, err)
			skipped++
			continue
		}

		// Compare in local time zone.
		gotRise := rs.Rise.In(loc)
		gotSet := rs.Set.In(loc)

		riseErr := diffMinutes(gotRise, refRise)
		setErr := diffMinutes(gotSet, refSet)

		riseStats.add(riseErr)
		setStats.add(setErr)

		riseSigned := diffMinutesSigned(gotRise, refRise)
		setSigned := diffMinutesSigned(gotSet, refSet)
		riseSignedStats.add(riseSigned)
		setSignedStats.add(setSigned)

		if *verbose {
			fmt.Printf("%s %s: rise err=%.2f min (got=%s ref=%s), set err=%.2f min (got=%s ref=%s)\n",
				dateStr, modeDesc,
				riseErr, gotRise.Format("15:04"), refRise.Format("15:04"),
				setErr, gotSet.Format("15:04"), refSet.Format("15:04"))
		}

		// --- Optional Moon phase info (for Moon runs only) ---
		var phaseFraction, phaseName, phaseElongation, phaseWaxing string

		if strings.EqualFold(*bodyS, "moon") {
			// Evaluate phase at local noon for this date.
			phaseTime := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, loc)
			mp, err := astroglide.MoonPhaseAt(phaseTime)
			if err != nil {
				log.Printf("row %d: failed to compute Moon phase: %v", i+1, err)
			} else {
				phaseFraction = fmt.Sprintf("%.6f", mp.Fraction)
				phaseName = mp.Name
				phaseElongation = fmt.Sprintf("%.3f", mp.Elongation)
				if mp.Waxing {
					phaseWaxing = "waxing"
				} else {
					phaseWaxing = "waning"
				}
			}
		}

		// --- Write per-row CSV if requested ---
		if outWriter != nil {
			rec := []string{
				dateStr,
				strings.ToUpper(*bodyS),
				modeDesc,
				fmt.Sprintf("%.6f", riseErr),
				fmt.Sprintf("%.6f", setErr),
				fmt.Sprintf("%.6f", riseSigned),
				fmt.Sprintf("%.6f", setSigned),
				phaseFraction,
				phaseName,
				phaseElongation,
				phaseWaxing,
			}
			if err := outWriter.Write(rec); err != nil {
				log.Printf("row %d: failed to write outcsv: %v", i+1, err)
			}
		}
	}

	fmt.Println("=== astroglide profiler summary ===")
	fmt.Printf("Mode:   %s\n", modeDesc)
	fmt.Printf("Lat/Lon: %.4f / %.4f\n", *lat, *lon)
	fmt.Printf("TZ:     %s\n", loc.String())
	fmt.Printf("Rows:   %d (processed), %d skipped\n", totalRows-skipped, skipped)

	if riseStats.count == 0 {
		fmt.Println("No valid rows to compute stats.")
		return
	}

	fmt.Println("\nRise error (minutes):")
	fmt.Printf("  count: %d\n", riseStats.count)
	fmt.Printf("  min:   %.3f\n", riseStats.min)
	fmt.Printf("  max:   %.3f\n", riseStats.max)
	fmt.Printf("  avg:   %.3f\n", riseStats.avg())

	fmt.Println("\nSet error (minutes):")
	fmt.Printf("  count: %d\n", setStats.count)
	fmt.Printf("  min:   %.3f\n", setStats.min)
	fmt.Printf("  max:   %.3f\n", setStats.max)
	fmt.Printf("  avg:   %.3f\n", setStats.avg())

	fmt.Println("\nRise signed error (minutes, our - ref):")
	fmt.Printf("  count: %d\n", riseSignedStats.count)
	fmt.Printf("  min:   %.3f\n", riseSignedStats.min)
	fmt.Printf("  max:   %.3f\n", riseSignedStats.max)
	fmt.Printf("  mean:  %.3f\n", riseSignedStats.mean())

	fmt.Println("\nSet signed error (minutes, our - ref):")
	fmt.Printf("  count: %d\n", setSignedStats.count)
	fmt.Printf("  min:   %.3f\n", setSignedStats.min)
	fmt.Printf("  max:   %.3f\n", setSignedStats.max)
	fmt.Printf("  mean:  %.3f\n", setSignedStats.mean())
}

func parseLocalTime(date time.Time, hhmm string, loc *time.Location) (time.Time, error) {
	// Expect HH:MM (optionally HH:MM:SS).
	layout := "15:04"
	if strings.Count(hhmm, ":") == 2 {
		layout = "15:04:05"
	}

	parsed, err := time.ParseInLocation(layout, hhmm, loc)
	if err != nil {
		return time.Time{}, err
	}
	// Combine parsed clock time with date.
	return time.Date(date.Year(), date.Month(), date.Day(),
		parsed.Hour(), parsed.Minute(), parsed.Second(), 0, loc), nil
}
