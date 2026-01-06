import { createClient } from "@libsql/client";
import { config } from "dotenv";
config({ path: ".env.local" });

const turso = createClient({
  url: process.env.TURSO_DATABASE_URL,
  authToken: process.env.TURSO_AUTH_TOKEN,
});

async function addIndexes() {
  console.log("Adding indexes to Turso...");

  await turso.executeMultiple(`
    CREATE INDEX IF NOT EXISTS idx_champion_stats_position ON champion_stats(team_position);
    CREATE INDEX IF NOT EXISTS idx_champion_stats_champ_pos ON champion_stats(champion_id, team_position);

    CREATE INDEX IF NOT EXISTS idx_champion_items_champ_pos ON champion_items(champion_id, team_position);

    CREATE INDEX IF NOT EXISTS idx_champion_item_slots_champ_pos ON champion_item_slots(champion_id, team_position);
    CREATE INDEX IF NOT EXISTS idx_champion_item_slots_champ_pos_slot ON champion_item_slots(champion_id, team_position, build_slot);

    CREATE INDEX IF NOT EXISTS idx_champion_matchups_champ_pos ON champion_matchups(champion_id, team_position);
    CREATE INDEX IF NOT EXISTS idx_champion_matchups_enemy ON champion_matchups(champion_id, team_position, enemy_champion_id);
  `);

  console.log("Done!");
}

addIndexes().catch(console.error);
