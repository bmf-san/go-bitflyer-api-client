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

	// Verify timestamp can be parsed as Unix milliseconds
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		t.Errorf("timestamp is not a valid Unix millisecond: %v", err)
	}

	// Verify timestamp is within 1 minute of current time
	now := time.Now().UnixMilli()
	if timestampInt < now-60000 || timestampInt > now+60000 {
		t.Errorf("timestamp %d is not within 1 minute of current time %d", timestampInt, now)
	}
}
