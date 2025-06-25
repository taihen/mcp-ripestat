package routinghistory

import (
	"context"
	"errors"
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
			name:     "valid ASN",
			resource: "AS3333",
			mockResponse: `{
				"status": "ok",
				"status_code": 200,
				"version": "1.0",
				"data_call_name": "routing-history",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"by_origin": [
						{
							"origin": "3333",
							"prefixes": [
								{
									"prefix": "193.0.0.0/21",
									"timelines": [
										{
											"starttime": "2000-01-01T00:00:00",
											"endtime": "2023-12-31T23:59:59",
											"full_peers_seeing": 50.0
										}
									]
								}
							]
						}
					],
					"resource": "AS3333"
				}
			}`,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:     "valid IP prefix",
			resource: "193.0.0.0/21",
			mockResponse: `{
				"status": "ok",
				"status_code": 200,
				"version": "1.0",
				"data_call_name": "routing-history",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"by_origin": [
						{
							"origin": "3333",
							"prefixes": [
								{
									"prefix": "193.0.0.0/21",
									"timelines": [
										{
											"starttime": "2000-01-01T00:00:00",
											"endtime": "2023-12-31T23:59:59",
											"full_peers_seeing": 50.0
										}
									]
								}
							]
						}
					],
					"resource": "193.0.0.0/21"
				}
			}`,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:         "empty resource",
			resource:     "",
			mockResponse: "",
			mockStatus:   http.StatusOK,
			wantErr:      true,
			errType:      ripestaterrors.ErrInvalidParameter,
		},
		{
			name:         "server error",
			resource:     "AS3333",
			mockResponse: `{"error": "internal server error"}`,
			mockStatus:   http.StatusInternalServerError,
			wantErr:      true,
			errType:      ripestaterrors.ErrServerError,
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
			routingHistoryClient := New(c)

			result, err := routingHistoryClient.Get(context.Background(), tt.resource)

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
				t.Error("Get() result is nil, want non-nil")
				return
			}

			if len(result.Data.ByOrigin) == 0 {
				t.Error("Get() result.Data.ByOrigin is empty, want non-empty")
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
	routingHistoryClient := New(c)

	_, err := routingHistoryClient.Get(context.Background(), "AS3333")
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}

	expectedURL := "https://stat.ripe.net/data/routing-history/data.json?resource=AS3333"
	if mockClient.lastURL != expectedURL {
		t.Errorf("Get() called URL = %s, want %s", mockClient.lastURL, expectedURL)
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
		return 0, nil
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

func TestGetRoutingHistory(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "valid ASN",
			resource: "AS3333",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRoutingHistory(context.Background(), tt.resource)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetRoutingHistory() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetRoutingHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("GetRoutingHistory() result is nil, want non-nil")
			}
		})
	}
}
