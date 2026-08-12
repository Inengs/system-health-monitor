// ai.go turns the current alerts into a short plain-language diagnosis
// via the Gemini API. Debounced: only re-queries when the alerting
// process set changes, or after suggestionTTL passes. No-ops cleanly if
// no API key is set — the rule-based banner already covers the basics.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	geminiURLFmt  = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"
	geminiModel   = "gemini-2.5-flash" // check ai.google.dev/gemini-api/docs/models for current free-tier models
	suggestionTTL = 5 * time.Minute
)

type Suggester struct {
	apiKey string
	client *http.Client

	mu          sync.Mutex
	lastKey     string
	lastText    string
	lastFetched time.Time
}

// NewSuggester reads GEMINI_API_KEY from the environment. Leave it
// unset during dev — Suggest() just returns "" until it's configured.
func NewSuggester() *Suggester {
	cfg, _ := LoadConfig()
	return &Suggester{
		apiKey: cfg.GeminiAPIKey,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// Suggest returns a cached or freshly-generated diagnosis for the given
// snapshot's alerts, or "" if there's nothing to say.
func (s *Suggester) Suggest(snap Snapshot) string {
	if s.apiKey == "" || len(snap.Alerts) == 0 {
		return ""
	}

	key := fingerprint(snap.Alerts)

	s.mu.Lock()
	if key == s.lastKey && time.Since(s.lastFetched) < suggestionTTL {
		text := s.lastText
		s.mu.Unlock()
		return text
	}
	s.mu.Unlock()

	text, err := s.fetch(snap)
	if err != nil {
		return "" // fail silent, rule-based banner still works
	}

	s.mu.Lock()
	s.lastKey, s.lastText, s.lastFetched = key, text, time.Now()
	s.mu.Unlock()
	return text
}

// SetAPIKey updates the key used for future requests (e.g. right after
// the user saves it in settings). Clears any cached suggestion made
// under the old key.
func (s *Suggester) SetAPIKey(key string) {
	s.mu.Lock()
	s.apiKey = key
	s.lastKey = ""
	s.mu.Unlock()
}

func fingerprint(alerts []Alert) string {
	names := make([]string, len(alerts))
	for i, a := range alerts {
		names[i] = a.Name
	}
	return strings.Join(names, ",")
}

func (s *Suggester) fetch(snap Snapshot) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{{
				"text": "You explain, in one short plain-language paragraph, why a " +
					"Windows laptop feels slow, based on process data. Suggest at " +
					"most one concrete action. No jargon, no PIDs, under 60 words.",
			}},
		},
		"contents": []map[string]any{{
			"role":  "user",
			"parts": []map[string]string{{"text": buildPrompt(snap)}},
		}},
		"generationConfig": map[string]any{"maxOutputTokens": 200},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	url := fmt.Sprintf(geminiURLFmt, geminiModel)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API returned %d", resp.StatusCode)
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}

func buildPrompt(snap Snapshot) string {
	var b strings.Builder
	b.WriteString("Top processes by CPU/memory right now:\n")
	n := len(snap.TopByCPU)
	if n > 5 {
		n = 5
	}
	for _, p := range snap.TopByCPU[:n] {
		fmt.Fprintf(&b, "- %s: %.0f%% CPU, %.0f%% memory\n", p.Friendly, p.CPUPercent, p.MemPercent)
	}
	b.WriteString("\nAlerts currently active:\n")
	for _, a := range snap.Alerts {
		fmt.Fprintf(&b, "- %s: %s\n", a.Name, a.Reason)
	}
	return b.String()
}