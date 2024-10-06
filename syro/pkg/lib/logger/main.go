// Package logger provides a logging interface for the project.
// Two implementations are provided: ConsoleLogger and MongoLogger.
package logger

import (
	"fmt"
	"time"
)

// Log struct for storing the log data
type Log struct {
	// When the log was created (UTC)
	Time time.Time `json:"time" bson:"time"`
	// Log level
	Level string `json:"level" bson:"level"`
	// Logged message
	Message string `json:"message" bson:"message"`
	// Source of the log (api, pooler, etc.)
	Source string `json:"source" bson:"source"`
	// Event of the log (api-auth-request, binance-eth-pooler, etc.)
	Event string `json:"event" bson:"event"`
	// (not logged to the console)
	EventID string `json:"event_id" bson:"event_id"`
	// Optional fields
	Fields LogFields `json:"fields" bson:"fields"`
}

type LogFields map[string]interface{}

func newLog(level string, msg, source, event, eventID string, fields ...LogFields) Log {
	log := Log{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: msg,
		Source:  source,
		Event:   event,
		EventID: eventID,
	}

	if len(fields) == 1 {
		log.Fields = fields[0]
	}

	return log
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

	var logFields string
	if log.Fields != nil {
		for k, v := range log.Fields {
			logFields += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	time := log.Time.In(settings.Location).Format(settings.TimeFormat)
	return fmt.Sprintf(" %s   %-6s %-10s  %-16s  %v %v\n", time, log.Level, log.Source, log.Event, log.Message, logFields)
}

// Logger interface implements the methods for logging
type Logger interface {
	Error(msg error, lf ...LogFields) error
	Info(msg string, lf ...LogFields) error
	Debug(msg string, lf ...LogFields) error
	Warn(msg string, lf ...LogFields) error
	Trace(msg string, lf ...LogFields) error

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

type LoggerProps struct {
	Settings *LoggerSettings
	Source   string
	Event    string
	EventID  string
}
