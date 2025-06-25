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
		ips := strings.Split(xff, ",")

		// Clean up IPs by trimming whitespace
		for i, ip := range ips {
			ips[i] = strings.TrimSpace(ip)
		}

		// Strategy:
		// 1. If exactly 2 IPs: Google Cloud Run format (client-ip, load-balancer-ip)
		// 2. If more than 2 IPs: Traditional approach (first IP is client)
		// 3. Fallback: First valid IP

		if len(ips) == 2 {
			// Likely Google Cloud Run format: client-ip, load-balancer-ip
			// Take the first IP as the client IP
			clientIP := ips[0]
			if isValidIP(clientIP) {
				return clientIP
			}
		} else if len(ips) > 2 {
			// Traditional multi-proxy scenario or existing header + Google Cloud Run
			// Standard approach: first IP is the original client
			firstIP := ips[0]
			if isValidIP(firstIP) {
				return firstIP
			}
		}

		// Fallback: find the first valid IP
		for _, ip := range ips {
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// Check X-Real-IP header (used by some proxies)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		xri = strings.TrimSpace(xri)
		if isValidIP(xri) {
			return xri
		}
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		cfip = strings.TrimSpace(cfip)
		if isValidIP(cfip) {
			return cfip
		}
	}

	// Fall back to RemoteAddr (direct connection)
	// RemoteAddr includes port, so we need to strip it
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	// If SplitHostPort fails, return RemoteAddr as-is
	return r.RemoteAddr
}

// isValidIP checks if the given string is a valid IP address.
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
