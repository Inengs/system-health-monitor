package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// TestCollectPrintsSnapshot is a manual eyeball check, not an assertion
// test. Run: go test -run TestCollectPrintsSnapshot -v
func TestCollectPrintsSnapshot(t *testing.T) {
	m := NewMonitor()
	m.Start(2 * time.Second)
	time.Sleep(3 * time.Second) // let it tick twice so CPU% isn't all zero
	b, _ := json.MarshalIndent(m.Snapshot(), "", "  ")
	fmt.Println(string(b))
}