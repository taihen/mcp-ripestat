package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/countryasns"
)

func TestCountryASNs_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("BasicQuery", func(t *testing.T) {
		resp, err := countryasns.GetCountryASNs(ctx, "nl", nil)
		if err != nil {
			t.Fatalf("GetCountryASNs failed: %v", err)
		}

		if resp.Status != "ok" {
			t.Errorf("Expected status 'ok', got %s", resp.Status)
		}

		if len(resp.Data.Countries) == 0 {
			t.Error("Expected at least one country in response")
		}

		country := resp.Data.Countries[0]
		if country.Resource != "nl" {
			t.Errorf("Expected resource 'nl', got %s", country.Resource)
		}

		if country.Stats.Registered <= 0 {
			t.Errorf("Expected positive number of registered ASNs, got %d", country.Stats.Registered)
		}

		if country.Stats.Routed <= 0 {
			t.Errorf("Expected positive number of routed ASNs, got %d", country.Stats.Routed)
		}
	})

	t.Run("DetailedQuery", func(t *testing.T) {
		opts := &countryasns.GetOptions{LOD: 1}
		resp, err := countryasns.GetCountryASNs(ctx, "nl", opts)
		if err != nil {
			t.Fatalf("GetCountryASNs with LOD=1 failed: %v", err)
		}

		if resp.Status != "ok" {
			t.Errorf("Expected status 'ok', got %s", resp.Status)
		}

		if len(resp.Data.Countries) == 0 {
			t.Error("Expected at least one country in response")
		}

		country := resp.Data.Countries[0]
		if country.Routed == "" {
			t.Error("Expected routed ASNs list to be populated with LOD=1")
		}
	})

	t.Run("SmallCountry", func(t *testing.T) {
		// Test with a smaller country that should have fewer ASNs
		resp, err := countryasns.GetCountryASNs(ctx, "va", nil) // Vatican City
		if err != nil {
			t.Fatalf("GetCountryASNs for Vatican failed: %v", err)
		}

		if resp.Status != "ok" {
			t.Errorf("Expected status 'ok', got %s", resp.Status)
		}

		// Vatican might have 0 ASNs, which is valid
		if len(resp.Data.Countries) > 0 {
			country := resp.Data.Countries[0]
			if country.Resource != "va" {
				t.Errorf("Expected resource 'va', got %s", country.Resource)
			}
		}
	})

	t.Run("InvalidCountryCode", func(t *testing.T) {
		_, err := countryasns.GetCountryASNs(ctx, "invalid", nil)
		if err == nil {
			t.Error("Expected error for invalid country code")
		}
	})
}
