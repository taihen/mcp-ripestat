package networkinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	ripeStatBaseURL = "https://stat.ripe.net/data/network-info/data.json"
)

var (
	defaultHTTPClient httpDoer = &http.Client{Timeout: 10 * time.Second}
	defaultBaseURL             = ripeStatBaseURL
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// GetNetworkInfo queries the RIPEstat network-info endpoint for the given resource (IP or prefix).
func GetNetworkInfo(ctx context.Context, resource string) (*NetworkInfoResponse, error) {
	return getNetworkInfoWithClient(ctx, resource, defaultHTTPClient, defaultBaseURL)
}

func getNetworkInfoWithClient(ctx context.Context, resource string, client httpDoer, baseURL string) (response *NetworkInfoResponse, err error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RIPEstat base URL: %w", err)
	}
	q := u.Query()
	q.Set("resource", resource)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call RIPEstat: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result NetworkInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
