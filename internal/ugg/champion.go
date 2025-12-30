package ugg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

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

	// Build URL
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

	var rawData map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to parse champion data: %w", err)
	}

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

// buildPathKey is used to group builds by first core item
type buildPathKey struct {
	firstItem int
}

// buildPathData aggregates data for a build path
type buildPathData struct {
	wins          float64
	games         float64
	bestGames     float64           // Best single-region games (for picking situational items)
	startingItems []int
	coreItems     []int
	statsData     []json.RawMessage // Stats from best region for situational items
}

// parseChampionData extracts build data from U.GG response
func (f *Fetcher) parseChampionData(rawData map[string]json.RawMessage, championID int, championName string, role string) (*BuildData, error) {
	roleID := roleToID(role)
	if roleID == "" {
		roleID = "5" // Default to mid
	}

	// Collect all builds grouped by first core item
	buildPaths := make(map[int]*buildPathData)

	for _, regionData := range rawData {
		var regionMap map[string]json.RawMessage
		if err := json.Unmarshal(regionData, &regionMap); err != nil {
			continue
		}

		roleData, ok := regionMap[roleID]
		if !ok {
			continue
		}

		var tierMap map[string]json.RawMessage
		if err := json.Unmarshal(roleData, &tierMap); err != nil {
			continue
		}

		tierData, ok := tierMap["3"] // Diamond+
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

		if len(statsData) <= 6 {
			continue
		}

		// Get wins/games from index 6
		wins, games := f.getWinsAndGames(statsData[6])
		if games == 0 {
			continue
		}

		// Get core items from index 3
		coreItems := f.parseItemsArray(statsData[3])
		if len(coreItems) == 0 {
			continue
		}

		firstItem := coreItems[0]

		// Get starting items from index 2
		startingItems := f.parseItemsArray(statsData[2])

		// Aggregate by first item
		if _, exists := buildPaths[firstItem]; !exists {
			buildPaths[firstItem] = &buildPathData{
				bestGames:     games,
				startingItems: startingItems,
				coreItems:     coreItems,
				statsData:     statsData,
			}
		}
		buildPaths[firstItem].wins += wins
		buildPaths[firstItem].games += games

		// Keep the stats from the region with most games for this build path
		if games > buildPaths[firstItem].bestGames {
			buildPaths[firstItem].bestGames = games
			buildPaths[firstItem].statsData = statsData
			buildPaths[firstItem].startingItems = startingItems
			buildPaths[firstItem].coreItems = coreItems
		}
	}

	if len(buildPaths) == 0 {
		return nil, fmt.Errorf("no data found for champion %d", championID)
	}

	// Convert to BuildPath slice and sort by games
	var builds []BuildPath
	for firstItem, data := range buildPaths {
		if data.games < 50 { // Filter out very low sample builds
			continue
		}

		winRate := 0.0
		if data.games > 0 {
			winRate = (data.wins / data.games) * 100
		}

		build := BuildPath{
			Name:          fmt.Sprintf("Build %d", firstItem), // Will be replaced with item name in app.go
			WinRate:       winRate,
			Games:         int(data.games),
			StartingItems: data.startingItems,
			CoreItems:     data.coreItems,
		}

		// Parse situational items from best stats
		if len(data.statsData) > 5 {
			f.parseSituationalItemsForPath(data.statsData[5], &build)
		}

		builds = append(builds, build)
	}

	// Sort by games descending
	sort.Slice(builds, func(i, j int) bool {
		return builds[i].Games > builds[j].Games
	})

	// Limit to top 3 builds
	if len(builds) > 3 {
		builds = builds[:3]
	}

	fmt.Printf("Found %d build paths for %s\n", len(builds), championName)

	return &BuildData{
		ChampionID:   championID,
		ChampionName: championName,
		Role:         role,
		Builds:       builds,
	}, nil
}

// parseItemsArray extracts item IDs from [?, ?, [items]] structure
func (f *Fetcher) parseItemsArray(data json.RawMessage) []int {
	var itemArray []json.RawMessage
	if err := json.Unmarshal(data, &itemArray); err != nil || len(itemArray) < 3 {
		return nil
	}

	var itemIDs []int
	if err := json.Unmarshal(itemArray[2], &itemIDs); err != nil {
		return nil
	}

	return itemIDs
}

// parseSituationalItemsForPath extracts 4th/5th/6th item options for a build path
func (f *Fetcher) parseSituationalItemsForPath(data json.RawMessage, build *BuildPath) {
	var slots []json.RawMessage
	if err := json.Unmarshal(data, &slots); err != nil || len(slots) < 3 {
		return
	}

	build.FourthItemOptions = f.parseSlotOptions(slots[0], 3)
	build.FifthItemOptions = f.parseSlotOptions(slots[1], 3)
	build.SixthItemOptions = f.parseSlotOptions(slots[2], 3)
}

// parseSlotOptions extracts top N items with win rates from a slot
func (f *Fetcher) parseSlotOptions(data json.RawMessage, limit int) []ItemOption {
	var options [][]float64
	if err := json.Unmarshal(data, &options); err != nil {
		return nil
	}

	var items []ItemOption
	for i, opt := range options {
		if i >= limit || len(opt) < 3 {
			break
		}
		itemID := int(opt[0])
		wins := opt[1]
		games := opt[2]
		winRate := 0.0
		if games > 0 {
			winRate = (wins / games) * 100
		}
		items = append(items, ItemOption{
			ItemID:  itemID,
			WinRate: winRate,
			Games:   int(games),
		})
	}
	return items
}
