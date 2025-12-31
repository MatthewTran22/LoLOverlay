-- Aggregated Champion Stats
-- PK: (patch, champion_id, team_position)
CREATE TABLE IF NOT EXISTS champion_stats (
    patch VARCHAR(10) NOT NULL,
    champion_id INT NOT NULL,
    team_position VARCHAR(20) NOT NULL,
    wins INT NOT NULL DEFAULT 0,
    matches INT NOT NULL DEFAULT 0,
    PRIMARY KEY (patch, champion_id, team_position)
);

-- Aggregated Item Stats per Champion
-- PK: (patch, champion_id, team_position, item_id)
CREATE TABLE IF NOT EXISTS champion_items (
    patch VARCHAR(10) NOT NULL,
    champion_id INT NOT NULL,
    team_position VARCHAR(20) NOT NULL,
    item_id INT NOT NULL,
    wins INT NOT NULL DEFAULT 0,
    matches INT NOT NULL DEFAULT 0,
    PRIMARY KEY (patch, champion_id, team_position, item_id)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_champion_stats_patch ON champion_stats(patch);
CREATE INDEX IF NOT EXISTS idx_champion_stats_champion ON champion_stats(champion_id);
CREATE INDEX IF NOT EXISTS idx_champion_items_patch ON champion_items(patch);
CREATE INDEX IF NOT EXISTS idx_champion_items_champion ON champion_items(champion_id);
