package utils

import (
	"fmt"
	"time"
)

func FuzzyTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	switch {
	case t.IsZero():
		return "Never"
	case duration < -0:
		return "In the future"
	case duration < time.Minute:
		return "Just now"
	case duration < 2*time.Minute:
		return "A minute ago"
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 2*time.Hour:
		return "An hour ago"
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration < 48*time.Hour:
		return "Yesterday"
	case duration < 48*time.Hour:
		return "Yesterday"
	default:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}
}
