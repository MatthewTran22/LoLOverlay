package data

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"os"

	_ "modernc.org/sqlite"
)

// ChampionDB manages the champion database
type ChampionDB struct {
	db *sql.DB
}

// ChampionInfo holds champion metadata
type ChampionInfo struct {
	Name       string
	DamageType string
	RoleTags   string
}

// NewChampionDB creates and initializes the champion database
func NewChampionDB() (*ChampionDB, error) {
	// Get user's app data directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	dbDir := filepath.Join(configDir, "GhostDraft")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "champions.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	cdb := &ChampionDB{db: db}
	if err := cdb.init(); err != nil {
		db.Close()
		return nil, err
	}

	return cdb, nil
}

// init creates the schema and populates data
func (c *ChampionDB) init() error {
	// Create table
	_, err := c.db.Exec(`
		CREATE TABLE IF NOT EXISTS champions (
			name TEXT PRIMARY KEY,
			damage_type TEXT NOT NULL,
			role_tags TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Check if data exists
	var count int
	c.db.QueryRow("SELECT COUNT(*) FROM champions").Scan(&count)
	if count > 0 {
		return nil // Already populated
	}

	// Populate with champion data
	return c.populate()
}

// populate inserts all champion data
func (c *ChampionDB) populate() error {
	champions := []ChampionInfo{
		{"Aatrox", "AD", "Bruiser, Engage"},
		{"Ahri", "AP", "Burst, Poke, Engage (Light)"},
		{"Akali", "AP", "Burst"},
		{"Akshan", "AD", "Burst, Poke"},
		{"Alistar", "AP", "Tank, Engage"},
		{"Ambessa", "AD", "Bruiser, Engage, Burst"},
		{"Amumu", "AP", "Tank, Engage, Burst"},
		{"Anivia", "AP", "Burst, Poke, Zone"},
		{"Annie", "AP", "Burst, Engage"},
		{"Aphelios", "AD", "Burst, Poke"},
		{"Ashe", "AD", "Poke, Engage"},
		{"Aurelion Sol", "AP", "Burst, Poke"},
		{"Aurora", "AP", "Burst, Poke, Mobility"},
		{"Azir", "AP", "Poke, Engage"},
		{"Bard", "AP", "Poke, Engage"},
		{"Bel'Veth", "AD", "Bruiser, Engage"},
		{"Blitzcrank", "AP", "Tank, Engage (Pick)"},
		{"Brand", "AP", "Burst, Poke"},
		{"Braum", "AP", "Tank, Engage (Counter)"},
		{"Briar", "AD", "Bruiser, Engage, Burst"},
		{"Caitlyn", "AD", "Poke, Burst"},
		{"Camille", "AD", "Bruiser, Engage, Burst"},
		{"Cassiopeia", "AP", "Burst, Engage (Counter)"},
		{"Cho'Gath", "AP/Tank", "Tank, Burst, Poke"},
		{"Corki", "AD/AP", "Poke, Burst"},
		{"Darius", "AD", "Bruiser"},
		{"Diana", "AP", "Burst, Engage, Bruiser"},
		{"Dr. Mundo", "AD/HP", "Tank, Poke"},
		{"Draven", "AD", "Burst"},
		{"Ekko", "AP", "Burst, Poke"},
		{"Elise", "AP", "Burst, Engage (Pick)"},
		{"Evelynn", "AP", "Burst"},
		{"Ezreal", "AD/AP", "Poke, Burst"},
		{"Fiddlesticks", "AP", "Burst, Engage"},
		{"Fiora", "AD", "Bruiser, Burst"},
		{"Fizz", "AP", "Burst, Engage"},
		{"Galio", "AP", "Tank, Engage, Burst"},
		{"Gangplank", "AD", "Burst, Poke"},
		{"Garen", "AD", "Bruiser, Burst"},
		{"Gnar", "AD", "Bruiser, Engage, Poke"},
		{"Gragas", "AP", "Tank, Burst, Engage, Poke"},
		{"Graves", "AD", "Burst, Bruiser"},
		{"Gwen", "AP", "Bruiser, Burst"},
		{"Hecarim", "AD", "Bruiser, Engage, Burst"},
		{"Heimerdinger", "AP", "Poke, Burst"},
		{"Hwei", "AP", "Poke, Burst, Engage"},
		{"Illaoi", "AD", "Bruiser, Poke"},
		{"Irelia", "AD", "Bruiser, Engage, Burst"},
		{"Ivern", "AP", "Support, Engage"},
		{"Janna", "AP", "Poke, Disengage"},
		{"Jarvan IV", "AD", "Bruiser, Engage, Burst"},
		{"Jax", "AD/AP", "Bruiser, Engage, Burst"},
		{"Jayce", "AD", "Poke, Burst"},
		{"Jhin", "AD", "Poke, Burst, Engage (Long Range)"},
		{"Jinx", "AD", "Poke, Burst"},
		{"K'Sante", "AD/Tank", "Tank, Bruiser, Engage"},
		{"Kai'Sa", "AD/AP", "Burst, Poke, Engage"},
		{"Kalista", "AD", "Burst, Engage"},
		{"Karma", "AP", "Poke, Engage (Speed)"},
		{"Karthus", "AP", "Burst, Poke"},
		{"Kassadin", "AP", "Burst"},
		{"Katarina", "AP/AD", "Burst"},
		{"Kayle", "AP/AD", "Burst"},
		{"Kayn", "AD", "Burst, Bruiser, Engage"},
		{"Kennen", "AP", "Burst, Engage, Poke"},
		{"Kha'Zix", "AD", "Burst, Poke"},
		{"Kindred", "AD", "Burst"},
		{"Kled", "AD", "Bruiser, Engage"},
		{"Kog'Maw", "AD/AP", "Poke, Burst"},
		{"LeBlanc", "AP", "Burst, Poke"},
		{"Lee Sin", "AD", "Bruiser, Engage, Burst"},
		{"Leona", "AP", "Tank, Engage"},
		{"Lillia", "AP", "Bruiser, Engage, Poke"},
		{"Lissandra", "AP", "Burst, Engage"},
		{"Lucian", "AD", "Burst, Poke"},
		{"Lulu", "AP", "Poke, Disengage"},
		{"Lux", "AP", "Burst, Poke, Engage (Pick)"},
		{"Malphite", "AP/Tank", "Tank, Engage, Burst"},
		{"Malzahar", "AP", "Burst, Engage (Pick)"},
		{"Maokai", "AP/Tank", "Tank, Engage, Poke"},
		{"Master Yi", "AD", "Burst, Bruiser"},
		{"Mel", "AP", "Burst, Poke"},
		{"Milio", "AP", "Poke, Disengage"},
		{"Miss Fortune", "AD", "Burst, Poke"},
		{"Mordekaiser", "AP", "Bruiser, Poke (Pull)"},
		{"Morgana", "AP", "Poke, Engage (Pick)"},
		{"Naafiri", "AD", "Burst, Poke, Engage"},
		{"Nami", "AP", "Poke, Engage"},
		{"Nasus", "AD", "Bruiser, Tank"},
		{"Nautilus", "AP", "Tank, Engage"},
		{"Neeko", "AP", "Burst, Engage, Poke"},
		{"Nidalee", "AP", "Poke, Burst"},
		{"Nilah", "AD", "Bruiser, Engage, Burst"},
		{"Nocturne", "AD", "Burst, Engage"},
		{"Nunu & Willump", "AP/Tank", "Tank, Engage, Burst"},
		{"Olaf", "AD", "Bruiser, Poke"},
		{"Orianna", "AP", "Burst, Poke, Engage"},
		{"Ornn", "Tank", "Tank, Engage, Burst"},
		{"Pantheon", "AD", "Bruiser, Burst, Engage"},
		{"Poppy", "Tank", "Tank, Engage, Burst"},
		{"Pyke", "AD", "Burst, Engage (Pick)"},
		{"Qiyana", "AD", "Burst, Engage"},
		{"Quinn", "AD", "Burst, Poke"},
		{"Rakan", "AP", "Engage, Mobility"},
		{"Rammus", "Tank", "Tank, Engage"},
		{"Rek'Sai", "AD", "Bruiser, Engage, Burst"},
		{"Rell", "Tank", "Tank, Engage"},
		{"Renata Glasc", "AP", "Poke, Disengage (Counter-Engage)"},
		{"Renekton", "AD", "Bruiser, Engage, Burst"},
		{"Rengar", "AD", "Burst, Engage"},
		{"Riven", "AD", "Bruiser, Burst, Engage"},
		{"Rumble", "AP", "Bruiser, Burst, Poke"},
		{"Ryze", "AP", "Burst, Poke"},
		{"Samira", "AD", "Burst, Engage"},
		{"Sejuani", "Tank", "Tank, Engage"},
		{"Senna", "AD", "Poke, Burst, Engage (Root)"},
		{"Seraphine", "AP", "Poke, Engage, Burst"},
		{"Sett", "AD", "Bruiser, Tank, Engage"},
		{"Shaco", "AD/AP", "Burst, Poke (AP)"},
		{"Shen", "Tank", "Tank, Engage"},
		{"Shyvana", "AP/AD", "Bruiser, Burst, Poke (AP)"},
		{"Singed", "AP", "Tank, Bruiser, Engage"},
		{"Sion", "AD/Tank", "Tank, Engage, Poke"},
		{"Sivir", "AD", "Poke, Burst, Engage (Ult)"},
		{"Skarner", "Tank", "Tank, Engage, Bruiser"},
		{"Smolder", "AD", "Poke, Burst"},
		{"Sona", "AP", "Poke, Engage"},
		{"Soraka", "AP", "Poke, Disengage"},
		{"Swain", "AP", "Bruiser, Engage, Poke"},
		{"Sylas", "AP", "Bruiser, Burst, Engage"},
		{"Syndra", "AP", "Burst, Poke, Engage (Pick)"},
		{"Tahm Kench", "Tank", "Tank, Poke, Engage"},
		{"Taliyah", "AP", "Burst, Poke, Engage (Wall)"},
		{"Talon", "AD", "Burst, Poke"},
		{"Taric", "AP/Tank", "Tank, Engage"},
		{"Teemo", "AP", "Poke, Burst"},
		{"Thresh", "AP/Tank", "Tank, Engage (Pick)"},
		{"Tristana", "AD", "Burst, Poke"},
		{"Trundle", "AD", "Bruiser, Tank"},
		{"Tryndamere", "AD", "Bruiser, Burst"},
		{"Twisted Fate", "AP/AD", "Burst, Poke, Engage (Pick)"},
		{"Twitch", "AD/AP", "Burst, Poke"},
		{"Udyr", "AD/AP", "Bruiser, Engage"},
		{"Urgot", "AD", "Bruiser, Poke"},
		{"Varus", "AD/AP", "Poke, Burst, Engage"},
		{"Vayne", "AD", "Burst"},
		{"Veigar", "AP", "Burst, Poke, Zone"},
		{"Vel'Koz", "AP", "Poke, Burst"},
		{"Vex", "AP", "Burst, Engage"},
		{"Vi", "AD", "Bruiser, Engage, Burst"},
		{"Viego", "AD", "Bruiser, Burst"},
		{"Viktor", "AP", "Burst, Poke, Zone"},
		{"Vladimir", "AP", "Burst, Bruiser"},
		{"Volibear", "AD/AP", "Bruiser, Engage"},
		{"Warwick", "AD/AP", "Bruiser, Engage"},
		{"Wukong", "AD", "Bruiser, Engage, Burst"},
		{"Xayah", "AD", "Burst, Poke"},
		{"Xerath", "AP", "Poke, Burst"},
		{"Xin Zhao", "AD", "Bruiser, Engage, Burst"},
		{"Yasuo", "AD", "Burst, Bruiser"},
		{"Yone", "AD", "Burst, Bruiser, Engage"},
		{"Yorick", "AD", "Bruiser, Poke"},
		{"Yuumi", "AP", "Poke, Engage"},
		{"Zac", "Tank", "Tank, Engage"},
		{"Zed", "AD", "Burst, Poke"},
		{"Zeri", "AD", "Poke, Burst"},
		{"Ziggs", "AP", "Poke, Burst"},
		{"Zilean", "AP", "Poke, Engage (Speed)"},
		{"Zoe", "AP", "Burst, Poke, Engage (Sleep)"},
		{"Zyra", "AP", "Poke, Burst, Engage"},
	}

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO champions (name, damage_type, role_tags) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, champ := range champions {
		_, err := stmt.Exec(champ.Name, champ.DamageType, champ.RoleTags)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetChampion returns info for a champion by name
func (c *ChampionDB) GetChampion(name string) (*ChampionInfo, error) {
	var info ChampionInfo
	err := c.db.QueryRow(
		"SELECT name, damage_type, role_tags FROM champions WHERE name = ?",
		name,
	).Scan(&info.Name, &info.DamageType, &info.RoleTags)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetDamageType returns the damage type for a champion
func (c *ChampionDB) GetDamageType(name string) string {
	var dmgType string
	err := c.db.QueryRow("SELECT damage_type FROM champions WHERE name = ?", name).Scan(&dmgType)
	if err != nil {
		return "Unknown"
	}
	return dmgType
}

// GetRoleTags returns the role tags for a champion
func (c *ChampionDB) GetRoleTags(name string) string {
	var tags string
	err := c.db.QueryRow("SELECT role_tags FROM champions WHERE name = ?", name).Scan(&tags)
	if err != nil {
		return ""
	}
	return tags
}

// Close closes the database connection
func (c *ChampionDB) Close() error {
	return c.db.Close()
}
