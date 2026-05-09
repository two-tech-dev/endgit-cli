/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/config"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the EndGit platform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Authenticate with EndGit")

		var token string

		prompt := &survey.Password{
			Message: "Enter your Personal Access Token (PAT):",
		}

		err := survey.AskOne(prompt, &token, survey.WithValidator(func(ans interface{}) error {
			str := ans.(string)
			if len(str) < 10 {
				return fmt.Errorf("invalid token format")
			}
			return nil
		}))

		if err != nil || token == "" {
			log.Warn("Login cancelled")
			return
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Saving credentials..."
		s.Start()

		cfg := config.GetConfig()
		cfg.APIToken = token
		err = config.SaveConfig(cfg)

		s.Stop()

		if err != nil {
			log.Fatal("Failed to save token", err)
		}

		log.Success("Successfully logged in. Token saved locally.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
