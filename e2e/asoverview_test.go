//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/asoverview"
)

func TestASOverviewE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := asoverview.GetASOverview(ctx, "3333")
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}
	if resp.Data.Resource != "3333" {
		t.Errorf("unexpected resource in data: %+v", resp.Data)
	}
}
