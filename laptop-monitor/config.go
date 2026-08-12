// config.go persists user-editable settings — currently just the
// optional Gemini API key — to a per-user JSON file. Works out of the
// box for anyone downloading the app; AI suggestions stay off until the
// user adds their own free key via the settings panel.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the full set of user-editable settings, persisted to disk.
type Config struct {
	GeminiAPIKey string `json:"geminiApiKey"`
}

// configPath resolves the per-user config file location, creating its
// parent directory if needed. os.UserConfigDir resolves to %AppData% on
// Windows and ~/.config on Linux — one code path for both target OSes.
func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(dir, "SystemHealthMonitor")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "config.json"), nil
}

// LoadConfig reads the config file, returning a zero-value Config (no
// error) if it doesn't exist yet — the normal state on first run.
func LoadConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save writes the config to disk as indented JSON, user-readable only.
func (c Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}