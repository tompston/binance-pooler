package timeset

import (
	"fmt"
	"time"
)

func StartOfNextDay(t time.Time) time.Time                { return StartOfDay(t).AddDate(0, 0, 1) }
func IsTodayOrInFuture(t time.Time) bool                  { return t.UTC().After(StartOfDay(time.Now().UTC())) }
func MilisToDuration(milis int64) time.Duration           { return time.Duration(milis) * time.Millisecond }
func DiffInMilliseconds(t1, t2 time.Time) int64           { return t2.Sub(t1).Milliseconds() }
func ExceedsDiffInHours(t1, t2 time.Time, hours int) bool { return t2.Sub(t1).Hours() > float64(hours) }
func MinSince(t time.Time) string                         { return fmt.Sprintf("%.2f min", time.Since(t).Seconds()/60) }
func SecSince(t time.Time) string                         { return fmt.Sprintf("%.2f sec", time.Since(t).Seconds()) }

// StartOfDay converts the input time to the start of the day.
func StartOfDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// UtcOffsetInLocation returns the UTC offset in hours for the given location.
// Possible inputs:
//   - "Europe/Riga"
//   - "America/Mexico_City"
func UtcOffsetInLocation(location string) (int, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		return 0, err
	}

	_, offset := time.Now().In(loc).Zone()
	return offset / 3600, nil
}

// Calculates the offset from the current time to a future time in milliseconds.
// If the given time is not in the future, the result will be negative.
func FutureOffsetMillis(future time.Time) int64 {
	now := time.Now().UTC()
	duration := future.Sub(now)
	return int64(duration / time.Millisecond)
}

// See documentation for ExpandTime
func ExpandDays(start time.Time, days int, filterFunc func(time.Time) bool) []time.Time {
	return ExpandTime(start, 24*time.Hour, days, filterFunc)
}

// ExpandTime takes a starting date, a duration to add, and the number of time values into the future.
// The optional filterFunc can be used to skip certain dates based on custom logic.
// If no filterFunc is provided, all dates are included (including the input date).
func ExpandTime(start time.Time, duration time.Duration, count int, filterFunc func(time.Time) bool) []time.Time {
	var times []time.Time
	for i := 0; i < count; i++ {
		nextTime := start.Add(duration * time.Duration(i))
		if filterFunc != nil && filterFunc(nextTime) {
			continue // Skip this date if filterFunc returns true
		}
		times = append(times, nextTime)
	}
	return times
}

// Convert Unix milliseconds to time.Time value
func UnixMillisToTime(unixMillis int64) time.Time {
	seconds := unixMillis / 1000
	nanoseconds := (unixMillis % 1000) * int64(time.Millisecond)
	return time.Unix(seconds, nanoseconds)
}

type TimeChunk struct {
	From time.Time
	To   time.Time
}

func ChunkTimeRange(from, to time.Time, interval time.Duration, maxReqPeriods, overlayPeriods int, withDebug ...bool) ([]TimeChunk, error) {
	if from.After(to) {
		return nil, fmt.Errorf("from date is after to date")
	}

	if interval <= 0 {
		return nil, fmt.Errorf("interval must be greater than 0")
	}

	var chunks []TimeChunk
	maxDuration := interval * time.Duration(maxReqPeriods) // Maximum duration of each chunk
	overlayTime := interval * time.Duration(overlayPeriods)

	debugEnabled := len(withDebug) == 1 && withDebug[0]

	if debugEnabled {
		fmt.Printf("maxDuration: %v, overlayTime: %v\n",
			maxDuration.Hours()/24, overlayTime.Hours()/24)
	}

	const format = "2006-01-02 15:04:05"

	for start := from; start.Before(to); {
		end := start.Add(maxDuration)
		if end.After(to) {
			end = to
		}

		chunks = append(chunks, TimeChunk{From: start, To: end})

		start = start.Add(maxDuration).Add(-overlayTime)
		end = end.Add(maxDuration).Add(-overlayTime)

		if debugEnabled {
			fmt.Printf("%v -> %v, diff: %v hours\n", start.Format(format), end.Format(format), end.Sub(start).Hours())
		}
	}

	return chunks, nil
}
