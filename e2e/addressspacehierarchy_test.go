//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/addressspacehierarchy"
)

func TestAddressSpaceHierarchyE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := addressspacehierarchy.GetAddressSpaceHierarchy(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("unexpected resource in data: %+v", resp.Data)
	}
	if resp.Data.RIR != "ripe" {
		t.Errorf("unexpected RIR in data: %+v", resp.Data)
	}
	if len(resp.Data.Exact) == 0 {
		t.Errorf("expected at least one exact entry: %+v", resp.Data)
	}
}
