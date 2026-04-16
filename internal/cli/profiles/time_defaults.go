package profiles

import (
	"fmt"
	"strings"
	"time"
)

func validateTimeDefaults(timezone string, timeOfDay string) error {
	if timezone != "" {
		if _, err := time.LoadLocation(strings.TrimSpace(timezone)); err != nil {
			return fmt.Errorf("invalid default timezone %q: %w", timezone, err)
		}
	}

	if timeOfDay != "" {
		value := strings.TrimSpace(timeOfDay)
		for _, layout := range []string{"15:04", "15:04:05"} {
			if _, err := time.Parse(layout, value); err == nil {
				return nil
			}
		}
		return fmt.Errorf("invalid default time-of-day %q: use HH:MM or HH:MM:SS", timeOfDay)
	}

	return nil
}
