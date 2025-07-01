//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/prefixoverview"
)

func TestPrefixOverviewE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := prefixoverview.GetPrefixOverview(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("unexpected resource in data: %+v", resp.Data)
	}
}
