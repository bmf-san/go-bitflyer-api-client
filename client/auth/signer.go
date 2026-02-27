package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// APICredentials holds API credentials
type APICredentials struct {
	APIKey    string
	APISecret string
}

// Signer implements request signing for bitFlyer API
type Signer struct {
	credentials APICredentials
}

// NewSigner creates a new signer with the given credentials
func NewSigner(credentials APICredentials) *Signer {
	return &Signer{
		credentials: credentials,
	}
}

// Sign signs the HTTP request with the required authentication headers
func (s *Signer) Sign(req *http.Request) error {
	timestamp := time.Now().Unix() // bitFlyer requires Unix timestamp in SECONDS
	method := req.Method
	// RequestURI includes query string (e.g. /v1/me/getpositions?product_code=FX_BTC_JPY).
	// bitFlyer requires the full path+query in the signed message.
	path := req.URL.RequestURI()
	body := ""

	if req.Body != nil {
		// Read and restore the body
		bodyReadCloser, err := req.GetBody()
		if err != nil {
			return fmt.Errorf("failed to get request body: %w", err)
		}
		defer func() {
			if closeErr := bodyReadCloser.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("failed to close request body: %w", closeErr)
			}
		}()

		bodyBytes, err := io.ReadAll(bodyReadCloser)
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}
		body = string(bodyBytes)
	}

	// Generate signature
	message := fmt.Sprintf("%d%s%s%s", timestamp, method, path, body)
	signature := s.generateHMAC(message)

	// Set authentication headers
	req.Header.Set("ACCESS-KEY", s.credentials.APIKey)
	req.Header.Set("ACCESS-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Set("ACCESS-SIGN", signature)
	req.Header.Set("Content-Type", "application/json")

	return nil
}

// generateHMAC generates HMAC signature using SHA256
func (s *Signer) generateHMAC(message string) string {
	mac := hmac.New(sha256.New, []byte(s.credentials.APISecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
