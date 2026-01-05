package main

import (
	"fmt"

	"ghostdraft/internal/data"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// fetchAndEmitBuild fetches matchup data from our database and emits it to frontend
func (a *App) fetchAndEmitBuild(championID int, championName string, role string, enemyChampionIDs []int) {
	fmt.Printf("Fetching matchup for %s (%s) vs %d enemies...\n", championName, role, len(enemyChampionIDs))

	patch := ""
	if a.statsProvider != nil {
		patch = a.statsProvider.GetPatch()
	}

	if len(enemyChampionIDs) == 0 {
		runtime.EventsEmit(a.ctx, "build:update", map[string]interface{}{
			"hasBuild":     true,
			"championName": championName,
			"role":         role,
			"winRate":      "-",
			"winRateLabel": "Waiting for enemy...",
			"patch":        patch,
		})
		fmt.Printf("No enemies detected yet for %s\n", championName)
		return
	}

	if a.statsProvider == nil {
		runtime.EventsEmit(a.ctx, "build:update", map[string]interface{}{
			"hasBuild": false,
			"error":    "Stats provider not available",
		})
		fmt.Println("Stats provider not available for matchups")
		return
	}

	// Fetch our matchups - this gives us all enemies we face in our role
	matchups, err := a.statsProvider.FetchAllMatchups(championID, role)
	if err != nil {
		runtime.EventsEmit(a.ctx, "build:update", map[string]interface{}{
			"hasBuild": false,
			"error":    err.Error(),
		})
		fmt.Printf("Failed to fetch matchups: %v\n", err)
		return
	}

	// Find enemy with highest game count in matchup data (likely lane opponent)
	var laneOpponentID int
	var matchupWR float64
	var matchupGames int
	for _, enemyID := range enemyChampionIDs {
		for _, m := range matchups {
			if m.EnemyChampionID == enemyID && m.Matches > matchupGames {
				laneOpponentID = enemyID
				matchupWR = m.WinRate
				matchupGames = m.Matches
			}
		}
	}
	if laneOpponentID > 0 {
		fmt.Printf("Lane opponent (highest games): %d (%.1f%% WR, %d games)\n", laneOpponentID, matchupWR, matchupGames)
	}

	if laneOpponentID == 0 {
		runtime.EventsEmit(a.ctx, "build:update", map[string]interface{}{
			"hasBuild":     true,
			"championName": championName,
			"role":         role,
			"winRate":      "-",
			"winRateLabel": "No lane opponent found",
			"patch":        patch,
		})
		fmt.Printf("No lane opponent found in matchup data for %s\n", championName)
		return
	}

	enemyName := a.champions.GetName(laneOpponentID)

	// Determine matchup status: winning (>51%), losing (<49%), even (49-51%)
	var matchupStatus string
	if matchupWR >= 51.0 {
		matchupStatus = "winning"
	} else if matchupWR <= 49.0 {
		matchupStatus = "losing"
	} else {
		matchupStatus = "even"
	}

	fmt.Printf("Matchup: %s vs %s = %.1f%% (%s, %d games)\n", championName, enemyName, matchupWR, matchupStatus, matchupGames)
	runtime.EventsEmit(a.ctx, "build:update", map[string]interface{}{
		"hasBuild":      true,
		"championName":  championName,
		"role":          role,
		"winRate":       fmt.Sprintf("%.1f%%", matchupWR),
		"winRateLabel":  fmt.Sprintf("vs %s", enemyName),
		"enemyName":     enemyName,
		"matchupStatus": matchupStatus,
		"patch":         patch,
	})
}

// fetchAndEmitCounterPicks fetches champions that counter the enemy laner
func (a *App) fetchAndEmitCounterPicks(enemyChampionID int, role string) {
	enemyName := a.champions.GetName(enemyChampionID)
	fmt.Printf("Fetching counter picks vs %s (%s)...\n", enemyName, role)

	if a.statsProvider == nil {
		fmt.Println("Stats provider not available for counter picks")
		runtime.EventsEmit(a.ctx, "counterpicks:update", map[string]interface{}{
			"hasData": false,
		})
		return
	}

	counterPicks, err := a.statsProvider.FetchCounterPicks(enemyChampionID, role, 6)
	if err != nil || len(counterPicks) == 0 {
		fmt.Printf("No counter pick data vs %s: %v\n", enemyName, err)
		runtime.EventsEmit(a.ctx, "counterpicks:update", map[string]interface{}{
			"hasData":   true,
			"enemyName": enemyName,
			"enemyIcon": a.champions.GetIconURL(enemyChampionID),
			"picks":     []map[string]interface{}{},
		})
		return
	}

	// Convert to frontend format
	var pickList []map[string]interface{}
	for _, m := range counterPicks {
		champName := a.champions.GetName(m.EnemyChampionID)
		pickList = append(pickList, map[string]interface{}{
			"championID":   m.EnemyChampionID,
			"championName": champName,
			"iconURL":      a.champions.GetIconURL(m.EnemyChampionID),
			"winRate":      m.WinRate,
			"games":        m.Matches,
		})
	}

	fmt.Printf("Counter picks vs %s: ", enemyName)
	for _, p := range pickList {
		fmt.Printf("%s (%.1f%%) ", p["championName"], p["winRate"])
	}
	fmt.Println()

	runtime.EventsEmit(a.ctx, "counterpicks:update", map[string]interface{}{
		"hasData":   true,
		"enemyName": enemyName,
		"enemyIcon": a.champions.GetIconURL(enemyChampionID),
		"picks":     pickList,
	})
}

// fetchAndEmitRecommendedBans fetches hardest counters and emits as recommended bans
func (a *App) fetchAndEmitRecommendedBans(championID int, role string) {
	championName := a.champions.GetName(championID)
	fmt.Printf("Fetching recommended bans for %s (%s)...\n", championName, role)

	// Use our stats provider for counter matchups
	if a.statsProvider == nil {
		fmt.Println("Stats provider not available for bans")
		runtime.EventsEmit(a.ctx, "bans:update", map[string]interface{}{
			"hasBans":      true,
			"championName": championName,
			"role":         role,
			"bans":         []map[string]interface{}{},
			"noData":       true,
		})
		return
	}

	matchups, err := a.statsProvider.FetchCounterMatchups(championID, role, 5)
	if err != nil || len(matchups) == 0 {
		fmt.Printf("No matchup data for %s %s: %v\n", championName, role, err)
		runtime.EventsEmit(a.ctx, "bans:update", map[string]interface{}{
			"hasBans":      true,
			"championName": championName,
			"role":         role,
			"bans":         []map[string]interface{}{},
			"noData":       true,
		})
		return
	}

	// Convert to frontend format
	var banList []map[string]interface{}
	for _, m := range matchups {
		enemyName := a.champions.GetName(m.EnemyChampionID)
		damageType := "Unknown"
		if a.championDB != nil {
			damageType = a.championDB.GetDamageType(enemyName)
		}
		banList = append(banList, map[string]interface{}{
			"championID":   m.EnemyChampionID,
			"championName": enemyName,
			"iconURL":      a.champions.GetIconURL(m.EnemyChampionID),
			"damageType":   damageType,
			"winRate":      m.WinRate,
			"games":        m.Matches,
		})
	}

	fmt.Printf("Counter matchups for %s: ", championName)
	for _, b := range banList {
		fmt.Printf("%s (%.1f%%) ", b["championName"], b["winRate"])
	}
	fmt.Println()

	runtime.EventsEmit(a.ctx, "bans:update", map[string]interface{}{
		"hasBans":      true,
		"championName": championName,
		"role":         role,
		"bans":         banList,
	})
}

// fetchAndEmitItems fetches item build from our stats database and emits to frontend
func (a *App) fetchAndEmitItems(championID int, championName string, role string) {
	fmt.Printf("Fetching items for %s (%s)...\n", championName, role)

	if a.statsProvider == nil {
		fmt.Println("Stats provider not available")
		runtime.EventsEmit(a.ctx, "items:update", map[string]interface{}{
			"hasItems": false,
		})
		return
	}

	buildData, err := a.statsProvider.FetchChampionData(championID, championName, role)
	if err != nil {
		fmt.Printf("No data for %s: %v\n", championName, err)
		runtime.EventsEmit(a.ctx, "items:update", map[string]interface{}{
			"hasItems": false,
		})
		return
	}

	// Helper to convert item IDs to frontend format
	convertItems := func(itemIDs []int) []map[string]interface{} {
		var result []map[string]interface{}
		for _, itemID := range itemIDs {
			result = append(result, map[string]interface{}{
				"id":      itemID,
				"name":    a.items.GetName(itemID),
				"iconURL": a.items.GetIconURL(itemID),
			})
		}
		return result
	}

	// Helper to convert item options with win rates
	convertItemOptions := func(options []data.ItemOption) []map[string]interface{} {
		var result []map[string]interface{}
		for _, opt := range options {
			result = append(result, map[string]interface{}{
				"id":      opt.ItemID,
				"name":    a.items.GetName(opt.ItemID),
				"iconURL": a.items.GetIconURL(opt.ItemID),
				"winRate": opt.WinRate,
				"games":   opt.Games,
			})
		}
		return result
	}

	// Convert all build paths
	var builds []map[string]interface{}
	for _, build := range buildData.Builds {
		// Name the build after the first core item
		buildName := "Build"
		if len(build.CoreItems) > 0 {
			buildName = a.items.GetName(build.CoreItems[0])
		}

		builds = append(builds, map[string]interface{}{
			"name":          buildName,
			"winRate":       build.WinRate,
			"games":         build.Games,
			"startingItems": convertItems(build.StartingItems),
			"coreItems":     convertItems(build.CoreItems),
			"fourthItems":   convertItemOptions(build.FourthItemOptions),
			"fifthItems":    convertItemOptions(build.FifthItemOptions),
			"sixthItems":    convertItemOptions(build.SixthItemOptions),
		})
	}

	fmt.Printf("Found %d build paths for %s\n", len(builds), championName)

	runtime.EventsEmit(a.ctx, "items:update", map[string]interface{}{
		"hasItems":     true,
		"championName": championName,
		"role":         role,
		"builds":       builds,
	})
}
