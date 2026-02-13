// Package timeutil provides time formatting utilities
package timeutil

import (
	"fmt"
	"time"
)

// FormatAge returns a human-readable string representing how long ago a time was
func FormatAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)

	switch {
	case d < time.Hour*24:
		return "today"
	case d < time.Hour*24*2:
		return "yesterday"
	case d < time.Hour*24*7:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < time.Hour*24*30:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case d < time.Hour*24*365:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(d.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
