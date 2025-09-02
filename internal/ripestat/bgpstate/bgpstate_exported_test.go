package bgpstate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/bgpstate"
	ripestaterrors "github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

func TestGetBGPState_EmptyResource(t *testing.T) {
	client := bgpstate.DefaultClient()

	opts := bgpstate.Options{
		Resource: "",
	}

	_, err := client.Get(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for empty resource parameter")
	}

	var targetErr *ripestaterrors.Error
	if !errors.As(err, &targetErr) {
		t.Errorf("Expected RIPEstat error, got %T", err)
	} else if targetErr.Message != ripestaterrors.ErrInvalidParameter.Message {
		t.Errorf("Expected InvalidParameter error, got %v", targetErr.Message)
	}
}

func TestGetBGPState_ValidPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client := bgpstate.DefaultClient()

	opts := bgpstate.Options{
		Resource: "140.78/16",
	}

	result, err := client.Get(context.Background(), opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected result to be non-nil")
		return
	}

	// Validate response structure.
	if result.Data.Resource == "" {
		t.Error("Expected resource to be set")
	}

	if result.Data.Timestamp == nil {
		t.Error("Expected timestamp to be set")
	}

	if result.Data.BGPState == nil {
		t.Error("Expected BGP state to not be nil")
	}

	// Validate BGP route structure if routes exist.
	if len(result.Data.BGPState) > 0 {
		route := result.Data.BGPState[0]

		if route.TargetPrefix == "" {
			t.Error("Expected target prefix to be set")
		}

		if route.SourceID == "" {
			t.Error("Expected source ID to be set")
		}

		if route.Path == nil {
			t.Error("Expected path to not be nil")
		}
	}
}

func TestGetBGPState_WithTimestamp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client := bgpstate.DefaultClient()

	opts := bgpstate.Options{
		Resource:  "140.78/16",
		Timestamp: "2024-01-01T00:00:00Z",
	}

	result, err := client.Get(context.Background(), opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected result to be non-nil")
	}
}

func TestGetBGPState_WithUnixTimestamps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client := bgpstate.DefaultClient()

	opts := bgpstate.Options{
		Resource:       "140.78/16",
		UnixTimestamps: true,
	}

	result, err := client.Get(context.Background(), opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected result to be non-nil")
	}
}

func TestGetBGPState_InvalidResource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client := bgpstate.DefaultClient()

	opts := bgpstate.Options{
		Resource: "invalid-resource",
	}

	_, err := client.Get(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid resource")
	}
}
