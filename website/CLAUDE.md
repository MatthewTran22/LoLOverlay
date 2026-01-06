# GhostDraft Website

Next.js companion website for browsing League of Legends champion statistics, builds, and matchup data.

## Tech Stack
- **Framework**: Next.js 15 with App Router
- **Styling**: Tailwind CSS 4 + Custom CSS (Hextech Arcane theme)
- **Database**: Turso (libSQL) - cloud SQLite, read-only
- **Caching**: Next.js `unstable_cache` (1 hour TTL)
- **Fonts**: Cinzel (display), Rajdhani (body)

## Project Structure

```
website/
├── src/
│   ├── app/
│   │   ├── globals.css           # Hextech Arcane theme + Tailwind
│   │   ├── layout.tsx            # Root layout with fonts
│   │   ├── page.tsx              # Landing page
│   │   ├── privacy/page.tsx      # Privacy policy
│   │   ├── terms/page.tsx        # Terms of service
│   │   ├── stats/
│   │   │   ├── page.tsx          # Champion tier list by role
│   │   │   └── [championId]/
│   │   │       └── page.tsx      # Champion detail page
│   │   ├── admin/
│   │   │   └── page.tsx          # Admin panel (read-only status)
│   │   └── api/
│   │       ├── stats/route.ts    # Stats info endpoint
│   │       ├── meta/route.ts     # Top champions by role
│   │       ├── champions/[id]/route.ts  # Champion build data
│   │       └── matchups/[id]/route.ts   # Champion matchup data
│   ├── components/
│   │   ├── Header.tsx            # Navigation header
│   │   └── Footer.tsx            # Site footer
│   └── lib/
│       ├── db.ts                 # Turso client connection
│       ├── stats.ts              # Stats query functions (async, cached)
│       └── champions.ts          # Champion ID→name mapping, utilities
├── scripts/
│   ├── push-to-turso.mjs         # Push local SQLite to Turso
│   └── add-indexes.mjs           # Add database indexes
└── public/                       # Static assets
```

## Pages

### `/` - Landing Page
Hero section with download CTA, feature highlights, and links to stats.

### `/stats` - Champion Tier List
- Role tabs (Top, Jungle, Mid, ADC, Support)
- Full champion list sorted by win rate
- Tier badges (S+, S, A, B, C, D)
- Links to individual champion pages

### `/stats/[championId]` - Champion Detail
- Champion header with icon, tier, win rate, pick rate
- Role tabs for multi-role champions
- Recommended build: 2 core items + boots + situational options
- Counters section (worst matchups)
- Best matchups section

### `/admin` - Admin Panel
- Database status (patch, champion count, matchup count)
- Read-only mode - data updates managed by Data Analyzer

## Key Files

### `lib/db.ts`
- Turso client singleton using `@libsql/client`
- `getDb()` - Returns Turso client
- `getCurrentPatch()` - Async, fetches current patch version
- `hasData()` - Async, checks if database has data

### `lib/stats.ts`
All functions are **async** and wrapped with `unstable_cache` (1 hour TTL):
- `fetchChampionData()` - Build paths with item options
- `fetchAllMatchups()` - All matchup win rates
- `fetchCounterMatchups()` - Worst matchups (counters)
- `fetchBestMatchups()` - Best matchups
- `fetchAllChampionsByRole()` - Full tier list
- `fetchChampionStats()` - Single champion stats
- `fetchChampionRoles()` - Roles a champion plays
- `fetchTopChampionsByRole()` - Top N champions by win rate
- `getStatsInfo()` - Patch and database stats

### `lib/ddragon.ts`
- Fetches latest version from Data Dragon API dynamically
- Loads champion and item data from Data Dragon
- Caches data for 1 hour (uses Next.js fetch caching)
- Exports `getDDragon()` for synchronous lookups after initial load

### `lib/champions.ts`
- Re-exports Data Dragon functions (getChampionName, getChampionIcon, getItemName, getItemIcon)
- Tier calculation (S+/S/A/B/C/D based on win rate + pick rate)
- Role display names and utility functions

## Database

### Turso (Cloud SQLite)
- Read-only access from website
- Data pushed from local SQLite via `scripts/push-to-turso.mjs`
- Indexes added via `scripts/add-indexes.mjs`

### Schema
```sql
champion_stats        -- Win rates by champion/position
champion_items        -- Item stats (overall)
champion_item_slots   -- Item stats by build slot (1-6)
champion_matchups     -- Matchup win rates
data_version          -- Current patch version
```

### Indexes
```sql
idx_champion_stats_position
idx_champion_stats_champ_pos
idx_champion_items_champ_pos
idx_champion_item_slots_champ_pos
idx_champion_item_slots_champ_pos_slot
idx_champion_matchups_champ_pos
idx_champion_matchups_enemy
```

## Build System

### Item Build Logic
Core build = 2 legendary items + best boots:
1. Query best boots across all slots
2. Get top item from slots 1-3 (excluding boots + duplicates)
3. Add boots to complete core
4. 4th/5th/6th options exclude core items

### Tier Calculation
```typescript
if (winRate >= 53 && pickRate >= 3) return "S+";
if (winRate >= 52 && pickRate >= 2) return "S";
if (winRate >= 51 && pickRate >= 1) return "A";
if (winRate >= 50) return "B";
if (winRate >= 48) return "C";
return "D";
```

## Data Flow

```
Desktop App (local stats.db)
         ↓
scripts/push-to-turso.mjs
         ↓
Turso Cloud Database
         ↓
Website queries via @libsql/client
         ↓
Results cached with unstable_cache (1 hour)
         ↓
Server components render pages
```

## Environment Variables

```
TURSO_DATABASE_URL=libsql://your-db.turso.io
TURSO_AUTH_TOKEN=your-token
```

## Scripts

### Push data to Turso
```bash
node scripts/push-to-turso.mjs
```
Reads from desktop app's `stats.db` and pushes to Turso in batches.

### Add indexes
```bash
node scripts/add-indexes.mjs
```
Creates indexes on Turso for faster queries.

## Build Commands

```bash
npm run dev      # Development server
npm run build    # Production build
npm run start    # Start production server
```

## Caching Strategy

- All stats queries use `unstable_cache` with 1 hour revalidation
- First request hits Turso (slow ~100-500ms)
- Subsequent requests within 1 hour are instant (cached)
- Pages also have `revalidate = 3600` for ISR

## CSS Classes

### Theme Classes (globals.css)
- `.hex-card` - Card with gold border
- `.btn-hextech` - Gold gradient button
- `.btn-outline` - Outlined button
- `.text-glow` - Gold text shadow
- `.text-glow-cyan` - Cyan text shadow
- `.wr-high` / `.wr-mid` / `.wr-low` - Win rate colors
- `.hover-line` - Underline hover effect

### Tailwind Custom Colors
Use CSS variables: `text-[var(--hextech-gold)]`, `bg-[var(--abyss)]`, etc.
