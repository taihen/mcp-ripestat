package rpkivalidation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name           string
		resource       string
		prefix         string
		serverResponse string
		serverStatus   int
		want           *APIResponse
		wantErr        bool
	}{
		{
			name:     "successful valid RPKI validation",
			resource: "3333",
			prefix:   "193.0.0.0/21",
			serverResponse: `{
				"status": "ok",
				"time": "2025-06-23T20:42:05.326049",
				"data": {
					"validating_roas": [
						{
							"origin": "3333",
							"prefix": "193.0.0.0/21",
							"max_length": 21,
							"validity": "valid"
						}
					],
					"status": "valid",
					"validator": "routinator",
					"resource": "3333",
					"prefix": "193.0.0.0/21"
				}
			}`,
			serverStatus: http.StatusOK,
			want: &APIResponse{
				Status:    "valid",
				Validator: "routinator",
				Resource:  "3333",
				Prefix:    "193.0.0.0/21",
				ValidatingROAs: []ValidatingROA{
					{
						Origin:    "3333",
						Prefix:    "193.0.0.0/21",
						MaxLength: 21,
						Validity:  "valid",
					},
				},
				FetchedAt: "2025-06-23T20:42:05.326049",
			},
			wantErr: false,
		},
		{
			name:     "successful invalid_asn RPKI validation",
			resource: "1234",
			prefix:   "8.8.8.0/24",
			serverResponse: `{
				"status": "ok",
				"time": "2025-06-23T20:42:11.586624",
				"data": {
					"validating_roas": [
						{
							"origin": "15169",
							"prefix": "8.8.8.0/24",
							"max_length": 24,
							"validity": "invalid_asn"
						}
					],
					"status": "invalid_asn",
					"validator": "routinator",
					"resource": "1234",
					"prefix": "8.8.8.0/24"
				}
			}`,
			serverStatus: http.StatusOK,
			want: &APIResponse{
				Status:    "invalid_asn",
				Validator: "routinator",
				Resource:  "1234",
				Prefix:    "8.8.8.0/24",
				ValidatingROAs: []ValidatingROA{
					{
						Origin:    "15169",
						Prefix:    "8.8.8.0/24",
						MaxLength: 24,
						Validity:  "invalid_asn",
					},
				},
				FetchedAt: "2025-06-23T20:42:11.586624",
			},
			wantErr: false,
		},
		{
			name:     "empty validating_roas",
			resource: "65000",
			prefix:   "192.0.2.0/24",
			serverResponse: `{
				"status": "ok",
				"time": "2025-06-23T20:42:05.326049",
				"data": {
					"validating_roas": null,
					"status": "unknown",
					"validator": "routinator",
					"resource": "65000",
					"prefix": "192.0.2.0/24"
				}
			}`,
			serverStatus: http.StatusOK,
			want: &APIResponse{
				Status:         "unknown",
				Validator:      "routinator",
				Resource:       "65000",
				Prefix:         "192.0.2.0/24",
				ValidatingROAs: []ValidatingROA{},
				FetchedAt:      "2025-06-23T20:42:05.326049",
			},
			wantErr: false,
		},
		{
			name:         "missing resource parameter",
			resource:     "",
			prefix:       "193.0.0.0/21",
			wantErr:      true,
			serverStatus: http.StatusOK,
		},
		{
			name:         "missing prefix parameter",
			resource:     "3333",
			prefix:       "",
			wantErr:      true,
			serverStatus: http.StatusOK,
		},
		{
			name:         "server error",
			resource:     "3333",
			prefix:       "193.0.0.0/21",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverStatus != http.StatusOK {
					w.WriteHeader(tt.serverStatus)
					return
				}

				// Verify query parameters
				if tt.resource != "" && r.URL.Query().Get("resource") != tt.resource {
					t.Errorf("Expected resource %s, got %s", tt.resource, r.URL.Query().Get("resource"))
				}
				if tt.prefix != "" && r.URL.Query().Get("prefix") != tt.prefix {
					t.Errorf("Expected prefix %s, got %s", tt.prefix, r.URL.Query().Get("prefix"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					_, _ = w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			c := NewClient(client.New(server.URL, &http.Client{}))
			got, err := c.Get(context.Background(), tt.resource, tt.prefix)

			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got == nil {
				t.Errorf("Client.Get() returned nil response")
				return
			}

			// Compare the results
			if got.Status != tt.want.Status {
				t.Errorf("Client.Get() Status = %v, want %v", got.Status, tt.want.Status)
			}
			if got.Validator != tt.want.Validator {
				t.Errorf("Client.Get() Validator = %v, want %v", got.Validator, tt.want.Validator)
			}
			if got.Resource != tt.want.Resource {
				t.Errorf("Client.Get() Resource = %v, want %v", got.Resource, tt.want.Resource)
			}
			if got.Prefix != tt.want.Prefix {
				t.Errorf("Client.Get() Prefix = %v, want %v", got.Prefix, tt.want.Prefix)
			}
			if got.FetchedAt != tt.want.FetchedAt {
				t.Errorf("Client.Get() FetchedAt = %v, want %v", got.FetchedAt, tt.want.FetchedAt)
			}

			// Compare ValidatingROAs
			if len(got.ValidatingROAs) != len(tt.want.ValidatingROAs) {
				t.Errorf("Client.Get() ValidatingROAs length = %v, want %v", len(got.ValidatingROAs), len(tt.want.ValidatingROAs))
			} else {
				for i, roa := range got.ValidatingROAs {
					wantROA := tt.want.ValidatingROAs[i]
					if roa.Origin != wantROA.Origin {
						t.Errorf("Client.Get() ValidatingROAs[%d].Origin = %v, want %v", i, roa.Origin, wantROA.Origin)
					}
					if roa.Prefix != wantROA.Prefix {
						t.Errorf("Client.Get() ValidatingROAs[%d].Prefix = %v, want %v", i, roa.Prefix, wantROA.Prefix)
					}
					if roa.MaxLength != wantROA.MaxLength {
						t.Errorf("Client.Get() ValidatingROAs[%d].MaxLength = %v, want %v", i, roa.MaxLength, wantROA.MaxLength)
					}
					if roa.Validity != wantROA.Validity {
						t.Errorf("Client.Get() ValidatingROAs[%d].Validity = %v, want %v", i, roa.Validity, wantROA.Validity)
					}
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	t.Run("with nil client", func(t *testing.T) {
		c := NewClient(nil)
		if c == nil {
			t.Error("NewClient() returned nil")
			return
		}
		if c.client == nil {
			t.Error("NewClient() client field is nil")
		}
	})

	t.Run("with existing client", func(t *testing.T) {
		existingClient := client.New("http://example.com", &http.Client{})
		c := NewClient(existingClient)
		if c == nil {
			t.Error("NewClient() returned nil")
			return
		}
		if c.client != existingClient {
			t.Error("NewClient() did not use provided client")
		}
	})
}

func TestDefaultClient(t *testing.T) {
	c := DefaultClient()
	if c == nil {
		t.Error("DefaultClient() returned nil")
		return
	}
	if c.client == nil {
		t.Error("DefaultClient() client field is nil")
	}
}

func TestGetRPKIValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := Response{
			Data: Data{
				Status:    "valid",
				Validator: "routinator",
				Resource:  "3333",
				Prefix:    "193.0.0.0/21",
				ValidatingROAs: []ValidatingROA{
					{
						Origin:    "3333",
						Prefix:    "193.0.0.0/21",
						MaxLength: 21,
						Validity:  "valid",
					},
				},
			},
		}
		response.Time = "2025-06-23T20:42:05.326049"

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a custom client that uses our test server
	customClient := client.New(server.URL, server.Client())
	rpkiClient := NewClient(customClient)

	// Test the client directly
	ctx := context.Background()
	got, err := rpkiClient.Get(ctx, "3333", "193.0.0.0/21")
	if err != nil {
		t.Errorf("GetRPKIValidation() error = %v", err)
		return
	}

	if got.Status != "valid" {
		t.Errorf("GetRPKIValidation() Status = %v, want %v", got.Status, "valid")
	}
	if got.Resource != "3333" {
		t.Errorf("GetRPKIValidation() Resource = %v, want %v", got.Resource, "3333")
	}
	if got.Prefix != "193.0.0.0/21" {
		t.Errorf("GetRPKIValidation() Prefix = %v, want %v", got.Prefix, "193.0.0.0/21")
	}
}
