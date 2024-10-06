package logger

import "fmt"

// Logger implementation for console
type ConsoleLogger struct {
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

func (logger *ConsoleLogger) log(level, msg string, lf ...LogFields) error {
	log := newLog(level, msg, logger.Source, logger.Event, logger.EventID, lf...)
	_, err := fmt.Print(log)
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

func (logger *ConsoleLogger) Info(msg string, lf ...LogFields) error {
	return logger.log(INFO, msg, lf...)
}

func (logger *ConsoleLogger) Debug(msg string, lf ...LogFields) error {
	return logger.log(DEBUG, msg, lf...)
}

func (logger *ConsoleLogger) Warn(msg string, lf ...LogFields) error {
	return logger.log(WARN, msg, lf...)
}

func (logger *ConsoleLogger) Trace(msg string, lf ...LogFields) error {
	return logger.log(TRACE, msg, lf...)
}

func (logger *ConsoleLogger) Error(err error, lf ...LogFields) error {
	if err == nil {
		return logger.log(ERROR, "<nil>", lf...)
	}

	return logger.log(ERROR, err.Error(), lf...)
}

func (logger *ConsoleLogger) LogExists(filter any) (bool, error) {
	return false, fmt.Errorf("method cannot be used with ConsoleLogger")
}

func (logger *ConsoleLogger) FindLogs(filter LogFilter, limit int64, skip int64) ([]Log, error) {
	return nil, fmt.Errorf("method cannot be used with ConsoleLogger")
}
