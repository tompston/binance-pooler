// Package logbook provides a logging interface for the project. Two
// implementations are provided: ConsoleLogger and MongoLogger.
// The package is not called something predictable like
// "logger" because there can be conflicts with other
// packages + logbooks are an actual thing (https://en.wikipedia.org/wiki/Logbook_(nautical))
package logbook

import (
	"fmt"
	"strings"
	"time"
)

// Log struct for storing the log data. Event, EventID, and Fields are optional.
type Log struct {
	Time    time.Time `json:"time" bson:"time"`                             // Time of the log (UTC)
	Level   Level     `json:"level" bson:"level"`                           // Log level
	Message string    `json:"message" bson:"message"`                       // Logged message
	Source  string    `json:"source,omitempty" bson:"source,omitempty"`     // Source of the log (api, pooler, etc.)
	Event   string    `json:"event,omitempty" bson:"event,omitempty"`       // Event of the log (api-auth-request, binance-eth-pooler, etc.)
	EventID string    `json:"event_id,omitempty" bson:"event_id,omitempty"` // (not logged to the console)
	Fields  Fields    `json:"fields,omitempty" bson:"fields,omitempty"`     // Optional fields
}

type Level int16

const (
	TRACE Level = 0
	DEBUG Level = 1
	INFO  Level = 2
	WARN  Level = 3
	ERROR Level = 4
	FATAL Level = 5
)

func (l Level) Prettify() string {
	switch l {
	case ERROR:
		return "error"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	case WARN:
		return "warn"
	case TRACE:
		return "trace"
	case FATAL:
		return "fatal"
	default:
		return "unknown"
	}
}

type Fields map[string]interface{}

func newLog(level Level, msg, source, event, eventID string, fields ...Fields) Log {
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

	// Removing string length reduces ns/op from 933 - 718 (29% faster)

	var b strings.Builder

	b.WriteString(log.Time.In(settings.Location).Format(settings.TimeFormat))
	b.WriteString("  ")
	b.WriteString(fmt.Sprintf("%-6s", log.Level.Prettify()))
	b.WriteString("  ")
	b.WriteString(fmt.Sprintf("%-12s", log.Source))
	b.WriteString(fmt.Sprintf("%-12s", log.Event))
	b.WriteString("  ")
	b.WriteString(log.Message)

	if log.Fields != nil {
		for k, v := range log.Fields {
			b.WriteString(" ")
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(fmt.Sprintf("%v", v))
		}
	}

	b.WriteString("\n")

	return b.String()
}

// Logger interface implements the methods for logging
type Logger interface {
	Error(msg string, lf ...Fields) error
	Info(msg string, lf ...Fields) error
	Debug(msg string, lf ...Fields) error
	Warn(msg string, lf ...Fields) error
	Trace(msg string, lf ...Fields) error
	Fatal(msg string, lf ...Fields) error

	GetTableName() string                     // GetTableName returns the name of the table where the logs are stored
	FindLogs(filter LogFilter) ([]Log, error) // FindLogs returns the logs that match the provided filter
	LogExists(filter any) (bool, error)       // LogExists checks if the log with the provided filter exists.
	GetProps() LoggerProps                    // GetProps returns the properties of the logger
	SetSource(v string) Logger                // SetSource sets the source of the log
	SetEvent(v string) Logger                 // SetEvent sets the event of the log
	SetEventID(v string) Logger               // SetEventID sets the event id of the log
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

type LoggerProps struct {
	Settings *LoggerSettings
	Source   string
	Event    string
	EventID  string
}
