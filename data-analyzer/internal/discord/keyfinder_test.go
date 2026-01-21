package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestParseAPIKey tests extracting RGAPI key from message content
func TestParseAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
		found    bool
	}{
		{
			name:     "valid key at start",
			content:  "RGAPI-12345678-abcd-1234-efgh-567890abcdef",
			expected: "RGAPI-12345678-abcd-1234-efgh-567890abcdef",
			found:    true,
		},
		{
			name:     "valid key with text before",
			content:  "Here's the new key: RGAPI-12345678-abcd-1234-efgh-567890abcdef",
			expected: "RGAPI-12345678-abcd-1234-efgh-567890abcdef",
			found:    true,
		},
		{
			name:     "valid key with text after",
			content:  "RGAPI-12345678-abcd-1234-efgh-567890abcdef please use this",
			expected: "RGAPI-12345678-abcd-1234-efgh-567890abcdef",
			found:    true,
		},
		{
			name:     "valid key in middle of text",
			content:  "The key is RGAPI-12345678-abcd-1234-efgh-567890abcdef and it should work",
			expected: "RGAPI-12345678-abcd-1234-efgh-567890abcdef",
			found:    true,
		},
		{
			name:     "no key present",
			content:  "Hello, this is just a regular message",
			expected: "",
			found:    false,
		},
		{
			name:     "partial key (too short)",
			content:  "RGAPI-1234",
			expected: "",
			found:    false,
		},
		{
			name:     "empty message",
			content:  "",
			expected: "",
			found:    false,
		},
		{
			name:     "key with newlines",
			content:  "New key:\nRGAPI-12345678-abcd-1234-efgh-567890abcdef\nEnjoy!",
			expected: "RGAPI-12345678-abcd-1234-efgh-567890abcdef",
			found:    true,
		},
		{
			name:     "multiple keys (returns first)",
			content:  "RGAPI-first-key-1234-5678-abcdefghijkl and RGAPI-second-key-5678-9012-lmnopqrstuvw",
			expected: "RGAPI-first-key-1234-5678-abcdefghijkl",
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, found := ParseAPIKey(tt.content)
			if found != tt.found {
				t.Errorf("Expected found=%v, got=%v", tt.found, found)
			}
			if key != tt.expected {
				t.Errorf("Expected key=%q, got=%q", tt.expected, key)
			}
		})
	}
}

// TestKeyFinder_PollChannel tests polling Discord channel for messages
func TestKeyFinder_PollChannel(t *testing.T) {
	messages := []DiscordMessage{
		{
			ID:        "123456789",
			Content:   "Hello, this is a test",
			Timestamp: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:        "123456790",
			Content:   "Here's the key: RGAPI-12345678-abcd-1234-efgh-567890abcdef",
			Timestamp: time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		if r.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Verify query parameters
		if r.URL.Query().Get("limit") == "" {
			t.Error("Expected limit parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}))
	defer server.Close()

	finder := NewKeyFinder("test-bot-token", "123456789", WithDiscordBaseURL(server.URL))

	ctx := context.Background()
	key, err := finder.PollForKey(ctx, time.Now().Add(-10*time.Minute))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if key != "RGAPI-12345678-abcd-1234-efgh-567890abcdef" {
		t.Errorf("Expected to find API key, got: %s", key)
	}
}

// TestKeyFinder_NoKeyFound tests when no key is in messages
func TestKeyFinder_NoKeyFound(t *testing.T) {
	messages := []DiscordMessage{
		{
			ID:        "123456789",
			Content:   "Hello, this is a test",
			Timestamp: time.Now().Format(time.RFC3339),
		},
		{
			ID:        "123456790",
			Content:   "Another message without a key",
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}))
	defer server.Close()

	finder := NewKeyFinder("test-bot-token", "123456789", WithDiscordBaseURL(server.URL))

	ctx := context.Background()
	key, err := finder.PollForKey(ctx, time.Now().Add(-10*time.Minute))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if key != "" {
		t.Errorf("Expected no key found, got: %s", key)
	}
}

// TestKeyFinder_EmptyChannel tests when channel has no messages
func TestKeyFinder_EmptyChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]DiscordMessage{})
	}))
	defer server.Close()

	finder := NewKeyFinder("test-bot-token", "123456789", WithDiscordBaseURL(server.URL))

	ctx := context.Background()
	key, err := finder.PollForKey(ctx, time.Now().Add(-10*time.Minute))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if key != "" {
		t.Errorf("Expected no key found, got: %s", key)
	}
}

// TestKeyFinder_FiltersByTimestamp tests that only messages after timestamp are considered
func TestKeyFinder_FiltersByTimestamp(t *testing.T) {
	now := time.Now()
	messages := []DiscordMessage{
		{
			ID:        "123456789",
			Content:   "Old key: RGAPI-old-key-1234-5678-abcdefghijkl",
			Timestamp: now.Add(-1 * time.Hour).Format(time.RFC3339), // Old message
		},
		{
			ID:        "123456790",
			Content:   "New key: RGAPI-new-key-1234-5678-abcdefghijkl",
			Timestamp: now.Add(-1 * time.Minute).Format(time.RFC3339), // Recent message
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}))
	defer server.Close()

	finder := NewKeyFinder("test-bot-token", "123456789", WithDiscordBaseURL(server.URL))

	// Only look for messages after 30 minutes ago
	ctx := context.Background()
	key, err := finder.PollForKey(ctx, now.Add(-30*time.Minute))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should find the new key, not the old one
	if key != "RGAPI-new-key-1234-5678-abcdefghijkl" {
		t.Errorf("Expected new key, got: %s", key)
	}
}

// TestKeyFinder_NetworkError tests handling of network errors
func TestKeyFinder_NetworkError(t *testing.T) {
	finder := NewKeyFinder("test-bot-token", "123456789", WithDiscordBaseURL("http://localhost:1"))

	ctx := context.Background()
	_, err := finder.PollForKey(ctx, time.Now())

	if err == nil {
		t.Error("Expected network error")
	}
}

// TestKeyFinder_ContextCancelled tests handling of cancelled context
func TestKeyFinder_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]DiscordMessage{})
	}))
	defer server.Close()

	finder := NewKeyFinder("test-bot-token", "123456789", WithDiscordBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := finder.PollForKey(ctx, time.Now())

	if err == nil {
		t.Error("Expected context cancelled error")
	}
}

// TestKeyFinder_RateLimited tests handling of Discord rate limiting
func TestKeyFinder_RateLimited(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]DiscordMessage{
			{
				ID:        "123",
				Content:   "RGAPI-12345678-abcd-1234-efgh-567890abcdef",
				Timestamp: time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	finder := NewKeyFinder("test-bot-token", "123456789", WithDiscordBaseURL(server.URL))

	ctx := context.Background()
	key, err := finder.PollForKey(ctx, time.Now().Add(-10*time.Minute))

	if err != nil {
		t.Fatalf("Expected success after retry, got: %v", err)
	}

	if key == "" {
		t.Error("Expected to find key after retry")
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got: %d", attempts)
	}
}

// TestKeyFinder_WaitForKey tests the blocking wait for key functionality
func TestKeyFinder_WaitForKey(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		// Return key on second call
		if callCount >= 2 {
			json.NewEncoder(w).Encode([]DiscordMessage{
				{
					ID:        "123",
					Content:   "RGAPI-12345678-abcd-1234-efgh-567890abcdef",
					Timestamp: time.Now().Format(time.RFC3339),
				},
			})
		} else {
			json.NewEncoder(w).Encode([]DiscordMessage{})
		}
	}))
	defer server.Close()

	finder := NewKeyFinder("test-bot-token", "123456789",
		WithDiscordBaseURL(server.URL),
		WithPollInterval(100*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key, err := finder.WaitForKey(ctx, time.Now().Add(-10*time.Minute))

	if err != nil {
		t.Fatalf("Expected to find key, got error: %v", err)
	}

	if key != "RGAPI-12345678-abcd-1234-efgh-567890abcdef" {
		t.Errorf("Expected API key, got: %s", key)
	}

	if callCount < 2 {
		t.Errorf("Expected at least 2 poll attempts, got: %d", callCount)
	}
}

// TestKeyFinder_WaitForKeyTimeout tests timeout while waiting for key
func TestKeyFinder_WaitForKeyTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]DiscordMessage{}) // Always empty
	}))
	defer server.Close()

	finder := NewKeyFinder("test-bot-token", "123456789",
		WithDiscordBaseURL(server.URL),
		WithPollInterval(50*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := finder.WaitForKey(ctx, time.Now())

	if err == nil {
		t.Error("Expected timeout error")
	}
}
