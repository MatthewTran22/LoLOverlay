package riot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	// API base URLs
	americasBaseURL = "https://americas.api.riotgames.com"
)

// Client is a Riot API client that handles 429 rate limit responses
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Riot API client
func NewClient() (*Client, error) {
	apiKey := os.Getenv("RIOT_API_KEY")
	if apiKey == "" {
		// Also check alternative env var name
		apiKey = os.Getenv("RIOT-DEV-KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("RIOT_API_KEY or RIOT-DEV-KEY environment variable not set")
	}

	// Show key prefix for debugging (don't show full key)
	if len(apiKey) > 10 {
		fmt.Printf("Using API key: %s...%s\n", apiKey[:8], apiKey[len(apiKey)-4:])
	}

	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// doRequest makes a request and handles 429 rate limit responses
func (c *Client) doRequest(ctx context.Context, url string, result interface{}) error {
	// Sleep to stay under 100 requests/2min rate limit (~1.2s between requests)
	time.Sleep(1300 * time.Millisecond)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Riot-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		// Rate limited - wait 30 seconds and retry
		fmt.Printf("      [429 Rate Limited] Waiting 30 seconds...\n")
		time.Sleep(30 * time.Second)
		return c.doRequest(ctx, url, result)
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("API returned 403 Forbidden - check if your API key is valid")
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("API returned 404 Not Found - player/match may not exist")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// GetAccountByRiotID fetches account info by Riot ID (gameName#tagLine)
func (c *Client) GetAccountByRiotID(ctx context.Context, gameName, tagLine string) (*AccountResponse, error) {
	url := fmt.Sprintf("%s/riot/account/v1/accounts/by-riot-id/%s/%s",
		americasBaseURL, gameName, tagLine)

	var account AccountResponse
	err := c.doRequest(ctx, url, &account)
	return &account, err
}

// GetMatchHistory fetches match IDs for a player
func (c *Client) GetMatchHistory(ctx context.Context, puuid string, count int) ([]string, error) {
	url := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?queue=420&count=%d",
		americasBaseURL, puuid, count)

	var matchIDs []string
	err := c.doRequest(ctx, url, &matchIDs)
	return matchIDs, err
}

// GetMatch fetches match details
func (c *Client) GetMatch(ctx context.Context, matchID string) (*MatchResponse, error) {
	url := fmt.Sprintf("%s/lol/match/v5/matches/%s", americasBaseURL, matchID)

	var match MatchResponse
	err := c.doRequest(ctx, url, &match)
	return &match, err
}

// GetTimeline fetches match timeline
func (c *Client) GetTimeline(ctx context.Context, matchID string) (*TimelineResponse, error) {
	url := fmt.Sprintf("%s/lol/match/v5/matches/%s/timeline", americasBaseURL, matchID)

	var timeline TimelineResponse
	err := c.doRequest(ctx, url, &timeline)
	return &timeline, err
}

// ExtractBuildOrder extracts item purchase order for a participant from timeline
func ExtractBuildOrder(timeline *TimelineResponse, participantID int) []int {
	var buildOrder []int
	seenItems := make(map[int]bool)

	for _, frame := range timeline.Info.Frames {
		for _, event := range frame.Events {
			if event.Type == "ITEM_PURCHASED" && event.ParticipantID == participantID {
				// Only include completed items, skip duplicates
				if IsCompletedItem(event.ItemID) && !seenItems[event.ItemID] {
					buildOrder = append(buildOrder, event.ItemID)
					seenItems[event.ItemID] = true
				}
			}
		}
	}

	return buildOrder
}
