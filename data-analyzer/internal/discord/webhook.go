package discord

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	json "github.com/goccy/go-json"
)

const (
	// Colors for Discord embeds
	colorRed   = 15158332 // 0xE74C3C - for errors/expiration
	colorGreen = 5763719  // 0x57F287 - for success

	// Default timeout for webhook requests
	defaultWebhookTimeout = 10 * time.Second

	// Max retries for rate limiting
	maxRetries = 3
)

// WebhookPayload represents a Discord webhook message
type WebhookPayload struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

// Embed represents a Discord embed
type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
}

// EmbedField represents a field in a Discord embed
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// EmbedFooter represents the footer of a Discord embed
type EmbedFooter struct {
	Text string `json:"text"`
}

// NewKeyExpiredPayload creates a payload for API key expiration notification
func NewKeyExpiredPayload(matchesCollected int, runtime time.Duration, lastReduceAgo time.Duration) WebhookPayload {
	return WebhookPayload{
		Content: "@here API Key Expired!",
		Embeds: []Embed{
			{
				Title: "🔑 API Key Expired",
				Color: colorRed,
				Fields: []EmbedField{
					{
						Name:   "Matches Collected",
						Value:  formatNumber(matchesCollected),
						Inline: true,
					},
					{
						Name:   "Runtime",
						Value:  formatDuration(runtime),
						Inline: true,
					},
					{
						Name:   "Last Reduce",
						Value:  formatDurationAgo(lastReduceAgo),
						Inline: true,
					},
				},
				Footer: &EmbedFooter{
					Text: "Reply with new RGAPI-xxx key to start fresh session",
				},
			},
		},
	}
}

// NewSessionStartedPayload creates a payload for new session started notification
func NewSessionStartedPayload(apiKey string, seedPlayer string) WebhookPayload {
	return WebhookPayload{
		Embeds: []Embed{
			{
				Title: "✅ New Session Started",
				Color: colorGreen,
				Fields: []EmbedField{
					{
						Name:   "New Key",
						Value:  maskAPIKey(apiKey) + " (validated)",
						Inline: true,
					},
					{
						Name:   "Seed Player",
						Value:  seedPlayer,
						Inline: true,
					},
				},
				Footer: &EmbedFooter{
					Text: "Fresh crawl beginning from top of ladder",
				},
			},
		},
	}
}

// WebhookClient sends notifications to Discord webhooks
type WebhookClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewWebhookClient creates a new WebhookClient
func NewWebhookClient(webhookURL string) *WebhookClient {
	return &WebhookClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: defaultWebhookTimeout,
		},
	}
}

// SendKeyExpiredNotification sends a key expiration notification
func (c *WebhookClient) SendKeyExpiredNotification(ctx context.Context, matchesCollected int, runtime time.Duration, lastReduceAgo time.Duration) error {
	payload := NewKeyExpiredPayload(matchesCollected, runtime, lastReduceAgo)
	return c.sendPayload(ctx, payload)
}

// SendNewSessionNotification sends a new session started notification
func (c *WebhookClient) SendNewSessionNotification(ctx context.Context, apiKey string, seedPlayer string) error {
	payload := NewSessionStartedPayload(apiKey, seedPlayer)
	return c.sendPayload(ctx, payload)
}

// sendPayload sends a webhook payload with retry on rate limiting
func (c *WebhookClient) sendPayload(ctx context.Context, payload WebhookPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", c.webhookURL, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		resp.Body.Close()

		// Success - Discord returns 204 No Content
		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			return nil
		}

		// Rate limited - wait and retry
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			waitDuration := time.Second // Default wait
			if retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					waitDuration = time.Duration(seconds) * time.Second
				}
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDuration):
				continue
			}
		}

		// Other error
		return fmt.Errorf("webhook request failed with status %d", resp.StatusCode)
	}

	return fmt.Errorf("webhook request failed after %d retries", maxRetries)
}

// formatNumber formats a number with commas (e.g., 47832 -> "47,832")
func formatNumber(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}

	// Simple comma formatting
	s := strconv.Itoa(n)
	var result bytes.Buffer
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// formatDuration formats a duration as "Xh Ym" (e.g., 18h 32m)
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// formatDurationAgo formats a duration as "X min ago" or "X sec ago"
func formatDurationAgo(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d sec ago", int(d.Seconds()))
	}
	return fmt.Sprintf("%d min ago", int(d.Minutes()))
}

// maskAPIKey masks an API key for display (e.g., "RGAPI-xxxx-xxxx" -> "RGAPI-...xxxx")
func maskAPIKey(key string) string {
	if len(key) <= 10 {
		return "****"
	}
	return key[:5] + "..." + key[len(key)-4:]
}
