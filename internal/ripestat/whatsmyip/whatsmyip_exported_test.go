package whatsmyip_test

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/whatsmyip"
)

func TestGetWhatsMyIP(_ *testing.T) {
	// Test the exported function
	ctx := context.Background()
	_, err := whatsmyip.GetWhatsMyIP(ctx)
	// We don't assert on the error since it might fail due to network issues
	// We just ensure the function doesn't panic
	_ = err
}

func TestGetWhatsMyIPWithClientIP(t *testing.T) {
	ctx := context.Background()
	clientIP := "192.0.2.1"
	result, err := whatsmyip.GetWhatsMyIPWithClientIP(ctx, clientIP)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IP != clientIP {
		t.Errorf("Expected IP '%s', got %s", clientIP, result.IP)
	}

	if result.FetchedAt != "" {
		t.Errorf("Expected empty FetchedAt for client IP override, got %s", result.FetchedAt)
	}
}

func TestDefaultClient(t *testing.T) {
	client := whatsmyip.DefaultClient()
	if client == nil {
		t.Error("DefaultClient() returned nil")
	}
}
