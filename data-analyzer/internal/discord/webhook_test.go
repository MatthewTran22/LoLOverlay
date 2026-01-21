package discord

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestKeyExpiredPayload_Format tests that the key expired payload matches expected Discord embed format
func TestKeyExpiredPayload_Format(t *testing.T) {
	payload := NewKeyExpiredPayload(47832, 18*time.Hour+32*time.Minute, 2*time.Minute)

	// Verify content has mention
	if !strings.Contains(payload.Content, "@here") {
		t.Error("Expected @here mention in content")
	}

	// Verify we have exactly one embed
	if len(payload.Embeds) != 1 {
		t.Fatalf("Expected 1 embed, got %d", len(payload.Embeds))
	}

	embed := payload.Embeds[0]

	// Verify title
	if !strings.Contains(embed.Title, "API Key Expired") {
		t.Errorf("Expected title to contain 'API Key Expired', got: %s", embed.Title)
	}

	// Verify color is red (15158332 = 0xE74C3C)
	if embed.Color != 15158332 {
		t.Errorf("Expected red color (15158332), got: %d", embed.Color)
	}

	// Verify fields
	if len(embed.Fields) < 3 {
		t.Fatalf("Expected at least 3 fields, got %d", len(embed.Fields))
	}

	// Verify matches collected field
	matchesField := embed.Fields[0]
	if matchesField.Name != "Matches Collected" {
		t.Errorf("Expected first field name 'Matches Collected', got: %s", matchesField.Name)
	}
	if matchesField.Value != "47,832" {
		t.Errorf("Expected matches value '47,832', got: %s", matchesField.Value)
	}
	if !matchesField.Inline {
		t.Error("Expected matches field to be inline")
	}

	// Verify runtime field
	runtimeField := embed.Fields[1]
	if runtimeField.Name != "Runtime" {
		t.Errorf("Expected second field name 'Runtime', got: %s", runtimeField.Name)
	}
	if runtimeField.Value != "18h 32m" {
		t.Errorf("Expected runtime value '18h 32m', got: %s", runtimeField.Value)
	}

	// Verify footer
	if !strings.Contains(embed.Footer.Text, "Reply with new RGAPI") {
		t.Errorf("Expected footer to contain instructions, got: %s", embed.Footer.Text)
	}
}

// TestKeyExpiredPayload_JSON tests that the payload serializes to valid JSON
func TestKeyExpiredPayload_JSON(t *testing.T) {
	payload := NewKeyExpiredPayload(1000, time.Hour, 5*time.Minute)

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Verify it's valid JSON by unmarshaling back
	var parsed WebhookPayload
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if parsed.Content != payload.Content {
		t.Error("Content mismatch after round-trip")
	}
}

// TestNewSessionPayload_Format tests that the new session payload matches expected format
func TestNewSessionPayload_Format(t *testing.T) {
	payload := NewSessionStartedPayload("RGAPI-1234-xxxx", "Challenger #1 Player")

	// Verify no @here for success message (less urgent)
	if strings.Contains(payload.Content, "@here") {
		t.Error("Success message should not have @here mention")
	}

	// Verify we have exactly one embed
	if len(payload.Embeds) != 1 {
		t.Fatalf("Expected 1 embed, got %d", len(payload.Embeds))
	}

	embed := payload.Embeds[0]

	// Verify title
	if !strings.Contains(embed.Title, "New Session Started") {
		t.Errorf("Expected title to contain 'New Session Started', got: %s", embed.Title)
	}

	// Verify color is green (5763719 = 0x57F287)
	if embed.Color != 5763719 {
		t.Errorf("Expected green color (5763719), got: %d", embed.Color)
	}

	// Verify key field is masked
	keyField := embed.Fields[0]
	if !strings.Contains(keyField.Value, "xxxx") {
		t.Errorf("Expected masked key, got: %s", keyField.Value)
	}

	// Verify seed player field
	seedField := embed.Fields[1]
	if seedField.Name != "Seed Player" {
		t.Errorf("Expected 'Seed Player' field, got: %s", seedField.Name)
	}
}

// TestWebhookClient_SendKeyExpired tests the HTTP call for key expired notification
func TestWebhookClient_SendKeyExpired(t *testing.T) {
	var receivedBody []byte
	var receivedContentType string
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent) // Discord returns 204 on success
	}))
	defer server.Close()

	client := NewWebhookClient(server.URL)

	err := client.SendKeyExpiredNotification(context.Background(), 1000, time.Hour, 5*time.Minute)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify HTTP method
	if receivedMethod != "POST" {
		t.Errorf("Expected POST method, got: %s", receivedMethod)
	}

	// Verify content type
	if receivedContentType != "application/json" {
		t.Errorf("Expected application/json content type, got: %s", receivedContentType)
	}

	// Verify body is valid JSON payload
	var payload WebhookPayload
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("Failed to parse sent payload: %v", err)
	}

	if len(payload.Embeds) == 0 {
		t.Error("Expected embeds in payload")
	}
}

// TestWebhookClient_SendNewSession tests the HTTP call for new session notification
func TestWebhookClient_SendNewSession(t *testing.T) {
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewWebhookClient(server.URL)

	err := client.SendNewSessionNotification(context.Background(), "RGAPI-test-key", "TopPlayer")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var payload WebhookPayload
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("Failed to parse sent payload: %v", err)
	}

	// Verify it's a success message (green color)
	if len(payload.Embeds) > 0 && payload.Embeds[0].Color != 5763719 {
		t.Error("Expected green color for success notification")
	}
}

// TestWebhookClient_WebhookError tests handling of webhook errors
func TestWebhookClient_WebhookError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Invalid webhook"}`))
	}))
	defer server.Close()

	client := NewWebhookClient(server.URL)

	err := client.SendKeyExpiredNotification(context.Background(), 1000, time.Hour, time.Minute)

	if err == nil {
		t.Error("Expected error for bad request")
	}
}

// TestWebhookClient_NetworkError tests handling of network errors
func TestWebhookClient_NetworkError(t *testing.T) {
	// Use an invalid URL
	client := NewWebhookClient("http://localhost:1") // Port 1 should be unreachable

	err := client.SendKeyExpiredNotification(context.Background(), 1000, time.Hour, time.Minute)

	if err == nil {
		t.Error("Expected network error")
	}
}

// TestWebhookClient_ContextCancelled tests handling of cancelled context
func TestWebhookClient_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewWebhookClient(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.SendKeyExpiredNotification(ctx, 1000, time.Hour, time.Minute)

	if err == nil {
		t.Error("Expected context cancelled error")
	}
}

// TestWebhookClient_RateLimited tests handling of Discord rate limiting
func TestWebhookClient_RateLimited(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewWebhookClient(server.URL)

	err := client.SendKeyExpiredNotification(context.Background(), 1000, time.Hour, time.Minute)

	// Should succeed after retry
	if err != nil {
		t.Errorf("Expected success after retry, got: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts (1 retry), got: %d", attempts)
	}
}
