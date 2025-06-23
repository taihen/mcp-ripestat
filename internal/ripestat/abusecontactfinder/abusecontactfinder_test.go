package abusecontactfinder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name           string
		resource       string
		serverResponse string
		statusCode     int
		wantContacts   []string
		wantErr        bool
	}{
		{
			name:     "successful response with contacts",
			resource: "193.0.0.0/21",
			serverResponse: `{
				"messages": [],
				"see_also": [],
				"version": "2.1",
				"data_call_name": "abuse-contact-finder",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"abuse_contacts": ["abuse@ripe.net"],
					"authoritative_rir": "ripe",
					"latest_time": "2025-06-23T20:09:57",
					"earliest_time": "2025-06-23T20:09:57",
					"parameters": {
						"resource": "193.0.0.0/21",
						"cache": null
					}
				},
				"query_id": "test-query-id",
				"process_time": 46,
				"server_id": "app194",
				"build_version": "main-2025.06.23",
				"status": "ok",
				"status_code": 200,
				"time": "2025-06-23T20:09:57.048781"
			}`,
			statusCode:   200,
			wantContacts: []string{"abuse@ripe.net"},
			wantErr:      false,
		},
		{
			name:     "successful response with empty contacts",
			resource: "192.168.1.0/24",
			serverResponse: `{
				"messages": [],
				"see_also": [],
				"version": "2.1",
				"data_call_name": "abuse-contact-finder",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"abuse_contacts": [],
					"authoritative_rir": "ripe",
					"latest_time": "2025-06-23T20:09:57",
					"earliest_time": "2025-06-23T20:09:57",
					"parameters": {
						"resource": "192.168.1.0/24",
						"cache": null
					}
				},
				"query_id": "test-query-id",
				"process_time": 46,
				"server_id": "app194",
				"build_version": "main-2025.06.23",
				"status": "ok",
				"status_code": 200,
				"time": "2025-06-23T20:09:57.048781"
			}`,
			statusCode:   200,
			wantContacts: []string{},
			wantErr:      false,
		},
		{
			name:     "successful response with null contacts",
			resource: "10.0.0.0/8",
			serverResponse: `{
				"messages": [],
				"see_also": [],
				"version": "2.1",
				"data_call_name": "abuse-contact-finder",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"abuse_contacts": null,
					"authoritative_rir": "ripe",
					"latest_time": "2025-06-23T20:09:57",
					"earliest_time": "2025-06-23T20:09:57",
					"parameters": {
						"resource": "10.0.0.0/8",
						"cache": null
					}
				},
				"query_id": "test-query-id",
				"process_time": 46,
				"server_id": "app194",
				"build_version": "main-2025.06.23",
				"status": "ok",
				"status_code": 200,
				"time": "2025-06-23T20:09:57.048781"
			}`,
			statusCode:   200,
			wantContacts: []string{},
			wantErr:      false,
		},
		{
			name:         "empty resource parameter",
			resource:     "",
			statusCode:   200,
			wantContacts: nil,
			wantErr:      true,
		},
		{
			name:           "server error",
			resource:       "193.0.0.0/21",
			serverResponse: `{"error": "internal server error"}`,
			statusCode:     500,
			wantContacts:   nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.serverResponse != "" {
					_, _ = w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			httpClient := &http.Client{}
			c := NewClient(client.New(server.URL, httpClient))

			got, err := c.Get(context.Background(), tt.resource)

			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Errorf("Client.Get() returned nil response")
					return
				}

				if len(got.Contacts) != len(tt.wantContacts) {
					t.Errorf("Client.Get() contacts length = %v, want %v", len(got.Contacts), len(tt.wantContacts))
					return
				}

				for i, contact := range got.Contacts {
					if contact != tt.wantContacts[i] {
						t.Errorf("Client.Get() contact[%d] = %v, want %v", i, contact, tt.wantContacts[i])
					}
				}

				if got.FetchedAt == "" {
					t.Errorf("Client.Get() FetchedAt is empty")
				}
			}
		})
	}
}

func TestGetAbuseContactFinder(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"messages": [],
			"see_also": [],
			"version": "2.1",
			"data_call_name": "abuse-contact-finder",
			"data_call_status": "supported",
			"cached": false,
			"data": {
				"abuse_contacts": ["abuse@example.com"],
				"authoritative_rir": "ripe",
				"latest_time": "2025-06-23T20:09:57",
				"earliest_time": "2025-06-23T20:09:57",
				"parameters": {
					"resource": "193.0.0.0/21",
					"cache": null
				}
			},
			"query_id": "test-query-id",
			"process_time": 46,
			"server_id": "app194",
			"build_version": "main-2025.06.23",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-23T20:09:57.048781"
		}`))
	}))
	defer ts.Close()

	// Create a custom client that uses our test server
	customClient := client.New(ts.URL, ts.Client())
	abuseContactClient := NewClient(customClient)

	// Test the client directly
	ctx := context.Background()
	resp, err := abuseContactClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Contacts) != 1 || resp.Contacts[0] != "abuse@example.com" {
		t.Errorf("expected contacts [abuse@example.com], got %v", resp.Contacts)
	}

	// Since we can't mock the GetAbuseContactFinder function directly,
	// we're effectively testing that DefaultClient() and Get() work together
	// which is what GetAbuseContactFinder does
}

func TestGetAbuseContactFinder_ConvenienceFunction(t *testing.T) {
	// Test the convenience function with an empty resource to trigger error path
	ctx := context.Background()
	_, err := GetAbuseContactFinder(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}
}

func TestNewClient(t *testing.T) {
	t.Run("with existing client", func(t *testing.T) {
		existingClient := client.DefaultClient()
		c := NewClient(existingClient)
		if c == nil {
			t.Errorf("NewClient() returned nil")
			return
		}
		if c.client != existingClient {
			t.Errorf("NewClient() did not use provided client")
		}
	})
}

func TestDefaultClient(t *testing.T) {
	c := DefaultClient()
	if c == nil {
		t.Errorf("DefaultClient() returned nil")
		return
	}
	if c.client == nil {
		t.Errorf("DefaultClient() client field is nil")
	}
}
