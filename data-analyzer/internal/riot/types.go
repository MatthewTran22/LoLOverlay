package riot

// AccountResponse represents the response from /riot/account/v1/accounts/by-riot-id
type AccountResponse struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

// MatchResponse represents the response from /lol/match/v5/matches/{matchId}
type MatchResponse struct {
	Metadata MatchMetadata `json:"metadata"`
	Info     MatchInfo     `json:"info"`
}

type MatchMetadata struct {
	MatchID      string   `json:"matchId"`
	Participants []string `json:"participants"` // PUUIDs
}

type MatchInfo struct {
	GameCreation  int64              `json:"gameCreation"`
	GameDuration  int                `json:"gameDuration"`
	GameVersion   string             `json:"gameVersion"`
	QueueID       int                `json:"queueId"`
	Participants  []MatchParticipant `json:"participants"`
}

type MatchParticipant struct {
	PUUID          string `json:"puuid"`
	RiotIdGameName string `json:"riotIdGameName"`
	RiotIdTagline  string `json:"riotIdTagline"`
	ChampionID     int    `json:"championId"`
	ChampionName   string `json:"championName"`
	TeamPosition   string `json:"teamPosition"` // TOP, JUNGLE, MIDDLE, BOTTOM, UTILITY
	Win            bool   `json:"win"`
	Item0          int    `json:"item0"`
	Item1          int    `json:"item1"`
	Item2          int    `json:"item2"`
	Item3          int    `json:"item3"`
	Item4          int    `json:"item4"`
	Item5          int    `json:"item5"`
	Item6          int    `json:"item6"` // Trinket
}

// TimelineResponse represents the response from /lol/match/v5/matches/{matchId}/timeline
type TimelineResponse struct {
	Metadata TimelineMetadata `json:"metadata"`
	Info     TimelineInfo     `json:"info"`
}

type TimelineMetadata struct {
	MatchID      string   `json:"matchId"`
	Participants []string `json:"participants"` // PUUIDs
}

type TimelineInfo struct {
	FrameInterval int             `json:"frameInterval"`
	Frames        []TimelineFrame `json:"frames"`
}

type TimelineFrame struct {
	Timestamp int              `json:"timestamp"`
	Events    []TimelineEvent  `json:"events"`
}

type TimelineEvent struct {
	Type          string `json:"type"`
	Timestamp     int    `json:"timestamp"`
	ParticipantID int    `json:"participantId,omitempty"`
	ItemID        int    `json:"itemId,omitempty"`
}

// Items that should be excluded from build order (consumables, components, etc.)
var ExcludedItems = map[int]bool{
	// Potions and consumables
	2003: true, // Health Potion
	2031: true, // Refillable Potion
	2033: true, // Corrupting Potion
	2055: true, // Control Ward
	2138: true, // Elixir of Iron
	2139: true, // Elixir of Sorcery
	2140: true, // Elixir of Wrath

	// Trinkets
	3340: true, // Stealth Ward
	3341: true, // Sweeping Lens
	3363: true, // Farsight Alteration
	3364: true, // Oracle Lens

	// Boots components
	1001: true, // Boots

	// Common early components (optional - can remove if you want to track all purchases)
	1036: true, // Long Sword
	1037: true, // Pickaxe
	1038: true, // BF Sword
	1052: true, // Amplifying Tome
	1058: true, // Needlessly Large Rod
	1026: true, // Blasting Wand
	1027: true, // Sapphire Crystal
	1028: true, // Ruby Crystal
	1029: true, // Cloth Armor
	1031: true, // Chain Vest
	1033: true, // Null-Magic Mantle
	1057: true, // Negatron Cloak
	1042: true, // Dagger
	1043: true, // Recurve Bow
	1018: true, // Cloak of Agility
	1053: true, // Vampiric Scepter
	1054: true, // Doran's Shield
	1055: true, // Doran's Blade
	1056: true, // Doran's Ring
	1082: true, // Dark Seal
	1083: true, // Cull
}

// IsCompletedItem returns true if the item is a completed item worth tracking
func IsCompletedItem(itemID int) bool {
	// Item ID 0 means empty slot
	if itemID == 0 {
		return false
	}
	// Exclude known consumables/components
	if ExcludedItems[itemID] {
		return false
	}
	// Generally, completed items have IDs >= 3000 (with some exceptions)
	// This is a heuristic - adjust as needed
	return itemID >= 2000
}
