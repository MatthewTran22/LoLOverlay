package stats

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ItemOption holds item ID with win rate
type ItemOption struct {
	ItemID  int
	WinRate float64
	Games   int
}

// BuildPath represents a single build path
type BuildPath struct {
	Name              string
	WinRate           float64
	Games             int
	StartingItems     []int
	CoreItems         []int
	FourthItemOptions []ItemOption
	FifthItemOptions  []ItemOption
	SixthItemOptions  []ItemOption
}

// BuildData holds champion build information
type BuildData struct {
	ChampionID   int
	ChampionName string
	Role         string
	Builds       []BuildPath
}

// Provider fetches build data from our PostgreSQL database
type Provider struct {
	pool         *pgxpool.Pool
	currentPatch string
}

// ItemStat represents aggregated item statistics
type ItemStat struct {
	ItemID   int
	Wins     int
	Matches  int
	WinRate  float64
	PickRate float64
}

// MatchupStat represents a champion matchup statistic
type MatchupStat struct {
	EnemyChampionID int
	Wins            int
	Matches         int
	WinRate         float64
}

// NewProvider creates a new stats provider
func NewProvider(databaseURL string) (*Provider, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Provider{pool: pool}, nil
}

// Close closes the database connection
func (p *Provider) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

// FetchPatch gets the latest patch from our database
func (p *Provider) FetchPatch() error {
	ctx := context.Background()

	var patch string
	err := p.pool.QueryRow(ctx, `
		SELECT patch FROM champion_stats
		ORDER BY patch DESC
		LIMIT 1
	`).Scan(&patch)

	if err != nil {
		return fmt.Errorf("failed to get patch: %w", err)
	}

	p.currentPatch = patch
	fmt.Printf("[Stats] Using patch: %s\n", patch)
	return nil
}

// GetPatch returns the current patch
func (p *Provider) GetPatch() string {
	return p.currentPatch
}

// roleToPosition converts role names to database team_position values
func roleToPosition(role string) string {
	switch role {
	case "top":
		return "TOP"
	case "jungle":
		return "JUNGLE"
	case "middle", "mid":
		return "MIDDLE"
	case "bottom", "adc":
		return "BOTTOM"
	case "utility", "support":
		return "UTILITY"
	default:
		return "MIDDLE"
	}
}

// FetchChampionData gets build data for a champion from our database
func (p *Provider) FetchChampionData(championID int, championName string, role string) (*BuildData, error) {
	ctx := context.Background()
	position := roleToPosition(role)

	// Get total games for this champion/position to calculate pick rates
	var totalGames int
	err := p.pool.QueryRow(ctx, `
		SELECT COALESCE(matches, 0) FROM champion_stats
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3
	`, p.currentPatch, championID, position).Scan(&totalGames)

	if err != nil || totalGames == 0 {
		// Try without patch filter if no data for current patch
		err = p.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(matches), 0) FROM champion_stats
			WHERE champion_id = $1 AND team_position = $2
		`, championID, position).Scan(&totalGames)

		if err != nil || totalGames == 0 {
			return nil, fmt.Errorf("no data for champion %d in position %s", championID, position)
		}
	}

	// Get item stats
	rows, err := p.pool.Query(ctx, `
		SELECT item_id, wins, matches
		FROM champion_items
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3
		ORDER BY matches DESC
	`, p.currentPatch, championID, position)

	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []ItemStat
	for rows.Next() {
		var item ItemStat
		if err := rows.Scan(&item.ItemID, &item.Wins, &item.Matches); err != nil {
			continue
		}
		if item.Matches > 0 {
			item.WinRate = float64(item.Wins) / float64(item.Matches) * 100
			item.PickRate = float64(item.Matches) / float64(totalGames) * 100
		}
		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no item data for champion %d", championID)
	}

	// Build the response
	build := p.constructBuildPath(items, totalGames)

	return &BuildData{
		ChampionID:   championID,
		ChampionName: championName,
		Role:         role,
		Builds:       []BuildPath{build},
	}, nil
}

// constructBuildPath creates a build path from item stats
func (p *Provider) constructBuildPath(items []ItemStat, totalGames int) BuildPath {
	// Separate items by category
	var boots []ItemStat
	var mythics []ItemStat // High-cost items (likely first items)
	var legendary []ItemStat

	for _, item := range items {
		if isBootsItem(item.ItemID) {
			boots = append(boots, item)
		} else if item.PickRate >= 30 { // Highly picked = likely core
			mythics = append(mythics, item)
		} else {
			legendary = append(legendary, item)
		}
	}

	// Sort by pick rate for core items
	sort.Slice(mythics, func(i, j int) bool {
		return mythics[i].PickRate > mythics[j].PickRate
	})

	// Sort legendary by a mix of pick rate and win rate
	sort.Slice(legendary, func(i, j int) bool {
		// Weight: 70% pick rate, 30% win rate
		scoreI := legendary[i].PickRate*0.7 + legendary[j].WinRate*0.3
		scoreJ := legendary[j].PickRate*0.7 + legendary[j].WinRate*0.3
		return scoreI > scoreJ
	})

	// Build core items (top 3 by pick rate)
	var coreItems []int
	for i := 0; i < len(mythics) && i < 3; i++ {
		coreItems = append(coreItems, mythics[i].ItemID)
	}

	// Calculate overall win rate from champion stats
	var winRate float64
	if totalGames > 0 && len(mythics) > 0 {
		// Use the win rate of the most picked item as proxy
		winRate = mythics[0].WinRate
	}

	// Fourth, Fifth, Sixth item options
	fourthItems := p.toItemOptions(legendary, 0, 5)
	fifthItems := p.toItemOptions(legendary, 0, 5)
	sixthItems := p.toItemOptions(legendary, 0, 5)

	// Starting items (boots if available)
	var startingItems []int
	if len(boots) > 0 {
		startingItems = append(startingItems, boots[0].ItemID)
	}

	return BuildPath{
		Name:              "Recommended Build",
		WinRate:           winRate,
		Games:             totalGames,
		StartingItems:     startingItems,
		CoreItems:         coreItems,
		FourthItemOptions: fourthItems,
		FifthItemOptions:  fifthItems,
		SixthItemOptions:  sixthItems,
	}
}

// toItemOptions converts ItemStats to ItemOptions
func (p *Provider) toItemOptions(items []ItemStat, start, count int) []ItemOption {
	var options []ItemOption
	end := start + count
	if end > len(items) {
		end = len(items)
	}

	for i := start; i < end; i++ {
		options = append(options, ItemOption{
			ItemID:  items[i].ItemID,
			WinRate: items[i].WinRate,
			Games:   items[i].Matches,
		})
	}
	return options
}

// isBootsItem checks if an item is boots
func isBootsItem(itemID int) bool {
	boots := map[int]bool{
		3006: true, // Berserker's Greaves
		3009: true, // Boots of Swiftness
		3020: true, // Sorcerer's Shoes
		3047: true, // Plated Steelcaps
		3111: true, // Mercury's Treads
		3117: true, // Mobility Boots
		3158: true, // Ionian Boots of Lucidity
	}
	return boots[itemID]
}

// HasData checks if we have data for a champion
func (p *Provider) HasData(championID int, role string) bool {
	ctx := context.Background()
	position := roleToPosition(role)

	var count int
	err := p.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM champion_items
		WHERE champion_id = $1 AND team_position = $2
	`, championID, position).Scan(&count)

	return err == nil && count > 0
}

// FetchMatchup returns the win rate for a specific champion vs enemy matchup
func (p *Provider) FetchMatchup(championID int, enemyChampionID int, role string) (*MatchupStat, error) {
	ctx := context.Background()
	position := roleToPosition(role)

	var m MatchupStat
	m.EnemyChampionID = enemyChampionID

	// Try current patch first
	err := p.pool.QueryRow(ctx, `
		SELECT wins, matches
		FROM champion_matchups
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3 AND enemy_champion_id = $4
	`, p.currentPatch, championID, position, enemyChampionID).Scan(&m.Wins, &m.Matches)

	if err != nil || m.Matches == 0 {
		// Try without patch filter
		err = p.pool.QueryRow(ctx, `
			SELECT SUM(wins), SUM(matches)
			FROM champion_matchups
			WHERE champion_id = $1 AND team_position = $2 AND enemy_champion_id = $3
		`, championID, position, enemyChampionID).Scan(&m.Wins, &m.Matches)

		if err != nil || m.Matches == 0 {
			return nil, fmt.Errorf("no matchup data for %d vs %d", championID, enemyChampionID)
		}
	}

	m.WinRate = float64(m.Wins) / float64(m.Matches) * 100
	return &m, nil
}

// FetchAllMatchups returns all matchup data for a champion in a role
func (p *Provider) FetchAllMatchups(championID int, role string) ([]MatchupStat, error) {
	ctx := context.Background()
	position := roleToPosition(role)

	rows, err := p.pool.Query(ctx, `
		SELECT enemy_champion_id, wins, matches
		FROM champion_matchups
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3
		ORDER BY matches DESC
	`, p.currentPatch, championID, position)

	if err != nil {
		// Try without patch filter
		rows, err = p.pool.Query(ctx, `
			SELECT enemy_champion_id, SUM(wins) as wins, SUM(matches) as matches
			FROM champion_matchups
			WHERE champion_id = $1 AND team_position = $2
			GROUP BY enemy_champion_id
			ORDER BY SUM(matches) DESC
		`, championID, position)

		if err != nil {
			return nil, fmt.Errorf("failed to query matchups: %w", err)
		}
	}
	defer rows.Close()

	var matchups []MatchupStat
	for rows.Next() {
		var m MatchupStat
		if err := rows.Scan(&m.EnemyChampionID, &m.Wins, &m.Matches); err != nil {
			continue
		}
		if m.Matches > 0 {
			m.WinRate = float64(m.Wins) / float64(m.Matches) * 100
		}
		matchups = append(matchups, m)
	}

	return matchups, nil
}

// FetchCounterMatchups returns the champions that counter the specified champion
// (i.e., matchups where the specified champion has the lowest win rate)
func (p *Provider) FetchCounterMatchups(championID int, role string, limit int) ([]MatchupStat, error) {
	ctx := context.Background()
	position := roleToPosition(role)

	if limit <= 0 {
		limit = 10
	}

	// Query matchups ordered by lowest win rate (hardest counters first)
	// Only include matchups with at least 10 games for statistical significance
	rows, err := p.pool.Query(ctx, `
		SELECT enemy_champion_id, wins, matches
		FROM champion_matchups
		WHERE patch = $1 AND champion_id = $2 AND team_position = $3 AND matches >= 10
		ORDER BY (wins::float / matches::float) ASC
		LIMIT $4
	`, p.currentPatch, championID, position, limit)

	if err != nil {
		// Try without patch filter
		rows, err = p.pool.Query(ctx, `
			SELECT enemy_champion_id, SUM(wins) as wins, SUM(matches) as matches
			FROM champion_matchups
			WHERE champion_id = $1 AND team_position = $2
			GROUP BY enemy_champion_id
			HAVING SUM(matches) >= 10
			ORDER BY (SUM(wins)::float / SUM(matches)::float) ASC
			LIMIT $3
		`, championID, position, limit)

		if err != nil {
			return nil, fmt.Errorf("failed to query matchups: %w", err)
		}
	}
	defer rows.Close()

	var matchups []MatchupStat
	for rows.Next() {
		var m MatchupStat
		if err := rows.Scan(&m.EnemyChampionID, &m.Wins, &m.Matches); err != nil {
			continue
		}
		if m.Matches > 0 {
			m.WinRate = float64(m.Wins) / float64(m.Matches) * 100
		}
		matchups = append(matchups, m)
	}

	return matchups, nil
}
