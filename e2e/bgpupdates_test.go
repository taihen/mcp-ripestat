//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/bgpupdates"
)

func TestBGPUpdatesE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := bgpupdates.GetBGPUpdates(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}

	if resp.Data.Resource != "8.8.8.0/24" {
		t.Errorf("Expected resource '8.8.8.0/24', got %s", resp.Data.Resource)
	}
	if resp.Data.NumUpdates < 0 {
		t.Errorf("Expected num_updates >= 0, got %d", resp.Data.NumUpdates)
	}
	if resp.Data.QueryStartTime.IsZero() {
		t.Error("Expected non-zero query start time")
	}
}
