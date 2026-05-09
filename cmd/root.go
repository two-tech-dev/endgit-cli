/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var endgit = color.New(color.FgCyan, color.Bold).Sprint("EndGit")

const Version = "v0.1.4"

var rootCmd = &cobra.Command{
	Use:     "endgit",
	Version: Version,
	Short:   endgit + " - The package manager for Endstone plugins",
	Long: endgit + ` - The package manager for Endstone plugins

A fast and modern CLI for managing Endstone plugins.
Install, update, publish, and search plugins directly from your terminal.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Invalid command", err)
	}
}

func init() {
	// Disable built-in completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
}
