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
