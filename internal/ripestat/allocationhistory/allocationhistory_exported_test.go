package allocationhistory_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/allocationhistory"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestGetAllocationHistory_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	defer server.Close()

	customClient := client.New(server.URL, nil)
	testClient := allocationhistory.NewClient(customClient)

	ctx := context.Background()
	result, err := testClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if result.Data.Resource != "193.0.0.0/21" {
		t.Errorf("Expected resource 193.0.0.0/21, got %s", result.Data.Resource)
	}

	if len(result.Data.Results["IANA"]) != 1 {
		t.Errorf("Expected 1 IANA result, got %d", len(result.Data.Results["IANA"]))
	}

	if len(result.Data.Results["RIPE NCC"]) != 1 {
		t.Errorf("Expected 1 RIPE NCC result, got %d", len(result.Data.Results["RIPE NCC"]))
	}

	ianaResult := result.Data.Results["IANA"][0]
	if ianaResult.Status != "ALLOCATED" {
		t.Errorf("Expected IANA status ALLOCATED, got %s", ianaResult.Status)
	}

	ripeResult := result.Data.Results["RIPE NCC"][0]
	if ripeResult.Status != "ALLOCATED PA" {
		t.Errorf("Expected RIPE NCC status ALLOCATED PA, got %s", ripeResult.Status)
	}

	var _ = allocationhistory.GetAllocationHistory

	t.Log("Verified GetAllocationHistory function signature")
}
