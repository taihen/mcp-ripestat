package announcedprefixes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGet_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"messages": [["info", "Results exclude routes with very low visibility."]],
			"see_also": [],
			"version": "1.2",
			"data_call_name": "announced-prefixes",
			"data_call_status": "supported",
			"cached": false,
			"data": {
				"prefixes": [
					{
						"prefix": "193.0.0.0/21",
						"timelines": [
							{
								"starttime": "2025-06-05T00:00:00",
								"endtime": "2025-06-19T00:00:00"
							}
						]
					},
					{
						"prefix": "193.0.10.0/23",
						"timelines": [
							{
								"starttime": "2025-06-05T00:00:00",
								"endtime": "2025-06-19T00:00:00"
							}
						]
					}
				],
				"query_starttime": "2025-06-05T00:00:00",
				"query_endtime": "2025-06-19T00:00:00",
				"resource": "3333",
				"latest_time": "2025-06-19T00:00:00",
				"earliest_time": "2000-08-01T00:00:00"
			},
			"query_id": "test-id",
			"process_time": 21,
			"server_id": "test-server",
			"build_version": "test-build",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-19T09:22:34.976849"
		}`))
	}))
	defer ts.Close()

	client := ts.Client()
	ctx := context.Background()
	resp, err := getWithClient(ctx, "AS3333", client, ts.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Data.Resource != "3333" {
		t.Errorf("expected resource 3333, got %s", resp.Data.Resource)
	}

	if len(resp.Data.Prefixes) != 2 {
		t.Errorf("expected 2 prefixes, got %d", len(resp.Data.Prefixes))
	}

	expectedPrefixes := []string{"193.0.0.0/21", "193.0.10.0/23"}
	for i, prefix := range resp.Data.Prefixes {
		if prefix.Prefix != expectedPrefixes[i] {
			t.Errorf("expected prefix %s, got %s", expectedPrefixes[i], prefix.Prefix)
		}

		if len(prefix.Timelines) != 1 {
			t.Errorf("expected 1 timeline, got %d", len(prefix.Timelines))
		}

		if prefix.Timelines[0].Starttime != "2025-06-05T00:00:00" {
			t.Errorf("expected starttime 2025-06-05T00:00:00, got %s", prefix.Timelines[0].Starttime)
		}

		if prefix.Timelines[0].Endtime != "2025-06-19T00:00:00" {
			t.Errorf("expected endtime 2025-06-19T00:00:00, got %s", prefix.Timelines[0].Endtime)
		}
	}
}

func TestGet_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	client := ts.Client()
	ctx := context.Background()
	_, err := getWithClient(ctx, "AS3333", client, ts.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 502") {
		t.Errorf("expected status code error, got %v", err)
	}
}

func TestGet_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"not_json":`))
	}))
	defer ts.Close()

	client := ts.Client()
	ctx := context.Background()
	_, err := getWithClient(ctx, "AS3333", client, ts.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestGet_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 50 * time.Millisecond}
	ctx := context.Background()
	_, err := getWithClient(ctx, "AS3333", client, ts.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGet_BadURL(t *testing.T) {
	ctx := context.Background()
	_, err := getWithClient(ctx, "AS3333", &http.Client{}, "http://localhost:1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGet_ExportedFunc(t *testing.T) {
	originalGet := Get
	defer func() { Get = originalGet }()

	Get = func(ctx context.Context, resource string) (*AnnouncedPrefixesResponse, error) {
		return &AnnouncedPrefixesResponse{
			Data: AnnouncedPrefixesData{
				Resource: resource,
			},
		}, nil
	}

	ctx := context.Background()
	resp, err := Get(ctx, "AS3333")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Data.Resource != "AS3333" {
		t.Errorf("expected resource AS3333, got %s", resp.Data.Resource)
	}
}
