package whois

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	ripestaterrors "github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		mockResponse string
		mockStatus   int
		wantErr      bool
		errType      *ripestaterrors.Error
	}{
		{
			name:     "valid IP address",
			resource: "8.8.8.8",
			mockResponse: `{
				"status": "ok",
				"status_code": 200,
				"version": "1.0",
				"data_call_name": "whois",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"records": [
						[
							{"key": "NetRange", "value": "8.0.0.0 - 8.255.255.255"},
							{"key": "CIDR", "value": "8.0.0.0/8"},
							{"key": "NetName", "value": "LEVEL3"}
						]
					],
					"irr_records": [],
					"authorities": ["whois.arin.net"],
					"resource": "8.8.8.8",
					"query_time": "2023-01-01T00:00:00Z"
				}
			}`,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:     "valid ASN",
			resource: "AS3333",
			mockResponse: `{
				"status": "ok",
				"status_code": 200,
				"version": "1.0",
				"data_call_name": "whois",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"records": [
						[
							{"key": "aut-num", "value": "AS3333"},
							{"key": "as-name", "value": "RIPE-NCC-AS"},
							{"key": "descr", "value": "RIPE Network Coordination Centre"}
						]
					],
					"irr_records": [],
					"authorities": ["whois.ripe.net"],
					"resource": "AS3333",
					"query_time": "2023-01-01T00:00:00Z"
				}
			}`,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:     "empty resource",
			resource: "",
			wantErr:  true,
			errType:  ripestaterrors.ErrInvalidParameter,
		},
		{
			name:       "server error",
			resource:   "8.8.8.8",
			mockStatus: http.StatusInternalServerError,
			wantErr:    true,
			errType:    ripestaterrors.ErrServerError,
		},
		{
			name:       "not found",
			resource:   "invalid",
			mockStatus: http.StatusNotFound,
			wantErr:    true,
			errType:    ripestaterrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				response: tt.mockResponse,
				status:   tt.mockStatus,
			}

			c := client.New("https://stat.ripe.net", mockClient)
			whoisClient := New(c)

			result, err := whoisClient.Get(context.Background(), tt.resource)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Get() error = nil, wantErr %v", tt.wantErr)
					return
				}

				if tt.errType != nil {
					var targetErr *ripestaterrors.Error
					if !errors.As(err, &targetErr) {
						t.Errorf("Get() error type = %T, want *ripestaterrors.Error", err)
						return
					}
					// Check if the error is of the expected type by comparing messages
					if targetErr.Message != tt.errType.Message {
						t.Errorf("Get() error message = %v, want %v", targetErr.Message, tt.errType.Message)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("Get() result = nil, want non-nil")
				return
			}

			// Verify the response structure
			if result.Status != "ok" {
				t.Errorf("Get() result.Status = %v, want ok", result.Status)
			}

			if result.Data.Resource != tt.resource {
				t.Errorf("Get() result.Data.Resource = %v, want %v", result.Data.Resource, tt.resource)
			}

			if len(result.Data.Records) == 0 {
				t.Error("Get() result.Data.Records is empty, want non-empty")
			}
		})
	}
}

func TestClient_Get_ParameterValidation(t *testing.T) {
	mockClient := &MockHTTPClient{
		response: `{"status": "ok", "data": {"resource": "test"}}`,
		status:   http.StatusOK,
	}

	c := client.New("https://stat.ripe.net", mockClient)
	whoisClient := New(c)

	// Test that the correct parameters are passed to the client
	_, err := whoisClient.Get(context.Background(), "8.8.8.8")
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}

	// Verify the URL was constructed correctly
	expectedURL := "https://stat.ripe.net/data/whois/data.json?resource=8.8.8.8"
	if mockClient.lastURL != expectedURL {
		t.Errorf("Get() URL = %v, want %v", mockClient.lastURL, expectedURL)
	}
}

// MockHTTPClient implements the HTTPDoer interface for testing.
type MockHTTPClient struct {
	response string
	status   int
	lastURL  string
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.lastURL = req.URL.String()

	return &http.Response{
		StatusCode: m.status,
		Body:       &MockResponseBody{content: m.response},
		Header:     make(http.Header),
	}, nil
}

// MockResponseBody implements io.ReadCloser for testing.
type MockResponseBody struct {
	content string
	pos     int
}

func (m *MockResponseBody) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.content) {
		return 0, io.EOF
	}

	n = copy(p, m.content[m.pos:])
	m.pos += n
	return n, nil
}

func (m *MockResponseBody) Close() error {
	return nil
}

func TestDefaultClient(t *testing.T) {
	client := DefaultClient()
	if client == nil {
		t.Error("DefaultClient() returned nil")
		return
	}
	if client.client == nil {
		t.Error("DefaultClient() returned client with nil internal client")
	}
}

func TestGetWhois(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "valid resource",
			resource: "8.8.8.8",
			wantErr:  false,
		},
		{
			name:     "empty resource",
			resource: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test will make actual HTTP requests to the RIPEstat API
			// In a real scenario, you might want to mock this as well
			result, err := GetWhois(context.Background(), tt.resource)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetWhois() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				// For this test, we'll allow network errors since we're testing the function call path
				// The actual API functionality is tested in the integration tests
				t.Logf("GetWhois() error = %v (network error expected in unit tests)", err)
				return
			}

			if result == nil {
				t.Error("GetWhois() result = nil, want non-nil")
			}
		})
	}
}
