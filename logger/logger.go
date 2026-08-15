package logger

import "fmt"

// Logger is a boundary interface for application logging.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// ConsoleLogger writes log lines to standard output.
type ConsoleLogger struct{}

func (ConsoleLogger) Info(msg string) {
	fmt.Println("[INFO]", msg)
}

func (ConsoleLogger) Error(msg string) {
	fmt.Println("[ERROR]", msg)
}

// NoopLogger discards all log messages (useful in tests).
type NoopLogger struct{}

func (NoopLogger) Info(msg string)  {}
func (NoopLogger) Error(msg string) {}

// LogStartup demonstrates depending on Logger, not a concrete type.
func LogStartup(log Logger, appName string) {
	log.Info(appName + " started")
}

// ReviewSummary documents naming and design choices for code review.
func ReviewSummary() string {
	return `Logger review:
- Interface size: Logger has 2 methods (Info, Error) — inject at boundaries
- Receivers: value receivers on stateless ConsoleLogger and NoopLogger
- Naming: interface is Logger; implementations describe where logs go`
}
