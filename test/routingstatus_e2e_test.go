//go:build e2e

package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/routingstatus"
)

func TestRoutingStatusE2E(t *testing.T) {
	baseURL := "https://stat.ripe.net"
	client := routingstatus.NewClient(baseURL, http.DefaultClient)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("ValidResource", func(t *testing.T) {
		resource := "193.0.0.0/21" // RIPE NCC allocation
		response, err := client.Get(ctx, resource)
		if err != nil {
			t.Fatalf("live call failed: %v", err)
		}
		if response == nil {
			t.Fatal("expected non-nil response")
		}

		if response.Data.Resource != resource {
			t.Errorf("expected resource %s, got %s", resource, response.Data.Resource)
		}
		if response.Data.FirstSeen.Origin == "" {
			t.Error("expected non-empty first_seen origin")
		}
		if response.Data.LastSeen.Origin == "" {
			t.Error("expected non-empty last_seen origin")
		}
		if response.Data.Visibility.V4.TotalRISPeers == 0 {
			t.Error("expected non-zero total RIS peers")
		}
		if len(response.Data.Origins) < 1 {
			t.Error("expected at least one origin")
		}
	})

	t.Run("EmptyResource", func(t *testing.T) {
		response, err := client.Get(ctx, "")
		if err == nil {
			t.Error("expected error for empty resource")
		}
		if response != nil {
			t.Errorf("expected nil response, got %+v", response)
		}
	})

	t.Run("InvalidResource", func(t *testing.T) {
		response, err := client.Get(ctx, "invalid-resource")
		// The API might still return a valid response object, but with potentially empty/default values
		// or it might return an error, both are acceptable
		if err != nil {
			t.Logf("invalid resource error: %v", err)
		} else if response == nil {
			t.Error("expected non-nil response for invalid resource")
		}
	})
}
