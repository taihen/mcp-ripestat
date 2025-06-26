package countryasns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get_ValidRequest(t *testing.T) {
	mockResponse := `{
		"status": "ok",
		"status_code": 200,
		"data": {
			"countries": [{
				"stats": {
					"registered": 1582,
					"routed": 1058
				},
				"resource": "nl"
			}],
			"resource": ["nl"],
			"query_time": "2025-06-25T00:00:00",
			"lod": ["0"],
			"cache": null,
			"latest_time": "2025-06-25T00:00:00"
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/country-asns/data.json" {
			t.Errorf("Expected path /data/country-asns/data.json, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("resource") != "nl" {
			t.Errorf("Expected resource=nl, got %s", r.URL.Query().Get("resource"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	c := client.New(server.URL, server.Client())
	client := NewClient(c)
	resp, err := client.Get(context.Background(), "nl", nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Expected status ok, got %s", resp.Status)
	}

	if len(resp.Data.Countries) != 1 {
		t.Fatalf("Expected 1 country, got %d", len(resp.Data.Countries))
	}

	country := resp.Data.Countries[0]
	if country.Resource != "nl" {
		t.Errorf("Expected resource nl, got %s", country.Resource)
	}

	if country.Stats.Registered != 1582 {
		t.Errorf("Expected 1582 registered ASNs, got %d", country.Stats.Registered)
	}

	if country.Stats.Routed != 1058 {
		t.Errorf("Expected 1058 routed ASNs, got %d", country.Stats.Routed)
	}
}

func TestClient_Get_WithLOD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("lod") != "1" {
			t.Errorf("Expected lod=1, got %s", r.URL.Query().Get("lod"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok", "status_code": 200, "data": {"countries": []}}`))
	}))
	defer server.Close()

	c := client.New(server.URL, server.Client())
	client := NewClient(c)
	opts := &GetOptions{LOD: 1}
	_, err := client.Get(context.Background(), "nl", opts)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	client := DefaultClient()
	_, err := client.Get(context.Background(), "", nil)

	if err == nil {
		t.Fatal("Expected error for empty resource")
	}

	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("Expected resource required error, got %s", err.Error())
	}
}

func TestClient_Get_InvalidLOD(t *testing.T) {
	client := DefaultClient()
	opts := &GetOptions{LOD: 2}
	_, err := client.Get(context.Background(), "nl", opts)

	if err == nil {
		t.Fatal("Expected error for invalid LOD")
	}

	if !strings.Contains(err.Error(), "lod parameter must be 0 or 1") {
		t.Errorf("Expected invalid LOD error, got %s", err.Error())
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := client.New(server.URL, server.Client())
	client := NewClient(c)
	_, err := client.Get(context.Background(), "nl", nil)

	if err == nil {
		t.Fatal("Expected error for HTTP 500")
	}

	if !strings.Contains(err.Error(), "failed to get country ASNs") {
		t.Errorf("Expected country ASNs error, got %s", err.Error())
	}
}

func TestNewClient_NilClient(t *testing.T) {
	client := NewClient(nil)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestDefaultClient(t *testing.T) {
	client := DefaultClient()
	if client == nil {
		t.Fatal("Expected non-nil default client")
	}
}
