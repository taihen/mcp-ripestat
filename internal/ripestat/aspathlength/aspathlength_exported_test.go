package aspathlength_test

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/aspathlength"
)

func TestGetASPathLength_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := aspathlength.GetASPathLength(ctx, "AS3333")
	if err != nil {
		t.Fatalf("Failed to get AS path length data: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.Data.Resource != "3333" {
		t.Errorf("Expected resource '3333', got %s", resp.Data.Resource)
	}

	if len(resp.Data.Stats) == 0 {
		t.Error("Expected at least one stat entry")
	}

	// Check the structure of the first stat entry
	if len(resp.Data.Stats) > 0 {
		stat := resp.Data.Stats[0]
		if stat.Count <= 0 {
			t.Errorf("Expected positive count, got %d", stat.Count)
		}
		if stat.Location == "" {
			t.Error("Expected location to be non-empty")
		}
		if stat.Stripped.Sum <= 0 {
			t.Errorf("Expected positive stripped sum, got %d", stat.Stripped.Sum)
		}
		if stat.Unstripped.Sum <= 0 {
			t.Errorf("Expected positive unstripped sum, got %d", stat.Unstripped.Sum)
		}
		if stat.Stripped.Min <= 0 {
			t.Errorf("Expected positive stripped min, got %d", stat.Stripped.Min)
		}
		if stat.Stripped.Max <= 0 {
			t.Errorf("Expected positive stripped max, got %d", stat.Stripped.Max)
		}
		if stat.Stripped.Avg <= 0 {
			t.Errorf("Expected positive stripped avg, got %f", stat.Stripped.Avg)
		}
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", resp.Status)
	}
}

func TestGetASPathLength_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := aspathlength.GetASPathLength(ctx, "AS3333")
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}
