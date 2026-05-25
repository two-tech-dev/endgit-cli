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
	"github.com/two-tech-dev/endgit-cli/internal/common"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

func downloadAndSave(s *spinner.Spinner, client *api.Client, url string, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		s.Stop()
		return "", fmt.Errorf("failed to create plugins directory: %w", err)
	}

	baseSuffix := s.Suffix

	onProgress := func(downloaded, total int64) {
		if total > 0 {
			percent := float64(downloaded) / float64(total) * 100
			s.Suffix = fmt.Sprintf("%s (%.1f%% of %.2fMB)", baseSuffix, percent, float64(total)/(1024*1024))
		} else {
			s.Suffix = fmt.Sprintf("%s (%.2fMB)", baseSuffix, float64(downloaded)/(1024*1024))
		}
	}

	filename, err := client.DownloadFile(url, destDir, onProgress)
	if err != nil {
		s.Stop()
		return "", fmt.Errorf("download failed: %w", err)
	}

	s.Stop()
	log.Infof("Saved to: %s", filepath.Join(destDir, filename))
	return filename, nil
}

func resolveInstallDir() string {
	cwd, err := os.Getwd()
	if err == nil && strings.EqualFold(filepath.Base(cwd), "plugins") {
		return "."
	}

	options := []string{
		"plugins/  (create if needed)",
		".         (current directory)",
	}

	var selected string
	prompt := &survey.Select{
		Message: "Where should the plugin be saved?",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return ""
	}

	if strings.HasPrefix(selected, "plugins/") {
		return "plugins"
	}
	return "."
}

func parsePluginInput(input string) (plugin, version, commit string, err error) {
	parts := strings.SplitN(input, "@", 2)
	plugin = strings.TrimSpace(parts[0])

	if plugin == "" {
		return "", "", "", fmt.Errorf("plugin name cannot be empty")
	}

	if len(parts) > 1 {
		value := strings.TrimSpace(parts[1])
		if value == "" {
			return "", "", "", fmt.Errorf("version or commit hash cannot be empty after @")
		}
		if len(value) >= 7 && common.IsHex(value) {
			commit = value
		} else {
			version = value
		}
	}

	return plugin, version, commit, nil
}

var installCmd = &cobra.Command{
	Use:   "install [plugin[@version|@commit]]",
	Short: "Download and install a plugin to the current directory",
	Long: `Download and install a plugin from the EndGit registry.

Examples:
  endgit install                        Interactive search and install
  endgit install my-plugin              Install the latest stable version
  endgit install my-plugin@1.2.0        Install a specific version
  endgit install my-plugin@abc1234      Install a specific dev build by commit`,
	Run: func(cmd *cobra.Command, args []string) {
		var plugin, version, commit string
		var err error

		if len(args) == 0 {
			plugin = interactiveInstall("")
			if plugin == "" {
				return
			}
		} else {
			input := args[0]
			if strings.Contains(input, "@") {
				plugin, version, commit, err = parsePluginInput(input)
				if err != nil {
					log.Fatal("Invalid input", err)
				}
			} else {
				plugin = interactiveInstall(input)
				if plugin == "" {
					return
				}
			}
		}

		client := api.NewClient()

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Fetching %s...", plugin)
		s.Start()

		p, err := client.GetPlugin(plugin)
		if err != nil {
			s.Stop()
			log.Fatal("Failed to fetch plugin", err)
		}
		s.Stop()

		destDir := resolveInstallDir()
		if destDir == "" {
			return
		}

		// DEV BUILD (COMMIT)
		if commit != "" {
			s.Suffix = fmt.Sprintf(" Searching build %.7s...", commit)
			s.Start()

			buildsResp, err := client.GetBuilds(plugin)
			if err != nil {
				s.Stop()
				log.Fatal("Failed to fetch builds", err)
			}

			var target *api.Build
			for i := range buildsResp.Data.Builds {
				b := &buildsResp.Data.Builds[i]
				if b.CommitHash == commit && b.Status == "SUCCESS" {
					target = b
					break
				}
			}

			if target == nil {
				s.Stop()
				log.Warnf("No successful build found for commit %.7s", commit)
				return
			}

			s.Suffix = fmt.Sprintf(" Downloading build #%d...", target.BuildNumber)

			url := target.ResolveArtifactURL()

			if _, err := downloadAndSave(s, client, url, destDir); err != nil {
				log.Fatal("Failed to download", err)
			}
			log.Successf("Installed %s dev build #%d (%s)", plugin, target.BuildNumber, commit[:7])
			return
		}

		// VERSION INSTALL
		if version != "" {
			s.Suffix = fmt.Sprintf(" Downloading %s@%s...", plugin, version)
			s.Start()

			downloadURL := fmt.Sprintf(
				"%s/download/%s/%s?platform=%s",
				client.BaseURL,
				plugin,
				version,
				runtime.GOOS,
			)

			if _, err := downloadAndSave(s, client, downloadURL, destDir); err != nil {
				log.Fatal("Failed to download", err)
			}
			log.Successf("Installed %s@%s", plugin, version)
			return
		}

		// LATEST STABLE
		s.Suffix = fmt.Sprintf(" Fetching %s...", plugin)
		s.Start()

		if p.LatestVersion == "" {
			s.Stop()
			log.Fatal("No published versions found", nil)
		}

		s.Suffix = fmt.Sprintf(" Downloading %s@%s...", plugin, p.LatestVersion)

		downloadURL := fmt.Sprintf(
			"%s/download/%s/%s?platform=%s",
			client.BaseURL,
			plugin,
			p.LatestVersion,
			runtime.GOOS,
		)

		if _, err := downloadAndSave(s, client, downloadURL, destDir); err != nil {
			log.Fatal("Failed to download", err)
		}
		log.Successf("Installed %s@%s", plugin, p.LatestVersion)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func interactiveInstall(query string) string {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	if query != "" {
		s.Suffix = fmt.Sprintf(" Searching \"%s\"...", query)
	} else {
		s.Suffix = " Fetching available plugins..."
	}
	s.Start()

	client := api.NewClient()
	response, err := client.GetPlugins(query)
	s.Stop()

	if err != nil {
		log.Fatal("Failed to fetch plugins", err)
	}

	plugins := response.Data.Plugins
	if len(plugins) == 0 {
		if query != "" {
			log.Warnf("No plugins found matching \"%s\"", query)
		} else {
			log.Warn("No plugins available")
		}
		return ""
	}

	items := make([]string, len(plugins))
	for i, p := range plugins {
		ver := p.LatestVersion
		if ver == "" {
			ver = "?.?.?"
		}
		items[i] = fmt.Sprintf("%s (v%s) - %d downloads", p.Name, ver, p.Downloads)
	}

	var selected string
	prompt := &survey.Select{
		Message:  "Select a plugin to install:",
		Options:  items,
		PageSize: 10,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return ""
	}

	for i, item := range items {
		if item == selected {
			return plugins[i].Name
		}
	}
	return ""
}
