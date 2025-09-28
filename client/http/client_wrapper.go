package http

import (
	"net/http"

	"github.com/bmf-san/go-bitflyer-api-client/v1/client/auth"
)

// AuthenticatedClient wraps the generated client with authentication
type AuthenticatedClient struct {
	client *ClientWithResponses
	signer *auth.Signer
}

// AuthOption is a function that modifies the authenticated client
type AuthOption func(*AuthenticatedClient)

// WithCustomHTTPClient sets a custom HTTP client
func WithCustomHTTPClient(httpClient *http.Client) AuthOption {
	return func(c *AuthenticatedClient) {
		customClient := &http.Client{
			Transport: &authenticatedTransport{
				base:   httpClient.Transport,
				signer: c.signer,
			},
			Timeout: httpClient.Timeout,
		}

		var err error
		c.client, err = NewClientWithResponses("https://api.bitflyer.com", WithHTTPClient(customClient))
		if err != nil {
			panic(err) // TODO: handle this better
		}
	}
}

// NewAuthenticatedClient creates a new authenticated API client
func NewAuthenticatedClient(credentials auth.APICredentials, baseURL string, opts ...AuthOption) (*AuthenticatedClient, error) {
	if baseURL == "" {
		baseURL = "https://api.bitflyer.com"
	}

	ac := &AuthenticatedClient{
		signer: auth.NewSigner(credentials),
	}

	// Create transport with authentication
	transport := &authenticatedTransport{
		base:   http.DefaultTransport,
		signer: ac.signer,
	}

	// Create client with auth transport
	httpClient := &http.Client{
		Transport: transport,
	}

	client, err := NewClientWithResponses(baseURL, WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	ac.client = client

	// Apply additional options
	for _, opt := range opts {
		opt(ac)
	}

	return ac, nil
}

// Client returns the underlying generated client
func (c *AuthenticatedClient) Client() *ClientWithResponses {
	return c.client
}

// authenticatedTransport is an http.RoundTripper that adds authentication
type authenticatedTransport struct {
	base   http.RoundTripper
	signer *auth.Signer
}

// RoundTrip implements http.RoundTripper
func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add authentication headers
	if err := t.signer.Sign(req); err != nil {
		return nil, err
	}

	// Use the base transport to perform the request
	if t.base == nil {
		t.base = http.DefaultTransport
	}
	return t.base.RoundTrip(req)
}
