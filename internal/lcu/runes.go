package lcu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RuneTreeData holds rune tree information
type RuneTreeData struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Slots []struct {
		Runes []struct {
			ID   int    `json:"id"`
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"runes"`
	} `json:"slots"`
}

// RuneRegistry holds rune ID to name mapping
type RuneRegistry struct {
	runes  map[int]string // rune ID -> name
	trees  map[int]string // tree ID -> name
	mu     sync.RWMutex
	loaded bool
}

// NewRuneRegistry creates a new rune registry
func NewRuneRegistry() *RuneRegistry {
	return &RuneRegistry{
		runes: make(map[int]string),
		trees: make(map[int]string),
	}
}

// Load fetches rune data from Data Dragon
func (r *RuneRegistry) Load() error {
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

	version := versions[0]

	// Get rune data
	runeURL := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/runesReforged.json", version)
	runeResp, err := client.Get(runeURL)
	if err != nil {
		return fmt.Errorf("failed to fetch runes: %w", err)
	}
	defer runeResp.Body.Close()

	var runeTrees []RuneTreeData
	if err := json.NewDecoder(runeResp.Body).Decode(&runeTrees); err != nil {
		return fmt.Errorf("failed to parse runes: %w", err)
	}

	// Build ID -> Name maps
	for _, tree := range runeTrees {
		r.trees[tree.ID] = tree.Name
		for _, slot := range tree.Slots {
			for _, rune := range slot.Runes {
				r.runes[rune.ID] = rune.Name
			}
		}
	}

	// Add stat shard names (these aren't in the API)
	r.runes[5008] = "Adaptive Force"
	r.runes[5005] = "Attack Speed"
	r.runes[5007] = "Ability Haste"
	r.runes[5002] = "Armor"
	r.runes[5003] = "Magic Resist"
	r.runes[5001] = "Health Scaling"
	r.runes[5011] = "Health"

	r.loaded = true
	fmt.Printf("Loaded %d runes from Data Dragon (v%s)\n", len(r.runes), version)
	return nil
}

// GetRuneName returns the rune name for a given ID
func (r *RuneRegistry) GetRuneName(id int) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if name, ok := r.runes[id]; ok {
		return name
	}
	return fmt.Sprintf("Rune %d", id)
}

// GetTreeName returns the rune tree name for a given ID
func (r *RuneRegistry) GetTreeName(id int) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if name, ok := r.trees[id]; ok {
		return name
	}
	return fmt.Sprintf("Tree %d", id)
}

// IsLoaded returns whether the registry has been loaded
func (r *RuneRegistry) IsLoaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loaded
}
