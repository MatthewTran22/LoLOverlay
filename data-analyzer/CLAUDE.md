# Data Analyzer - Match History Collection

## Purpose
Collect match history data from Riot API to build our own item statistics instead of relying on U.GG.

## Quick Start

```bash
# Start PostgreSQL
docker-compose up -d

# Collect data using Riot IDs (recommended)
go run cmd/collector/main.go --riot-ids="Faker#KR1,Player#NA1"

# Or using PUUIDs directly
go run cmd/collector/main.go --puuids="PUUID_HERE"

# Start web viewer
go run cmd/server/main.go
# Open http://localhost:8080
```

## Environment Variables
Create `.env` file:
```
RIOT_API_KEY=RGAPI-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
DATABASE_URL=postgres://analyzer:analyzer123@localhost:5432/lol_matches?sslmode=disable
```

## Project Structure
```
data-analyzer/
├── cmd/
│   ├── collector/    # CLI to fetch match data
│   └── server/       # Web UI server
├── internal/
│   ├── riot/         # Riot API client
│   │   ├── client.go # HTTP client with rate limiting
│   │   └── types.go  # API response structs
│   └── db/           # Database layer
│       ├── db.go     # Connection setup
│       ├── schema.sql
│       └── queries.go
└── web/              # Static HTML/CSS
```

## Database Schema

### matches
- `match_id` - Riot match ID (e.g., "NA1_12345678")
- `game_version` - Patch version
- `game_duration` - Duration in seconds
- `game_creation` - Unix timestamp

### participants
- `match_id` - Foreign key to matches
- `champion_id` - Champion ID
- `champion_name` - Champion name
- `team_position` - TOP, JUNGLE, MIDDLE, BOTTOM, UTILITY
- `win` - Boolean
- `item0-5` - Final items
- `build_order` - JSONB array of item IDs in purchase order

## Riot API Endpoints Used

1. **Account Lookup**: `/riot/account/v1/accounts/by-riot-id/{gameName}/{tagLine}`
2. **Match History**: `/lol/match/v5/matches/by-puuid/{puuid}/ids?queue=420&count=20`
3. **Match Details**: `/lol/match/v5/matches/{matchId}`
4. **Match Timeline**: `/lol/match/v5/matches/{matchId}/timeline`

## Rate Limits (Dev Key)
- 20 requests/second
- 100 requests/2 minutes

The client implements a token bucket rate limiter to stay within these limits.

## Data Collection Strategy
- Target: Emerald+ ranked players
- Queue: Ranked Solo/Duo only (queue 420)
- Matches per player: 20
- Data captured: Win/loss + full build path from timeline
