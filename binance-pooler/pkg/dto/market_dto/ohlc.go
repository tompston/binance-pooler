package market_dto

import (
	"binance-pooler/pkg/lib/mongodb"
	"fmt"
	"sort"
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
	Symbol                   string `json:"symbol" bson:"symbol"`
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

	return fmt.Sprintf("symbol: %v, time: %v, interval: %v, o: %v, h: %v, l: %v, c: %v, v: %v",
		o.Symbol, o.StartTime, o.Interval, o.Open, o.High, o.Low, o.Close, o.Volume)
}

func NewOhlcRow(symbol string, startTime, endTime time.Time, open, high, low, close, volume float64) (*OhlcRow, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is empty")
	}

	timeseries, err := mongodb.NewTimeseriesFields(startTime, endTime)
	if err != nil {
		return nil, err
	}

	return &OhlcRow{
		TimeseriesFields: timeseries,
		Symbol:           symbol,
		OHLC:             *NewOHLC(open, high, low, close, volume),
	}, nil
}

func (r *OhlcRow) SetBaseAssetVolume(vol float64) { r.BaseAssetVolume = &vol }
func (r *OhlcRow) SetNumberOfTrades(num int64)    { r.NumberOfTrades = &num }

func (*Mongo) CreateOhlcIndexes(coll *mongo.Collection) error {
	return mongodb.TimeseriesIndexes().
		Add("symbol").
		Add(mongodb.START_TIME, "symbol", "interval").
		Create(coll)
}

func (*Mongo) UpsertOhlcRows(data []OhlcRow, coll *mongo.Collection) (*mongodb.UpsertLog, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to upsert")
	}

	// sort the data for consecutive upserts
	sort.Slice(data, func(i, j int) bool {
		return data[i].StartTime.Before(data[j].StartTime)
	})

	var models []mongo.WriteModel
	for _, row := range data {
		if row.Symbol == "" {
			return nil, fmt.Errorf("symbol is empty")
		}

		filter := bson.M{"symbol": row.Symbol, mongodb.START_TIME: row.StartTime, "interval": row.Interval}
		update := bson.M{"$set": row}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	start := time.Now()

	_, err := coll.BulkWrite(ctx, models)
	log := mongodb.NewUpsertLog(coll, data[0].StartTime, data[len(data)-1].StartTime, len(data), start)
	return log, err
}
