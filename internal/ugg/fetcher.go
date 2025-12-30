package ugg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	patchesURL   = "https://static.bigbrain.gg/assets/lol/riot_patch_update/prod/ugg/patches.json"
	statsBaseURL = "https://stats2.u.gg/lol"
	apiVersion   = "1.5"
	statsVersion = "1.5.0"
)

// Fetcher handles U.GG data fetching
type Fetcher struct {
	client       *http.Client
	currentPatch string
	mu           sync.RWMutex
	cache        map[string]*BuildData
	matchupCache map[string][]MatchupData
}

// NewFetcher creates a new U.GG fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:        make(map[string]*BuildData),
		matchupCache: make(map[string][]MatchupData),
	}
}

// FetchPatch fetches the current patch version from U.GG
func (f *Fetcher) FetchPatch() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	req, err := http.NewRequest("GET", patchesURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch patches: %w", err)
	}
	defer resp.Body.Close()

	var patches []string
	if err := json.NewDecoder(resp.Body).Decode(&patches); err != nil {
		return fmt.Errorf("failed to parse patches: %w", err)
	}

	if len(patches) == 0 {
		return fmt.Errorf("no patches available")
	}

	f.currentPatch = patches[0]
	fmt.Printf("U.GG: Current patch is %s\n", f.currentPatch)
	return nil
}

// GetPatch returns the current patch
func (f *Fetcher) GetPatch() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.currentPatch
}

// ClearCache clears the cached data
func (f *Fetcher) ClearCache() {
	f.mu.Lock()
	f.cache = make(map[string]*BuildData)
	f.matchupCache = make(map[string][]MatchupData)
	f.mu.Unlock()
}

// roleToID converts role name to U.GG role ID
func roleToID(role string) string {
	roleMap := map[string]string{
		"top":     "4",
		"jungle":  "1",
		"middle":  "5",
		"bottom":  "3",
		"utility": "2",
		"":        "",
	}
	return roleMap[role]
}

// getWinsAndGames extracts wins and games from stats data
func (f *Fetcher) getWinsAndGames(data json.RawMessage) (float64, float64) {
	var stats []float64
	if err := json.Unmarshal(data, &stats); err != nil || len(stats) < 2 {
		return 0, 0
	}
	return stats[0], stats[1]
}

// idToRole converts U.GG role ID to role name
func idToRole(id string) string {
	roleMap := map[string]string{
		"4": "top",
		"1": "jungle",
		"5": "middle",
		"3": "bottom",
		"2": "utility",
	}
	return roleMap[id]
}

// GetChampionPrimaryRole returns the most played role for a champion
func (f *Fetcher) GetChampionPrimaryRole(championID int) (string, error) {
	f.mu.RLock()
	patch := f.currentPatch
	f.mu.RUnlock()

	if patch == "" {
		if err := f.FetchPatch(); err != nil {
			return "", err
		}
		f.mu.RLock()
		patch = f.currentPatch
		f.mu.RUnlock()
	}

	url := fmt.Sprintf("%s/%s/overview/%s/ranked_solo_5x5/%d/%s.json",
		statsBaseURL, apiVersion, patch, championID, statsVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("U.GG returned status %d", resp.StatusCode)
	}

	var rawData map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return "", err
	}

	// Count games per role across all regions
	roleGames := make(map[string]float64)
	roles := []string{"1", "2", "3", "4", "5"} // jungle, utility, bottom, top, middle

	for _, regionData := range rawData {
		var regionMap map[string]json.RawMessage
		if err := json.Unmarshal(regionData, &regionMap); err != nil {
			continue
		}

		// First level is tier - get Diamond+ (tier "3")
		tierData, ok := regionMap["3"]
		if !ok {
			continue
		}

		var tierMap map[string]json.RawMessage
		if err := json.Unmarshal(tierData, &tierMap); err != nil {
			continue
		}

		// Now iterate through roles
		for _, roleID := range roles {
			roleData, ok := tierMap[roleID]
			if !ok {
				continue
			}

			var roleContent []json.RawMessage
			if err := json.Unmarshal(roleData, &roleContent); err != nil || len(roleContent) == 0 {
				continue
			}

			var statsData []json.RawMessage
			if err := json.Unmarshal(roleContent[0], &statsData); err != nil {
				continue
			}

			if len(statsData) > 6 {
				_, games := f.getWinsAndGames(statsData[6])
				roleGames[roleID] += games
			}
		}
	}

	// Find role with most games
	var bestRole string
	var bestGames float64
	for roleID, games := range roleGames {
		if games > bestGames {
			bestGames = games
			bestRole = roleID
		}
	}

	if bestRole == "" {
		return "", fmt.Errorf("no role data for champion %d", championID)
	}

	return idToRole(bestRole), nil
}
