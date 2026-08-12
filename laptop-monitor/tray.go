// tray.go implements the system tray icon: the ready/exit callbacks
// energye/systray expects, plus a background poller that swaps the icon
// between OK and warning based on the Monitor's latest alerts.
package main

import (
	"time"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// onTrayReady wires up the tray icon, its menu, and an icon-state poller.
// Called once by systray.Run in main.go.
func onTrayReady(app *App, iconOK, iconWarning []byte) {
	systray.SetIcon(iconOK)
	systray.SetTooltip("System Health Monitor")

	show := systray.AddMenuItem("Show", "Show details")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "Quit")

	// energye/systray uses callback registration, not channels — no
	// goroutine/select needed here (unlike getlantern/systray's
	// ClickedCh pattern).
	show.Click(func() {
		runtime.WindowShow(app.ctx)
	})
	quit.Click(func() {
		runtime.Quit(app.ctx)
	})

	// Icon state poller: switch to the warning icon whenever the latest
	// snapshot has any active alerts, back to OK once they clear.
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		warning := false
		for range ticker.C {
			hasAlerts := len(app.monitor.Snapshot().Alerts) > 0
			if hasAlerts == warning {
				continue
			}
			warning = hasAlerts
			if warning {
				systray.SetIcon(iconWarning)
			} else {
				systray.SetIcon(iconOK)
			}
		}
	}()
}

// onTrayExit runs when systray's loop ends (after Quit()). Nothing to
// clean up here — runtime.Quit already tears down the window/process.
func onTrayExit() {}