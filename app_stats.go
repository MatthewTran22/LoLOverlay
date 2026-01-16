package main

import (
	"fmt"

	"ghostdraft/internal/lcu"
)

// ForceStatsUpdate clears the cache and refreshes stats data from Turso
func (a *App) ForceStatsUpdate() string {
	if a.statsProvider == nil {
		return "Stats provider not initialized"
	}

	// Clear the query cache
	a.statsProvider.ClearCache()

	// Refetch patch info
	if err := a.statsProvider.FetchPatch(); err != nil {
		return fmt.Sprintf("Failed to refresh: %v", err)
	}

	return fmt.Sprintf("Cache cleared, using patch %s", a.statsProvider.GetPatch())
}

// GetPersonalStats returns aggregated personal stats from recent match history
func (a *App) GetPersonalStats() *lcu.PersonalStats {
	emptyStats := &lcu.PersonalStats{HasData: false}

	if !a.lcuClient.IsConnected() {
		return emptyStats
	}

	history, err := a.lcuClient.FetchMatchHistory(20)
	if err != nil {
		fmt.Printf("Failed to fetch match history: %v\n", err)
		return emptyStats
	}

	return lcu.CalculatePersonalStats(history, a.champions)
}
