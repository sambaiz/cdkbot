package logger

import "go.uber.org/zap"

// Loggerer is interface of Logger
type Loggerer interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

// Logger logs
type Logger struct {
	logger *zap.Logger
}

// New Logger
func New() *Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return &Logger{
		logger: logger,
	}
}

// Info log
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Error log
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// MockLogger is logger for test
type MockLogger struct{}

// Info log
func (l MockLogger) Info(msg string, fields ...zap.Field) {}

// Error log
func (l MockLogger) Error(msg string, fields ...zap.Field) {}
