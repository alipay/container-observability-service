package utils

import "time"

const (
	layout = "2006-01-02T15:04:05.000Z"
)

// ParseTime parse string into time using layout "2006-01-02T15:04:05.000Z"
func ParseTime(s string) (time.Time, error) {
	return time.Parse(layout, s)
}

// FormatTime format time to UTC time and with million seconds
func FormatTime(t time.Time) string {
	return t.Format(layout)
}
