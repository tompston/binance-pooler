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

func (logger *MongoLogger) GetProps() LoggerProps {
	return LoggerProps{
		Settings: logger.Settings,
		Source:   logger.Source,
		Event:    logger.Event,
		EventID:  logger.EventID,
	}
}

func (logger *MongoLogger) SetSource(v string) Logger {
	logger.Source = v
	return logger
}

func (logger *MongoLogger) SetEvent(v string) Logger {
	logger.Event = v
	return logger
}

func (logger *MongoLogger) SetEventID(v string) Logger {
	logger.EventID = v
	return logger
}

// Todo: how to implement the saving of key-value pairs?
func (logger *MongoLogger) log(level, msg string, lf ...LogFields) error {
	log := newLog(level, msg, logger.Source, logger.Event, logger.EventID, lf...)
	_, err := logger.Coll.InsertOne(context.Background(), log)
	fmt.Print(log.String(logger))
	return err
}

func (logger *MongoLogger) LogExists(filter any) (bool, error) {
	if _, ok := filter.(bson.M); !ok {
		return false, errors.New("filter must have a bson.M type")
	}

	var log Log
	if err := logger.Coll.FindOne(context.Background(), filter).Decode(&log); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return !log.Time.IsZero(), nil
}

func (logger *MongoLogger) Info(msg string, lf ...LogFields) error {
	return logger.log(INFO, msg, lf...)
}

func (logger *MongoLogger) Debug(msg string, lf ...LogFields) error {
	return logger.log(DEBUG, msg, lf...)
}

func (logger *MongoLogger) Warn(msg string, lf ...LogFields) error {
	return logger.log(WARN, msg, lf...)
}

func (logger *MongoLogger) Trace(msg string, lf ...LogFields) error {
	return logger.log(TRACE, msg, lf...)
}

func (logger *MongoLogger) Error(err error, lf ...LogFields) error {
	if err == nil {
		return logger.log(ERROR, "<nil>", lf...)
	}

	return logger.log(ERROR, err.Error(), lf...)
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
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
	Log  Log       `json:"log" bson:"log"`
}

// FindLogs returns logs that match the filter
func (logger *MongoLogger) FindLogs(filter LogFilter, limit int64, skip int64) ([]Log, error) {

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
		SetLimit(limit).
		SetSkip(skip)

	var docs []Log
	cursor, err := logger.Coll.Find(context.Background(), queryFilter, opts)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &docs)
	return docs, err
}
