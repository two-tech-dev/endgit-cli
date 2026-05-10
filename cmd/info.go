/*
Copyright © 2026 Two Tech Studio
*/
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/two-tech-dev/endgit-cli/internal/api"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

var infoCmd = &cobra.Command{
	Use:   "info <plugin>",
	Short: "Show detailed information about a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Fetching %s...", name)
		s.Start()

		client := api.NewClient()
		p, err := client.GetPlugin(name)

		s.Stop()

		if err != nil {
			log.Fatal("Failed to fetch plugin", err)
		}

		bold := color.New(color.Bold)
		dim := color.New(color.FgHiBlack)

		// Plugin header
		typeBadge := color.HiGreenString("[PY]")
		if strings.ToUpper(p.PluginType) != "PYTHON" {
			typeBadge = color.HiBlueString("[C++]")
		}

		displayName := p.DisplayName
		if displayName == "" {
			displayName = p.Name
		}

		fmt.Println()
		bold.Printf("  %s %s", typeBadge, displayName)
		if p.LatestVersion != "" {
			fmt.Printf("  %s", color.MagentaString("v%s", p.LatestVersion))
		}
		fmt.Println()

		// Slug / Registry name
		dim.Printf("  %s\n", p.Slug)
		fmt.Println()

		// Description
		if p.Description != "" {
			fmt.Printf("  %s\n\n", p.Description)
		}

		// Metadata table
		label := color.New(color.FgCyan)

		printField := func(key, value string) {
			if value != "" {
				label.Printf("  %-14s", key)
				fmt.Println(value)
			}
		}

		printField("Author", formatAuthor(p.Author))
		printField("License", p.License)
		printField("Type", p.PluginType)
		printField("Status", p.Status)
		printField("Downloads", fmt.Sprintf("%d", p.Downloads))
		printField("Stars", fmt.Sprintf("%d", p.Stars))
		printField("Trust Score", formatTrustScore(p.TrustScore))

		if p.RepoURL != "" {
			printField("Repository", p.RepoURL)
		}

		if len(p.Tags) > 0 {
			printField("Tags", strings.Join(p.Tags, ", "))
		}
		if len(p.Keywords) > 0 {
			printField("Keywords", strings.Join(p.Keywords, ", "))
		}

		fmt.Println()

		// Install hint
		fmt.Printf("  Install: %s\n\n",
			color.HiWhiteString("endgit install %s", p.Name),
		)
	},
}

func formatAuthor(a api.Author) string {
	if a.DisplayName != "" && a.DisplayName != a.Username {
		return fmt.Sprintf("%s (%s)", a.DisplayName, a.Username)
	}
	if a.Username != "" {
		return a.Username
	}
	return "unknown"
}

func formatTrustScore(score int) string {
	switch {
	case score >= 80:
		return color.GreenString("%d/100 ★", score)
	case score >= 50:
		return color.YellowString("%d/100", score)
	default:
		return color.RedString("%d/100", score)
	}
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
