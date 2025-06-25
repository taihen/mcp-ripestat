package whatsmyip

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get(t *testing.T) {
	// Create a test server that returns a mock response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/whats-my-ip/data.json" {
			t.Errorf("Expected path '/data/whats-my-ip/data.json', got %s", r.URL.Path)
		}

		response := `{
			"data": {
				"ip": "203.0.113.1"
			},
			"time": "2023-01-01T12:00:00Z"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create a client with the test server
	httpClient := &http.Client{}
	c := client.New(server.URL, httpClient)
	whatsMyIPClient := NewClient(c)

	ctx := context.Background()
	result, err := whatsMyIPClient.Get(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IP != "203.0.113.1" {
		t.Errorf("Expected IP '203.0.113.1', got %s", result.IP)
	}

	if result.FetchedAt != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected FetchedAt '2023-01-01T12:00:00Z', got %s", result.FetchedAt)
	}
}

func TestClient_GetWithClientIP(t *testing.T) {
	c := NewClient(nil)

	ctx := context.Background()
	clientIP := "192.0.2.1"
	result, err := c.GetWithClientIP(ctx, clientIP)

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

func TestClient_GetWithClientIP_EmptyIP(t *testing.T) {
	// Create a test server that returns a mock response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := `{
			"data": {
				"ip": "203.0.113.1"
			},
			"time": "2023-01-01T12:00:00Z"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create a client with the test server
	httpClient := &http.Client{}
	c := client.New(server.URL, httpClient)
	whatsMyIPClient := NewClient(c)

	ctx := context.Background()
	result, err := whatsMyIPClient.GetWithClientIP(ctx, "")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IP != "203.0.113.1" {
		t.Errorf("Expected IP '203.0.113.1', got %s", result.IP)
	}

	if result.FetchedAt != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected FetchedAt '2023-01-01T12:00:00Z', got %s", result.FetchedAt)
	}
}

func TestClient_Get_ServerError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Create a client with the test server
	httpClient := &http.Client{}
	c := client.New(server.URL, httpClient)
	whatsMyIPClient := NewClient(c)

	ctx := context.Background()
	_, err := whatsMyIPClient.Get(ctx)

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

func TestGetWhatsMyIP(_ *testing.T) {
	// This test will likely fail without internet, but we test the function exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := GetWhatsMyIP(ctx)
	// We don't assert on the error since it might fail due to network issues
	// We just ensure the function doesn't panic
	_ = err
}

func TestGetWhatsMyIPWithClientIP(t *testing.T) {
	ctx := context.Background()
	clientIP := "192.0.2.1"
	result, err := GetWhatsMyIPWithClientIP(ctx, clientIP)

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

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "X-Forwarded-For single IP",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.1",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 192.168.1.1, 10.0.0.1",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.2",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.2",
		},
		{
			name: "CF-Connecting-IP",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.3",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.3",
		},
		{
			name: "X-Forwarded-For takes precedence",
			headers: map[string]string{
				"X-Forwarded-For":  "203.0.113.1",
				"X-Real-IP":        "203.0.113.2",
				"CF-Connecting-IP": "203.0.113.3",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.1",
		},
		{
			name:       "No proxy headers - use RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "203.0.113.4:12345",
			expected:   "203.0.113.4",
		},
		{
			name:       "No proxy headers - RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "203.0.113.5",
			expected:   "203.0.113.5",
		},
		{
			name: "X-Forwarded-For with spaces",
			headers: map[string]string{
				"X-Forwarded-For": " 203.0.113.6 , 192.168.1.1 ",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.6",
		},
		{
			name: "Google Cloud Run format - client IP and load balancer IP",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.7, 35.191.0.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expected:   "203.0.113.7",
		},
		{
			name: "Google Cloud Run format - existing header with client and LB IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 203.0.113.8, 35.191.0.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.100", // First IP in traditional multi-proxy case
		},
		{
			name: "Invalid IP in X-Forwarded-For should be skipped",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip, 203.0.113.9, 35.191.0.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expected:   "203.0.113.9",
		},
		{
			name: "All invalid IPs should fallback to RemoteAddr",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip, another-invalid",
			},
			remoteAddr: "203.0.113.10:12345",
			expected:   "203.0.113.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := ExtractClientIP(req)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	c := client.DefaultClient()
	whatsMyIPClient := NewClient(c)

	if whatsMyIPClient.client != c {
		t.Error("Expected client to be set correctly")
	}
}

func TestNewClient_NilClient(t *testing.T) {
	whatsMyIPClient := NewClient(nil)

	if whatsMyIPClient.client == nil {
		t.Error("Expected client to be initialized when nil is passed")
	}
}

func TestDefaultClient(t *testing.T) {
	whatsMyIPClient := DefaultClient()

	if whatsMyIPClient.client == nil {
		t.Error("Expected client to be initialized")
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"203.0.113.1", true},
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"2001:db8::1", true},
		{"::1", true},
		{"invalid-ip", false},
		{"256.256.256.256", false},
		{"", false},
		{"192.168.1", false},
		{"192.168.1.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isValidIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isValidIP(%q) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}
