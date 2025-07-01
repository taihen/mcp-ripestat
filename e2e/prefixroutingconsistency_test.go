//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/prefixroutingconsistency"
)

func TestPrefixRoutingConsistencyE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("ValidResource", func(t *testing.T) {
		resource := "193.0.0.0/21" // RIPE NCC allocation
		response, err := prefixroutingconsistency.GetPrefixRoutingConsistency(ctx, resource)
		if err != nil {
			t.Fatalf("live call failed: %v", err)
		}
		if response == nil {
			t.Fatal("expected non-nil response")
		}

		if response.Data.Resource != resource {
			t.Errorf("expected resource %s, got %s", resource, response.Data.Resource)
		}
		if response.Data.QueryStartTime == "" {
			t.Error("expected non-empty query_starttime")
		}
		if response.Data.QueryEndTime == "" {
			t.Error("expected non-empty query_endtime")
		}
		// Routes might be empty for some resources, so we don't require it
	})

	t.Run("EmptyResource", func(t *testing.T) {
		response, err := prefixroutingconsistency.GetPrefixRoutingConsistency(ctx, "")
		if err == nil {
			t.Error("expected error for empty resource")
		}
		if response != nil {
			t.Errorf("expected nil response, got %+v", response)
		}
	})

	t.Run("InvalidResource", func(t *testing.T) {
		response, err := prefixroutingconsistency.GetPrefixRoutingConsistency(ctx, "invalid-resource")
		// The API might still return a valid response object, but with potentially empty/default values
		// or it might return an error, both are acceptable
		if err != nil {
			t.Logf("invalid resource error: %v", err)
		} else if response == nil {
			t.Error("expected non-nil response for invalid resource")
		}
	})
}
