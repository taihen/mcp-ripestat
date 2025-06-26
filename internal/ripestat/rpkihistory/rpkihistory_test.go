package rpkihistory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		mockResponse string
		wantErr      bool
		errContains  string
	}{
		{
			name:     "valid prefix",
			resource: "193.0.22.0/23",
			mockResponse: `{
				"messages": [],
				"see_also": [],
				"version": "0.1",
				"data_call_name": "rpki-history",
				"data_call_status": "development",
				"cached": false,
				"data": {
					"timeseries": [
						{
							"prefix": "193.0.22.0/23",
							"time": "2015-02-11T00:00:00Z",
							"vrp_count": 1,
							"count": 1,
							"family": 4,
							"max_length": 23
						}
					]
				}
			}`,
			wantErr: false,
		},
		{
			name:        "empty resource",
			resource:    "",
			wantErr:     true,
			errContains: "resource parameter is required",
		},
		{
			name:        "invalid resource",
			resource:    "invalid-resource",
			wantErr:     true,
			errContains: "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.mockResponse != "" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.mockResponse))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			httpClient := client.New(server.URL, server.Client())
			c := NewClient(httpClient)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := c.Get(ctx, tt.resource)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Get() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Get() error = %v, want error containing %s", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("Get() result is nil")
				return
			}

			if result.Data.Timeseries == nil {
				t.Error("Get() result.Data.Timeseries is nil")
				return
			}

			if len(result.Data.Timeseries) == 0 {
				t.Error("Get() result.Data.Timeseries is empty")
				return
			}

			entry := result.Data.Timeseries[0]
			if entry.Prefix != "193.0.22.0/23" {
				t.Errorf("Get() result.Data.Timeseries[0].Prefix = %s, want 193.0.22.0/23", entry.Prefix)
			}
			if entry.Family != 4 {
				t.Errorf("Get() result.Data.Timeseries[0].Family = %d, want 4", entry.Family)
			}
		})
	}
}

func TestClient_GetWithHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Internal Server Error"}`))
	}))
	defer server.Close()

	httpClient := client.New(server.URL, server.Client())
	c := NewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.Get(ctx, "193.0.22.0/23")
	if err == nil {
		t.Error("Get() expected error for HTTP 500, got nil")
	}
}

func TestClient_GetWithInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"invalid": json}`))
	}))
	defer server.Close()

	httpClient := client.New(server.URL, server.Client())
	c := NewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.Get(ctx, "193.0.22.0/23")
	if err == nil {
		t.Error("Get() expected error for invalid JSON, got nil")
	}
}

func TestClient_GetWithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := client.New(server.URL, &http.Client{Timeout: 100 * time.Millisecond})
	c := NewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := c.Get(ctx, "193.0.22.0/23")
	if err == nil {
		t.Error("Get() expected timeout error, got nil")
	}
}

func TestGetRPKIHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]interface{}{
			"messages":         []string{},
			"see_also":         []string{},
			"version":          "0.1",
			"data_call_name":   "rpki-history",
			"data_call_status": "development",
			"cached":           false,
			"data": map[string]interface{}{
				"timeseries": []map[string]interface{}{
					{
						"prefix":     "193.0.22.0/23",
						"time":       "2015-02-11T00:00:00Z",
						"vrp_count":  1,
						"count":      1,
						"family":     4,
						"max_length": 23,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	httpClient := client.New(server.URL, server.Client())
	c := NewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := c.Get(ctx, "193.0.22.0/23")
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
		return
	}

	if result == nil {
		t.Error("Get() result is nil")
		return
	}

	if len(result.Data.Timeseries) == 0 {
		t.Error("Get() result.Data.Timeseries is empty")
		return
	}

	entry := result.Data.Timeseries[0]
	if entry.Prefix != "193.0.22.0/23" {
		t.Errorf("Get() result.Data.Timeseries[0].Prefix = %s, want 193.0.22.0/23", entry.Prefix)
	}
}

func TestNewClient(t *testing.T) {
	httpClient := client.New("http://example.com", nil)
	c := NewClient(httpClient)

	if c == nil {
		t.Error("NewClient() returned nil")
		return
	}
	if c.client != httpClient {
		t.Error("NewClient() did not set client correctly")
	}
}

func TestNewClientWithNil(t *testing.T) {
	c := NewClient(nil)
	if c == nil {
		t.Error("NewClient(nil) returned nil")
		return
	}
	if c.client == nil {
		t.Error("NewClient(nil) did not set default client")
	}
}

func TestDefaultClient(t *testing.T) {
	c := DefaultClient()
	if c == nil {
		t.Error("DefaultClient() returned nil")
		return
	}
	if c.client == nil {
		t.Error("DefaultClient() client is nil")
	}
}
