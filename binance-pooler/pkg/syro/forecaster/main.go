package forecaster

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Forecast struct {
	StartTime  time.Time      `json:"start_time" bson:"start_time"` // For which start_time the forecast is made
	Interval   int64          `json:"interval" bson:"interval"`     // Resolution of the forecast (milliseconds)
	Offset     int64          `json:"offset" bson:"offset"`         // start_time - now
	UserOffset *int64         `json:"u_offset" bson:"u_offset"`     // Optional specified offset by the user
	Source     string         `json:"source" bson:"source"`         // Source of the forecast (who submitted the forecast)
	Model      string         `json:"model" bson:"model"`           // Model name which generated the forecast
	Variable   string         `json:"var" bson:"var"`               // Variable for which the forecast is made
	SubVar     *string        `json:"sub_var" bson:"sub_var"`       // Optional sub-variable (eg wind_forecast -> wind_speed)
	Value      float64        `json:"value" bson:"value"`           // Forecasted value
	Meta       map[string]any `json:"meta" bson:"meta"`             // Additional metadata
}

type NewForecastsBody struct {
	// Fields which are constant for all forecasts
	Interval int64          `json:"interval"`
	Source   string         `json:"source"`
	Model    string         `json:"model"`
	Variable string         `json:"variable"`
	SubVar   string         `json:"sub_var"`
	Meta     map[string]any `json:"meta"`
	Data     []struct {
		// Json field names are shortened to reduce the size of the request
		StartTime  time.Time `json:"t"` // todo: refactor to a timestamp for parsing efficiency?
		Value      float64   `json:"v"`
		UserOffset *int64    `json:"uo"` // save the value if it's specified
	} `json:"data"`
}

// Optional settings for saving forecasts
type SaveOptions struct {
	ExcludePast     bool  // Disallow saving forecasts with start_time in the past
	MaxFutureOffset int64 // Maximum allowed future offset (milliseconds)
	MaxPastOffset   int64 // Maximum allowed past offset (milliseconds)
}

// Parse the ForecastBody and return a slice of Forecast structs
func ParseBody(body NewForecastsBody, opt ...SaveOptions) ([]Forecast, error) {

	if len(body.Data) == 0 {
		return nil, fmt.Errorf("no forecasts to save")
	}

	if body.Source == "" {
		return nil, fmt.Errorf("source is required")
	}

	if body.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if body.Variable == "" {
		return nil, fmt.Errorf("variable is required")
	}

	if body.Interval <= 0 {
		return nil, fmt.Errorf("interval is required and must be greater than 0")
	}

	// var forecasts []Forecast
	forecasts := make([]Forecast, 0, len(body.Data))
	optionsExist := len(opt) == 1
	now := time.Now()

	for _, f := range body.Data {
		offset := now.Sub(f.StartTime).Milliseconds()

		// Do the filtering of forecasts before creating new structs, for optimization
		if optionsExist {
			options := opt[0]
			if options.ExcludePast && offset < 0 {
				continue
			}

			if options.MaxFutureOffset > 0 && offset > options.MaxFutureOffset {
				continue
			}

			if options.MaxPastOffset > 0 && offset < -options.MaxPastOffset {
				continue
			}
		}

		if f.StartTime.IsZero() {
			return nil, fmt.Errorf("start_time is required")
		}

		forecasts = append(forecasts, Forecast{
			StartTime:  f.StartTime.UTC(),
			Interval:   body.Interval,
			Offset:     offset,
			UserOffset: f.UserOffset,
			Source:     body.Source,
			Model:      body.Model,
			Variable:   body.Variable,
			SubVar:     &body.SubVar,
			Value:      f.Value,
			Meta:       body.Meta,
		})
	}

	return forecasts, nil
}

// --- mongo storage ---

type MongoStorage struct {
	coll *mongo.Collection
}

func NewMongoStorage(coll *mongo.Collection) *MongoStorage {
	return &MongoStorage{coll: coll}
}

func (m *MongoStorage) CreateIndexes(f Forecast) error {
	return nil
}

func (m *MongoStorage) InsertForecasts(f Forecast) error {
	_, err := m.coll.InsertOne(context.Background(), f)
	return err
}

func (m *MongoStorage) QueryForecasts() error {
	return nil
}
