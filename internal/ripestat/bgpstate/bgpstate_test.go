package bgpstate

import (
	"context"
	"errors"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	ripestaterrors "github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

func TestNewClient(t *testing.T) {
	c := NewClient(nil)
	if c == nil {
		t.Error("Expected client to be created, got nil")
		return
	}

	if c.client == nil {
		t.Error("Expected client.client to be set")
	}
}

func TestDefaultClient(t *testing.T) {
	c := DefaultClient()
	if c == nil {
		t.Error("Expected client to be created, got nil")
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	c := DefaultClient()

	opts := Options{
		Resource: "",
	}

	_, err := c.Get(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for empty resource parameter")
	}

	targetErr, ok := errors.AsType[*ripestaterrors.Error](err)
	if !ok {
		t.Errorf("Expected RIPEstat error, got %T", err)
	} else if targetErr.Message != ripestaterrors.ErrInvalidParameter.Message {
		t.Errorf("Expected InvalidParameter error, got %v", targetErr.Message)
	}
}

func TestClient_Get_ValidResource(t *testing.T) {
	// Skip this test in CI/CD or when testing without network access.
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	c := DefaultClient()

	opts := Options{
		Resource: "140.78/16",
	}

	result, err := c.Get(context.Background(), opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected result to be non-nil")
		return
	}

	if result.Data.Resource == "" {
		t.Error("Expected resource to be set in response")
	}

	if result.Data.BGPState == nil {
		t.Error("Expected BGP state to be set in response")
	}
}

func TestClient_Get_WithOptions(t *testing.T) {
	// Skip this test in CI/CD or when testing without network access.
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	c := DefaultClient()

	opts := Options{
		Resource:       "140.78/16",
		UnixTimestamps: true,
	}

	result, err := c.Get(context.Background(), opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected result to be non-nil")
		return
	}

	if result.Data.Resource == "" {
		t.Error("Expected resource to be set in response")
	}
}

func TestClient_Get_WithCustomClient(t *testing.T) {
	customClient := client.New("", nil)
	c := NewClient(customClient)

	if c.client != customClient {
		t.Error("Expected custom client to be used")
	}
}
