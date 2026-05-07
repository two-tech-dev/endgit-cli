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

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/two-tech-dev/endgit-cli/internal/api"
	"github.com/two-tech-dev/endgit-cli/internal/common"
)

func resolveExt(pluginType string) string {
	switch strings.ToLower(pluginType) {
	case "python":
		return ".whl"
	default:
		switch runtime.GOOS {
		case "windows":
			return ".dll"
		case "darwin":
			return ".dylib"
		default:
			return ".so"
		}
	}
}

func downloadAndSave(s *spinner.Spinner, client *api.Client, url string, filename string) {
	if err := os.MkdirAll("plugins", 0755); err != nil {
		s.Stop()
		color.Red("Failed to create plugins directory: %v", err)
		os.Exit(1)
	}

	file := filepath.Join("plugins", filename)
	baseSuffix := s.Suffix

	onProgress := func(downloaded, total int64) {
		if total > 0 {
			percent := float64(downloaded) / float64(total) * 100
			s.Suffix = fmt.Sprintf("%s (%.1f%% of %.2fMB)", baseSuffix, percent, float64(total)/(1024*1024))
		} else {
			s.Suffix = fmt.Sprintf("%s (%.2fMB)", baseSuffix, float64(downloaded)/(1024*1024))
		}
	}

	if err := client.DownloadFile(url, file, onProgress); err != nil {
		s.Stop()
		color.Red("Download failed: %v", err)
		os.Exit(1)
	}

	s.Stop()
	color.HiBlack("Saved to: %s", file)
}

var installCmd = &cobra.Command{
	Use:   "install <plugin[@version|@commit]>",
	Short: "Download and install a plugin to the current directory",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		input := args[0]

		var plugin, version, commit string

		parts := strings.Split(input, "@")
		plugin = parts[0]

		if len(parts) > 1 {
			value := parts[1]

			if len(value) >= 7 && common.IsHex(value) {
				commit = value
			} else {
				version = value
			}
		}

		client := api.NewClient()

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		p, err := client.GetPlugin(plugin)
		if err != nil {
			color.Red("Failed to fetch plugin: %v", err)
			return
		}
		ext := resolveExt(p.PluginType)

		// DEV BUILD (COMMIT)
		if commit != "" {

			s.Suffix = fmt.Sprintf(" Searching build %.7s...", commit)
			s.Start()

			buildsResp, err := client.GetBuilds(plugin)
			if err != nil {
				s.Stop()
				color.Red("Failed to fetch builds: %v", err)
				return
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
				color.Yellow("No successful build found for commit %.7s", commit)
				return
			}

			s.Suffix = fmt.Sprintf(" Downloading build #%d...", target.BuildNumber)

			url := target.ResolveArtifactURL()
			filename := fmt.Sprintf("%s-build%d-%s%s", plugin, target.BuildNumber, commit[:7], ext)

			downloadAndSave(s, client, url, filename)
			color.Green("Installed dev build %s #%d", plugin, target.BuildNumber)
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

			filename := fmt.Sprintf("%s-%s%s", plugin, version, ext)
			downloadAndSave(s, client, downloadURL, filename)
			color.Green("Installed %s@%s", plugin, version)
			return
		}

		// LATEST STABLE
		s.Suffix = fmt.Sprintf(" Fetching %s...", plugin)
		s.Start()

		if err != nil {
			s.Stop()
			color.Red("Failed to fetch plugin: %v", err)
			return
		}

		if p.LatestVersion == "" {
			s.Stop()
			color.Red("No published versions found")
			return
		}

		s.Suffix = fmt.Sprintf(" Downloading %s@%s...", plugin, p.LatestVersion)

		downloadURL := fmt.Sprintf(
			"%s/download/%s/%s?platform=%s",
			client.BaseURL,
			plugin,
			p.LatestVersion,
			runtime.GOOS,
		)

		filename := fmt.Sprintf("%s-%s%s", plugin, p.LatestVersion, ext)
		downloadAndSave(s, client, downloadURL, filename)
		color.Green("Installed %s@%s", plugin, p.LatestVersion)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
