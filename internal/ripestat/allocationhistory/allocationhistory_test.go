package allocationhistory

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"messages":[],
			"see_also":[],
			"version":"1.1",
			"data_call_name":"allocation-history",
			"data_call_status":"supported",
			"cached":false,
			"data":{
				"results":{
					"IANA":[{
						"resource":"193.0.0.0/21",
						"status":"ALLOCATED",
						"timelines":[{
							"starttime":"1993-05-01T00:00:00Z",
							"endtime":"2025-06-15T23:59:59Z"
						}]
					}],
					"RIPE NCC":[{
						"resource":"193.0.0.0/21",
						"status":"ALLOCATED PA",
						"timelines":[{
							"starttime":"1993-05-01T00:00:00Z",
							"endtime":"2025-06-15T23:59:59Z"
						}]
					}]
				},
				"resource":"193.0.0.0/21",
				"query_starttime":"1993-05-01T00:00:00Z",
				"query_endtime":"2025-06-15T23:59:59Z"
			},
			"query_id":"test-id",
			"process_time":1,
			"server_id":"test-server",
			"build_version":"test-build",
			"status":"ok",
			"status_code":200,
			"time":"2025-06-15T16:31:58.741967"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	allocationHistoryClient := NewClient(c)

	ctx := context.Background()
	resp, err := allocationHistoryClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("expected resource 193.0.0.0/21, got %s", resp.Data.Resource)
	}
	if len(resp.Data.Results["IANA"]) != 1 {
		t.Errorf("expected 1 IANA result, got %d", len(resp.Data.Results["IANA"]))
	}
	if len(resp.Data.Results["RIPE NCC"]) != 1 {
		t.Errorf("expected 1 RIPE NCC result, got %d", len(resp.Data.Results["RIPE NCC"]))
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	allocationHistoryClient := NewClient(c)

	ctx := context.Background()
	_, err := allocationHistoryClient.Get(ctx, "193.0.0.0/21")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP status: 502") {
		t.Errorf("expected status code error, got %v", err)
	}
}

func TestClient_Get_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"not_json":`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	allocationHistoryClient := NewClient(c)

	ctx := context.Background()
	_, err := allocationHistoryClient.Get(ctx, "193.0.0.0/21")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestClient_Get_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messages":[],"see_also":[],"version":"1.1","data_call_name":"allocation-history","data_call_status":"supported","cached":false,"data":{"results":{},"resource":"193.0.0.0/21","query_starttime":"1993-05-01T00:00:00Z","query_endtime":"2025-06-15T23:59:59Z"},"query_id":"test-id","process_time":1,"server_id":"test-server","build_version":"test-build","status":"ok","status_code":200,"time":"2025-06-15T16:31:58.741967"}`))
	}))
	defer ts.Close()

	httpClient := &http.Client{Timeout: 50 * time.Millisecond}
	c := client.New(ts.URL, httpClient)
	allocationHistoryClient := NewClient(c)

	ctx := context.Background()
	_, err := allocationHistoryClient.Get(ctx, "193.0.0.0/21")

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGetAllocationHistory_Exported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messages":[],"see_also":[],"version":"1.1","data_call_name":"allocation-history","data_call_status":"supported","cached":false,"data":{"results":{},"resource":"193.0.0.0/21","query_starttime":"1993-05-01T00:00:00Z","query_endtime":"2025-06-15T23:59:59Z"},"query_id":"test-id","process_time":1,"server_id":"test-server","build_version":"test-build","status":"ok","status_code":200,"time":"2025-06-15T16:31:58.741967"}`))
	}))
	defer ts.Close()

	customClient := client.New(ts.URL, ts.Client())
	allocationHistoryClient := NewClient(customClient)

	ctx := context.Background()
	resp, err := allocationHistoryClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("expected resource 193.0.0.0/21, got %s", resp.Data.Resource)
	}
}

func TestGetAllocationHistory_ConvenienceFunction(t *testing.T) {
	ctx := context.Background()
	_, err := GetAllocationHistory(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	c := DefaultClient()
	ctx := context.Background()
	_, err := c.Get(ctx, "")

	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}
}
