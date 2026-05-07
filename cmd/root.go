/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var endgit = color.New(color.FgCyan, color.Bold).Sprint("EndGit")

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "endgit",
	Version: Version,
	Short: endgit + " - The package manager for Endstone plugins",
	Long: endgit + ` - The package manager for Endstone plugins

A fast and modern CLI for managing Endstone plugins.
Install, update, publish, and search plugins directly from your terminal.`,
}

func printInvalidCommand(args []string) {
	red := color.New(color.FgRed, color.Bold)

	fmt.Fprintf(
		os.Stderr,
		"%s Invalid command: %s\nSee --help for a list of available commands.\n",
		red.Sprint("Error:"),
		color.New(color.FgWhite).Sprint(strings.Join(args, " ")),
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printInvalidCommand(os.Args[1:])
		os.Exit(1)
	}
}

func init() {
	// Disable built-in completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
}
