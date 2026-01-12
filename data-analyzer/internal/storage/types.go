package storage

// RawMatch represents a flattened match record for JSONL storage
// One record per participant (10 rows per match)
type RawMatch struct {
	// Match identifiers
	MatchID      string `json:"matchId"`
	GameVersion  string `json:"gameVersion"`
	GameDuration int    `json:"gameDuration"`
	GameCreation int64  `json:"gameCreation"`

	// Participant data
	PUUID        string `json:"puuid"`
	GameName     string `json:"gameName,omitempty"`
	TagLine      string `json:"tagLine,omitempty"`
	ChampionID   int    `json:"championId"`
	ChampionName string `json:"championName"`
	TeamPosition string `json:"teamPosition"` // TOP, JUNGLE, MIDDLE, BOTTOM, UTILITY
	Win          bool   `json:"win"`

	// Final items (used for item stats and build inference)
	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`

	// BuildOrder contains the order items were purchased (from timeline, ~20% of matches)
	// Used for champion_item_slots table (1st item, 2nd item, etc.)
	BuildOrder []int `json:"buildOrder,omitempty"`
}

// GetFinalItems returns the final inventory items as a slice (excluding empty slots)
func (r *RawMatch) GetFinalItems() []int {
	items := []int{r.Item0, r.Item1, r.Item2, r.Item3, r.Item4, r.Item5}
	result := make([]int, 0, 6)
	for _, item := range items {
		if item > 0 {
			result = append(result, item)
		}
	}
	return result
}
