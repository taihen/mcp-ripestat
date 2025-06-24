package mcp

import (
	"encoding/json"
	"fmt"
)

// JSON-RPC 2.0 message types

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// Notification represents a JSON-RPC 2.0 notification (no ID).
type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Error represents a JSON-RPC 2.0 error.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes.
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// MCP-specific error codes.
const (
	InitializationError = -32000
	ProtocolError       = -32001
	ResourceError       = -32002
	ToolError           = -32003
)

// NewRequest creates a new JSON-RPC request.
func NewRequest(method string, params interface{}, id interface{}) *Request {
	return &Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

// NewResponse creates a new JSON-RPC response.
func NewResponse(result interface{}, id interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

// NewErrorResponse creates a new JSON-RPC error response.
func NewErrorResponse(code int, message string, data interface{}, id interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
}

// NewNotification creates a new JSON-RPC notification.
func NewNotification(method string, params interface{}) *Notification {
	return &Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// ValidateRequest validates a JSON-RPC request.
func (r *Request) ValidateRequest() error {
	if r.JSONRPC != "2.0" {
		return fmt.Errorf("invalid jsonrpc version: %s", r.JSONRPC)
	}
	if r.Method == "" {
		return fmt.Errorf("method is required")
	}
	if r.ID == nil {
		return fmt.Errorf("id is required for requests")
	}
	return nil
}

// IsNotification checks if this is a notification (has no ID).
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// ParseMessage parses a JSON message into appropriate JSON-RPC type.
func ParseMessage(data []byte) (interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Check if it's a response (has result or error field)
	if _, hasResult := raw["result"]; hasResult {
		var resp Response
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("invalid response: %w", err)
		}
		return &resp, nil
	}

	if _, hasError := raw["error"]; hasError {
		var resp Response
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("invalid error response: %w", err)
		}
		return &resp, nil
	}

	// Check if it has an ID (request) or not (notification)
	if _, hasID := raw["id"]; hasID {
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}
		return &req, nil
	}

	// It's a notification
	var notif Notification
	if err := json.Unmarshal(data, &notif); err != nil {
		return nil, fmt.Errorf("invalid notification: %w", err)
	}
	return &notif, nil
}
