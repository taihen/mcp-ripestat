package abusecontactfinder_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/abusecontactfinder"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestAbuseContactFinder_Integration(t *testing.T) {
	// Create a test server that mimics the RIPEstat API
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
				"abuse_contacts": ["abuse@ripe.net", "security@example.org"],
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

	// Create a client that uses our test server
	httpClient := ts.Client()
	c := client.New(ts.URL, httpClient)
	abuseContactClient := abusecontactfinder.NewClient(c)

	// Test with a valid IP prefix
	ctx := context.Background()
	resp, err := abuseContactClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the response structure
	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if len(resp.Contacts) != 2 {
		t.Errorf("expected 2 contacts, got %d", len(resp.Contacts))
	}

	expectedContacts := []string{"abuse@ripe.net", "security@example.org"}
	for i, expected := range expectedContacts {
		if i >= len(resp.Contacts) || resp.Contacts[i] != expected {
			t.Errorf("expected contact[%d] = %s, got %v", i, expected, resp.Contacts)
		}
	}

	if resp.FetchedAt == "" {
		t.Errorf("expected FetchedAt to be set")
	}
}

func TestAbuseContactFinder_EmptyContacts(t *testing.T) {
	// Create a test server that returns empty contacts
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
		}`))
	}))
	defer ts.Close()

	// Create a client that uses our test server
	httpClient := ts.Client()
	c := client.New(ts.URL, httpClient)
	abuseContactClient := abusecontactfinder.NewClient(c)

	// Test with a resource that has no abuse contacts
	ctx := context.Background()
	resp, err := abuseContactClient.Get(ctx, "192.168.1.0/24")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify empty contacts are handled correctly
	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if len(resp.Contacts) != 0 {
		t.Errorf("expected 0 contacts, got %d", len(resp.Contacts))
	}

	// Ensure it's an empty slice, not nil
	if resp.Contacts == nil {
		t.Errorf("expected empty slice, got nil")
	}
}

func TestAbuseContactFinder_ErrorHandling(t *testing.T) {
	// Test with invalid resource parameter
	c := abusecontactfinder.DefaultClient()
	ctx := context.Background()

	_, err := c.Get(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}

	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}
}
