package auth

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name        string
		credentials APICredentials
		method      string
		path        string
		body        string
		wantErr     bool
	}{
		{
			name: "valid signature",
			credentials: APICredentials{
				APIKey:    "key123",
				APISecret: "secret123",
			},
			method:  "GET",
			path:    "/v1/me/getbalance",
			body:    "",
			wantErr: false,
		},
		{
			name: "valid signature with POST and body",
			credentials: APICredentials{
				APIKey:    "key123",
				APISecret: "secret123",
			},
			method:  "POST",
			path:    "/v1/me/sendchildorder",
			body:    `{"product_code":"BTC_JPY","child_order_type":"LIMIT","side":"BUY","price":30000,"size":0.1}`,
			wantErr: false,
		},
		{
			name: "valid signature with GET and query params",
			credentials: APICredentials{
				APIKey:    "key123",
				APISecret: "secret123",
			},
			method:  "GET",
			path:    "/v1/me/getpositions?product_code=FX_BTC_JPY",
			body:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer := NewSigner(tt.credentials)

			req, err := http.NewRequest(tt.method, "https://api.bitflyer.com"+tt.path, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			err = signer.Sign(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify signature exists
			accessKey := req.Header.Get("ACCESS-KEY")
			if accessKey != tt.credentials.APIKey {
				t.Errorf("ACCESS-KEY header = %v, want %v", accessKey, tt.credentials.APIKey)
			}

			timestamp := req.Header.Get("ACCESS-TIMESTAMP")
			if timestamp == "" {
				t.Error("ACCESS-TIMESTAMP header is empty")
			}

			signature := req.Header.Get("ACCESS-SIGN")
			if signature == "" {
				t.Error("ACCESS-SIGN header is empty")
			}
		})
	}
}

func TestSignature_TimestampFormat(t *testing.T) {
	signer := NewSigner(APICredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	req, _ := http.NewRequest("GET", "https://api.bitflyer.com/v1/me/getbalance", nil)
	err := signer.Sign(req)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	timestamp := req.Header.Get("ACCESS-TIMESTAMP")
	if timestamp == "" {
		t.Error("ACCESS-TIMESTAMP header is empty")
	}

	// Verify timestamp can be parsed as Unix seconds (not milliseconds)
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		t.Errorf("timestamp is not a valid Unix second: %v", err)
	}

	// Verify timestamp is within 60 seconds of current time
	// bitFlyer requires ACCESS-TIMESTAMP in seconds, not milliseconds
	now := time.Now().Unix()
	if timestampInt < now-60 || timestampInt > now+60 {
		t.Errorf("timestamp %d is not within 60 seconds of current Unix time %d", timestampInt, now)
	}
}

// TestSignature_QueryParamsIncluded verifies that query parameters are included
// in the signed message. Two requests to the same path but different query strings
// must produce different signatures.
func TestSignature_QueryParamsIncluded(t *testing.T) {
	signer := NewSigner(APICredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	req1, _ := http.NewRequest("GET", "https://api.bitflyer.com/v1/me/getpositions?product_code=BTC_JPY", nil)
	req2, _ := http.NewRequest("GET", "https://api.bitflyer.com/v1/me/getpositions?product_code=FX_BTC_JPY", nil)

	if err := signer.Sign(req1); err != nil {
		t.Fatalf("Sign() req1 failed: %v", err)
	}
	// Use same timestamp header to isolate the query string difference
	ts := req1.Header.Get("ACCESS-TIMESTAMP")
	req2.Header.Set("ACCESS-TIMESTAMP", ts)
	if err := signer.Sign(req2); err != nil {
		t.Fatalf("Sign() req2 failed: %v", err)
	}

	sig1 := req1.Header.Get("ACCESS-SIGN")
	sig2 := req2.Header.Get("ACCESS-SIGN")
	if sig1 == sig2 {
		t.Error("Expected different signatures for different query params, but got the same signature")
	}
}
