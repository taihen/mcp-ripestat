package mcp

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

const (
	InitializationError             = -32000
	ProtocolError                   = -32001 // Legacy implementation-defined error.
	ResourceError                   = -32002
	ToolError                       = -32003
	HeaderMismatchError             = -32020
	UnsupportedProtocolVersionError = -32022
)

func NewRequest(method string, params interface{}, id interface{}) *Request {
	return &Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

func NewResponse(result interface{}, id interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

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

func NewNotification(method string, params interface{}) *Notification {
	return &Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

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

func (r *Request) IsNotification() bool {
	return r.ID == nil
}

func ParseMessage(data []byte) (interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

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

	if _, hasID := raw["id"]; hasID {
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}
		return &req, nil
	}

	var notif Notification
	if err := json.Unmarshal(data, &notif); err != nil {
		return nil, fmt.Errorf("invalid notification: %w", err)
	}
	return &notif, nil
}
