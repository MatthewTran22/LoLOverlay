package db

import (
	"context"
)

// AggregatedChampionStats from the reducer
type AggregatedChampionStats struct {
	Patch        string  `json:"patch"`
	ChampionID   int     `json:"championId"`
	TeamPosition string  `json:"teamPosition"`
	Wins         int     `json:"wins"`
	Matches      int     `json:"matches"`
	WinRate      float64 `json:"winRate"`
}

// AggregatedItemStats from the reducer
type AggregatedItemStats struct {
	Patch        string  `json:"patch"`
	ChampionID   int     `json:"championId"`
	TeamPosition string  `json:"teamPosition"`
	ItemID       int     `json:"itemId"`
	Wins         int     `json:"wins"`
	Matches      int     `json:"matches"`
	WinRate      float64 `json:"winRate"`
	PickRate     float64 `json:"pickRate"` // % of games this item was built
}

// GetAggregatedChampionStats returns champion stats from the reducer
func (db *DB) GetAggregatedChampionStats(ctx context.Context, patch string) ([]AggregatedChampionStats, error) {
	query := `
		SELECT patch, champion_id, team_position, wins, matches
		FROM champion_stats
		WHERE ($1 = '' OR patch = $1)
		ORDER BY matches DESC
	`

	rows, err := db.pool.Query(ctx, query, patch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []AggregatedChampionStats
	for rows.Next() {
		var s AggregatedChampionStats
		if err := rows.Scan(&s.Patch, &s.ChampionID, &s.TeamPosition, &s.Wins, &s.Matches); err != nil {
			return nil, err
		}
		if s.Matches > 0 {
			s.WinRate = float64(s.Wins) / float64(s.Matches) * 100
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetAggregatedItemStats returns item stats for a champion from the reducer
func (db *DB) GetAggregatedItemStats(ctx context.Context, patch string, championID int, position string) ([]AggregatedItemStats, error) {
	// First get total matches for this champion/position to calculate pick rate
	var totalMatches int
	err := db.pool.QueryRow(ctx, `
		SELECT COALESCE(matches, 0) FROM champion_stats
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3
	`, patch, championID, position).Scan(&totalMatches)
	if err != nil {
		totalMatches = 0
	}

	query := `
		SELECT patch, champion_id, team_position, item_id, wins, matches
		FROM champion_items
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3
		ORDER BY matches DESC
	`

	rows, err := db.pool.Query(ctx, query, patch, championID, position)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []AggregatedItemStats
	for rows.Next() {
		var s AggregatedItemStats
		if err := rows.Scan(&s.Patch, &s.ChampionID, &s.TeamPosition, &s.ItemID, &s.Wins, &s.Matches); err != nil {
			return nil, err
		}
		if s.Matches > 0 {
			s.WinRate = float64(s.Wins) / float64(s.Matches) * 100
		}
		if totalMatches > 0 {
			s.PickRate = float64(s.Matches) / float64(totalMatches) * 100
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetPatches returns all available patches
func (db *DB) GetPatches(ctx context.Context) ([]string, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT DISTINCT patch FROM champion_stats ORDER BY patch DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patches []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		patches = append(patches, p)
	}
	return patches, nil
}

// AggregatedMatchupStats from the reducer
type AggregatedMatchupStats struct {
	Patch           string  `json:"patch"`
	ChampionID      int     `json:"championId"`
	TeamPosition    string  `json:"teamPosition"`
	EnemyChampionID int     `json:"enemyChampionId"`
	Wins            int     `json:"wins"`
	Matches         int     `json:"matches"`
	WinRate         float64 `json:"winRate"`
}

// GetAggregatedMatchupStats returns matchup stats for a champion
// Returns best (highest WR) and worst (lowest WR) matchups
func (db *DB) GetAggregatedMatchupStats(ctx context.Context, patch string, championID int, position string, limit int) (best []AggregatedMatchupStats, worst []AggregatedMatchupStats, err error) {
	minGames := 10 // Minimum games for statistical relevance

	// Get best matchups (highest win rate)
	bestQuery := `
		SELECT patch, champion_id, team_position, enemy_champion_id, wins, matches
		FROM champion_matchups
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3 AND matches >= $4
		ORDER BY (wins::float / matches::float) DESC
		LIMIT $5
	`

	rows, err := db.pool.Query(ctx, bestQuery, patch, championID, position, minGames, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s AggregatedMatchupStats
		if err := rows.Scan(&s.Patch, &s.ChampionID, &s.TeamPosition, &s.EnemyChampionID, &s.Wins, &s.Matches); err != nil {
			return nil, nil, err
		}
		if s.Matches > 0 {
			s.WinRate = float64(s.Wins) / float64(s.Matches) * 100
		}
		best = append(best, s)
	}

	// Get worst matchups (lowest win rate)
	worstQuery := `
		SELECT patch, champion_id, team_position, enemy_champion_id, wins, matches
		FROM champion_matchups
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3 AND matches >= $4
		ORDER BY (wins::float / matches::float) ASC
		LIMIT $5
	`

	rows2, err := db.pool.Query(ctx, worstQuery, patch, championID, position, minGames, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var s AggregatedMatchupStats
		if err := rows2.Scan(&s.Patch, &s.ChampionID, &s.TeamPosition, &s.EnemyChampionID, &s.Wins, &s.Matches); err != nil {
			return nil, nil, err
		}
		if s.Matches > 0 {
			s.WinRate = float64(s.Wins) / float64(s.Matches) * 100
		}
		worst = append(worst, s)
	}

	return best, worst, nil
}

// GetStatsOverview returns a summary of the aggregated data
func (db *DB) GetStatsOverview(ctx context.Context) (map[string]interface{}, error) {
	overview := make(map[string]interface{})

	// Count unique champions
	var champCount int
	err := db.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT champion_id) FROM champion_stats`).Scan(&champCount)
	if err != nil {
		champCount = 0
	}
	overview["uniqueChampions"] = champCount

	// Count total matches aggregated
	var totalMatches int
	err = db.pool.QueryRow(ctx, `SELECT COALESCE(SUM(matches), 0) / 10 FROM champion_stats`).Scan(&totalMatches)
	if err != nil {
		totalMatches = 0
	}
	overview["totalMatchesAggregated"] = totalMatches

	// Count patches
	var patchCount int
	err = db.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT patch) FROM champion_stats`).Scan(&patchCount)
	if err != nil {
		patchCount = 0
	}
	overview["patches"] = patchCount

	// Count item stats entries
	var itemStatCount int
	err = db.pool.QueryRow(ctx, `SELECT COUNT(*) FROM champion_items`).Scan(&itemStatCount)
	if err != nil {
		itemStatCount = 0
	}
	overview["itemStatEntries"] = itemStatCount

	return overview, nil
}
