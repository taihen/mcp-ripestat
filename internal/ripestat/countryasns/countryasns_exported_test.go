package countryasns_test

import (
	"context"
	"strings"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/countryasns"
)

func TestGetCountryASNs_EmptyResource(t *testing.T) {
	_, err := countryasns.GetCountryASNs(context.Background(), "", nil)

	if err == nil {
		t.Fatal("Expected error for empty resource")
	}

	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("Expected resource required error, got %s", err.Error())
	}
}

func TestGetCountryASNs_InvalidLOD(t *testing.T) {
	opts := &countryasns.GetOptions{LOD: -1}
	_, err := countryasns.GetCountryASNs(context.Background(), "nl", opts)

	if err == nil {
		t.Fatal("Expected error for invalid LOD")
	}

	if !strings.Contains(err.Error(), "lod parameter must be 0 or 1") {
		t.Errorf("Expected invalid LOD error, got %s", err.Error())
	}
}
