/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

type PluginJSON struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	API         []string `json:"api"`
	Main        string   `json:"main"`
}

type EndgitJSON struct {
	DisplayName string `json:"displayName"`
	PluginType  string `json:"pluginType"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Endstone plugin configuration",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Initialize Endstone Plugin")

		dir, err := os.Getwd()
		if err != nil {
			log.Fatal("Failed to determine current directory", err)
		}
		currentDir := filepath.Base(dir)

		reader := bufio.NewReader(os.Stdin)

		read := func(prompt, def string) string {
			if def != "" {
				fmt.Printf("%s [%s]: ", prompt, def)
			} else {
				fmt.Printf("%s: ", prompt)
			}
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "" {
				return def
			}
			return input
		}

		name := read("Plugin Name (slug)", sanitizeSlug(currentDir))
		displayName := read("Display Name", currentDir)
		version := read("Version", "1.0.0")
		description := read("Description", "")
		apiVersion := read("Required Endstone API version", "^0.5.0")

		fmt.Println("Plugin Type:")
		fmt.Println("1) Python")
		fmt.Println("2) C++")
		fmt.Print("Select [1-2]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		pluginType := "PYTHON"
		if choice == "2" {
			pluginType = "CPP"
		}

		if name == "" {
			log.Warn("Initialization cancelled")
			return
		}

		plugin := PluginJSON{
			Name:        name,
			Version:     version,
			Description: description,
			API:         []string{apiVersion},
			Main:        "src.main",
		}

		if pluginType == "CPP" {
			plugin.Main = "EndstonePlugin"
		}

		pluginPath := filepath.Join(dir, "plugin.json")
		endgitPath := filepath.Join(dir, "endgit.json")

		// Warn if files already exist
		if _, err := os.Stat(pluginPath); err == nil {
			log.Warn("plugin.json already exists and will be overwritten")
		}
		if _, err := os.Stat(endgitPath); err == nil {
			log.Warn("endgit.json already exists and will be overwritten")
		}

		if err := writeJSON(pluginPath, plugin); err != nil {
			log.Fatal("Failed to write plugin.json", err)
		}

		endgitCfg := EndgitJSON{
			DisplayName: displayName,
			PluginType:  pluginType,
		}

		if err := writeJSON(endgitPath, endgitCfg); err != nil {
			log.Fatal("Failed to write endgit.json", err)
		}

		log.Success("Created plugin.json and endgit.json")
		log.Info("You can now run 'endgit publish' to push your plugin")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func sanitizeSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	var out strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out.WriteRune(r)
		}
	}
	return out.String()
}
