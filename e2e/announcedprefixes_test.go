//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/announcedprefixes"
)

func TestAnnouncedPrefixesE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := announcedprefixes.GetAnnouncedPrefixes(ctx, "AS3333")
	if err != nil {
		t.Fatalf("live call failed: %v", err)
	}
	if len(resp.Data.Prefixes) == 0 {
		t.Errorf("unexpected empty prefixes list: %+v", resp.Data)
	}
	// Verify that the resource matches what we requested
	if resp.Data.Resource != "3333" {
		t.Errorf("unexpected resource in response: expected 3333, got %s", resp.Data.Resource)
	}
}
