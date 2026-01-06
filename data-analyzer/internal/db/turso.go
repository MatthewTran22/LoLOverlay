package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// TursoClient wraps a connection to Turso
type TursoClient struct {
	db *sql.DB
}

// NewTursoClient creates a new Turso client
func NewTursoClient(url, authToken string) (*TursoClient, error) {
	connStr := url
	if authToken != "" {
		connStr = fmt.Sprintf("%s?authToken=%s", url, authToken)
	}

	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Turso: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping Turso: %w", err)
	}

	return &TursoClient{db: db}, nil
}

// Close closes the Turso connection
func (c *TursoClient) Close() error {
	return c.db.Close()
}

// CreateTables creates the required tables if they don't exist
func (c *TursoClient) CreateTables(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS data_version (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			patch TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS champion_stats (
			patch TEXT NOT NULL,
			champion_id INTEGER NOT NULL,
			team_position TEXT NOT NULL,
			wins INTEGER NOT NULL DEFAULT 0,
			matches INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (patch, champion_id, team_position)
		)`,
		`CREATE TABLE IF NOT EXISTS champion_items (
			patch TEXT NOT NULL,
			champion_id INTEGER NOT NULL,
			team_position TEXT NOT NULL,
			item_id INTEGER NOT NULL,
			wins INTEGER NOT NULL DEFAULT 0,
			matches INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (patch, champion_id, team_position, item_id)
		)`,
		`CREATE TABLE IF NOT EXISTS champion_item_slots (
			patch TEXT NOT NULL,
			champion_id INTEGER NOT NULL,
			team_position TEXT NOT NULL,
			item_id INTEGER NOT NULL,
			build_slot INTEGER NOT NULL,
			wins INTEGER NOT NULL DEFAULT 0,
			matches INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (patch, champion_id, team_position, item_id, build_slot)
		)`,
		`CREATE TABLE IF NOT EXISTS champion_matchups (
			patch TEXT NOT NULL,
			champion_id INTEGER NOT NULL,
			team_position TEXT NOT NULL,
			enemy_champion_id INTEGER NOT NULL,
			wins INTEGER NOT NULL DEFAULT 0,
			matches INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (patch, champion_id, team_position, enemy_champion_id)
		)`,
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_champion_stats_position ON champion_stats(team_position)`,
		`CREATE INDEX IF NOT EXISTS idx_champion_stats_champ_pos ON champion_stats(champion_id, team_position)`,
		`CREATE INDEX IF NOT EXISTS idx_champion_items_champ_pos ON champion_items(champion_id, team_position)`,
		`CREATE INDEX IF NOT EXISTS idx_champion_item_slots_champ_pos ON champion_item_slots(champion_id, team_position)`,
		`CREATE INDEX IF NOT EXISTS idx_champion_item_slots_champ_pos_slot ON champion_item_slots(champion_id, team_position, build_slot)`,
		`CREATE INDEX IF NOT EXISTS idx_champion_matchups_champ_pos ON champion_matchups(champion_id, team_position)`,
		`CREATE INDEX IF NOT EXISTS idx_champion_matchups_enemy ON champion_matchups(champion_id, team_position, enemy_champion_id)`,
	}

	for _, query := range queries {
		if _, err := c.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// ClearData deletes all existing data
func (c *TursoClient) ClearData(ctx context.Context) error {
	tables := []string{"data_version", "champion_stats", "champion_items", "champion_item_slots", "champion_matchups"}
	for _, table := range tables {
		if _, err := c.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			return fmt.Errorf("failed to clear %s: %w", table, err)
		}
	}
	return nil
}

// SetDataVersion sets the current patch version
func (c *TursoClient) SetDataVersion(ctx context.Context, patch string) error {
	_, err := c.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO data_version (id, patch, updated_at) VALUES (1, ?, ?)`,
		patch, time.Now().UTC().Format(time.RFC3339))
	return err
}

// ChampionStat represents a champion stat row
type ChampionStat struct {
	Patch        string
	ChampionID   int
	TeamPosition string
	Wins         int
	Matches      int
}

// ChampionItem represents a champion item row
type ChampionItem struct {
	Patch        string
	ChampionID   int
	TeamPosition string
	ItemID       int
	Wins         int
	Matches      int
}

// ChampionItemSlot represents a champion item slot row
type ChampionItemSlot struct {
	Patch        string
	ChampionID   int
	TeamPosition string
	ItemID       int
	BuildSlot    int
	Wins         int
	Matches      int
}

// ChampionMatchup represents a champion matchup row
type ChampionMatchup struct {
	Patch           string
	ChampionID      int
	TeamPosition    string
	EnemyChampionID int
	Wins            int
	Matches         int
}

// InsertChampionStats inserts champion stats in batches
func (c *TursoClient) InsertChampionStats(ctx context.Context, stats []ChampionStat) error {
	const batchSize = 100

	for i := 0; i < len(stats); i += batchSize {
		end := i + batchSize
		if end > len(stats) {
			end = len(stats)
		}
		batch := stats[i:end]

		tx, err := c.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO champion_stats (patch, champion_id, team_position, wins, matches) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, s := range batch {
			if _, err := stmt.ExecContext(ctx, s.Patch, s.ChampionID, s.TeamPosition, s.Wins, s.Matches); err != nil {
				stmt.Close()
				tx.Rollback()
				return err
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// InsertChampionItems inserts champion items in batches
func (c *TursoClient) InsertChampionItems(ctx context.Context, items []ChampionItem) error {
	const batchSize = 100

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]

		tx, err := c.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO champion_items (patch, champion_id, team_position, item_id, wins, matches) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, item := range batch {
			if _, err := stmt.ExecContext(ctx, item.Patch, item.ChampionID, item.TeamPosition, item.ItemID, item.Wins, item.Matches); err != nil {
				stmt.Close()
				tx.Rollback()
				return err
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// InsertChampionItemSlots inserts champion item slots in batches
func (c *TursoClient) InsertChampionItemSlots(ctx context.Context, slots []ChampionItemSlot) error {
	const batchSize = 100

	for i := 0; i < len(slots); i += batchSize {
		end := i + batchSize
		if end > len(slots) {
			end = len(slots)
		}
		batch := slots[i:end]

		tx, err := c.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO champion_item_slots (patch, champion_id, team_position, item_id, build_slot, wins, matches) VALUES (?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, slot := range batch {
			if _, err := stmt.ExecContext(ctx, slot.Patch, slot.ChampionID, slot.TeamPosition, slot.ItemID, slot.BuildSlot, slot.Wins, slot.Matches); err != nil {
				stmt.Close()
				tx.Rollback()
				return err
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// InsertChampionMatchups inserts champion matchups in batches
func (c *TursoClient) InsertChampionMatchups(ctx context.Context, matchups []ChampionMatchup) error {
	const batchSize = 100

	for i := 0; i < len(matchups); i += batchSize {
		end := i + batchSize
		if end > len(matchups) {
			end = len(matchups)
		}
		batch := matchups[i:end]

		tx, err := c.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO champion_matchups (patch, champion_id, team_position, enemy_champion_id, wins, matches) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, m := range batch {
			if _, err := stmt.ExecContext(ctx, m.Patch, m.ChampionID, m.TeamPosition, m.EnemyChampionID, m.Wins, m.Matches); err != nil {
				stmt.Close()
				tx.Rollback()
				return err
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}
