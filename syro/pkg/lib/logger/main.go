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

// Log struct for storing the log data
type Log struct {
	// When the log was created (UTC)
	Time time.Time `json:"time" bson:"time"`
	// Log level
	Level string `json:"level" bson:"level"`
	// Logged message
	Message any `json:"message" bson:"message"`
	// Source of the log (api, pooler, etc.)
	Source string `json:"source" bson:"source"`
	// Event of the log (request, response, etc.)
	Event string `json:"event" bson:"event"`
	// (not logged to the console)
	EventID string `json:"event_id" bson:"event_id"`
}

func newLog(level string, msg any, source, event, eventID string) Log {
	return Log{Time: time.Now().UTC(), Level: level, Message: msg, Source: source, Event: event, EventID: eventID}
}

// String method converts the log to a string, using the provided logger settings.
func (log Log) String(logger Logger) string {
	// Use the default settings by default if the settings are not correct
	settings := DefaultLoggerSettings

	// if the logger is not nil and has it has settings with a defined location, use them
	if logger != nil {
		props := logger.GetProps()

		if props.Settings != nil && props.Settings.Location != nil {
			settings = props.Settings
		}
	}

	time := log.Time.In(settings.Location).Format(settings.TimeFormat)
	return fmt.Sprintf(" %s   %-6s %-10s  %-16s  %v\n", time, log.Level, log.Source, log.Event, log.Message)
}

// Logger interface implements the methods for logging
type Logger interface {
	Error(msg error) error
	Info(msg any) error
	Debug(msg any) error
	Warn(msg any) error
	Trace(msg any) error

	// GetProps returns the properties of the logger
	GetProps() LoggerProps
	// LogExists method checks if the log with the provided filter exists.
	LogExists(filter any) (bool, error)
	// FindLogs method returns the logs that match the provided filter
	FindLogs(filter LogFilter, limit int64, skip int64) ([]Log, error)
	// SetSource sets the source of the log
	SetSource(v string) Logger
	// SetEvent sets the event of the log
	SetEvent(v string) Logger
	// SetEventID sets the event id of the log
	SetEventID(v string) Logger
}

// LoggerSettings struct for storing the settings for the logger which are
// used when printing the log to the console.
type LoggerSettings struct {
	Location   *time.Location
	TimeFormat string
}

// DefaultLoggerSettings are the default settings for the logger, used if the
// settings are not provided or location is nil.
var DefaultLoggerSettings = &LoggerSettings{
	Location:   time.UTC,
	TimeFormat: "2006-01-02 15:04:05",
}

const (
	ERROR = "error"
	INFO  = "info"
	DEBUG = "debug"
	WARN  = "warn"
	TRACE = "trace"
)

// Logger implementation for console
type ConsoleLogger struct {
	Settings *LoggerSettings
	Source   string
	Event    string
	EventID  string
}

type LoggerProps struct {
	Settings *LoggerSettings
	Source   string
	Event    string
	EventID  string
}

func NewConsoleLogger(s *LoggerSettings) *ConsoleLogger { return &ConsoleLogger{Settings: s} }
func (logger *ConsoleLogger) GetProps() LoggerProps {
	return LoggerProps{
		Settings: logger.Settings,
		Source:   logger.Source,
		Event:    logger.Event,
		EventID:  logger.EventID,
	}
}

func (logger *ConsoleLogger) log(level string, data any) error {
	_, err := fmt.Print(newLog(level, data, logger.Source, logger.Event, logger.EventID).String(logger))
	return err
}

func (logger *ConsoleLogger) SetSource(v string) Logger {
	logger.Source = v
	return logger
}

func (logger *ConsoleLogger) SetEvent(v string) Logger {
	logger.Event = v
	return logger
}

func (logger *ConsoleLogger) SetEventID(v string) Logger {
	logger.EventID = v
	return logger
}

func (logger *ConsoleLogger) Info(msg any) error  { return logger.log(INFO, msg) }
func (logger *ConsoleLogger) Debug(msg any) error { return logger.log(DEBUG, msg) }
func (logger *ConsoleLogger) Warn(msg any) error  { return logger.log(WARN, msg) }
func (logger *ConsoleLogger) Trace(msg any) error { return logger.log(TRACE, msg) }
func (logger *ConsoleLogger) Error(err error) error {
	if err == nil {
		return logger.log(ERROR, "nil")
	}

	return logger.log(ERROR, err.Error())
}

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

func (logger *MongoLogger) log(level string, msg any) error {
	log := newLog(level, msg, logger.Source, logger.Event, logger.EventID)
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

func (logger *MongoLogger) Info(msg any) error  { return logger.log(INFO, msg) }
func (logger *MongoLogger) Debug(msg any) error { return logger.log(DEBUG, msg) }
func (logger *MongoLogger) Warn(msg any) error  { return logger.log(WARN, msg) }
func (logger *MongoLogger) Trace(msg any) error { return logger.log(TRACE, msg) }
func (logger *MongoLogger) Error(err error) error {
	if err == nil {
		return logger.log(ERROR, "nil")
	}

	return logger.log(ERROR, err.Error())
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
