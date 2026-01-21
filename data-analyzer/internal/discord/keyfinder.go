package discord

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	json "github.com/goccy/go-json"
)

const (
	// Discord API base URL
	defaultDiscordBaseURL = "https://discord.com/api/v10"

	// Default poll interval for waiting for key
	defaultPollInterval = 10 * time.Second

	// Default timeout for Discord API requests
	defaultDiscordTimeout = 10 * time.Second

	// Number of messages to fetch per poll
	defaultMessageLimit = 3
)

// Regex pattern to match Riot API keys
// Format: RGAPI-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars after RGAPI-)
var apiKeyPattern = regexp.MustCompile(`RGAPI-[a-zA-Z0-9-]{20,50}`)

// DiscordMessage represents a message from the Discord API
type DiscordMessage struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Author    struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"author"`
}

// KeyFinder polls a Discord channel for API key messages
type KeyFinder struct {
	botToken     string
	channelID    string
	baseURL      string
	pollInterval time.Duration
	httpClient   *http.Client
}

// KeyFinderOption configures a KeyFinder
type KeyFinderOption func(*KeyFinder)

// WithDiscordBaseURL sets a custom Discord API base URL (for testing)
func WithDiscordBaseURL(url string) KeyFinderOption {
	return func(f *KeyFinder) {
		f.baseURL = url
	}
}

// WithPollInterval sets the polling interval for WaitForKey
func WithPollInterval(interval time.Duration) KeyFinderOption {
	return func(f *KeyFinder) {
		f.pollInterval = interval
	}
}

// NewKeyFinder creates a new KeyFinder
func NewKeyFinder(botToken, channelID string, opts ...KeyFinderOption) *KeyFinder {
	f := &KeyFinder{
		botToken:     botToken,
		channelID:    channelID,
		baseURL:      defaultDiscordBaseURL,
		pollInterval: defaultPollInterval,
		httpClient: &http.Client{
			Timeout: defaultDiscordTimeout,
		},
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// ParseAPIKey extracts a Riot API key from message content
// Returns the key and whether a key was found
func ParseAPIKey(content string) (string, bool) {
	match := apiKeyPattern.FindString(content)
	if match == "" {
		// Debug: log what we're trying to match
		if len(content) > 0 && len(content) < 100 {
			log.Printf("[KeyFinder] No match in: %q", content)
		}
		return "", false
	}
	return match, true
}

// PollForKey checks the Discord channel for a new API key
// Returns the key if found, empty string if not found, or error on failure
func (f *KeyFinder) PollForKey(ctx context.Context, since time.Time) (string, error) {
	messages, err := f.fetchMessages(ctx)
	if err != nil {
		return "", err
	}

	log.Printf("[KeyFinder] Fetched %d messages:", len(messages))
	for i, msg := range messages {
		log.Printf("[KeyFinder]   %d. %s: %q", i+1, msg.Author.Username, msg.Content)
	}

	// Look for API key in any message (most recent first)
	for _, msg := range messages {
		key, found := ParseAPIKey(msg.Content)
		if found {
			log.Printf("[KeyFinder] Found key in message from %s", msg.Author.Username)
			return key, nil
		}
	}

	return "", nil
}

// WaitForKey polls the channel until a key is found or context is cancelled
func (f *KeyFinder) WaitForKey(ctx context.Context, since time.Time) (string, error) {
	log.Printf("[KeyFinder] Starting to poll channel %s for new API key (checking every %v)", f.channelID, f.pollInterval)

	for {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		key, err := f.PollForKey(ctx, since)
		if err != nil {
			log.Printf("[KeyFinder] Error polling: %v", err)
		}
		if key != "" {
			log.Printf("[KeyFinder] Found API key!")
			return key, nil
		}

		log.Printf("[KeyFinder] No key found, sleeping %v...", f.pollInterval)
		time.Sleep(f.pollInterval)
	}
}

// SendMessage sends a message to the Discord channel
func (f *KeyFinder) SendMessage(ctx context.Context, content string) error {
	url := fmt.Sprintf("%s/channels/%s/messages", f.baseURL, f.channelID)

	payload := map[string]string{"content": content}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+f.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Discord API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendEmbed sends an embed message to the Discord channel
func (f *KeyFinder) SendEmbed(ctx context.Context, payload WebhookPayload) error {
	url := fmt.Sprintf("%s/channels/%s/messages", f.baseURL, f.channelID)

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+f.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Discord API returned status %d", resp.StatusCode)
	}

	return nil
}

// fetchMessages fetches recent messages from the Discord channel
func (f *KeyFinder) fetchMessages(ctx context.Context) ([]DiscordMessage, error) {
	url := fmt.Sprintf("%s/channels/%s/messages?limit=%d", f.baseURL, f.channelID, defaultMessageLimit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+f.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		waitDuration := time.Second
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				waitDuration = time.Duration(seconds) * time.Second
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(waitDuration):
			return f.fetchMessages(ctx) // Retry
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Discord API returned status %d", resp.StatusCode)
	}

	var messages []DiscordMessage
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return messages, nil
}
