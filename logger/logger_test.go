package logger

import "testing"

func TestLoggerImplementations(t *testing.T) {
	tests := []struct {
		name   string
		logger Logger
	}{
		{name: "console logger", logger: ConsoleLogger{}},
		{name: "noop logger", logger: NoopLogger{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logger.Info("ready")
			tt.logger.Error("failed")
			LogStartup(tt.logger, "cli-calculator")
		})
	}
}

func TestLoggerInterface(t *testing.T) {
	var _ Logger = ConsoleLogger{}
	var _ Logger = NoopLogger{}
}
