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
	"github.com/spf13/cobra"

	"github.com/two-tech-dev/endgit-cli/internal/api"
	"github.com/two-tech-dev/endgit-cli/internal/common"
	"github.com/two-tech-dev/endgit-cli/internal/log"
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

func buildFilename(name, version, pluginType string) string {
	ext := resolveExt(pluginType)
	if strings.ToLower(pluginType) == "python" {
		safeName := strings.ReplaceAll(name, "-", "_")
		return fmt.Sprintf("%s-%s-py3-none-any%s", safeName, version, ext)
	}
	return fmt.Sprintf("%s-%s%s", name, version, ext)
}

func downloadAndSave(s *spinner.Spinner, client *api.Client, url string, filename string) error {
	if err := os.MkdirAll("plugins", 0o755); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create plugins directory: %w", err)
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
		return fmt.Errorf("download failed: %w", err)
	}

	s.Stop()
	log.Infof("Saved to: %s", file)
	return nil
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
			s.Stop()
			log.Fatal("Failed to fetch plugin", err)
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
			devVersion := fmt.Sprintf("build%d.%s", target.BuildNumber, commit[:7])
			filename := buildFilename(plugin, devVersion, p.PluginType)

			if err := downloadAndSave(s, client, url, filename); err != nil {
				log.Fatal("Failed to download", err)
			}
			log.Successf("Installed dev build %s #%d", plugin, target.BuildNumber)
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

			filename := buildFilename(plugin, version, p.PluginType)
			if err := downloadAndSave(s, client, downloadURL, filename); err != nil {
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

		filename := buildFilename(plugin, p.LatestVersion, p.PluginType)
		if err := downloadAndSave(s, client, downloadURL, filename); err != nil {
			log.Fatal("Failed to download", err)
		}
		log.Successf("Installed %s@%s", plugin, p.LatestVersion)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
