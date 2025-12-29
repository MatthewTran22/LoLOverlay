package ugg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	patchesURL     = "https://static.bigbrain.gg/assets/lol/riot_patch_update/prod/ugg/patches.json"
	statsBaseURL   = "https://stats2.u.gg/lol"
	apiVersion     = "1.5"
	statsVersion   = "1.5.0"
)

// BuildData holds champion build information from U.GG
type BuildData struct {
	ChampionID    int      `json:"championId"`
	ChampionName  string   `json:"championName"`
	Role          string   `json:"role"`
	WinRate       float64  `json:"winRate"`
	PickRate      float64  `json:"pickRate"`
	Runes         RuneData `json:"runes"`
	StartingItems []int    `json:"startingItems"`
	CoreItems     []int    `json:"coreItems"`
	Counters      []int    `json:"counters"`
}

// RuneData holds rune information
type RuneData struct {
	PrimaryStyle    int   `json:"primaryStyle"`
	SecondaryStyle  int   `json:"secondaryStyle"`
	PrimaryPerks    []int `json:"primaryPerks"`
	SecondaryPerks  []int `json:"secondaryPerks"`
	StatShards      []int `json:"statShards"`
}

// Fetcher handles U.GG data fetching
type Fetcher struct {
	client       *http.Client
	currentPatch string
	mu           sync.RWMutex
	cache        map[string]*BuildData // key: "championId-role"
}

// NewFetcher creates a new U.GG fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]*BuildData),
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

// FetchChampionData fetches build data for a champion
func (f *Fetcher) FetchChampionData(championID int, championName string, role string) (*BuildData, error) {
	f.mu.RLock()
	patch := f.currentPatch
	f.mu.RUnlock()

	if patch == "" {
		if err := f.FetchPatch(); err != nil {
			return nil, err
		}
		f.mu.RLock()
		patch = f.currentPatch
		f.mu.RUnlock()
	}

	// Check cache
	cacheKey := fmt.Sprintf("%d-%s", championID, role)
	f.mu.RLock()
	if cached, ok := f.cache[cacheKey]; ok {
		f.mu.RUnlock()
		return cached, nil
	}
	f.mu.RUnlock()

	// Build URL: https://stats2.u.gg/lol/1.5/overview/15_24/ranked_solo_5x5/233/1.5.0.json
	url := fmt.Sprintf("%s/%s/overview/%s/ranked_solo_5x5/%d/%s.json",
		statsBaseURL, apiVersion, patch, championID, statsVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch champion data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("U.GG returned status %d", resp.StatusCode)
	}

	// Parse the response - U.GG returns nested data by region/role
	var rawData map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to parse champion data: %w", err)
	}

	// Parse the build data
	buildData, err := f.parseChampionData(rawData, championID, championName, role)
	if err != nil {
		return nil, err
	}

	// Cache it
	f.mu.Lock()
	f.cache[cacheKey] = buildData
	f.mu.Unlock()

	return buildData, nil
}

// parseChampionData extracts build data from U.GG response
func (f *Fetcher) parseChampionData(rawData map[string]json.RawMessage, championID int, championName string, role string) (*BuildData, error) {
	// U.GG data structure: data[regionId][roleId][tierId] = [statsArray, timestamp]
	// We'll use world stats (12) and map role to role ID
	// Role IDs: 1=jungle, 2=supp, 3=adc, 4=top, 5=mid
	// Tier IDs: 1=all ranks, 10=platinum+, etc.

	roleMap := map[string]string{
		"top":     "4",
		"jungle":  "1",
		"middle":  "5",
		"bottom":  "3",
		"utility": "2",
		"":        "",
	}

	roleID := roleMap[role]

	// Aggregate wins/games across all regions for the target role
	var totalWins, totalGames float64
	var bestStatsData []json.RawMessage // Keep the best stats for runes/items
	var bestGames float64

	// Search all regions and aggregate data
	for regionID, regionData := range rawData {
		var regionMap map[string]json.RawMessage
		if err := json.Unmarshal(regionData, &regionMap); err != nil {
			continue
		}

		// Try the specific role first, then try all roles to find any data
		rolesToTry := []string{roleID}
		if roleID == "" {
			rolesToTry = []string{"5", "4", "3", "1", "2"} // mid, top, adc, jungle, support
		}

		for _, tryRole := range rolesToTry {
			if tryRole == "" {
				continue
			}
			roleData, ok := regionMap[tryRole]
			if !ok {
				continue
			}

			var tierMap map[string]json.RawMessage
			if err := json.Unmarshal(roleData, &tierMap); err != nil {
				continue
			}

			// Use tier 3 (main data tier with most games)
			tierData, ok := tierMap["3"]
			if !ok {
				continue
			}

			// The tier data is [statsArray, timestamp]
			var tierContent []json.RawMessage
			if err := json.Unmarshal(tierData, &tierContent); err != nil {
				continue
			}

			if len(tierContent) > 0 {
				var statsData []json.RawMessage
				if err := json.Unmarshal(tierContent[0], &statsData); err != nil {
					continue
				}
				if len(statsData) > 6 {
					// Get wins/games at index 6 [wins, games]
					wins, games := f.getWinsAndGames(statsData[6])
					if games > 0 && tryRole == roleID {
						totalWins += wins
						totalGames += games
						fmt.Printf("Region %s, Role %s: %.0f wins / %.0f games\n", regionID, tryRole, wins, games)

						// Keep the stats with most games for runes/items
						if games > bestGames {
							bestGames = games
							bestStatsData = statsData
						}
					}
				}
			}
			// Only use the target role for aggregation
			if tryRole == roleID {
				break
			}
		}
	}

	// If no data for target role, try to find any role
	if len(bestStatsData) == 0 {
		for _, regionData := range rawData {
			var regionMap map[string]json.RawMessage
			if err := json.Unmarshal(regionData, &regionMap); err != nil {
				continue
			}
			for _, tryRole := range []string{"5", "4", "3", "1", "2"} {
				roleData, ok := regionMap[tryRole]
				if !ok {
					continue
				}
				var tierMap map[string]json.RawMessage
				if err := json.Unmarshal(roleData, &tierMap); err != nil {
					continue
				}
				tierData, ok := tierMap["3"]
				if !ok {
					continue
				}
				var tierContent []json.RawMessage
				if err := json.Unmarshal(tierData, &tierContent); err != nil || len(tierContent) == 0 {
					continue
				}
				var statsData []json.RawMessage
				if err := json.Unmarshal(tierContent[0], &statsData); err != nil {
					continue
				}
				if len(statsData) > 6 {
					wins, games := f.getWinsAndGames(statsData[6])
					if games > bestGames {
						bestGames = games
						bestStatsData = statsData
						totalWins = wins
						totalGames = games
					}
				}
			}
			if len(bestStatsData) > 0 {
				break
			}
		}
	}

	if len(bestStatsData) == 0 {
		return nil, fmt.Errorf("no data found for champion %d", championID)
	}

	statsData := bestStatsData
	fmt.Printf("Aggregated: %.0f wins / %.0f games = %.1f%% WR\n", totalWins, totalGames, (totalWins/totalGames)*100)

	// The statsData is an array with various stats at specific indices
	// Based on U.GG's structure:
	// [0] = general stats [wins, games, ...]
	// [1] = runes
	// [2] = summoner spells
	// [3] = starting items
	// [4] = core items
	// ...etc

	build := &BuildData{
		ChampionID:   championID,
		ChampionName: championName,
		Role:         role,
	}

	// Set aggregated win rate
	if totalGames > 0 {
		build.WinRate = (totalWins / totalGames) * 100
		build.PickRate = totalGames
	}

	// Parse runes (index 0)
	if len(statsData) > 0 {
		f.parseRunes(statsData[0], build)
	}

	// Parse stat shards (index 9)
	if len(statsData) > 9 {
		f.parseStatShards(statsData[9], build)
	}

	// Parse starting items (index 2)
	if len(statsData) > 2 {
		f.parseItems(statsData[2], &build.StartingItems)
	}

	// Parse core items (index 3)
	if len(statsData) > 3 {
		f.parseItems(statsData[3], &build.CoreItems)
	}

	return build, nil
}

// parseRunes extracts rune data
func (f *Fetcher) parseRunes(data json.RawMessage, build *BuildData) {
	// Rune data structure: [[?, ?, primaryTree, secondaryTree, [perks...]], wins, games, ...]
	var runeData []json.RawMessage
	if err := json.Unmarshal(data, &runeData); err != nil || len(runeData) == 0 {
		return
	}

	var runeArray []json.RawMessage
	if err := json.Unmarshal(runeData[0], &runeArray); err != nil || len(runeArray) < 5 {
		return
	}

	var primaryTree, secondaryTree int
	json.Unmarshal(runeArray[2], &primaryTree)
	json.Unmarshal(runeArray[3], &secondaryTree)

	var perks []int
	json.Unmarshal(runeArray[4], &perks)

	if len(perks) >= 6 {
		build.Runes = RuneData{
			PrimaryStyle:   primaryTree,
			SecondaryStyle: secondaryTree,
			PrimaryPerks:   perks[0:4], // keystone + 3 minor runes
			SecondaryPerks: perks[4:6], // 2 secondary runes
			StatShards:     []int{},    // will be filled from stat shards data
		}
	}
}

// parseStatShards extracts stat shard data
func (f *Fetcher) parseStatShards(data json.RawMessage, build *BuildData) {
	// Stat shards structure: [[?, ?, [shard1, shard2, shard3]], ...]
	var shardData []json.RawMessage
	if err := json.Unmarshal(data, &shardData); err != nil || len(shardData) == 0 {
		return
	}

	var shardArray []json.RawMessage
	if err := json.Unmarshal(shardData[0], &shardArray); err != nil || len(shardArray) < 3 {
		return
	}

	var shards []interface{}
	if err := json.Unmarshal(shardArray[2], &shards); err != nil {
		return
	}

	// Convert shards to int (they might be strings)
	for _, s := range shards {
		switch v := s.(type) {
		case float64:
			build.Runes.StatShards = append(build.Runes.StatShards, int(v))
		case string:
			var id int
			fmt.Sscanf(v, "%d", &id)
			build.Runes.StatShards = append(build.Runes.StatShards, id)
		}
	}
}

// parseItems extracts item IDs
func (f *Fetcher) parseItems(data json.RawMessage, items *[]int) {
	// Item data structure: [[matchCount, ?, [item1, item2, ...]], wins, games, ...]
	var itemData []json.RawMessage
	if err := json.Unmarshal(data, &itemData); err != nil || len(itemData) == 0 {
		return
	}

	var itemArray []json.RawMessage
	if err := json.Unmarshal(itemData[0], &itemArray); err != nil || len(itemArray) < 3 {
		return
	}

	var itemIDs []int
	if err := json.Unmarshal(itemArray[2], &itemIDs); err != nil {
		return
	}

	*items = itemIDs
}

// parseWinRate extracts win rate
func (f *Fetcher) parseWinRate(data json.RawMessage, build *BuildData) {
	// Win rate data structure: [wins, games]
	var stats []float64
	if err := json.Unmarshal(data, &stats); err != nil || len(stats) < 2 {
		return
	}

	wins := stats[0]
	games := stats[1]

	if games > 0 {
		build.WinRate = (wins / games) * 100
		build.PickRate = games
	}
	fmt.Printf("Win rate parsed: %.0f wins / %.0f games = %.1f%%\n", wins, games, build.WinRate)
}

// ClearCache clears the cached data
func (f *Fetcher) ClearCache() {
	f.mu.Lock()
	f.cache = make(map[string]*BuildData)
	f.mu.Unlock()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getGamesCount extracts the number of games from stats data at index 6
func (f *Fetcher) getGamesCount(data json.RawMessage) float64 {
	var stats []float64
	if err := json.Unmarshal(data, &stats); err != nil || len(stats) < 2 {
		return 0
	}
	return stats[1] // games is at index 1: [wins, games]
}

// getWinsAndGames extracts wins and games from stats data
func (f *Fetcher) getWinsAndGames(data json.RawMessage) (float64, float64) {
	var stats []float64
	if err := json.Unmarshal(data, &stats); err != nil || len(stats) < 2 {
		return 0, 0
	}
	return stats[0], stats[1] // [wins, games]
}
