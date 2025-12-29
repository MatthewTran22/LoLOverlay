# GhostDraft

A lightweight League of Legends overlay that displays matchup win rates during champion select.

## Features

- **Auto-connects** to the League Client via the LCU API
- **Detects your role** and shows matchup data for your lane opponent
- **Win rate display** sourced from U.GG (Diamond+ data)
- **Minimal overlay** - appears only during champion select, hides otherwise
- **Resizable & draggable** - position it wherever works for you

## Screenshot

The overlay shows:
- Your assigned role (Top, Jungle, Mid, ADC, Support)
- Win rate against your lane opponent
- Color-coded: green (winning), red (losing), orange (even)

## Installation

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Build from source

```bash
# Clone the repo
git clone https://github.com/yourusername/ghostdraft.git
cd ghostdraft

# Install frontend dependencies
cd frontend && npm install && cd ..

# Build the app
wails build
```

The executable will be in `build/bin/`.

### Development

```bash
wails dev
```

## Usage

1. Run GhostDraft
2. Launch League of Legends
3. Enter champion select - the overlay will appear automatically
4. The overlay shows your matchup win rate once enemies start picking

## How it works

- Reads the League Client lockfile to connect via WebSocket
- Listens for champion select events
- Fetches matchup data from U.GG's public stats API
- Identifies your lane opponent by matching role + highest game count in matchup data

## Tech Stack

- **Backend**: Go + [Wails](https://wails.io/)
- **Frontend**: Vanilla JS + CSS
- **Data**: U.GG Stats API, Riot Data Dragon

## License

MIT
