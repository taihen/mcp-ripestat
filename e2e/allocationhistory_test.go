//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/allocationhistory"
)

func TestAllocationHistoryE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := allocationhistory.GetAllocationHistory(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}
	if resp.Data.Resource == "" || len(resp.Data.Results) == 0 {
		t.Errorf("unexpected empty data: %+v", resp.Data)
	}
}
