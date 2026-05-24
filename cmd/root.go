/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"runtime"

	box "github.com/box-cli-maker/box-cli-maker/v3"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var endgit = "EndGit"

var bannerBox = box.NewBox().
		Style(box.Bold).
		Padding(2, 1).
		TitlePosition(box.Top).
		ContentAlign(box.Center).
		Color(box.Cyan).
		TitleColor(box.BrightCyan).
		ContentColor(box.BrightWhite)

var versionBox = box.NewBox().
		Style(box.Single).
		Padding(2, 0).
		TitlePosition(box.Inside).
		ContentAlign(box.Left).
		Color(box.Cyan).
		TitleColor(box.BrightCyan).
		ContentColor(box.BrightWhite)

// Version is set at build time via -ldflags.
// Falls back to "dev" for local builds.
var Version = "0.1.9"

var rootCmd = &cobra.Command{
	Use:     "endgit",
	Version: Version,
	Short:   endgit + " - The package manager for Endstone plugins",
	Long: func() string {
		out := bannerBox.MustRender(endgit, "The package manager for Endstone plugins\n\nA fast and modern CLI for managing Endstone plugins.\nInstall, update, publish, and search plugins\n directly from your terminal.")
		return out
	}(),
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

	// Custom version template with OS/arch info in a box
	rootCmd.SetVersionTemplate(func() string {
		out := versionBox.MustRender(
			endgit+" CLI",
			fmt.Sprintf("Version:  %s\nPlatform: %s/%s", Version, runtime.GOOS, runtime.GOARCH),
		)
		return out
	}())
}
