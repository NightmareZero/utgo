package util

import (
	"strings"
	"time"
)

func FormatDateTime(s string) (time.Time, error) {
	if strings.Contains(s, "T") {
		return time.Parse(time.RFC3339, s)
	} else if strings.Contains(s, "-") && strings.Contains(s, ":") {
		return time.Parse(time.DateTime, s)
	} else if strings.Contains(s, "-") {
		return time.Parse(time.DateOnly, s)
	} else {
		return time.Parse(time.TimeOnly, s)
	}
}
