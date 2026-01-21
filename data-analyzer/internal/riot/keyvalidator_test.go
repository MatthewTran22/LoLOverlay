package riot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestValidateKey_ValidKey tests that a valid API key passes validation
func TestValidateKey_ValidKey(t *testing.T) {
	// Create mock server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the API key header is set
		if r.Header.Get("X-Riot-Token") == "" {
			t.Error("Expected X-Riot-Token header to be set")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"NA1","name":"North America","locales":["en_US"]}`))
	}))
	defer server.Close()

	validator := NewKeyValidator(WithBaseURL(server.URL))

	valid, err := validator.ValidateKey(context.Background(), "RGAPI-test-key")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !valid {
		t.Error("Expected key to be valid")
	}
}

// TestValidateKey_InvalidKey tests that an invalid/expired API key fails validation
func TestValidateKey_InvalidKey(t *testing.T) {
	// Create mock server that returns 403 Forbidden
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"status":{"message":"Forbidden","status_code":403}}`))
	}))
	defer server.Close()

	validator := NewKeyValidator(WithBaseURL(server.URL))

	valid, err := validator.ValidateKey(context.Background(), "RGAPI-expired-key")

	if err != nil {
		t.Errorf("Expected no error for invalid key, got: %v", err)
	}
	if valid {
		t.Error("Expected key to be invalid")
	}
}

// TestValidateKey_Unauthorized tests that 401 response marks key as invalid
func TestValidateKey_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":{"message":"Unauthorized","status_code":401}}`))
	}))
	defer server.Close()

	validator := NewKeyValidator(WithBaseURL(server.URL))

	valid, err := validator.ValidateKey(context.Background(), "RGAPI-bad-key")

	if err != nil {
		t.Errorf("Expected no error for unauthorized key, got: %v", err)
	}
	if valid {
		t.Error("Expected key to be invalid for 401 response")
	}
}

// TestValidateKey_NetworkError tests that network errors return an error (not invalid)
func TestValidateKey_NetworkError(t *testing.T) {
	// Create a server that immediately closes connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close the connection without responding
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	validator := NewKeyValidator(WithBaseURL(server.URL))

	valid, err := validator.ValidateKey(context.Background(), "RGAPI-test-key")

	// Network errors should return an error, not just invalid
	if err == nil {
		t.Error("Expected network error to be returned")
	}
	if valid {
		t.Error("Expected key to not be valid on network error")
	}
}

// TestValidateKey_Timeout tests that timeouts return an error
func TestValidateKey_Timeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	validator := NewKeyValidator(
		WithBaseURL(server.URL),
		WithTimeout(100*time.Millisecond),
	)

	ctx := context.Background()
	valid, err := validator.ValidateKey(ctx, "RGAPI-test-key")

	// Timeout should return an error
	if err == nil {
		t.Error("Expected timeout error to be returned")
	}
	if valid {
		t.Error("Expected key to not be valid on timeout")
	}
}

// TestValidateKey_EmptyKey tests that empty key returns error
func TestValidateKey_EmptyKey(t *testing.T) {
	validator := NewKeyValidator()

	valid, err := validator.ValidateKey(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty key")
	}
	if valid {
		t.Error("Expected empty key to be invalid")
	}
}

// TestValidateKey_ServerError tests that 5xx errors return an error (not invalid)
func TestValidateKey_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":{"message":"Internal Server Error","status_code":500}}`))
	}))
	defer server.Close()

	validator := NewKeyValidator(WithBaseURL(server.URL))

	valid, err := validator.ValidateKey(context.Background(), "RGAPI-test-key")

	// Server errors should return an error (we don't know if key is valid)
	if err == nil {
		t.Error("Expected server error to be returned")
	}
	if valid {
		t.Error("Expected key to not be valid on server error")
	}
}

// TestValidateKey_ContextCancelled tests that cancelled context is handled
func TestValidateKey_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	validator := NewKeyValidator(WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	valid, err := validator.ValidateKey(ctx, "RGAPI-test-key")

	if err == nil {
		t.Error("Expected context cancelled error")
	}
	if valid {
		t.Error("Expected key to not be valid on cancelled context")
	}
}
