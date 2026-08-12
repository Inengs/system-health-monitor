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
}

// NewApp returns an App with an idle Monitor — call startup to begin
// polling once the Wails runtime context is available.
func NewApp() *App {
	return &App{monitor: NewMonitor()}
}

// startup runs once, after Wails' runtime is ready. Stores the context
// (needed for runtime.* calls in tray.go) and starts the background
// poller.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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