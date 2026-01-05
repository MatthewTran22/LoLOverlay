package main

import (
	"fmt"

	"ghostdraft/internal/data"
)

// MetaChampion represents a champion in the meta list
type MetaChampion struct {
	ChampionID   int     `json:"championId"`
	ChampionName string  `json:"championName"`
	IconURL      string  `json:"iconURL"`
	WinRate      float64 `json:"winRate"`
	PickRate     float64 `json:"pickRate"`
	Games        int     `json:"games"`
}

// MetaData represents the top champions for all roles
type MetaData struct {
	Patch   string                    `json:"patch"`
	HasData bool                      `json:"hasData"`
	Roles   map[string][]MetaChampion `json:"roles"`
}

// ChampionDetailItem represents an item in a build
type ChampionDetailItem struct {
	ItemID  int     `json:"itemId"`
	Name    string  `json:"name"`
	IconURL string  `json:"iconURL"`
	WinRate float64 `json:"winRate"`
	Games   int     `json:"games"`
}

// ChampionDetailMatchup represents a matchup
type ChampionDetailMatchup struct {
	ChampionID   int     `json:"championId"`
	ChampionName string  `json:"championName"`
	IconURL      string  `json:"iconURL"`
	WinRate      float64 `json:"winRate"`
	Games        int     `json:"games"`
}

// ChampionDetails represents detailed info for a champion
type ChampionDetails struct {
	HasData      bool                    `json:"hasData"`
	ChampionID   int                     `json:"championId"`
	ChampionName string                  `json:"championName"`
	Role         string                  `json:"role"`
	CoreItems    []ChampionDetailItem    `json:"coreItems"`
	FourthItems  []ChampionDetailItem    `json:"fourthItems"`
	FifthItems   []ChampionDetailItem    `json:"fifthItems"`
	SixthItems   []ChampionDetailItem    `json:"sixthItems"`
	Counters     []ChampionDetailMatchup `json:"counters"`
	GoodMatchups []ChampionDetailMatchup `json:"goodMatchups"`
}

// BuildItem represents an item in a build
type BuildItem struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	IconURL string  `json:"iconURL"`
	WinRate float64 `json:"winRate,omitempty"`
	Games   int     `json:"games,omitempty"`
}

// BuildPath represents a single build path
type BuildPath struct {
	Name          string      `json:"name"`
	WinRate       float64     `json:"winRate"`
	Games         int         `json:"games"`
	StartingItems []BuildItem `json:"startingItems"`
	CoreItems     []BuildItem `json:"coreItems"`
	FourthItems   []BuildItem `json:"fourthItems"`
	FifthItems    []BuildItem `json:"fifthItems"`
	SixthItems    []BuildItem `json:"sixthItems"`
}

// ChampionBuildData represents build data for a champion
type ChampionBuildData struct {
	HasItems     bool        `json:"hasItems"`
	ChampionName string      `json:"championName"`
	ChampionID   int         `json:"championId"`
	Role         string      `json:"role"`
	IconURL      string      `json:"iconURL"`
	SplashURL    string      `json:"splashURL"`
	Builds       []BuildPath `json:"builds"`
}

// GetMetaChampions returns the top 5 champions by win rate for each role
func (a *App) GetMetaChampions() MetaData {
	result := MetaData{
		HasData: false,
		Roles:   make(map[string][]MetaChampion),
	}

	if a.statsProvider == nil {
		return result
	}

	result.Patch = a.statsProvider.GetPatch()

	roleData, err := a.statsProvider.FetchAllRolesTopChampions(5)
	if err != nil {
		return result
	}

	for role, champs := range roleData {
		var metaChamps []MetaChampion
		for _, c := range champs {
			name := a.champions.GetName(c.ChampionID)
			icon := a.champions.GetIconURL(c.ChampionID)
			metaChamps = append(metaChamps, MetaChampion{
				ChampionID:   c.ChampionID,
				ChampionName: name,
				IconURL:      icon,
				WinRate:      c.WinRate,
				PickRate:     c.PickRate,
				Games:        c.Matches,
			})
		}
		result.Roles[role] = metaChamps
	}

	result.HasData = true
	return result
}

// GetChampionBuild returns build data for a champion in the same format as items:update
func (a *App) GetChampionBuild(championID int, role string) ChampionBuildData {
	result := ChampionBuildData{
		HasItems:   false,
		ChampionID: championID,
		Builds:     []BuildPath{},
	}

	champName := a.champions.GetName(championID)
	result.ChampionName = champName
	result.Role = role
	result.IconURL = a.champions.GetIconURL(championID)
	result.SplashURL = a.champions.GetSplashURL(championID)

	if a.statsProvider == nil {
		return result
	}

	buildData, err := a.statsProvider.FetchChampionData(championID, champName, role)
	if err != nil || buildData == nil || len(buildData.Builds) == 0 {
		return result
	}

	result.HasItems = true

	// Helper to convert item IDs to BuildItem
	convertItems := func(itemIDs []int) []BuildItem {
		var items []BuildItem
		for _, itemID := range itemIDs {
			items = append(items, BuildItem{
				ID:      itemID,
				Name:    a.items.GetName(itemID),
				IconURL: a.items.GetIconURL(itemID),
			})
		}
		return items
	}

	// Helper to convert item options with win rates
	convertItemOptions := func(options []data.ItemOption) []BuildItem {
		var items []BuildItem
		for _, opt := range options {
			items = append(items, BuildItem{
				ID:      opt.ItemID,
				Name:    a.items.GetName(opt.ItemID),
				IconURL: a.items.GetIconURL(opt.ItemID),
				WinRate: opt.WinRate,
				Games:   opt.Games,
			})
		}
		return items
	}

	// Convert all build paths
	for _, build := range buildData.Builds {
		buildName := "Build"
		if len(build.CoreItems) > 0 {
			buildName = a.items.GetName(build.CoreItems[0])
		}

		result.Builds = append(result.Builds, BuildPath{
			Name:          buildName,
			WinRate:       build.WinRate,
			Games:         build.Games,
			StartingItems: convertItems(build.StartingItems),
			CoreItems:     convertItems(build.CoreItems),
			FourthItems:   convertItemOptions(build.FourthItemOptions),
			FifthItems:    convertItemOptions(build.FifthItemOptions),
			SixthItems:    convertItemOptions(build.SixthItemOptions),
		})
	}

	return result
}

// GetChampionDetails returns detailed build and matchup info for a champion
func (a *App) GetChampionDetails(championID int, role string) ChampionDetails {
	result := ChampionDetails{
		HasData:      false,
		ChampionID:   championID,
		Role:         role,
		CoreItems:    []ChampionDetailItem{},
		FourthItems:  []ChampionDetailItem{},
		FifthItems:   []ChampionDetailItem{},
		SixthItems:   []ChampionDetailItem{},
		Counters:     []ChampionDetailMatchup{},
		GoodMatchups: []ChampionDetailMatchup{},
	}

	if a.statsProvider == nil {
		return result
	}

	champName := a.champions.GetName(championID)
	result.ChampionName = champName

	// Fetch build data
	buildData, err := a.statsProvider.FetchChampionData(championID, champName, role)
	if err == nil && buildData != nil && len(buildData.Builds) > 0 {
		result.HasData = true
		build := buildData.Builds[0]

		// Core items
		for _, itemID := range build.CoreItems {
			result.CoreItems = append(result.CoreItems, ChampionDetailItem{
				ItemID:  itemID,
				Name:    a.items.GetName(itemID),
				IconURL: a.items.GetIconURL(itemID),
			})
		}

		// 4th item options
		for _, opt := range build.FourthItemOptions[:min(3, len(build.FourthItemOptions))] {
			result.FourthItems = append(result.FourthItems, ChampionDetailItem{
				ItemID:  opt.ItemID,
				Name:    a.items.GetName(opt.ItemID),
				IconURL: a.items.GetIconURL(opt.ItemID),
				WinRate: opt.WinRate,
				Games:   opt.Games,
			})
		}

		// 5th item options
		for _, opt := range build.FifthItemOptions[:min(3, len(build.FifthItemOptions))] {
			result.FifthItems = append(result.FifthItems, ChampionDetailItem{
				ItemID:  opt.ItemID,
				Name:    a.items.GetName(opt.ItemID),
				IconURL: a.items.GetIconURL(opt.ItemID),
				WinRate: opt.WinRate,
				Games:   opt.Games,
			})
		}

		// 6th item options
		for _, opt := range build.SixthItemOptions[:min(3, len(build.SixthItemOptions))] {
			result.SixthItems = append(result.SixthItems, ChampionDetailItem{
				ItemID:  opt.ItemID,
				Name:    a.items.GetName(opt.ItemID),
				IconURL: a.items.GetIconURL(opt.ItemID),
				WinRate: opt.WinRate,
				Games:   opt.Games,
			})
		}
	}

	// Fetch counters (champions that beat you) - separate from allMatchups
	counters, err := a.statsProvider.FetchCounterMatchups(championID, role, 6)
	if err != nil {
		fmt.Printf("Failed to fetch counters for %s: %v\n", champName, err)
	} else {
		fmt.Printf("Fetched %d counters for %s:\n", len(counters), champName)
		for _, m := range counters {
			enemyName := a.champions.GetName(m.EnemyChampionID)
			iconURL := a.champions.GetIconURL(m.EnemyChampionID)
			fmt.Printf("  - %s: %.1f%% WR (%d games)\n", enemyName, m.WinRate, m.Matches)
			result.Counters = append(result.Counters, ChampionDetailMatchup{
				ChampionID:   m.EnemyChampionID,
				ChampionName: enemyName,
				IconURL:      iconURL,
				WinRate:      m.WinRate,
				Games:        m.Matches,
			})
		}
		if len(counters) > 0 {
			result.HasData = true
		}
	}

	// Fetch good matchups (champions you beat)
	allMatchups, err := a.statsProvider.FetchAllMatchups(championID, role)
	if err == nil && len(allMatchups) > 0 {
		result.HasData = true

		// Sort allMatchups by win rate descending
		for i := 0; i < len(allMatchups); i++ {
			for j := i + 1; j < len(allMatchups); j++ {
				if allMatchups[j].WinRate > allMatchups[i].WinRate {
					allMatchups[i], allMatchups[j] = allMatchups[j], allMatchups[i]
				}
			}
		}
		for i := 0; i < min(5, len(allMatchups)); i++ {
			m := allMatchups[i]
			if m.Matches < 20 {
				continue
			}
			enemyName := a.champions.GetName(m.EnemyChampionID)
			iconURL := a.champions.GetIconURL(m.EnemyChampionID)
			result.GoodMatchups = append(result.GoodMatchups, ChampionDetailMatchup{
				ChampionID:   m.EnemyChampionID,
				ChampionName: enemyName,
				IconURL:      iconURL,
				WinRate:      m.WinRate,
				Games:        m.Matches,
			})
		}
	}

	return result
}
