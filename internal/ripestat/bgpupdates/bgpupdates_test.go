package bgpupdates

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	ripestaterrors "github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

func TestClient_Get_Success(t *testing.T) {
	mockResponse := `{
		"data": {
			"resource": "8.8.8.8",
			"query_starttime": "2024-01-01T00:00:00Z",
			"query_endtime": "2024-01-01T23:59:59Z",
			"updates": [
				{
					"seq": 1,
					"timestamp": "2024-01-01T12:00:00Z",
					"type": "announcement",
					"attrs": {
						"source_id": "rrc00",
						"target_prefix": "8.8.8.0/24",
						"path": [15169, 3356],
						"community": ["15169:3356"]
					}
				}
			],
			"nr_updates": 1
		},
		"messages": [],
		"status": "ok",
		"status_code": 200
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	client := NewClient(c)

	ctx := context.Background()
	resp, err := client.Get(ctx, "8.8.8.8")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Data.Resource != "8.8.8.8" {
		t.Errorf("Expected resource '8.8.8.8', got %s", resp.Data.Resource)
	}
	if resp.Data.NumUpdates != 1 {
		t.Errorf("Expected 1 update, got %d", resp.Data.NumUpdates)
	}
	if len(resp.Data.Updates) != 1 {
		t.Errorf("Expected 1 update in slice, got %d", len(resp.Data.Updates))
	}
	if resp.Data.Updates[0].Attributes.SourceID != "rrc00" {
		t.Errorf("Expected source_id 'rrc00', got %s", resp.Data.Updates[0].Attributes.SourceID)
	}
	if resp.Data.Updates[0].Attributes.TargetPrefix != "8.8.8.0/24" {
		t.Errorf("Expected target_prefix '8.8.8.0/24', got %s", resp.Data.Updates[0].Attributes.TargetPrefix)
	}
}

func TestClient_GetWithOptions_ForwardsTimeBounds(t *testing.T) {
	var gotStartTime, gotEndTime string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotStartTime = r.URL.Query().Get("starttime")
		gotEndTime = r.URL.Query().Get("endtime")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {"resource": "8.8.8.0/24", "updates": [], "nr_updates": 0},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer ts.Close()

	c := NewClient(client.New(ts.URL, ts.Client()))
	_, err := c.GetWithOptions(context.Background(), "8.8.8.0/24", &GetOptions{
		StartTime: "2026-08-14T10:00:00Z",
		EndTime:   "2026-08-21T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("GetWithOptions() error = %v", err)
	}
	if gotStartTime != "2026-08-14T10:00:00Z" {
		t.Errorf("starttime = %q", gotStartTime)
	}
	if gotEndTime != "2026-08-21T10:00:00Z" {
		t.Errorf("endtime = %q", gotEndTime)
	}
}

func TestClient_Get_RejectsExcessiveUpdateCount(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {"resource": "8.8.8.0/24", "updates": [], "nr_updates": 10001},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer ts.Close()

	c := NewClient(client.New(ts.URL, ts.Client()))
	result, err := c.Get(context.Background(), "8.8.8.0/24")
	if err == nil {
		t.Fatal("Get() expected an excessive update count error")
	}
	if result != nil {
		t.Errorf("Get() result = %#v, want nil", result)
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	client := NewClient(nil)
	ctx := context.Background()

	resp, err := client.Get(ctx, "")

	if resp != nil {
		t.Error("Expected nil response for empty resource")
	}
	if err == nil {
		t.Fatal("Expected error for empty resource, got nil")
	}
	targetErr, ok := errors.AsType[*ripestaterrors.Error](err)
	if !ok {
		t.Errorf("Expected ripestaterrors.Error, got %T", err)
	} else if targetErr.Message != ripestaterrors.ErrInvalidParameter.Message {
		t.Errorf("Expected InvalidParameter error, got %v", targetErr.Message)
	}
}

func TestClient_Get_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	client := NewClient(c)

	ctx := context.Background()
	resp, err := client.Get(ctx, "8.8.8.8")

	if resp != nil {
		t.Error("Expected nil response for server error")
	}
	if err == nil {
		t.Fatal("Expected error for server error, got nil")
	}
	targetErr, ok := errors.AsType[*ripestaterrors.Error](err)
	if !ok {
		t.Errorf("Expected ripestaterrors.Error, got %T", err)
	} else if targetErr.Message != ripestaterrors.ErrServerError.Message {
		t.Errorf("Expected ServerError, got %v", targetErr.Message)
	}
}

func TestClient_Get_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	client := NewClient(c)

	ctx := context.Background()
	resp, err := client.Get(ctx, "8.8.8.8")

	if resp != nil {
		t.Error("Expected nil response for invalid JSON")
	}
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
	targetErr, ok := errors.AsType[*ripestaterrors.Error](err)
	if !ok {
		t.Errorf("Expected ripestaterrors.Error, got %T", err)
	} else if targetErr.Message != ripestaterrors.ErrServerError.Message {
		t.Errorf("Expected ServerError, got %v", targetErr.Message)
	}
}

func TestDefaultClient(t *testing.T) {
	client := DefaultClient()
	if client == nil {
		t.Error("Expected non-nil client")
		return
	}
	if client.client == nil {
		t.Error("Expected non-nil client.client")
	}
}

func TestGetBGPUpdates_EmptyResource(t *testing.T) {
	ctx := context.Background()
	resp, err := GetBGPUpdates(ctx, "")

	if resp != nil {
		t.Error("Expected nil response for empty resource")
	}
	if err == nil {
		t.Fatal("Expected error for empty resource, got nil")
	}
	targetErr, ok := errors.AsType[*ripestaterrors.Error](err)
	if !ok {
		t.Errorf("Expected ripestaterrors.Error, got %T", err)
	} else if targetErr.Message != ripestaterrors.ErrInvalidParameter.Message {
		t.Errorf("Expected InvalidParameter error, got %v", targetErr.Message)
	}
}
