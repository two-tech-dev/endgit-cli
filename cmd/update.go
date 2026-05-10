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
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update EndGit to the latest version",
	Long:  "Checks for the latest release of EndGit and automatically downloads and installs it.",
	Run: func(cmd *cobra.Command, args []string) {
		repo := "two-tech-dev/endgit-cli"

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Checking for updates..."
		s.Start()

		client := api.NewClient()

		// Check what the latest version is
		latestTag, err := client.GetLatestReleaseTag(repo)
		if err != nil {
			s.Stop()
			log.Fatal("Failed to check for updates", err)
		}

		// Compare with current version
		currentVersion := strings.TrimPrefix(Version, "v")
		latestVersion := strings.TrimPrefix(latestTag, "v")

		if currentVersion == latestVersion {
			s.Stop()
			log.Successf("You are already on the latest version (%s)", Version)
			return
		}

		if Version == "dev" {
			s.Stop()
			log.Warn("Running a development build — skipping auto-update")
			log.Infof("Latest release: %s", latestTag)
			log.Info("Download from: https://github.com/two-tech-dev/endgit-cli/releases/latest")
			return
		}

		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}
		assetName := fmt.Sprintf("endgit-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)

		latestURL, err := client.GetLatestReleaseAssetURL(repo, assetName)
		if err != nil {
			s.Stop()
			log.Fatal("Failed to find update binary for your platform", err)
		}

		s.Stop()
		log.Infof("Updating %s → %s", Version, latestTag)
		fmt.Println()

		// Resolve install path
		installPath, err := resolveInstallPath()
		if err != nil {
			log.Fatal("Could not determine install path", err)
		}

		installDir := filepath.Dir(installPath)
		if err := os.MkdirAll(installDir, 0o755); err != nil {
			log.Fatal("Failed to create install directory", err)
		}

		// On Windows the running binary is locked and can't be overwritten.
		// Rename it out of the way first (Windows allows renaming locked files).
		oldPath := installPath + ".old"
		if runtime.GOOS == "windows" {
			if _, err := os.Stat(installPath); err == nil {
				os.Remove(oldPath) // clean up any previous .old file
				if err := os.Rename(installPath, oldPath); err != nil {
					log.Fatal("Failed to move current binary for replacement", err)
				}
			}
		}

		s.Suffix = " Downloading update..."
		s.Start()

		if err := client.DownloadFile(latestURL, installPath, nil); err != nil {
			s.Stop()
			// Restore the old binary if download failed
			if runtime.GOOS == "windows" {
				if _, statErr := os.Stat(oldPath); statErr == nil {
					os.Rename(oldPath, installPath)
				}
			}
			log.Fatal("Failed to download update", err)
		}

		// Clean up the old binary
		if runtime.GOOS == "windows" {
			os.Remove(oldPath)
		}

		s.Stop()
		log.Success("EndGit updated successfully!")
		log.Infof("Installed to: %s", installPath)
	},
}

func resolveInstallPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
		}
		return filepath.Join(localAppData, "endgit", "endgit.exe"), nil

	default: // linux, darwin
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not find home directory: %w", err)
		}
		return filepath.Join(home, ".local", "bin", "endgit"), nil
	}
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
