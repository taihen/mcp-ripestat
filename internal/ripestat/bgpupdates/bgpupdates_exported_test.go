package bgpupdates_test

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/bgpupdates"
)

func TestGetBGPUpdates_Integration(t *testing.T) {
	ctx := context.Background()
	resp, err := bgpupdates.GetBGPUpdates(ctx, "8.8.8.8")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
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
	if resp.Data.QueryEndTime.IsZero() {
		t.Error("Expected non-zero query end time")
	}
	if !resp.Data.QueryStartTime.Before(resp.Data.QueryEndTime.Time) {
		t.Error("Expected query start time to be before end time")
	}
}

func TestGetBGPUpdates_Integration_WithPrefix(t *testing.T) {
	ctx := context.Background()
	resp, err := bgpupdates.GetBGPUpdates(ctx, "8.8.8.0/24")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Data.Resource != "8.8.8.0/24" {
		t.Errorf("Expected resource '8.8.8.0/24', got %s", resp.Data.Resource)
	}
	if resp.Data.NumUpdates < 0 {
		t.Errorf("Expected num_updates >= 0, got %d", resp.Data.NumUpdates)
	}
}

func TestGetBGPUpdates_Integration_InvalidResource(t *testing.T) {
	ctx := context.Background()
	resp, err := bgpupdates.GetBGPUpdates(ctx, "invalid-resource")

	if resp != nil {
		t.Error("Expected nil response for invalid resource")
	}
	if err == nil {
		t.Fatal("Expected error for invalid resource, got nil")
	}
}

func TestGetBGPUpdates_Integration_EmptyResource(t *testing.T) {
	ctx := context.Background()
	resp, err := bgpupdates.GetBGPUpdates(ctx, "")

	if resp != nil {
		t.Error("Expected nil response for empty resource")
	}
	if err == nil {
		t.Fatal("Expected error for empty resource, got nil")
	}
}

func TestGetBGPUpdates_Integration_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	resp, err := bgpupdates.GetBGPUpdates(ctx, "8.8.8.8")

	if resp != nil {
		t.Error("Expected nil response for timeout")
	}
	if err == nil {
		t.Fatal("Expected error for timeout, got nil")
	}
}
