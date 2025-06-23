package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	testCases := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarning, "WARNING"},
		{LogLevelError, "ERROR"},
		{LogLevelNone, "NONE"},
		{LogLevel(99), "UNKNOWN(99)"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.level.String() != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, tc.level.String())
			}
		})
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := NewLogger(LogLevelInfo, nil)

	if logger.GetLevel() != LogLevelInfo {
		t.Errorf("Expected initial level to be INFO, got %v", logger.GetLevel())
	}

	logger.SetLevel(LogLevelDebug)
	if logger.GetLevel() != LogLevelDebug {
		t.Errorf("Expected level to be DEBUG after SetLevel, got %v", logger.GetLevel())
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelDebug, &buf)

	logger.Debug("test debug message")
	if !strings.Contains(buf.String(), "test debug message") {
		t.Errorf("Expected debug message to be logged, got %q", buf.String())
	}

	buf.Reset()
	logger.SetLevel(LogLevelInfo)
	logger.Debug("this should not be logged")
	if buf.String() != "" {
		t.Errorf("Expected no debug message when level is INFO, got %q", buf.String())
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelInfo, &buf)

	logger.Info("test info message")
	if !strings.Contains(buf.String(), "test info message") {
		t.Errorf("Expected info message to be logged, got %q", buf.String())
	}

	buf.Reset()
	logger.SetLevel(LogLevelWarning)
	logger.Info("this should not be logged")
	if buf.String() != "" {
		t.Errorf("Expected no info message when level is WARNING, got %q", buf.String())
	}
}

func TestLogger_Warning(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelWarning, &buf)

	logger.Warning("test warning message")

	if !strings.Contains(buf.String(), "test warning message") {
		t.Errorf("Expected warning message to be logged, got %q", buf.String())
	}

	buf.Reset()
	logger.SetLevel(LogLevelError)
	logger.Warning("this should not be logged")
	if buf.String() != "" {
		t.Errorf("Expected no warning message when level is ERROR, got %q", buf.String())
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelError, &buf)

	logger.Error("test error message")
	if !strings.Contains(buf.String(), "test error message") {
		t.Errorf("Expected error message to be logged, got %q", buf.String())
	}

	buf.Reset()
	logger.SetLevel(LogLevelNone)
	logger.Error("this should not be logged")
	if buf.String() != "" {
		t.Errorf("Expected no error message when level is NONE, got %q", buf.String())
	}
}

func TestDefaultLogger(t *testing.T) {
	// Save the original default logger
	originalLogger := DefaultLogger
	defer func() {
		DefaultLogger = originalLogger
	}()

	var buf bytes.Buffer
	DefaultLogger = NewLogger(LogLevelDebug, &buf)

	Debug("test debug function")

	if !strings.Contains(buf.String(), "test debug function") {
		t.Errorf("Expected debug message to be logged via Debug function, got %q", buf.String())
	}

	buf.Reset()
	Info("test info function")
	if !strings.Contains(buf.String(), "test info function") {
		t.Errorf("Expected info message to be logged via Info function, got %q", buf.String())
	}

	buf.Reset()
	Warning("test warning function")

	if !strings.Contains(buf.String(), "test warning function") {
		t.Errorf("Expected warning message to be logged via Warning function, got %q", buf.String())
	}

	buf.Reset()
	Error("test error function")
	if !strings.Contains(buf.String(), "test error function") {
		t.Errorf("Expected error message to be logged via Error function, got %q", buf.String())
	}
}

func TestSetDefaultLogLevel(t *testing.T) {
	// Save the original default logger
	originalLogger := DefaultLogger
	defer func() {
		DefaultLogger = originalLogger
	}()

	var buf bytes.Buffer
	DefaultLogger = NewLogger(LogLevelDebug, &buf)

	Debug("debug message before")
	if !strings.Contains(buf.String(), "debug message before") {
		t.Errorf("Expected debug message to be logged before changing level, got %q", buf.String())
	}

	buf.Reset()
	SetDefaultLogLevel(LogLevelError)

	Debug("debug message after")
	if buf.String() != "" {
		t.Errorf("Expected no debug message after changing level to ERROR, got %q", buf.String())
	}

	buf.Reset()
	Error("error message after")
	if !strings.Contains(buf.String(), "error message after") {
		t.Errorf("Expected error message to be logged after changing level to ERROR, got %q", buf.String())
	}
}

func TestLogger_FormatWithArgs(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelDebug, &buf)

	logger.Debug("test %s with %d args", "message", 2)
	if !strings.Contains(buf.String(), "test message with 2 args") {
		t.Errorf("Expected formatted debug message, got %q", buf.String())
	}

	buf.Reset()
	logger.Info("info %s: %.2f", "value", 3.14159)
	if !strings.Contains(buf.String(), "info value: 3.14") {
		t.Errorf("Expected formatted info message, got %q", buf.String())
	}
}
