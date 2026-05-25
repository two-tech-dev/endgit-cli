/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/api"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var upgradeCmd = &cobra.Command{
	Use:     "upgrade [plugin]",
	Aliases: []string{"up"},
	Short:   "Upgrade installed plugins to the latest version",
	Long: `Upgrade one or all installed plugins to their latest version.

Examples:
  endgit upgrade              Upgrade all installed plugins
  endgit upgrade my-plugin    Upgrade a specific plugin`,
	Run: func(cmd *cobra.Command, args []string) {
		entries, err := os.ReadDir("plugins")
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn("No plugins directory found.")
				return
			}
			log.Fatal("Failed to read plugins directory", err)
		}

		type installedPlugin struct {
			name    string
			version string
			file    string
		}

		var installed []installedPlugin

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

			if len(args) > 0 && !strings.EqualFold(p.name, args[0]) {
				continue
			}

			installed = append(installed, p)
		}

		if len(installed) == 0 {
			if len(args) > 0 {
				log.Warnf("Plugin %q not found in ./plugins", args[0])
			} else {
				log.Warn("No plugins installed to upgrade.")
			}
			return
		}

		client := api.NewClient()

		type upgradeable struct {
			local  installedPlugin
			remote *api.Plugin
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Checking for updates..."
		s.Start()

		var available []upgradeable
		for _, p := range installed {
			remote, err := client.GetPlugin(p.name)
			if err != nil {
				continue
			}
			if remote.LatestVersion != "" && remote.LatestVersion != p.version {
				available = append(available, upgradeable{local: p, remote: remote})
			}
		}
		s.Stop()

		if len(available) == 0 {
			log.SuccessBox("Up to Date", "All plugins are on the latest version.")
			return
		}

		var toUpgrade []upgradeable

		if len(args) > 0 || len(available) == 1 {
			toUpgrade = available
		} else {
			items := make([]string, len(available))
			for i, u := range available {
				items[i] = fmt.Sprintf("%s  %s → %s", u.local.name, u.local.version, u.remote.LatestVersion)
			}

			var selected []string
			prompt := &survey.MultiSelect{
				Message:  fmt.Sprintf("%d update(s) available. Select plugins to upgrade:", len(available)),
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
			for i, item := range items {
				if selectedSet[item] {
					toUpgrade = append(toUpgrade, available[i])
				}
			}
		}

		upgraded := 0
		for _, u := range toUpgrade {
			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = fmt.Sprintf(" Upgrading %s %s → %s...", u.local.name, u.local.version, u.remote.LatestVersion)
			s.Start()

			downloadURL := fmt.Sprintf(
				"%s/download/%s/%s?platform=%s",
				client.BaseURL,
				u.local.name,
				u.remote.LatestVersion,
				runtime.GOOS,
			)

			newFilename, err := downloadAndSave(s, client, downloadURL)
			if err != nil {
				log.Warnf("Failed to upgrade %s: %v", u.local.name, err)
				continue
			}

			oldPath := filepath.Join("plugins", u.local.file)
			newPath := filepath.Join("plugins", newFilename)
			if oldPath != newPath {
				os.Remove(oldPath)
			}

			log.Successf("Upgraded %s: v%s → v%s", u.local.name, u.local.version, u.remote.LatestVersion)
			upgraded++
		}

		if upgraded > 0 {
			log.Successf("Upgraded %d plugin(s)", upgraded)
		}
	},
}

func parsePluginFilename(base, ext string) struct {
	name    string
	version string
	file    string
} {
	filename := base + ext

	if ext == ".whl" {
		parts := strings.SplitN(base, "-", 3)
		if len(parts) >= 2 {
			name := strings.TrimPrefix(parts[0], "endstone_")
			return struct {
				name    string
				version string
				file    string
			}{name, parts[1], filename}
		}
	}

	parts := strings.Split(base, "-")
	if len(parts) >= 2 {
		version := parts[len(parts)-1]
		name := strings.Join(parts[:len(parts)-1], "-")
		return struct {
			name    string
			version string
			file    string
		}{name, version, filename}
	}

	return struct {
		name    string
		version string
		file    string
	}{base, "unknown", filename}
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
