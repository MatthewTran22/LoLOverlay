package main

import (
	"fmt"
	"os"

	"ghostdraft/internal/data"
	"ghostdraft/internal/lcu"
)

// ForceStatsUpdate forces a redownload of stats data
func (a *App) ForceStatsUpdate() string {
	if a.statsDB == nil {
		return "Stats database not initialized"
	}

	manifestURL := os.Getenv("STATS_MANIFEST_URL")
	if manifestURL == "" {
		manifestURL = data.DefaultManifestURL
	}

	if err := a.statsDB.ForceUpdate(manifestURL); err != nil {
		return fmt.Sprintf("Update failed: %v", err)
	}

	// Recreate stats provider with new data
	if a.statsDB.HasData() {
		provider, err := data.NewStatsProvider(a.statsDB)
		if err != nil {
			return fmt.Sprintf("Provider creation failed: %v", err)
		}
		if provider.GetPatch() == "" {
			provider.FetchPatch()
		}
		a.statsProvider = provider
		return fmt.Sprintf("Updated to %s", provider.GetPatch())
	}

	return "Update completed but no data available"
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
