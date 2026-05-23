/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
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
	Long:  "Checks for the latest release of EndGit and runs the official installer to update.",
	Run: func(cmd *cobra.Command, args []string) {
		repo := "two-tech-dev/endgit-cli"

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Checking for updates..."
		s.Start()

		client := api.NewClient()

		latestTag, err := client.GetLatestReleaseTag(repo)
		if err != nil {
			s.Stop()
			log.Fatal("Failed to check for updates", err)
		}

		currentVersion := strings.TrimPrefix(Version, "v")
		latestVersion := strings.TrimPrefix(latestTag, "v")

		if currentVersion == latestVersion {
			s.Stop()
			log.SuccessBox(
				"Up to Date",
				fmt.Sprintf("You are already on the latest version (%s)", Version),
			)
			return
		}

		s.Stop()
		log.Infof("Updating %s → %s", Version, latestTag)
		fmt.Println()

		s.Suffix = " Running installer..."
		s.Start()

		if runtime.GOOS == "windows" {
			err = runWindowsInstaller()
		} else {
			err = runUnixInstaller()
		}

		s.Stop()

		if err != nil {
			log.Fatal("Update failed", err)
		}

		log.SuccessBox(
			"Update Complete",
			fmt.Sprintf("EndGit has been updated to %s", latestTag),
		)
	},
}

func runUnixInstaller() error {
	c := exec.Command("bash", "-c", "curl -sSL https://endgit.dev/installer.sh | bash")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func runWindowsInstaller() error {
	c := exec.Command("powershell", "-NoProfile", "-Command", "irm https://endgit.dev/installer.ps1 | iex")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
