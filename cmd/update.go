/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/api"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update EndGit to the latest version",
	Long:  `Checks for the latest release of EndGit and automatically downloads and installs it.`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := "two-tech-dev/endgit-cli"

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Checking for updates..."
		s.Start()

		client := api.NewClient()

		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}
		assetName := fmt.Sprintf("endgit-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)

		latestURL, err := client.GetLatestReleaseAssetURL(repo, assetName)
		if err != nil {
			s.Stop()
			color.Red("Failed to check for updates: %v", err)
			return
		}

		if latestURL == "" {
			s.Stop()
			color.Green("You are already on the latest version of EndGit.")
			return
		}

		s.Stop()
		color.Cyan("→ A new version is available! Installing update...")
		fmt.Println()

		// Resolve install path
		installPath, err := resolveInstallPath()
		if err != nil {
			color.Red("Could not determine install path: %v", err)
			return
		}

		installDir := filepath.Dir(installPath)
		if err := os.MkdirAll(installDir, 0755); err != nil {
			color.Red("Failed to create install directory: %v", err)
			return
		}

		s.Suffix = " Downloading update..."
		s.Start()

		if err := client.DownloadFile(latestURL, installPath, nil); err != nil {
			s.Stop()
			color.Red("Failed to download update: %v", err)
			return
		}

		s.Stop()
		color.Green("EndGit updated successfully!")
		fmt.Printf("  Installed to: %s\n", installPath)
	},
}

func resolveInstallPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("%%LOCALAPPDATA%% is not set")
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
