# Astroglide

> *"Because calculating celestial events should be smooth and friction-free."* 🌙✨

A Go library for computing astronomical events including sunrise, sunset, moonrise, moonset, twilight periods, and moon phases. Designed for use with [MyWeatherDash](https://github.com/thurmanmarka/MyWeatherDash) and other applications requiring accurate celestial calculations.

**Warning**: This package is for astronomical calculations only. For other types of gliding, please consult your local pharmacy.

## Features

*Slip into astronomical accuracy with ease:*

- **Sun Events** ☀️
  - Sunrise and sunset times (because mornings are hard enough)
  - Civil, nautical, and astronomical twilight (for every shade of dark)
  - Golden hour and blue hour calculations (Instagram won't filter itself)
  
- **Moon Events** 🌙
  - Moonrise and moonset times (lunacy has a schedule)
  - Moon phase calculations (illumination, phase name, waxing/waning)
  - Distance-adjusted horizon corrections (the Moon plays hard to get)
  
- **High Accuracy** 🎯
  - Topocentric corrections for the Moon (because perspective matters)
  - Geographic location support (latitude, longitude, elevation)
  - Time zone aware calculations (unlike your jet-lagged relatives)

## Installation

```bash
go get github.com/thurmanmarka/astroglide
```

## Quick Start

### Sunrise and Sunset

*Glide smoothly into solar calculations:*

```go
package main

import (
    "fmt"
    "time"
    "github.com/thurmanmarka/astroglide"
)

func main() {
    // Phoenix, Arizona coordinates (where the sun has no chill)
    location := astroglide.Coordinates{
        Lat: 33.4484,
        Lon: -112.0740,
    }
    
    // Get today's sunrise and sunset
    date := time.Now()
    rs, err := astroglide.SlideIntoSunset(location, date)
    if err != nil {
        panic(err) // The sun will rise again, your code might not
    }
    
    fmt.Printf("Sunrise: %s\n", rs.Rise.Format(time.Kitchen))
    fmt.Printf("Sunset:  %s\n", rs.Set.Format(time.Kitchen))
}
```

### Moon Phase

*Get illuminated about lunar illumination:*

```go
phase, err := astroglide.MoonPhaseAt(time.Now())
if err != nil {
    panic(err) // Houston, we have a problem
}

fmt.Printf("Moon Phase: %s\n", phase.Name)
fmt.Printf("Illumination: %.1f%%\n", phase.Fraction * 100)
// Now you know whether to blame the full moon for your weird week
```

### Moonrise and Moonset

*Because werewolves need to plan their schedules too:*

```go
location := astroglide.Coordinates{
    Lat: 33.4484,
    Lon: -112.0740,
}

date := time.Now()
rs, err := astroglide.RiseSetFor(astroglide.Moon, location, date)
if err != nil {
    panic(err) // No moonrise means no excuses for that behavior
}

fmt.Printf("Moonrise: %s\n", rs.Rise.Format(time.Kitchen))
fmt.Printf("Moonset:  %s\n", rs.Set.Format(time.Kitchen))
```

## API Reference

### Types

#### `Coordinates`
Represents an observer's location on Earth.

```go
type Coordinates struct {
    Lat       float64 // degrees, north positive
    Lon       float64 // degrees, east positive (west negative)
    Elevation float64 // meters above sea level (reserved for future use)
}
```

#### `Body`
Celestial bodies available for calculations.

```go
const (
    Sun Body = iota
    Moon
)
```

#### `RiseSet`
Holds rise and set times for a celestial body.

```go
type RiseSet struct {
    Rise time.Time
    Set  time.Time
}
```

#### `MoonPhase`
Describes the Moon's illumination and phase.

```go
type MoonPhase struct {
    Time       time.Time // the instant this phase is evaluated at
    Fraction   float64   // illuminated fraction [0..1]
    Elongation float64   // Sun-Moon angular separation in degrees [0..180]
    Waxing     bool      // true if waxing, false if waning
    Name       string    // e.g. "New Moon", "Waxing Crescent", "Full Moon"
}
```

#### `TwilightKind`
Types of twilight based on Sun altitude below the horizon.

```go
const (
    TwilightCivil         // Sun at -6°
    TwilightNautical      // Sun at -12°
    TwilightAstronomical  // Sun at -18°
)
```

### Functions

#### `RiseSetFor(body Body, loc Coordinates, date time.Time) (RiseSet, error)`
Computes rise and set times for a celestial body at a given location and date.

#### `SlideIntoSunset(loc Coordinates, date time.Time) (RiseSet, error)`
Convenience function for computing sunrise and sunset. *The name is the best part of this function.*

#### `TwilightFor(loc Coordinates, date time.Time, kind TwilightKind) (RiseSet, error)`
Computes twilight times (dawn and dusk) for a given twilight type.

#### `GoldenHourFor(loc Coordinates, date time.Time) (DaylightPhases, error)`
Computes golden hour intervals (Sun altitude between -4° and +6°).

#### `BlueHourFor(loc Coordinates, date time.Time) (DaylightPhases, error)`
Computes blue hour intervals (Sun altitude between -6° and -4°).

#### `MoonPhaseAt(t time.Time) (MoonPhase, error)`
Computes the Moon's phase and illumination at a specific time.

## Command Line Tool

Astroglide includes a CLI tool for quick calculations:

### Installation

```bash
go install github.com/thurmanmarka/astroglide/cmd/astroglide@latest
```

### Usage

#### Sun/Moon Rise and Set

```bash
# Default (Sun rise/set for today at 0,0)
astroglide -lat 33.4484 -lon -112.0740

# Specific date
astroglide -lat 33.4484 -lon -112.0740 -date 2025-12-25

# Moon rise/set
astroglide -lat 33.4484 -lon -112.0740 -body moon

# JSON output
astroglide -lat 33.4484 -lon -112.0740 -json
```

#### Moon Phase

```bash
# Current moon phase in UTC
astroglide phase

# Specific time and timezone
astroglide phase -tz America/Phoenix -time "2025-12-25T18:00"
```

## Implementation Details

### Accuracy

*We're astronomically precise (but we're not launching rockets here):*

- **Sun calculations**: Approximately ±1 minute accuracy for rise/set times (more punctual than your average meeting)
- **Moon calculations**: Tuned for Phoenix, Arizona (2025) with distance-dependent horizon corrections (because the Moon social-distances too)
- **Topocentric corrections**: Applied for Moon position to account for observer location on Earth's surface (yes, where you stand actually matters)
- **Atmospheric refraction**: Included in horizon calculations (the atmosphere bends light, and the truth)

### Algorithm Levels

The library is designed to evolve from simple approximate algorithms (Level 1) to high-precision ephemeris-grade models (Level 3). Currently implements Level 1 algorithms with plans for future enhancement.

*Think of it as a journey from "eyeballing it" to "NASA would approve."*

### Internal Structure

- `internal/sun`: Solar position and event calculations
- `internal/moon`: Lunar position, phase, and event calculations
- `internal/solver`: Generic altitude event solver (rise/set/twilight)
- `internal/timeutil`: Time and angle conversion utilities

## Examples

See the `cmd/astroglide` and `cmd/astroglide-profiler` directories for complete working examples.

## Testing

The project includes comprehensive test coverage:

```bash
go test ./...
```

Test files include validation against known astronomical data for Phoenix, Arizona in 2025.

## Error Handling

The library returns `ErrNoRiseNoSet` when a celestial body does not rise or set on a given date at a location (e.g., polar regions during certain seasons).

*Sometimes the Sun just doesn't show up. We've all been there.*

## Contributing

This project is part of the MyWeatherDash ecosystem. Contributions, issues, and feature requests are welcome.

## License

See LICENSE file for details.

## Related Projects

- [MyWeatherDash](https://github.com/thurmanmarka/MyWeatherDash) - Weather dashboard using Astroglide for astronomical data

## Credits

Astronomical algorithms based on Jean Meeus' "Astronomical Algorithms" and other standard references in the field.

The name? Well, that's all us. We wanted something that would make astronomers and pharmacists do a double-take. Mission accomplished. 🎯

*Making celestial mechanics smoother since 2025.*
