package ugg

// ItemOption holds item ID with win rate
type ItemOption struct {
	ItemID  int
	WinRate float64
	Games   int
}

// BuildPath represents a single build path (e.g., AP vs AD)
type BuildPath struct {
	Name              string       // e.g., "Kraken Build", "AP Build"
	WinRate           float64
	Games             int
	StartingItems     []int
	CoreItems         []int
	FourthItemOptions []ItemOption
	FifthItemOptions  []ItemOption
	SixthItemOptions  []ItemOption
}

// BuildData holds champion build information from U.GG
type BuildData struct {
	ChampionID   int         `json:"championId"`
	ChampionName string      `json:"championName"`
	Role         string      `json:"role"`
	Builds       []BuildPath `json:"builds"` // Multiple build paths
}

// MatchupData holds matchup information
type MatchupData struct {
	EnemyChampionID int     `json:"enemyChampionId"`
	Wins            float64 `json:"wins"`
	Games           float64 `json:"games"`
	WinRate         float64 `json:"winRate"`
}
