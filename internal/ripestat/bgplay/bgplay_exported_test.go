package bgplay_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/bgplay"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestBGPlay_Integration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"messages": [],
			"see_also": [],
			"version": "1.3",
			"data_call_name": "bgplay",
			"data_call_status": "supported",
			"cached": false,
			"data": {
				"resource": "193.0.6.0/24",
				"query_starttime": "2025-06-16T16:00:00",
				"query_endtime": "2025-06-16T17:00:00",
				"target_prefix": "193.0.6.0/24",
				"rrcs": [0, 1, 3, 5, 7, 10, 12, 13, 14, 15, 16, 18, 19, 20, 21, 23, 24, 25, 26],
				"initial_state": [
					{
						"target_prefix": "193.0.6.0/24",
						"source_id": "rrc00-1",
						"path": [3333],
						"community": []
					},
					{
						"target_prefix": "193.0.6.0/24",
						"source_id": "rrc01-2",
						"path": [1299, 3333],
						"community": ["1299:3000"]
					}
				],
				"events": [
					{
						"type": "A",
						"timestamp": "2025-06-16T16:15:30",
						"attrs": {
							"target_prefix": "193.0.6.0/24",
							"source_id": "rrc03-1",
							"path": [286, 3333],
							"community": []
						}
					},
					{
						"type": "W",
						"timestamp": "2025-06-16T16:45:15",
						"attrs": {
							"target_prefix": "193.0.6.0/24",
							"source_id": "rrc01-2",
							"path": [1299, 3333],
							"community": ["1299:3000"]
						}
					}
				]
			},
			"query_id": "20250616201149-test-bgplay-integration",
			"process_time": 8,
			"server_id": "app195",
			"build_version": "main-2025.05.26",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-16T20:11:49.678721"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	bgplayClient := bgplay.New(c)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := bgplayClient.Get(ctx, "193.0.6.0/24")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Data.Resource != "193.0.6.0/24" {
		t.Errorf("Expected resource '193.0.6.0/24', got '%s'", result.Data.Resource)
	}

	if result.Data.TargetPrefix != "193.0.6.0/24" {
		t.Errorf("Expected target prefix '193.0.6.0/24', got '%s'", result.Data.TargetPrefix)
	}

	if len(result.Data.InitialState) != 2 {
		t.Errorf("Expected 2 initial state records, got %d", len(result.Data.InitialState))
	}

	if len(result.Data.Events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(result.Data.Events))
	}

	if result.Data.Events[0].Type != "A" {
		t.Errorf("Expected first event type 'A', got '%s'", result.Data.Events[0].Type)
	}

	if result.Data.Events[1].Type != "W" {
		t.Errorf("Expected second event type 'W', got '%s'", result.Data.Events[1].Type)
	}

	if len(result.Data.RRCs) == 0 {
		t.Error("Expected RRCs list to be populated")
	}

	if result.Cached {
		t.Error("Expected cached to be false")
	}

	if result.DataCallName != "bgplay" {
		t.Errorf("Expected data_call_name 'bgplay', got '%s'", result.DataCallName)
	}
}
