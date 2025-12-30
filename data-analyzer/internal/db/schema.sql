-- Match History Data Analyzer Schema

-- Matches (Ranked Solo only)
CREATE TABLE IF NOT EXISTS matches (
    match_id VARCHAR(50) PRIMARY KEY,
    game_version VARCHAR(20),
    game_duration INT,
    game_creation BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Participant data with build order
CREATE TABLE IF NOT EXISTS participants (
    id SERIAL PRIMARY KEY,
    match_id VARCHAR(50) REFERENCES matches(match_id) ON DELETE CASCADE,
    puuid VARCHAR(100),
    game_name VARCHAR(100),
    tag_line VARCHAR(20),
    champion_id INT,
    champion_name VARCHAR(50),
    team_position VARCHAR(20),  -- TOP, JUNGLE, MIDDLE, BOTTOM, UTILITY
    win BOOLEAN,
    -- Final items (slot 0-5, excluding trinket)
    item0 INT DEFAULT 0,
    item1 INT DEFAULT 0,
    item2 INT DEFAULT 0,
    item3 INT DEFAULT 0,
    item4 INT DEFAULT 0,
    item5 INT DEFAULT 0,
    -- Build order from timeline (JSON array of item IDs in purchase order)
    build_order JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_participants_champion ON participants(champion_id);
CREATE INDEX IF NOT EXISTS idx_participants_position ON participants(team_position);
CREATE INDEX IF NOT EXISTS idx_participants_win ON participants(win);
CREATE INDEX IF NOT EXISTS idx_participants_puuid ON participants(puuid);
CREATE INDEX IF NOT EXISTS idx_matches_creation ON matches(game_creation);
