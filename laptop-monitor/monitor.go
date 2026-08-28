// monitor.go collects process, CPU, and memory data on a fixed interval
// and exposes the latest reading through Snapshot(). It has no Wails or
// systray dependency, so it can be tested with `go test` before any UI
// is wired up.
package main

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	// CPUThreshold and MemThreshold are the percentages above which a
	// process triggers an Alert.
	CPUThreshold = 80.0
	MemThreshold = 80.0
)

// CloseWhitelist lists process names (lowercase) that RequestClose will
// always refuse to terminate, mirroring process_monitoring.py's
// KILL_WHITELIST.
var CloseWhitelist = map[string]bool{
	"system":               true,
	"system idle process":  true,
	"explorer.exe":         true,
	"winlogon.exe":         true,
	"csrss.exe":            true,
	"svchost.exe":          true,
	"msmpeng.exe":          true, // Windows Defender
	"wininit.exe":          true,
}

// FriendlyNames maps raw executable names to plain-language labels shown
// to non-technical users. Anything missing falls back to the raw name.
var FriendlyNames = map[string]string{
	"chrome.exe":        "Google Chrome",
	"msedge.exe":        "Microsoft Edge",
	"firefox.exe":       "Firefox",
	"explorer.exe":      "Windows File Explorer",
	"msmpeng.exe":       "Windows Defender (antivirus — spikes are normal during scans)",
	"searchindexer.exe": "Windows Search (indexing your files)",
}

// ProcInfo is a single process's stats, shaped for direct JSON transfer
// to the frontend.
type ProcInfo struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name"`
	Friendly    string  `json:"friendly"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemPercent  float32 `json:"memPercent"`
	SafeToClose bool    `json:"safeToClose"`
}

// Alert flags a process that crossed a threshold on the most recent tick.
type Alert struct {
	PID    int32  `json:"pid"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
	Time   string `json:"time"`
}

// Snapshot is the full picture returned to the frontend on each poll.
type Snapshot struct {
	TopByCPU []ProcInfo `json:"topByCpu"`
	TopByMem []ProcInfo `json:"topByMem"`
	Alerts   []Alert    `json:"alerts"`
}

// Monitor owns the latest Snapshot and refreshes it on a ticker in the
// background. Safe for concurrent reads via Snapshot().
type Monitor struct {
	mu     sync.RWMutex
	latest Snapshot
	seenAlerts map[string]bool // fingerprint of currently-active alerts
	onNewAlert func(Alert)
}

// NewMonitor returns an idle Monitor — call Start to begin polling.
func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start(interval time.Duration) {
	go func() {
		for {
			snap := m.collect()

			m.mu.Lock()
			m.latest = snap
			fresh, next := m.diffAlerts(snap.Alerts)
			m.seenAlerts = next
			cb := m.onNewAlert
			m.mu.Unlock()

			if cb != nil {
				for _, a := range fresh {
					cb(a)
				}
			}

			time.Sleep(interval)
		}
	}()
}

// diffAlerts returns only the alerts not present in the previous tick's
// seenAlerts, plus the new full set for next tick's comparison. Keyed by
// PID+reason, so a process that clears and later re-triggers is treated
// as fresh again rather than staying permanently silenced.
func (m *Monitor) diffAlerts(alerts []Alert) ([]Alert, map[string]bool) {
	next := make(map[string]bool, len(alerts))
	var fresh []Alert
	for _, a := range alerts {
		key := fmt.Sprintf("%d|%s", a.PID, a.Reason)
		next[key] = true
		if !m.seenAlerts[key] {
			fresh = append(fresh, a)
		}
	}
	return fresh, next
}

// Snapshot returns the most recently collected reading.
func (m *Monitor) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.latest
}

// collect walks all running processes once and builds a fresh Snapshot.
//
// Note: gopsutil's CPUPercent() needs two calls spaced apart to mean
// anything — the very first tick after startup reads ~0 for every
// process. That's expected, not a bug; it settles from the second tick
// onward since Monitor reuses the same long-lived loop.
func (m *Monitor) collect() Snapshot {
	procs, _ := process.Processes()
	infos := make([]ProcInfo, 0, len(procs))

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue // process exited mid-scan, or access denied
		}
		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()
		lower := lowerName(name)

		infos = append(infos, ProcInfo{
			PID:         p.Pid,
			Name:        name,
			Friendly:    friendlyName(lower, name),
			CPUPercent:  cpuPct,
			MemPercent:  memPct,
			SafeToClose: !CloseWhitelist[lower],
		})
	}

	byCPU := append([]ProcInfo{}, infos...)
	sort.Slice(byCPU, func(i, j int) bool { return byCPU[i].CPUPercent > byCPU[j].CPUPercent })

	byMem := append([]ProcInfo{}, infos...)
	sort.Slice(byMem, func(i, j int) bool { return byMem[i].MemPercent > byMem[j].MemPercent })

	return Snapshot{
		TopByCPU: topN(byCPU, 10),
		TopByMem: topN(byMem, 10),
		Alerts:   checkThresholds(infos),
	}
}

// checkThresholds returns an Alert for every process currently above
// CPUThreshold or MemThreshold.
func checkThresholds(procs []ProcInfo) []Alert {
	var alerts []Alert
	now := time.Now().Format(time.RFC3339)
	for _, p := range procs {
		if p.CPUPercent > CPUThreshold {
			alerts = append(alerts, Alert{p.PID, p.Name, "CPU usage is unusually high", now})
		}
		if float64(p.MemPercent) > MemThreshold {
			alerts = append(alerts, Alert{p.PID, p.Name, "Memory usage is unusually high", now})
		}
	}
	return alerts
}

// topN returns at most n items from s.
func topN(s []ProcInfo, n int) []ProcInfo {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// friendlyName looks up lower (an already-lowercased process name) in
// FriendlyNames, falling back to the raw name if there's no entry.
func friendlyName(lower, raw string) string {
	if fn, ok := FriendlyNames[lower]; ok {
		return fn
	}
	return raw
}

// lowerName lowercases an ASCII process name without pulling in the
// strings/unicode packages for a case this simple.
func lowerName(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		out[i] = c
	}
	return string(out)
}

// SetOnNewAlert registers a callback fired once per alert the first tick
// it appears — not on every poll while it's still active. Call before
// Start.
func (m *Monitor) SetOnNewAlert(fn func(Alert)) {
	m.mu.Lock()
	m.onNewAlert = fn
	m.mu.Unlock()
}