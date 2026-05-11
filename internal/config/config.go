/*
Copyright © 2026 Two Tech Studio
*/
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// EndGitConfig represents the EndGit CLI configuration.
type EndGitConfig struct {
	APIToken     string `json:"apiToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	APIURL       string `json:"apiUrl,omitempty"`
}

const defaultAPIURL = "https://api.endgit.dev"

func configDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(dir, ".endgit"), nil
}

func configFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// GetConfig loads the EndGit configuration from disk.
// Returns default config if file doesn't exist or can't be read.
func GetConfig() EndGitConfig {
	file, err := configFile()
	if err != nil {
		return defaultConfig()
	}

	// If file doesn't exist → return default
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return defaultConfig()
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return defaultConfig()
	}

	var cfg EndGitConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig()
	}

	// Ensure default API URL always exists
	if cfg.APIURL == "" {
		cfg.APIURL = defaultAPIURL
	}

	return cfg
}

// SaveConfig writes the EndGit configuration to disk.
func SaveConfig(config EndGitConfig) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	file, err := configFile()
	if err != nil {
		return err
	}

	// Create config dir if missing
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(file, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func defaultConfig() EndGitConfig {
	return EndGitConfig{
		APIURL: defaultAPIURL,
	}
}
