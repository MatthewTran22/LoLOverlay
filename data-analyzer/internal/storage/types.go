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

	// Final items
	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`

	// Build order from timeline (completed items in purchase order)
	BuildOrder []int `json:"buildOrder"`
}
