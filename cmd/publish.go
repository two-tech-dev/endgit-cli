/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a plugin to the EndGit registry",
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Publishing plugins is not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}
