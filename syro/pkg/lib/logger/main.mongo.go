package logger

import (
	"context"
	"errors"
	"fmt"
	"syro/pkg/lib/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Logger implementation for mongodb
type MongoLogger struct {
	Coll     *mongo.Collection
	Settings *LoggerSettings
	Source   string
	Event    string
	EventID  string
}

func NewMongoLogger(coll *mongo.Collection, settings *LoggerSettings) *MongoLogger {
	return &MongoLogger{Coll: coll, Settings: settings}
}

func (lg *MongoLogger) GetProps() LoggerProps {
	return LoggerProps{
		Settings: lg.Settings,
		Source:   lg.Source,
		Event:    lg.Event,
		EventID:  lg.EventID,
	}
}

func (lg *MongoLogger) SetSource(v string) Logger {
	lg.Source = v
	return lg
}

func (lg *MongoLogger) SetEvent(v string) Logger {
	lg.Event = v
	return lg
}

func (lg *MongoLogger) SetEventID(v string) Logger {
	lg.EventID = v
	return lg
}

func (lg *MongoLogger) log(level, msg string, lf ...LogFields) error {
	log := newLog(level, msg, lg.Source, lg.Event, lg.EventID, lf...)
	_, err := lg.Coll.InsertOne(context.Background(), log)
	fmt.Print(log.String(lg))
	return err
}

func (lg *MongoLogger) LogExists(filter any) (bool, error) {
	if _, ok := filter.(bson.M); !ok {
		return false, errors.New("filter must have a bson.M type")
	}

	var log Log
	if err := lg.Coll.FindOne(context.Background(), filter).Decode(&log); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return !log.Time.IsZero(), nil
}

func (lg *MongoLogger) Info(msg string, lf ...LogFields) error  { return lg.log(INFO, msg, lf...) }
func (lg *MongoLogger) Debug(msg string, lf ...LogFields) error { return lg.log(DEBUG, msg, lf...) }
func (lg *MongoLogger) Warn(msg string, lf ...LogFields) error  { return lg.log(WARN, msg, lf...) }
func (lg *MongoLogger) Trace(msg string, lf ...LogFields) error { return lg.log(TRACE, msg, lf...) }

func (lg *MongoLogger) Error(err error, lf ...LogFields) error {
	if err == nil {
		return lg.log(ERROR, "<nil>", lf...)
	}

	return lg.log(ERROR, err.Error(), lf...)
}

func CreateMongoIndexes(coll *mongo.Collection) error {
	return mongodb.NewIndexes().
		Add("time").
		Add("level").
		Add("source").
		Add("event").
		Add("event_id").
		Create(coll)
}

type LogFilter struct {
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Limit int64     `json:"limit"`
	Skip  int64     `json:"skip"`
	Log   Log       `json:"log" bson:"log"`
}

// FindLogs returns logs that match the filter
func (lg *MongoLogger) FindLogs(filter LogFilter) ([]Log, error) {

	queryFilter := bson.M{}

	// if the from and to fields are not zero, add them to the query filter
	if !filter.From.IsZero() && !filter.To.IsZero() {
		if filter.From.After(filter.To) {
			return nil, errors.New("from date cannot be after to date")
		}

		queryFilter["time"] = bson.M{"$gte": filter.From, "$lte": filter.To}
	}

	if filter.Log.Level != "" {
		queryFilter["level"] = filter.Log.Level
	}

	if filter.Log.Source != "" {
		queryFilter["source"] = filter.Log.Source
	}

	if filter.Log.Event != "" {
		queryFilter["event"] = filter.Log.Event
	}

	if filter.Log.EventID != "" {
		queryFilter["event_id"] = filter.Log.EventID
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "time", Value: -1}}). // sort by time field in descending order
		SetLimit(filter.Limit).
		SetSkip(filter.Skip)

	var docs []Log
	cursor, err := lg.Coll.Find(context.Background(), queryFilter, opts)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &docs)
	return docs, err
}
