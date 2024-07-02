// Package logger provides a logging interface for the project.
// Two implementations are provided: ConsoleLogger and MongoLogger.
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

// Logger interface implements the methods for logging
type Logger interface {
	Error(msg error) error
	Info(msg any) error
	Debug(msg any) error
	Warn(msg any) error
	Trace(msg any) error

	// GetSettings returns the settings for the logger, which are used when printing the log to the console.
	GetSettings() *LoggerSettings

	// LogExists method checks if the log with the provided filter exists.
	LogExists(filter any) (bool, error)

	// FindLogs method returns the logs that match the provided filter
	FindLogs(filter LogFilter, limit int64, skip int64) ([]Log, error)
}

// LoggerSettings struct for storing the settings for the logger which are used when printing the log
// to the console.
type LoggerSettings struct {
	Location   *time.Location
	TimeFormat string
}

// DefaultLoggerSettings are the default settings for the logger, used if the settings are not provided
// or location is nil.
var DefaultLoggerSettings = &LoggerSettings{
	Location:   time.UTC,
	TimeFormat: "2006-01-02 15:04:05",
}

const (
	ERROR = "ERROR"
	INFO  = "INFO"
	DEBUG = "DEBUG"
	WARN  = "WARN"
	TRACE = "TRACE"
)

// Log struct for storing the log data
type Log struct {
	// When the log was created (UTC)
	Time time.Time `json:"time" bson:"time"`
	// Log level
	Level string `json:"level" bson:"level"`
	// Logged value
	Data any `json:"data" bson:"data"`
	// Source of the log (api, pooler, etc.)
	Source string `json:"source" bson:"source"`
	// Event of the log (request, response, etc.)
	Event string `json:"event" bson:"event"`
	// (not logged to the console)
	EventID string `json:"event_id" bson:"event_id"`
}

func newLog(level string, data any, source, event, eventID string) Log {
	return Log{Time: time.Now().UTC(), Level: level, Data: data, Source: source, Event: event, EventID: eventID}
}

// String method converts the log to a string, using the provided logger settings.
func (log Log) String(logger Logger) string {
	// Use the default settings by default if the settings are not correct
	settings := DefaultLoggerSettings

	// if the logger is not nil and has it has settings with a defined location, use them
	if logger != nil {
		if logger.GetSettings() != nil && logger.GetSettings().Location != nil {
			settings = logger.GetSettings()
		}
	}

	time := log.Time.In(settings.Location).Format(settings.TimeFormat)
	return fmt.Sprintf(" %s   %-6s %-10s  %-16s  %v\n", time, log.Level, log.Source, log.Event, log.Data)
}

// Logger implementation for console
type ConsoleLogger struct {
	Settings *LoggerSettings
	Source   string
	Event    string
	EventID  string
}

func NewConsoleLogger(settings *LoggerSettings) *ConsoleLogger {
	return &ConsoleLogger{Settings: settings}
}

func (logger *ConsoleLogger) GetSettings() *LoggerSettings {
	return logger.Settings
}

func (logger *ConsoleLogger) log(level string, data any) error {
	fmt.Print(newLog(level, data, logger.Source, logger.Event, logger.EventID).String(logger))
	return nil
}

func (logger *ConsoleLogger) SetSource(v string) *ConsoleLogger {
	logger.Source = v
	return logger
}

func (logger *ConsoleLogger) SetEvent(v string) *ConsoleLogger {
	logger.Event = v
	return logger
}

func (logger *ConsoleLogger) SetEventID(v string) *ConsoleLogger {
	logger.EventID = v
	return logger
}

func (logger *ConsoleLogger) Error(v error) error { return logger.log(ERROR, v.Error()) }
func (logger *ConsoleLogger) Info(v any) error    { return logger.log(INFO, v) }
func (logger *ConsoleLogger) Debug(v any) error   { return logger.log(DEBUG, v) }
func (logger *ConsoleLogger) Warn(v any) error    { return logger.log(WARN, v) }
func (logger *ConsoleLogger) Trace(v any) error   { return logger.log(TRACE, v) }

func (logger *ConsoleLogger) LogExists(filter any) (bool, error) {
	return false, fmt.Errorf("method cannot be used with ConsoleLogger")
}

func (logger *ConsoleLogger) FindLogs(filter LogFilter, limit int64, skip int64) ([]Log, error) {
	return nil, fmt.Errorf("method cannot be used with ConsoleLogger")
}

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

func (logger *MongoLogger) GetSettings() *LoggerSettings {
	return logger.Settings
}

func (logger *MongoLogger) SetSource(v string) *MongoLogger {
	logger.Source = v
	return logger
}

func (logger *MongoLogger) SetEvent(v string) *MongoLogger {
	logger.Event = v
	return logger
}

func (logger *MongoLogger) SetEventID(v string) *MongoLogger {
	logger.EventID = v
	return logger
}

func (logger *MongoLogger) log(level string, data any) error {
	log := newLog(level, data, logger.Source, logger.Event, logger.EventID)
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

	return log != (Log{}), nil
}

func (logger *MongoLogger) Error(v error) error { return logger.log(ERROR, v.Error()) }
func (logger *MongoLogger) Info(v any) error    { return logger.log(INFO, v) }
func (logger *MongoLogger) Debug(v any) error   { return logger.log(DEBUG, v) }
func (logger *MongoLogger) Warn(v any) error    { return logger.log(WARN, v) }
func (logger *MongoLogger) Trace(v any) error   { return logger.log(TRACE, v) }

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

	var logs []Log
	cursor, err := logger.Coll.Find(context.Background(), queryFilter, opts)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(context.Background(), &logs); err != nil {
		return nil, err
	}

	return logs, nil
}
