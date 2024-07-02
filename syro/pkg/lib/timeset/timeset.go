package timeset

import (
	"fmt"
	"time"
)

// StartOfDay converts the input time to the start of the day.
func StartOfDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func StartOfNextDay(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, 1)
}

func IsTodayOrInFuture(t time.Time) bool {
	return t.UTC().After(StartOfDay(time.Now().UTC()))
}

// DiffInMilliseconds calculates the time difference between the start and end time in milliseconds
func DiffInMilliseconds(t1, t2 time.Time) int64 { return t2.Sub(t1).Milliseconds() }

// ExceedsDiffInHours checks if the time difference between the start
// and end time exceeds the allowed difference in hours.
func ExceedsDiffInHours(t1, t2 time.Time, hours int) bool { return t2.Sub(t1).Hours() > float64(hours) }

func IsFromYesterday(t time.Time) bool {
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	return t.Format("2006-01-02") == yesterday.Format("2006-01-02")
}

// EndOfMonth returns the last day of the month for the given date.
func EndOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).AddDate(0, 1, -1)
}

// StartOfMonth returns the first day of the month for the given date.
func StartOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func EndOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 0, 0, date.UTC().Location())
}

// MinSince returns a string that represents the time passed since the input time
func MinSince(t time.Time) string { return fmt.Sprintf("%.2f min", time.Since(t).Seconds()/60) }

// SecSince returns a string that represents the time passed since the input time
func SecSince(t time.Time) string { return fmt.Sprintf("%.2f sec", time.Since(t).Seconds()) }

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

func IsInNextMonth(t time.Time) bool {
	now := time.Now().UTC()
	nextMonth := now.AddDate(0, 1, 0)
	return t.Year() == nextMonth.Year() && t.Month() == nextMonth.Month()
}

// Calculates the offset from the current time to a future time in milliseconds.
// If the given time is not in the future, the result will be negative.
func FutureOffsetMillis(future time.Time) int64 {
	now := time.Now().UTC()
	duration := future.Sub(now)
	return int64(duration / time.Millisecond)
}

// Round the milliseconds to the closest hour.
func RoundMillisToClosestHour(milliseconds int64) int64 {
	hourInMilli := int64(3600000) // Number of milliseconds in an hour
	// Calculate the half-hour mark to determine rounding direction
	halfHourInMilli := hourInMilli / 2

	// Adding half an hour worth of milliseconds before division to ensure rounding to nearest hour
	if milliseconds >= 0 {
		return (milliseconds + halfHourInMilli) / hourInMilli * hourInMilli
	}
	return (milliseconds - halfHourInMilli) / hourInMilli * hourInMilli
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

func MinToMillis(min int64) int64  { return min * 60 * 1000 }
func MilisToMin(milis int64) int64 { return milis / 60000 }

// Convert Unix milliseconds to time.Time value
func UnixMillisToTime(unixMillis int64) time.Time {
	seconds := unixMillis / 1000
	nanoseconds := (unixMillis % 1000) * int64(time.Millisecond)
	return time.Unix(seconds, nanoseconds)
}
