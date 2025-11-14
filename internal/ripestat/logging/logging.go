
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)


type LogLevel int

const (

	LogLevelDebug LogLevel = iota

	LogLevelInfo

	LogLevelWarning

	LogLevelError

	LogLevelNone
)


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


type Logger struct {
	mu       sync.Mutex
	level    LogLevel
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	writer   io.Writer
}



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


func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}


func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.level
}


func (l *Logger) Debug(format string, v ...interface{}) {
	if l.GetLevel() <= LogLevelDebug {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.debugLog.Printf(format, v...)
	}
}


func (l *Logger) Info(format string, v ...interface{}) {
	if l.GetLevel() <= LogLevelInfo {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.infoLog.Printf(format, v...)
	}
}


func (l *Logger) Warning(format string, v ...interface{}) {
	if l.GetLevel() <= LogLevelWarning {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.warnLog.Printf(format, v...)
	}
}


func (l *Logger) Error(format string, v ...interface{}) {
	if l.GetLevel() <= LogLevelError {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.errorLog.Printf(format, v...)
	}
}


var DefaultLogger = NewLogger(LogLevelInfo, nil)


func SetDefaultLogLevel(level LogLevel) {
	DefaultLogger.SetLevel(level)
}


func Debug(format string, v ...interface{}) {
	DefaultLogger.Debug(format, v...)
}


func Info(format string, v ...interface{}) {
	DefaultLogger.Info(format, v...)
}


func Warning(format string, v ...interface{}) {
	DefaultLogger.Warning(format, v...)
}


func Error(format string, v ...interface{}) {
	DefaultLogger.Error(format, v...)
}
