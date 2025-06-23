// Package errors provides error handling for the RIPEstat API client.
package errors

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewError(t *testing.T) {
	err := NewError("test error", http.StatusBadRequest)

	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got %q", err.Message)
	}

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}

	if err.Err != nil {
		t.Errorf("Expected nil wrapped error, got %v", err.Err)
	}
}

func TestError_WithError(t *testing.T) {
	baseErr := errors.New("wrapped error")
	err := NewError("test error", http.StatusBadRequest).WithError(baseErr)

	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got %q", err.Message)
	}

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}

	if err.Err != baseErr {
		t.Errorf("Expected wrapped error %v, got %v", baseErr, err.Err)
	}
}

func TestError_WithError_Nil(t *testing.T) {
	err := NewError("test error", http.StatusBadRequest).WithError(nil)

	if err.Err != nil {
		t.Errorf("Expected nil wrapped error, got %v", err.Err)
	}
}

func TestError_Error(t *testing.T) {
	// Test without wrapped error
	err := NewError("test error", http.StatusBadRequest)
	if err.Error() != "test error" {
		t.Errorf("Expected error string 'test error', got %q", err.Error())
	}

	// Test with wrapped error
	wrappedErr := errors.New("wrapped error")
	err = err.WithError(wrappedErr)
	expected := "test error: wrapped error"

	if err.Error() != expected {
		t.Errorf("Expected error string %q, got %q", expected, err.Error())
	}
}

func TestError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	err := NewError("test error", http.StatusBadRequest).WithError(wrappedErr)

	if !errors.Is(err, wrappedErr) {
		t.Errorf("Expected errors.Is to return true for wrapped error")
	}

	if unwrapped := err.Unwrap(); unwrapped != wrappedErr {
		t.Errorf("Expected Unwrap to return wrapped error %v, got %v", wrappedErr, unwrapped)
	}
}

func TestFromHTTPResponse(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		expectedErrMsg string
	}{
		{"BadRequest", http.StatusBadRequest, "invalid parameter"},
		{"Unauthorized", http.StatusUnauthorized, "unauthorized"},
		{"Forbidden", http.StatusForbidden, "forbidden"},
		{"NotFound", http.StatusNotFound, "resource not found"},
		{"Timeout", http.StatusGatewayTimeout, "request timed out"},
		{"ServerError", http.StatusInternalServerError, "server error"},
		{"OtherClientError", 418, "custom message"}, // I'm a teapot
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			resp.WriteHeader(tc.statusCode)

			err := FromHTTPResponse(resp.Result(), "custom message")

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			var e *Error
			if !errors.As(err, &e) {
				t.Errorf("Expected error to be of type *Error")
			}

			if e, ok := err.(*Error); ok {
				if e.StatusCode != tc.statusCode {
					t.Errorf("Expected status code %d, got %d", tc.statusCode, e.StatusCode)
				}

				errStr := fmt.Sprintf("%s", err)
				if !strings.Contains(errStr, tc.expectedErrMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tc.expectedErrMsg, err.Error())
				}

				if !strings.Contains(errStr, fmt.Sprintf("HTTP status: %d", tc.statusCode)) {
					t.Errorf("Expected error message to contain HTTP status, got %q", err.Error())
				}
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	testCases := []struct {
		err        *Error
		message    string
		statusCode int
	}{
		{ErrInvalidParameter, "invalid parameter", http.StatusBadRequest},
		{ErrNotFound, "resource not found", http.StatusNotFound},
		{ErrUnauthorized, "unauthorized", http.StatusUnauthorized},
		{ErrForbidden, "forbidden", http.StatusForbidden},
		{ErrServerError, "server error", http.StatusInternalServerError},
		{ErrTimeout, "request timed out", http.StatusGatewayTimeout},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			if tc.err.Message != tc.message {
				t.Errorf("Expected message %q, got %q", tc.message, tc.err.Message)
			}

			if tc.err.StatusCode != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, tc.err.StatusCode)
			}
		})
	}
}
