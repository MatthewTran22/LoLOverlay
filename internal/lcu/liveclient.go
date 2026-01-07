package lcu

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LiveClientPlayer represents a player from the live client API
type LiveClientPlayer struct {
	ChampionName    string `json:"championName"`
	IsBot           bool   `json:"isBot"`
	IsDead          bool   `json:"isDead"`
	Items           []LiveClientItem `json:"items"`
	Level           int    `json:"level"`
	Position        string `json:"position"`
	RawChampionName string `json:"rawChampionName"`
	RespawnTimer    float64 `json:"respawnTimer"`
	Scores          LiveClientScores `json:"scores"`
	SummonerName    string `json:"summonerName"`
	Team            string `json:"team"`
}

// LiveClientItem represents an item from the live client API
type LiveClientItem struct {
	CanUse       bool   `json:"canUse"`
	Consumable   bool   `json:"consumable"`
	Count        int    `json:"count"`
	DisplayName  string `json:"displayName"`
	ItemID       int    `json:"itemID"`
	Price        int    `json:"price"`
	RawDescription string `json:"rawDescription"`
	RawDisplayName string `json:"rawDisplayName"`
	Slot         int    `json:"slot"`
}

// LiveClientScores represents player scores
type LiveClientScores struct {
	Assists    int     `json:"assists"`
	CreepScore int     `json:"creepScore"`
	Deaths     int     `json:"deaths"`
	Kills      int     `json:"kills"`
	WardScore  float64 `json:"wardScore"`
}

// LiveClient handles communication with the live client API (localhost:2999)
type LiveClient struct {
	httpClient *http.Client
}

// NewLiveClient creates a new live client
func NewLiveClient() *LiveClient {
	return &LiveClient{
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// GetAllPlayers fetches all players from the live game
func (c *LiveClient) GetAllPlayers() ([]LiveClientPlayer, error) {
	resp, err := c.httpClient.Get("https://127.0.0.1:2999/liveclientdata/playerlist")
	if err != nil {
		return nil, fmt.Errorf("live client not available: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var players []LiveClientPlayer
	if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
		return nil, fmt.Errorf("failed to parse players: %w", err)
	}

	return players, nil
}

// GetActivePlayer returns the active player's summoner name
func (c *LiveClient) GetActivePlayer() (string, error) {
	resp, err := c.httpClient.Get("https://127.0.0.1:2999/liveclientdata/activeplayername")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var name string
	if err := json.NewDecoder(resp.Body).Decode(&name); err != nil {
		return "", err
	}

	return name, nil
}

// IsGameRunning checks if a live game is running
func (c *LiveClient) IsGameRunning() bool {
	resp, err := c.httpClient.Get("https://127.0.0.1:2999/liveclientdata/activeplayername")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}
