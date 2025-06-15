//go:build e2e

package test

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/networkinfo"
)

func TestNetworkInfoE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := networkinfo.GetNetworkInfo(ctx, "140.78.90.50")
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}
	if resp.Data.Prefix == "" || len(resp.Data.ASNs) == 0 {
		t.Errorf("unexpected empty data: %+v", resp.Data)
	}
}
