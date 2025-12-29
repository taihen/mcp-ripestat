package whatsmyip

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	EndpointPath = "/data/whats-my-ip/data.json"
)

// ProxyConfig holds configuration for trusted proxy validation.
type ProxyConfig struct {
	// TrustedProxies contains CIDR ranges of trusted proxies.
	// Only requests from these IPs will have their X-Forwarded-For headers trusted.
	TrustedProxies []*net.IPNet
}

var (
	// defaultProxyConfig holds the default proxy configuration.
	defaultProxyConfig *ProxyConfig
	proxyConfigOnce    sync.Once
)

// DefaultProxyConfig returns the default proxy configuration.
// It reads trusted proxies from the TRUSTED_PROXIES environment variable,
// which should be a comma-separated list of CIDR ranges.
// Common private ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8)
// are trusted by default for local/container deployments.
func DefaultProxyConfig() *ProxyConfig {
	proxyConfigOnce.Do(func() {
		defaultProxyConfig = &ProxyConfig{
			TrustedProxies: make([]*net.IPNet, 0),
		}

		// Default private ranges commonly used in container/proxy deployments
		defaultCIDRs := []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"127.0.0.0/8",
			"::1/128",
			"fc00::/7",
		}

		// Add user-configured trusted proxies from environment
		if envProxies := os.Getenv("TRUSTED_PROXIES"); envProxies != "" {
			for _, cidr := range strings.Split(envProxies, ",") {
				cidr = strings.TrimSpace(cidr)
				if cidr != "" {
					defaultCIDRs = append(defaultCIDRs, cidr)
				}
			}
		}

		for _, cidr := range defaultCIDRs {
			_, network, err := net.ParseCIDR(cidr)
			if err == nil {
				defaultProxyConfig.TrustedProxies = append(defaultProxyConfig.TrustedProxies, network)
			}
		}
	})
	return defaultProxyConfig
}

// IsTrustedProxy checks if the given IP is from a trusted proxy.
func (c *ProxyConfig) IsTrustedProxy(ip string) bool {
	if c == nil || len(c.TrustedProxies) == 0 {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, network := range c.TrustedProxies {
		if network.Contains(parsedIP) {
			return true
		}
	}
	return false
}

// extractRemoteIP extracts the IP address from RemoteAddr (host:port format).
func extractRemoteIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}

type Client struct {
	client *client.Client
}

func NewClient(c *client.Client) *Client {
	if c == nil {
		c = client.DefaultClient()
	}

	return &Client{client: c}
}

func DefaultClient() *Client {
	return NewClient(nil)
}

func (c *Client) Get(ctx context.Context) (*APIResponse, error) {
	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, nil, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get IP address: %w", err))
	}

	apiResponse := &APIResponse{
		IP:        response.Data.IP,
		FetchedAt: response.Time,
	}

	return apiResponse, nil
}

func (c *Client) GetWithClientIP(ctx context.Context, clientIP string) (*APIResponse, error) {
	if clientIP == "" {
		return c.Get(ctx)
	}

	apiResponse := &APIResponse{
		IP:        clientIP,
		FetchedAt: "",
	}

	return apiResponse, nil
}

func GetWhatsMyIP(ctx context.Context) (*APIResponse, error) {
	return DefaultClient().Get(ctx)
}

func GetWhatsMyIPWithClientIP(ctx context.Context, clientIP string) (*APIResponse, error) {
	return DefaultClient().GetWithClientIP(ctx, clientIP)
}

// ExtractClientIP extracts the real client IP from an HTTP request.
// It only trusts proxy headers (X-Forwarded-For, X-Real-IP, CF-Connecting-IP)
// if the request comes from a trusted proxy as configured in DefaultProxyConfig.
func ExtractClientIP(r *http.Request) string {
	return ExtractClientIPWithConfig(r, DefaultProxyConfig())
}

// ExtractClientIPWithConfig extracts the real client IP using the provided proxy config.
// If config is nil, proxy headers are not trusted.
func ExtractClientIPWithConfig(r *http.Request, config *ProxyConfig) string {
	remoteIP := extractRemoteIP(r.RemoteAddr)

	// Only trust proxy headers if the request comes from a trusted proxy
	if config == nil || !config.IsTrustedProxy(remoteIP) {
		// Not from a trusted proxy - return the direct connection IP
		return remoteIP
	}

	// Request is from a trusted proxy - we can trust the proxy headers
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for i, ip := range ips {
			ips[i] = strings.TrimSpace(ip)
		}

		// Return the first valid IP (the original client)
		for _, ip := range ips {
			if isValidIP(ip) {
				return ip
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		xri = strings.TrimSpace(xri)
		if isValidIP(xri) {
			return xri
		}
	}

	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		cfip = strings.TrimSpace(cfip)
		if isValidIP(cfip) {
			return cfip
		}
	}

	// Fallback to the remote address
	return remoteIP
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
