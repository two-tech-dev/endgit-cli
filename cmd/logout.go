/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/config"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication credentials",
	Run: func(cmd *cobra.Command, args []string) {
		color.New(color.FgHiCyan, color.Bold).Println("Logging out of EndGit")

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Removing stored credentials..."
		s.Start()

		err := config.SaveConfig(config.EndGitConfig{
			APIToken: "",
		})

		s.Stop()

		if err != nil {
			color.Red("Failed to logout: %v", err)
			return
		}

		color.Green("Successfully logged out.")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
