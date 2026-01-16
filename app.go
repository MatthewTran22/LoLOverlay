package main

import (
	"context"
	"fmt"

	"ghostdraft/internal/data"
	"ghostdraft/internal/lcu"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx              context.Context
	lcuClient        *lcu.Client
	wsClient         *lcu.WebSocketClient
	liveClient       *lcu.LiveClient
	champions        *lcu.ChampionRegistry
	items            *lcu.ItemRegistry
	championDB       *data.ChampionDB
	tursoClient      *data.TursoClient     // Turso database connection
	statsProvider    *data.StatsProvider   // Stats queries (uses Turso with caching)
	stopPoll         chan struct{}
	lastFetchedChamp    int
	lastFetchedEnemy    int
	lastBanFetchKey     string
	lastItemFetchKey    string
	lastCounterFetchKey string
	windowVisible       bool

	// Champ select state - passed to in-game
	lockedChampionID   int
	lockedChampionName string
	lockedPosition     string

	// User identity - stored on LCU connection
	currentPUUID string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		lcuClient:     lcu.NewClient(),
		wsClient:      lcu.NewWebSocketClient(),
		liveClient:    lcu.NewLiveClient(),
		champions:     lcu.NewChampionRegistry(),
		items:         lcu.NewItemRegistry(),
		stopPoll:      make(chan struct{}),
		windowVisible: true,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize champion database
	if db, err := data.NewChampionDB(); err != nil {
		fmt.Printf("Failed to initialize champion DB: %v\n", err)
	} else {
		a.championDB = db
		fmt.Println("Champion database initialized")
	}

	// Position and size window relative to screen
	screens, err := runtime.ScreenGetAll(ctx)
	if err == nil && len(screens) > 0 {
		screen := screens[0]
		// Size: ~20% width, ~55% height (roughly matches champ select sidebar)
		width := screen.Size.Width * 20 / 100
		height := screen.Size.Height * 55 / 100
		runtime.WindowSetSize(ctx, width, height)
		// Position at right edge, vertically centered
		x := screen.Size.Width - width - 20
		y := (screen.Size.Height - height) / 2
		runtime.WindowSetPosition(ctx, x, y)
	}

	// Load data from Data Dragon in parallel
	go func() {
		if err := a.champions.Load(); err != nil {
			fmt.Printf("Failed to load champions: %v\n", err)
		}
	}()
	go func() {
		if err := a.items.Load(); err != nil {
			fmt.Printf("Failed to load items: %v\n", err)
		}
	}()

	// Initialize stats database and check for updates
	go a.initStats()

	// Set up champ select handler
	a.wsClient.SetChampSelectHandler(a.onChampSelectUpdate)

	// Set up gameflow handler
	a.wsClient.SetGameflowHandler(a.onGameflowUpdate)

	// Start polling for League Client
	go a.pollForLeagueClient()

	// Register global hotkey (Ctrl+O to toggle visibility)
	a.RegisterToggleHotkey()
}

// initStats initializes the Turso connection and stats provider
func (a *App) initStats() {
	// Connect to Turso
	tursoClient, err := data.NewTursoClient()
	if err != nil {
		fmt.Printf("Failed to connect to Turso: %v\n", err)
		return
	}
	a.tursoClient = tursoClient

	// Create stats provider with caching
	provider, err := data.NewStatsProvider(tursoClient)
	if err != nil {
		fmt.Printf("Stats provider not available: %v\n", err)
		return
	}

	// Fetch current patch from Turso
	if err := provider.FetchPatch(); err != nil {
		fmt.Printf("No stats patch available: %v\n", err)
		return
	}

	a.statsProvider = provider
	fmt.Printf("Stats provider ready (patch %s)\n", provider.GetPatch())
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	close(a.stopPoll)
	a.wsClient.Disconnect()
	a.lcuClient.Disconnect()
	if a.championDB != nil {
		a.championDB.Close()
	}
	if a.tursoClient != nil {
		a.tursoClient.Close()
	}
}
