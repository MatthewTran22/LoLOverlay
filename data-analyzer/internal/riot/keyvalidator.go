package riot

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	// Default validation endpoint (LoL Status API - lightweight)
	defaultStatusEndpoint = "/lol/status/v4/platform-data"
	defaultValidationURL  = na1BaseURL + defaultStatusEndpoint

	// Default timeout for validation requests
	defaultValidationTimeout = 10 * time.Second
)

// KeyValidator validates Riot API keys by making a test request
type KeyValidator struct {
	httpClient *http.Client
	baseURL    string
}

// KeyValidatorOption configures a KeyValidator
type KeyValidatorOption func(*KeyValidator)

// WithBaseURL sets a custom base URL (useful for testing)
func WithBaseURL(url string) KeyValidatorOption {
	return func(v *KeyValidator) {
		v.baseURL = url
	}
}

// WithTimeout sets a custom timeout for validation requests
func WithTimeout(timeout time.Duration) KeyValidatorOption {
	return func(v *KeyValidator) {
		v.httpClient.Timeout = timeout
	}
}

// NewKeyValidator creates a new KeyValidator with the given options
func NewKeyValidator(opts ...KeyValidatorOption) *KeyValidator {
	v := &KeyValidator{
		httpClient: &http.Client{
			Timeout: defaultValidationTimeout,
		},
		baseURL: na1BaseURL,
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// ValidateKey validates an API key by making a test request to the Riot API.
// Returns:
//   - (true, nil) if the key is valid
//   - (false, nil) if the key is invalid (401/403)
//   - (false, error) if there was a network/server error (key validity unknown)
func (v *KeyValidator) ValidateKey(ctx context.Context, apiKey string) (bool, error) {
	if apiKey == "" {
		return false, fmt.Errorf("API key cannot be empty")
	}

	url := v.baseURL + defaultStatusEndpoint
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Riot-Token", apiKey)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Key is valid
		return true, nil

	case http.StatusUnauthorized, http.StatusForbidden:
		// Key is invalid or expired (401/403)
		return false, nil

	default:
		// Server error or unexpected response - we can't determine if key is valid
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
