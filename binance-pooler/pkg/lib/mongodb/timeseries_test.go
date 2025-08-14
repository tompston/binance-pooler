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

//	TODO: fix these tests
// func TestGapFinder(t *testing.T) {
// 	db, err := SetupMongdbTest()
// 	if err != nil {
// 		t.Fatalf("could not setup mongodb: %v", err)
// 	}
// 	defer db.Conn().Disconnect(context.Background())

// 	t.Run(" Test Timeseries gap checker function", func(t *testing.T) {
// 		name := "timeseries_gap_test"
// 		coll := db.TestCollection(name)

// 		// drop the collection to remove any previous data
// 		if err := coll.Drop(context.Background()); err != nil {
// 			t.Fatalf("failed to drop collection %v: %v", name, err)
// 		}

// 		exampleDay := func(hour int) time.Time {
// 			return time.Date(2023, 10, 1, hour, 0, 0, 0, time.UTC)
// 		}

// 		newRow := func(hour, intervalInHours int) TimeseriesFields {
// 			fields, _ := NewTimeseriesFields(exampleDay(hour), exampleDay(hour+intervalInHours))
// 			return fields
// 		}

// 		records := []interface{}{
// 			// 1-hour intervals (gap between 1 and 2)
// 			newRow(0, 1),
// 			newRow(2, 1),
// 			newRow(3, 1),

// 			// 2-hour intervals (gap between 6 and 10)
// 			newRow(2, 2),
// 			newRow(4, 2),
// 			newRow(10, 2),
// 			newRow(12, 2),
// 		}
// 		_, err := coll.InsertMany(context.Background(), records)
// 		if err != nil {
// 			t.Fatalf("failed to insert records: %v", err)
// 		}

// 		gaps, err := FindGaps(coll)
// 		if err != nil {
// 			t.Fatalf("failed to find gaps: %v", err)
// 		}

// 		// Validate the gaps
// 		expectedGaps := map[int64][]GapInfo{
// 			3600000: {
// 				{StartOfGap: exampleDay(1), EndOfGap: exampleDay(2)},
// 			},
// 			7200000: {
// 				{StartOfGap: exampleDay(6), EndOfGap: exampleDay(10)},
// 			},
// 		}

// 		// assert.Equal(t, expectedGaps, gaps)

// 		if len(expectedGaps) != len(gaps) {
// 			t.Fatalf("expected %d gaps but got %d", len(expectedGaps), len(gaps))
// 		}
// 	})
// }
