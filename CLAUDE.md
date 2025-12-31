# GhostDraft - Project Overview

A Wails-based League of Legends overlay that provides real-time champion select assistance.

## Tech Stack
- **Backend**: Go with Wails v2
- **Frontend**: Vanilla JavaScript (no framework)
- **Data Sources**: Riot LCU API, PostgreSQL (own match data), Data Dragon

## Project Structure

```
├── app.go                 # Main application logic, event handlers
├── main.go                # Wails app entry point
├── frontend/
│   └── src/
│       ├── main.js        # UI logic, event listeners, DOM updates
│       └── style.css      # Styling
├── internal/
│   ├── lcu/
│   │   ├── client.go      # LCU HTTP client (connects to League Client)
│   │   ├── websocket.go   # LCU WebSocket (champ select events)
│   │   ├── champions.go   # ChampionRegistry - ID→name/icon from Data Dragon
│   │   ├── items.go       # ItemRegistry - ID→name/icon from Data Dragon
│   │   └── types.go       # LCU data structures
│   ├── stats/
│   │   └── provider.go    # PostgreSQL queries for builds, matchups, counters
│   └── data/
│       └── champions.go   # SQLite DB for static champion data (damage types, tags)
├── data-analyzer/         # Separate module for collecting match data
```

## Key Data Flows

### 1. Champion Select Updates
```
LCU WebSocket → websocket.go:handleChampSelectEvent
             → app.go:onChampSelectUpdate
             → Emits: champselect:update, bans:update, items:update, etc.
             → Frontend listeners update DOM
```

### 2. Stats Provider (PostgreSQL)
The `internal/stats/provider.go` queries our own match data:
- **FetchChampionData**: Item builds with win rates and pick rates
- **FetchAllMatchups**: All matchup win rates for a champion
- **FetchCounterMatchups**: Top N counters (lowest win rate matchups)
- **FetchMatchup**: Specific champion vs enemy win rate

### Database Tables (managed by data-analyzer/reducer)
```sql
champion_stats    -- Champion win rates by patch/position
champion_items    -- Item stats per champion/position
champion_matchups -- Matchup win rates between champions
```

### 3. Frontend Events
```javascript
EventsOn('lcu:status', updateStatus);
EventsOn('champselect:update', updateChampSelect);
EventsOn('build:update', updateBuild);
EventsOn('bans:update', updateBans);
EventsOn('items:update', updateItems);
EventsOn('teamcomp:update', updateTeamComp);
EventsOn('fullcomp:update', updateFullComp);
```

## UI Tabs
1. **Matchup** - Counter matchups (champions with lowest WR against you), live matchup WR vs lane opponent
2. **Build** - Core items, situational items with win rates
3. **Team Comp** - Team archetype analysis (when all locked in)

## Important Implementation Notes

### Stats Provider
- Connects to PostgreSQL at `DATABASE_URL` env var or default localhost
- Fetches current patch from database on startup
- Falls back to aggregated data across patches if current patch has no data

### Item Filtering
- Only shows "completed" items (not components)
- Uses Data Dragon to identify items with no "into" field and cost >= 1000g

### Caching Keys
- `lastFetchedChamp` - prevents refetching same champion
- `lastBanFetchKey` - `"{champId}-{role}"` for bans
- `lastItemFetchKey` - `"{champId}-{role}"` for items

## SQLite Database
Located at `{UserConfigDir}/GhostDraft/champions.db`
- Auto-created on first run
- Stores: champion name, damage type (AP/AD/Mixed/Tank), role tags
- Used for team comp analysis

## Environment Variables
```
DATABASE_URL=postgres://analyzer:analyzer123@localhost:5432/lol_matches?sslmode=disable
```

## Common Tasks

### Adding a new event
1. Create emit in `app.go`: `runtime.EventsEmit(a.ctx, "event:name", data)`
2. Add listener in `main.js`: `EventsOn('event:name', handlerFunction)`
3. Create handler function to update DOM

### Adding a new tab
1. Add button in `main.js` HTML: `<button class="tab-btn" data-tab="tabname">Label</button>`
2. Add content div: `<div class="tab-content" id="tab-tabname">...</div>`
3. Tab switching is automatic via existing JS

## Build Commands
```bash
go build ./...           # Check Go compiles
wails dev                # Run in dev mode
wails build              # Build production binary
```

## Data Pipeline
See `data-analyzer/CLAUDE.md` for the match collection and aggregation pipeline.
