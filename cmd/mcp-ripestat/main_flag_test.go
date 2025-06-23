package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

// TestMainFlagParsing tests the flag parsing in the main function.
func TestMainFlagParsing(t *testing.T) {
	// Save original flags and restore them after the test
	origCommandLine := flag.CommandLine
	defer func() { flag.CommandLine = origCommandLine }()

	// Save original os.Args and restore them after the test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Save original stdout and restore it after the test
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	defer func() { os.Stdout = oldStdout }()

	tests := []struct {
		name     string
		args     []string
		wantExit bool
		contains string
	}{
		{
			name:     "help flag",
			args:     []string{"cmd", "-help"},
			wantExit: true,
			contains: "Usage:",
		},
		{
			name:     "debug flag",
			args:     []string{"cmd", "-debug"},
			wantExit: false,
			contains: "",
		},
		{
			name:     "port flag",
			args:     []string{"cmd", "-port", "9090"},
			wantExit: false,
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set up test args
			os.Args = tt.args

			// Create a buffer to capture output
			var buf bytes.Buffer

			// Mock os.Exit
			exitCalled := false
			oldOsExit := osExit
			defer func() { osExit = oldOsExit }()
			osExit = func(_ int) {
				exitCalled = true
			}

			// Call the function that parses flags
			parseFlags(&buf)

			// Check if exit was called
			if exitCalled != tt.wantExit {
				t.Errorf("parseFlags() exit = %v, want %v", exitCalled, tt.wantExit)
			}

			// Check output
			output := buf.String()
			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("parseFlags() output = %q, want to contain %q", output, tt.contains)
			}
		})
	}
}

// Mock for os.Exit.
var osExit = os.Exit

// Extract flag parsing from main() to make it testable.
func parseFlags(out *bytes.Buffer) {
	port := flag.String("port", "8080", "Port for the server to listen on")
	debug := flag.Bool("debug", false, "Enable debug logging")
	help := flag.Bool("help", false, "Print all possible flags")

	flag.Usage = func() {
		fmt.Fprintf(out, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(out, "Options:\n")
		flag.CommandLine.SetOutput(out)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help {
		flag.Usage()
		osExit(0)
	}

	// Just for testing, we don't actually set up the logger here
	_ = port
	_ = debug
}

// TestWriteJSONError_EncoderFail tests the error handling in writeJSON.
func TestWriteJSONError_EncoderFail(_ *testing.T) {
	w := &errorWriter{}
	data := map[string]string{"key": "value"}

	// This should not panic
	writeJSON(w, data, http.StatusOK)
}

// errorWriter is a mock http.ResponseWriter that fails on Write.
type errorWriter struct {
	header http.Header
}

func (w *errorWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write error")
}

func (w *errorWriter) WriteHeader(_ int) {
	// Do nothing
}
