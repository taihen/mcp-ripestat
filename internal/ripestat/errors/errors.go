package errors

import (
	"fmt"
	"net/http"
)

var (
	ErrInvalidParameter = NewError("invalid parameter", http.StatusBadRequest)
	ErrNotFound         = NewError("resource not found", http.StatusNotFound)
	ErrUnauthorized     = NewError("unauthorized", http.StatusUnauthorized)
	ErrForbidden        = NewError("forbidden", http.StatusForbidden)
	ErrServerError      = NewError("server error", http.StatusInternalServerError)
	ErrTimeout          = NewError("request timed out", http.StatusGatewayTimeout)
)

type Error struct {
	Message    string
	StatusCode int
	Err        error
}

func NewError(message string, statusCode int) *Error {
	return &Error{
		Message:    message,
		StatusCode: statusCode,
	}
}

func (e *Error) WithError(err error) *Error {
	if err == nil {
		return e
	}

	return &Error{
		Message:    e.Message,
		StatusCode: e.StatusCode,
		Err:        err,
	}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func FromHTTPResponse(resp *http.Response, defaultMessage string) error {
	var baseErr *Error

	switch resp.StatusCode {
	case http.StatusBadRequest:
		baseErr = ErrInvalidParameter
	case http.StatusUnauthorized:
		baseErr = ErrUnauthorized
	case http.StatusForbidden:
		baseErr = ErrForbidden
	case http.StatusNotFound:
		baseErr = ErrNotFound
	case http.StatusGatewayTimeout:
		baseErr = ErrTimeout
	default:
		if resp.StatusCode >= 500 {
			baseErr = ErrServerError
		} else {
			baseErr = NewError(defaultMessage, resp.StatusCode)
		}
	}

	return baseErr.WithError(fmt.Errorf("HTTP status: %d", resp.StatusCode))
}
