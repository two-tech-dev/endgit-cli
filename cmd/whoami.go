/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/api"
	"github.com/two-tech-dev/endgit-cli/internal/config"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently logged in user",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()

		if cfg.APIToken == "" {
			log.Warn("Not logged in. Run 'endgit login' to authenticate.")
			return
		}

		client := api.NewClient()
		user, err := client.GetMe()
		if err != nil {
			log.Fatal("Failed to fetch user info", err)
		}

		fmt.Printf("\n  Logged in as %s\n", color.New(color.FgHiWhite, color.Bold).Sprintf("@%s", user.Username))
		if user.DisplayName != "" && user.DisplayName != user.Username {
			fmt.Printf("  %s\n", color.HiBlackString(user.DisplayName))
		}
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
