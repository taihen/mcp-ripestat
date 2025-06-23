package rpkivalidation_test

import (
	"context"
	"strings"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/rpkivalidation"
)

func TestGetRPKIValidation_ConvenienceFunction(t *testing.T) {
	// Test the convenience function with empty resource to trigger error path
	ctx := context.Background()
	_, err := rpkivalidation.GetRPKIValidation(ctx, "", "193.0.0.0/21")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}

	// Test the convenience function with empty prefix to trigger error path
	_, err = rpkivalidation.GetRPKIValidation(ctx, "3333", "")
	if err == nil {
		t.Fatal("expected error for empty prefix, got nil")
	}
	if !strings.Contains(err.Error(), "prefix parameter is required") {
		t.Errorf("expected prefix required error, got %v", err)
	}
}
