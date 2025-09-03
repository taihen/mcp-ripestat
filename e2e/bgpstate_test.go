//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/bgpstate"
)

func TestBGPStateE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := bgpstate.DefaultClient()
	opts := bgpstate.Options{
		Resource: "140.78/16",
	}

	resp, err := client.Get(ctx, opts)
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}

	if resp.Data.Resource == "" {
		t.Errorf("unexpected resource in data: %+v", resp.Data)
	}

	if resp.Data.Timestamp == nil {
		t.Errorf("expected timestamp to be set: %+v", resp.Data)
	}

	if resp.Data.BGPState == nil {
		t.Errorf("expected BGP state to be set: %+v", resp.Data)
	}
}
