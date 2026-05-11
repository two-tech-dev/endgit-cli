/*
Copyright © 2026 Two Tech Studio
*/
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	defaultAPIURL = "https://api.endgit.dev"
	serviceName   = "endgit-cli"
	tokenKey      = "api-token"
	refreshKey    = "refresh-token"
)

// EndGitConfig represents the EndGit CLI configuration.
type EndGitConfig struct {
	APIToken     string `json:"apiToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	APIURL       string `json:"apiUrl,omitempty"`
}

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

	// Try to load tokens from keyring
	if token, err := keyring.Get(serviceName, tokenKey); err == nil {
		cfg.APIToken = token
	}
	if refresh, err := keyring.Get(serviceName, refreshKey); err == nil {
		cfg.RefreshToken = refresh
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

	// Handle tokens in keyring
	keyringAvailable := true

	if config.APIToken != "" {
		if err := keyring.Set(serviceName, tokenKey, config.APIToken); err != nil {
			keyringAvailable = false
		}
	} else {
		_ = keyring.Delete(serviceName, tokenKey)
	}

	if config.RefreshToken != "" {
		if err := keyring.Set(serviceName, refreshKey, config.RefreshToken); err != nil {
			keyringAvailable = false
		}
	} else {
		_ = keyring.Delete(serviceName, refreshKey)
	}

	// Create a copy for JSON storage
	jsonCfg := config
	if keyringAvailable {
		// If keyring worked, we don't need tokens in the JSON file
		jsonCfg.APIToken = ""
		jsonCfg.RefreshToken = ""
	}

	data, err := json.MarshalIndent(jsonCfg, "", "  ")
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
