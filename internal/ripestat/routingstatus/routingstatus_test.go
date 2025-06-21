package routingstatus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	baseURL := "https://example.com"
	httpClient := &http.Client{}

	client := NewClient(baseURL, httpClient)

	if client.baseURL != baseURL {
		t.Errorf("expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.httpClient != httpClient {
		t.Errorf("expected httpClient %v, got %v", httpClient, client.httpClient)
	}
}

func TestClient_Get(t *testing.T) {
	parsedTime, _ := time.Parse(time.RFC3339, "2025-06-21T16:00:00Z")
	mockTime := CustomTime{Time: parsedTime}

	tests := []struct {
		name          string
		client        *Client
		resource      string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		expected      *Response
		expectError   bool
	}{
		{
			name:     "valid request",
			resource: "193.0.0.0/21",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != endpoint {
					t.Errorf("expected path %s, got %s", endpoint, r.URL.Path)
				}

				query := r.URL.Query()
				if query.Get("resource") != "193.0.0.0/21" {
					t.Errorf("expected resource 193.0.0.0/21, got %s", query.Get("resource"))
				}

				response := map[string]interface{}{
					"messages":         []interface{}{},
					"see_also":         []interface{}{},
					"version":          "1.0",
					"data_call_name":   "routing-status",
					"data_call_status": "supported",
					"cached":           false,
					"data": map[string]interface{}{
						"first_seen": map[string]interface{}{
							"prefix": "193.0.0.0/21",
							"origin": "3333",
							"time":   mockTime,
						},
						"last_seen": map[string]interface{}{
							"prefix": "193.0.0.0/21",
							"origin": "3333",
							"time":   mockTime,
						},
						"visibility": map[string]interface{}{
							"v4": map[string]interface{}{
								"ris_peers_seeing": 342,
								"total_ris_peers":  343,
							},
							"v6": map[string]interface{}{
								"ris_peers_seeing": 0,
								"total_ris_peers":  0,
							},
						},
						"origins": []interface{}{
							map[string]interface{}{
								"origin":        3333,
								"route_objects": []string{"RIPE"},
							},
						},
						"less_specifics": []interface{}{},
						"more_specifics": []interface{}{},
						"resource":       "193.0.0.0/21",
						"query_time":     mockTime,
					},
					"query_id":      "query-123",
					"process_time":  25,
					"server_id":     "app123",
					"build_version": "1.0.0",
					"status":        "ok",
					"status_code":   200,
					"time":          "2025-06-21T16:00:00Z",
				}

				json.NewEncoder(w).Encode(response)
			},
			expected: &Response{
				Messages:       []interface{}{},
				SeeAlso:        []interface{}{},
				Version:        "1.0",
				DataCallName:   "routing-status",
				DataCallStatus: "supported",
				Cached:         false,
				Data: Data{
					FirstSeen: RouteInfo{
						Prefix: "193.0.0.0/21",
						Origin: "3333",
						Time:   mockTime,
					},
					LastSeen: RouteInfo{
						Prefix: "193.0.0.0/21",
						Origin: "3333",
						Time:   mockTime,
					},
					Visibility: Visibility{
						V4: AddressVisibility{
							RISPeersSeeing: 342,
							TotalRISPeers:  343,
						},
						V6: AddressVisibility{
							RISPeersSeeing: 0,
							TotalRISPeers:  0,
						},
					},
					Origins: []Origin{
						{
							Origin:       3333,
							RouteObjects: []string{"RIPE"},
						},
					},
					LessSpecifics: []any{},
					MoreSpecifics: []any{},
					Resource:      "193.0.0.0/21",
					QueryTime:     mockTime,
				},
				QueryID:      "query-123",
				ProcessTime:  25,
				ServerID:     "app123",
				BuildVersion: "1.0.0",
				Status:       "ok",
				StatusCode:   200,
				Time:         "2025-06-21T16:00:00Z",
			},
			expectError: false,
		},
		{
			name:     "empty resource",
			resource: "",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called")
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:     "server error",
			resource: "193.0.0.0/21",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:     "invalid json",
			resource: "193.0.0.0/21",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"data":invalid json}`))
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:     "invalid url",
			client:   NewClient("http://[::1]:namedport", http.DefaultClient),
			resource: "193.0.0.0/21",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called")
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "http do error",
			client: &Client{
				baseURL: "http://localhost:12345", // Assuming this port is not in use
				httpClient: &http.Client{
					Timeout: 1 * time.Microsecond, // Very short timeout to force error
				},
			},
			resource: "193.0.0.0/21",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called")
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *Client
			
			// Use predefined client or create a new one
			if tt.client != nil {
				client = tt.client
			} else {
				server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
				defer server.Close()
				client = NewClient(server.URL, server.Client())
			}

			result, err := client.Get(context.Background(), tt.resource)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("expected %+v, got %+v", tt.expected, result)
				}
			}
		})
	}
}
