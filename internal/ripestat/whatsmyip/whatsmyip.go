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
	EndpointPath = "/data/whats-my-ip/data.json"
)

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

func ExtractClientIP(r *http.Request) string {

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {

		ips := strings.Split(xff, ",")

		for i, ip := range ips {
			ips[i] = strings.TrimSpace(ip)
		}

		if len(ips) == 2 {

			clientIP := ips[0]
			if isValidIP(clientIP) {
				return clientIP
			}
		} else if len(ips) > 2 {

			firstIP := ips[0]
			if isValidIP(firstIP) {
				return firstIP
			}
		}

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

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
