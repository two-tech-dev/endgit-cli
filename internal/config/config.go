/*
Copyright © 2026 Two Tech Studio
*/
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type EndGitConfig struct {
	APIToken string `json:"apiToken,omitempty"`
	APIURL   string `json:"apiUrl,omitempty"`
}

func configDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
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

func GetConfig() EndGitConfig {
	file, err := configFile()
	if err != nil {
		return defaultConfig()
	}

	// if file doesn't exist → return default
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

	// ensure default API URL always exists
	if cfg.APIURL == "" {
		cfg.APIURL = "https://api.endgit.dev"
	}

	return cfg
}

func SaveConfig(update EndGitConfig) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	file, err := configFile()
	if err != nil {
		return err
	}

	// create config dir if missing
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	current := GetConfig()

	// merge (like {...current, ...config})
	if update.APIToken != "" {
		current.APIToken = update.APIToken
	}
	if update.APIURL != "" {
		current.APIURL = update.APIURL
	}

	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0644)
}

func defaultConfig() EndGitConfig {
	return EndGitConfig{
		APIURL: "https://api.endgit.dev",
	}
}
