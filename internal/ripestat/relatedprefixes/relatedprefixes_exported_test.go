package relatedprefixes_test

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/relatedprefixes"
)

func TestGetRelatedPrefixes_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
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

	t.Logf("Found %d related prefixes for %s", len(result.Data.Prefixes), resource)
}

func TestClient_Get_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	resource := "193.0.0.0/21"

	client := relatedprefixes.DefaultClient()
	result, err := client.Get(ctx, resource)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if result == nil {
		t.Fatal("Get() returned nil result")
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

	t.Logf("Found %d related prefixes for %s", len(result.Data.Prefixes), resource)
}
