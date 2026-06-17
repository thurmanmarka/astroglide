package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/thurmanmarka/astroglide"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter date/time (YYYY-MM-DD or YYYY-MM-DDTHH:MM, blank=now, q=quit): ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "q" || input == "quit" {
			break
		}

		var t time.Time
		if input == "" {
			t = time.Now()
		} else {
			var err error
			for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02"} {
				t, err = time.ParseInLocation(layout, input, time.Local)
				if err == nil {
					break
				}
			}
			if err != nil {
				fmt.Printf("  couldn't parse %q — try YYYY-MM-DD or YYYY-MM-DDTHH:MM\n\n", input)
				continue
			}
		}

		phase, err := astroglide.MoonPhaseAt(t)
		if err != nil {
			fmt.Printf("  error: %v\n\n", err)
			continue
		}

		trend := "Waning"
		if phase.Waxing {
			trend = "Waxing"
		}

		fmt.Printf("  %s\n", phase.Name)
		fmt.Printf("  Illuminated : %.1f%%\n", phase.Fraction*100)
		fmt.Printf("  Elongation  : %.2f°\n", phase.Elongation)
		fmt.Printf("  Trend       : %s\n\n", trend)
	}
}
