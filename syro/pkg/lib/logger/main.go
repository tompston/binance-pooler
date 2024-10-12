// Package logger provides a logging interface for the project.
// Two implementations are provided: ConsoleLogger and MongoLogger.
package logger

import (
	"fmt"
	"strings"
	"time"
)

// Log struct for storing the log data. Event, EventID, and Fields are optional.
type Log struct {
	Time    time.Time `json:"time" bson:"time"`                             // Time of the log (UTC)
	Level   string    `json:"level" bson:"level"`                           // Log level
	Message string    `json:"message" bson:"message"`                       // Logged message
	Source  string    `json:"source,omitempty" bson:"source,omitempty"`     // Source of the log (api, pooler, etc.)
	Event   string    `json:"event,omitempty" bson:"event,omitempty"`       // Event of the log (api-auth-request, binance-eth-pooler, etc.)
	EventID string    `json:"event_id,omitempty" bson:"event_id,omitempty"` // (not logged to the console)
	Fields  Fields    `json:"fields,omitempty" bson:"fields,omitempty"`     // Optional fields
}

type Fields map[string]interface{}

func newLog(level string, msg, source, event, eventID string, fields ...Fields) Log {
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
	b.WriteString(log.Level)
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
	Error(msg error, lf ...Fields) error
	Info(msg string, lf ...Fields) error
	Debug(msg string, lf ...Fields) error
	Warn(msg string, lf ...Fields) error
	Trace(msg string, lf ...Fields) error

	// GetProps returns the properties of the logger
	GetProps() LoggerProps
	// LogExists method checks if the log with the provided filter exists.
	LogExists(filter any) (bool, error)
	// FindLogs method returns the logs that match the provided filter
	FindLogs(filter LogFilter) ([]Log, error)
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

/*

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

	var Fields string
	for k, v := range log.Fields {
		Fields += fmt.Sprintf(" %s=%v", k, v)
	}

	time := log.Time.In(settings.Location).Format(settings.TimeFormat)
	return fmt.Sprintf(" %s   %-6s %-10s  %-16s  %v %v\n", time, log.Level, log.Source, log.Event, log.Message, Fields)

	// Removing string length reduces ns/op from 933 - 718 (29% faster)
	// return fmt.Sprintf(" %s   %s %s  %s  %v %v\n", time, log.Level, log.Source, log.Event, log.Message, Fields)

	// var b strings.Builder

	// if log.Fields != nil {
	// 	for k, v := range log.Fields {
	// 		b.WriteString(" ")
	// 		b.WriteString(k)
	// 		b.WriteString("=")
	// 		b.WriteString(fmt.Sprintf("%v", v))
	// 	}
	// }

	// time := log.Time.In(settings.Location).Format(settings.TimeFormat)
	// return fmt.Sprintf(" %s   %-6s %-10s  %-16s  %v %v\n", time, log.Level, log.Source, log.Event, log.Message, b.String())
}

*/
