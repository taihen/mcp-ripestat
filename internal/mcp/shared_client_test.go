package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func testWhatsMyIPSharedCache(
	t *testing.T,
	ripeClient *client.Client,
	call func(context.Context) (bool, error),
) {
	t.Helper()

	hits := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != "/data/whats-my-ip/data.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","status_code":200,"data":{"ip":"203.0.113.5"},"time":"2026-09-24T12:00:00"}`))
	}))
	defer ts.Close()

	ripeClient.BaseURL = ts.URL
	ripeClient.HTTPClient = ts.Client()
	ripeClient.RetryConfig.RetryCount = 0

	ctx := context.Background()
	for _, name := range []string{"first", "second"} {
		ok, err := call(ctx)
		if err != nil || !ok {
			t.Fatalf("%s WhatsMyIP call = (ok=%v, err=%v)", name, ok, err)
		}
	}
	if hits != 1 {
		t.Fatalf("whats-my-ip upstream hits = %d, want 1", hits)
	}
}
