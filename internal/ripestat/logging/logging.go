// Package logging provides standardized logging for the RIPEstat API client.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// LogLevel represents the severity level of a log message.
type LogLevel int

const (
	// LogLevelDebug is the most verbose log level.
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for general operational information.
	LogLevelInfo
	// LogLevelWarning is for important but non-critical issues.
	LogLevelWarning
	// LogLevelError is for critical issues that require attention.
	LogLevelError
	// LogLevelNone disables all logging.
	LogLevelNone
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	case LogLevelNone:
		return "NONE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", l)
	}
}

// Logger is a simple logger for the RIPEstat API client.
type Logger struct {
	mu       sync.Mutex
	level    LogLevel
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	writer   io.Writer
}

// NewLogger creates a new Logger with the specified log level and writer.
// If writer is nil, os.Stderr is used.
func NewLogger(level LogLevel, writer io.Writer) *Logger {
	if writer == nil {
		writer = os.Stderr
	}

	return &Logger{
		level:    level,
		debugLog: log.New(writer, "[DEBUG] ", log.LstdFlags),
		infoLog:  log.New(writer, "[INFO] ", log.LstdFlags),
		warnLog:  log.New(writer, "[WARNING] ", log.LstdFlags),
		errorLog: log.New(writer, "[ERROR] ", log.LstdFlags),
		writer:   writer,
	}
}

// SetLevel sets the log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level.
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.level
}

// Debug logs a debug message if the log level is LogLevelDebug or lower.
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.GetLevel() <= LogLevelDebug {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.debugLog.Printf(format, v...)
	}
}

// Info logs an info message if the log level is LogLevelInfo or lower.
func (l *Logger) Info(format string, v ...interface{}) {
	if l.GetLevel() <= LogLevelInfo {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.infoLog.Printf(format, v...)
	}
}

// Warning logs a warning message if the log level is LogLevelWarning or lower.
func (l *Logger) Warning(format string, v ...interface{}) {
	if l.GetLevel() <= LogLevelWarning {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.warnLog.Printf(format, v...)
	}
}

// Error logs an error message if the log level is LogLevelError or lower.
func (l *Logger) Error(format string, v ...interface{}) {
	if l.GetLevel() <= LogLevelError {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.errorLog.Printf(format, v...)
	}
}

// DefaultLogger is the default logger used by the RIPEstat API client.
var DefaultLogger = NewLogger(LogLevelInfo, nil)

// SetDefaultLogLevel sets the log level for the default logger.
func SetDefaultLogLevel(level LogLevel) {
	DefaultLogger.SetLevel(level)
}

// Debug logs a debug message to the default logger.
func Debug(format string, v ...interface{}) {
	DefaultLogger.Debug(format, v...)
}

// Info logs an info message to the default logger.
func Info(format string, v ...interface{}) {
	DefaultLogger.Info(format, v...)
}

// Warning logs a warning message to the default logger.
func Warning(format string, v ...interface{}) {
	DefaultLogger.Warning(format, v...)
}

// Error logs an error message to the default logger.
func Error(format string, v ...interface{}) {
	DefaultLogger.Error(format, v...)
}
