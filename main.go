package main

import (
	"embed"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load .env file for development (ignored if not present)
	if err := godotenv.Load(); err != nil {
		fmt.Println("[Config] No .env file found (using build-time or system env vars)")
	} else {
		fmt.Println("[Config] Loaded .env file")
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "GhostDraft",
		Width:         280,
		Height:        120,
		StartHidden:   false,
		Frameless:     true,
		AlwaysOnTop:   true,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			DisableWindowIcon:            true,
			WebviewIsTransparent:         true,
			WindowIsTranslucent:          true,
			DisableFramelessWindowDecorations: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
