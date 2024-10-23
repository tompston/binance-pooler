package logbook

import "fmt"

// Logger implementation for console
type ConsoleLogger struct {
	Settings *LoggerSettings
	Source   string
	Event    string
	EventID  string
}

func NewConsoleLogger(s *LoggerSettings) *ConsoleLogger { return &ConsoleLogger{Settings: s} }
func (lg *ConsoleLogger) GetProps() LoggerProps {
	return LoggerProps{
		Settings: lg.Settings,
		Source:   lg.Source,
		Event:    lg.Event,
		EventID:  lg.EventID,
	}
}

func (lg *ConsoleLogger) log(level, msg string, lf ...Fields) error {
	log := newLog(level, msg, lg.Source, lg.Event, lg.EventID, lf...)
	_, err := fmt.Print(log.String(lg))
	return err
}

func (lg *ConsoleLogger) SetSource(v string) Logger {
	lg.Source = v
	return lg
}

func (lg *ConsoleLogger) SetEvent(v string) Logger {
	lg.Event = v
	return lg
}

func (lg *ConsoleLogger) SetEventID(v string) Logger {
	lg.EventID = v
	return lg
}

func (lg *ConsoleLogger) Info(msg string, lf ...Fields) error  { return lg.log(INFO, msg, lf...) }
func (lg *ConsoleLogger) Debug(msg string, lf ...Fields) error { return lg.log(DEBUG, msg, lf...) }
func (lg *ConsoleLogger) Warn(msg string, lf ...Fields) error  { return lg.log(WARN, msg, lf...) }
func (lg *ConsoleLogger) Trace(msg string, lf ...Fields) error { return lg.log(TRACE, msg, lf...) }

func (lg *ConsoleLogger) Error(err error, lf ...Fields) error {
	if err == nil {
		return lg.log(ERROR, "<nil>", lf...)
	}

	return lg.log(ERROR, err.Error(), lf...)
}

func (lg *ConsoleLogger) LogExists(filter any) (bool, error) {
	return false, fmt.Errorf("method cannot be used with ConsoleLogger")
}

func (lg *ConsoleLogger) FindLogs(filter LogFilter) ([]Log, error) {
	return nil, fmt.Errorf("method cannot be used with ConsoleLogger")
}
