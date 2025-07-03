package e2e

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/relatedprefixes"
)

func TestRelatedPrefixes_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	ctx := context.Background()
	resource := "193.0.0.0/21"

	result, err := relatedprefixes.GetRelatedPrefixes(ctx, resource)
	if err != nil {
		t.Fatalf("GetRelatedPrefixes() failed: %v", err)
	}

	if result == nil {
		t.Fatal("GetRelatedPrefixes() returned nil result")
	}

	if result.Data.Resource != resource {
		t.Errorf("Resource = %v, want %v", result.Data.Resource, resource)
	}

	if result.Data.QueryTime == "" {
		t.Error("QueryTime is empty")
	}

	for i, prefix := range result.Data.Prefixes {
		if prefix.Prefix == "" {
			t.Errorf("Prefix[%d].Prefix is empty", i)
		}
		if prefix.OriginASN == "" {
			t.Errorf("Prefix[%d].OriginASN is empty", i)
		}
		if prefix.Relationship == "" {
			t.Errorf("Prefix[%d].Relationship is empty", i)
		}
	}

	t.Logf("E2E test passed: Found %d related prefixes for %s", len(result.Data.Prefixes), resource)
}

func TestRelatedPrefixes_E2E_InvalidResource(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	ctx := context.Background()
	resource := ""

	_, err := relatedprefixes.GetRelatedPrefixes(ctx, resource)
	if err == nil {
		t.Error("GetRelatedPrefixes() expected error for empty resource, got nil")
	}

	t.Logf("E2E test passed: Empty resource correctly returned error: %v", err)
}
