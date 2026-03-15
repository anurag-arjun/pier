// Package config manages user configuration loading and saving.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig is the global application configuration.
type AppConfig struct {
	PiPath       string `json:"pi_path,omitempty"`
	BrPath       string `json:"br_path,omitempty"`
	DefaultModel string `json:"default_model"`
	Theme        string `json:"theme"` // "light" | "dark"
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() AppConfig {
	return AppConfig{
		DefaultModel: "claude-sonnet-4-5",
		Theme:        "dark",
	}
}

// ConfigDir returns the config directory (~/.config/pier/).
func ConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "pier"), nil
}

// Load reads config from ~/.config/pier/config.json.
// Returns defaults if file doesn't exist.
func Load() (AppConfig, error) {
	dir, err := ConfigDir()
	if err != nil {
		return DefaultConfig(), err
	}

	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}
	return cfg, nil
}

// Save writes config to ~/.config/pier/config.json.
func Save(cfg AppConfig) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0644)
}
