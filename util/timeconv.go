package util

import (
	"strings"
	"time"
)

var (
	DateTime = "2006-01-02 15:04:05"
	DateOnly = "2006-01-02"
	TimeOnly = "15:04:05"
)

func FormatDateTime(s string) (time.Time, error) {
	if strings.Contains(s, "T") {
		return time.Parse(time.RFC3339, s)
	} else if strings.Contains(s, "-") && strings.Contains(s, ":") {
		return time.Parse(DateTime, s)
	} else if strings.Contains(s, "-") {
		return time.Parse(DateOnly, s)
	} else {
		return time.Parse(TimeOnly, s)
	}
}
