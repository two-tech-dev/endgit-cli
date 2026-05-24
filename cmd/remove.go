/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var removeCmd = &cobra.Command{
	Use:     "remove [plugin]",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove an installed plugin",
	Long: `Remove one or more installed plugins.

Examples:
  endgit remove my-plugin    Remove a specific plugin
  endgit remove              Interactive select from installed plugins`,
	Run: func(cmd *cobra.Command, args []string) {
		entries, err := os.ReadDir("plugins")
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn("No plugins directory found.")
				return
			}
			log.Fatal("Failed to read plugins directory", err)
		}

		type pluginFile struct {
			name string
			file string
		}

		var installed []pluginFile
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			if ext != ".whl" && ext != ".so" && ext != ".dll" && ext != ".dylib" {
				continue
			}
			base := strings.TrimSuffix(name, ext)
			p := parsePluginFilename(base, ext)
			installed = append(installed, pluginFile{name: p.name, file: name})
		}

		if len(installed) == 0 {
			log.Warn("No plugins installed in ./plugins")
			return
		}

		var targets []pluginFile

		if len(args) > 0 {
			target := strings.ToLower(args[0])
			for _, p := range installed {
				if strings.EqualFold(p.name, target) {
					targets = append(targets, p)
				}
			}
			if len(targets) == 0 {
				log.Warnf("Plugin %q not found in ./plugins", args[0])
				return
			}
		} else {
			items := make([]string, len(installed))
			for i, p := range installed {
				items[i] = p.file
			}

			var selected []string
			prompt := &survey.MultiSelect{
				Message:  "Select plugins to remove:",
				Options:  items,
				PageSize: 10,
			}
			if err := survey.AskOne(prompt, &selected); err != nil || len(selected) == 0 {
				return
			}

			selectedSet := make(map[string]bool)
			for _, s := range selected {
				selectedSet[s] = true
			}
			for _, p := range installed {
				if selectedSet[p.file] {
					targets = append(targets, p)
				}
			}
		}

		for _, t := range targets {
			path := filepath.Join("plugins", t.file)
			if err := os.Remove(path); err != nil {
				log.Warnf("Failed to remove %s: %v", t.file, err)
				continue
			}
			log.Infof("Removed: %s", t.file)
		}

		log.SuccessBox("Plugin Removed", fmt.Sprintf("Removed %d file(s)", len(targets)))
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
