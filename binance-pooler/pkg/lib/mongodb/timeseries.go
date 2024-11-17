// The vast majority of the database is populated with timeseries data. This
// package defines a reusable type which can be used to create new
// timeseries data.
package mongodb

import (
	"binance-pooler/pkg/syro/timeset"
	"binance-pooler/pkg/syro/utils"
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TimeseriesFields holds filelds which are used for timeseries data.
type TimeseriesFields struct {
	StartTime time.Time `json:"start_time" bson:"start_time"` // StartTime is the start of the measurement
	Interval  int64     `json:"interval" bson:"interval"`     // Interval is the frequency of the measurement in milliseconds
}

// Name of the json + bson fields
const (
	START_TIME = "start_time"
	INTERVAL   = "interval"
)

// NewTimeseriesFields returns a new TimeseriesFields struct for which the time values
// are always UTC (so that all of the collections have standardized time
// values and the interval value which is calculated from the passed
// in start and end time.
func NewTimeseriesFields(startTime, endTime time.Time) (TimeseriesFields, error) {
	if startTime.IsZero() || endTime.IsZero() {
		return TimeseriesFields{}, fmt.Errorf("start or end time is zero")
	}

	if startTime.After(endTime) {
		return TimeseriesFields{}, fmt.Errorf("start time is after the end time")
	}

	return TimeseriesFields{
		StartTime: startTime.UTC(),
		Interval:  endTime.Sub(startTime).Milliseconds(),
	}, nil
}

// Indexes returns a new IndexBuilder for the timeseries collection.
func TimeseriesIndexes() *IndexBuilder {
	return NewIndexes().Add(START_TIME).Add("interval")
}

// CreateTimeseriesIndexes creates indexes for interval and time fields
func CreateTimeseriesIndexes(coll *mongo.Collection) error {
	return TimeseriesIndexes().Create(coll)
}

type GapInfo struct {
	StartOfGap time.Time
	EndOfGap   time.Time
}

func (g GapInfo) String() string {
	return g.StartOfGap.Format("2006-01-02 15:04:05") + " - " + g.EndOfGap.Format("2006-01-02 15:04:05")
}

func findGapsInIntervalGroup(records []TimeseriesFields) []GapInfo {
	var gaps []GapInfo
	for i := 0; i < len(records)-1; i++ {
		currStartTime := records[i].StartTime
		nextStartTime := records[i+1].StartTime

		expectedNextStartTime := currStartTime.Add(time.Duration(records[i].Interval) * time.Millisecond)

		if expectedNextStartTime.Before(nextStartTime) {
			gaps = append(gaps, GapInfo{
				StartOfGap: expectedNextStartTime,
				EndOfGap:   nextStartTime,
			})
		}
	}

	return gaps
}

func UnixMilisToTime(unixMilli int64) time.Time {
	// Convert milliseconds to seconds and nanoseconds
	seconds := unixMilli / 1000
	nanoseconds := (unixMilli % 1000) * 1000000
	return time.Unix(seconds, nanoseconds)
}

// FindGaps in the given collection. This is done based on the
// time and interval field. Gaps are cheched for each unique
// interval seperately.
func FindGaps(coll *mongo.Collection, customFilter ...bson.M) (map[int64][]GapInfo, error) {
	ctx := context.Background()

	var filter bson.M
	if len(customFilter) == 1 {
		filter = customFilter[0]
	}

	// Find unique intervals in the collection
	intervalCursor, err := coll.Distinct(ctx, "interval", filter)
	if err != nil {
		return nil, err
	}

	var intervals []int64
	for _, item := range intervalCursor {
		if intVal, ok := item.(int64); ok {
			intervals = append(intervals, intVal)
		}
	}

	gapsMap := make(map[int64][]GapInfo)

	for _, interval := range intervals {
		filter := bson.M{"interval": interval}

		// merge the passed down filter with the interval filter if it exists
		if len(customFilter) == 1 {
			additionalFilter := customFilter[0]
			for k, v := range additionalFilter {
				filter[k] = v
			}
		}

		cursor, err := coll.Find(ctx, filter, options.Find().SetSort(bson.M{START_TIME: 1}))
		if err != nil {
			return nil, err
		}

		var rec []TimeseriesFields
		for cursor.Next(ctx) {
			var record TimeseriesFields
			if err = cursor.Decode(&record); err != nil {
				return nil, err
			}
			rec = append(rec, record)
		}

		if err = cursor.Err(); err != nil {
			return nil, err
		}
		cursor.Close(ctx)

		gaps := findGapsInIntervalGroup(rec)
		if len(gaps) > 0 {
			gapsMap[interval] = gaps
		}
	}

	return gapsMap, nil
}

// GetLatestStartTime returns the most recent time value from a document from a
// collection which is sorted by time field (Descending).
//
// If the returned time.Time value is today and setToStartOfToday is set to true, then
// it is set to the start of today, so that the service would refetch the data from
// the start of today. This is done because most of the scraped data sources
// update the data of today constantly.
func GetLatestStartTime(defaultStart time.Time, coll *mongo.Collection, filter bson.M, setToStartOfToday bool, logger_fn ...func(any)) (time.Time, error) {

	sort := bson.M{START_TIME: -1}

	var row TimeseriesFields
	err := GetLastDocumentWithTypes(coll, sort, filter, &row)
	if err != nil {
		// if the query fails because there are no documents in the result, return the
		// default start date and no errors.
		if strings.Contains(err.Error(), "no documents in result") {
			msg := noRecordsFoundMsg(coll.Name(), defaultStart, filter)
			utils.LogIfArgExists(msg, logger_fn)
			return defaultStart, nil
		}

		// If could not query the db for other reason, return error
		err := fmt.Errorf("could not get the last document from the %v collection: %v", coll.Name(), err)
		return time.Time{}, err
	}

	startTime := row.StartTime
	if startTime.IsZero() {
		return time.Time{}, fmt.Errorf("mongo query returned a zero value for the time field")
	}

	// if the latest row in the collection has the time value of today or is in the future, set it to the
	// start of today. This is done so that the info queried from the external sources is updated, as the
	// majority of sources don't have all of the data for today available for most of the day.
	if setToStartOfToday && timeset.IsTodayOrInFuture(startTime) {
		msg := lastDocStartTimeIsTodayMsg(startTime, coll.Name())
		utils.LogIfArgExists(msg, logger_fn)
		startOfToday := timeset.StartOfDay(time.Now())
		return startOfToday, err
	}

	// if the returned time.Time value is a valid date and is not today, return the last
	// time value from the collection
	return startTime, err
}

func noRecordsFoundMsg(collName string, defaultStartDate time.Time, filter bson.M) string {
	return fmt.Sprintf("no records found in the %v collection with %v filter. Returning default start date of %v",
		collName, filter, defaultStartDate.Format("2006-01-02"))
}

func lastDocStartTimeIsTodayMsg(collectionLatestDate time.Time, collName string) string {
	return fmt.Sprintf("latest document in the %v collection is today or in the future ( %v ) . Returning start of today",
		collName, collectionLatestDate.Format("2006-01-01 05:00"))
}
