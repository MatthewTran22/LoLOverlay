# Data Analyzer - Match History Collection

## Purpose
Collect match history data from Riot API to build our own item statistics instead of relying on U.GG.

## Architecture: Rotated JSONL + Incremental Reducer

```
Riot API → Spider Loop → FileRotator → JSONL files
                              ↓
                         hot/ → warm/ → cold/
```

### Storage Lifecycle
- **hot/** - Active writes (current JSONL file)
- **warm/** - Closed files awaiting processing
- **cold/** - Compressed archives (.jsonl.gz)

### File Rotation Triggers
- 1,000 matches (10,000 participant records) per file
- 1 hour max file age
- Graceful shutdown

## Quick Start

```bash
# Collect data (writes to ./data by default)
go run cmd/collector/main.go --riot-id="Player#NA1"

# Custom data directory
go run cmd/collector/main.go --riot-id="Player#NA1" --data-dir="/path/to/data"

# Options
#   --count=20         Matches per player (default: 20)
#   --max-players=100  Max players to process (default: 100)
#   --data-dir=./data  Base directory for hot/warm/cold storage
```

## Environment Variables
Create `.env` file:
```
RIOT_API_KEY=RGAPI-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

## Project Structure
```
data-analyzer/
├── cmd/
│   ├── collector/       # Spider crawler CLI
│   └── server/          # Web UI server (legacy, uses PostgreSQL)
├── internal/
│   ├── riot/            # Riot API client
│   │   ├── client.go    # HTTP client with rate limiting
│   │   └── types.go     # API response structs
│   ├── storage/         # JSONL file rotation
│   │   ├── rotator.go   # FileRotator implementation
│   │   └── types.go     # RawMatch struct
│   └── db/              # PostgreSQL layer (legacy)
└── web/                 # Static HTML/CSS
```

## Data Format

### JSONL Records (one per participant)
```json
{
  "matchId": "NA1_12345678",
  "gameVersion": "14.24.1",
  "gameDuration": 1847,
  "gameCreation": 1703001234567,
  "puuid": "...",
  "gameName": "Player",
  "tagLine": "NA1",
  "championId": 103,
  "championName": "Ahri",
  "teamPosition": "MIDDLE",
  "win": true,
  "item0": 3089, "item1": 3157, "item2": 3020, "item3": 3165, "item4": 3135, "item5": 3907,
  "buildOrder": [3089, 3157, 3165, 3135, 3907]
}
```

## Collection Strategy

1. **Spider/Snowball approach**: Start with 1 player, discover others from matches
2. **In-memory deduplication**: Track visitedMatchIDs to avoid re-fetching
3. **Rate limiting**: 15 req/sec, 90 req/2min (conservative under 20/100 limits)
4. **Graceful shutdown**: Ctrl+C flushes current file to warm/

## Riot API Endpoints Used

1. **Account Lookup**: `/riot/account/v1/accounts/by-riot-id/{gameName}/{tagLine}`
2. **Match History**: `/lol/match/v5/matches/by-puuid/{puuid}/ids?queue=420&count=20`
3. **Match Details**: `/lol/match/v5/matches/{matchId}`
4. **Match Timeline**: `/lol/match/v5/matches/{matchId}/timeline`

## Rate Limits (Dev Key)
- 20 requests/second
- 100 requests/2 minutes
- 30-second wait on 429 responses

## Future: Batch Reducer

The reducer will process warm/ files to compute aggregated stats:
- Champion win rates by position
- Common build paths
- Item synergies

This allows re-running analysis without re-fetching from Riot API.
