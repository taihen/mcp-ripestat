//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/asroutingconsistency"
)

func TestASRoutingConsistencyE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := asroutingconsistency.GetASRoutingConsistency(ctx, "AS3333")
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}
	if resp.Data.Prefixes == nil {
		t.Errorf("unexpected nil prefixes in data: %+v", resp.Data)
	}
	if resp.Data.Imports == nil {
		t.Errorf("unexpected nil imports in data: %+v", resp.Data)
	}
}
