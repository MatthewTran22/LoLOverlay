import { createClient } from "@libsql/client";
import Database from "better-sqlite3";

const LOCAL_DB_PATH = "C:\\Users\\yourl\\AppData\\Roaming\\GhostDraft\\stats.db";
const BATCH_SIZE = 100;

// Load from .env.local
import { config } from "dotenv";
config({ path: ".env.local" });

const turso = createClient({
  url: process.env.TURSO_DATABASE_URL,
  authToken: process.env.TURSO_AUTH_TOKEN,
});

const local = new Database(LOCAL_DB_PATH, { readonly: true });

async function batchInsert(tableName, rows, buildArgs) {
  for (let i = 0; i < rows.length; i += BATCH_SIZE) {
    const batch = rows.slice(i, i + BATCH_SIZE);
    const statements = batch.map(row => buildArgs(row));
    await turso.batch(statements);
    process.stdout.write(`\r   - ${Math.min(i + BATCH_SIZE, rows.length)}/${rows.length} rows`);
  }
  console.log();
}

async function pushToTurso() {
  console.log("Connected to local DB:", LOCAL_DB_PATH);
  console.log("Pushing to Turso:", process.env.TURSO_DATABASE_URL);
  console.log("Batch size:", BATCH_SIZE);

  // Create tables
  console.log("\n1. Creating tables...");
  await turso.executeMultiple(`
    CREATE TABLE IF NOT EXISTS data_version (
      id INTEGER PRIMARY KEY CHECK (id = 1),
      patch TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS champion_stats (
      patch TEXT NOT NULL,
      champion_id INTEGER NOT NULL,
      team_position TEXT NOT NULL,
      wins INTEGER NOT NULL DEFAULT 0,
      matches INTEGER NOT NULL DEFAULT 0,
      PRIMARY KEY (patch, champion_id, team_position)
    );

    CREATE TABLE IF NOT EXISTS champion_items (
      patch TEXT NOT NULL,
      champion_id INTEGER NOT NULL,
      team_position TEXT NOT NULL,
      item_id INTEGER NOT NULL,
      wins INTEGER NOT NULL DEFAULT 0,
      matches INTEGER NOT NULL DEFAULT 0,
      PRIMARY KEY (patch, champion_id, team_position, item_id)
    );

    CREATE TABLE IF NOT EXISTS champion_item_slots (
      patch TEXT NOT NULL,
      champion_id INTEGER NOT NULL,
      team_position TEXT NOT NULL,
      item_id INTEGER NOT NULL,
      build_slot INTEGER NOT NULL,
      wins INTEGER NOT NULL DEFAULT 0,
      matches INTEGER NOT NULL DEFAULT 0,
      PRIMARY KEY (patch, champion_id, team_position, item_id, build_slot)
    );

    CREATE TABLE IF NOT EXISTS champion_matchups (
      patch TEXT NOT NULL,
      champion_id INTEGER NOT NULL,
      team_position TEXT NOT NULL,
      enemy_champion_id INTEGER NOT NULL,
      wins INTEGER NOT NULL DEFAULT 0,
      matches INTEGER NOT NULL DEFAULT 0,
      PRIMARY KEY (patch, champion_id, team_position, enemy_champion_id)
    );
  `);

  // Clear existing data
  console.log("2. Clearing existing data...");
  await turso.executeMultiple(`
    DELETE FROM data_version;
    DELETE FROM champion_stats;
    DELETE FROM champion_items;
    DELETE FROM champion_item_slots;
    DELETE FROM champion_matchups;
  `);

  // Copy data_version
  console.log("3. Copying data_version...");
  const version = local.prepare("SELECT * FROM data_version").all();
  for (const row of version) {
    await turso.execute({
      sql: "INSERT INTO data_version (id, patch, updated_at) VALUES (?, ?, ?)",
      args: [row.id, row.patch, row.updated_at],
    });
  }
  console.log(`   - ${version.length} rows`);

  // Copy champion_stats
  console.log("4. Copying champion_stats...");
  const stats = local.prepare("SELECT * FROM champion_stats").all();
  await batchInsert("champion_stats", stats, (row) => ({
    sql: "INSERT INTO champion_stats (patch, champion_id, team_position, wins, matches) VALUES (?, ?, ?, ?, ?)",
    args: [row.patch, row.champion_id, row.team_position, row.wins, row.matches],
  }));

  // Copy champion_items
  console.log("5. Copying champion_items...");
  const items = local.prepare("SELECT * FROM champion_items").all();
  await batchInsert("champion_items", items, (row) => ({
    sql: "INSERT INTO champion_items (patch, champion_id, team_position, item_id, wins, matches) VALUES (?, ?, ?, ?, ?, ?)",
    args: [row.patch, row.champion_id, row.team_position, row.item_id, row.wins, row.matches],
  }));

  // Copy champion_item_slots
  console.log("6. Copying champion_item_slots...");
  const slots = local.prepare("SELECT * FROM champion_item_slots").all();
  await batchInsert("champion_item_slots", slots, (row) => ({
    sql: "INSERT INTO champion_item_slots (patch, champion_id, team_position, item_id, build_slot, wins, matches) VALUES (?, ?, ?, ?, ?, ?, ?)",
    args: [row.patch, row.champion_id, row.team_position, row.item_id, row.build_slot, row.wins, row.matches],
  }));

  // Copy champion_matchups
  console.log("7. Copying champion_matchups...");
  const matchups = local.prepare("SELECT * FROM champion_matchups").all();
  await batchInsert("champion_matchups", matchups, (row) => ({
    sql: "INSERT INTO champion_matchups (patch, champion_id, team_position, enemy_champion_id, wins, matches) VALUES (?, ?, ?, ?, ?, ?)",
    args: [row.patch, row.champion_id, row.team_position, row.enemy_champion_id, row.wins, row.matches],
  }));

  console.log("\nDone! Data pushed to Turso.");

  local.close();
}

pushToTurso().catch(console.error);
