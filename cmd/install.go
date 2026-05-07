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

		ext := ".so"
		if runtime.GOOS == "windows" {
			ext = ".dll"
		}

		// =========================
		// DEV BUILD (COMMIT)
		// =========================
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

			url := common.ResolveArtifactURL(target)

			data, err := common.DownloadFile(url)
			if err != nil {
				s.Stop()
				color.Red("Download failed: %v", err)
				return
			}

			_ = os.MkdirAll("plugins", 0755)

			file := filepath.Join(
				"plugins",
				fmt.Sprintf("%s-build%d-%s%s", plugin, target.BuildNumber, commit[:7], ext),
			)

			_ = os.WriteFile(file, data, 0644)

			s.Stop()

			color.Green("Installed dev build %s #%d", plugin, target.BuildNumber)
			color.HiBlack("Saved to: %s", file)
			return
		}

		// =========================
		// VERSION INSTALL
		// =========================
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

			data, err := common.DownloadFile(downloadURL)
			if err != nil {
				s.Stop()
				color.Red("Download failed: %v", err)
				return
			}

			_ = os.MkdirAll("plugins", 0755)

			file := filepath.Join(
				"plugins",
				fmt.Sprintf("%s-%s%s", plugin, version, ext),
			)

			_ = os.WriteFile(file, data, 0644)

			s.Stop()

			color.Green("Installed %s@%s", plugin, version)
			color.HiBlack("Saved to: %s", file)
			return
		}

		// =========================
		// LATEST STABLE
		// =========================
		s.Suffix = fmt.Sprintf(" Fetching %s...", plugin)
		s.Start()

		p, err := client.GetPlugin(plugin)
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

		downloadURL := fmt.Sprintf(
			"%s/download/%s/%s?platform=%s",
			client.BaseURL,
			plugin,
			p.LatestVersion,
			runtime.GOOS,
		)

		data, err := common.DownloadFile(downloadURL)
		if err != nil {
			s.Stop()
			color.Red("Download failed: %v", err)
			return
		}

		_ = os.MkdirAll("plugins", 0755)

		file := filepath.Join(
			"plugins",
			fmt.Sprintf("%s-%s%s", plugin, p.LatestVersion, ext),
		)

		_ = os.WriteFile(file, data, 0644)

		s.Stop()

		color.Green("Installed %s@%s", plugin, p.LatestVersion)
		color.HiBlack("Saved to: %s", file)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
