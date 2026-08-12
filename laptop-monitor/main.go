// Command laptop-monitor is a Windows system-tray app that watches CPU
// and memory usage and explains, in plain language, what's slowing the
// machine down.
package main

import (
	"embed"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// assets holds the built React frontend (frontend/dist), embedded into the
// binary so the .exe is self-contained — no separate files to ship.
//
//go:embed all:frontend/dist
var assets embed.FS

// Tray icons, embedded so systray works without reading from disk at
// runtime (matters once this is a single distributed .exe).
//
//go:embed build/tray-ok.ico
var iconOK []byte

//go:embed build/tray-warning.ico
var iconWarning []byte

func main() {
	app := NewApp()

	// systray.Run blocks internally, so it needs its own goroutine.
	// Started before wails.Run so the tray icon exists as soon as the
	// process launches, even before the window system has spun up.
	go systray.Run(
		func() { onTrayReady(app, iconOK, iconWarning) },
		onTrayExit,
	)

	err := wails.Run(&options.App{
		Title:     "System Health Monitor",
		Width:     420,
		Height:    560,
		MinWidth:  380,
		MinHeight: 480,

		AssetServer: &assetserver.Options{
			Assets: assets,
		},

		// StartHidden + HideWindowOnClose: launches straight into the
		// tray (no window flash on login) and clicking the window's
		// close button hides it instead of quitting — quitting only
		// happens via the tray menu's "Quit".
		StartHidden:       true,
		HideWindowOnClose: true,

		OnStartup: app.startup,
		Bind:      []interface{}{app},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}