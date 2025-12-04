package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/thurmanmarka/astroglide"
)

func main() {
	log.SetFlags(0)

	// Backwards-compatible behavior:
	// - If no args or first arg starts with "-", run rise/set mode (old style).
	// - Otherwise treat the first arg as a subcommand (e.g. "phase").
	if len(os.Args) < 2 || strings.HasPrefix(os.Args[1], "-") {
		runRiseSet(os.Args[1:])
		return
	}

	switch os.Args[1] {
	case "phase":
		runPhase(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `astroglide – astro goodies

Usage:
  astroglide [flags]           # Sun/Moon rise/set (legacy/default mode)
  astroglide phase [flags]     # Moon phase / illumination

Default mode flags (rise/set):
  -lat float
        latitude in degrees (north positive)
  -lon float
        longitude in degrees (east positive, west negative)
  -date string
        date in YYYY-MM-DD (optional, defaults to today in local time)
  -body string
        celestial body: sun or moon (default "sun")
  -event string
        event: rise, set, or both (default "both")
  -json
        output result as JSON

For phase mode:
  astroglide phase -h
`)
}

// ---------------------
// Rise/set (default) mode
// ---------------------

func runRiseSet(args []string) {
	fs := flag.NewFlagSet("astroglide", flag.ExitOnError)

	lat := fs.Float64("lat", 0, "latitude in degrees (north positive)")
	lon := fs.Float64("lon", 0, "longitude in degrees (east positive, west negative)")
	dateS := fs.String("date", "", "date in YYYY-MM-DD (optional, defaults to today in local time)")
	bodyS := fs.String("body", "sun", "celestial body: sun or moon")
	event := fs.String("event", "both", "event: rise, set, or both")
	jsonOut := fs.Bool("json", false, "output result as JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: astroglide [flags]

Flags:
`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

	if *lat == 0 && *lon == 0 {
		log.Println("warning: lat=0 lon=0 (Gulf of Guinea). Use -lat and -lon to set a real location.")
	}

	// Default date: today in local time.
	var date time.Time
	if *dateS == "" {
		now := time.Now()
		date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	} else {
		var err error
		date, err = time.ParseInLocation("2006-01-02", *dateS, time.Local)
		if err != nil {
			log.Fatalf("invalid -date %q: %v", *dateS, err)
		}
	}

	// Parse body
	var body astroglide.Body
	switch strings.ToLower(*bodyS) {
	case "sun":
		body = astroglide.Sun
	case "moon":
		body = astroglide.Moon
	default:
		log.Fatalf("unsupported body %q (use sun or moon)", *bodyS)
	}

	coords := astroglide.Coordinates{
		Lat: *lat,
		Lon: *lon,
		// Elevation reserved for future use
	}

	rs, err := astroglide.RiseSetFor(body, coords, date)
	if err != nil {
		log.Fatalf("error computing rise/set: %v", err)
	}

	if *jsonOut {
		printJSON(body, coords, date, *event, rs)
	} else {
		printHuman(body, coords, date, *event, rs)
	}
}

// ---------------------
// Phase subcommand
// ---------------------

func runPhase(args []string) {
	fs := flag.NewFlagSet("phase", flag.ExitOnError)

	tzName := fs.String("tz", "UTC", "IANA time zone name (e.g. America/Phoenix)")
	timeStr := fs.String("time", "", "Time in RFC3339 or 'YYYY-MM-DDTHH:MM' (optional, defaults to now in tz)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: astroglide phase [flags]

Flags:
`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}

	loc, err := time.LoadLocation(*tzName)
	if err != nil {
		log.Fatalf("invalid time zone %q: %v", *tzName, err)
	}

	var tLocal time.Time
	if *timeStr == "" {
		// Now in the given time zone
		tLocal = time.Now().In(loc)
	} else {
		// Try a couple of common formats
		layouts := []string{
			time.RFC3339,
			"2006-01-02T15:04",
			"2006-01-02 15:04",
			"2006-01-02",
		}
		var parseErr error
		for _, layout := range layouts {
			tLocal, parseErr = time.ParseInLocation(layout, *timeStr, loc)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			log.Fatalf("could not parse -time %q: %v", *timeStr, parseErr)
		}
	}

	phase, err := astroglide.MoonPhaseAt(tLocal)
	if err != nil {
		log.Fatalf("MoonPhaseAt failed: %v", err)
	}

	fmt.Printf("Moon phase at %s (%s)\n", phase.Time.Format(time.RFC3339), loc.String())
	fmt.Printf("  Name       : %s\n", phase.Name)
	fmt.Printf("  Fraction   : %.3f (%.1f%% illuminated)\n", phase.Fraction, phase.Fraction*100)
	fmt.Printf("  Elongation : %.2f°\n", phase.Elongation)
	if phase.Waxing {
		fmt.Printf("  Trend      : Waxing (illumination increasing)\n")
	} else {
		fmt.Printf("  Trend      : Waning (illumination decreasing)\n")
	}
}

// ---------------------
// Shared helpers
// ---------------------

func printHuman(body astroglide.Body, coords astroglide.Coordinates, date time.Time, event string, rs astroglide.RiseSet) {
	bodyName := map[astroglide.Body]string{
		astroglide.Sun:  "Sun",
		astroglide.Moon: "Moon",
	}[body]

	fmt.Printf("%s rise/set for lat=%.6f lon=%.6f\n", bodyName, coords.Lat, coords.Lon)
	fmt.Printf("Date: %s (%s)\n\n", date.Format("2006-01-02"), date.Location())

	event = strings.ToLower(event)
	switch event {
	case "rise":
		fmt.Printf("Rise: %s\n", rs.Rise.Format(time.RFC3339))
	case "set":
		fmt.Printf("Set:  %s\n", rs.Set.Format(time.RFC3339))
	case "both":
		fmt.Printf("Rise: %s\n", rs.Rise.Format(time.RFC3339))
		fmt.Printf("Set:  %s\n", rs.Set.Format(time.RFC3339))
	default:
		fmt.Fprintf(os.Stderr, "unknown event %q, showing both\n", event)
		fmt.Printf("Rise: %s\n", rs.Rise.Format(time.RFC3339))
		fmt.Printf("Set:  %s\n", rs.Set.Format(time.RFC3339))
	}
}

type jsonOutput struct {
	Body      string             `json:"body"`
	Latitude  float64            `json:"latitude"`
	Longitude float64            `json:"longitude"`
	Date      string             `json:"date"` // YYYY-MM-DD
	Rise      *time.Time         `json:"rise,omitempty"`
	Set       *time.Time         `json:"set,omitempty"`
	Timezone  string             `json:"timezone"`
	Raw       astroglide.RiseSet `json:"raw"`
}

func printJSON(body astroglide.Body, coords astroglide.Coordinates, date time.Time, event string, rs astroglide.RiseSet) {
	bodyName := map[astroglide.Body]string{
		astroglide.Sun:  "sun",
		astroglide.Moon: "moon",
	}[body]

	out := jsonOutput{
		Body:      bodyName,
		Latitude:  coords.Lat,
		Longitude: coords.Lon,
		Date:      date.Format("2006-01-02"),
		Timezone:  date.Location().String(),
		Raw:       rs,
	}

	e := strings.ToLower(event)
	switch e {
	case "rise":
		out.Rise = &rs.Rise
	case "set":
		out.Set = &rs.Set
	case "both":
		out.Rise = &rs.Rise
		out.Set = &rs.Set
	default:
		out.Rise = &rs.Rise
		out.Set = &rs.Set
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		log.Fatalf("failed to encode JSON: %v", err)
	}
}
