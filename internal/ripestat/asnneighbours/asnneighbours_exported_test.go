package asnneighbours_test

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/asnneighbours"
)

func ExampleGetASNNeighbours() {
	ctx := context.Background()
	resp, err := asnneighbours.GetASNNeighbours(ctx, "AS1205", 0, "")
	if err != nil {
		// Handle error
		return
	}

	// Use the response
	_ = resp.Resource
	_ = resp.Neighbours
	_ = resp.NeighbourCounts
}

func ExampleGetASNNeighbours_withLOD() {
	ctx := context.Background()
	resp, err := asnneighbours.GetASNNeighbours(ctx, "AS1205", 1, "")
	if err != nil {
		// Handle error
		return
	}

	// With LOD=1, neighbours include additional details
	for _, neighbour := range resp.Neighbours {
		_ = neighbour.ASN
		_ = neighbour.Type
		if neighbour.Power != nil {
			_ = *neighbour.Power
		}
		if neighbour.V4Peers != nil {
			_ = *neighbour.V4Peers
		}
		if neighbour.V6Peers != nil {
			_ = *neighbour.V6Peers
		}
	}
}

func ExampleGetASNNeighbours_withQueryTime() {
	ctx := context.Background()
	resp, err := asnneighbours.GetASNNeighbours(ctx, "AS1205", 0, "2024-01-01T00:00:00")
	if err != nil {
		// Handle error
		return
	}

	// Historical data for the specified time
	_ = resp.QueryTime
	_ = resp.Neighbours
}

func TestExportedFunctions(t *testing.T) {
	// Test that exported functions are accessible from external packages
	ctx := context.Background()

	// Test with invalid parameters to ensure error handling works
	_, err := asnneighbours.GetASNNeighbours(ctx, "", 0, "")
	if err == nil {
		t.Error("expected error for empty resource")
	}

	_, err = asnneighbours.GetASNNeighbours(ctx, "AS1205", 5, "")
	if err == nil {
		t.Error("expected error for invalid lod")
	}
}
