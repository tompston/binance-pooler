package market_dto

import (
	"context"
	"fmt"
	"syro/pkg/lib/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// OHLC represents the open, high, low, close, and volume of a market
type OHLC struct {
	Open   float64 `json:"o" bson:"o"`
	High   float64 `json:"h" bson:"h"`
	Low    float64 `json:"l" bson:"l"`
	Close  float64 `json:"c" bson:"c"`
	Volume float64 `json:"v" bson:"v"`
}

func NewOHLC(open, high, low, close, volume float64) *OHLC {
	return &OHLC{Open: open, High: high, Low: low, Close: close, Volume: volume}
}

// timeseries data stored in the db
type OhlcRow struct {
	mongodb.TimeseriesFields `bson:",inline"`
	ID                       string `json:"id" bson:"id"`
	OHLC                     `bson:",inline"`
	// Optional fields that are not always available
	BaseAssetVolume *float64 `json:"bv" bson:"bv"`
	NumberOfTrades  *int64   `json:"n" bson:"n"`
}

// Pretty print the OhlcRow for debugging
func (o *OhlcRow) String() string {
	if o == nil {
		return "<nil>"
	}

	return fmt.Sprintf("id: %v, time: %v, interval: %v, o: %v, h: %v, l: %v, c: %v, v: %v",
		o.ID, o.StartTime, o.Interval, o.Open, o.High, o.Low, o.Close, o.Volume)
}

func NewOhlcRow(id string, startTime, endTime time.Time, open, high, low, close, volume float64) (*OhlcRow, error) {
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}

	timeseries, err := mongodb.NewTimeseriesFields(startTime, endTime)
	if err != nil {
		return nil, err
	}

	return &OhlcRow{
		TimeseriesFields: timeseries,
		ID:               id,
		OHLC:             *NewOHLC(open, high, low, close, volume),
	}, nil
}

func (r *OhlcRow) SetBaseAssetVolume(vol float64) { r.BaseAssetVolume = &vol }
func (r *OhlcRow) SetNumberOfTrades(num int64)    { r.NumberOfTrades = &num }

func (m *Mongo) CreateOhlcIndexes(coll *mongo.Collection) error {
	return mongodb.TimeseriesIndexes().
		Add("id").
		Add(mongodb.START_TIME, "id").
		Create(coll)
}

func (m *Mongo) GetLatestOhlcStartTime(id string, defaultStartTime time.Time, coll *mongo.Collection, loggerFn func(any)) (time.Time, error) {
	return mongodb.GetLatestStartTime(defaultStartTime, coll, bson.M{"id": id}, false, loggerFn)
}

func (m *Mongo) UpsertOhlcRows(data []OhlcRow, coll *mongo.Collection) (*mongodb.UpsertLog, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to upsert")
	}

	upsertFn := func(row OhlcRow) error {
		if row.ID == "" {
			return fmt.Errorf("id is empty")
		}

		filter := bson.M{"id": row.ID, mongodb.START_TIME: row.StartTime, "interval": row.Interval}
		_, err := coll.UpdateOne(context.Background(), filter, bson.M{"$set": row}, mongodb.UpsertOpt)
		return err
	}

	start := time.Now()

	for _, row := range data {
		if err := upsertFn(row); err != nil {
			return nil, err
		}
	}

	return mongodb.NewUpsertLog(coll,
		data[0].StartTime, data[len(data)-1].StartTime,
		len(data), start), nil
}
