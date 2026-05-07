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

	"github.com/fatih/color"
	"github.com/spf13/cobra"
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

		color.Cyan("Initialize Endstone Plugin\n")

		dir, _ := os.Getwd()
		currentDir := filepath.Base(dir)

		reader := bufio.NewReader(os.Stdin)

		read := func(prompt, def string) string {
			fmt.Printf("%s [%s]: ", prompt, def)
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
			color.Red("Initialization cancelled.")
			os.Exit(1)
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

		writeJSON(pluginPath, plugin)

		endgit := EndgitJSON{
			DisplayName: displayName,
			PluginType:  pluginType,
		}

		writeJSON(endgitPath, endgit)

		color.Green("\nSuccessfully created plugin.json and endgit.json.")
		color.HiBlack("You can now run endgit publish to push your plugin.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func writeJSON(path string, v any) {
	data, _ := json.MarshalIndent(v, "", "  ")
	_ = os.WriteFile(path, data, 0644)
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
