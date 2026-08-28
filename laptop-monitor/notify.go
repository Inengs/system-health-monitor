// notify.go sends a toast notification the first time a process crosses
// a threshold — wired to Monitor.SetOnNewAlert so it fires once per new
// alert, not on every 3-second poll while the alert persists.
package main

import (
	"fmt"

	"github.com/gen2brain/beeep"
)

// Notifier sends toast notifications for newly-detected alerts.
type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

// Notify sends a toast for a single alert. Errors are logged, not
// surfaced — a failed notification shouldn't interrupt monitoring.
func (n *Notifier) Notify(a Alert) {
	msg := fmt.Sprintf("%s: %s", a.Name, a.Reason)
	if err := beeep.Notify("System Health Monitor", msg, ""); err != nil {
		fmt.Println("notify: failed to send toast:", err)
	}
}