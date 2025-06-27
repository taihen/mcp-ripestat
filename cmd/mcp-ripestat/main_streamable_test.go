package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/mcp"
)

func TestValidateStreamableHTTP(t *testing.T) {
	testCases := []struct {
		name           string
		setupRequest   func() *http.Request
		expectValid    bool
		expectedStatus int
	}{
		{
			name: "valid request without origin",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "http://example.com/mcp", nil)
				req.Header.Set("MCP-Protocol-Version", "2025-06-18")
				return req
			},
			expectValid: true,
		},
		{
			name: "valid request with allowed origin",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "http://example.com/mcp", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("MCP-Protocol-Version", "2025-06-18")
				return req
			},
			expectValid: true,
		},
		{
			name: "invalid origin",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "http://example.com/mcp", nil)
				req.Header.Set("Origin", "https://malicious.com")
				return req
			},
			expectValid:    false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "unsupported protocol version",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "http://example.com/mcp", nil)
				req.Header.Set("MCP-Protocol-Version", "2024-01-01")
				return req
			},
			expectValid:    false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "default protocol version",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "http://example.com/mcp", nil)
				return req
			},
			expectValid: true,
		},
		{
			name: "backward compatible protocol version",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "http://example.com/mcp", nil)
				req.Header.Set("MCP-Protocol-Version", "2025-03-26")
				return req
			},
			expectValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.setupRequest()
			recorder := httptest.NewRecorder()

			result := validateStreamableHTTP(recorder, req)

			if tc.expectValid {
				if !result {
					t.Error("Expected validation to pass")
				}
				if recorder.Code != 0 && recorder.Code != http.StatusOK {
					t.Errorf("Expected no error response, got status %d", recorder.Code)
				}
			} else {
				if result {
					t.Error("Expected validation to fail")
				}
				if recorder.Code != tc.expectedStatus {
					t.Errorf("Expected status %d, got %d", tc.expectedStatus, recorder.Code)
				}
			}
		})
	}
}

func TestIsValidOrigin(t *testing.T) {
	testCases := []struct {
		name     string
		origin   string
		expected bool
	}{
		{
			name:     "localhost http",
			origin:   "http://localhost:3000",
			expected: true,
		},
		{
			name:     "localhost https",
			origin:   "https://localhost:8080",
			expected: true,
		},
		{
			name:     "127.0.0.1 http",
			origin:   "http://127.0.0.1:8080",
			expected: true,
		},
		{
			name:     "127.0.0.1 https",
			origin:   "https://127.0.0.1:3000",
			expected: true,
		},
		{
			name:     "cursor domain",
			origin:   "https://cursor.sh",
			expected: true,
		},
		{
			name:     "claude domain",
			origin:   "https://claude.ai",
			expected: true,
		},
		{
			name:     "disallowed domain",
			origin:   "https://malicious.com",
			expected: false,
		},
		{
			name:     "empty origin",
			origin:   "",
			expected: false,
		},
		{
			name:     "localhost without port",
			origin:   "http://localhost",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidOrigin(tc.origin)
			if result != tc.expected {
				t.Errorf("Expected %v for origin '%s', got %v", tc.expected, tc.origin, result)
			}
		})
	}
}

func TestIsSupportedProtocolVersion(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "current version",
			version:  "2025-06-18",
			expected: true,
		},
		{
			name:     "backward compatible version",
			version:  "2025-03-26",
			expected: true,
		},
		{
			name:     "unsupported old version",
			version:  "2024-01-01",
			expected: false,
		},
		{
			name:     "unsupported future version",
			version:  "2026-01-01",
			expected: false,
		},
		{
			name:     "empty version",
			version:  "",
			expected: false,
		},
		{
			name:     "invalid format",
			version:  "invalid",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isSupportedProtocolVersion(tc.version)
			if result != tc.expected {
				t.Errorf("Expected %v for version '%s', got %v", tc.expected, tc.version, result)
			}
		})
	}
}

func TestGetOrCreateSession(t *testing.T) {
	testCases := []struct {
		name                    string
		existingSessionID       string
		expectNewSession        bool
		expectSessionInResponse bool
	}{
		{
			name:                    "no existing session",
			existingSessionID:       "",
			expectNewSession:        true,
			expectSessionInResponse: true,
		},
		{
			name:                    "existing session",
			existingSessionID:       "existing-session-123",
			expectNewSession:        false,
			expectSessionInResponse: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://example.com/mcp", nil)
			if tc.existingSessionID != "" {
				req.Header.Set("MCP-Session-ID", tc.existingSessionID)
			}

			recorder := httptest.NewRecorder()
			sessionID := getOrCreateSession(req, recorder)

			if tc.expectNewSession {
				if sessionID == "" {
					t.Error("Expected new session ID to be generated")
				}
				if sessionID == tc.existingSessionID {
					t.Error("Expected new session ID, got existing one")
				}
			} else if sessionID != tc.existingSessionID {
				t.Errorf("Expected existing session ID '%s', got '%s'", tc.existingSessionID, sessionID)
			}

			responseSessionID := recorder.Header().Get("MCP-Session-ID")
			if tc.expectSessionInResponse {
				if responseSessionID == "" {
					t.Error("Expected session ID in response header")
				}
			} else {
				if responseSessionID != "" {
					t.Error("Expected no session ID in response header for existing session")
				}
			}
		})
	}
}

func TestGenerateSessionID(t *testing.T) {
	// Generate multiple session IDs and verify they are unique
	sessionIDs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		sessionID := generateSessionID()
		if sessionID == "" {
			t.Error("Generated session ID should not be empty")
		}
		if sessionIDs[sessionID] {
			t.Errorf("Generated duplicate session ID: %s", sessionID)
		}
		sessionIDs[sessionID] = true
	}

	// Verify session ID format (should be hex string of reasonable length)
	sessionID := generateSessionID()
	if len(sessionID) < 16 {
		t.Errorf("Session ID seems too short: %s", sessionID)
	}
}

func TestHandleCORS(t *testing.T) {
	testCases := []struct {
		name               string
		origin             string
		expectOriginHeader bool
	}{
		{
			name:               "valid origin",
			origin:             "http://localhost:3000",
			expectOriginHeader: true,
		},
		{
			name:               "invalid origin",
			origin:             "https://malicious.com",
			expectOriginHeader: false,
		},
		{
			name:               "no origin",
			origin:             "",
			expectOriginHeader: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", "http://example.com/mcp", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			recorder := httptest.NewRecorder()
			handleCORS(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Errorf("Expected status OK, got %d", recorder.Code)
			}

			if tc.expectOriginHeader {
				allowOrigin := recorder.Header().Get("Access-Control-Allow-Origin")
				if allowOrigin != tc.origin {
					t.Errorf("Expected Allow-Origin '%s', got '%s'", tc.origin, allowOrigin)
				}
			} else {
				allowOrigin := recorder.Header().Get("Access-Control-Allow-Origin")
				if allowOrigin != "" {
					t.Errorf("Expected no Allow-Origin header, got '%s'", allowOrigin)
				}
			}

			// Check other CORS headers
			expectedHeaders := map[string]string{
				"Access-Control-Allow-Methods": "POST, GET, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, MCP-Protocol-Version, MCP-Session-ID",
				"Access-Control-Max-Age":       "86400",
			}

			for header, expectedValue := range expectedHeaders {
				actualValue := recorder.Header().Get(header)
				if actualValue != expectedValue {
					t.Errorf("Expected %s '%s', got '%s'", header, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestMCPHandler_MethodRouting(t *testing.T) {
	server := mcp.NewServer("test-server", "1.0.0", false)

	testCases := []struct {
		name           string
		method         string
		origin         string
		body           string
		expectedStatus int
	}{
		{
			name:           "POST request",
			method:         "POST",
			origin:         "http://localhost:3000",
			body:           "{}",
			expectedStatus: http.StatusNoContent, // Empty JSON returns no content
		},
		{
			name:           "GET request",
			method:         "GET",
			origin:         "http://localhost:3000",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed, // Only POST allowed now
		},
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			origin:         "http://localhost:3000",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed, // Only POST allowed now
		},
		{
			name:           "unsupported method",
			method:         "PUT",
			origin:         "http://localhost:3000",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid origin",
			method:         "POST",
			origin:         "https://malicious.com",
			body:           "{}",
			expectedStatus: http.StatusNoContent, // Now processes as regular MCP
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, "http://example.com/mcp", strings.NewReader(tc.body))
			} else {
				req = httptest.NewRequest(tc.method, "http://example.com/mcp", nil)
			}

			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			recorder := httptest.NewRecorder()
			mcpHandler(recorder, req, server)

			if recorder.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, recorder.Code)
			}
		})
	}
}
