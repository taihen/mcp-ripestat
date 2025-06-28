package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/bgplay"
)

func TestBGPlay_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// TODO: BGPlay API is flaky in CI environments, likely due to rate limiting
	// or IP blocking of GitHub Actions runners. The API works fine locally but
	// consistently returns HTTP 500 errors during CI runs. This should be
	// re-enabled once we implement proper mocking or retry logic.
	// See: https://github.com/taihen/mcp-ripestat/issues/XXX
	if os.Getenv("CI") != "" {
		t.Skip("skipping BGPlay E2E test in CI due to external API rate limiting")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testCases := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "Valid IP address",
			resource: "8.8.8.8",
			wantErr:  false,
		},
		{
			name:     "Valid IP prefix",
			resource: "193.0.6.0/24",
			wantErr:  false,
		},
		{
			name:     "Empty resource",
			resource: "",
			wantErr:  true,
		},
		{
			name:     "Invalid resource",
			resource: "invalid-resource",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bgplay.GetBGPlay(ctx, tc.resource)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for resource '%s', got nil", tc.resource)
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error for resource '%s', got %v", tc.resource, err)
				return
			}

			if result == nil {
				t.Errorf("Expected result for resource '%s', got nil", tc.resource)
				return
			}

			if result.DataCallName != "bgplay" {
				t.Errorf("Expected data_call_name 'bgplay', got '%s'", result.DataCallName)
			}

			if result.Data.Resource == "" {
				t.Error("Expected resource to be populated in response")
			}

			if result.Data.QueryStartTime == "" {
				t.Error("Expected query_starttime to be populated in response")
			}

			if result.Data.QueryEndTime == "" {
				t.Error("Expected query_endtime to be populated in response")
			}

			t.Logf("BGPlay result for %s: target_prefix=%s, initial_state_count=%d, events_count=%d",
				tc.resource, result.Data.TargetPrefix, len(result.Data.InitialState), len(result.Data.Events))
		})
	}
}

func TestBGPlay_E2E_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// TODO: Skip BGPlay timeout test in CI for same reasons as main BGPlay test
	if os.Getenv("CI") != "" {
		t.Skip("skipping BGPlay E2E timeout test in CI due to external API rate limiting")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := bgplay.GetBGPlay(ctx, "8.8.8.8")
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}
