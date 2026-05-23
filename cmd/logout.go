/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/config"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of EndGit",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Logging out of EndGit")

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Removing stored credentials..."
		s.Start()

		cfg := config.GetConfig()
		cfg.APIToken = ""
		cfg.RefreshToken = ""
		err := config.SaveConfig(cfg)

		s.Stop()

		if err != nil {
			log.Fatal("Failed to logout", err)
		}

		log.SuccessBox("Logged Out", "Successfully logged out of EndGit")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
