// Package whatsmyip provides access to the RIPEstat whats-my-ip API.
package whatsmyip

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for whats-my-ip.
	EndpointPath = "/data/whats-my-ip/data.json"
)

// Client provides methods to interact with the RIPEstat whats-my-ip API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat whats-my-ip API.
func NewClient(c *client.Client) *Client {
	if c == nil {
		c = client.DefaultClient()
	}

	return &Client{client: c}
}

// DefaultClient returns a new Client with default settings.
func DefaultClient() *Client {
	return NewClient(nil)
}

// Get fetches the caller's public IP address.
func (c *Client) Get(ctx context.Context) (*APIResponse, error) {
	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, nil, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get IP address: %w", err))
	}

	// Transform the response to the expected API format
	apiResponse := &APIResponse{
		IP:        response.Data.IP,
		FetchedAt: response.Time,
	}

	return apiResponse, nil
}

// GetWithClientIP fetches the caller's public IP address, but allows overriding the client IP
// for cases where the server is behind a proxy and needs to respect X-Forwarded-For headers.
func (c *Client) GetWithClientIP(ctx context.Context, clientIP string) (*APIResponse, error) {
	if clientIP == "" {
		return c.Get(ctx)
	}

	// For whats-my-ip, we need to make the request appear to come from the specified IP
	// This is typically handled by the RIPEstat service based on the source IP of the request
	// Since we can't change the source IP in our HTTP client, we'll return the provided IP directly
	// This is the expected behavior when behind a proxy
	apiResponse := &APIResponse{
		IP:        clientIP,
		FetchedAt: "", // We don't have a timestamp from RIPEstat in this case
	}

	return apiResponse, nil
}

// GetWhatsMyIP is a convenience function that uses the default client to get the caller's IP address.
func GetWhatsMyIP(ctx context.Context) (*APIResponse, error) {
	return DefaultClient().Get(ctx)
}

// GetWhatsMyIPWithClientIP is a convenience function that uses the default client to get the caller's IP address
// with support for client IP override (for proxy scenarios).
func GetWhatsMyIPWithClientIP(ctx context.Context, clientIP string) (*APIResponse, error) {
	return DefaultClient().GetWithClientIP(ctx, clientIP)
}

// ExtractClientIP extracts the real client IP from HTTP request headers,
// respecting X-Forwarded-For and other proxy headers.
func ExtractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (most common for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// The first IP is typically the original client
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (used by some proxies)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}

	// Fall back to RemoteAddr (direct connection)
	// RemoteAddr includes port, so we need to strip it
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	// If SplitHostPort fails, return RemoteAddr as-is
	return r.RemoteAddr
}
