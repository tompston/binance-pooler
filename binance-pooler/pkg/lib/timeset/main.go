package timeset

import (
	"fmt"
	"time"
)

func MilisToDuration(milis int64) time.Duration           { return time.Duration(milis) * time.Millisecond }
func ExceedsDiffInHours(t1, t2 time.Time, hours int) bool { return t2.Sub(t1).Hours() > float64(hours) }

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

	debug := len(withDebug) == 1 && withDebug[0]

	if debug {
		fmt.Printf("maxDuration: %v, overlayTime: %v\n", maxDuration.Hours()/24, overlayTime.Hours()/24)
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

		if debug {
			fmt.Printf("%v -> %v, diff: %v hours\n", start.Format(format), end.Format(format), end.Sub(start).Hours())
		}
	}

	return chunks, nil
}
