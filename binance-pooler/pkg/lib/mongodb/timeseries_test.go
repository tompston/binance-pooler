package mongodb

import (
	"testing"
	"time"
)

func TestTimeseries(t *testing.T) {
	startTime := time.Date(2023, 9, 20, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(1 * time.Hour)
	fields, err := NewTimeseriesFields(startTime, endTime)
	if err != nil {
		t.Fatalf("Failed to create TimeseriesFields: %v", err)
	}

	t.Run(" Test Timeseries field creation", func(t *testing.T) {
		// Check if StartTime and EndTime are in UTC
		if fields.StartTime.Location() != time.UTC {
			t.Error("StartTime is not in UTC")
		}

		// Check if Interval is correctly calculated
		expectedInterval := 3600000
		if fields.Interval != int64(expectedInterval) {
			t.Errorf("Expected interval to be %d but got %d", expectedInterval, fields.Interval)
		}
	})

}
