package mcp

import (
	"encoding/json"
	"testing"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest("test", map[string]string{"key": "value"}, 1)
	
	if req.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC to be '2.0', got %s", req.JSONRPC)
	}
	if req.Method != "test" {
		t.Errorf("Expected Method to be 'test', got %s", req.Method)
	}
	if req.ID != 1 {
		t.Errorf("Expected ID to be 1, got %v", req.ID)
	}
}

func TestNewResponse(t *testing.T) {
	resp := NewResponse("result", 1)
	
	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC to be '2.0', got %s", resp.JSONRPC)
	}
	if resp.Result != "result" {
		t.Errorf("Expected Result to be 'result', got %v", resp.Result)
	}
	if resp.ID != 1 {
		t.Errorf("Expected ID to be 1, got %v", resp.ID)
	}
}

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse(InvalidRequest, "Invalid request", "extra data", 1)
	
	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC to be '2.0', got %s", resp.JSONRPC)
	}
	if resp.Error == nil {
		t.Fatal("Expected Error to be non-nil")
	}
	if resp.Error.Code != InvalidRequest {
		t.Errorf("Expected Error.Code to be %d, got %d", InvalidRequest, resp.Error.Code)
	}
	if resp.Error.Message != "Invalid request" {
		t.Errorf("Expected Error.Message to be 'Invalid request', got %s", resp.Error.Message)
	}
	if resp.ID != 1 {
		t.Errorf("Expected ID to be 1, got %v", resp.ID)
	}
}

func TestNewNotification(t *testing.T) {
	notif := NewNotification("test", map[string]string{"key": "value"})
	
	if notif.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC to be '2.0', got %s", notif.JSONRPC)
	}
	if notif.Method != "test" {
		t.Errorf("Expected Method to be 'test', got %s", notif.Method)
	}
}

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		wantErr bool
	}{
		{
			name: "valid request",
			request: Request{
				JSONRPC: "2.0",
				Method:  "test",
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "invalid jsonrpc version",
			request: Request{
				JSONRPC: "1.0",
				Method:  "test",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "missing method",
			request: Request{
				JSONRPC: "2.0",
				Method:  "",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "missing id",
			request: Request{
				JSONRPC: "2.0",
				Method:  "test",
				ID:      nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.ValidateRequest()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantType  string
		wantErr   bool
	}{
		{
			name:     "parse request",
			data:     `{"jsonrpc": "2.0", "method": "test", "id": 1}`,
			wantType: "*mcp.Request",
			wantErr:  false,
		},
		{
			name:     "parse response",
			data:     `{"jsonrpc": "2.0", "result": "success", "id": 1}`,
			wantType: "*mcp.Response",
			wantErr:  false,
		},
		{
			name:     "parse error response",
			data:     `{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid request"}, "id": 1}`,
			wantType: "*mcp.Response",
			wantErr:  false,
		},
		{
			name:     "parse notification",
			data:     `{"jsonrpc": "2.0", "method": "notification"}`,
			wantType: "*mcp.Notification",
			wantErr:  false,
		},
		{
			name:     "invalid json",
			data:     `{invalid json}`,
			wantType: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseMessage([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				msgType := ""
				switch msg.(type) {
				case *Request:
					msgType = "*mcp.Request"
				case *Response:
					msgType = "*mcp.Response"
				case *Notification:
					msgType = "*mcp.Notification"
				}
				if msgType != tt.wantType {
					t.Errorf("ParseMessage() got type %s, want %s", msgType, tt.wantType)
				}
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		code int
		want int
	}{
		{"ParseError", ParseError, -32700},
		{"InvalidRequest", InvalidRequest, -32600},
		{"MethodNotFound", MethodNotFound, -32601},
		{"InvalidParams", InvalidParams, -32602},
		{"InternalError", InternalError, -32603},
		{"InitializationError", InitializationError, -32000},
		{"ProtocolError", ProtocolError, -32001},
		{"ResourceError", ResourceError, -32002},
		{"ToolError", ToolError, -32003},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Errorf("Error code %s = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}

func TestJSONSerialization(t *testing.T) {
	req := NewRequest("test", map[string]string{"key": "value"}, 1)
	
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	
	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse marshaled request: %v", err)
	}
	
	parsedReq, ok := parsed.(*Request)
	if !ok {
		t.Fatalf("Parsed message is not a Request")
	}
	
	if parsedReq.Method != req.Method {
		t.Errorf("Method mismatch: got %s, want %s", parsedReq.Method, req.Method)
	}
	// JSON unmarshaling converts numbers to float64, so we need to compare accordingly
	if parsedReq.ID.(float64) != float64(req.ID.(int)) {
		t.Errorf("ID mismatch: got %v, want %v", parsedReq.ID, req.ID)
	}
}

func TestIsNotification(t *testing.T) {
	// Test request with ID
	req := &Request{
		JSONRPC: "2.0",
		Method:  "test",
		ID:      1,
	}
	if req.IsNotification() {
		t.Error("Request with ID should not be a notification")
	}
	
	// Test request without ID (nil ID)
	reqNoID := &Request{
		JSONRPC: "2.0",
		Method:  "test",
		ID:      nil,
	}
	if !reqNoID.IsNotification() {
		t.Error("Request with nil ID should be a notification")
	}
}

func TestParseMessage_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		wantType string
		wantErr  bool
	}{
		{
			name:     "empty object",
			data:     `{}`,
			wantType: "*mcp.Notification",
			wantErr:  false,
		},
		{
			name:     "notification with method only",
			data:     `{"jsonrpc": "2.0", "method": "test"}`,
			wantType: "*mcp.Notification",
			wantErr:  false,
		},
		{
			name:     "request with null ID",
			data:     `{"jsonrpc": "2.0", "method": "test", "id": null}`,
			wantType: "*mcp.Request",
			wantErr:  false,
		},
		{
			name:     "response with null result",
			data:     `{"jsonrpc": "2.0", "result": null, "id": 1}`,
			wantType: "*mcp.Response",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseMessage([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				msgType := ""
				switch msg.(type) {
				case *Request:
					msgType = "*mcp.Request"
				case *Response:
					msgType = "*mcp.Response"
				case *Notification:
					msgType = "*mcp.Notification"
				}
				if msgType != tt.wantType {
					t.Errorf("ParseMessage() got type %s, want %s", msgType, tt.wantType)
				}
			}
		})
	}
}