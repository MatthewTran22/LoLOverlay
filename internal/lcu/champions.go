package lcu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ChampionData holds champion information
type ChampionData struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// ChampionRegistry holds the champion ID to name mapping
type ChampionRegistry struct {
	champions map[int]string // key -> name (key is the numeric ID)
	mu        sync.RWMutex
	loaded    bool
}

// NewChampionRegistry creates a new champion registry
func NewChampionRegistry() *ChampionRegistry {
	return &ChampionRegistry{
		champions: make(map[int]string),
	}
}

// Load fetches champion data from Data Dragon
func (r *ChampionRegistry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	client := &http.Client{Timeout: 10 * time.Second}

	// Get latest version
	versionsResp, err := client.Get("https://ddragon.leagueoflegends.com/api/versions.json")
	if err != nil {
		return fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer versionsResp.Body.Close()

	var versions []string
	if err := json.NewDecoder(versionsResp.Body).Decode(&versions); err != nil {
		return fmt.Errorf("failed to parse versions: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no versions available")
	}

	latestVersion := versions[0]

	// Get champion data
	champURL := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", latestVersion)
	champResp, err := client.Get(champURL)
	if err != nil {
		return fmt.Errorf("failed to fetch champions: %w", err)
	}
	defer champResp.Body.Close()

	var champData struct {
		Data map[string]ChampionData `json:"data"`
	}
	if err := json.NewDecoder(champResp.Body).Decode(&champData); err != nil {
		return fmt.Errorf("failed to parse champions: %w", err)
	}

	// Build ID -> Name map
	for _, champ := range champData.Data {
		key, err := strconv.Atoi(champ.Key)
		if err != nil {
			continue
		}
		r.champions[key] = champ.Name
	}

	r.loaded = true
	fmt.Printf("Loaded %d champions from Data Dragon (v%s)\n", len(r.champions), latestVersion)
	return nil
}

// GetName returns the champion name for a given ID
func (r *ChampionRegistry) GetName(id int) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if name, ok := r.champions[id]; ok {
		return name
	}
	return fmt.Sprintf("Champion %d", id)
}

// IsLoaded returns whether the registry has been loaded
func (r *ChampionRegistry) IsLoaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loaded
}
