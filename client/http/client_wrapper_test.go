package http

import (
	"net/http"
	"testing"
	"time"

	"go-bitflyer-api-client/client/auth"
)

func TestNewAuthenticatedClient(t *testing.T) {
	tests := []struct {
		name        string
		credentials auth.APICredentials
		baseURL     string
		wantErr     bool
	}{
		{
			name: "success with default URL",
			credentials: auth.APICredentials{
				APIKey:    "test-key",
				APISecret: "test-secret",
			},
			baseURL: "",
			wantErr: false,
		},
		{
			name: "success with custom URL",
			credentials: auth.APICredentials{
				APIKey:    "test-key",
				APISecret: "test-secret",
			},
			baseURL: "https://api.custom.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAuthenticatedClient(tt.credentials, tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuthenticatedClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if client == nil {
				t.Error("NewAuthenticatedClient() returned nil client")
				return
			}

			// Check basic client configuration
			if client.signer == nil {
				t.Error("signer is nil")
			}
			if client.client == nil {
				t.Error("client is nil")
			}
		})
	}
}

func TestWithCustomHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	credentials := auth.APICredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	client, err := NewAuthenticatedClient(credentials, "", WithCustomHTTPClient(customClient))
	if err != nil {
		t.Fatalf("NewAuthenticatedClient() error = %v", err)
	}

	// Check if client is generated correctly
	if client.client == nil {
		t.Error("client is nil")
	}

	// Check if custom client is set correctly
	if client.Client() == nil {
		t.Error("Client() returned nil")
	}
}

func TestAuthenticatedTransport_RoundTrip(t *testing.T) {
	credentials := auth.APICredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	transport := &authenticatedTransport{
		base:   http.DefaultTransport,
		signer: auth.NewSigner(credentials),
	}

	req, err := http.NewRequest("GET", "https://api.bitflyer.com/v1/me/getbalance", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, err = transport.RoundTrip(req)
	// Actual API call is not made, so error is expected
	// Error is intentionally ignored as it's expected behavior
	_ = err
	// Here, check if authentication headers are set correctly
	if req.Header.Get("ACCESS-KEY") != credentials.APIKey {
		t.Error("ACCESS-KEY header not set correctly")
	}
	if req.Header.Get("ACCESS-TIMESTAMP") == "" {
		t.Error("ACCESS-TIMESTAMP header not set")
	}
	if req.Header.Get("ACCESS-SIGN") == "" {
		t.Error("ACCESS-SIGN header not set")
	}
}
