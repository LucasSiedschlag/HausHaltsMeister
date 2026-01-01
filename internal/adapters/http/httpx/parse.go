package httpx

import (
	"strings"
	"time"
)

func ParseDateTime(value string) (time.Time, error) {
	if strings.Contains(value, "T") {
		return time.Parse(time.RFC3339, value)
	}
	return time.Parse("2006-01-02", value)
}

func ParseMonth(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}
