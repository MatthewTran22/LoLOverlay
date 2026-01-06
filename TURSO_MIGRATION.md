# Turso Migration Plan for Website

## Overview
Migrate the website from local SQLite (better-sqlite3) to Turso DB (read-only).
- **Website**: Read-only access to Turso
- **Data Analyzer**: Will write to Turso (separate task)
- **Desktop App**: Unchanged (keeps local SQLite + JSON download)

## Environment Setup (DONE)
- `.env.local` created with `TURSO_DATABASE_URL` and `TURSO_AUTH_TOKEN`

## Tasks

### 1. Install @libsql/client package
```bash
cd website
npm install @libsql/client
npm uninstall better-sqlite3 @types/better-sqlite3
```

### 2. Replace `website/src/lib/db.ts`
Replace entire file with this simplified read-only version:

```typescript
import { createClient, Client } from "@libsql/client";

// Database singleton
let db: Client | null = null;

export function getDb(): Client {
  if (!db) {
    db = createClient({
      url: process.env.TURSO_DATABASE_URL!,
      authToken: process.env.TURSO_AUTH_TOKEN,
    });
    console.log("[Stats] Turso client initialized");
  }
  return db;
}

export async function getCurrentPatch(): Promise<string> {
  try {
    const client = getDb();
    const result = await client.execute("SELECT patch FROM data_version WHERE id = 1");
    return (result.rows[0]?.patch as string) || "";
  } catch {
    return "";
  }
}

export async function hasData(): Promise<boolean> {
  try {
    const client = getDb();
    const result = await client.execute("SELECT COUNT(*) as count FROM champion_stats");
    return (result.rows[0]?.count as number) > 0;
  } catch {
    return false;
  }
}
```

### 3. Update `website/src/lib/stats.ts`
Convert all functions to async. Key changes:
- Import `getDb` (no longer `getCurrentPatch` - get it from db.ts async)
- All functions become `async`
- Replace `db.prepare(...).get()` with `await client.execute(...)`
- Replace `db.prepare(...).all()` with `await client.execute(...)`
- Access results via `result.rows[0]` instead of direct return

Example conversion:
```typescript
// BEFORE (sync, better-sqlite3)
export function fetchChampionStats(championId: number, role: string): ChampionWinRate | null {
  const db = getDb();
  const row = db.prepare(`SELECT ...`).get(championId, position) as { ... };
  return row;
}

// AFTER (async, @libsql/client)
export async function fetchChampionStats(championId: number, role: string): Promise<ChampionWinRate | null> {
  const client = getDb();
  const result = await client.execute({
    sql: `SELECT ...`,
    args: [championId, position]
  });
  const row = result.rows[0];
  if (!row) return null;
  return { ... };
}
```

### 4. Update pages/components that use stats functions
Files to update (add `await` to all stats function calls):
- `website/src/app/stats/page.tsx`
- `website/src/app/stats/[championId]/page.tsx`
- `website/src/app/admin/page.tsx`
- `website/src/app/api/stats/route.ts`
- `website/src/app/api/meta/route.ts`
- `website/src/app/api/champions/[id]/route.ts`
- `website/src/app/api/matchups/[id]/route.ts`

### 5. Remove admin update functionality
The admin page currently has "Check for Updates" and "Force Update" buttons.
These should be removed or disabled since the website is now read-only.
Data updates will come from the Data Analyzer.

### 6. Update `.gitignore`
Add `.env.local` if not already there.

## Schema Reference (for Data Analyzer later)
```sql
CREATE TABLE data_version (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  patch TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE champion_stats (
  patch TEXT NOT NULL,
  champion_id INTEGER NOT NULL,
  team_position TEXT NOT NULL,
  wins INTEGER NOT NULL DEFAULT 0,
  matches INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (patch, champion_id, team_position)
);

CREATE TABLE champion_items (
  patch TEXT NOT NULL,
  champion_id INTEGER NOT NULL,
  team_position TEXT NOT NULL,
  item_id INTEGER NOT NULL,
  wins INTEGER NOT NULL DEFAULT 0,
  matches INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (patch, champion_id, team_position, item_id)
);

CREATE TABLE champion_item_slots (
  patch TEXT NOT NULL,
  champion_id INTEGER NOT NULL,
  team_position TEXT NOT NULL,
  item_id INTEGER NOT NULL,
  build_slot INTEGER NOT NULL,
  wins INTEGER NOT NULL DEFAULT 0,
  matches INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (patch, champion_id, team_position, item_id, build_slot)
);

CREATE TABLE champion_matchups (
  patch TEXT NOT NULL,
  champion_id INTEGER NOT NULL,
  team_position TEXT NOT NULL,
  enemy_champion_id INTEGER NOT NULL,
  wins INTEGER NOT NULL DEFAULT 0,
  matches INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (patch, champion_id, team_position, enemy_champion_id)
);
```

## Notes
- Turso client is async (unlike better-sqlite3 which is sync)
- Use `client.execute({ sql: "...", args: [...] })` for parameterized queries
- Results come back as `result.rows` array
- Column access: `row.column_name` or `row['column_name']`
