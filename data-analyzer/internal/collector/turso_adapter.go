package collector

import (
	"context"
	"fmt"
	"log"

	"data-analyzer/internal/db"
)

// TursoDataPusher adapts TursoClient to the DataPusher interface for use with TursoPusher
type TursoDataPusher struct {
	client *db.TursoClient
}

// NewTursoDataPusher creates a new TursoDataPusher
func NewTursoDataPusher(client *db.TursoClient) *TursoDataPusher {
	return &TursoDataPusher{client: client}
}

// PushAggData pushes aggregated data to Turso
func (p *TursoDataPusher) PushAggData(ctx context.Context, data *AggData) error {
	if data == nil {
		return nil
	}

	log.Printf("[TursoPusher] Starting push: %d champion stats, %d item stats, %d item slot stats, %d matchup stats",
		len(data.ChampionStats), len(data.ItemStats), len(data.ItemSlotStats), len(data.MatchupStats))

	// Ensure tables exist
	if err := p.client.CreateTables(ctx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Drop indexes for faster bulk insert
	if err := p.client.DropIndexes(ctx); err != nil {
		log.Printf("[TursoPusher] Warning: failed to drop indexes: %v", err)
	}

	// Push champion stats
	if len(data.ChampionStats) > 0 {
		stats := make([]db.ChampionStat, 0, len(data.ChampionStats))
		for k, v := range data.ChampionStats {
			stats = append(stats, db.ChampionStat{
				Patch:        k.Patch,
				ChampionID:   k.ChampionID,
				TeamPosition: k.TeamPosition,
				Wins:         v.Wins,
				Matches:      v.Matches,
			})
		}
		if err := p.client.InsertChampionStats(ctx, stats); err != nil {
			return fmt.Errorf("failed to insert champion stats: %w", err)
		}
		log.Printf("[TursoPusher] Inserted %d champion stats", len(stats))
	}

	// Push item stats
	if len(data.ItemStats) > 0 {
		items := make([]db.ChampionItem, 0, len(data.ItemStats))
		for k, v := range data.ItemStats {
			items = append(items, db.ChampionItem{
				Patch:        k.Patch,
				ChampionID:   k.ChampionID,
				TeamPosition: k.TeamPosition,
				ItemID:       k.ItemID,
				Wins:         v.Wins,
				Matches:      v.Matches,
			})
		}
		if err := p.client.InsertChampionItems(ctx, items); err != nil {
			return fmt.Errorf("failed to insert champion items: %w", err)
		}
		log.Printf("[TursoPusher] Inserted %d item stats", len(items))
	}

	// Push item slot stats
	if len(data.ItemSlotStats) > 0 {
		slots := make([]db.ChampionItemSlot, 0, len(data.ItemSlotStats))
		for k, v := range data.ItemSlotStats {
			slots = append(slots, db.ChampionItemSlot{
				Patch:        k.Patch,
				ChampionID:   k.ChampionID,
				TeamPosition: k.TeamPosition,
				ItemID:       k.ItemID,
				BuildSlot:    k.BuildSlot,
				Wins:         v.Wins,
				Matches:      v.Matches,
			})
		}
		if err := p.client.InsertChampionItemSlots(ctx, slots); err != nil {
			return fmt.Errorf("failed to insert champion item slots: %w", err)
		}
		log.Printf("[TursoPusher] Inserted %d item slot stats", len(slots))
	}

	// Push matchup stats
	if len(data.MatchupStats) > 0 {
		matchups := make([]db.ChampionMatchup, 0, len(data.MatchupStats))
		for k, v := range data.MatchupStats {
			matchups = append(matchups, db.ChampionMatchup{
				Patch:           k.Patch,
				ChampionID:      k.ChampionID,
				TeamPosition:    k.TeamPosition,
				EnemyChampionID: k.EnemyChampionID,
				Wins:            v.Wins,
				Matches:         v.Matches,
			})
		}
		if err := p.client.InsertChampionMatchups(ctx, matchups); err != nil {
			return fmt.Errorf("failed to insert champion matchups: %w", err)
		}
		log.Printf("[TursoPusher] Inserted %d matchup stats", len(matchups))
	}

	// Update data version
	if data.DetectedPatch != "" {
		if err := p.client.SetDataVersion(ctx, data.DetectedPatch); err != nil {
			log.Printf("[TursoPusher] Warning: failed to set data version: %v", err)
		}
	}

	// Recreate indexes
	if err := p.client.CreateIndexes(ctx); err != nil {
		log.Printf("[TursoPusher] Warning: failed to recreate indexes: %v", err)
	}

	log.Printf("[TursoPusher] Push complete for patch %s", data.DetectedPatch)
	return nil
}
