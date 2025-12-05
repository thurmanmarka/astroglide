package astroglide_test

import (
	"fmt"
	"time"

	"github.com/thurmanmarka/astroglide"
	// <-- replace with your actual module path
)

// ExampleSlideIntoSunset demonstrates computing sunrise and sunset for a location.
func ExampleSlideIntoSunset() {
	loc := astroglide.Coordinates{
		Lat: 40.7128,  // New York City latitude
		Lon: -74.0060, // New York City longitude
	}

	// Use a local date; the time zone is taken from the date's Location.
	locNY, _ := time.LoadLocation("America/New_York")
	date := time.Date(2025, time.November, 30, 0, 0, 0, 0, locNY)

	rs, err := astroglide.SlideIntoSunset(loc, date)
	if err != nil {
		panic(err)
	}

	fmt.Println("Sunrise:", rs.Rise.Format(time.RFC3339))
	fmt.Println("Sunset:", rs.Set.Format(time.RFC3339))
	// Intentionally no // Output: block so this stays a documentation example
	// and is not validated as a test.
}

// ExampleRiseSetFor demonstrates using the generic RiseSetFor API.
func ExampleRiseSetFor() {
	loc := astroglide.Coordinates{
		Lat: 33.4484,   // Phoenix, AZ
		Lon: -112.0740, // Phoenix longitude
	}

	locPHX, _ := time.LoadLocation("America/Phoenix")
	date := time.Date(2025, time.November, 30, 0, 0, 0, 0, locPHX)

	rs, err := astroglide.RiseSetFor(astroglide.Sun, loc, date)
	if err != nil {
		panic(err)
	}

	fmt.Println("Sunrise:", rs.Rise.Format(time.RFC3339))
	fmt.Println("Sunset:", rs.Set.Format(time.RFC3339))
	// Again, no // Output: so future algorithm changes don't break tests.
}

// ExampleDaylightHours demonstrates calculating daylight duration.
func ExampleDaylightHours() {
	loc := astroglide.Coordinates{
		Lat: 33.4484,   // Phoenix, AZ
		Lon: -112.0740, // Phoenix longitude
	}

	locPHX, _ := time.LoadLocation("America/Phoenix")

	// Summer solstice
	summer := time.Date(2025, time.June, 21, 0, 0, 0, 0, locPHX)
	summerHours, _ := astroglide.DaylightHours(loc, summer)
	fmt.Printf("Summer solstice daylight: %.2f hours\n", summerHours)

	// Winter solstice
	winter := time.Date(2025, time.December, 21, 0, 0, 0, 0, locPHX)
	winterHours, _ := astroglide.DaylightHours(loc, winter)
	fmt.Printf("Winter solstice daylight: %.2f hours\n", winterHours)
}
