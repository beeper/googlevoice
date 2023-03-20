package util

import (
	"strconv"
	"time"
)

// StringMilliTimestampToTime returns a time.Time object of the given
// millisecond timestamp string.
func StringMilliTimestampToTime(s string) time.Time {
	t, _ := strconv.ParseInt(s, 10, 64)
	return time.UnixMilli(t)
}
