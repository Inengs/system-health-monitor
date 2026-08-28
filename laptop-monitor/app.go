// app.go defines the App struct Wails binds to the frontend. Its
// exported methods are auto-generated into
// frontend/wailsjs/go/main/App.d.ts on every `wails dev`/`wails build`.
package main

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// App is the single binding surface between Go and the frontend.
type App struct {
	ctx     context.Context
	monitor *Monitor
	suggester *Suggester
}

// NewApp returns an App with an idle Monitor — call startup to begin
// polling once the Wails runtime context is available.
func NewApp() *App {
	return &App{monitor: NewMonitor(), suggester: NewSuggester()}
}

// startup runs once, after Wails' runtime is ready. Stores the context
// (needed for runtime.* calls in tray.go) and starts the background
// poller.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.monitor.SetOnNewAlert(NewNotifier().Notify)
	a.monitor.Start(3 * time.Second)
}

// GetSnapshot returns the latest process/CPU/memory reading. Callable
// from the frontend as window.go.main.App.GetSnapshot().
func (a *App) GetSnapshot() Snapshot {
	return a.monitor.Snapshot()
}

// RequestClose attempts to terminate the process with the given PID.
// The frontend must have already confirmed with the user — this
// function does not ask again. Refuses anything on CloseWhitelist.
func (a *App) RequestClose(pid int32) string {
	p, err := process.NewProcess(pid)
	if err != nil {
		return "Couldn't find that process — it may have already closed."
	}
	name, err := p.Name()
	if err != nil {
		return "Couldn't identify that process."
	}
	if CloseWhitelist[lowerName(name)] {
		return "This is a core Windows process and can't be closed here."
	}
	if err := p.Terminate(); err != nil {
		return "Couldn't close it — you may need administrator permission."
	}
	return "Closed."
}


// GetConfig returns current settings for the settings panel to display.
func (a *App) GetConfig() Config {
	cfg, _ := LoadConfig()
	return cfg
}

// SaveConfig persists settings and applies the new key immediately —
// no restart needed. Returns "" on success, an error message otherwise.
func (a *App) SaveConfig(cfg Config) string {
	if err := cfg.Save(); err != nil {
		return "Couldn't save settings: " + err.Error()
	}
	a.suggester.SetAPIKey(cfg.GeminiAPIKey)
	return ""
}